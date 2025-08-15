//go:build !unit

package items_test

import (
	"net/http/httptest"
	"sync/atomic"
	"testing"
	_ "unsafe"

	"bou.ke/monkey"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestService_startResult_concurrency(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		items: [{id: 1, type: Task, default_language_tag: 'fr', requires_explicit_entry: 0}]
		groups: [{id: 3, type: User, root_activity_id: 1}]
		users: [{group_id: 3, login: john}]
		permissions_generated: [{group_id: 3, item_id: 1, can_view_generated: content}]
		attempts: [{participant_id: 3, id: 0}]
		sessions: [{session_id: 3, user_id: 3}]
		access_tokens: [{token: 'token_john', session_id: 3, expires_at: '9999-12-31 23:59:59'}]`)
	defer func() { _ = db.Close() }()

	// app server
	application, err := app.New()
	if err != nil {
		t.Fatalf("Unable to load a hooked app: %v", err)
	}
	defer func() { _ = application.Database.Close() }()
	appServer := httptest.NewServer(application.HTTPHandler)
	defer appServer.Close()

	monkey.Patch(service.SchedulePropagation, func(*database.DataStore, string, []string) {})
	defer monkey.UnpatchAll()

	onBeforeInsertingResultInResultStartHook.Store(func() {
		onBeforeInsertingResultInResultStartHook.Store(func() {})
		testhelpers.VerifyTestHTTPRequestWithToken(t, appServer, "token_john", 200,
			"POST", "/items/1/start-result?attempt_id=0", nil, nil)
	})
	defer onBeforeInsertingResultInResultStartHook.Store(func() {})

	testhelpers.VerifyTestHTTPRequestWithToken(t, appServer, "token_john", 200,
		"POST", "/items/1/start-result?attempt_id=0", nil, nil)
}

//nolint:gochecknoglobals // this is a link to a global variable to store the default hook, used for testing purposes only
//go:linkname onBeforeInsertingResultInResultStartHook github.com/France-ioi/AlgoreaBackend/v2/app/api/items.onBeforeInsertingResultInResultStartHook
var onBeforeInsertingResultInResultStartHook atomic.Value
