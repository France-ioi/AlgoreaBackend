package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) rejectRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.acceptOrRejectRequests(w, r, database.AdminRefusesRequest)
}
