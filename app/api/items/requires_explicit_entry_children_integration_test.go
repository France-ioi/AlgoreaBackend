//go:build !unit

package items_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

// Reproduces / guards against a mismatch where get_item returned requires_explicit_entry=true
// for a child while get_children returned false for the same item (SQL omitted the column).
func TestRequiresExplicitEntry_ChildrenVsItem(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 11, type: User}]
		users: [{group_id: 11, login: jdoe}]
		items:
			- {id: 100, type: Chapter, default_language_tag: en, requires_explicit_entry: 0}
			- {id: 200, type: Chapter, default_language_tag: en, requires_explicit_entry: 1, duration: "03:00:00"}
		items_strings:
			- {item_id: 100, language_tag: en, title: Parent}
			- {item_id: 200, language_tag: en, title: Explicit-entry child}
		items_items:
			- {parent_item_id: 100, child_item_id: 200, child_order: 1}
		permissions_generated:
			- {group_id: 11, item_id: 100, can_view_generated: solution}
			- {group_id: 11, item_id: 200, can_view_generated: solution}
		attempts: [{id: 0, participant_id: 11}]
		results:
			- {attempt_id: 0, participant_id: 11, item_id: 100, started_at: "2020-01-01 00:00:00"}
			- {attempt_id: 0, participant_id: 11, item_id: 200, started_at: "2020-01-01 00:00:00"}
		sessions: [{session_id: 1, user_id: 11}]
		access_tokens: [{token: tok, session_id: 1, expires_at: "9999-12-31 23:59:59"}]
	`)
	defer func() { _ = db.Close() }()

	application, err := app.New()
	require.NoError(t, err)
	defer func() { _ = application.Database.Close() }()
	appServer := httptest.NewServer(application.HTTPHandler)
	defer appServer.Close()

	headers := map[string][]string{"Authorization": {"Bearer tok"}}

	childrenResp, childrenBody, err := testhelpers.SendTestHTTPRequest(
		appServer, "GET", "/items/100/children?attempt_id=0", headers, nil)
	require.NoError(t, err)
	_ = childrenResp.Body.Close()
	require.Equal(t, http.StatusOK, childrenResp.StatusCode, childrenBody)

	itemResp, itemBody, err := testhelpers.SendTestHTTPRequest(
		appServer, "GET", "/items/200", headers, nil)
	require.NoError(t, err)
	_ = itemResp.Body.Close()
	require.Equal(t, http.StatusOK, itemResp.StatusCode, itemBody)

	var children []struct {
		ID                    string `json:"id"`
		RequiresExplicitEntry bool   `json:"requires_explicit_entry"`
	}
	require.NoError(t, json.Unmarshal([]byte(childrenBody), &children))
	require.Len(t, children, 1)
	require.Equal(t, "200", children[0].ID)

	var item struct {
		ID                    string `json:"id"`
		RequiresExplicitEntry bool   `json:"requires_explicit_entry"`
	}
	require.NoError(t, json.Unmarshal([]byte(itemBody), &item))

	t.Logf("children requires_explicit_entry=%v; item requires_explicit_entry=%v",
		children[0].RequiresExplicitEntry, item.RequiresExplicitEntry)

	assert.True(t, item.RequiresExplicitEntry, "get_item should report requires_explicit_entry=true")
	assert.True(t, children[0].RequiresExplicitEntry,
		"get_children should report requires_explicit_entry=true for the same child")
}
