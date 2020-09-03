package cmd

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func mockReadAddonDescriptor(FileName string) (abaputils.AddonDescriptor, error) {
	var addonDescriptor abaputils.AddonDescriptor
	var err error
	switch FileName {
	case "success":
		{
			addonDescriptor = abaputils.AddonDescriptor{
				AddonProduct:     "/DRNMSPC/PRD01",
				AddonVersionYAML: "3.2.1",
				Repositories: []abaputils.Repository{
					{
						Name:        "/DRNMSPC/COMP01",
						VersionYAML: "1.2.3",
					},
				},
			}
		}
	case "failing":
		{
			err = errors.New("error in ReadAddonDescriptor")
		}
	}
	return addonDescriptor, err
}
func TestCheckCVsStep(t *testing.T) {
	t.Run("step success", func(t *testing.T) {
		client := &abaputils.ClientMock{
			Body:       responseCheckCVs,
			Token:      "myToken",
			StatusCode: 200,
		}

		var config abapAddonAssemblyKitCheckCVsOptions
		config.AddonDescriptorFileName = "success"
		var cpe abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment
		err := runAbapAddonAssemblyKitCheckCVs(&config, nil, client, &cpe, mockReadAddonDescriptor)
		assert.NoError(t, err, "Did not expect error")

		var addonDescriptorFinal abaputils.AddonDescriptor
		json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.Equal(t, "0001", addonDescriptorFinal.Repositories[0].Version)
		assert.Equal(t, "0002", addonDescriptorFinal.Repositories[0].SpLevel)
		assert.Equal(t, "0003", addonDescriptorFinal.Repositories[0].PatchLevel)
	})
	t.Run("step error - in ReadAddonDescriptor", func(t *testing.T) {
		var config abapAddonAssemblyKitCheckCVsOptions
		config.AddonDescriptorFileName = "failing"
		var cpe abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment
		err := runAbapAddonAssemblyKitCheckCVs(&config, nil, &abaputils.ClientMock{}, &cpe, mockReadAddonDescriptor)
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

		var config abapAddonAssemblyKitCheckCVsOptions
		config.AddonDescriptorFileName = "success"
		var cpe abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment
		err := runAbapAddonAssemblyKitCheckCVs(&config, nil, client, &cpe, mockReadAddonDescriptor)
		assert.Error(t, err, "Did expect error")
	})
}

func TestInitCV(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &abaputils.ClientMock{}
		repo := abaputils.Repository{
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.2.3",
		}
		var c cv
		c.init(repo, *conn)
		assert.Equal(t, "/DRNMSPC/COMP01", c.Name)
		assert.Equal(t, "1.2.3", c.VersionYAML)
	})
}

func TestValidateCV(t *testing.T) {
	t.Run("test validate", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &abaputils.ClientMock{
			Body:       responseCheckCVs,
			Token:      "myToken",
			StatusCode: 200,
		}

		var c cv
		c.connector = *conn
		c.Name = "/DRNMSPC/COMP01"
		c.VersionYAML = "1.2.3"
		err := c.validate()
		assert.NoError(t, err)
		assert.Equal(t, c.Version, "0001")
		assert.Equal(t, c.SpLevel, "0002")
		assert.Equal(t, c.PatchLevel, "0003")
	})
}

func TestValidateCVError(t *testing.T) {
	t.Run("test validate with error", func(t *testing.T) {
		conn := new(connector)
		conn.Client = &abaputils.ClientMock{
			Body:       "ErrorBody",
			Token:      "myToken",
			StatusCode: 400,
			Error:      errors.New("Validation failed"),
		}

		var c cv
		c.connector = *conn
		c.Name = "/DRNMSPC/COMP01"
		c.VersionYAML = "1.2.3"
		err := c.validate()
		assert.Error(t, err)
		assert.Equal(t, c.Version, "")
		assert.Equal(t, c.SpLevel, "")
		assert.Equal(t, c.PatchLevel, "")
	})
}

func TestCopyFieldsCV(t *testing.T) {
	t.Run("test copyFieldsToRepo", func(t *testing.T) {
		repo := abaputils.Repository{
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.2.3",
		}
		var c cv
		c.Version = "0001"
		c.SpLevel = "0002"
		c.PatchLevel = "0003"
		c.copyFieldsToRepo(&repo)
		assert.Equal(t, "0001", repo.Version)
		assert.Equal(t, "0002", repo.SpLevel)
		assert.Equal(t, "0003", repo.PatchLevel)
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

var responseCheckCVs = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/SoftwareComponentVersionSet(Name='%2FDRNMSPC%2FCOMP01',Version='0001')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/SoftwareComponentVersionSet(Name='%2FDRNMSPC%2FCOMP01',Version='0001')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.SoftwareComponentVersion"
        },
        "Name": "/DRNMSPC/COMP01",
        "Version": "0001",
        "SpLevel": "0002",
        "PatchLevel": "0003",
        "Vendor": "",
        "VendorType": ""
    }
}`
