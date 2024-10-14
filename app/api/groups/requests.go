package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

func renderGroupGroupTransitionResults(w http.ResponseWriter, r *http.Request, results database.GroupGroupTransitionResults) {
	response := service.Response[database.GroupGroupTransitionResults]{
		Success: true,
		Message: "updated",
		Data:    results,
	}
	render.Respond(w, r, &response)
}
