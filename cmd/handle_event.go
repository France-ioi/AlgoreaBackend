package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/event"
	"github.com/France-ioi/AlgoreaBackend/v2/app/groupresultsexport"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

func init() { //nolint:gochecknoinits // cobra suggests using init functions to add commands
	handleEventCmd := &cobra.Command{
		Use:   "handle-event [environment]",
		Short: "handle an EventBridge event from stdin",
		Long: `Reads an EventBridge event JSON from stdin, dispatches on detail-type,
and runs the matching worker (e.g. group_results_export_requested).`,
		Args: cobra.MaximumNArgs(1),
		RunE: runHandleEventCommand,
	}
	rootCmd.AddCommand(handleEventCmd)
}

type eventBridgeEnvelope struct {
	// EventBridge uses kebab-case for this field.
	DetailType string          `json:"detail-type"` //nolint:tagliatelle // EventBridge envelope field name
	Detail     json.RawMessage `json:"detail"`
}

type eventDetailEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func runHandleEventCommand(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		appenv.SetEnv(args[0])
	}
	appenv.SetDefaultEnv("dev")

	application, err := app.New()
	defer func() {
		if application != nil && application.Database != nil {
			_ = application.Database.Close()
		}
	}()
	if err != nil {
		return err
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("cannot read event from stdin: %w", err)
	}

	ctx := application.ContextWithEventDispatcher(createContextWithLogger(application.Config))
	return handleEventJSON(ctx, application, raw)
}

func handleEventJSON(ctx context.Context, application *app.Application, raw []byte) error {
	var envelope eventBridgeEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("invalid EventBridge event JSON: %w", err)
	}

	detailType := envelope.DetailType
	var detail eventDetailEnvelope
	if len(envelope.Detail) > 0 {
		if err := json.Unmarshal(envelope.Detail, &detail); err != nil {
			return fmt.Errorf("invalid EventBridge detail JSON: %w", err)
		}
		if detailType == "" {
			detailType = detail.Type
		}
	}

	logging.EntryFromContext(ctx).WithField("detail_type", detailType).Info("handling event")

	switch detailType {
	case "group_results_export_requested":
		return handleGroupResultsExportRequested(ctx, application, detail.Payload)
	default:
		logging.EntryFromContext(ctx).WithField("detail_type", detailType).Error("unknown event detail type")
		return fmt.Errorf("unknown event detail type: %s", detailType)
	}
}

func handleGroupResultsExportRequested(
	ctx context.Context, application *app.Application, payloadRaw json.RawMessage,
) error {
	var payload groupresultsexport.RequestedPayload
	if len(payloadRaw) > 0 {
		if err := json.Unmarshal(payloadRaw, &payload); err != nil {
			dispatchRecoverableExportFailure(ctx, payloadRaw)
			return fmt.Errorf("invalid group_results_export_requested payload: %w", err)
		}
	}

	tokenConfig, err := app.TokenConfig(application.Config)
	if err != nil {
		dispatchExportFailureCompletion(ctx, payload, "internal")
		return fmt.Errorf("unable to load token config: %w", err)
	}

	store := database.NewDataStore(application.Database)
	return groupresultsexport.Run(ctx, store, tokenConfig.PublicKey, payload)
}

func dispatchRecoverableExportFailure(ctx context.Context, payloadRaw json.RawMessage) {
	var partial struct {
		ExportID string `json:"export_id"`
		Token    string `json:"token"`
	}
	if err := json.Unmarshal(payloadRaw, &partial); err != nil || partial.ExportID == "" || partial.Token == "" {
		return
	}
	dispatchExportFailureCompletion(ctx, groupresultsexport.RequestedPayload{
		ExportID: partial.ExportID,
		Token:    partial.Token,
	}, "internal")
}

func dispatchExportFailureCompletion(
	ctx context.Context, payload groupresultsexport.RequestedPayload, errorCode string,
) {
	if payload.ExportID == "" || payload.Token == "" {
		return
	}
	event.Dispatch(ctx, event.TypeGroupResultsExportCompleted, map[string]interface{}{
		"export_id":  payload.ExportID,
		"status":     "failure",
		"token":      payload.Token,
		"user_id":    nil,
		"group_id":   nil,
		"group_name": nil,
		"items":      nil,
		"filename":   nil,
		"size_bytes": nil,
		"error":      errorCode,
	})
}
