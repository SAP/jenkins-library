package cmd

import (
	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func newBtpCreateServiceBindingUtils() btp.BTPUtils {
	e := &btp.Executor{}
	btpUtils := btp.NewBTPUtils(e)
	return *btpUtils
}

func btpCreateServiceBinding(config btpCreateServiceBindingOptions, telemetryData *telemetry.CustomData) {
	btpUtils := newBtpCreateServiceBindingUtils()

	err := runBtpCreateServiceBinding(&config, telemetryData, btpUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runBtpCreateServiceBinding(config *btpCreateServiceBindingOptions, telemetryData *telemetry.CustomData, utils btp.BTPUtils) error {
	btpConfig := btp.CreateServiceBindingOptions{
		Url:             config.Url,
		Subdomain:       config.Subdomain,
		Subaccount:      config.Subaccount,
		User:            config.User,
		Password:        config.Password,
		Tenant:          config.Tenant,
		Parameters:      config.CreateServiceConfig,
		Timeout:         config.Timeout,
		PollInterval:    config.PollInterval,
		ServiceInstance: config.ServiceInstanceName,
		BindingName:     config.ServiceBindingName,
	}

	_, err := utils.CreateServiceBinding(btpConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create BTP service binding")
	}

	log.Entry().Info("Service binding creation completed successfully")

	return nil
}
