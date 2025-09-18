package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestReleasePackagesStep(t *testing.T) {
	var config abapAddonAssemblyKitReleasePackagesOptions
	var cpe abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment
	bundle := aakaas.NewAakBundleMock()
	bundle.SetBody(responseRelease)
	utils := bundle.GetUtils()

	config.Username = "dummyUser"
	config.Password = "dummyPassword"
	t.Run("step success", func(t *testing.T) {
		//arrange
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
		//act
		err := runAbapAddonAssemblyKitReleasePackages(&config, &utils, &cpe)
		//assert
		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		err = json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.NoError(t, err)
		assert.Equal(t, "R", addonDescriptorFinal.Repositories[0].Status)
	})

	t.Run("step error - invalid input", func(t *testing.T) {
		//arrange
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Status: "L",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		//act
		err := runAbapAddonAssemblyKitReleasePackages(&config, &utils, &cpe)
		//assert
		assert.Error(t, err, "Did expect error")
		assert.Equal(t, err.Error(), "Parameter missing. Please provide the name of the package which should be released")
	})

	t.Run("step error - timeout single", func(t *testing.T) {
		//arrange
		bundle.SetError("Release not finished")
		bundle.SetMaxRuntime(1 * time.Microsecond)
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
		//act
		err := runAbapAddonAssemblyKitReleasePackages(&config, &utils, &cpe)
		//assert
		assert.Error(t, err, "Did expect error")
		assert.Equal(t, err.Error(), "Release of all packages failed/timed out - Aborting as abapEnvironmentAssembleConfirm step is not needed: Timed out")
	})
}
func TestReleasePackagesStepMix(t *testing.T) {
	var config abapAddonAssemblyKitReleasePackagesOptions
	var cpe abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment
	bundle := aakaas.NewAakBundleMock()
	bundle.SetBody(responseRelease)
	utils := bundle.GetUtils()

	config.Username = "dummyUser"
	config.Password = "dummyPassword"
	t.Run("step error - timeout mix", func(t *testing.T) {
		//arrange
		bundle.SetBodyList([]string{responseRelease, responseRelease}) //Head + Post
		bundle.SetMaxRuntime(500 * time.Microsecond)
		bundle.SetErrorInsteadOfDumpToTrue()
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName: "SAPK-002AAINDRNMSPC",
					Status:      "L",
				},
				{
					PackageName: "SAPK-001AAINDRNMSPC",
					Status:      "L",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		//act
		err := runAbapAddonAssemblyKitReleasePackages(&config, &utils, &cpe)
		//assert
		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		err = json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.NoError(t, err)
		assert.Equal(t, "R", addonDescriptorFinal.Repositories[0].Status)
		assert.Equal(t, "L", addonDescriptorFinal.Repositories[1].Status)
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
