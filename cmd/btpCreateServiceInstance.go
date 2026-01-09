package cmd

import (
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
		Parameters:       config.CreateServiceConfig,
		Timeout:          config.Timeout,
		PollInterval:     config.PollInterval,
	}

	_, err := utils.CreateServiceInstance(btpConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create BTP service instance")
	}

	log.Entry().Info("Service creation completed successfully")

	return nil
}
