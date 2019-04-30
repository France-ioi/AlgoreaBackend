package currentuser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_escapeLikeString(t *testing.T) {
	type args struct {
		s               string
		escapeCharacter byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "all characters",
			args: args{s: "|some _string_ 100%|", escapeCharacter: '|'},
			want: "||some |_string|_ 100|%||",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := escapeLikeString(tt.args.s, tt.args.escapeCharacter)
			assert.Equal(t, tt.want, got)
		})
	}
}
