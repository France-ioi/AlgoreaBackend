package groups

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
)

type groupResponseRow struct {
	ID int `json:"id"`
}

func (ctx *Ctx) getAll(w http.ResponseWriter, r *http.Request) {
	groups := ctx.dbGetAllGroups()
	render.Respond(w, r, groups)
}

func (ctx *Ctx) dbGetAllGroups() []groupResponseRow {
	rows, err := ctx.db.Query(`SELECT id FROM groups`)
	if err != nil {
		log.Fatal(err) // TODO
	}
	defer rows.Close()

	var groups []groupResponseRow
	for rows.Next() {
		var group groupResponseRow
		err := rows.Scan(&group.ID)
		if err != nil {
			log.Fatal(err) // TODO
		}
		groups = append(groups, group)
	}
	return groups
}
