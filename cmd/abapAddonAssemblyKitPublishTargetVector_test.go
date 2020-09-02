package cmd

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPublishTargetVectorStep(t *testing.T) {

	t.Run("step success prod", func(t *testing.T) {

		config := abapAddonAssemblyKitPublishTargetVectorOptions{
			ScopeTV: "P",
		}
		addonDescriptor := abaputils.AddonDescriptor{
			TargetVectorID: "dummy",
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)

		client := &abaputils.ClientMock{
			Body:       "dummy",
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, client)

		assert.NoError(t, err, "Did not expect error")
	})

	t.Run("step success test", func(t *testing.T) {

		config := abapAddonAssemblyKitPublishTargetVectorOptions{
			ScopeTV: "T",
		}
		addonDescriptor := abaputils.AddonDescriptor{
			TargetVectorID: "dummy",
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)

		client := &abaputils.ClientMock{
			Body:       "dummy",
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, client)

		assert.NoError(t, err, "Did not expect error")
	})

	t.Run("step fail http", func(t *testing.T) {

		config := abapAddonAssemblyKitPublishTargetVectorOptions{
			ScopeTV: "T",
		}
		addonDescriptor := abaputils.AddonDescriptor{
			TargetVectorID: "dummy",
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)

		client := &abaputils.ClientMock{
			Body:       "dummy",
			Error:      errors.New("dummy"),
			Token:      "myToken",
			StatusCode: 400,
		}

		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, client)

		assert.Error(t, err, "Must end with error")
	})

	t.Run("step fail no id", func(t *testing.T) {

		config := abapAddonAssemblyKitPublishTargetVectorOptions{}

		client := &abaputils.ClientMock{
			Body:       "dummy",
			Error:      errors.New("dummy"),
			Token:      "myToken",
			StatusCode: 400,
		}

		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, client)

		assert.Error(t, err, "Must end with error")
	})
}
