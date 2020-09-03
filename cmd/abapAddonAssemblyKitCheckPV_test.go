package cmd

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCheckPVStep(t *testing.T) {
	t.Run("step success", func(t *testing.T) {
		client := &abaputils.ClientMock{
			Body:       responseCheckPV,
			Token:      "myToken",
			StatusCode: 200,
		}

		var config abapAddonAssemblyKitCheckPVOptions
		config.AddonDescriptorFileName = "success"
		var cpe abapAddonAssemblyKitCheckPVCommonPipelineEnvironment
		err := runAbapAddonAssemblyKitCheckPV(&config, nil, client, &cpe, mockReadAddonDescriptor)
		assert.NoError(t, err, "Did not expect error")

		var addonDescriptorFinal abaputils.AddonDescriptor
		json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.Equal(t, "0003", addonDescriptorFinal.AddonVersion)
		assert.Equal(t, "0002", addonDescriptorFinal.AddonSpsLevel)
		assert.Equal(t, "0001", addonDescriptorFinal.AddonPatchLevel)
	})
	t.Run("step error - in ReadAddonDescriptor", func(t *testing.T) {
		var config abapAddonAssemblyKitCheckPVOptions
		config.AddonDescriptorFileName = "failing"
		var cpe abapAddonAssemblyKitCheckPVCommonPipelineEnvironment
		err := runAbapAddonAssemblyKitCheckPV(&config, nil, &abaputils.ClientMock{}, &cpe, mockReadAddonDescriptor)
		assert.Error(t, err, "Did expect error")
		assert.Equal(t, err.Error(), "error in ReadAddonDescriptor")
	})
	t.Run("step error - in validate", func(t *testing.T) {
		client := &abaputils.ClientMock{
			Body:       "ErrorBody",
			Token:      "myToken",
			StatusCode: 400,
			Error:      errors.New("error during validation"),
		}
		var config abapAddonAssemblyKitCheckPVOptions
		config.AddonDescriptorFileName = "success"
		var cpe abapAddonAssemblyKitCheckPVCommonPipelineEnvironment
		err := runAbapAddonAssemblyKitCheckPV(&config, nil, client, &cpe, mockReadAddonDescriptor)
		assert.Error(t, err, "Did expect error")
	})
}

func TestInitPV(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &abaputils.ClientMock{}
		prodvers := abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSPC/PRD01",
			AddonVersionYAML: "3.2.1",
		}

		var pv pv
		pv.init(prodvers, *conn)
		assert.Equal(t, "/DRNMSPC/PRD01", pv.Name)
		assert.Equal(t, "3.2.1", pv.VersionYAML)
	})
}

func TestValidatePV(t *testing.T) {
	t.Run("test validate", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &abaputils.ClientMock{
			Body:       responseCheckPV,
			Token:      "myToken",
			StatusCode: 200,
		}
		var pv pv
		pv.connector = *conn
		pv.Name = "/DRNMSPC/PRD01"
		pv.VersionYAML = "3.2.1"
		err := pv.validate()
		assert.NoError(t, err)
		assert.Equal(t, pv.Version, "0003")
		assert.Equal(t, pv.SpsLevel, "0002")
		assert.Equal(t, pv.PatchLevel, "0001")
	})
}

func TestValidatePVError(t *testing.T) {
	t.Run("test validate with error", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &abaputils.ClientMock{
			Body:       "ErrorBody",
			Token:      "myToken",
			StatusCode: 400,
			Error:      errors.New("Validation failed"),
		}
		var pv pv
		pv.connector = *conn
		pv.Name = "/DRNMSPC/PRD01"
		pv.VersionYAML = "3.2.1"
		err := pv.validate()
		assert.Error(t, err)
		assert.Equal(t, pv.Version, "")
		assert.Equal(t, pv.SpsLevel, "")
		assert.Equal(t, pv.PatchLevel, "")
	})
}

func TestCopyFieldsPV(t *testing.T) {
	t.Run("test copyFieldsToRepo", func(t *testing.T) {
		prodVers := abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSPC/PRD01",
			AddonVersionYAML: "1.2.3",
		}
		var pv pv
		pv.Version = "0003"
		pv.SpsLevel = "0002"
		pv.PatchLevel = "0001"
		pv.copyFieldsToRepo(&prodVers)
		assert.Equal(t, "0003", prodVers.AddonVersion)
		assert.Equal(t, "0002", prodVers.AddonSpsLevel)
		assert.Equal(t, "0001", prodVers.AddonPatchLevel)
	})
}

var responseCheckPV = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/ProductVersionSet(Name='%2FDRNMSPC%2FPRD01',Version='0001')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/ProductVersionSet(Name='%2FDRNMSPC%2FPRD01',Version='0001')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.ProductVersion"
        },
        "Name": "/DRNMSPC/PRD01",
        "Version": "0003",
        "SpsLevel": "0002",
        "PatchLevel": "0001",
        "Vendor": "",
        "VendorType": ""
    }
}`
