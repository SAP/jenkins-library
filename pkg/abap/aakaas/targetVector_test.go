package aakaas

import (
	"net/http"
	"testing"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

func TestTargetVectorInitExisting(t *testing.T) {
	t.Run("ID is set", func(t *testing.T) {
		//arrange
		id := "dummyID"
		targetVector := new(TargetVector)
		//act
		targetVector.InitExisting(id)
		//assert
		assert.Equal(t, id, targetVector.ID)
	})
}

func TestTargetVectorInitNew(t *testing.T) {
	t.Run("Ensure values not initial", func(t *testing.T) {
		//arrange
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
		targetVector := new(TargetVector)
		//act
		err := targetVector.InitNew(&addonDescriptor)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, "dummy", targetVector.ProductVersion)
	})
	t.Run("Fail if values initial", func(t *testing.T) {
		//arrange
		addonDescriptor := abaputils.AddonDescriptor{}
		targetVector := new(TargetVector)
		//act
		err := targetVector.InitNew(&addonDescriptor)
		//assert
		assert.Error(t, err)
	})
}

func TestTargetVectorGet(t *testing.T) {
	//arrange global
	targetVector := new(TargetVector)
	conn := new(abapbuild.Connector)

	t.Run("Ensure error if ID is initial", func(t *testing.T) {
		//arrange
		targetVector.ID = ""
		//act
		err := targetVector.GetTargetVector(conn)
		//assert
		assert.Error(t, err)
	})
	t.Run("Normal Get Test Success", func(t *testing.T) {
		//arrange
		targetVector.ID = "W7Q00207512600000353"
		mc := abapbuild.NewMockClient()
		mc.AddData(AAKaaSGetTVPublishTestSuccess)
		conn.Client = &mc
		//act
		err := targetVector.GetTargetVector(conn)
		//assert
		assert.NoError(t, err)
		assert.Equal(t, TargetVectorPublishStatusSuccess, targetVector.PublishStatus)
		assert.Equal(t, TargetVectorStatusTest, targetVector.Status)
	})
	t.Run("Error Get", func(t *testing.T) {
		//arrange
		targetVector.ID = "W7Q00207512600000353"
		mc := abapbuild.NewMockClient()
		conn.Client = &mc
		//act
		err := targetVector.GetTargetVector(conn)
		//assert
		assert.Error(t, err)
	})
}

func TestTargetVectorPollForStatus(t *testing.T) {
	//arrange global
	targetVector := new(TargetVector)
	conn := new(abapbuild.Connector)
	conn.MaxRuntimeInMinutes = time.Duration(1 * time.Second)
	conn.PollIntervalsInSeconds = time.Duration(50 * time.Microsecond)

	t.Run("Normal Poll", func(t *testing.T) {
		//arrange
		mc := abapbuild.NewMockClient()
		mc.AddData(AAKaaSGetTVPublishRunning)
		mc.AddData(AAKaaSGetTVPublishTestSuccess)
		conn.Client = &mc
		//act
		err := targetVector.PollForStatus(conn, TargetVectorStatusTest)
		//assert
		assert.NoError(t, err)
	})
}

/****************
 Mock Client Data
****************/

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
