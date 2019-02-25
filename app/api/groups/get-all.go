package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getAll(w http.ResponseWriter, r *http.Request) service.APIError {

	var groups []struct {
		ID   int    `json:"id"   sql:"column:ID"`
		Name string `json:"name" sql:"column:sName"`
	}

	db := srv.Store.Groups().Select("ID, sName")
	db = db.Scan(&groups)
	if db.Error() != nil {
		return service.ErrUnexpected(db.Error())
	}

	render.Respond(w, r, groups)
	return service.NoError
}
