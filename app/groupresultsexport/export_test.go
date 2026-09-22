package groupresultsexport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/event"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
)

func signedGroupResultsToken(t *testing.T, payload *payloads.GroupResultsToken) string {
	t.Helper()
	signed, err := (&token.Token[payloads.GroupResultsToken]{Payload: *payload}).
		Sign(tokentest.AlgoreaPlatformPrivateKeyParsed())
	require.NoError(t, err)
	return signed
}

func testContextWithMockDispatcher(t *testing.T) (context.Context, *event.MockDispatcher) {
	t.Helper()
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	mock := event.NewMockDispatcher()
	return event.ContextWithDispatcher(ctx, mock), mock
}

func tlsUploadClient(server *httptest.Server) *http.Client {
	client := server.Client()
	client.Timeout = uploadHTTPClientTimeout
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}

func TestLoadUserByID(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)

	mock.ExpectQuery("^" + regexp.QuoteMeta(
		"SELECT "+strings.TrimSpace(userSelectColumns)+" FROM `users` WHERE (users.group_id = ?) LIMIT 1",
	)).
		WithArgs(int64(21)).
		WillReturnRows(sqlmock.NewRows([]string{
			"login", "login_id", "is_admin", "group_id", "access_group_id",
			"temp_user", "notifications_read_at", "default_language",
		}).AddRow("owner", nil, false, int64(21), nil, false, nil, "en"))

	user, err := loadUserByID(store, 21)
	require.NoError(t, err)
	assert.Equal(t, int64(21), user.GroupID)
	assert.Equal(t, "owner", user.Login)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadUserByID_NotFound(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)

	mock.ExpectQuery("SELECT .* FROM `users`").
		WithArgs(int64(404)).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := loadUserByID(store, 404)
	require.Error(t, err)
	assert.True(t, gorm.IsRecordNotFoundError(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadUserByID_DBError(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	store := database.NewDataStore(db)

	mock.ExpectQuery("SELECT .* FROM `users`").
		WithArgs(int64(21)).
		WillReturnError(errors.New("db down"))

	_, err := loadUserByID(store, 21)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestParseGroupResultsToken_Success(t *testing.T) {
	tokenStr := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID:  "123",
		GroupID: "456",
		ItemIDs: []string{"210", "220"},
		Exp:     time.Now().Add(time.Hour).Unix(),
	})
	payload, err := parseGroupResultsToken(tokenStr, tokentest.AlgoreaPlatformPublicKeyParsed())
	require.NoError(t, err)
	assert.Equal(t, "123", payload.UserID)
	assert.Equal(t, "456", payload.GroupID)
	assert.Equal(t, []string{"210", "220"}, payload.ItemIDs)
	assert.NotEmpty(t, payload.Date, "date claim is auto-set by token.Generate")
}

func TestParseGroupResultsToken_ExpCheckFailure(t *testing.T) {
	old := validateExpImpl
	defer func() { validateExpImpl = old }()
	validateExpImpl = func(int64) error { return errors.New("the token has expired") }

	tokenStr := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID: "123", GroupID: "456", ItemIDs: []string{}, Exp: time.Now().Add(time.Hour).Unix(),
	})
	_, err := parseGroupResultsToken(tokenStr, tokentest.AlgoreaPlatformPublicKeyParsed())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestValidateGroupResultsTokenExp(t *testing.T) {
	require.NoError(t, validateGroupResultsTokenExp(time.Now().Add(time.Hour).Unix()))
	err := validateGroupResultsTokenExp(time.Now().Add(-time.Hour).Unix())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestParseGroupResultsToken_InvalidSignature(t *testing.T) {
	_, err := parseGroupResultsToken("not-a-token", tokentest.AlgoreaPlatformPublicKeyParsed())
	require.Error(t, err)
}

func TestParseItemIDs(t *testing.T) {
	ids, err := parseItemIDs([]string{"1", "2"})
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2}, ids)

	_, err = parseItemIDs([]string{"abc"})
	require.Error(t, err)
}

func TestContentDispositionAttachment(t *testing.T) {
	assert.Equal(t,
		`attachment; filename="groups_progress_with_answers_for_group-11-and_child_items_of-210.zip"`,
		contentDispositionAttachment("groups_progress_with_answers_for_group-11-and_child_items_of-210.zip"),
	)
}

func TestValidateHTTPSUploadURL(t *testing.T) {
	require.NoError(t, validateHTTPSUploadURL("https://bucket.s3.amazonaws.com/key?X-Amz-Signature=x"))
	require.Error(t, validateHTTPSUploadURL("http://bucket.s3.amazonaws.com/key"))
	require.Error(t, validateHTTPSUploadURL("https:///no-host"))
	require.Error(t, validateHTTPSUploadURL("://bad"))
}

func TestUploadZIP_SuccessAndFailure(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "upload-test-*.zip")
	require.NoError(t, err)
	_, err = tmpFile.WriteString("zip-content")
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	info, err := os.Stat(tmpFile.Name())
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
			assert.Equal(t, http.MethodPut, req.Method)
			assert.Equal(t, "application/zip", req.Header.Get("Content-Type"))
			assert.Equal(t, `attachment; filename="export.zip"`, req.Header.Get("Content-Disposition"))
			assert.Equal(t, info.Size(), req.ContentLength)
			body, readErr := io.ReadAll(req.Body)
			assert.NoError(t, readErr)
			assert.Equal(t, "zip-content", string(body))
			responseWriter.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		oldClient := httpClient
		httpClient = tlsUploadClient(server)
		defer func() { httpClient = oldClient }()
		assert.NoError(t, uploadZIP(context.Background(), server.URL, tmpFile.Name(), "export.zip", info.Size()))
	})

	t.Run("http rejected", func(t *testing.T) {
		err := uploadZIP(context.Background(), "http://127.0.0.1:1/put", tmpFile.Name(), "export.zip", info.Size())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "refusing non-https")
	})

	t.Run("non2xx", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
			responseWriter.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()
		oldClient := httpClient
		httpClient = tlsUploadClient(server)
		defer func() { httpClient = oldClient }()
		err := uploadZIP(context.Background(), server.URL, tmpFile.Name(), "export.zip", info.Size())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})

	t.Run("302 not followed", func(t *testing.T) {
		var targetHits atomic.Int32
		target := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			targetHits.Add(1)
		}))
		defer target.Close()

		redirect := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
			responseWriter.Header().Set("Location", target.URL)
			responseWriter.WriteHeader(http.StatusFound)
		}))
		defer redirect.Close()

		oldClient := httpClient
		httpClient = tlsUploadClient(redirect)
		defer func() { httpClient = oldClient }()

		err := uploadZIP(context.Background(), redirect.URL, tmpFile.Name(), "export.zip", info.Size())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "302")
		assert.Equal(t, int32(0), targetHits.Load())
	})

	t.Run("missing file", func(t *testing.T) {
		err := uploadZIP(context.Background(), "https://example.com/put",
			filepath.Join(os.TempDir(), "missing-file.zip"), "x.zip", 0)
		require.Error(t, err)
	})

	t.Run("bad url", func(t *testing.T) {
		err := uploadZIP(context.Background(), "://bad", tmpFile.Name(), "x.zip", info.Size())
		require.Error(t, err)
	})
}

func TestGenerateZIP_RecoversPanic(t *testing.T) {
	_, err := generateZIP(io.Discard, nil, &database.User{}, 1, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "panicked")
}

func TestRun_Success(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	tokenStr := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID:  "21",
		GroupID: "11",
		ItemIDs: []string{"210"},
		Exp:     time.Now().Add(time.Hour).Unix(),
	})

	oldLoad, oldGen, oldUpload := loadUserByIDImpl, generateZIPImpl, uploadZIPImpl
	defer func() {
		loadUserByIDImpl, generateZIPImpl, uploadZIPImpl = oldLoad, oldGen, oldUpload
	}()

	loadUserByIDImpl = func(_ *database.DataStore, userID int64) (*database.User, error) {
		assert.Equal(t, int64(21), userID)
		return &database.User{GroupID: 21, Login: "owner"}, nil
	}
	filename := "groups_progress_with_answers_for_group-11-and_child_items_of-210.zip"
	generateZIPImpl = func(
		writer io.Writer, _ *database.DataStore, _ *database.User, groupID int64, itemParentIDs []int64,
	) (groups.GroupProgressZIPMeta, error) {
		assert.Equal(t, int64(11), groupID)
		assert.Equal(t, []int64{210}, itemParentIDs)
		_, _ = writer.Write([]byte("PK\x03\x04fake-zip"))
		return groups.GroupProgressZIPMeta{
			Filename:  filename,
			GroupName: "Classe 3B",
			Items:     []groups.GroupProgressZIPItemMeta{{ID: "210", Title: "Chapitre 1"}},
		}, nil
	}
	var uploaded bool
	uploadZIPImpl = func(_ context.Context, uploadURL, _, gotFilename string, sizeBytes int64) error {
		uploaded = true
		assert.Equal(t, "https://upload.example/put", uploadURL)
		assert.Equal(t, filename, gotFilename)
		assert.Positive(t, sizeBytes)
		return nil
	}

	err := Run(ctx, nil, tokentest.AlgoreaPlatformPublicKeyParsed(), RequestedPayload{
		ExportID:  "export-1",
		Token:     tokenStr,
		UploadURL: "https://upload.example/put",
	})
	require.NoError(t, err)
	assert.True(t, uploaded)

	events := mock.GetEvents()
	require.Len(t, events, 1)
	payload := events[0].Payload
	assert.Equal(t, event.TypeGroupResultsExportCompleted, events[0].Type)
	assert.Equal(t, statusSuccess, payload["status"])
	assert.Nil(t, payload["error"])
	assert.Equal(t, "export-1", payload["export_id"])
	assert.Equal(t, tokenStr, payload["token"])
	assert.Equal(t, "Classe 3B", payload["group_name"])
	assert.Equal(t, "21", payload["user_id"])
	assert.Equal(t, "11", payload["group_id"])
	assert.Equal(t, filename, payload["filename"])
	assert.Equal(t, int64(len("PK\x03\x04fake-zip")), payload["size_bytes"])
	items, ok := payload["items"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, items, 1)
	assert.Equal(t, "210", items[0]["id"])
	assert.Equal(t, "Chapitre 1", items[0]["title"])
}

func TestRun_FailurePaths(t *testing.T) {
	publicKey := tokentest.AlgoreaPlatformPublicKeyParsed()
	validToken := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID:  "21",
		GroupID: "11",
		ItemIDs: []string{"210"},
		Exp:     time.Now().Add(time.Hour).Unix(),
	})

	tests := []struct {
		name              string
		payload           RequestedPayload
		setup             func()
		expectedError     string
		expectedGroupName interface{}
	}{
		{
			name:          "missing fields",
			payload:       RequestedPayload{ExportID: "e"},
			expectedError: errorInternal,
		},
		{
			name: "invalid token",
			payload: RequestedPayload{
				ExportID: "e", Token: "bad", UploadURL: "https://x.example/put",
			},
			expectedError: errorTokenInvalid,
		},
		{
			name: "invalid user id in token",
			payload: RequestedPayload{
				ExportID: "e", UploadURL: "https://x.example/put",
				Token: signedGroupResultsToken(t, &payloads.GroupResultsToken{
					UserID: "abc", GroupID: "11", ItemIDs: []string{},
					Exp: time.Now().Add(time.Hour).Unix(),
				}),
			},
			expectedError: errorTokenInvalid,
		},
		{
			name: "invalid group id in token",
			payload: RequestedPayload{
				ExportID: "e", UploadURL: "https://x.example/put",
				Token: signedGroupResultsToken(t, &payloads.GroupResultsToken{
					UserID: "21", GroupID: "xyz", ItemIDs: []string{},
					Exp: time.Now().Add(time.Hour).Unix(),
				}),
			},
			expectedError: errorTokenInvalid,
		},
		{
			name: "invalid item id in token",
			payload: RequestedPayload{
				ExportID: "e", UploadURL: "https://x.example/put",
				Token: signedGroupResultsToken(t, &payloads.GroupResultsToken{
					UserID: "21", GroupID: "11", ItemIDs: []string{"nope"},
					Exp: time.Now().Add(time.Hour).Unix(),
				}),
			},
			expectedError: errorTokenInvalid,
		},
		{
			name: "unknown user",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return nil, gorm.ErrRecordNotFound
				}
			},
			expectedError: errorTokenInvalid,
		},
		{
			name: "db error loading user",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return nil, errors.New("connection refused")
				}
			},
			expectedError: errorInternal,
		},
		{
			name: "http upload_url rejected",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "http://insecure.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return &database.User{GroupID: 21}, nil
				}
			},
			expectedError: errorUploadFailed,
		},
		{
			name: "upload expired",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
				UploadExpiresAt: 1,
			},
			setup: func() {
				nowUnixMilliImpl = func() int64 { return 2 }
			},
			expectedError: errorUploadFailed,
		},
		{
			name: "too many users",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return &database.User{GroupID: 21}, nil
				}
				generateZIPImpl = stubZIPErrWithMeta(
					groups.ErrTooManyUsersInProgressZIP,
					groups.GroupProgressZIPMeta{GroupName: "Classe 3B"},
				)
			},
			expectedError:     errorTooManyUsers,
			expectedGroupName: "Classe 3B",
		},
		{
			name: "too many items",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return &database.User{GroupID: 21}, nil
				}
				generateZIPImpl = stubZIPErrWithMeta(
					groups.ErrTooManyItemsInProgressZIP,
					groups.GroupProgressZIPMeta{GroupName: "Classe 3B"},
				)
			},
			expectedError:     errorTooManyItems,
			expectedGroupName: "Classe 3B",
		},
		{
			name: "generation internal error",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return &database.User{GroupID: 21}, nil
				}
				generateZIPImpl = stubZIPErr(errors.New("boom"))
			},
			expectedError: errorInternal,
		},
		{
			name: "upload failed",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return &database.User{GroupID: 21}, nil
				}
				generateZIPImpl = stubZIPOK()
				uploadZIPImpl = func(context.Context, string, string, string, int64) error {
					return errors.New("upload broken")
				}
			},
			expectedError: errorUploadFailed,
		},
		{
			name: "upload deadline exceeded",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return &database.User{GroupID: 21}, nil
				}
				generateZIPImpl = stubZIPOK()
				uploadZIPImpl = func(context.Context, string, string, string, int64) error {
					return context.DeadlineExceeded
				}
			},
			expectedError: errorInternal,
		},
		{
			name: "temp dir unavailable",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
					return &database.User{GroupID: 21}, nil
				}
				tempDir = filepath.Join(os.TempDir(), "missing-dir-for-group-results-export-test")
			},
			expectedError: errorInternal,
		},
		{
			name: "context already canceled",
			payload: RequestedPayload{
				ExportID: "e", Token: validToken, UploadURL: "https://x.example/put",
			},
			setup: func() {
				// Checked via a pre-canceled ctx in the test body below.
			},
			expectedError: errorInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldLoad, oldGen, oldUpload, oldTemp, oldNow := loadUserByIDImpl, generateZIPImpl, uploadZIPImpl, tempDir, nowUnixMilliImpl
			defer func() {
				loadUserByIDImpl, generateZIPImpl, uploadZIPImpl, tempDir, nowUnixMilliImpl = oldLoad, oldGen, oldUpload, oldTemp, oldNow
			}()
			if tt.setup != nil {
				tt.setup()
			}

			ctx, mock := testContextWithMockDispatcher(t)
			if tt.name == "context already canceled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			err := Run(ctx, nil, publicKey, tt.payload)
			require.NoError(t, err)
			events := mock.GetEvents()
			require.Len(t, events, 1)
			assert.Equal(t, statusFailure, events[0].Payload["status"])
			assert.Equal(t, tt.expectedError, events[0].Payload["error"])
			if tt.expectedGroupName != nil {
				assert.Equal(t, tt.expectedGroupName, events[0].Payload["group_name"])
			}
		})
	}
}

func stubZIPErr(err error) zipGeneratorFunc {
	return func(io.Writer, *database.DataStore, *database.User, int64, []int64) (groups.GroupProgressZIPMeta, error) {
		return groups.GroupProgressZIPMeta{}, err
	}
}

func stubZIPErrWithMeta(err error, meta groups.GroupProgressZIPMeta) zipGeneratorFunc {
	return func(io.Writer, *database.DataStore, *database.User, int64, []int64) (groups.GroupProgressZIPMeta, error) {
		return meta, err
	}
}

func stubZIPOK() zipGeneratorFunc {
	return func(writer io.Writer, _ *database.DataStore, _ *database.User, _ int64, _ []int64) (groups.GroupProgressZIPMeta, error) {
		_, _ = writer.Write([]byte("data"))
		return groups.GroupProgressZIPMeta{Filename: "f.zip", GroupName: "g"}, nil
	}
}

func TestRun_RecoversPanicAndDispatches(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	tokenStr := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID:  "21",
		GroupID: "11",
		ItemIDs: []string{},
		Exp:     time.Now().Add(time.Hour).Unix(),
	})

	oldLoad := loadUserByIDImpl
	defer func() { loadUserByIDImpl = oldLoad }()
	loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
		panic("unexpected")
	}

	err := Run(ctx, nil, tokentest.AlgoreaPlatformPublicKeyParsed(), RequestedPayload{
		ExportID: "e", Token: tokenStr, UploadURL: "https://x.example/put",
	})
	require.NoError(t, err)
	events := mock.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, statusFailure, events[0].Payload["status"])
	assert.Equal(t, errorInternal, events[0].Payload["error"])
}

func TestRun_StatFailureAfterClose(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	tokenStr := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID: "21", GroupID: "11", ItemIDs: []string{}, Exp: time.Now().Add(time.Hour).Unix(),
	})
	oldLoad, oldGen := loadUserByIDImpl, generateZIPImpl
	defer func() { loadUserByIDImpl, generateZIPImpl = oldLoad, oldGen }()
	loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
		return &database.User{GroupID: 21}, nil
	}
	generateZIPImpl = func(
		writer io.Writer, _ *database.DataStore, _ *database.User, _ int64, _ []int64,
	) (groups.GroupProgressZIPMeta, error) {
		file, ok := writer.(*os.File)
		require.True(t, ok)
		name := file.Name()
		_, _ = file.WriteString("x")
		require.NoError(t, os.Rename(name, name+".moved"))
		t.Cleanup(func() { _ = os.Remove(name + ".moved") })
		return groups.GroupProgressZIPMeta{Filename: "f.zip"}, nil
	}
	err := Run(ctx, nil, tokentest.AlgoreaPlatformPublicKeyParsed(), RequestedPayload{
		ExportID: "e", Token: tokenStr, UploadURL: "https://x.example/put",
	})
	require.NoError(t, err)
	events := mock.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, errorInternal, events[0].Payload["error"])
}

func TestRun_CloseError(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	tokenStr := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID: "21", GroupID: "11", ItemIDs: []string{}, Exp: time.Now().Add(time.Hour).Unix(),
	})
	oldLoad, oldGen := loadUserByIDImpl, generateZIPImpl
	defer func() { loadUserByIDImpl, generateZIPImpl = oldLoad, oldGen }()
	loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
		return &database.User{GroupID: 21}, nil
	}
	generateZIPImpl = func(
		writer io.Writer, _ *database.DataStore, _ *database.User, _ int64, _ []int64,
	) (groups.GroupProgressZIPMeta, error) {
		file, ok := writer.(*os.File)
		require.True(t, ok)
		_, _ = file.WriteString("x")
		require.NoError(t, file.Close())
		return groups.GroupProgressZIPMeta{Filename: "f.zip"}, nil
	}
	err := Run(ctx, nil, tokentest.AlgoreaPlatformPublicKeyParsed(), RequestedPayload{
		ExportID: "e", Token: tokenStr, UploadURL: "https://x.example/put",
	})
	require.NoError(t, err)
	events := mock.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, errorInternal, events[0].Payload["error"])
}

func TestUploadZIP_HTTPClientError(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "upload-err-*.zip")
	require.NoError(t, err)
	_, _ = tmpFile.WriteString("x")
	_ = tmpFile.Close()
	info, _ := os.Stat(tmpFile.Name())

	oldClient := httpClient
	defer func() { httpClient = oldClient }()
	httpClient = &http.Client{
		Timeout: uploadHTTPClientTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		}),
	}
	err = uploadZIP(context.Background(), "https://example.com", tmpFile.Name(), "f.zip", info.Size())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "network down")
}

func TestUploadZIP_NewRequestError(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "upload-req-*.zip")
	require.NoError(t, err)
	_, _ = tmpFile.WriteString("x")
	_ = tmpFile.Close()
	info, _ := os.Stat(tmpFile.Name())

	// http.NewRequestWithContext rejects a nil context.
	err = uploadZIP(nil, "https://example.com/put", tmpFile.Name(), "f.zip", info.Size()) //nolint:staticcheck // intentional nil ctx
	require.Error(t, err)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestParseGroupResultsToken_InvalidPayloadMap(t *testing.T) {
	raw := token.Generate(map[string]interface{}{
		"user_id":  "1",
		"group_id": "1",
		"item_ids": "not-an-array",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}, tokentest.AlgoreaPlatformPrivateKeyParsed())
	_, err := parseGroupResultsToken(string(raw), tokentest.AlgoreaPlatformPublicKeyParsed())
	require.Error(t, err)
}

func TestContextWithExportDeadline_RespectsUploadExpiresAt(t *testing.T) {
	uploadExpiry := time.Now().Add(2 * time.Minute).UnixMilli()
	ctx, cancel := contextWithExportDeadline(context.Background(), uploadExpiry)
	defer cancel()
	deadline, ok := ctx.Deadline()
	require.True(t, ok)
	assert.WithinDuration(t, time.UnixMilli(uploadExpiry), deadline, time.Second)
}

func TestContextWithExportDeadline_DefaultCap(t *testing.T) {
	before := time.Now()
	ctx, cancel := contextWithExportDeadline(context.Background(), 0)
	defer cancel()
	deadline, ok := ctx.Deadline()
	require.True(t, ok)
	assert.WithinDuration(t, before.Add(exportMaxDuration), deadline, 2*time.Second)
}

func TestNewUploadHTTPClient_NoRedirects(t *testing.T) {
	client := newUploadHTTPClient()
	require.NotNil(t, client.CheckRedirect)
	assert.Equal(t, http.ErrUseLastResponse, client.CheckRedirect(nil, nil))
	assert.Equal(t, uploadHTTPClientTimeout, client.Timeout)
}

func TestRun_ContextCanceledAfterAuth(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	ctx, cancel := context.WithCancel(ctx)
	tokenStr := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID: "21", GroupID: "11", ItemIDs: []string{}, Exp: time.Now().Add(time.Hour).Unix(),
	})
	oldLoad := loadUserByIDImpl
	defer func() { loadUserByIDImpl = oldLoad }()
	loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
		cancel()
		return &database.User{GroupID: 21}, nil
	}
	err := Run(ctx, nil, tokentest.AlgoreaPlatformPublicKeyParsed(), RequestedPayload{
		ExportID: "e", Token: tokenStr, UploadURL: "https://x.example/put",
	})
	require.NoError(t, err)
	require.Len(t, mock.GetEvents(), 1)
	assert.Equal(t, errorInternal, mock.GetEvents()[0].Payload["error"])
}

func TestRun_ContextCanceledAfterZIP(t *testing.T) {
	ctx, mock := testContextWithMockDispatcher(t)
	ctx, cancel := context.WithCancel(ctx)
	tokenStr := signedGroupResultsToken(t, &payloads.GroupResultsToken{
		UserID: "21", GroupID: "11", ItemIDs: []string{}, Exp: time.Now().Add(time.Hour).Unix(),
	})
	oldLoad, oldGen := loadUserByIDImpl, generateZIPImpl
	defer func() { loadUserByIDImpl, generateZIPImpl = oldLoad, oldGen }()
	loadUserByIDImpl = func(_ *database.DataStore, _ int64) (*database.User, error) {
		return &database.User{GroupID: 21}, nil
	}
	generateZIPImpl = func(
		writer io.Writer, _ *database.DataStore, _ *database.User, _ int64, _ []int64,
	) (groups.GroupProgressZIPMeta, error) {
		_, _ = writer.Write([]byte("x"))
		cancel()
		return groups.GroupProgressZIPMeta{Filename: "f.zip", GroupName: "g"}, nil
	}
	err := Run(ctx, nil, tokentest.AlgoreaPlatformPublicKeyParsed(), RequestedPayload{
		ExportID: "e", Token: tokenStr, UploadURL: "https://x.example/put",
	})
	require.NoError(t, err)
	require.Len(t, mock.GetEvents(), 1)
	assert.Equal(t, errorInternal, mock.GetEvents()[0].Payload["error"])
}
