//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestRunAbapAddonAssemblyKitCheck(t *testing.T) {
	var config abapAddonAssemblyKitCheckOptions
	var cpe abapAddonAssemblyKitCheckCommonPipelineEnvironment
	bundle := aakaas.NewAakBundleMock()
	utils := bundle.GetUtils()
	config.Username = "dummyUser"
	config.Password = "dummyPassword"

	t.Run("happy path", func(t *testing.T) {
		config.AddonDescriptorFileName = "addon.yml.mock"
		bundle.SetBody(aakaas.ResponseCheck)
		bundle.MockAddonDescriptor = abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSPC/PRD01",
			AddonVersionYAML: "2.0.0",
			Repositories: []abaputils.Repository{
				{
					Name:        "/DRNMSPC/COMP01",
					VersionYAML: "2.0.0",
				},
				{
					Name:        "/DRNMSPC/COMP02",
					VersionYAML: "1.0.0",
				},
			},
		}

		err := runAbapAddonAssemblyKitCheck(&config, nil, utils, &cpe)

		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		config.AddonDescriptorFileName = "addon.yml.mock"
		bundle.SetBody(aakaas.ResponseCheck)
		bundle.MockAddonDescriptor = abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSPC/PRD01",
			AddonVersionYAML: "2.0.0",
			Repositories:     []abaputils.Repository{
				//no repos should fail during pvh creation...
			},
		}

		err := runAbapAddonAssemblyKitCheck(&config, nil, utils, &cpe)

		assert.EqualError(t, err, "addonDescriptor must contain at least one software component repository")
	})
}
