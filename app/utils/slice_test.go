package utils

import "testing"

func TestContains(t *testing.T) {
	type args struct {
		slice []string
		str   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should return false if the string is not in the slice",
			args: args{
				slice: []string{"test", "string"},
				str:   "test1",
			},
			want: false,
		},
		{
			name: "should return false if the slice is empty",
			args: args{
				slice: []string{},
				str:   "test",
			},
			want: false,
		},
		{
			name: "should return true if the string is at the end of the slice",
			args: args{
				slice: []string{"test", "string"},
				str:   "string",
			},
			want: true,
		},
		{
			name: "should return true if the string is at the beginning of the slice",
			args: args{
				slice: []string{"test", "string"},
				str:   "test",
			},
			want: true,
		},
		{
			name: "should return true if the string is in the middle of the slice",
			args: args{
				slice: []string{"test", "string", "test1"},
				str:   "string",
			},
			want: true,
		},
		{
			name: "should return true if the string is the only element of the slice",
			args: args{
				slice: []string{"test"},
				str:   "test",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.slice, tt.args.str); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
