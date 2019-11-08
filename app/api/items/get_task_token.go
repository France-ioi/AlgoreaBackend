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

// swagger:operation GET /items/{item_id}/task-token items itemTaskTokenGet
// ---
// summary: Get a task token with a refreshed active attempt
// description: >
//
//   * If there is no row for the current user and the given item in `users_items`, the service creates one.
//
//   * If the active attempt (`active_attempt_id`) is not set in the `users_items` for the item and the user,
//   the service chooses the most recent one among all the user's attempts (or the team's attempts if
//   `items.has_attempts`=1)  for the given item. If no attempts found, the new one gets created and chosen as active.
//
//   * Then `started_at` (if it is NULL) and `latest_activity_at` of `groups_attempts` & `user_items` are set to the current time.
//
//   * Finally, the service returns a task token with fresh data for the active attempt for the given item.
//
//
//   Depending on the `items.has_attempts` the active attempt is linked to the user's self group (if `items.has_attempts`=0)
//   or to the user’s team (if `items.has_attempts`=1). The user’s team is a user's parent group with `groups.team_item_id`
//   pointing to one of the item's ancestors or the item itself.
//
//
//   Restrictions:
//
//     * the user should have at least partial access to the item,
//     * the item should be either 'Task' or 'Course',
//     * for items with `has_attempts`=1 the user's team should exist when a new attempt is being created,
//
//   otherwise the 'forbidden' error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     description: "OK. Success response with the fresh task token"
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
func (srv *Service) getTaskToken(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	var itemInfo struct {
		HasAttempts       bool
		AccessSolutions   bool
		HintsAllowed      bool
		TextID            *string
		URL               string
		SupportedLangProg *string
	}
	err = srv.Store.Items().ByID(itemID).WhereUserHasViewPermissionOnItems(user, "content").
		Where("items.type IN('Task','Course')").
		Select(`
			can_view_generated_value = ? AS access_solutions,
			has_attempts, hints_allowed, text_id, url, supported_lang_prog`,
			srv.Store.PermissionsGranted().ViewIndexByKind("solution")).
		Take(&itemInfo).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	var groupsAttemptInfo struct {
		ID               int64
		HintsRequested   *string
		HintsCachedCount int32 `gorm:"column:hints_cached"`
	}
	var activeAttemptID *int64
	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		userItemStore := store.UserItems()
		err = userItemStore.Where("user_id = ?", user.GroupID).Where("item_id = ?", itemID).
			WithWriteLock().PluckFirst("active_attempt_id", &activeAttemptID).Error()

		// No active attempt set in `users_items` so we should choose or create one
		if gorm.IsRecordNotFoundError(err) {
			groupID := user.GroupID

			// if items.has_attempts = 1, we use use a team group instead of the user's self group
			if itemInfo.HasAttempts {
				err = store.Groups().TeamGroupForItemAndUser(itemID, user).PluckFirst("groups.id", &groupID).Error()
				if gorm.IsRecordNotFoundError(err) {
					apiError = service.ErrForbidden(errors.New("no team found for the user"))
					return err // rollback
				}
				service.MustNotBeError(err)
			}

			// find the freshest one among all the group's attempts for the item
			var attemptID int64
			groupAttemptScope := store.GroupAttempts().
				Where("group_id = ?", groupID).Where("item_id = ?", itemID)
			err = groupAttemptScope.Order("latest_activity_at DESC").
				Select("id, hints_requested, hints_cached").Limit(1).
				Take(&groupsAttemptInfo).Error()

			// if no attempt found, create a new one
			if gorm.IsRecordNotFoundError(err) {
				attemptID, err = store.GroupAttempts().CreateNew(groupID, itemID)
				service.MustNotBeError(err)
			} else { // otherwise, update groups_attempts.started_at (if it is NULL) & groups_attempts.latest_activity_at
				attemptID = groupsAttemptInfo.ID
			}
			activeAttemptID = &attemptID
		}
		service.MustNotBeError(err)

		// update groups_attempts
		service.MustNotBeError(store.GroupAttempts().ByID(*activeAttemptID).UpdateColumn(map[string]interface{}{
			"started_at":         gorm.Expr("IFNULL(started_at, ?)", database.Now()),
			"latest_activity_at": database.Now(),
		}).Error())

		// update users_items.active_attempt_id
		service.MustNotBeError(userItemStore.SetActiveAttempt(user.GroupID, itemID, *activeAttemptID))

		// propagate compute users_items
		service.MustNotBeError(store.GroupAttempts().ComputeAllGroupAttempts())

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
		UserID:             strconv.FormatInt(user.GroupID, 10),
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

	render.Respond(w, r, map[string]interface{}{
		"task_token": signedTaskToken,
	})
	return service.NoError
}

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool       { return &b }
