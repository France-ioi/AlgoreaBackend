package items

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation GET /items/{item_id}/dependencies items itemDependenciesView
//
//	---
//	summary: Get dependent items for an item
//	description: Lists dependent items for the specified item
//						 and the current user's (or the team's given in `as_team_id`) interactions with them
//						 (from tables `items`, `item_dependencies`, `items_string`, `results`, `permissions_generated`).
//						 Only items visible to the current user (or to the `{as_team_id}` team) are shown.
//						 If `{watched_group_id}` is given, some additional info about the given group's results on the items is shown.
//
//
//						 * The current user (or the team given in `as_team_id`) should have at least 'info' permissions on the specified item,
//							 otherwise the 'forbidden' response is returned.
//
//						 * If `as_team_id` is given, it should be a user's parent team group,
//							 otherwise the "forbidden" error is returned.
//
//						 * If `{watched_group_id}` is given, the user should ba a manager of the group with the 'can_watch_members' permission,
//							 otherwise the "forbidden" error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//		- name: watched_group_id
//			in: query
//			type: integer
//	responses:
//		"200":
//			description: OK. Success response with dependent items
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/prerequisiteOrDependencyItem"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getItemDependencies(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	return srv.getItemPrerequisitesOrDependencies(rw, httpReq, "item_id", "dependent_item_id")
}
