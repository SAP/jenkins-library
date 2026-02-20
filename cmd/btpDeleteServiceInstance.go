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
	}

	err := utils.DeleteServiceInstance(btpConfig)
	if err != nil {
		return errors.Wrap(err, "failed to delete BTP service instance")
	}

	log.Entry().Info("Service deletion completed successfully")

	return nil
}
