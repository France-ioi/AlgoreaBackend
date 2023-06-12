package utils

import "testing"

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "should capitalize the first letter",
			str:  "test string",
			want: "Test string",
		},
		{
			name: "should keep as it is if already capitalized",
			str:  "Test string",
			want: "Test string",
		},
		{
			name: "should handle empty strings",
			str:  "",
			want: "",
		},
		{
			name: "should keep non-letter character as they are",
			str:  ".test",
			want: ".test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Capitalize(tt.str); got != tt.want {
				t.Errorf("Capitalize() = %v, want %v", got, tt.want)
			}
		})
	}
}
