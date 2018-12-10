package groups

import (
	"net/http"

	"github.com/go-chi/render"

	s "github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getAll(w http.ResponseWriter, r *http.Request) s.APIError {

	groups := []struct {
		ID   int    `json:"id"   sql:"column:ID"`
		Name string `json:"name" sql:"column:sName"`
	}{}

	db := srv.Store.Groups().All().Select("ID, sName")
	db = db.Scan(&groups)
	if db.Error != nil {
		return s.ErrUnexpected(db.Error)
	}

	render.Respond(w, r, groups)
	return s.NoError
}
