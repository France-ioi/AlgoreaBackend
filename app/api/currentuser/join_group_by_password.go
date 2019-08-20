package currentuser

import (
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /current-user/group-memberships groups users groupsJoinByPassword
// ---
// summary: Join a team using a password
// description:
//   Lets a user to join a team group by a password.
//   On success the service inserts a row into `groups_groups` (or updates an existing one)
//   with `sType`=`requestAccepted` and `sStatusDate` = current UTC time.
//   It also refreshes the access rights.
//
//   * If there is no team with `bFreeAccess` = 1, `sPasswordEnd` > NOW() (or NULL), and `sPassword` = `password`,
//     the forbidden error is returned.
//
//   * If there is already a row in `groups_groups` with the found team as a parent
//     and the authenticated user’s selfGroup’s ID as a child with `sType`=`invitationAccepted`/`requestAccepted`/`direct`,
//     the unprocessable entity error is returned.
// parameters:
// - name: password
//   in: query
//   type: string
//   required: true
// responses:
//   "201":
//     description: Created. The request has successfully created the group relation.
//     schema:
//       "$ref": "#/definitions/createdResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "422":
//     "$ref": "#/responses/unprocessableEntityResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) joinGroupByPassword(w http.ResponseWriter, r *http.Request) service.APIError {
	password, err := service.ResolveURLQueryGetStringField(r, "password")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	if user.SelfGroupID == nil {
		return service.InsufficientAccessRightsError
	}

	apiError := service.NoError
	var results database.GroupGroupTransitionResults
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var groupInfo struct {
			ID                  int64 `gorm:"column:ID"`
			PasswordEndIsNull   bool  `gorm:"column:bPasswordEndIsNull"`
			PasswordTimerIsNull bool  `gorm:"column:bPasswordTimerIsNull"`
		}
		errInTransaction := store.Groups().WithWriteLock().
			Where("sType = 'Team'").Where("bFreeAccess").
			Where("sPassword LIKE ?", password).
			Where("sPasswordEnd IS NULL OR NOW() < sPasswordEnd").
			Select("ID, sPasswordEnd IS NULL AS bPasswordEndIsNull, sPasswordTimer IS NULL AS bPasswordTimerIsNull").
			Take(&groupInfo).Error()
		if gorm.IsRecordNotFoundError(errInTransaction) {
			logging.GetLogEntry(r).Warnf("A user with ID = %d tried to join a group using a wrong/expired password", user.ID)
			apiError = service.InsufficientAccessRightsError
			return errInTransaction
		}
		service.MustNotBeError(errInTransaction)

		if groupInfo.PasswordEndIsNull && !groupInfo.PasswordTimerIsNull {
			service.MustNotBeError(store.Groups().ByID(groupInfo.ID).
				UpdateColumn("sPasswordEnd", gorm.Expr("ADDTIME(NOW(), sPasswordTimer)")).Error())
		}
		results, errInTransaction = store.GroupGroups().Transition(
			database.UserJoinsGroupByPassword, groupInfo.ID, []int64{*user.SelfGroupID}, user.ID)
		return errInTransaction
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	return service.RenderGroupGroupTransitionResult(w, r, results[*user.SelfGroupID], true, false)
}
