package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation GET /current-user/dump users currentUserDataExport
//
//	---
//	summary: Export the short version of the current user's data
//	description: >
//		Returns a downloadable JSON file with all the short version of the current user's data.
//		The content returned is just the dump of raw entries of tables related to the user
//
//		* `current_user` (from `users`): all attributes;
//		* `managed_groups`: `id` and `name` for every descendant of groups managed by the user;
//		* `joined_groups`: `id` and `name` for every ancestor of user’s `group_id`;
//		* `groups_groups`: where the user’s `group_id` is the `child_group_id`, all attributes + `groups.name`;
//		* `group_managers`: where the user’s `group_id` is the `manager_id`, all attributes + `groups.name`;
//
//		In case of unexpected error (e.g. a DB error), the response will be a malformed JSON like
//		```{"current_user":{"success":false,"message":"Internal Server Error","error_text":"Some error"}```
//	produces:
//		- application/json
//	responses:
//		"200":
//			description: The returned data dump file
//			schema:
//				type: file
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getDump(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.getDumpCommon(r, w, false)
}
