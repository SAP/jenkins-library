//go:build unit
// +build unit

package cmd

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestCheckCVsStep(t *testing.T) {
	var config abapAddonAssemblyKitCheckCVsOptions
	var cpe abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment
	bundle := aakaas.NewAakBundleMock()
	bundle.SetBody(aakaas.ResponseCheckCVs)
	utils := bundle.GetUtils()
	config.Username = "dummyUser"
	config.Password = "dummyPassword"
	t.Run("step success", func(t *testing.T) {
		config.AddonDescriptorFileName = "success"
		err := runAbapAddonAssemblyKitCheckCVs(&config, &utils, &cpe)
		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		err = json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.NoError(t, err)
		assert.Equal(t, "0001", addonDescriptorFinal.Repositories[0].Version)
		assert.Equal(t, "0002", addonDescriptorFinal.Repositories[0].SpLevel)
		assert.Equal(t, "0003", addonDescriptorFinal.Repositories[0].PatchLevel)
		assert.Equal(t, "HUGO1234", addonDescriptorFinal.Repositories[0].CommitID)
	})
	t.Run("step error - in validate(no CommitID)", func(t *testing.T) {
		config.AddonDescriptorFileName = "noCommitID"
		err := runAbapAddonAssemblyKitCheckCVs(&config, &utils, &cpe)
		assert.Error(t, err, "Must end with error")
		assert.Contains(t, err.Error(), "CommitID missing in repo")
	})
	t.Run("step error - in ReadAddonDescriptor", func(t *testing.T) {
		config.AddonDescriptorFileName = "failing"
		err := runAbapAddonAssemblyKitCheckCVs(&config, &utils, &cpe)
		assert.Error(t, err, "Must end with error")
		assert.Contains(t, "error in ReadAddonDescriptor", err.Error())
	})
	t.Run("step error - in validate", func(t *testing.T) {
		config.AddonDescriptorFileName = "success"
		bundle.SetBody("ErrorBody")
		bundle.SetError("error during validation")
		err := runAbapAddonAssemblyKitCheckCVs(&config, &utils, &cpe)
		assert.Error(t, err, "Must end with error")
	})
}

func TestCombineYAMLRepositoriesWithCPEProduct(t *testing.T) {
	t.Run("test combineYAMLRepositoriesWithCPEProduct", func(t *testing.T) {
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Name:        "/DRNMSPC/COMP01",
					VersionYAML: "1.2.3",
				},
				{
					Name:        "/DRNMSPC/COMP02",
					VersionYAML: "3.2.1",
				},
			},
		}
		addonDescriptorFromCPE := abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSP/PROD",
			AddonVersionYAML: "1.2.3",
		}
		finalAddonDescriptor := combineYAMLRepositoriesWithCPEProduct(addonDescriptor, addonDescriptorFromCPE)
		assert.Equal(t, "/DRNMSP/PROD", finalAddonDescriptor.AddonProduct)
		assert.Equal(t, "1.2.3", finalAddonDescriptor.AddonVersionYAML)
		assert.Equal(t, "/DRNMSPC/COMP01", finalAddonDescriptor.Repositories[0].Name)
		assert.Equal(t, "/DRNMSPC/COMP02", finalAddonDescriptor.Repositories[1].Name)
		assert.Equal(t, "1.2.3", finalAddonDescriptor.Repositories[0].VersionYAML)
		assert.Equal(t, "3.2.1", finalAddonDescriptor.Repositories[1].VersionYAML)
	})
}
