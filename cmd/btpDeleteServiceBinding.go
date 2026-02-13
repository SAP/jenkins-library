package cmd

import (
	"github.com/SAP/jenkins-library/pkg/btp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func newBtpDeleteServiceBindingUtils() btp.BTPUtils {
	e := &btp.Executor{}
	btpUtils := btp.NewBTPUtils(e)
	return *btpUtils
}

func btpDeleteServiceBinding(config btpDeleteServiceBindingOptions, telemetryData *telemetry.CustomData) {
	btpUtils := newBtpDeleteServiceBindingUtils()

	err := runBtpDeleteServiceBinding(&config, telemetryData, btpUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runBtpDeleteServiceBinding(config *btpDeleteServiceBindingOptions, telemetryData *telemetry.CustomData, utils btp.BTPUtils) error {
	btpConfig := btp.DeleteServiceBindingOptions{
		Url:              config.Url,
		Subdomain:        config.Subdomain,
		Subaccount:       config.Subaccount,
		User:             config.User,
		Password:         config.Password,
		IdentityProvider: config.Idp,
		BindingName:      config.ServiceBindingName,
		Timeout:          config.Timeout,
		PollInterval:     config.PollInterval,
		ServiceInstance:  config.ServiceInstanceName,
	}

	err := utils.DeleteServiceBinding(btpConfig)
	if err != nil {
		return errors.Wrap(err, "failed to delete BTP service binding")
	}

	log.Entry().Info("Service binding deletion completed successfully")

	return nil
}
