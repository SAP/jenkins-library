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
