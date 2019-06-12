package groups

import (
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Information of the group to be modified
type GroupUpdateInput struct {
	// Type of the group, among Class, Team, Club, Friends, Other
	// example: Other
	Type        string  `json:"type" sql:"column:sType" valid:"in(Class|Team|Club|Friends|Other)"`
	Name        string  `json:"name" sql:"column:sName"`
	Grade       int32   `json:"grade" sql:"column:iGrade"`
	Description *string `json:"description" sql:"column:sDescription"`
	Opened      bool    `json:"opened" sql:"column:bOpened"`
	// If changed from true to false, automatically switch all requests to join
	// this group from requestSent to requestRefused
	FreeAccess bool `json:"free_access" sql:"column:bFreeAccess"`
	// Duration after the first use of the password it will expire
	// example: 79:56:22
	PasswordTimer *string    `json:"password_timer" sql:"column:sPasswordTimer" valid:"matches(^\\d{2}:[0-5]\\d:[0-5]\\d$)"`
	PasswordEnd   *time.Time `json:"password_end" sql:"column:sPasswordEnd"`
	// Path to which the user should be sent when joining this group.
	// example: /group-join/welcome-joe
	RedirectPath *string `json:"redirect_path" sql:"column:sRedirectPath" valid:"matches(^(\\d+(/\\d+)*)*$)"`
	OpenContest  bool    `json:"open_contest" sql:"column:bOpenContest"`
}

// swagger:operation PUT /groups/{group_id} groupEdit
// ---
// summary: Edit group information
// description: Edit group information.
//   Requires the user to be the owner of the group.
// parameters:
// - name: group_id
//   in: path
//   required: true
//   type: string
//   pattern: \d+
// - name: group information
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/GroupUpdateInput"
// responses:
//   default:
//     "$ref": "#/responses/defaultResponse"
func (srv *Service) updateGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	formData, err := validateUpdateGroupInput(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	apiErr := service.NoError

	err = srv.Store.InTransaction(func(s *database.DataStore) error {
		var errInTransaction error
		groupStore := s.Groups()

		var currentGroupData []struct {
			FreeAccess bool `sql:"column:bFreeAccess"`
		}

		if errInTransaction = groupStore.OwnedBy(user).
			Select("groups.bFreeAccess").WithWriteLock().
			Where("groups.ID = ?", groupID).Limit(1).Scan(&currentGroupData).Error(); errInTransaction != nil {
			if errInTransaction == database.ErrUserNotFound {
				apiErr = service.InsufficientAccessRightsError
				return apiErr.Error // rollback
			}
			return errInTransaction // rollback
		}
		if len(currentGroupData) < 1 {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		dbMap := formData.ConstructMapForDB()
		if errInTransaction := refuseSentGroupRequestsIfNeeded(
			groupStore, groupID, dbMap, currentGroupData[0].FreeAccess); errInTransaction != nil {
			return errInTransaction // rollback
		}

		// update the group
		if errInTransaction := groupStore.Where("ID = ?", groupID).Updates(dbMap).Error(); errInTransaction != nil {
			return errInTransaction // rollback
		}

		return nil // commit
	})

	if apiErr != service.NoError {
		return apiErr
	}

	if err != nil {
		return service.ErrUnexpected(err)
	}

	response := service.Response{Success: true, Message: "updated"}
	render.Respond(w, r, &response)

	return service.NoError
}

// refuseSentGroupRequestsIfNeeded automatically switches all requests to join this group
// from requestSent to requestRefused
// if free_access is changed from true to false
func refuseSentGroupRequestsIfNeeded(
	store *database.GroupStore, groupID int64, dbMap map[string]interface{}, previousFreeAccessValue bool) error {
	// if bFreeAccess is going to be changed from true to false
	if newFreeAccess, ok := dbMap["bFreeAccess"]; ok && !newFreeAccess.(bool) && previousFreeAccessValue {
		// refuse sent group requests
		return store.GroupGroups().
			Where("sType = \"requestSent\"").
			Where("idGroupParent = ?", groupID).
			UpdateColumn("sType", "requestRefused").Error()
	}
	return nil
}

func validateUpdateGroupInput(r *http.Request) (*formdata.FormData, error) {
	formData := formdata.NewFormData(&GroupUpdateInput{})
	return formData, formData.ParseJSONRequestData(r)
}
