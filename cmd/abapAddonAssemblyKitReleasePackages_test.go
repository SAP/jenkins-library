package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestReleasePackagesStep(t *testing.T) {
	var config abapAddonAssemblyKitReleasePackagesOptions
	var cpe abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment
	client := &abaputils.ClientMock{
		Body:       responseRelease,
		Token:      "myToken",
		StatusCode: 200,
	}
	maxRuntime := time.Duration(1 * time.Second)
	pollingInterval := time.Duration(1 * time.Microsecond)
	config.Username = "dummyUser"
	config.Password = "dummyPassword"
	t.Run("step success", func(t *testing.T) {
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName: "SAPK-002AAINDRNMSPC",
					Status:      "L",
				},
				{
					PackageName: "SAPK-001AAINDRNMSPC",
					Status:      "R",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)

		err := runAbapAddonAssemblyKitReleasePackages(&config, nil, client, &cpe, maxRuntime, pollingInterval)

		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.Equal(t, "R", addonDescriptorFinal.Repositories[0].Status)
	})

	t.Run("step error - invalid input", func(t *testing.T) {
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Status: "L",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)

		err := runAbapAddonAssemblyKitReleasePackages(&config, nil, client, &cpe, maxRuntime, pollingInterval)
		assert.Error(t, err, "Did expect error")
		assert.Equal(t, err.Error(), "Parameter missing. Please provide the name of the package which should be released")
	})

	t.Run("step error - timeout", func(t *testing.T) {
		client.Error = errors.New("Release not finished")
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName: "SAPK-001AAINDRNMSPC",
					Status:      "L",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)

		maxRuntime := time.Duration(2 * time.Second)
		pollingInterval := time.Duration(1 * time.Second)
		err := runAbapAddonAssemblyKitReleasePackages(&config, nil, client, &cpe, maxRuntime, pollingInterval)
		assert.Error(t, err, "Did expect error")
		assert.Equal(t, err.Error(), "Timed out")
	})
}

var responseRelease = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.OcsPackage"
        },
        "Name": "SAPK-001AAINDRNMSPC",
        "Type": "AOI",
        "Component": "/DRNMSPC/COMP01",
        "Release": "0001",
        "Level": "0000",
        "Status": "R",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`
