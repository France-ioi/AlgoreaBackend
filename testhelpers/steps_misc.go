package testhelpers

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/token"
	"github.com/France-ioi/AlgoreaBackend/app/tokentest"
)

func (ctx *TestContext) RunFallbackServer() error { // nolint
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Got-Query", r.URL.Path)
	}))
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		return err
	}

	_ = os.Setenv("ALGOREA_REVERSEPROXY.SERVER", backendURL.String()) // nolint
	ctx.setupApp()
	return nil
}

func (ctx *TestContext) IAmUserWithID(id int64) error { // nolint
	ctx.userID = id
	db, err := database.Open(ctx.db())
	if err != nil {
		return err
	}
	return database.NewDataStore(db).Sessions().InsertMap(map[string]interface{}{
		"sAccessToken":    testAccessToken,
		"idUser":          ctx.userID,
		"sExpirationDate": gorm.Expr("? + INTERVAL 7200 SECOND", database.Now()),
	})
}

func (ctx *TestContext) TimeNow(timeStr string) error { // nolint
	testTime, err := time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		monkey.Patch(time.Now, func() time.Time { return testTime })
	}
	return err
}

func (ctx *TestContext) TimeIsFrozen() error { // nolint
	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })
	return nil
}

func (ctx *TestContext) TheGeneratedGroupPasswordIs(generatedPassword string) error { // nolint
	monkey.Patch(groups.GenerateGroupPassword, func() (string, error) { return generatedPassword, nil }) // nolint:unparam
	return nil
}

var multipleStringsRegexp = regexp.MustCompile(`^((?:\s*,\s*)?"([^"]*)")`)

func (ctx *TestContext) TheGeneratedGroupPasswordsAre(generatedPasswords string) error { // nolint
	currentIndex := 0
	monkey.Patch(groups.GenerateGroupPassword, func() (string, error) {
		currentIndex++
		password := multipleStringsRegexp.FindStringSubmatch(generatedPasswords)
		if password == nil {
			return "", errors.New("not enough generated passwords")
		}
		generatedPasswords = generatedPasswords[len(password[1]):]
		return password[2], nil
	})
	return nil
}

func (ctx *TestContext) TheGeneratedAuthKeyIs(generatedString string) error { // nolint
	monkey.Patch(auth.GenerateKey, func() (string, error) { return generatedString, nil }) // nolint:unparam
	return nil
}

func (ctx *TestContext) TheGeneratedAuthKeysAre(generatedStrings string) error { // nolint
	currentIndex := 0
	monkey.Patch(auth.GenerateKey, func() (string, error) {
		currentIndex++
		randomString := multipleStringsRegexp.FindStringSubmatch(generatedStrings)
		if randomString == nil {
			return "", errors.New("not enough generated random strings")
		}
		generatedStrings = generatedStrings[len(randomString[1]):]
		return randomString[2], nil
	})
	return nil
}

func (ctx *TestContext) LogsShouldContain(docString *gherkin.DocString) error { // nolint
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

func (ctx *TestContext) SignedTokenIsDistributed(varName, signerName string, docString *gherkin.DocString) error { // nolint
	var privateKey *rsa.PrivateKey
	signerName = strings.TrimSpace(signerName)
	switch signerName {
	case "the app":
		privateKey = ctx.application.TokenConfig.PrivateKey
	case "the task platform":
		privateKey = tokentest.TaskPlatformPrivateKeyParsed
	default:
		return fmt.Errorf("unknown signer: %q. Only \"the app\" and \"the task platform\" are supported", signerName)
	}

	data, err := ctx.preprocessString(docString.Content)
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

func (ctx *TestContext) TheApplicationConfigIs(body *gherkin.DocString) error { // nolint
	viperConfig := viper.New()
	viperConfig.SetConfigType("yaml")
	preprocessedConfig, err := ctx.preprocessString(body.Content)
	if err != nil {
		return err
	}
	err = viperConfig.MergeConfig(strings.NewReader(preprocessedConfig))
	if err != nil {
		return err
	}
	return viperConfig.UnmarshalExact(ctx.application.Config)
}
