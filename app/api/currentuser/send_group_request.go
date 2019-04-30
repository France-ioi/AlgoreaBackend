package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) sendGroupRequest(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, createGroupRequestAction)
}
