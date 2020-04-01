package cloudfoundry

import (
	"testing"
)

func TestCloudFoundry(t *testing.T) {
	//execRunner := mock.ExecMockRunner{}
	t.Run("CF Login: success case", func(t *testing.T) {
		cfconfig := CloudFoundryLoginOptions{
			CfAPIEndpoint: "https://api.cf.sap.hana.ondemand.com",
			CfOrg:         "Steampunk-2-jenkins-test",
			CfSpace:       "Test",
			Username:      "P2001217173",
			Password:      "ABAPsaas1!",
		}

		Login(cfconfig)
	})
}
