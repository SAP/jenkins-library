package cloudfoundry

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
)

func TestCloudFoundry(t *testing.T) {
	//execRunner := mock.ExecMockRunner{}
	t.Run("CF Login: success case", func(t *testing.T) {
		cfconfig := CloudFoundryReadServiceKeyOptions{
			CfAPIEndpoint:     "https://api.cf.sap.hana.ondemand.com",
			CfOrg:             "Steampunk-2-jenkins-test",
			CfSpace:           "Test",
			Username:          "P2001217173",
			Password:          "ABAPsaas1!",
			CfServiceInstance: "ATCTest",
			CfServiceKey:      "TestKey",
		}

		var abapServiceKey ServiceKey
		var err error

		abapServiceKey, err = ReadServiceKey(cfconfig, false)
		if err == nil {
			log.Entry().WithField("ServiceKey", abapServiceKey.URL).Info("ServiceKey")
		}
	})
}
