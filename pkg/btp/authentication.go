package btp

import (
	"errors"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

func NewBTPUtils(exec ExecRunner) *BTPUtils {
	b := new(BTPUtils)
	b.Exec = exec

	configOptions := ConfigOptions{
		Format: "json",
	}
	b.SetConfig(configOptions)
	return b
}

func (btp *BTPUtils) Login(options LoginOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	if options.Url == "" || options.Subdomain == "" || options.User == "" || options.Password == "" {
		return fmt.Errorf("Failed to login to BTP: %w", errors.New("Parameters missing. Please provide the CLI URL, Subdomain, Space, User and Password"))
	}

	log.Entry().Info("Logging in to BTP")

	builder := NewBTPCommandBuilder().
		WithAction("login").
		WithURL(options.Url).
		WithSubdomain(options.Subdomain).
		WithUser(options.User).
		WithPassword(options.Password)

	if options.Tenant != "" {
		builder = builder.WithTenant(options.Tenant)
	}

	btpLoginScript, _ := builder.Build()

	log.Entry().WithField("CLI URL:", options.Url).WithField("Subdomain", options.Subdomain).WithField("User", options.User).WithField("Tenant", options.Tenant)

	err := btp.Exec.Run(btpLoginScript)

	if err != nil {
		return fmt.Errorf("Failed to login to BTP: %w", err)
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
		return fmt.Errorf("Failed to Logout of BTP: %w", err)
	}
	log.Entry().Info("Logged out successfully")
	return nil
}

func (btp *BTPUtils) SetConfig(options ConfigOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	if options.Format == "" {
		return fmt.Errorf("Failed to set the configuration of the BTP CLI: %w", errors.New("Parameters missing. Please provide the Format"))
	}

	builder := NewBTPCommandBuilder().
		WithAction("set config").
		WithFormat(options.Format).
		WithVerbose()

	btpConfigScript, _ := builder.Build()

	log.Entry().WithField("Format:", options.Format)

	err := btp.Exec.Run(btpConfigScript)

	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("Failed to define the configuration of the BTP CLI: %w", err)
	}
	log.Entry().Info("Configuration successfully defined.")
	return nil
}

type BTPUtils struct {
	Exec ExecRunner
}

type LoginOptions struct {
	Url       string
	Subdomain string
	User      string
	Password  string
	Tenant    string
}

type ConfigOptions struct {
	Format  string
	Verbose bool
}
