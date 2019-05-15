package testhelpers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"time"

	"bou.ke/monkey"

	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
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
	return nil
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
