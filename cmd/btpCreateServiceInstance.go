package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func newBtpCreateServiceInstanceUtils() btp.BTPUtils {
	e := &btp.Executor{}
	btpUtils := btp.NewBTPUtils(e)
	return *btpUtils
}

func btpCreateServiceInstance(config btpCreateServiceInstanceOptions, telemetryData *telemetry.CustomData) {
	btpUtils := newBtpCreateServiceInstanceUtils()

	err := runBtpCreateServiceInstance(&config, telemetryData, btpUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runBtpCreateServiceInstance(config *btpCreateServiceInstanceOptions, telemetryData *telemetry.CustomData, utils btp.BTPUtils) error {

	btpConfig := btp.CreateServiceInstanceOptions{
		Url:              config.Url,
		Subdomain:        config.Subdomain,
		Subaccount:       config.Subaccount,
		User:             config.User,
		Password:         config.Password,
		IdentityProvider: config.Idp,
		PlanName:         config.PlanName,
		OfferingName:     config.OfferingName,
		InstanceName:     config.ServiceInstanceName,
		Timeout:          config.Timeout,
		PollInterval:     config.PollInterval,
		MaxRetries:       6,
		MaxBadRequests:   10,
	}

	if config.Parameters != "" {
		btpConfig.Parameters = config.Parameters
	} else {
		abapParameters, err := generateBTPServiceParameterString(config)
		if err != nil {
			return errors.Wrap(err, "failed to generate service parameters")
		}

		if abapParameters != "" {
			btpConfig.Parameters = abapParameters
		}
	}

	_, err := utils.CreateServiceInstance(btpConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create BTP service instance")
	}

	log.Entry().Info("Service creation completed successfully")

	return nil
}

func generateBTPServiceParameterString(config *btpCreateServiceInstanceOptions) (string, error) {
	//Check if no value provided and return an empty string
	if config.AbapSystemAdminEmail == "" &&
		config.AbapSystemDescription == "" &&
		config.AbapSystemIsDevelopmentAllowed == false &&
		config.AbapSystemID == "" &&
		config.AbapSystemSizeOfPersistence == 0 &&
		config.AbapSystemSizeOfRuntime == 0 {
		return "", nil
	}

	params := btpSystemParameters{
		AdminEmail:           config.AbapSystemAdminEmail,
		Description:          config.AbapSystemDescription,
		IsDevelopmentAllowed: &config.AbapSystemIsDevelopmentAllowed,
		SapSystemName:        config.AbapSystemID,
		SizeOfPersistence:    config.AbapSystemSizeOfPersistence,
		SizeOfRuntime:        config.AbapSystemSizeOfRuntime,
	}

	serviceParameters, err := json.Marshal(params)
	serviceParametersString := string(serviceParameters)
	log.Entry().Debugf("Service Parameters: %s", serviceParametersString)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("Could not generate parameter string for the cloud foundry cli: %w", err)
	}

	return serviceParametersString, nil
}

type btpSystemParameters struct {
	AdminEmail           string `json:"admin_email,omitempty"`
	Description          string `json:"description,omitempty"`
	IsDevelopmentAllowed *bool  `json:"is_development_allowed,omitempty"`
	SapSystemName        string `json:"sapsystemname,omitempty"`
	SizeOfPersistence    int    `json:"size_of_persistence,omitempty"`
	SizeOfRuntime        int    `json:"size_of_runtime,omitempty"`
}
