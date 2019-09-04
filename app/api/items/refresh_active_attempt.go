package items

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// swagger:operation PUT /items/{item_id}/active-attempt items itemActiveAttemptRefresh
// ---
// summary: Refresh an active attempt and fetch a task token
// description: >
//
//   * If there is no row for the current user and the given item in `users_items`, the service creates one.
//
//   * If the active attempt (`idAttemptActive`) is not set in the `users_items` for the item and the user,
//   the service chooses the most recent one among all the user's attempts (or the team's attempts if
//   `items.bHasAttempts`=1)  for the given item. If no attempts found, the new one gets created and chosen as active.
//
//   * Then `sStartDate` and `sLastActivity` of `groups_attempts` & `user_items` are set to the current time.
//
//   * Finally, the service returns a task token with fresh data for the active attempt for the given item.
//
//
//   Depending on the `items.bHasAttempts` the active attempt is linked to the user's self group (if `items.bHasAttempts`=0)
//   or to the user’s team (if `items.bHasAttempts`=1). The user’s team is a user's parent group with `groups.idTeamItem`
//   pointing to one of the item's ancestors or the item itself.
//
//
//   Restrictions:
//
//     * the user should have at least partial access to the item,
//     * the item should be either 'Task' or 'Course',
//     * for items with `bHasAttempts`=1 the user's team should exist when a new attempt is being created,
//
//   otherwise the 'forbidden' error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     description: "Updated. Success response with the fresh task token"
//     schema:
//       type: object
//       required: [success, message, data]
//       properties:
//         success:
//           description: "true"
//           type: boolean
//           enum: [true]
//         message:
//           description: updated
//           type: string
//           enum: [updated]
//         data:
//           type: object
//           required: [task_token]
//           properties:
//             task_token:
//               type: string
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) refreshActiveAttempt(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	var itemInfo struct {
		HasAttempts       bool    `gorm:"column:bHasAttempts"`
		AccessSolutions   bool    `gorm:"column:accessSolutions"`
		HintsAllowed      bool    `gorm:"column:bHintsAllowed"`
		TextID            *string `gorm:"column:sTextId"`
		URL               string  `gorm:"column:sUrl"`
		SupportedLangProg *string `gorm:"column:sSupportedLangProg"`
	}
	err = srv.Store.Items().Visible(user).Where("ID = ?", itemID).
		Where("partialAccess > 0 OR fullAccess > 0").
		Where("items.sType IN('Task','Course')").
		Select("accessSolutions, bHasAttempts, bHintsAllowed, sTextId, sUrl, sSupportedLangProg").
		Take(&itemInfo).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	var groupsAttemptInfo struct {
		ID               int64   `gorm:"column:ID"`
		HintsRequested   *string `gorm:"column:sHintsRequested"`
		HintsCachedCount int32   `gorm:"column:nbHintsCached"`
	}
	var activeAttemptID *int64
	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		userItemStore := store.UserItems()
		service.MustNotBeError(userItemStore.CreateIfMissing(user.ID, itemID))
		service.MustNotBeError(userItemStore.Where("idUser = ?", user.ID).Where("idItem = ?", itemID).
			WithWriteLock().PluckFirst("idAttemptActive", &activeAttemptID).Error())
		if activeAttemptID == nil {
			groupID := *user.SelfGroupID // not null since we have passed the access rights checking
			if itemInfo.HasAttempts {
				err = store.Groups().TeamGroupForItemAndUser(itemID, user).PluckFirst("groups.ID", &groupID).Error()
				if gorm.IsRecordNotFoundError(err) {
					apiError = service.ErrForbidden(errors.New("no team found for the user"))
					return err // rollback
				}
				service.MustNotBeError(err)
			}
			var attemptID int64
			groupAttemptScope := store.GroupAttempts().
				Where("idGroup = ?", groupID).Where("idItem = ?", itemID)
			err = groupAttemptScope.Order("sLastActivityDate DESC").
				Select("ID, sHintsRequested, nbHintsCached").Limit(1).
				Take(&groupsAttemptInfo).Error()
			if gorm.IsRecordNotFoundError(err) {
				attemptID, err = store.GroupAttempts().CreateNew(groupID, itemID)
				service.MustNotBeError(err)
			} else {
				attemptID = groupsAttemptInfo.ID
				service.MustNotBeError(store.GroupAttempts().ByID(attemptID).UpdateColumn(map[string]interface{}{
					"sStartDate":        gorm.Expr("IFNULL(sStartDate, ?)", database.Now()),
					"sLastActivityDate": database.Now(),
				}).Error())
			}
			activeAttemptID = &attemptID
		}
		service.MustNotBeError(userItemStore.Where("idUser = ?", user.ID).Where("idItem = ?", itemID).
			UpdateColumn(map[string]interface{}{
				"idAttemptActive":   *activeAttemptID,
				"sStartDate":        gorm.Expr("IFNULL(sStartDate, ?)", database.Now()),
				"sLastActivityDate": database.Now(),
			}).Error())
		service.MustNotBeError(store.GroupAttempts().After())
		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	taskToken := token.Task{
		AccessSolutions:    &itemInfo.AccessSolutions,
		SubmissionPossible: ptrBool(true),
		HintsAllowed:       &itemInfo.HintsAllowed,
		HintsRequested:     groupsAttemptInfo.HintsRequested,
		HintsGivenCount:    ptrString(strconv.Itoa(int(groupsAttemptInfo.HintsCachedCount))),
		IsAdmin:            ptrBool(false),
		ReadAnswers:        ptrBool(true),
		UserID:             strconv.FormatInt(user.ID, 10),
		LocalItemID:        strconv.FormatInt(itemID, 10),
		ItemID:             itemInfo.TextID,
		AttemptID:          strconv.FormatInt(*activeAttemptID, 10),
		ItemURL:            itemInfo.URL,
		SupportedLangProg:  itemInfo.SupportedLangProg,
		RandomSeed:         strconv.FormatInt(*activeAttemptID, 10),
		PlatformName:       srv.TokenConfig.PlatformName,
	}
	signedTaskToken, err := taskToken.Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(map[string]interface{}{
		"task_token": signedTaskToken,
	})))
	return service.NoError
}

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool       { return &b }
