package groups

import (
	"crypto/rand"
	"github.com/go-chi/render"
	"math/big"
	"net/http"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) changePassword(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := srv.checkThatUserOwnsTheGroup(user, groupID); apiError != service.NoError {
		return apiError
	}

retry:
	newPassword, err := GenerateGroupPassword()
	service.MustNotBeError(err)

	// `CREATE UNIQUE INDEX sPassword ON groups(sPassword)` must be done
	err = srv.Store.Groups().Where("ID = ?", groupID).Updates(map[string]interface{}{"sPassword": newPassword}).Error()
	if err != nil && strings.Contains(err.Error(), "Duplicate entry") {
		goto retry
	}
	service.MustNotBeError(err)

	render.Respond(w, r, struct {
		Password string `json:"password"`
	}{newPassword})

	return service.NoError
}

// GenerateGroupPassword generate a random password for a group
func GenerateGroupPassword() (string, error) {
	const allowedCharacters = "3456789abcdefghijkmnpqrstuvwxy" // copied from the JS code
	const allowedCharactersLength = len(allowedCharacters)
	const passwordLength = 10

	result := make([]byte, 0, passwordLength)
	for i := 0; i < passwordLength; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(allowedCharactersLength)))
		if err != nil {
			return "", err
		}
		result = append(result, allowedCharacters[index.Int64()])
	}
	return string(result), nil
}
