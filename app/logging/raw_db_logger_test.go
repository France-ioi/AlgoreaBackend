package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_prepareRawDBLoggerValuesMap(t *testing.T) {
	tests := []struct {
		name    string
		keyvals []interface{}
		want    map[string]interface{}
	}{
		{
			name: "simple",
			keyvals: []interface{}{
				"query", "SELECT * FROM users WHERE users.ID=? and users.sName=?",
				"args", "{[int64 1], [string \"Joe\"]}",
			},
			want: map[string]interface{}{
				"query": "SELECT * FROM users WHERE users.ID='1' and users.sName='Joe'",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareRawDBLoggerValuesMap(tt.keyvals)
			assert.Equal(t, tt.want, got)
		})
	}
}
