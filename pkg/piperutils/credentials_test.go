package piperutils

import (
	"testing"
)

func TestEncodeUsernamePassword(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{args: args{username: "anything", password: "something"}, want: "YW55dGhpbmc6c29tZXRoaW5n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeUsernamePassword(tt.args.username, tt.args.password); got != tt.want {
				t.Errorf("EncodeUsernamePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeToken(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{args: args{token: "anything"}, want: "YW55dGhpbmc="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeString(tt.args.token); got != tt.want {
				t.Errorf("EncodeToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
