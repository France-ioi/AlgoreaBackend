package groups

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"strings"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

const maxNumberOfRetriesForCodeGenerator = 3

// swagger:operation POST /groups/{group_id}/code groups groupCodeCreate
//
//	---
//	summary: Create a new group code
//	description: >
//
//		Creates a new code using a set of allowed characters [3456789abcdefghijkmnpqrstuvwxy].
//		Makes sure it doesn’t correspond to any existing group code. Saves it for the given group and returns it.
//
//
//		The authenticated user should be a manager of `group_id` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"200":
//			description: OK. The new code has been set.
//			schema:
//				type: object
//				properties:
//					code:
//						type: string
//				required:
//					- code
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
func (srv *Service) createCode(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var err error
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	service.MustNotBeError(checkThatUserCanManageTheGroupMemberships(store, user, groupID))

	var newCode string
	service.MustNotBeError(store.InTransaction(func(store *database.DataStore) error {
		for retryCount := 1; ; retryCount++ {
			if retryCount > maxNumberOfRetriesForCodeGenerator {
				generatorErr := errors.New("the code generator is broken")
				logging.GetLogEntry(httpRequest).Error(generatorErr)
				return generatorErr
			}

			newCode, err = GenerateGroupCode()
			service.MustNotBeError(err)

			err = store.Groups().Where("id = ?", groupID).UpdateColumn(map[string]interface{}{"code": newCode}).Error()
			if err != nil && strings.Contains(err.Error(), "Duplicate entry") {
				continue
			}
			service.MustNotBeError(err)

			break
		}
		return nil
	}))

	render.Respond(responseWriter, httpRequest, struct {
		Code string `json:"code"`
	}{newCode})

	return nil
}

// GenerateGroupCode generate a random code for a group.
func GenerateGroupCode() (string, error) {
	const allowedCharacters = "3456789abcdefghijkmnpqrstuvwxy" // copied from the JS code
	const allowedCharactersLength = len(allowedCharacters)
	const codeLength = 10

	result := make([]byte, 0, codeLength)
	for i := 0; i < codeLength; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(allowedCharactersLength)))
		if err != nil {
			return "", err
		}
		result = append(result, allowedCharacters[index.Int64()])
	}
	return string(result), nil
}
