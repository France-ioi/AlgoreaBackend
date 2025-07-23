//go:build !prod

package testhelpers

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bou.ke/monkey"
	"github.com/cucumber/godog"
	"github.com/go-chi/chi"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
)

// TimeNow stubs time.Now and mocks the DB function NOW() with the provided time.
func (ctx *TestContext) TimeNow(timeStr string) error {
	if err := ctx.ServerTimeNow(timeStr); err != nil {
		return err
	}
	ctx.dbTimePatches = append(ctx.dbTimePatches, MockDBTime(time.Now().UTC().Format(time.DateTime+".999999999")))
	return nil
}

// ServerTimeNow stubs time.Now with the provided time.
func (ctx *TestContext) ServerTimeNow(timeStr string) error {
	timeStr = ctx.preprocessString(timeStr)
	testTime, err := time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		monkey.Patch(time.Now, func() time.Time { return testTime })
	}
	return err
}

// TimeIsFrozen stubs time.Now with the current time and mocks the DB function NOW().
func (ctx *TestContext) TimeIsFrozen() error {
	if err := ctx.ServerTimeIsFrozen(); err != nil {
		return err
	}
	ctx.dbTimePatches = append(ctx.dbTimePatches, MockDBTime(time.Now().UTC().Format(time.DateTime+".999999999")))
	return nil
}

// ServerTimeIsFrozen stubs time.Now with the current time.
func (ctx *TestContext) ServerTimeIsFrozen() error {
	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })
	return nil
}

// TheGeneratedGroupCodeIs stubs groups.GenerateGroupCode to return the provided code instead of a random one.
func (ctx *TestContext) TheGeneratedGroupCodeIs(generatedCode string) error {
	monkey.Patch(groups.GenerateGroupCode, func() (string, error) { return generatedCode, nil })
	return nil
}

// TheGeneratedGroupCodesAre stubs groups.GenerateGroupCode to generate the provided codes instead of random ones.
// generatedCodes is in the following form:
// example for three codes: "code1","code2","code3"
// with an arbitrary number of codes.
func (ctx *TestContext) TheGeneratedGroupCodesAre(generatedCodes string) error {
	ctx.generatedGroupCodeIndex = -1
	var parsedGeneratedCode []string
	if err := json.Unmarshal([]byte(fmt.Sprintf("[%s]", generatedCodes)), &parsedGeneratedCode); err != nil {
		return err
	}
	monkey.Patch(groups.GenerateGroupCode, func() (string, error) {
		ctx.generatedGroupCodeIndex++

		if ctx.generatedGroupCodeIndex >= len(parsedGeneratedCode) {
			return "", errors.New("not enough generated codes")
		}
		return parsedGeneratedCode[ctx.generatedGroupCodeIndex], nil
	})
	return nil
}

// TheGeneratedAuthKeyIs stubs auth.GenerateKey to return the provided auth key instead of a random one.
func (ctx *TestContext) TheGeneratedAuthKeyIs(generatedKey string) error {
	monkey.Patch(auth.GenerateKey, func() (string, error) { return generatedKey, nil })
	return nil
}

// LogsShouldContain checks that the logs contain a provided string.
func (ctx *TestContext) LogsShouldContain(docString *godog.DocString) error {
	preprocessed := ctx.preprocessString(docString.Content)
	stringToSearch := strings.TrimSpace(preprocessed)
	logs := ctx.logsHook.GetAllStructuredLogs()
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
		privateKey = tokentest.TaskPlatformPrivateKeyParsed()
	default:
		panic(fmt.Errorf("unknown signer: %q. Only \"the app\" and \"the task platform\" are supported", signerName))
	}

	return privateKey
}

// SignedTokenIsDistributed declares a signed token and puts it in a global variable.
// This allows later use inside a request, or a comparison with a response.
func (ctx *TestContext) SignedTokenIsDistributed(
	varName, signerName string,
	jsonPayload *godog.DocString,
) error {
	privateKey := ctx.getPrivateKeyOf(signerName)

	data := ctx.preprocessString(jsonPayload.Content)

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return err
	}

	ctx.templateSet.AddGlobal(varName, token.Generate(payload, privateKey))

	return nil
}

// FalsifiedSignedTokenIsDistributed generates a falsified token and sets it in a global template variable.
func (ctx *TestContext) FalsifiedSignedTokenIsDistributed(
	varName, signerName string,
	jsonPayload *godog.DocString,
) error {
	privateKey := ctx.getPrivateKeyOf(signerName)

	data := ctx.preprocessString(jsonPayload.Content)

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return err
	}

	generatedToken := token.Generate(payload, privateKey)

	// A token is of the following form:
	// HEADER.PAYLOAD.SIGNATURE (separated by dots)
	// To falsify the token, we increment the last byte of the payload.
	lastPayloadPosition := strings.LastIndex(string(generatedToken), ".") - 1
	generatedToken[lastPayloadPosition]++ // falsify the token.

	ctx.templateSet.AddGlobal(varName, generatedToken)

	return nil
}

// TheApplicationConfigIs specifies variables of the app configuration given in YAML format.
func (ctx *TestContext) TheApplicationConfigIs(yamlConfig *godog.DocString) error {
	config := viper.New()
	config.SetConfigType("yaml")
	preprocessedConfig := ctx.preprocessString(ctx.replaceReferencesWithIDs(yamlConfig.Content))
	err := config.MergeConfig(strings.NewReader(preprocessedConfig))
	if err != nil {
		return err
	}

	// Only 'domain' and 'auth' changes are currently supported
	if config.IsSet("auth") {
		ctx.application.ReplaceAuthConfig(config, ctx.logger)
	}
	if config.IsSet("domains") {
		ctx.application.ReplaceDomainsConfig(config, ctx.logger)
	}

	return nil
}

// TheContextVariableIs sets a context variable in the request http.Request as the provided value.
// Can be retrieved from the request with r.Context().Value(service.APIServiceContextVariableName("variableName")).
func (ctx *TestContext) TheContextVariableIs(variableName, value string) error {
	preprocessed := ctx.preprocessString(value)

	oldHTTPHandler := ctx.application.HTTPHandler
	ctx.application.HTTPHandler = chi.NewRouter().With(func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			oldHTTPHandler.ServeHTTP(writer, request.WithContext(context.WithValue(request.Context(),
				service.APIServiceContextVariableName(variableName), preprocessed)))
		})
	}).(*chi.Mux)
	ctx.application.HTTPHandler.Mount("/", oldHTTPHandler)
	return nil
}
