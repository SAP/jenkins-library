package btp

import (
	"bytes"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

func (btp *BTPUtils) CreateServiceBinding(options CreateServiceBindingOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	loginOptions := LoginOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return "", errors.Wrap(err, "Login to BTP failed")
	}

	// we are logged in --> create service binding
	parametersCheck := options.Subaccount == "" ||
		options.BindingName == "" ||
		options.Timeout == 0 ||
		options.PollInterval == 0

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.BindingName == "" {
			missingParams = append(missingParams, "BindingName")
		}
		if options.Timeout == 0 {
			missingParams = append(missingParams, "Timeout")
		}
		if options.PollInterval == 0 {
			missingParams = append(missingParams, "PollInterval")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return "", errors.Wrap(errors.New(errorMsg), "Failed to create service binding")
	}

	log.Entry().WithField("subaccount", options.Subaccount).
		WithField("name", options.BindingName).
		WithField("parameter", options.Parameters)

	builder := NewBTPCommandBuilder().
		WithAction("create").
		WithTarget("services/binding").
		WithName(options.BindingName).
		WithServiceInstanceName(options.ServiceInstance).
		WithSubAccount(options.Subaccount)

	if options.Parameters != "" {
		builder = builder.WithParameters(options.Parameters)
	}

	btpCreateBindingScript, _ := builder.Build()

	err = btp.Exec.RunSync(RunSyncOptions{
		CmdScript:      btpCreateBindingScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		LoginFunc: func() error {
			return btp.Login(loginOptions)
		},
		CheckFunc: func() CheckResponse {
			return IsServiceBindingCreated(btp, GetServiceBindingOptions{
				Url:              options.Url,
				Subdomain:        options.Subdomain,
				User:             options.User,
				Password:         options.Password,
				IdentityProvider: options.IdentityProvider,
				Subaccount:       options.Subaccount,
				BindingName:      options.BindingName,
			})
		},
		IgnoreErrorOnFirstCall: true,
	})

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", errors.Wrapf(err, "Creation of service binding failed for binding : %s", options.BindingName)
	}

	// parse and return service binding
	serviceBindingJSON, err := GetJSON(btp.Exec.GetStdoutValue())

	if err != nil {
		return "", errors.Wrap(err, "Parsing service binding JSON failed")
	}

	err = btp.Logout()
	if err != nil {
		return serviceBindingJSON, errors.Wrap(err, "Logout of BTP failed")
	}

	return serviceBindingJSON, nil
}

func (btp *BTPUtils) GetServiceBinding(options GetServiceBindingOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	loginOptions := LoginOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return "", errors.Wrap(err, "Login to BTP failed")
	}
	var serviceBindingBytes bytes.Buffer
	btp.Exec.Stdout(&serviceBindingBytes)

	// we are logged in --> read service binding
	res, err := btp.RunGetServiceBinding(options)

	if err != nil {
		// error while getting service binding
		return res, errors.Wrap(err, "Retrieving service binding failed")
	}

	err = btp.Logout()
	if err != nil {
		return res, errors.Wrap(err, "Logout of BTP failed")
	}

	return res, nil
}

/*
Actually runs the get service binding command and returns the result without login/logout
*/
func (btp *BTPUtils) RunGetServiceBinding(options GetServiceBindingOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	var serviceBindingBytes bytes.Buffer
	btp.Exec.Stdout(&serviceBindingBytes)

	parametersCheck := options.Subaccount == "" ||
		options.BindingName == ""

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.BindingName == "" {
			missingParams = append(missingParams, "BindingName")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return "", errors.Wrap(errors.New(errorMsg), "Failed to read service binding")
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	builder := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithName(options.BindingName)

	btpGetBindingScript, _ := builder.Build()
	err := btp.Exec.Run(btpGetBindingScript)

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", errors.Wrapf(err, "Retrieve service binding %v failed", options.BindingName)
	}

	// parse and return service binding
	serviceBindingJSON, err := GetJSON(serviceBindingBytes.String())

	if err != nil {
		return "", errors.Wrap(err, "Parsing service binding JSON failed")
	}

	return serviceBindingJSON, nil
}

func (btp *BTPUtils) DeleteServiceBinding(options DeleteServiceBindingOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	loginOptions := LoginOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return errors.Wrap(err, "Login to BTP failed")
	}

	// we are logged in --> delete service binding
	parametersCheck := options.Subaccount == "" ||
		options.BindingName == ""

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.BindingName == "" {
			missingParams = append(missingParams, "BindingName")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return errors.Wrap(errors.New(errorMsg), "Failed to delete service binding")
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	btpDeleteBindingScript, _ := NewBTPCommandBuilder().
		WithAction("delete").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithName(options.BindingName).WithConfirm().Build()

	err = btp.Exec.RunSync(RunSyncOptions{
		CmdScript:      btpDeleteBindingScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		LoginFunc: func() error {
			return btp.Login(loginOptions)
		},
		CheckFunc: func() CheckResponse {
			return IsServiceBindingDeleted(btp, GetServiceBindingOptions{
				Url:              options.Url,
				Subdomain:        options.Subdomain,
				User:             options.User,
				Password:         options.Password,
				IdentityProvider: options.IdentityProvider,
				Subaccount:       options.Subaccount,
				BindingName:      options.BindingName,
			})
		},
	})

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "Failed to delete Service-Binding: %v", options.BindingName)
	}

	err = btp.Logout()
	if err != nil {
		return errors.Wrapf(err, "Logout of BTP failed")
	}

	return nil
}

func (btp *BTPUtils) CreateServiceInstance(options CreateServiceInstanceOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	loginOptions := LoginOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return "", errors.Wrap(err, "Login to BTP failed")
	}

	// we are logged in --> create service instance
	parametersCheck := options.Subaccount == "" ||
		options.PlanName == "" ||
		options.OfferingName == "" ||
		options.InstanceName == "" ||
		options.Timeout == 0 ||
		options.PollInterval == 0

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.PlanName == "" {
			missingParams = append(missingParams, "PlanName")
		}
		if options.OfferingName == "" {
			missingParams = append(missingParams, "OfferingName")
		}
		if options.InstanceName == "" {
			missingParams = append(missingParams, "InstanceName")
		}
		if options.Timeout == 0 {
			missingParams = append(missingParams, "Timeout")
		}
		if options.PollInterval == 0 {
			missingParams = append(missingParams, "PollInterval")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return "", errors.Wrap(errors.New(errorMsg), "Failed to create service instance")
	}

	log.Entry().WithField("subaccount", options.Subaccount).
		WithField("planName", options.PlanName).
		WithField("offeringName", options.OfferingName).
		WithField("name", options.InstanceName).
		WithField("parameters", options.Parameters)

	builder := NewBTPCommandBuilder().
		WithAction("create").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount).
		WithPlanName(options.PlanName).
		WithOfferingName(options.OfferingName)

	if options.Parameters != "" {
		builder = builder.WithParameters(options.Parameters)
	}

	btpCreateInstanceScript, _ := builder.Build()

	err = btp.Exec.RunSync(RunSyncOptions{
		CmdScript:      btpCreateInstanceScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		LoginFunc: func() error {
			return btp.Login(loginOptions)
		},
		CheckFunc: func() CheckResponse {
			return IsServiceInstanceCreated(btp, GetServiceInstanceOptions{
				Url:              options.Url,
				Subdomain:        options.Subdomain,
				User:             options.User,
				Password:         options.Password,
				IdentityProvider: options.IdentityProvider,
				Subaccount:       options.Subaccount,
				InstanceName:     options.InstanceName,
			})
		},
	})

	if err != nil {
		// error while getting service instance
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", errors.Wrapf(err, "Creation of service instance failed")
	}

	// parse and return service instance
	serviceInstanceJSON, err := GetJSON(btp.Exec.GetStdoutValue())

	if err != nil {
		return "", errors.Wrap(err, "Parsing service instance JSON failed")
	}

	err = btp.Logout()
	if err != nil {
		return serviceInstanceJSON, errors.Wrap(err, "Logout of BTP failed")
	}

	return serviceInstanceJSON, nil
}

func (btp *BTPUtils) GetServiceInstance(options GetServiceInstanceOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	loginOptions := LoginOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return "", errors.Wrap(err, "Login to BTP failed")
	}

	// we are logged in --> read service instance
	res, err := btp.RunGetServiceInstance(options)

	if err != nil {
		return res, errors.Wrap(err, "Retieving service instance failed")
	}

	err = btp.Logout()
	if err != nil {
		return res, errors.Wrap(err, "Logout of BTP failed")
	}

	return res, nil
}

/*
Actually runs the get service binding command and returns the result without login/logout
*/
func (btp *BTPUtils) RunGetServiceInstance(options GetServiceInstanceOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	var serviceInstanceBytes bytes.Buffer
	btp.Exec.Stdout(&serviceInstanceBytes)

	parametersCheck := options.Subaccount == "" ||
		options.InstanceName == ""

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.InstanceName == "" {
			missingParams = append(missingParams, "InstanceName")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return "", errors.Wrap(errors.New(errorMsg), "Failed to retrieve service instance")
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.InstanceName)

	builder := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount)

	btpGetServiceScript, _ := builder.Build()
	err := btp.Exec.Run(btpGetServiceScript)

	if err != nil {
		// error while getting service instance
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", errors.Wrapf(err, "Retrieve service instance failed")
	}

	// parse and return service instance
	serviceInstanceJSON, err := GetJSON(serviceInstanceBytes.String())

	if err != nil {
		return "", errors.Wrap(err, "Parsing service instance JSON failed")
	}

	return serviceInstanceJSON, nil
}

func (btp *BTPUtils) DeleteServiceInstance(options DeleteServiceInstanceOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	loginOptions := LoginOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
	}
	err := btp.Login(loginOptions)

	if err != nil {
		// error while trying to run btp login
		return errors.Wrapf(err, "Login to BTP failed")
	}

	// we are logged in --> delete service instance
	parametersCheck := options.Subaccount == "" ||
		options.InstanceName == ""

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.InstanceName == "" {
			missingParams = append(missingParams, "InstanceName")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return errors.Wrap(errors.New(errorMsg), "Failed to delete service instance")
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.InstanceName)

	btpGetServiceScript, _ := NewBTPCommandBuilder().
		WithAction("delete").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount).
		WithConfirm().Build()

	err = btp.Exec.RunSync(RunSyncOptions{
		CmdScript:      btpGetServiceScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		LoginFunc: func() error {
			return btp.Login(loginOptions)
		},
		CheckFunc: func() CheckResponse {
			return IsServiceInstanceDeleted(btp, GetServiceInstanceOptions{
				Url:              options.Url,
				Subdomain:        options.Subdomain,
				User:             options.User,
				Password:         options.Password,
				IdentityProvider: options.IdentityProvider,
				Subaccount:       options.Subaccount,
				InstanceName:     options.InstanceName,
			})
		},
	})

	if err != nil {
		// error while deleting service instance
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "Checking if Service-Instance was deleted failed")
	}

	err = btp.Logout()
	if err != nil {
		return errors.Wrapf(err, "Logout of BTP failed")
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
	Url              string
	Subdomain        string
	Subaccount       string
	ServiceInstance  string
	BindingName      string
	Parameters       string
	User             string
	Password         string
	IdentityProvider string
	Timeout          int
	PollInterval     int
}

type GetServiceBindingOptions struct {
	Url              string
	Subdomain        string
	Subaccount       string
	BindingName      string
	User             string
	Password         string
	IdentityProvider string
}

type DeleteServiceBindingOptions struct {
	Url              string
	Subdomain        string
	Subaccount       string
	BindingName      string
	User             string
	Password         string
	IdentityProvider string
	Timeout          int
	PollInterval     int
}

type CreateServiceInstanceOptions struct {
	Url              string
	Subdomain        string
	User             string
	Password         string
	IdentityProvider string
	Subaccount       string
	PlanName         string
	OfferingName     string
	InstanceName     string
	Parameters       string
	Timeout          int
	PollInterval     int
}

type GetServiceInstanceOptions struct {
	Url              string
	Subdomain        string
	User             string
	Password         string
	IdentityProvider string
	Subaccount       string
	InstanceName     string
}

type DeleteServiceInstanceOptions struct {
	Url              string
	Subdomain        string
	User             string
	Password         string
	IdentityProvider string
	Subaccount       string
	InstanceName     string
	Timeout          int
	PollInterval     int
}
