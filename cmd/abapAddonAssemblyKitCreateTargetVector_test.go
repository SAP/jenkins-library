//go:build unit

package cmd

import (
	"encoding/json"
	"testing"

	"errors"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestCreateTargetVectorStep(t *testing.T) {
	//setup
	config := abapAddonAssemblyKitCreateTargetVectorOptions{
		Username: "dummy",
		Password: "dummy",
	}
	addonDescriptor := abaputils.AddonDescriptor{
		AddonProduct:    "dummy",
		AddonVersion:    "dummy",
		AddonSpsLevel:   "dummy",
		AddonPatchLevel: "dummy",
		TargetVectorID:  "dummy",
		Repositories: []abaputils.Repository{
			{
				Name:        "dummy",
				Version:     "dummy",
				SpLevel:     "dummy",
				PatchLevel:  "dummy",
				PackageName: "dummy",
			},
		},
	}
	adoDesc, _ := json.Marshal(addonDescriptor)
	config.AddonDescriptor = string(adoDesc)

	client := &abaputils.ClientMock{
		Body: responseCreateTargetVector,
	}

	cpe := abapAddonAssemblyKitCreateTargetVectorCommonPipelineEnvironment{}

	t.Run("step success test", func(t *testing.T) {
		//act
		err := runAbapAddonAssemblyKitCreateTargetVector(&config, client, &cpe)
		//assert
		assert.NoError(t, err, "Did not expect error")

		resultAddonDescriptor := abaputils.AddonDescriptor{}
		err = json.Unmarshal([]byte(cpe.abap.addonDescriptor), &resultAddonDescriptor)
		assert.NoError(t, err)
		assert.Equal(t, "W7Q00207512600000262", resultAddonDescriptor.TargetVectorID)
	})

	t.Run("step success test", func(t *testing.T) {
		//arrange
		client := &abaputils.ClientMock{
			Body:  responseCreateTargetVector,
			Error: errors.New("dummy"),
		}
		//act
		err := runAbapAddonAssemblyKitCreateTargetVector(&config, client, &cpe)
		//assert
		assert.Error(t, err, "Must end with error")
	})

	t.Run("step error init product", func(t *testing.T) {
		//arrange
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		//act
		err := runAbapAddonAssemblyKitCreateTargetVector(&config, client, &cpe)
		//assert
		assert.Error(t, err, "Must end with error")
	})

	t.Run("step error init component", func(t *testing.T) {
		//arrange
		addonDescriptor := abaputils.AddonDescriptor{
			AddonProduct:    "dummy",
			AddonVersion:    "dummy",
			AddonSpsLevel:   "dummy",
			AddonPatchLevel: "dummy",
			TargetVectorID:  "dummy",
			Repositories: []abaputils.Repository{
				{
					Name:        "dummy",
					Version:     "dummy",
					SpLevel:     "dummy",
					PatchLevel:  "dummy",
					PackageName: "dummy",
				},
				{},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		//act
		err := runAbapAddonAssemblyKitCreateTargetVector(&config, client, &cpe)
		//assert
		assert.Error(t, err, "Must end with error")
	})

}

var responseCreateTargetVector = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/TargetVectorSet('W7Q00207512600000262')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/TargetVectorSet('W7Q00207512600000262')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.TargetVector"
        },
        "Id": "W7Q00207512600000262",
        "Vendor": "0000203069",
        "ProductName": "/DRNMSPC/PRD01",
        "ProductVersion": "0001",
        "SpsLevel": "0000",
        "PatchLevel": "0000",
        "Status": "G",
        "Content": {
            "results": [
                {
                    "__metadata": {
                        "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/TargetVectorContentSet(Id='W7Q00207512600000262',ScName='%2FDRNMSPC%2FCOMP01')",
                        "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/TargetVectorContentSet(Id='W7Q00207512600000262',ScName='%2FDRNMSPC%2FCOMP01')",
                        "type": "SSDA.AAS_ODATA_PACKAGE_SRV.TargetVectorContent"
                    },
                    "Id": "W7Q00207512600000262",
                    "ScName": "/DRNMSPC/COMP01",
                    "ScVersion": "0001",
                    "DeliveryPackage": "SAPK-001AAINDRNMSPC",
                    "SpLevel": "0000",
                    "PatchLevel": "0000",
                    "Header": {
                        "__deferred": {
                            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/TargetVectorContentSet(Id='W7Q00207512600000262',ScName='%2FDRNMSPC%2FCOMP01')/Header"
                        }
                    }
                }
            ]
        }
    }
}`
