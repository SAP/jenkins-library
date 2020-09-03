package cmd

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestCreateTargetVectorStep(t *testing.T) {

	t.Run("step success test", func(t *testing.T) {

		config := abapAddonAssemblyKitCreateTargetVectorOptions{}
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

		var jTV jsontargetVector
		jTV.Tv = &targetVector{
			ID: "dummy",
		}
		dummyBody, _ := json.Marshal(jTV)

		client := &abaputils.ClientMock{
			Body:       string(dummyBody),
			Token:      "myToken",
			StatusCode: 200,
		}

		cpe := abapAddonAssemblyKitCreateTargetVectorCommonPipelineEnvironment{}

		err := runAbapAddonAssemblyKitCreateTargetVector(&config, nil, client, &cpe)

		assert.NoError(t, err, "Did not expect error")
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
