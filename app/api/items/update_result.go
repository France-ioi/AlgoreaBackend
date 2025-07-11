package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// resultUpdateRequest is the expected input for result updating
// swagger:model resultUpdateRequest
type resultUpdateRequest struct {
	// Toggle a help request on/off
	HelpRequested bool `json:"help_requested"`
}

// swagger:operation PUT /items/{item_id}/attempts/{attempt_id} items resultUpdate
//
//	---
//	summary: Update attempt result properties
//	description: >
//		Modifies values of an attempt result's properties a participant is able to modify.
//
//		Restrictions:
//
//			* `{as_team_id}` (if given) should be the current user's team,
//			* the participant should have a `results` row for the `{item_id}`-`{attempt_id}` pair,
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
//			in: query
//			type: integer
//			format: int64
//		- in: body
//			name: data
//			required: true
//			description: Result properties to modify
//			schema:
//				"$ref": "#/definitions/resultUpdateRequest"
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
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
func (srv *Service) updateResult(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var err error

	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	attemptID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID := service.ParticipantIDFromContext(httpRequest.Context())

	input := resultUpdateRequest{}
	formData := formdata.NewFormData(&input)
	err = formData.ParseJSONRequestData(httpRequest)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	err = srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		resultScope := store.Results().
			Where("participant_id = ?", participantID).
			Where("attempt_id = ?", attemptID).
			Where("item_id = ?", itemID)
		var found bool
		found, err = resultScope.WithExclusiveWriteLock().HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrAPIInsufficientAccessRights // rollback
		}

		data := formData.ConstructMapForDB()
		if len(data) > 0 {
			service.MustNotBeError(resultScope.UpdateColumn(data).Error())
		}
		return nil
	})
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.UpdateSuccess[*struct{}](nil)))
	return nil
}
