//go:build !prod

package testhelpers

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bou.ke/monkey"
	"github.com/cucumber/messages-go/v10"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/token"
	"github.com/France-ioi/AlgoreaBackend/app/tokentest"
)

// IAmUserWithID sets the current logged user to the one with the provided ID.
func (ctx *TestContext) IAmUserWithID(userID int64) error {
	ctx.userID = userID
	ctx.user = strconv.FormatInt(userID, 10)

	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")
		return store.Sessions().InsertMap(map[string]interface{}{
			"access_token": testAccessToken,
			"user_id":      ctx.userID,
			"issued_at":    database.Now(),
			"expires_at":   gorm.Expr("? + INTERVAL 7200 SECOND", database.Now()),
		})
	})
}

// TimeNow stubs time.Now to the provided time.
func (ctx *TestContext) TimeNow(timeStr string) error {
	testTime, err := time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		monkey.Patch(time.Now, func() time.Time { return testTime })
	}
	return err
}

// TimeIsFrozen stubs time.Now to the current time.
func (ctx *TestContext) TimeIsFrozen() error {
	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })
	return nil
}

// TheGeneratedGroupCodeIs stubs groups.GenerateGroupCode to return the provided code instead of a random one.
func (ctx *TestContext) TheGeneratedGroupCodeIs(generatedCode string) error {
	monkey.Patch(groups.GenerateGroupCode, func() (string, error) { return generatedCode, nil })
	return nil
}

var multipleStringsRegexp = regexp.MustCompile(`^((?:\s*,\s*)?"([^"]*)")`)

// TheGeneratedGroupCodesAre stubs groups.GenerateGroupCode to generate the provided codes instead of random ones.
// generatedCodes is in the following form:
// example for three codes: "code1","code2","code3"
// with an arbitrary number of codes.
func (ctx *TestContext) TheGeneratedGroupCodesAre(generatedCodes string) error {
	currentIndex := 0
	monkey.Patch(groups.GenerateGroupCode, func() (string, error) {
		currentIndex++
		code := multipleStringsRegexp.FindStringSubmatch(generatedCodes)
		if code == nil {
			return "", errors.New("not enough generated codes")
		}
		generatedCodes = generatedCodes[len(code[1]):]
		return code[2], nil
	})
	return nil
}

// TheGeneratedAuthKeyIs stubs auth.GenerateKey to return the provided auth key instead of a random one.
func (ctx *TestContext) TheGeneratedAuthKeyIs(generatedKey string) error {
	monkey.Patch(auth.GenerateKey, func() (string, error) { return generatedKey, nil })
	return nil
}

// LogsShouldContain checks that the logs contain a provided string.
func (ctx *TestContext) LogsShouldContain(docString *messages.PickleStepArgument_PickleDocString) error {
	preprocessed, err := ctx.preprocessString(docString.Content)
	if err != nil {
		return err
	}
	stringToSearch := strings.TrimSpace(preprocessed)
	logs := ctx.logsHook.GetAllLogs()
	if !strings.Contains(logs, stringToSearch) {
		return fmt.Errorf("cannot find %q in logs:\n%s", stringToSearch, logs)
	}
	return nil
}

// getPrivateKeyOf gets the test private key  of the app or the task platform.
func (ctx *TestContext) getPrivateKeyOf(signerName string) *rsa.PrivateKey {
	var privateKey *rsa.PrivateKey
	signerName = strings.TrimSpace(signerName)
	switch signerName {
	case "the app":
		config, _ := app.TokenConfig(ctx.application.Config)
		privateKey = config.PrivateKey
	case "the task platform":
		privateKey = tokentest.TaskPlatformPrivateKeyParsed
	default:
		panic(fmt.Errorf("unknown signer: %q. Only \"the app\" and \"the task platform\" are supported", signerName))
	}

	return privateKey
}

// SignedTokenIsDistributed declares a signed token and puts it in a global variable.
// This allows later use inside a request, or a comparison with a response.
func (ctx *TestContext) SignedTokenIsDistributed(
	varName, signerName string,
	jsonPayload *messages.PickleStepArgument_PickleDocString,
) error {
	privateKey := ctx.getPrivateKeyOf(signerName)

	data, err := ctx.preprocessString(jsonPayload.Content)
	if err != nil {
		return err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return err
	}

	ctx.templateSet.AddGlobal(varName, token.Generate(payload, privateKey))

	return nil
}

// TheApplicationConfigIs specifies variables of the app configuration given in YAML format.
func (ctx *TestContext) TheApplicationConfigIs(yamlConfig *messages.PickleStepArgument_PickleDocString) error {
	config := viper.New()
	config.SetConfigType("yaml")
	preprocessedConfig, err := ctx.preprocessString(yamlConfig.Content)
	if err != nil {
		return err
	}
	err = config.MergeConfig(strings.NewReader(preprocessedConfig))
	if err != nil {
		return err
	}

	// Only 'domain' and 'auth' changes are currently supported
	if config.IsSet("auth") {
		ctx.application.ReplaceAuthConfig(config)
	}
	if config.IsSet("domains") {
		ctx.application.ReplaceDomainsConfig(config)
	}

	return nil
}

// TheContextVariableIs sets a context variable in the request http.Request as the provided value.
// Can be retrieved from the request with r.Context().Value("variableName").
func (ctx *TestContext) TheContextVariableIs(variableName, value string) error {
	preprocessed, err := ctx.preprocessString(value)
	if err != nil {
		return err
	}

	oldHTTPHandler := ctx.application.HTTPHandler
	ctx.application.HTTPHandler = chi.NewRouter().With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			oldHTTPHandler.ServeHTTP(writer, request.WithContext(context.WithValue(request.Context(), variableName, preprocessed)))
		})
	}).(*chi.Mux)
	ctx.application.HTTPHandler.Mount("/", oldHTTPHandler)
	return nil
}
