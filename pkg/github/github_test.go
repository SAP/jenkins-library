package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientBuilder(t *testing.T) {
	type args struct {
		token   string
		baseURL string
	}
	tests := []struct {
		name string
		args args
		want *ClientBuilder
	}{
		{
			name: "token and baseURL",
			args: args{
				token:   "test_token",
				baseURL: "https://test.com/",
			},
			want: &ClientBuilder{
				token:   "test_token",
				baseURL: "https://test.com/",
			},
		},
		{
			name: "baseURL without prefix",
			args: args{
				token:   "test_token",
				baseURL: "https://test.com",
			},
			want: &ClientBuilder{
				token:   "test_token",
				baseURL: "https://test.com/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewClientBuilder(tt.args.token, tt.args.baseURL), "NewClientBuilder(%v, %v)", tt.args.token, tt.args.baseURL)
		})
	}
}
