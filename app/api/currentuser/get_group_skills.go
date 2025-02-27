package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model skillsViewResponseRow
type skillsViewResponseRow struct {
	*groupInfoForRootItem

	// required: true
	Skill *rootItem `json:"skill"`
}

// swagger:operation GET /current-user/group-memberships/skills group-memberships skillsView
//
//	---
//	summary: List root skills
//	description:
//		If `{watched_group_id}` is not given, the service returns the list of root skills of the groups the current user
//		(or `{as_team_id}`) belongs to or manages.
//		Otherwise, the service returns the list of root skills (visible to the current user or `{as_team_id}`)
//		of all ancestor groups of the watched group which are also
//		ancestors or descendants of at least one group that the current user manages explicitly.
//		Permissions returned for skills are related to the current user (or `{as_team_id}`).
//		Only one of `{as_team_id}` and `{watched_group_id}` can be given.
//
//
//		If `{as_team_id}` is given, it should be a user's parent team group, otherwise the "forbidden" error is returned.
//
//
//		If `{watched_group_id}` is given, the user should ba a manager (implicitly) of the group with the 'can_watch_members' permission,
//		otherwise the "forbidden" error is returned.
//	parameters:
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//		- name: watched_group_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"200":
//			description: OK. Success response with an array of root skills
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/skillsViewResponseRow"
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
func (srv *Service) getRootSkills(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.getRootItems(w, r, false)
}
