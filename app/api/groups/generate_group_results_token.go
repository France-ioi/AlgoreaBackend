package groups

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

const groupResultsTokenLifetime = time.Hour

// swagger:operation POST /groups/{group_id}/group-results-token groups groupResultsTokenGenerate
//
//	---
//	summary: Generate a group results token
//	description: >
//		Generates a signed JWS token that authorizes obtaining the group's progress-with-answers export
//		for the given `{parent_item_ids}` (async export flow).
//
//
//		Restrictions (same as `groupGroupProgressWithAnswersZIP`):
//
//		* The current user should be a manager of the group (or of one of its ancestors)
//		  with `can_watch_members` set to true,
//
//		* The current user should have `can_watch` >= 'answer' on each of `{parent_item_ids}` items,
//
//		* The export is limited to 100 users and 100 items in the visible descendant subtree.
//		  If either limit is exceeded, a distinct 400 error is returned.
//
//		Otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: parent_item_ids
//			in: query
//			type: array
//			required: true
//			items:
//				type: integer
//				format: int64
//	responses:
//		"201":
//			"$ref": "#/responses/groupResultsTokenResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) generateGroupResultsToken(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if !user.CanWatchGroupMembers(store, groupID) {
		return service.ErrAPIInsufficientAccessRights
	}

	itemParentIDs, err := resolveAndCheckParentIDs(store, httpRequest, user, "answer")
	service.MustNotBeError(err)

	// checkGroupProgressZIPLimits only returns ErrTooManyItems/UsersInProgressZIP (or nil);
	// unexpected DB failures panic via MustNotBeError inside the helpers.
	if err := checkGroupProgressZIPLimits(store, user, groupID, itemParentIDs); err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemIDs := make([]string, len(itemParentIDs))
	for i, id := range itemParentIDs {
		itemIDs[i] = strconv.FormatInt(id, 10)
	}

	expiresIn := int32(groupResultsTokenLifetime / time.Second)
	expirationTime := time.Now().Add(groupResultsTokenLifetime)

	groupResultsToken, err := (&token.Token[payloads.GroupResultsToken]{Payload: payloads.GroupResultsToken{
		UserID:  strconv.FormatInt(user.GroupID, 10),
		GroupID: strconv.FormatInt(groupID, 10),
		ItemIDs: itemIDs,
		Exp:     expirationTime.Unix(),
	}}).Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.CreationSuccess(map[string]interface{}{
		"group_results_token": groupResultsToken,
		"expires_in":          expiresIn,
	})))

	return nil
}
