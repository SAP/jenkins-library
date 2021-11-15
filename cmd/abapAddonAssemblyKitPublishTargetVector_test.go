package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPublishTargetVectorStep(t *testing.T) {
	//setup
	config := abapAddonAssemblyKitPublishTargetVectorOptions{
		TargetVectorScope: "P",
		Username:          "dummy",
		Password:          "dummy",
	}
	addonDescriptor := abaputils.AddonDescriptor{
		TargetVectorID: "W7Q00207512600000353",
	}
	adoDesc, _ := json.Marshal(addonDescriptor)
	config.AddonDescriptor = string(adoDesc)

	t.Run("step success prod", func(t *testing.T) {
		//arrange
		mc := abapbuild.NewMockClient()
		mc.AddData(aakaas.AAKaaSHead)
		mc.AddData(aakaas.AAKaaSTVPublishProdPost)
		mc.AddData(aakaas.AAKaaSGetTVPublishRunning)
		mc.AddData(aakaas.AAKaaSGetTVPublishProdSuccess)

		//act
		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, &mc, time.Duration(1*time.Second), time.Duration(1*time.Microsecond))
		//assert
		assert.NoError(t, err, "Did not expect error")
	})

	t.Run("step success test", func(t *testing.T) {
		//arrange
		config.TargetVectorScope = "T"
		mc := abapbuild.NewMockClient()
		mc.AddData(aakaas.AAKaaSHead)
		mc.AddData(aakaas.AAKaaSTVPublishTestPost)
		mc.AddData(aakaas.AAKaaSGetTVPublishRunning)
		mc.AddData(aakaas.AAKaaSGetTVPublishTestSuccess)
		//act
		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, &mc, time.Duration(1*time.Second), time.Duration(1*time.Microsecond))
		//assert
		assert.NoError(t, err, "Did not expect error")
	})

	t.Run("step fail http", func(t *testing.T) {
		//arrange
		client := &abaputils.ClientMock{
			Body:  "dummy",
			Error: errors.New("dummy"),
		}
		//act
		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, client, time.Duration(1*time.Second), time.Duration(1*time.Microsecond))
		//assert
		assert.Error(t, err, "Must end with error")
	})

	t.Run("step fail no id", func(t *testing.T) {
		//arrange
		config := abapAddonAssemblyKitPublishTargetVectorOptions{}
		mc := abapbuild.NewMockClient()
		//act
		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, &mc, time.Duration(1*time.Second), time.Duration(1*time.Microsecond))
		//assert
		assert.Error(t, err, "Must end with error")
	})
}
