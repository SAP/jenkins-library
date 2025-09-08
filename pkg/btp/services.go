package btp

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/btputils"
	"github.com/SAP/jenkins-library/pkg/log"
)

func (btp *BTPUtils) CreateServiceBinding(options CreateServiceBindingOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	if options.Subdomain == "" ||
		options.BindingName == "" ||
		options.Parameters == "" ||
		options.Timeout == 0 ||
		options.PollInterval == 0 {
		return "", fmt.Errorf("Failed to login to BTP: %w", errors.New("Parameters missing. Please provide the Subdomain, BindingName, Parameters, Timeout and PollInterval"))
	}

	loginOptions := LoginOptions{
		Url:       options.Url,
		Subdomain: options.Subdomain,
		User:      options.User,
		Password:  options.Password,
		Tenant:    options.Tenant,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return "", fmt.Errorf("Login to BTP failed: %w", err)
	}
	var serviceBindingBytes bytes.Buffer
	btp.Exec.Stdout(&serviceBindingBytes)

	// we are logged in --> create service binding
	log.Entry().WithField("subaccount", options.Subaccount).
		WithField("name", options.BindingName).
		WithField("parameter", options.Parameters)

	btpCreateBindingScript, _ := NewBTPCommandBuilder().
		WithAction("create").
		WithTarget("services/binding").
		WithName(options.BindingName).
		WithServiceInstanceName(options.ServiceInstance).
		WithSubAccount(options.Subaccount).
		WithParameters(options.Parameters).Build()

	err = btp.Exec.RunSync(btputils.RunSyncOptions{
		CmdScript:      btpCreateBindingScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		CheckFunc: func() bool {
			return CheckServiceBindingCreated(btp, GetServiceBindingOptions{
				Url:         options.Url,
				Subdomain:   options.Subdomain,
				User:        options.User,
				Password:    options.Password,
				Tenant:      options.Tenant,
				Subaccount:  options.Subaccount,
				BindingName: options.BindingName,
			})
		},
	})

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("Creation of service binding failed: %w", err)
	}

	// parse and return service binding
	serviceBindingJSON, err := GetJSON(serviceBindingBytes.String())

	if err != nil {
		return "", err
	}

	err = btp.Logout()
	if err != nil {
		return serviceBindingJSON, fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return serviceBindingJSON, nil
}

func (btp *BTPUtils) GetServiceBinding(options GetServiceBindingOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	if options.Subaccount == "" ||
		options.BindingName == "" {
		return "", fmt.Errorf("Failed to login to BTP: %w", errors.New("Parameters missing. Please provide the Subaccount, and BindingName"))
	}

	loginOptions := LoginOptions{
		Url:       options.Url,
		Subdomain: options.Subdomain,
		User:      options.User,
		Password:  options.Password,
		Tenant:    options.Tenant,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return "", fmt.Errorf("Login to BTP failed: %w", err)
	}
	var serviceBindingBytes bytes.Buffer
	btp.Exec.Stdout(&serviceBindingBytes)

	// we are logged in --> read service binding
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	builder := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithName(options.BindingName)

	btpGetBindingScript, _ := builder.Build()
	err = btp.Exec.Run(btpGetBindingScript)

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("Retrieve service binding failed: %w", err)
	}

	// parse and return service binding
	serviceBindingJSON, err := GetJSON(serviceBindingBytes.String())

	if err != nil {
		return "", err
	}

	err = btp.Logout()
	if err != nil {
		return serviceBindingJSON, fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return serviceBindingJSON, nil
}

func (btp *BTPUtils) DeleteServiceBinding(options DeleteServiceBindingOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	if options.Subaccount == "" ||
		options.BindingName == "" ||
		options.Timeout == 0 ||
		options.PollInterval == 0 {
		return fmt.Errorf("Failed to login to BTP: %w", errors.New("Parameters missing. Please provide the Subdomain, BindingName, Timeout and PollInterval"))
	}

	loginOptions := LoginOptions{
		Url:       options.Url,
		Subdomain: options.Subdomain,
		User:      options.User,
		Password:  options.Password,
		Tenant:    options.Tenant,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return fmt.Errorf("Login to BTP failed: %w", err)
	}
	var serviceBindingBytes bytes.Buffer
	btp.Exec.Stdout(&serviceBindingBytes)

	// we are logged in --> delete service binding
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	btpDeleteBindingScript, _ := NewBTPCommandBuilder().
		WithAction("delete").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithName(options.BindingName).WithConfirm().Build()

	err = btp.Exec.RunSync(btputils.RunSyncOptions{
		CmdScript:      btpDeleteBindingScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		CheckFunc: func() bool {
			return CheckServiceBindingDeleted(btp, GetServiceBindingOptions{
				Url:         options.Url,
				Subdomain:   options.Subdomain,
				User:        options.User,
				Password:    options.Password,
				Tenant:      options.Tenant,
				Subaccount:  options.Subaccount,
				BindingName: options.BindingName,
			})
		},
	})

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("Error occurred while deleting the Service-Binding: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		return fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return nil
}

func (btp *BTPUtils) CreateServiceInstance(options CreateServiceInstanceOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	if options.Subaccount == "" ||
		options.PlanName == "" ||
		options.OfferingName == "" ||
		options.InstanceName == "" ||
		options.Parameters == "" ||
		options.Timeout == 0 ||
		options.PollInterval == 0 {
		return "", fmt.Errorf("Failed to login to BTP: %w", errors.New("Parameters missing. Please provide the Subaccount, PlanName, OfferingName, InstanceName, Parameters, Timeout and PollInterval"))
	}

	loginOptions := LoginOptions{
		Url:       options.Url,
		Subdomain: options.Subdomain,
		User:      options.User,
		Password:  options.Password,
		Tenant:    options.Tenant,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return "", fmt.Errorf("Login to BTP failed: %w", err)
	}
	var serviceInstanceBytes bytes.Buffer
	btp.Exec.Stdout(&serviceInstanceBytes)

	// we are logged in --> create service instance
	log.Entry().WithField("subaccount", options.Subaccount).
		WithField("planName", options.PlanName).
		WithField("offeringName", options.OfferingName).
		WithField("name", options.InstanceName).
		WithField("parameters", options.Parameters)

	btpCreateInstanceScript, _ := NewBTPCommandBuilder().
		WithAction("create").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount).
		WithParameters(options.Parameters).
		WithPlanName(options.PlanName).WithOfferingName(options.OfferingName).
		Build()

	err = btp.Exec.RunSync(btputils.RunSyncOptions{
		CmdScript:      btpCreateInstanceScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		CheckFunc: func() bool {
			return CheckServiceInstanceCreated(btp, GetServiceInstanceOptions{
				Url:          options.Url,
				Subdomain:    options.Subdomain,
				User:         options.User,
				Password:     options.Password,
				Tenant:       options.Tenant,
				Subaccount:   options.Subaccount,
				InstanceName: options.InstanceName,
			})
		},
	})

	if err != nil {
		// error while getting service instance
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("Creation of service instance failed: %w", err)
	}

	// parse and return service instance
	serviceInstanceJSON, err := GetJSON(serviceInstanceBytes.String())

	if err != nil {
		return "", err
	}

	err = btp.Logout()
	if err != nil {
		return serviceInstanceJSON, fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return serviceInstanceJSON, nil
}

func (btp *BTPUtils) GetServiceInstance(options GetServiceInstanceOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	if options.Subaccount == "" ||
		options.InstanceName == "" {
		return "", fmt.Errorf("Failed to login to BTP: %w", errors.New("Parameters missing. Please provide the Subaccount, and InstanceName"))
	}

	loginOptions := LoginOptions{
		Url:       options.Url,
		Subdomain: options.Subdomain,
		User:      options.User,
		Password:  options.Password,
		Tenant:    options.Tenant,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return "", fmt.Errorf("Login to BTP failed: %w", err)
	}
	var serviceInstanceBytes bytes.Buffer
	btp.Exec.Stdout(&serviceInstanceBytes)

	// we are logged in --> read service instance
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.InstanceName)

	builder := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount)

	btpGetServiceScript, _ := builder.Build()
	err = btp.Exec.Run(btpGetServiceScript)

	if err != nil {
		// error while getting service instance
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("Retrieve service instance failed: %w", err)
	}

	// parse and return service instance
	serviceInstanceJSON, err := GetJSON(serviceInstanceBytes.String())

	if err != nil {
		return "", err
	}

	err = btp.Logout()
	if err != nil {
		return serviceInstanceJSON, fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return serviceInstanceJSON, nil
}

func (btp *BTPUtils) DeleteServiceInstance(options DeleteServiceInstanceOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	if options.Subaccount == "" ||
		options.InstanceName == "" {
		return fmt.Errorf("Failed to login to BTP: %w", errors.New("Parameters missing. Please provide the Subaccount, and InstanceName"))
	}

	loginOptions := LoginOptions{
		Url:       options.Url,
		Subdomain: options.Subdomain,
		User:      options.User,
		Password:  options.Password,
		Tenant:    options.Tenant,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return fmt.Errorf("Login to BTP failed: %w", err)
	}
	var serviceInstanceBytes bytes.Buffer
	btp.Exec.Stdout(&serviceInstanceBytes)

	// we are logged in --> delete service instance
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.InstanceName)

	btpGetServiceScript, _ := NewBTPCommandBuilder().
		WithAction("delete").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount).
		WithConfirm().Build()

	err = btp.Exec.RunSync(btputils.RunSyncOptions{
		CmdScript:      btpGetServiceScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		CheckFunc: func() bool {
			return CheckServiceInstanceDeleted(btp, GetServiceInstanceOptions{
				Url:          options.Url,
				Subdomain:    options.Subdomain,
				User:         options.User,
				Password:     options.Password,
				Tenant:       options.Tenant,
				Subaccount:   options.Subaccount,
				InstanceName: options.InstanceName,
			})
		},
	})

	if err != nil {
		// error while deleting service instance
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("Checking if Service-Instance was deleted failed: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		return fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return nil
}

func GetJSON(value string) (string, error) {
	var serviceBindingJSON string

	if len(value) > 0 {
		// parse and return service key
		var lines []string = strings.Split(value, "\n")
		serviceBindingJSON = strings.Join(lines, "")

		return serviceBindingJSON, nil
	}

	return "", errors.New("The returned value is empty")
}

type CreateServiceBindingOptions struct {
	Url             string
	Subdomain       string
	Subaccount      string
	ServiceInstance string
	BindingName     string
	Parameters      string
	User            string
	Password        string
	Tenant          string
	Timeout         int
	PollInterval    int
}

type GetServiceBindingOptions struct {
	Url         string
	Subdomain   string
	Subaccount  string
	BindingName string
	User        string
	Password    string
	Tenant      string
}

type DeleteServiceBindingOptions struct {
	Url          string
	Subdomain    string
	Subaccount   string
	BindingName  string
	User         string
	Password     string
	Tenant       string
	Timeout      int
	PollInterval int
}

type CreateServiceInstanceOptions struct {
	Url          string
	Subdomain    string
	User         string
	Password     string
	Tenant       string
	Subaccount   string
	PlanName     string
	OfferingName string
	InstanceName string
	Parameters   string
	Timeout      int
	PollInterval int
}

type GetServiceInstanceOptions struct {
	Url          string
	Subdomain    string
	User         string
	Password     string
	Tenant       string
	Subaccount   string
	InstanceName string
}

type DeleteServiceInstanceOptions struct {
	Url          string
	Subdomain    string
	User         string
	Password     string
	Tenant       string
	Subaccount   string
	InstanceName string
	Timeout      int
	PollInterval int
}
