package groups

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/go-chi/render"
	"net/http"
)

func renderGroupGroupTransitionResults(w http.ResponseWriter, r *http.Request, results database.GroupGroupTransitionResults) {
	response := service.Response{
		Success: true,
		Message: "updated",
		Data:    results,
	}
	render.Respond(w, r, &response)
}
