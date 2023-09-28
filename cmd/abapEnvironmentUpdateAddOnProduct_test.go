package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockedConfig = abapEnvironmentUpdateAddOnProductOptions{
	LandscapePortalAPIServiceKey: `{
		"url": "api.landscape-portal.stagingaws.hanavlab.ondemand.com",
		"uaa": {
		    "clientid": "clientid",
		    "url": "https://some-url.com",
		    "certificate": "-----BEGIN CERTIFICATE-----\nsome-certificate\n-----END CERTIFICATE-----\n",
		    "certurl": "https://some-certurl.com",
		    "credential-type": "x509",
		    "key": "-----BEGIN RSA PRIVATE KEY-----\nsome-key\n-----END RSA PRIVATE KEY-----\n"
		},
		"vendor": "SAP"
	    }`,
	AbapSystemNumber: "abapSystemNumber",
	AddonDescriptorFileName: "addon.yml",
	AddonDescriptor: "addonDescriptor",
}

var servKey serviceKey

func TestRunPrepareToGetLPAPIAccessToken(t *testing.T) {
	t.Run("Successfully parse the service key JSON", func(t *testing.T) {

	})
}
func TestRunAbapEnvironmentUpdateAddOnProduct(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentUpdateAddOnProductOptions{}

		utils := newAbapEnvironmentUpdateAddOnProductTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runAbapEnvironmentUpdateAddOnProduct(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentUpdateAddOnProductOptions{}

		utils := newAbapEnvironmentUpdateAddOnProductTestsUtils()

		// test
		err := runAbapEnvironmentUpdateAddOnProduct(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
