package database

import (
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSessionStore_InsertNewOAuth(t *testing.T) {
	tests := []struct {
		name             string
		issuer           string
		cookieAttributes SessionCookieAttributes
		dbError          error
	}{
		{name: "success without cookie"},
		{
			name:   "success with cookie",
			issuer: "login-module",
			cookieAttributes: SessionCookieAttributes{
				UseCookie: true,
				Secure:    false,
				SameSite:  false,
				Domain:    "somedomain.com",
				Path:      "/path/",
			},
		},
		{
			name:   "success with cookie 2",
			issuer: "backend",
			cookieAttributes: SessionCookieAttributes{
				UseCookie: true,
				Secure:    true,
				SameSite:  false,
				Domain:    "somedomain1.com",
				Path:      "/path1/",
			},
		},
		{
			name:   "success with cookie 3",
			issuer: "backend",
			cookieAttributes: SessionCookieAttributes{
				UseCookie: true,
				Secure:    false,
				SameSite:  true,
				Domain:    "somedomain1.com",
				Path:      "/path1/",
			},
		},
		{name: "error", dbError: errors.New("some error")},
	}
	for _, test := range tests {
		test := test

		db, mock := NewDBMock()
		defer func() { _ = db.Close() }()

		userID := int64(123456)
		token := "accesstoken"
		secondsUntilExpiry := int32(1234)
		expectedExec := mock.ExpectExec("^"+regexp.QuoteMeta(
			"INSERT INTO `sessions` "+
				"(`access_token`, `cookie_domain`, `cookie_path`, `cookie_same_site`, `cookie_secure`, `expires_at`, "+
				"`issued_at`, `issuer`, `use_cookie`, `user_id`) VALUES "+
				"(?, ?, ?, ?, ?, NOW() + INTERVAL ? SECOND, NOW(), ?, ?, ?)")+"$").
			WithArgs(token, stringOrNil(test.cookieAttributes.Domain),
				stringOrNil(test.cookieAttributes.Path),
				test.cookieAttributes.SameSite, test.cookieAttributes.Secure, secondsUntilExpiry,
				test.issuer, test.cookieAttributes.UseCookie, userID)

		if test.dbError != nil {
			expectedExec.WillReturnError(test.dbError)
		} else {
			expectedExec.WillReturnResult(sqlmock.NewResult(1, 1))
		}

		err := NewDataStore(db).Sessions().InsertNewOAuth(userID, token, secondsUntilExpiry, test.issuer, &test.cookieAttributes)
		assert.Equal(t, test.dbError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestSessionCookieAttributes_SessionCookie(t *testing.T) {
	now := time.Now()
	patch := monkey.Patch(time.Now, func() time.Time { return now })
	defer patch.Unpatch()

	type fields struct {
		UseCookie bool
		Secure    bool
		SameSite  bool
		Domain    string
		Path      string
	}
	type args struct {
		token              string
		secondsUntilExpiry int32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *http.Cookie
	}{
		{
			name:   "normal 1",
			fields: fields{Secure: false, SameSite: true, Domain: "example.org", Path: "/api/"},
			args:   args{token: "sometoken", secondsUntilExpiry: 12345},
			want: &http.Cookie{
				Name: "access_token", Value: "sometoken", Path: "/api/", Domain: "example.org",
				Expires: time.Now().Add(12345 * time.Second), MaxAge: 12345, Secure: false, HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			},
		},
		{
			name:   "normal 2",
			fields: fields{Secure: true, SameSite: false, Domain: "example.com", Path: "/"},
			args:   args{token: "anothertoken", secondsUntilExpiry: 7200},
			want: &http.Cookie{
				Name: "access_token", Value: "anothertoken", Path: "/", Domain: "example.com",
				Expires: time.Now().Add(7200 * time.Second), MaxAge: 7200, Secure: true, HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			},
		},
		{
			name:   "expired",
			fields: fields{Secure: true, SameSite: false, Domain: "example.com", Path: "/"},
			args:   args{token: "anothertoken", secondsUntilExpiry: -1},
			want: &http.Cookie{
				Name: "access_token", Value: "anothertoken", Path: "/", Domain: "example.com",
				Expires: time.Now().Add(-1 * time.Second), MaxAge: -1, Secure: true, HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			attributes := &SessionCookieAttributes{
				UseCookie: tt.fields.UseCookie,
				Secure:    tt.fields.Secure,
				SameSite:  tt.fields.SameSite,
				Domain:    tt.fields.Domain,
				Path:      tt.fields.Path,
			}
			got := attributes.SessionCookie(tt.args.token, tt.args.secondsUntilExpiry)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_stringOrNil(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want interface{}
	}{
		{name: "not empty", arg: "s", want: "s"},
		{name: "empty", arg: "", want: nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := stringOrNil(tt.arg)
			assert.Equal(t, tt.want, got)
		})
	}
}
