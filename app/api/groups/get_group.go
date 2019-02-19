package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	query := srv.Store.GroupAncestors().OwnedByUserID(user.UserID).
		Joins("JOIN groups ON idGroupChild=groups.ID").
		Where("idGroupChild = ?", groupID).Select(
		`groups.ID, groups.sName, groups.iGrade, groups.sDescription, groups.sDateCreated, groups.sType,
     groups.sRedirectPath, groups.bOpened, groups.bFreeAccess,
     groups.sPassword, groups.sPasswordTimer, groups.sPasswordEnd, groups.bOpenContest`).Limit(1)

	var result []map[string]interface{}
	if err := query.ScanIntoSliceOfMaps(&result).Error(); err != nil {
		return service.ErrUnexpected(err)
	}
	if len(result) == 0 {
		return service.ErrForbidden(errors.New("insufficient access rights"))
	}
	render.Respond(w, r, service.ConvertMapFromDBToJSON(result[0]))

	return service.NoError
}
