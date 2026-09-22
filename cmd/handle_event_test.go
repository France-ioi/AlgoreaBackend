package cmd

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/event"
	"github.com/France-ioi/AlgoreaBackend/v2/app/groupresultsexport"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestHandleEventJSON_UnknownDetailType(t *testing.T) {
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	err := handleEventJSON(ctx, &app.Application{}, []byte(`{"detail-type":"something_else","detail":{"type":"something_else","payload":{}}}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown event detail type")
}

func TestHandleEventJSON_InvalidJSON(t *testing.T) {
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	err := handleEventJSON(ctx, &app.Application{}, []byte(`{`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid EventBridge event JSON")
}

func TestHandleEventJSON_InvalidDetail(t *testing.T) {
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	err := handleEventJSON(ctx, &app.Application{}, []byte(`{"detail-type":"x","detail":"not-an-object"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid EventBridge detail JSON")
}

func TestHandleEventJSON_DetailTypeFromDetail(t *testing.T) {
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	err := handleEventJSON(ctx, &app.Application{}, []byte(`{"detail":{"type":"unhandled","payload":{}}}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unhandled")
}

func TestHandleEventJSON_GroupResultsExportRequested(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, _, hook := logging.NewContextWithNewMockLogger()
	var ran bool
	monkey.Patch(groupresultsexport.Run, func(
		ctx context.Context, _ *database.DataStore, _ *rsa.PublicKey, payload groupresultsexport.RequestedPayload,
	) error {
		ran = true
		assert.Equal(t, "inbound-req-id", middleware.GetReqID(ctx))
		assert.Equal(t, "export-1", payload.ExportID)
		assert.Equal(t, "tok", payload.Token)
		assert.Equal(t, "https://upload.example", payload.UploadURL)
		assert.Equal(t, int64(1758132000000), payload.UploadExpiresAt)
		return nil
	})
	monkey.Patch(app.TokenConfig, func(*viper.Viper) (*token.Config, error) {
		return &token.Config{PublicKey: tokentest.AlgoreaPlatformPublicKeyParsed()}, nil
	})
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	err := handleEventJSON(ctx, &app.Application{Database: db}, []byte(`{
		"detail-type":"group_results_export_requested",
		"detail":{"type":"group_results_export_requested","request_id":"inbound-req-id","payload":{
			"export_id":"export-1","token":"tok","upload_url":"https://upload.example",
			"upload_expires_at":1758132000000
		}}
	}`))
	require.NoError(t, err)
	assert.True(t, ran)
	require.NoError(t, mock.ExpectationsWereMet())

	foundHandlingLog := false
	for _, entry := range hook.AllEntries() {
		if entry.Message == "handling event" {
			foundHandlingLog = true
			assert.Equal(t, "inbound-req-id", entry.Data["req_id"])
			break
		}
	}
	assert.True(t, foundHandlingLog)
}

func TestHandleEventJSON_MissingRequestID(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, mockDispatcher := testContextWithMockDispatcher(t)
	var ran bool
	monkey.Patch(groupresultsexport.Run, func(
		context.Context, *database.DataStore, *rsa.PublicKey, groupresultsexport.RequestedPayload,
	) error {
		ran = true
		return nil
	})
	defer monkey.UnpatchAll()

	err := handleEventJSON(ctx, &app.Application{}, []byte(`{
		"detail-type":"group_results_export_requested",
		"detail":{"type":"group_results_export_requested","payload":{
			"export_id":"export-1","token":"tok","upload_url":"https://upload.example"
		}}
	}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing request_id in event envelope")
	assert.False(t, ran)
	assert.Empty(t, mockDispatcher.GetEvents())
}

func TestHandleEventJSON_PropagatesRequestIDToCompletionEvent(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	err := handleEventJSON(ctx, &app.Application{}, []byte(`{
		"detail-type":"group_results_export_requested",
		"detail":{"type":"group_results_export_requested","request_id":"envelope-req-id","payload":{
			"export_id":"e1","token":"tok","upload_url":{"nested":true}
		}}
	}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid group_results_export_requested payload")
	events := mock.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, event.TypeGroupResultsExportCompleted, events[0].Type)
	assert.Equal(t, "envelope-req-id", events[0].RequestID)
}

func TestHandleEventJSON_EmptyDetail(t *testing.T) {
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	err := handleEventJSON(ctx, &app.Application{},
		[]byte(`{"detail-type":"group_results_export_requested"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing request_id in event envelope")
}

func TestHandleGroupResultsExportRequested_InvalidPayload(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	err := handleGroupResultsExportRequested(ctx, &app.Application{}, json.RawMessage(`{"export_id":[]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid group_results_export_requested payload")
	assert.Empty(t, mock.GetEvents())
}

func TestHandleGroupResultsExportRequested_InvalidPayload_DispatchesWhenRecoverable(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	// export_id is valid string; upload_url wrong type → unmarshal fails but we recover ids.
	err := handleGroupResultsExportRequested(ctx, &app.Application{}, json.RawMessage(
		`{"export_id":"e1","token":"tok","upload_url":{"nested":true}}`,
	))
	require.Error(t, err)
	events := mock.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, event.TypeGroupResultsExportCompleted, events[0].Type)
	assert.Equal(t, "test-req-id", events[0].RequestID)
	assert.Equal(t, "failure", events[0].Payload["status"])
	assert.Equal(t, "internal", events[0].Payload["error"])
	assert.Equal(t, "e1", events[0].Payload["export_id"])
	assert.Equal(t, "tok", events[0].Payload["token"])
}

func TestHandleGroupResultsExportRequested_InvalidPayload_NoDispatchWithoutToken(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	err := handleGroupResultsExportRequested(ctx, &app.Application{}, json.RawMessage(
		`{"export_id":"e1","upload_url":{"nested":true}}`,
	))
	require.Error(t, err)
	assert.Empty(t, mock.GetEvents())
}

func TestHandleGroupResultsExportRequested_TokenConfigError_EmptyPayload(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, mock := testContextWithMockDispatcher(t)
	monkey.Patch(app.TokenConfig, func(*viper.Viper) (*token.Config, error) {
		return nil, errors.New("token config broken")
	})
	defer monkey.UnpatchAll()

	err := handleGroupResultsExportRequested(ctx, &app.Application{Config: viper.New()}, json.RawMessage(`{}`))
	require.Error(t, err)
	assert.Empty(t, mock.GetEvents())
}

func TestHandleGroupResultsExportRequested_TokenConfigError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, mock := testContextWithMockDispatcher(t)
	expected := errors.New("token config broken")
	monkey.Patch(app.TokenConfig, func(*viper.Viper) (*token.Config, error) {
		return nil, expected
	})
	defer monkey.UnpatchAll()

	err := handleGroupResultsExportRequested(ctx, &app.Application{Config: viper.New()}, json.RawMessage(
		`{"export_id":"e","token":"t","upload_url":"https://x.example"}`,
	))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load token config")
	events := mock.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "test-req-id", events[0].RequestID)
	assert.Equal(t, "internal", events[0].Payload["error"])
}

func TestHandleGroupResultsExportRequested_RunError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	ctx, _, _ := logging.NewContextWithNewMockLogger()
	expected := errors.New("run failed")
	monkey.Patch(app.TokenConfig, func(*viper.Viper) (*token.Config, error) {
		return &token.Config{PublicKey: tokentest.AlgoreaPlatformPublicKeyParsed()}, nil
	})
	monkey.Patch(groupresultsexport.Run, func(
		context.Context, *database.DataStore, *rsa.PublicKey, groupresultsexport.RequestedPayload,
	) error {
		return expected
	})
	defer monkey.UnpatchAll()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	err := handleGroupResultsExportRequested(ctx, &app.Application{Database: db}, json.RawMessage(
		`{"export_id":"e","token":"t","upload_url":"https://x.example"}`,
	))
	assert.Equal(t, expected, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunHandleEventCommand_AppNewError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	expected := errors.New("boot failed")
	monkey.Patch(app.New, func(...*logging.Logger) (*app.Application, error) {
		return nil, expected
	})
	defer monkey.UnpatchAll()

	assert.Equal(t, expected, runHandleEventCommand(&cobra.Command{}, nil))
}

func TestRunHandleEventCommand_Success(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockDispatcher := event.NewMockDispatcher()
	application := &app.Application{
		Database:        db,
		EventDispatcher: mockDispatcher,
		EventInstance:   "test",
		Config:          viper.New(),
	}
	monkey.Patch(app.New, func(...*logging.Logger) (*app.Application, error) {
		return application, nil
	})
	var ran bool
	monkey.Patch(groupresultsexport.Run, func(
		ctx context.Context, _ *database.DataStore, _ *rsa.PublicKey, _ groupresultsexport.RequestedPayload,
	) error {
		ran = true
		assert.NotNil(t, event.DispatcherFromContext(ctx))
		assert.Equal(t, "cli-req-id", middleware.GetReqID(ctx))
		return nil
	})
	monkey.Patch(app.TokenConfig, func(*viper.Viper) (*token.Config, error) {
		return &token.Config{PublicKey: tokentest.AlgoreaPlatformPublicKeyParsed()}, nil
	})
	defer monkey.UnpatchAll()

	oldStdin := os.Stdin
	stdinReader, stdinWriter, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = stdinReader
	defer func() { os.Stdin = oldStdin }()

	go func() {
		_, _ = stdinWriter.WriteString(`{
			"detail-type":"group_results_export_requested",
			"detail":{"type":"group_results_export_requested","request_id":"cli-req-id","payload":{
				"export_id":"e","token":"t","upload_url":"https://x.example"
			}}
		}`)
		_ = stdinWriter.Close()
	}()

	require.NoError(t, runHandleEventCommand(&cobra.Command{}, []string{"test"}))
	assert.True(t, ran)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunHandleEventCommand_StdinReadError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	application := &app.Application{Database: db}
	monkey.Patch(app.New, func(...*logging.Logger) (*app.Application, error) {
		return application, nil
	})
	defer monkey.UnpatchAll()

	oldStdin := os.Stdin
	stdinReader, stdinWriter, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = stdinReader
	defer func() { os.Stdin = oldStdin }()
	require.NoError(t, stdinWriter.Close())
	require.NoError(t, stdinReader.Close())

	err = runHandleEventCommand(&cobra.Command{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read event from stdin")
	require.NoError(t, mock.ExpectationsWereMet())
}

func testContextWithMockDispatcher(t *testing.T) (context.Context, *event.MockDispatcher) {
	t.Helper()
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	mock := event.NewMockDispatcher()
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-req-id")
	return event.ContextWithDispatcher(ctx, mock), mock
}
