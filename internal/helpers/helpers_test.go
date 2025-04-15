package helpers

import (
	"testing"
)

func TestIsURLValid(t *testing.T) {
	type args struct {
		rawURL string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Ok https",
			args: args{rawURL: "https://example.com"},
			want: true,
		},
		{
			name: "Ok http",
			args: args{rawURL: "http://example.com"},
			want: true,
		},
		{
			name: "Empty  string",
			args: args{rawURL: ""},
			want: false,
		},
		{
			name: "not url",
			args: args{rawURL: "abasd123"},
			want: false,
		},
		{
			name: "missing schema",
			args: args{rawURL: "foo://example.com"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsURLValid(tt.args.rawURL); got != tt.want {
				t.Errorf("IsURLValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
