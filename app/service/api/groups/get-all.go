package groups

import (
	"net/http"

	s "github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/go-chi/render"
)

type groupResponseRow struct {
	ID int `json:"id"`
}

func (ctx *Ctx) getAll(w http.ResponseWriter, r *http.Request) {
	groups, err := ctx.dbGetAllGroups()
	if err != nil {
		render.Render(w, r, s.ErrServer(err))
		return
	}
	render.Respond(w, r, groups)
}

func (ctx *Ctx) dbGetAllGroups() ([]groupResponseRow, error) {
	rows, err := ctx.db.Query(`SELECT id FROM groups`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []groupResponseRow
	for rows.Next() {
		var group groupResponseRow
		err := rows.Scan(&group.ID)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}
