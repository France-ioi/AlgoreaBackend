package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /groups/{parent_group_id}/relations/{child_group_id} group-memberships groupAddChild
//
//	---
//	summary: Add a subgroup
//	description: >
//		Add a group as a child to another group.
//		Lets a group admin add another group as a child and refreshes the access rights afterwards.
//
//
//		Restrictions (otherwise the 'forbidden' error is returned):
//			* the authenticated user should be a manager of both `parent_group_id` and `child_group_id,
//			* the authenticated user should have `can_manage` >= 'memberships' on the `parent_group_id`,
//			* the authenticated user should have `can_manage` = 'memberships_and_group' on the `child_group_id`,
//			* the parent group should not be of type "User" or "Team",
//			* the child group should not be of types "Base" or "User"
//				(since users should join groups only by code or by invitation/request),
//			* the action should not create cycles in the groups relations graph.
//	parameters:
//		- name: parent_group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: child_group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"201":
//			description: Created. The request has successfully created the group relation.
//			schema:
//				"$ref": "#/definitions/createdResponse"
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
func (srv *Service) addChild(w http.ResponseWriter, r *http.Request) error {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	childGroupID, err := service.ResolveURLQueryPathInt64Field(r, "child_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if parentGroupID == childGroupID {
		return service.ErrInvalidRequest(errors.New("a group cannot become its own parent"))
	}

	user := srv.GetUser(r)

	err = srv.GetStore(r).InTransaction(func(s *database.DataStore) error {
		var errInTransaction error
		service.MustNotBeError(checkThatUserHasRightsForDirectRelation(s, user, parentGroupID, childGroupID, createRelation))

		errInTransaction = s.GroupGroups().CreateRelation(parentGroupID, childGroupID)
		if errInTransaction == database.ErrRelationCycle {
			return service.ErrForbidden(errInTransaction) // rollback
		}
		return errInTransaction
	})

	service.MustNotBeError(err)
	service.MustNotBeError(render.Render(w, r, service.CreationSuccess[*struct{}](nil)))

	return nil
}
