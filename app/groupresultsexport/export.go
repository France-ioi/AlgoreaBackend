// Package groupresultsexport implements the async group-results ZIP export worker.
package groupresultsexport

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/event"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

const (
	errorTokenInvalid   = "token_invalid"
	errorTooManyEntries = "too_many_entries"
	errorUploadFailed   = "upload_failed"
	errorInternal       = "internal"
	statusSuccess       = "success"
	statusFailure       = "failure"
	// Below the worker Lambda's 900s limit so defer can still dispatch completion.
	exportMaxDuration = 14 * time.Minute
	// Caps a single PUT; the request also observes ctx (export deadline / upload_expires_at).
	uploadHTTPClientTimeout = 10 * time.Minute
	userSelectColumns       = `
		users.login,
		users.login_id,
		users.is_admin,
		users.group_id,
		users.access_group_id,
		users.temp_user,
		users.notifications_read_at,
		users.default_language`
)

// RequestedPayload is the inbound EventBridge payload for group_results_export_requested.
type RequestedPayload struct {
	ExportID  string `json:"export_id"`
	Token     string `json:"token"`
	UploadURL string `json:"upload_url"`
	// UploadExpiresAt is Unix milliseconds (plan §3.3); 0 means unset.
	UploadExpiresAt int64 `json:"upload_expires_at"`
}

type zipGeneratorFunc func(
	writer io.Writer, store *database.DataStore, user *database.User, groupID int64, itemParentIDs []int64,
) (groups.GroupProgressZIPMeta, error)

type uploaderFunc func(ctx context.Context, uploadURL, filePath, filename string, sizeBytes int64) error

type userLoaderFunc func(store *database.DataStore, userID int64) (*database.User, error)

// Overridable for unit tests.
var (
	//nolint:gochecknoglobals // overridable in tests
	generateZIPImpl zipGeneratorFunc = generateZIP
	//nolint:gochecknoglobals // overridable in tests
	uploadZIPImpl uploaderFunc = uploadZIP
	//nolint:gochecknoglobals // overridable in tests
	loadUserByIDImpl userLoaderFunc = loadUserByID
	//nolint:gochecknoglobals // overridable in tests; https-only, no redirects, explicit Timeout
	httpClient = newUploadHTTPClient()
	//nolint:gochecknoglobals // overridable in tests
	tempDir = "/tmp"
	//nolint:gochecknoglobals // overridable in tests
	validateExpImpl = validateGroupResultsTokenExp
	//nolint:gochecknoglobals // overridable in tests
	nowUnixMilliImpl = func() int64 { return time.Now().UnixMilli() }
)

func newUploadHTTPClient() *http.Client {
	return &http.Client{
		Timeout: uploadHTTPClientTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// Run processes a group_results_export_requested event: verify the token, build the ZIP,
// PUT it to the presigned URL, and dispatch group_results_export_completed exactly once.
// Export failures are reported via the completion event; Run returns nil so EventBridge does not
// re-run after a completed failure dispatch. Panics are recovered into error=internal.
func Run(
	ctx context.Context, store *database.DataStore, publicKey *rsa.PublicKey, payload RequestedPayload,
) (err error) {
	ctx, cancel := contextWithExportDeadline(ctx, payload.UploadExpiresAt)
	defer cancel()

	completion := newFailureCompletion(payload)

	var tempPath string
	defer func() {
		if recovered := recover(); recovered != nil {
			logging.EntryFromContext(ctx).Errorf("group results export panicked: %v", recovered)
			setFailure(completion, errorInternal)
			err = nil
		}
		if completion["status"] == statusFailure {
			logging.EntryFromContext(ctx).WithField("error", completion["error"]).
				WithField("export_id", completion["export_id"]).
				Error("group results export failed")
		}
		event.Dispatch(ctx, event.TypeGroupResultsExportCompleted, completion)
		if tempPath != "" {
			_ = os.Remove(tempPath)
		}
	}()

	tempPath = runExport(ctx, store, publicKey, payload, completion)
	return nil
}

func contextWithExportDeadline(ctx context.Context, uploadExpiresAtMs int64) (context.Context, context.CancelFunc) {
	deadline := time.Now().Add(exportMaxDuration)
	if uploadExpiresAtMs > 0 {
		uploadExpiry := time.UnixMilli(uploadExpiresAtMs)
		if uploadExpiry.Before(deadline) {
			deadline = uploadExpiry
		}
	}
	return context.WithDeadline(ctx, deadline)
}

func runExport(
	ctx context.Context, store *database.DataStore, publicKey *rsa.PublicKey,
	payload RequestedPayload, completion map[string]interface{},
) (tempPath string) {
	user, groupID, itemParentIDs, ok := prepareExport(ctx, store, publicKey, payload, completion)
	if !ok {
		return ""
	}

	tempFile, createErr := os.CreateTemp(tempDir, "group-results-export-*.zip")
	if createErr != nil {
		setFailure(completion, errorInternal)
		return ""
	}
	tempPath = tempFile.Name()

	meta, failCode := buildAndCloseZIP(tempFile, store, user, groupID, itemParentIDs)
	applyPartialMeta(completion, meta)
	if failCode != "" {
		setFailure(completion, failCode)
		return tempPath
	}

	if err := ctx.Err(); err != nil {
		setFailure(completion, errorInternal)
		return tempPath
	}

	finishExportUpload(ctx, payload.UploadURL, tempPath, meta, completion)
	return tempPath
}

func prepareExport(
	ctx context.Context, store *database.DataStore, publicKey *rsa.PublicKey,
	payload RequestedPayload, completion map[string]interface{},
) (user *database.User, groupID int64, itemParentIDs []int64, ok bool) {
	if payload.ExportID == "" || payload.Token == "" || payload.UploadURL == "" {
		setFailure(completion, errorInternal)
		return nil, 0, nil, false
	}
	if payload.UploadExpiresAt > 0 && nowUnixMilliImpl() >= payload.UploadExpiresAt {
		setFailure(completion, errorUploadFailed)
		return nil, 0, nil, false
	}
	if err := ctx.Err(); err != nil {
		setFailure(completion, errorInternal)
		return nil, 0, nil, false
	}

	user, groupID, itemParentIDs, failCode := resolveAuthorizedExport(store, publicKey, payload.Token, completion)
	if failCode != "" {
		setFailure(completion, failCode)
		return nil, 0, nil, false
	}
	if err := validateHTTPSUploadURL(payload.UploadURL); err != nil {
		logging.EntryFromContext(ctx).WithError(err).Error("group results export rejected upload_url")
		setFailure(completion, errorUploadFailed)
		return nil, 0, nil, false
	}
	if err := ctx.Err(); err != nil {
		setFailure(completion, errorInternal)
		return nil, 0, nil, false
	}
	return user, groupID, itemParentIDs, true
}

func finishExportUpload(
	ctx context.Context, uploadURL, tempPath string, meta groups.GroupProgressZIPMeta,
	completion map[string]interface{},
) {
	fileInfo, statErr := os.Stat(tempPath)
	if statErr != nil {
		setFailure(completion, errorInternal)
		return
	}
	sizeBytes := fileInfo.Size()

	if uploadErr := uploadZIPImpl(ctx, uploadURL, tempPath, meta.Filename, sizeBytes); uploadErr != nil {
		logging.EntryFromContext(ctx).WithError(uploadErr).Error("group results export upload failed")
		if errors.Is(uploadErr, context.DeadlineExceeded) || errors.Is(uploadErr, context.Canceled) {
			setFailure(completion, errorInternal)
		} else {
			setFailure(completion, errorUploadFailed)
		}
		return
	}
	setSuccess(completion, meta, sizeBytes)
}

func newFailureCompletion(payload RequestedPayload) map[string]interface{} {
	return map[string]interface{}{
		"export_id":  payload.ExportID,
		"status":     statusFailure,
		"token":      payload.Token,
		"user_id":    nil,
		"group_id":   nil,
		"group_name": nil,
		"items":      nil,
		"filename":   nil,
		"size_bytes": nil,
		"error":      errorInternal,
	}
}

func setFailure(completion map[string]interface{}, code string) {
	completion["status"] = statusFailure
	completion["error"] = code
	completion["filename"] = nil
	completion["size_bytes"] = nil
}

func applyPartialMeta(completion map[string]interface{}, meta groups.GroupProgressZIPMeta) {
	if meta.GroupName != "" {
		completion["group_name"] = meta.GroupName
	}
	if len(meta.Items) > 0 {
		items := make([]map[string]interface{}, len(meta.Items))
		for i, item := range meta.Items {
			items[i] = map[string]interface{}{"id": item.ID, "title": item.Title}
		}
		completion["items"] = items
	}
}

func setSuccess(completion map[string]interface{}, meta groups.GroupProgressZIPMeta, sizeBytes int64) {
	applyPartialMeta(completion, meta)
	completion["status"] = statusSuccess
	completion["error"] = nil
	completion["filename"] = meta.Filename
	completion["size_bytes"] = sizeBytes
}

func resolveAuthorizedExport(
	store *database.DataStore, publicKey *rsa.PublicKey, tokenStr string, completion map[string]interface{},
) (user *database.User, groupID int64, itemParentIDs []int64, failCode string) {
	tokenPayload, parseErr := parseGroupResultsToken(tokenStr, publicKey)
	if parseErr != nil {
		return nil, 0, nil, errorTokenInvalid
	}
	completion["user_id"] = tokenPayload.UserID
	completion["group_id"] = tokenPayload.GroupID

	userID, parseErr := strconv.ParseInt(tokenPayload.UserID, 10, 64)
	if parseErr != nil {
		return nil, 0, nil, errorTokenInvalid
	}
	groupID, parseErr = strconv.ParseInt(tokenPayload.GroupID, 10, 64)
	if parseErr != nil {
		return nil, 0, nil, errorTokenInvalid
	}
	itemParentIDs, parseErr = parseItemIDs(tokenPayload.ItemIDs)
	if parseErr != nil {
		return nil, 0, nil, errorTokenInvalid
	}

	user, loadErr := loadUserByIDImpl(store, userID)
	if loadErr != nil {
		if gorm.IsRecordNotFoundError(loadErr) {
			return nil, 0, nil, errorTokenInvalid
		}
		return nil, 0, nil, errorInternal
	}
	return user, groupID, itemParentIDs, ""
}

func buildAndCloseZIP(
	tempFile *os.File, store *database.DataStore, user *database.User, groupID int64, itemParentIDs []int64,
) (meta groups.GroupProgressZIPMeta, failCode string) {
	meta, genErr := generateZIPImpl(tempFile, store, user, groupID, itemParentIDs)
	closeErr := tempFile.Close()
	if genErr != nil {
		switch {
		case errors.Is(genErr, groups.ErrTooManyProgressZIPEntries):
			return meta, errorTooManyEntries
		default:
			return meta, errorInternal
		}
	}
	if closeErr != nil {
		return meta, errorInternal
	}
	return meta, ""
}

func parseGroupResultsToken(tokenStr string, publicKey *rsa.PublicKey) (*payloads.GroupResultsToken, error) {
	// ParseAndValidate checks signature and the date claim (yesterday/today/tomorrow UTC).
	// JWT libraries may also enforce exp; we still check payload.Exp explicitly per contract.
	claims, err := token.ParseAndValidate([]byte(tokenStr), publicKey)
	if err != nil {
		return nil, err
	}
	var payload payloads.GroupResultsToken
	if err := payloads.ParseMap(claims, &payload); err != nil {
		return nil, err
	}
	if err := validateExpImpl(payload.Exp); err != nil {
		return nil, err
	}
	return &payload, nil
}

func validateGroupResultsTokenExp(exp int64) error {
	if time.Now().Unix() > exp {
		return errors.New("the token has expired")
	}
	return nil
}

func parseItemIDs(itemIDStrings []string) ([]int64, error) {
	itemIDs := make([]int64, len(itemIDStrings))
	for i, s := range itemIDStrings {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid item_id %q: %w", s, err)
		}
		itemIDs[i] = id
	}
	return itemIDs, nil
}

func loadUserByID(store *database.DataStore, userID int64) (*database.User, error) {
	var user database.User
	err := store.Users().ByID(userID).Select(userSelectColumns).Take(&user).Error()
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func generateZIP(
	writer io.Writer, store *database.DataStore, user *database.User, groupID int64, itemParentIDs []int64,
) (meta groups.GroupProgressZIPMeta, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("zip generation panicked: %v", recovered)
		}
	}()
	return groups.GenerateGroupProgressWithAnswersZIP(writer, store, user, groupID, itemParentIDs)
}

func validateHTTPSUploadURL(uploadURL string) error {
	parsed, err := url.Parse(uploadURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("refusing non-https upload_url")
	}
	return nil
}

// contentDispositionAttachment formats Content-Disposition for the S3 presigned PUT.
// Must match serverless exactly: attachment; filename="<name>" (RFC 6266 / plan §3.4),
// not Go %q quoting which escapes differently.
func contentDispositionAttachment(filename string) string {
	//nolint:gocritic // sprintfQuotedString: must match S3 presign, not Go %q
	return fmt.Sprintf(`attachment; filename="%s"`, filename)
}

func uploadZIP(ctx context.Context, uploadURL, filePath, filename string, sizeBytes int64) error {
	if err := validateHTTPSUploadURL(uploadURL); err != nil {
		return err
	}

	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, file)
	if err != nil {
		return err
	}
	req.ContentLength = sizeBytes
	req.Header.Set("Content-Type", "application/zip")
	req.Header.Set("Content-Disposition", contentDispositionAttachment(filename))

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upload returned HTTP %d", resp.StatusCode)
	}
	return nil
}
