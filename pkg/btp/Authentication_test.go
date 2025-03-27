package btp

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func loginMockCleanup(m *mock.BtpExecuterMock) {
	m.ShouldFailOnCommand = map[string]error{}
	m.StdoutReturn = map[string]string{}
	m.Calls = []mock.BtpExecCall{}
}

func TestBTPLoginCheck(t *testing.T) {

	m := &mock.BtpExecuterMock{
		StdoutReturn: map[string]string{
			"btp login": "Login successful",
		},
	}

	m.Stdout(new(bytes.Buffer))

	t.Run("CF Login check: logged in", func(t *testing.T) {

		defer loginMockCleanup(m)

		btp := BTPUtils{Exec: m, loggedIn: true}

		loggedIn, err := btp.LoginCheck()
		if assert.NoError(t, err) {
			assert.True(t, loggedIn)
		}
	})

	t.Run("CF Login check: not logged in", func(t *testing.T) {

		defer loginMockCleanup(m)

		cf := BTPUtils{Exec: m, loggedIn: false}

		loggedIn, err := cf.LoginCheck()
		if assert.NoError(t, err) {
			assert.False(t, loggedIn)
		}
	})
}

func TestBTPLogin(t *testing.T) {

	m := &mock.BtpExecuterMock{
		StdoutReturn: map[string]string{},
	}

	m.Stdout(new(bytes.Buffer))

	t.Run("BTP Login: missing parameter", func(t *testing.T) {

		defer loginMockCleanup(m)

		cfconfig := LoginOptions{}
		btp := BTPUtils{Exec: m}
		err := btp.Login(cfconfig)
		assert.EqualError(t, err, "Failed to login to BTP: Parameters missing. Please provide the CLI URL, Subdomain, Space, User and Password")
	})
	t.Run("BTP Login: failure", func(t *testing.T) {

		defer loginMockCleanup(m)

		m.ShouldFailOnCommand = map[string]error{"btp login .*": fmt.Errorf("wrong password or account does not exist")}

		btpConfig := LoginOptions{
			Url:       "https://api.endpoint.com",
			Subdomain: "xxx",
			User:      "john@example.com",
			Password:  "xxx",
		}

		btp := BTPUtils{Exec: m}
		err := btp.Login(btpConfig)
		if assert.EqualError(t, err, "Failed to login to BTP: wrong password or account does not exist") {
			assert.False(t, btp.loggedIn)
		}
	})

	t.Run("BTP Login: success", func(t *testing.T) {

		defer loginMockCleanup(m)

		m.StdoutReturn = map[string]string{"btp login .*": "Authentication successful"}

		btpConfig := LoginOptions{
			Url:       "https://api.endpoint.com",
			Subdomain: "xxx",
			User:      "john@example.com",
			Password:  "xxx",
		}

		btp := BTPUtils{Exec: m}
		err := btp.Login(btpConfig)

		if assert.NoError(t, err) {
			assert.True(t, btp.loggedIn)
		}
	})
}

func TestBTPLogout(t *testing.T) {

	m := &mock.BtpExecuterMock{}

	t.Run("BTP Logout", func(t *testing.T) {
		btp := BTPUtils{Exec: m}
		err := btp.Logout()

		if assert.NoError(t, err) {
			assert.False(t, btp.loggedIn)
		}
	})
}
