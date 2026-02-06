package btp

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loginMockCleanup(m *BtpExecutorMock) {
	m.ShouldFailOnCommand = map[string]error{}
	m.StdoutReturn = map[string]string{}
	m.Calls = []BtpExecCall{}
}

func TestBTPLogin(t *testing.T) {

	m := &BtpExecutorMock{
		StdoutReturn: map[string]string{},
	}

	m.Stdout(new(bytes.Buffer))
	m.Stderr(new(bytes.Buffer))

	t.Run("BTP Login: missing parameter", func(t *testing.T) {
		defer loginMockCleanup(m)

		btpConfig := LoginOptions{}
		btp := BTPUtils{Exec: m}
		err := btp.Login(btpConfig)
		assert.EqualError(t, err, "Failed to login to BTP: Parameters missing. Please provide: Url, Subdomain, User, Password")
	})

	t.Run("BTP Login: failure", func(t *testing.T) {
		defer loginMockCleanup(m)

		m.ShouldFailOnCommand = map[string]error{"btp .* login .+": errors.New("wrong password or account does not exist")}

		btpConfig := LoginOptions{
			Url:       "https://api.endpoint.com",
			Subdomain: "xxx",
			User:      "john@example.com",
			Password:  "xxx",
		}

		btp := BTPUtils{Exec: m}
		err := btp.Login(btpConfig)
		assert.EqualError(t, err, "Failed to login to BTP: wrong password or account does not exist")
	})

	t.Run("BTP Login: success", func(t *testing.T) {

		defer loginMockCleanup(m)

		m.StdoutReturn = map[string]string{"btp .* login .+": "Authentication successful"}

		btpConfig := LoginOptions{
			Url:       "https://api.endpoint.com",
			Subdomain: "xxx",
			User:      "john@example.com",
			Password:  "xxx",
		}

		btp := BTPUtils{Exec: m}
		err := btp.Login(btpConfig)

		assert.NoError(t, err)
	})
}

func TestBTPLogout(t *testing.T) {

	m := &BtpExecutorMock{}

	t.Run("BTP Logout", func(t *testing.T) {
		btp := BTPUtils{Exec: m}
		err := btp.Logout()

		assert.NoError(t, err)
	})
}
