package auth

import (
	"net/http"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

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
				Name: "access_token", Value: "1!sometoken!example.org!/api/", Path: "/api/", Domain: "example.org",
				Expires: time.Now().Add(12345 * time.Second), MaxAge: 12345, Secure: false, HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			},
		},
		{
			name:   "normal 2",
			fields: fields{Secure: true, SameSite: false, Domain: "example.com", Path: "/"},
			args:   args{token: "anothertoken", secondsUntilExpiry: 7200},
			want: &http.Cookie{
				Name: "access_token", Value: "2!anothertoken!example.com!/", Path: "/", Domain: "example.com",
				Expires: time.Now().Add(7200 * time.Second), MaxAge: 7200, Secure: true, HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			},
		},
		{
			name:   "expired",
			fields: fields{Secure: true, SameSite: false, Domain: "example.com", Path: "/"},
			args:   args{token: "anothertoken", secondsUntilExpiry: -1},
			want: &http.Cookie{
				Name: "access_token", Value: "2!anothertoken!example.com!/", Path: "/", Domain: "example.com",
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

func Test_unmarshalSessionCookieValue_InvalidValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "three parts", value: "!!"},
		{name: "the first part is empty", value: "!!!"},
		{name: "five parts", value: "!!!!"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			gotValue, gotAttributes := unmarshalSessionCookieValue(test.value)
			assert.Empty(t, gotValue)
			assert.False(t, gotAttributes.UseCookie)
		})
	}
}
