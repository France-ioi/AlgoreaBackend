package groups

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"strings"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) changePassword(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	var newPassword string
	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		for retryCount := 1; ; retryCount++ {
			if retryCount > 3 {
				generatorErr := errors.New("the password generator is broken")
				logging.GetLogEntry(r).Error(generatorErr)
				return generatorErr
			}

			newPassword, err = GenerateGroupPassword()
			service.MustNotBeError(err)

			// `CREATE UNIQUE INDEX sPassword ON groups(sPassword)` must be done
			err = store.Groups().Where("ID = ?", groupID).Updates(map[string]interface{}{"sPassword": newPassword}).Error()
			if err != nil && strings.Contains(err.Error(), "Duplicate entry") {
				continue
			}
			service.MustNotBeError(err)

			break
		}
		return nil
	}))

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
