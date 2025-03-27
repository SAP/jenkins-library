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

type ServiceInstanceOptions struct {
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
		_b = &Executer{}
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
	serviceBindingJSON, err := GetServiceBindingJSON(serviceBindingBytes.String())

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
		_b = &Executer{}
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
	serviceBindingJSON, err := GetServiceBindingJSON(serviceBindingBytes.String())

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
		_b = &Executer{}
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
		_b = &Executer{}
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
	serviceInstanceJSON, err := GetServiceInstanceJSON(serviceInstanceBytes.String())

	if err != nil {
		return "", err
	}

	err = btp.Logout()
	if err != nil {
		return serviceInstanceJSON, fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return serviceInstanceJSON, err
}

func (btp *BTPUtils) GetServiceInstance(options ServiceInstanceOptions) (string, error) {

	_b := btp.Exec

	if _b == nil {
		_b = &Executer{}
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
	serviceInstanceJSON, err := GetServiceInstanceJSON(serviceInstanceBytes.String())

	if err != nil {
		return "", err
	}

	err = btp.Logout()
	if err != nil {
		return serviceInstanceJSON, fmt.Errorf("Logout of BTP failed: %w", err)
	}

	return serviceInstanceJSON, err
}

func (btp *BTPUtils) DeleteServiceInstance(options ServiceInstanceOptions) error {

	_b := btp.Exec

	if _b == nil {
		_b = &Executer{}
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

func GetServiceBindingJSON(value string) (string, error) {
	if len(value) > 0 {
		var lines []string = strings.Split(value, "\n")

		var data BindingResponseData
		serviceBindingJSON, err := ConvertYAMLToJSON(strings.Join(lines[:len(lines)-2], "\n"), &data)

		if err != nil {
			return "", err
		}
		return serviceBindingJSON, nil
	}

	return "", errors.New("The returned value is empty")
}

func GetServiceInstanceJSON(value string) (string, error) {
	if len(value) > 0 {
		var lines []string = strings.Split(value, "\n")

		var data InstanceResponseData
		serviceInstanceJSON, err := ConvertYAMLToJSON(strings.Join(lines[:len(lines)-2], "\n"), &data)

		if err != nil {
			return "", err
		}
		return serviceInstanceJSON, nil
	}

	return "", errors.New("The returned value is empty")
}

type BindingResponseData struct {
	Context struct {
		CrmCustomerID     interface{} `yaml:"crm_customer_id"`
		EnvType           string      `yaml:"env_type"`
		GlobalAccountID   string      `yaml:"global_account_id"`
		InstanceName      string      `yaml:"instance_name"`
		LicenseType       string      `yaml:"license_type"`
		Origin            string      `yaml:"origin"`
		Platform          string      `yaml:"platform"`
		Region            string      `yaml:"region"`
		ServiceInstanceID string      `yaml:"service_instance_id"`
		SubaccountID      string      `yaml:"subaccount_id"`
		Subdomain         string      `yaml:"subdomain"`
		ZoneID            string      `yaml:"zone_id"`
	} `yaml:"context"`
	CreatedAt struct {
	} `yaml:"created_at"`
	CreatedBy   string `yaml:"created_by"`
	Credentials struct {
		Abap struct {
			CommunicationArrangementID       string `yaml:"communication_arrangement_id"`
			CommunicationInboundUserAuthMode int    `yaml:"communication_inbound_user_auth_mode"`
			CommunicationInboundUserID       string `yaml:"communication_inbound_user_id"`
			CommunicationScenarioID          string `yaml:"communication_scenario_id"`
			CommunicationSystemID            string `yaml:"communication_system_id"`
			CommunicationType                string `yaml:"communication_type"`
			Password                         string `yaml:"password"`
			Username                         string `yaml:"username"`
		} `yaml:"abap"`
		Binding struct {
			Env     string `yaml:"env"`
			ID      string `yaml:"id"`
			Type    string `yaml:"type"`
			Version string `yaml:"version"`
		} `yaml:"binding"`
		PreserveHostHeader bool   `yaml:"preserve_host_header"`
		SapCloudService    string `yaml:"sap.cloud.service"`
		Systemid           string `yaml:"systemid"`
		URL                string `yaml:"url"`
	} `yaml:"credentials"`
	ID            string `yaml:"id"`
	Labels        string `yaml:"labels"`
	LastOperation struct {
		CorrelationID string `yaml:"correlation_id"`
		CreatedAt     struct {
		} `yaml:"created_at"`
		DeletionScheduled struct {
		} `yaml:"deletion_scheduled"`
		ID                  string `yaml:"id"`
		PlatformID          string `yaml:"platform_id"`
		Ready               bool   `yaml:"ready"`
		Reschedule          bool   `yaml:"reschedule"`
		RescheduleTimestamp struct {
		} `yaml:"reschedule_timestamp"`
		ResourceID   string `yaml:"resource_id"`
		ResourceType string `yaml:"resource_type"`
		State        string `yaml:"state"`
		Type         string `yaml:"type"`
		UpdatedAt    struct {
		} `yaml:"updated_at"`
	} `yaml:"last_operation"`
	Name              string `yaml:"name"`
	Ready             bool   `yaml:"ready"`
	ServiceInstanceID string `yaml:"service_instance_id"`
	SubaccountID      string `yaml:"subaccount_id"`
	UpdatedAt         struct {
	} `yaml:"updated_at"`
}

type InstanceResponseData struct {
	Context struct {
		CrmCustomerID   interface{} `yaml:"crm_customer_id"`
		EnvType         string      `yaml:"env_type"`
		GlobalAccountID string      `yaml:"global_account_id"`
		InstanceName    string      `yaml:"instance_name"`
		LicenseType     string      `yaml:"license_type"`
		Origin          string      `yaml:"origin"`
		Platform        string      `yaml:"platform"`
		Region          string      `yaml:"region"`
		SubaccountID    string      `yaml:"subaccount_id"`
		Subdomain       string      `yaml:"subdomain"`
		ZoneID          string      `yaml:"zone_id"`
	} `yaml:"context"`
	CreatedAt struct {
	} `yaml:"created_at"`
	CreatedBy     string `yaml:"created_by"`
	DashboardURL  string `yaml:"dashboard_url"`
	ID            string `yaml:"id"`
	Labels        string `yaml:"labels"`
	LastOperation struct {
		CorrelationID string `yaml:"correlation_id"`
		CreatedAt     struct {
		} `yaml:"created_at"`
		DeletionScheduled struct {
		} `yaml:"deletion_scheduled"`
		Description         string `yaml:"description"`
		ID                  string `yaml:"id"`
		PlatformID          string `yaml:"platform_id"`
		Ready               bool   `yaml:"ready"`
		Reschedule          bool   `yaml:"reschedule"`
		RescheduleTimestamp struct {
		} `yaml:"reschedule_timestamp"`
		ResourceID   string `yaml:"resource_id"`
		ResourceType string `yaml:"resource_type"`
		State        string `yaml:"state"`
		Type         string `yaml:"type"`
		UpdatedAt    struct {
		} `yaml:"updated_at"`
	} `yaml:"last_operation"`
	Name          string `yaml:"name"`
	PlatformID    string `yaml:"platform_id"`
	Protected     string `yaml:"protected"`
	Ready         bool   `yaml:"ready"`
	ServicePlanID string `yaml:"service_plan_id"`
	SubaccountID  string `yaml:"subaccount_id"`
	UpdatedAt     struct {
	} `yaml:"updated_at"`
	Usable bool `yaml:"usable"`
}
