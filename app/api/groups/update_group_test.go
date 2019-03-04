package groups

import (
	"net/http"
	"strings"
	"testing"
)

func Test_validateUpdateGroupInput(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{"sType=Class", `{"type":"Class"}`, false},
		{"sType=Team", `{"type":"Team"}`, false},
		{"sType=Club", `{"type":"Club"}`, false},
		{"sType=Friends", `{"type":"Friends"}`, false},
		{"sType=Other", `{"type":"Other"}`, false},

		{"sType=Root", `{"type":"Root"}`, true},
		{"sType=UserSelf", `{"type":"UserSelf"}`, true},
		{"sType=UserAdmin", `{"type":"UserAdmin"}`, true},
		{"sType=RootSelf", `{"type":"RootSelf"}`, true},
		{"sType=RootAdmin", `{"type":"RootAdmin"}`, true},
		{"sType=unknown", `{"type":"unknown"}`, true},
		{"sType=", `{"type":""}`, true},

		{"sPasswordTimer=99:59:59", `{"password_timer":"99:59:59"}`, false},
		{"sPasswordTimer=00:00:00", `{"password_timer":"00:00:00"}`, false},

		{"sPasswordTimer=99:60:59", `{"password_timer":"99:60:59"}`, true},
		{"sPasswordTimer=99:59:60", `{"password_timer":"99:59:60"}`, true},
		{"sPasswordTimer=59:59", `{"password_timer":"59:59"}`, true},
		{"sPasswordTimer=59", `{"password_timer":"59"}`, true},
		{"sPasswordTimer=59", `{"password_timer":"invalid"}`, true},
		{"sPasswordTimer=", `{"password_timer":""}`, true},

		{"sRedirectPath=9", `{"redirect_path":"9"}`, false},
		{"sRedirectPath=1234567890", `{"redirect_path":"1234567890"}`, false},
		{"sRedirectPath=1234567890/0", `{"redirect_path":"1234567890/0"}`, false},
		{"sRedirectPath=0/1234567890", `{"redirect_path":"0/1234567890"}`, false},
		{"sRedirectPath=1234567890/1234567890", `{"redirect_path":"1234567890/1234567890"}`, false},
		// empty strings are allowed (there are some in the DB)
		{"sRedirectPath=", `{"redirect_path":""}`, false},

		{"sRedirectPath=invalid", `{"redirect_path":"invalid"}`, true},
		{"sRedirectPath=1A", `{"redirect_path":"1A"}`, true},
		{"sRedirectPath=1A/2B", `{"redirect_path":"1A/2B"}`, true},
		{"sRedirectPath=1234567890/1234567890/1", `{"redirect_path":"1234567890/1234567890/1"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("POST", "/", strings.NewReader(tt.json))
			_, err := validateUpdateGroupInput(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUpdateGroupInput() error = %#v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
