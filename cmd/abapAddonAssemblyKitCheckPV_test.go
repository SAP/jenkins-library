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

func TestCheckPVStep(t *testing.T) {
	var config abapAddonAssemblyKitCheckPVOptions
	var cpe abapAddonAssemblyKitCheckPVCommonPipelineEnvironment
	bundle := aakaas.NewAakBundleMock()
	bundle.SetBody(aakaas.ResponseCheckPV)
	utils := bundle.GetUtils()
	config.Username = "dummyUser"
	config.Password = "dummyPassword"
	t.Run("step success", func(t *testing.T) {
		config.AddonDescriptorFileName = "success"
		err := runAbapAddonAssemblyKitCheckPV(&config, utils, &cpe)
		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		err = json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.NoError(t, err)
		assert.Equal(t, "0003", addonDescriptorFinal.AddonVersion)
		assert.Equal(t, "0002", addonDescriptorFinal.AddonSpsLevel)
		assert.Equal(t, "0001", addonDescriptorFinal.AddonPatchLevel)
	})
	t.Run("step error - in ReadAddonDescriptor", func(t *testing.T) {
		config.AddonDescriptorFileName = "failing"
		err := runAbapAddonAssemblyKitCheckPV(&config, utils, &cpe)
		assert.Error(t, err, "Did expect error")
		assert.Equal(t, err.Error(), "error in ReadAddonDescriptor")
	})
	t.Run("step error - in validate", func(t *testing.T) {
		config.AddonDescriptorFileName = "success"
		bundle.SetBody("ErrorBody")
		bundle.SetError("error during validation")
		err := runAbapAddonAssemblyKitCheckPV(&config, utils, &cpe)
		assert.Error(t, err, "Did expect error")
	})
}
