package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /items/{item_id}/attempts/{attempt_id}/publish items resultPublish
//
//	---
//	summary: Publish a result to LTI
//	description: >
//		Publishes score (divided by 100) obtained for the item within the attempt to LTI (via the login module).
//
//
//			Restrictions:
//
//		* if `as_team_id` is given, it should be a user's parent team group,
//		* the current user should have at least 'content' access on each of the `{item_id}` item,
//		* the current user should have non-empty `login_id`,
//
//		otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: attempt_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			description: fails with 'bad request' error if given, this service does not currently support team work
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"200":
//			"$ref": "#/responses/publishedOrFailedResponse"
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
func (srv *Service) publishResult(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var err error

	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	if user.LoginID == nil {
		return service.ErrAPIInsufficientAccessRights
	}
	store := srv.GetStore(httpRequest)

	found, err := store.Permissions().MatchingUserAncestors(user).WherePermissionIsAtLeast("view", "content").
		Where("item_id = ?", itemID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}

	attemptID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if service.ParticipantIDFromContext(httpRequest.Context()) != user.GroupID {
		return service.ErrInvalidRequest(errors.New("the service doesn't support 'as_team_id'"))
	}

	var score float32
	err = store.Results().ByID(user.GroupID, attemptID, itemID).PluckFirst("score_computed", &score).Error()
	if !gorm.IsRecordNotFoundError(err) {
		service.MustNotBeError(err)
	}

	const maxScore = 100.0
	result, err := loginmodule.NewClient(srv.AuthConfig.GetString("loginModuleURL")).SendLTIResult(
		httpRequest.Context(),
		srv.AuthConfig.GetString("clientID"),
		srv.AuthConfig.GetString("clientSecret"),
		*user.LoginID, itemID, score/maxScore,
	)
	service.MustNotBeError(err)

	message := "published"
	if !result {
		message = "failed"
	}
	render.Respond(responseWriter, httpRequest, &service.Response[*struct{}]{Success: result, Message: message})
	return nil
}
