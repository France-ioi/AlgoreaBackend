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

func (srv *Service) fetchActiveAttempt(w http.ResponseWriter, r *http.Request) service.APIError {
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
				err = store.Groups().TeamGroupByItemAndUser(itemID, user).PluckFirst("groups.ID", &groupID).Error()
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
