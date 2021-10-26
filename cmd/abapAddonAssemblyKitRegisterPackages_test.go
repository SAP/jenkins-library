package cmd

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRegisterPackagesStep(t *testing.T) {
	var config abapAddonAssemblyKitRegisterPackagesOptions
	var cpe abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment
	t.Run("step success", func(t *testing.T) {
		client := &abaputils.ClientMock{
			BodyList: []string{responseRegisterPackagesPost, "myToken", "dummyResponseUpload"},
		}
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName:    "SAPK-002AAINDRNMSPC",
					Status:         "P",
					SarXMLFilePath: "exists",
				},
				{
					PackageName: "SAPK-001AAINDRNMSPC",
					Status:      "R",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)

		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.Equal(t, "L", addonDescriptorFinal.Repositories[0].Status)
	})
	t.Run("step error - null file", func(t *testing.T) {
		client := &abaputils.ClientMock{
			BodyList: []string{responseRegisterPackagesPost, "myToken", "dummyResponseUpload"},
		}
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName:    "SAPK-002AAINDRNMSPC",
					Status:         "P",
					SarXMLFilePath: "null",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)

		assert.Error(t, err, "Did expect error")
	})
	t.Run("step error - uploadSarFiles", func(t *testing.T) {
		client := &abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Failure during upload of SAR file"),
		}
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName:    "SAPK-002AAINDRNMSPC",
					Status:         "P",
					SarXMLFilePath: "exists",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)
		assert.Error(t, err, "Did expect error")
	})
	t.Run("step error - registerPackages - invalid input", func(t *testing.T) {
		client := &abaputils.ClientMock{
			BodyList: []string{"myToken", "dummyResponseUpload"},
		}
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Status:         "P",
					SarXMLFilePath: "exists",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)
		assert.Error(t, err, "Did expect error")
	})
}
