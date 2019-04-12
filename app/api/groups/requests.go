package groups

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/go-chi/render"
	"net/http"
)

type groupGroupTransitionJSONResults struct {
	Success   map[int64]bool `json:"success,omitempty"`
	Unchanged map[int64]bool `json:"unchanged,omitempty"`
	Invalid   map[int64]bool `json:"invalid,omitempty"`
	Cycle     map[int64]bool `json:"cycle,omitempty"`
}

func renderGroupGroupTransitionResults(w http.ResponseWriter, r *http.Request, results *database.GroupGroupTransitionResults) {
	response := service.Response{
		Success: true,
		Message: "updated",
		Data: groupGroupTransitionJSONResults{
			Success:   results.Success,
			Unchanged: results.Unchanged,
			Invalid:   results.Invalid,
			Cycle:     results.Cycle,
		},
	}
	render.Respond(w, r, &response)
}
