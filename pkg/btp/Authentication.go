package btp

import (
	"errors"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

type BTPUtils struct {
	Exec     ExecRunner
	loggedIn bool
}

type LoginOptions struct {
	Url       string
	Subdomain string
	User      string
	Password  string
	Tenant    string
}

func (btp *BTPUtils) LoginCheck() (bool, error) {
	return btp.loggedIn, nil
}

func (btp *BTPUtils) Login(options LoginOptions) error {
	var err error

	_r := btp.Exec

	if _r == nil {
		_r = &Executer{}
	}

	if options.Url == "" || options.Subdomain == "" || options.User == "" || options.Password == "" {
		return fmt.Errorf("Failed to login to BTP: %w", errors.New("Parameters missing. Please provide the CLI URL, Subdomain, Space, User and Password"))
	}

	var loggedIn bool

	loggedIn, err = btp.LoginCheck()

	if loggedIn == true {
		return err
	}

	if err == nil {
		log.Entry().Info("Logging in to BTP")

		builder := NewBTPCommandBuilder().
			WithAction("login").
			WithURL(options.Url).
			WithSubdomain(options.Subdomain).
			WithUser(options.User).
			WithPassword(options.Password)

		btpLoginScript, _ := builder.Build()

		log.Entry().WithField("CLI URL:", options.Url).WithField("Subdomain", options.Subdomain).WithField("User", options.User).WithField("Password", options.Password).WithField("Tenant", options.Tenant)

		err = _r.Run(btpLoginScript)
	}

	if err != nil {
		return fmt.Errorf("Failed to login to BTP: %w", err)
	}
	log.Entry().Info("Logged in successfully to BTP..")
	btp.loggedIn = true
	return nil
}

func (btp *BTPUtils) Logout() error {

	_r := btp.Exec

	if _r == nil {
		_r = &Executer{}
	}

	log.Entry().Info("Logout of BTP")

	builder := NewBTPCommandBuilder().
		WithAction("logout")

	btpLogoutScript, _ := builder.Build()

	err := _r.Run(btpLogoutScript)

	if err != nil {
		return fmt.Errorf("Failed to Logout of BTP: %w", err)
	}
	log.Entry().Info("Logged out successfully")
	btp.loggedIn = false
	return nil
}
