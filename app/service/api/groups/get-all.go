package groups

import (
	"net/http"

	s "github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/go-chi/render"
)

type GroupResponseRow struct {
	Id   int    `json:"id" db:"ID"`
	Name string `json:"name" db:"sName"`
}

func (srv *GroupsService) getAll(w http.ResponseWriter, r *http.Request) {
	groups := []GroupResponseRow{}
	err := srv.Store.GetAll(&groups)

	if err != nil {
		render.Render(w, r, s.ErrServer(err))
		return
	}
	render.Respond(w, r, groups)
}
