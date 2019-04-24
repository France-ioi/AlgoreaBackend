package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func renderGroupGroupTransitionResults(w http.ResponseWriter, r *http.Request, results database.GroupGroupTransitionResults) {
	response := service.Response{
		Success: true,
		Message: "updated",
		Data:    results,
	}
	render.Respond(w, r, &response)
}
