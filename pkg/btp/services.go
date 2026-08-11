package btp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

func (btp *BTPUtils) CreateServiceBinding(options CreateServiceBindingOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}
	var responseBytes bytes.Buffer
	btp.Exec.Stdout(&responseBytes)

	var errorResBytes bytes.Buffer
	btp.Exec.Stderr(&errorResBytes)

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
		return "", fmt.Errorf("login to BTP failed: %w", err)
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
		if options.ServiceInstance == "" {
			missingParams = append(missingParams, "ServiceInstance")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return "", fmt.Errorf("failed to create service binding: %s", errorMsg)
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

	instanceData, err := GetServiceInstanceData(btp, GetServiceInstanceOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
		Subaccount:       options.Subaccount,
		InstanceName:     options.ServiceInstance,
	})

	if err != nil {
		return "", fmt.Errorf("failed to get service instance data: %w", err)
	}

	err = btp.Exec.RunSync(RunSyncOptions{
		CmdScript:      btpCreateBindingScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		LoginFunc: func() error {
			pollLoginOptions := loginOptions
			pollLoginOptions.Silent = true
			return btp.Login(pollLoginOptions)
		},
		CheckFunc: func() CheckResponse {
			return IsServiceBindingCreated(btp, GetServiceBindingOptions{
				Url:               options.Url,
				Subdomain:         options.Subdomain,
				User:              options.User,
				Password:          options.Password,
				IdentityProvider:  options.IdentityProvider,
				Subaccount:        options.Subaccount,
				BindingName:       options.BindingName,
				ServiceInstance:   options.ServiceInstance,
				ServiceInstanceId: instanceData.ID,
			})
		},
		IgnoreErrorOnFirstCall: true,
		maxRetries:             options.MaxRetries,
		maxBadRequests:         options.MaxBadRequests,
	})

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("creation of service binding failed for binding : %s: %w", options.BindingName, err)
	}

	// parse and return service binding
	serviceBindingJSON, err := GetJSON(btp.Exec.GetStdoutValue())

	if err != nil {
		return "", fmt.Errorf("parsing service binding JSON failed: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		return serviceBindingJSON, fmt.Errorf("logout of BTP failed: %w", err)
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
		return "", fmt.Errorf("login to BTP failed: %w", err)
	}

	// we are logged in --> read service binding
	res, err := btp.RunGetServiceBinding(options)

	if err != nil {
		// error while getting service binding
		return res, fmt.Errorf("retrieving service binding failed: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		return res, fmt.Errorf("logout of BTP failed: %w", err)
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

	var errorResBytes bytes.Buffer
	btp.Exec.Stderr(&errorResBytes)

	parametersCheck := options.Subaccount == "" ||
		(options.BindingName == "" && options.BindingId == "")

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.BindingName == "" && options.BindingId == "" {
			missingParams = append(missingParams, "BindingName or BindingId")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return "", fmt.Errorf("failed to read service binding: %s", errorMsg)
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	builder := NewBTPCommandBuilder().
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount)

	serviceBindingJSON := ""

	checkIfListMode := options.BindingName != "" && (options.ServiceInstanceId != "" || options.ServiceInstance != "")
	if checkIfListMode {
		serviceBindings, err := btp.RunListServiceBindings(ListServiceBindingOptions{
			Url:               options.Url,
			Subdomain:         options.Subdomain,
			User:              options.User,
			Password:          options.Password,
			IdentityProvider:  options.IdentityProvider,
			Subaccount:        options.Subaccount,
			ServiceInstance:   options.ServiceInstance,
			ServiceInstanceId: options.ServiceInstanceId,
			BindingName:       options.BindingName,
		})

		if err != nil {
			log.SetErrorCategory(log.ErrorService)
			return "", fmt.Errorf("listing service bindings failed: %w", err)
		}

		if len(serviceBindings) == 0 {
			log.SetErrorCategory(log.ErrorService)
			return "", errors.New("no service binding found with the given name and service instance id")
		}

		data := serviceBindings[0]
		foundedServiceBindingBytes, err := json.Marshal(data)
		if err != nil {
			log.SetErrorCategory(log.ErrorService)
			return "", fmt.Errorf("failed to marshal service binding data: %w", err)
		}
		serviceBindingJSON = string(foundedServiceBindingBytes)
	} else {
		builder = builder.WithAction("get")

		if options.BindingId != "" {
			builder = builder.WithID(options.BindingId)
		} else {
			builder = builder.WithName(options.BindingName)
		}

		btpGetBindingScript, _ := builder.Build()
		err := btp.Exec.Run(btpGetBindingScript)

		if err != nil {
			// error while getting service binding
			log.SetErrorCategory(log.ErrorService)
			return "", fmt.Errorf("retrieve service binding %v failed: %w", options.BindingName, err)
		}

		// parse and return service binding
		serviceBindingJSON, err = GetJSON(serviceBindingBytes.String())

		if err != nil {
			log.SetErrorCategory(log.ErrorService)
			return "", fmt.Errorf("parsing service binding JSON failed: %w", err)
		}
	}
	return serviceBindingJSON, nil
}

func (btp *BTPUtils) RunListServiceBindings(options ListServiceBindingOptions) ([]ServiceBindingData, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}

	parametersCheck := options.Subaccount == "" || (options.ServiceInstance == "" && options.ServiceInstanceId == "")
	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.ServiceInstance == "" && options.ServiceInstanceId == "" {
			missingParams = append(missingParams, "ServiceInstance or ServiceInstanceId")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return nil, fmt.Errorf("failed to list service bindings: %s", errorMsg)
	}

	instanceID := options.ServiceInstanceId
	if instanceID == "" {
		instanceData, err := GetServiceInstanceData(btp, GetServiceInstanceOptions{
			Url:              options.Url,
			Subdomain:        options.Subdomain,
			User:             options.User,
			Password:         options.Password,
			IdentityProvider: options.IdentityProvider,
			Subaccount:       options.Subaccount,
			InstanceName:     options.ServiceInstance,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to get service instance data: %w", err)
		}

		instanceID = instanceData.ID
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("serviceInstanceId", instanceID)

	fieldsFilter := ""
	if options.BindingName != "" {
		fieldsFilter = fmt.Sprintf("\"service_instance_id eq '%s' and name eq '%s'\"", instanceID, options.BindingName)
	} else {
		fieldsFilter = fmt.Sprintf("\"service_instance_id eq '%s'\"", instanceID)
	}

	builder := NewBTPCommandBuilder().
		WithAction("list").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithFieldsFilter(fieldsFilter)

	btpGetBindingScript, _ := builder.Build()

	var serviceBindingBytes bytes.Buffer
	btp.Exec.Stdout(&serviceBindingBytes)

	var errorResBytes bytes.Buffer
	btp.Exec.Stderr(&errorResBytes)

	err := btp.Exec.Run(btpGetBindingScript)
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return nil, fmt.Errorf("list service bindings failed for instance id: %s: %w", instanceID, err)
	}

	serviceBindingJSON, err := GetJSON(serviceBindingBytes.String())
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return nil, fmt.Errorf("parsing service bindings JSON failed: %w", err)
	}

	serviceBindingData := []ServiceBindingData{}
	err = json.Unmarshal([]byte(serviceBindingJSON), &serviceBindingData)
	if err != nil {
		// fallback: if result is a single object instead of array
		singleBinding := ServiceBindingData{}
		err2 := json.Unmarshal([]byte(serviceBindingJSON), &singleBinding)
		if err2 != nil {
			log.SetErrorCategory(log.ErrorService)
			return nil, fmt.Errorf("failed to unmarshal service bindings JSON: %w", err)
		}
		serviceBindingData = append(serviceBindingData, singleBinding)
	}

	return serviceBindingData, nil
}

func (btp *BTPUtils) ListServiceBindings(options ListServiceBindingOptions) ([]ServiceBindingData, error) {
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
	if err := btp.Login(loginOptions); err != nil {
		return nil, fmt.Errorf("login to BTP failed: %w", err)
	}

	bindings, err := btp.RunListServiceBindings(options)
	if err != nil {
		return nil, fmt.Errorf("retrieving service bindings failed: %w", err)
	}

	if err := btp.Logout(); err != nil {
		return bindings, fmt.Errorf("logout of BTP failed: %w", err)
	}

	return bindings, nil
}

func (btp *BTPUtils) DeleteServiceBinding(options DeleteServiceBindingOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}
	var responseBytes bytes.Buffer
	btp.Exec.Stdout(&responseBytes)

	var errorResBytes bytes.Buffer
	btp.Exec.Stderr(&errorResBytes)

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
		return fmt.Errorf("login to BTP failed: %w", err)
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
		if options.ServiceInstance == "" {
			missingParams = append(missingParams, "ServiceInstance")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return fmt.Errorf("failed to delete service binding: %s", errorMsg)
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.BindingName)

	instanceData, err := GetServiceInstanceData(btp, GetServiceInstanceOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
		Subaccount:       options.Subaccount,
		InstanceName:     options.ServiceInstance,
	})

	if err != nil {
		return fmt.Errorf("failed to get service instance data: %w", err)
	}

	bindingData, err := GetServiceBindingData(btp, GetServiceBindingOptions{
		Url:               options.Url,
		Subdomain:         options.Subdomain,
		User:              options.User,
		Password:          options.Password,
		IdentityProvider:  options.IdentityProvider,
		Subaccount:        options.Subaccount,
		BindingName:       options.BindingName,
		ServiceInstanceId: instanceData.ID,
	})

	if err != nil {
		return fmt.Errorf("service binding not found: %w", err)
	}

	btpDeleteBindingScript, _ := NewBTPCommandBuilder().
		WithAction("delete").
		WithTarget("services/binding").
		WithSubAccount(options.Subaccount).
		WithID(bindingData.ID).WithConfirm().Build()

	err = btp.Exec.RunSync(RunSyncOptions{
		CmdScript:      btpDeleteBindingScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		LoginFunc: func() error {
			pollLoginOptions := loginOptions
			pollLoginOptions.Silent = true
			return btp.Login(pollLoginOptions)
		},
		CheckFunc: func() CheckResponse {
			return IsServiceBindingDeleted(btp, GetServiceBindingOptions{
				Url:              options.Url,
				Subdomain:        options.Subdomain,
				User:             options.User,
				Password:         options.Password,
				IdentityProvider: options.IdentityProvider,
				Subaccount:       options.Subaccount,
				BindingId:        options.BindingName,
			})
		},
		maxRetries:     options.MaxRetries,
		maxBadRequests: options.MaxBadRequests,
	})

	if err != nil {
		// error while getting service binding
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to delete Service-Binding: %v: %w", options.BindingName, err)
	}

	err = btp.Logout()
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("logout of BTP failed: %w", err)
	}

	return nil
}

func (btp *BTPUtils) CreateServiceInstance(options CreateServiceInstanceOptions) (string, error) {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}
	var responseBytes bytes.Buffer
	btp.Exec.Stdout(&responseBytes)

	var errorResBytes bytes.Buffer
	btp.Exec.Stderr(&errorResBytes)

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
		return "", fmt.Errorf("login to BTP failed: %w", err)
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

		return "", fmt.Errorf("failed to create service instance: %s", errorMsg)
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
			pollLoginOptions := loginOptions
			pollLoginOptions.Silent = true
			return btp.Login(pollLoginOptions)
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
		maxRetries:     options.MaxRetries,
		maxBadRequests: options.MaxBadRequests,
	})

	if err != nil {
		// error while getting service instance
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("creation of service instance failed: %w", err)
	}

	// parse and return service instance
	serviceInstanceJSON, err := GetJSON(btp.Exec.GetStdoutValue())

	if err != nil {
		return "", fmt.Errorf("parsing service instance JSON failed: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		return serviceInstanceJSON, fmt.Errorf("logout of BTP failed: %w", err)
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
		return "", fmt.Errorf("login to BTP failed: %w", err)
	}

	// we are logged in --> read service instance
	res, err := btp.RunGetServiceInstance(options)

	if err != nil {
		return res, fmt.Errorf("retrieving service instance failed: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		return res, fmt.Errorf("logout of BTP failed: %w", err)
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

	var errorResBytes bytes.Buffer
	btp.Exec.Stderr(&errorResBytes)

	parametersCheck := options.Subaccount == "" ||
		(options.InstanceName == "" && options.InstanceId == "")

	if parametersCheck {
		errorMsg := "Parameters missing. Please provide: "
		missingParams := []string{}
		if options.Subaccount == "" {
			missingParams = append(missingParams, "Subaccount")
		}
		if options.InstanceName == "" && options.InstanceId == "" {
			missingParams = append(missingParams, "InstanceName or InstanceId")
		}
		errorMsg += strings.Join(missingParams, ", ")

		return "", fmt.Errorf("failed to retrieve service instance: %s", errorMsg)
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.InstanceName)

	builder := NewBTPCommandBuilder().
		WithAction("get").
		WithTarget("services/instance").
		WithSubAccount(options.Subaccount)

	if options.InstanceId != "" {
		builder = builder.WithID(options.InstanceId)
	} else {
		builder = builder.WithName(options.InstanceName)
	}

	btpGetServiceScript, _ := builder.Build()
	err := btp.Exec.Run(btpGetServiceScript)

	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("retrieving service instance failed: %w", err)
	}

	// parse and return service instance
	serviceInstanceJSON, err := GetJSON(serviceInstanceBytes.String())

	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return "", fmt.Errorf("parsing service instance JSON failed: %w", err)
	}

	return serviceInstanceJSON, nil
}

func (btp *BTPUtils) DeleteServiceInstance(options DeleteServiceInstanceOptions) error {
	if btp.Exec == nil {
		btp.Exec = &Executor{}
	}
	var responseBytes bytes.Buffer
	btp.Exec.Stdout(&responseBytes)

	var errorResBytes bytes.Buffer
	btp.Exec.Stderr(&errorResBytes)

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
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("login to BTP failed: %w", err)
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

		return fmt.Errorf("failed to delete service instance: %s", errorMsg)
	}

	log.Entry().WithField("subaccount", options.Subaccount).WithField("name", options.InstanceName)

	instanceData, err := GetServiceInstanceData(btp, GetServiceInstanceOptions{
		Url:              options.Url,
		Subdomain:        options.Subdomain,
		User:             options.User,
		Password:         options.Password,
		IdentityProvider: options.IdentityProvider,
		Subaccount:       options.Subaccount,
		InstanceName:     options.InstanceName,
	})

	if err != nil {
		return fmt.Errorf("service instance not found: %w", err)
	}

	btpGetServiceScript, _ := NewBTPCommandBuilder().
		WithAction("delete").
		WithTarget("services/instance").
		WithID(instanceData.ID).
		WithSubAccount(options.Subaccount).
		WithConfirm().Build()

	err = btp.Exec.RunSync(RunSyncOptions{
		CmdScript:      btpGetServiceScript,
		TimeoutSeconds: options.Timeout,
		PollInterval:   options.PollInterval,
		LoginFunc: func() error {
			pollLoginOptions := loginOptions
			pollLoginOptions.Silent = true
			return btp.Login(pollLoginOptions)
		},
		CheckFunc: func() CheckResponse {
			return IsServiceInstanceDeleted(btp, GetServiceInstanceOptions{
				Url:              options.Url,
				Subdomain:        options.Subdomain,
				User:             options.User,
				Password:         options.Password,
				IdentityProvider: options.IdentityProvider,
				Subaccount:       options.Subaccount,
				InstanceId:       instanceData.ID,
			})
		},
		maxRetries:     options.MaxRetries,
		maxBadRequests: options.MaxBadRequests,
	})

	if err != nil {
		// error while deleting service instance
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("checking if Service-Instance was deleted failed: %w", err)
	}

	err = btp.Logout()
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("logout of BTP failed: %w", err)
	}

	return nil
}

func GetJSON(value string) (string, error) {
	var jsonContent string

	if len(value) > 0 {
		// parse and return service key
		var lines []string = strings.Split(value, "\n")
		jsonContent = strings.Join(lines, "")

		return jsonContent, nil
	}

	return "", errors.New("the returned value is empty")
}

func GetServiceInstanceData(btp *BTPUtils, options GetServiceInstanceOptions) (ServiceInstanceData, error) {
	serviceInstanceJSON, err := btp.RunGetServiceInstance(options)

	if err != nil {
		return ServiceInstanceData{}, fmt.Errorf("failed to get service instance: %w", err)
	}

	instanceData := ServiceInstanceData{}
	err = json.Unmarshal([]byte(serviceInstanceJSON), &instanceData)
	if err != nil {
		return ServiceInstanceData{}, fmt.Errorf("failed to unmarshal service instance JSON: %w", err)
	}
	return instanceData, nil
}

func GetServiceBindingData(btp *BTPUtils, options GetServiceBindingOptions) (ServiceBindingData, error) {
	serviceBindingJSON, err := btp.RunGetServiceBinding(options)

	if err != nil {
		return ServiceBindingData{}, fmt.Errorf("failed to get service binding: %w", err)
	}

	instanceData := ServiceBindingData{}
	err = json.Unmarshal([]byte(serviceBindingJSON), &instanceData)
	if err != nil {
		return ServiceBindingData{}, fmt.Errorf("failed to unmarshal service binding JSON: %w", err)
	}
	return instanceData, nil
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
	MaxRetries       int
	MaxBadRequests   int
}

type GetServiceBindingOptions struct {
	Url               string
	Subdomain         string
	Subaccount        string
	BindingName       string
	User              string
	Password          string
	IdentityProvider  string
	BindingId         string
	ServiceInstance   string
	ServiceInstanceId string
}

type ListServiceBindingOptions struct {
	Url               string
	Subdomain         string
	Subaccount        string
	User              string
	Password          string
	IdentityProvider  string
	ServiceInstance   string
	ServiceInstanceId string
	BindingName       string
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
	ServiceInstance  string
	MaxRetries       int
	MaxBadRequests   int
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
	MaxRetries       int
	MaxBadRequests   int
}

type GetServiceInstanceOptions struct {
	Url              string
	Subdomain        string
	User             string
	Password         string
	IdentityProvider string
	Subaccount       string
	InstanceName     string
	InstanceId       string
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
	MaxRetries       int
	MaxBadRequests   int
}
