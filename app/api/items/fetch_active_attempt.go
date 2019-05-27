package items

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
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
	if err = user.Load(); err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

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

	var userItemInfo struct {
		ActiveAttemptID  *int64  `gorm:"column:idAttemptActive"`
		HintsRequested   *string `gorm:"column:sHintsRequested"`
		HintsCachedCount int32   `gorm:"column:nbHintsCached"`
	}
	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		userItemStore := store.UserItems()
		service.MustNotBeError(userItemStore.CreateIfMissing(user.UserID, itemID))
		service.MustNotBeError(userItemStore.Where("idUser = ?", user.UserID).Where("idItem = ?", itemID).
			WithWriteLock().Select("idAttemptActive, sHintsRequested, nbHintsCached").
			Take(&userItemInfo).Error())
		if userItemInfo.ActiveAttemptID == nil {
			groupID, _ := user.SelfGroupID()
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
			err = groupAttemptScope.Order("sLastActivityDate DESC").Limit(1).
				PluckFirst("ID", &attemptID).Error()
			if gorm.IsRecordNotFoundError(err) {
				attemptID, err = store.GroupAttempts().CreateNew(groupID, itemID)
				service.MustNotBeError(err)
			} else {
				service.MustNotBeError(store.GroupAttempts().ByID(attemptID).UpdateColumn(map[string]interface{}{
					"sStartDate":        gorm.Expr("IFNULL(sStartDate, NOW())"),
					"sLastActivityDate": gorm.Expr("NOW()"),
				}).Error())
			}
			userItemInfo.ActiveAttemptID = &attemptID
		}
		service.MustNotBeError(userItemStore.Where("idUser = ?", user.UserID).Where("idItem = ?", itemID).
			UpdateColumn(map[string]interface{}{
				"idAttemptActive":   *userItemInfo.ActiveAttemptID,
				"sStartDate":        gorm.Expr("IFNULL(sStartDate, NOW())"),
				"sLastActivityDate": gorm.Expr("NOW()"),
			}).Error())
		service.MustNotBeError(store.GroupAttempts().After())
		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	var tokenAccessSolutions *formdata.Anything
	if itemInfo.AccessSolutions {
		tokenAccessSolutions = formdata.AnythingFromString(`"1"`)
	} else {
		tokenAccessSolutions = formdata.AnythingFromString(`"0"`)
	}

	boolValues := map[bool]string{false: "0", true: "1"}

	taskToken := token.Task{
		AccessSolutions:    tokenAccessSolutions,
		SubmissionPossible: func(b bool) *bool { return &b }(true),
		HintsAllowed:       ptrString(boolValues[itemInfo.HintsAllowed]),
		HintsRequested:     userItemInfo.HintsRequested,
		HintsGivenCount:    ptrString(strconv.Itoa(int(userItemInfo.HintsCachedCount))),
		IsAdmin:            formdata.AnythingFromString("false"),
		ReadAnswers:        formdata.AnythingFromString("true"),
		UserID:             strconv.FormatInt(user.UserID, 10),
		LocalItemID:        strconv.FormatInt(itemID, 10),
		ItemID:             itemInfo.TextID,
		AttemptID:          ptrString(strconv.FormatInt(*userItemInfo.ActiveAttemptID, 10)),
		ItemURL:            itemInfo.URL,
		SupportedLangProg:  itemInfo.SupportedLangProg,
		RandomSeed:         strconv.FormatInt(*userItemInfo.ActiveAttemptID, 10),
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
