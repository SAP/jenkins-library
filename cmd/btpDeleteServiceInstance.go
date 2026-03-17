package cmd

import (
	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func newBtpDeleteServiceInstanceUtils() btp.BTPUtils {
	e := &btp.Executor{}
	btpUtils := btp.NewBTPUtils(e)
	return *btpUtils
}

func btpDeleteServiceInstance(config btpDeleteServiceInstanceOptions, telemetryData *telemetry.CustomData) {
	btpUtils := newBtpDeleteServiceInstanceUtils()

	err := runBtpDeleteServiceInstance(&config, telemetryData, btpUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runBtpDeleteServiceInstance(config *btpDeleteServiceInstanceOptions, telemetryData *telemetry.CustomData, utils btp.BTPUtils) error {
	btpConfig := btp.DeleteServiceInstanceOptions{
		Url:              config.Url,
		Subdomain:        config.Subdomain,
		Subaccount:       config.Subaccount,
		User:             config.User,
		Password:         config.Password,
		IdentityProvider: config.Idp,
		InstanceName:     config.ServiceInstanceName,
		Timeout:          config.Timeout,
		PollInterval:     config.PollInterval,
		MaxRetries:       6,
		MaxBadRequests:   10,
	}

	if config.DeleteServiceBindings {
		serviceBindings, err := utils.ListServiceBindings(btp.ListServiceBindingOptions{
			Url:              config.Url,
			Subdomain:        config.Subdomain,
			Subaccount:       config.Subaccount,
			User:             config.User,
			Password:         config.Password,
			IdentityProvider: config.Idp,
			ServiceInstance:  config.ServiceInstanceName,
		})
		if err != nil {
			return errors.Wrap(err, "failed to list service bindings of the service instance")
		}

		if len(serviceBindings) > 0 {
			log.Entry().Info("Found service bindings for the service instance")

			err := btpDeleteServiceBindings(*config, serviceBindings, telemetryData, utils)
			if err != nil {
				return errors.Wrap(err, "failed to delete service bindings")
			}
		}
	}

	err := utils.DeleteServiceInstance(btpConfig)
	if err != nil {
		return errors.Wrap(err, "failed to delete BTP service instance")
	}

	log.Entry().Info("Service deletion completed successfully")

	return nil
}

func btpDeleteServiceBindings(config btpDeleteServiceInstanceOptions, serviceBindings []btp.ServiceBindingData, telemetryData *telemetry.CustomData, utils btp.BTPUtils) error {
	log.Entry().Info("Deleting inherent Service Bindings of the Service Instance")

	for _, serviceBinding := range serviceBindings {
		log.Entry().WithField("bindingName", serviceBinding.Name).Info("Deleting Service Binding")
		deleteConfig := btpDeleteServiceBindingOptions{
			Url:                 config.Url,
			Subdomain:           config.Subdomain,
			Subaccount:          config.Subaccount,
			User:                config.User,
			Password:            config.Password,
			Idp:                 config.Idp,
			ServiceInstanceName: config.ServiceInstanceName,
			ServiceBindingName:  serviceBinding.Name,
			Timeout:             config.Timeout,
			PollInterval:        config.PollInterval,
		}

		err := runBtpDeleteServiceBinding(&deleteConfig, telemetryData, utils)
		if err != nil {
			return err
		}
	}

	return nil
}
