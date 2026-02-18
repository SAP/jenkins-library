package btp

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

func (btp *BTPUtils) Login(options LoginOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	parametersCheck := options.Url == "" ||
		options.Subdomain == "" ||
		options.User == "" ||
		options.Password == ""

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Url == "" {
			missingParams = append(missingParams, "Url")
		}
		if options.Subdomain == "" {
			missingParams = append(missingParams, "Subdomain")
		}
		if options.User == "" {
			missingParams = append(missingParams, "User")
		}
		if options.Password == "" {
			missingParams = append(missingParams, "Password")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return errors.Wrap(errors.New(errorMsg), "Failed to login to BTP")
	}

	log.Entry().Info("Logging in to BTP")

	builder := NewBTPCommandBuilder().
		WithAction("login").
		WithURL(options.Url).
		WithSubdomain(options.Subdomain).
		WithUser(options.User).
		WithPassword(options.Password)

	if options.IdentityProvider != "" {
		builder = builder.WithIdentityProvider(options.IdentityProvider)
	}

	btpLoginScript, _ := builder.Build()

	log.Entry().WithField("CLI URL:", options.Url).WithField("Subdomain", options.Subdomain).WithField("User", options.User).WithField("IdentityProvider", options.IdentityProvider)

	err := btp.Exec.Run(btpLoginScript)

	if err != nil {
		return errors.Wrap(err, "Failed to login to BTP")
	}
	log.Entry().Info("Logged in successfully to BTP.")
	return nil
}

func (btp *BTPUtils) Logout() error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	log.Entry().Info("Logout of BTP")

	builder := NewBTPCommandBuilder().
		WithAction("logout")

	btpLogoutScript, _ := builder.Build()

	err := btp.Exec.Run(btpLogoutScript)

	if err != nil {
		return errors.Wrap(err, "Failed to Logout of BTP")
	}
	log.Entry().Info("Logged out successfully")
	return nil
}

type LoginOptions struct {
	Url              string
	Subdomain        string
	User             string
	Password         string
	IdentityProvider string
}
