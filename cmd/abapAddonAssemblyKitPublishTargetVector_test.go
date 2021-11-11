package cmd

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

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
		mc.AddData(AAKaaSHead)
		mc.AddData(AAKaaSPublishProdPost)
		mc.AddData(AAKaaSGetTVPublishRunning)
		mc.AddData(AAKaaSGetTVPublishProdSuccess)

		//act
		err := runAbapAddonAssemblyKitPublishTargetVector(&config, nil, &mc, time.Duration(1*time.Second), time.Duration(1*time.Microsecond))
		//assert
		assert.NoError(t, err, "Did not expect error")
	})

	t.Run("step success test", func(t *testing.T) {
		//arrange
		config.TargetVectorScope = "T"
		mc := abapbuild.NewMockClient()
		mc.AddData(AAKaaSHead)
		mc.AddData(AAKaaSPublishTestPost)
		mc.AddData(AAKaaSGetTVPublishRunning)
		mc.AddData(AAKaaSGetTVPublishTestSuccess)
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

/************
 Mock Client
************/

var AAKaaSHead = abapbuild.MockData{
	Method: `HEAD`,
	Url:    `/odata/aas_ocs_package`,
	Body: `<?xml version="1.0"?>
	<HTTP_BODY/>`,
	StatusCode: 200,
	Header:     http.Header{"x-csrf-token": {"HRfJP0OhB9C9mHs2RRqUzw=="}},
}

var AAKaaSPublishTestPost = abapbuild.MockData{
	Method: `POST`,
	Url:    `/odata/aas_ocs_package/PublishTargetVector?Id='W7Q00207512600000353'&Scope='T'`,
	Body: `{
		"d": {
			"Id": "W7Q00207512600000353",
			"Vendor": "0000029218",
			"ProductName": "/DRNMSPC/PRD01",
			"ProductVersion": "0001",
			"SpsLevel": "0000",
			"PatchLevel": "0000",
			"Status": "G",
			"PublishStatus": "R"
		}
	}`,
	StatusCode: 200,
}

var AAKaaSPublishProdPost = abapbuild.MockData{
	Method: `POST`,
	Url:    `/odata/aas_ocs_package/PublishTargetVector?Id='W7Q00207512600000353'&Scope='P'`,
	Body: `{
		"d": {
			"Id": "W7Q00207512600000353",
			"Vendor": "0000029218",
			"ProductName": "/DRNMSPC/PRD01",
			"ProductVersion": "0001",
			"SpsLevel": "0000",
			"PatchLevel": "0000",
			"Status": "G",
			"PublishStatus": "R"
		}
	}`,
	StatusCode: 200,
}

var AAKaaSGetTVPublishRunning = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/TargetVectorSet('W7Q00207512600000353')`,
	Body: `{
		"d": {
			"Id": "W7Q00207512600000353",
			"Vendor": "0000029218",
			"ProductName": "/DRNMSPC/PRD01",
			"ProductVersion": "0001",
			"SpsLevel": "0000",
			"PatchLevel": "0000",
			"Status": "G",
			"PublishStatus": "R"
		}
	}`,
	StatusCode: 200,
}

var AAKaaSGetTVPublishTestSuccess = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/TargetVectorSet('W7Q00207512600000353')`,
	Body: `{
		"d": {
			"Id": "W7Q00207512600000353",
			"Vendor": "0000029218",
			"ProductName": "/DRNMSPC/PRD01",
			"ProductVersion": "0001",
			"SpsLevel": "0000",
			"PatchLevel": "0000",
			"Status": "T",
			"PublishStatus": "S"
		}
	}`,
	StatusCode: 200,
}

var AAKaaSGetTVPublishProdSuccess = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/TargetVectorSet('W7Q00207512600000353')`,
	Body: `{
		"d": {
			"Id": "W7Q00207512600000353",
			"Vendor": "0000029218",
			"ProductName": "/DRNMSPC/PRD01",
			"ProductVersion": "0001",
			"SpsLevel": "0000",
			"PatchLevel": "0000",
			"Status": "P",
			"PublishStatus": "S"
		}
	}`,
	StatusCode: 200,
}

var templateMockData = abapbuild.MockData{
	Method:     `GET`,
	Url:        ``,
	Body:       ``,
	StatusCode: 200,
}
