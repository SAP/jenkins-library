package btp

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

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

type CreateServiceInstanceOptions struct {
	Url             string
	Subdomain       string
	ServiceInstance string
	User            string
	Password        string
	Tenant          string
	Subaccount      string
	PlanName        string
	OfferingName    string
	InstanceName    string
	Parameters      string
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

func (btp *BTPUtils) CreateServiceBinding(options CreateServiceBindingOptions) (string, error) {

	_b := btp.Exec

	if _b == nil {
		_b = &Executor{}
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
	_b.Stdout(&serviceBindingBytes)

	// we are logged in --> create service binding
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	btpCreateBindingScript, _ := NewBTPCommandBuilder().
		WithAction("create").
		WithTarget("services/binding").
		WithName(options.BindingName).
		WithSubAccount(options.Subaccount).
		WithParameters(options.Parameters).Build()

	btpVerifScript, _ := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/binding").
		WithName(options.BindingName).
		WithSubAccount(options.Subaccount).Build()

	err = _b.RunSync(btpCreateBindingScript, btpVerifScript, 1, 30, false)

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

	return serviceBindingJSON, err
}

func (btp *BTPUtils) GetServiceBinding(options GetServiceBindingOptions) (string, error) {

	_b := btp.Exec

	if _b == nil {
		_b = &Executor{}
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
	_b.Stdout(&serviceBindingBytes)

	// we are logged in --> read service binding
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	builder := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithName(options.BindingName)

	btpGetBindingScript, _ := builder.Build()
	err = _b.Run(btpGetBindingScript)

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

	return serviceBindingJSON, err
}

func (btp *BTPUtils) DeleteServiceBinding(options GetServiceBindingOptions) error {

	_b := btp.Exec

	if _b == nil {
		_b = &Executor{}
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
	_b.Stdout(&serviceBindingBytes)

	// we are logged in --> delete service binding
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	btpDeleteBindingScript, _ := NewBTPCommandBuilder().
		WithAction("delete").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithName(options.BindingName).WithConfirm().Build()

	btpCheckScript, _ := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithName(options.BindingName).Build()

	err = _b.RunSync(btpDeleteBindingScript, btpCheckScript, 1, 30, true)

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("Checking if Service-Binding was deleted failed: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		return fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return err
}

func (btp *BTPUtils) CreateServiceInstance(options CreateServiceInstanceOptions) (string, error) {

	_b := btp.Exec

	if _b == nil {
		_b = &Executor{}
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
	_b.Stdout(&serviceInstanceBytes)

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

	btpVerifScript, _ := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount).Build()

	err = _b.RunSync(btpCreateInstanceScript, btpVerifScript, 1, 30, false)

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

	return serviceInstanceJSON, err
}

func (btp *BTPUtils) GetServiceInstance(options GetServiceInstanceOptions) (string, error) {

	_b := btp.Exec

	if _b == nil {
		_b = &Executor{}
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
	_b.Stdout(&serviceInstanceBytes)

	// we are logged in --> read service instance
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.InstanceName)

	builder := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount)

	btpGetServiceScript, _ := builder.Build()
	err = _b.Run(btpGetServiceScript)

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

	return serviceInstanceJSON, err
}

func (btp *BTPUtils) DeleteServiceInstance(options GetServiceInstanceOptions) error {

	_b := btp.Exec

	if _b == nil {
		_b = &Executor{}
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
	_b.Stdout(&serviceInstanceBytes)

	// we are logged in --> delete service instance
	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.InstanceName)

	btpGetServiceScript, _ := NewBTPCommandBuilder().
		WithAction("delete").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount).
		WithConfirm().Build()

	btpCheckScript, _ := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/instance").
		WithName(options.InstanceName).
		WithSubAccount(options.Subaccount).Build()

	err = _b.RunSync(btpGetServiceScript, btpCheckScript, 1, 60, true)

	if err != nil {
		// error while deleting service instance
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("Checking if Service-Instance was deleted failed: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		return fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return err
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
