package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractErrorStack(t *testing.T) {
	t.Run("Valid Json String", func(t *testing.T) {
		//act
		errorString := extractErrorStackFromJsonData([]byte(validResponse))
		//assert
		assert.Equal(t, `[0] Data Provider: SOME_PACKAGE and 0 other packages do not exist in system XXX
[1] Reading SCs of Packages failed
`, errorString)
	})

	t.Run("No Json String", func(t *testing.T) {
		//arrange
		var noJson = "ERROR 404: Unauthorized"
		//act
		errorString := extractErrorStackFromJsonData([]byte(noJson))
		//assert
		assert.Equal(t, noJson, errorString)
	})

	t.Run("step by step", func(t *testing.T) {
		my_error := new(GW_error)
		err := my_error.FromJson([]byte(validResponse))
		assert.NoError(t, err)

		assert.Equal(t, "/IWBEP/CM_MGW_RT/022", my_error.Error.Code)
		assert.Equal(t, "en", my_error.Error.Message.Lang)
		assert.Equal(t, "Data Provider: SOME_PACKAGE and 0 other packages do not exist in system XXX", my_error.Error.Message.Value)
		assert.Equal(t, "BC-UPG-ADDON", my_error.Error.Innererror.Application.Component_id)
		assert.Equal(t, "/BUILD/", my_error.Error.Innererror.Application.Service_namespace)
		assert.Equal(t, "CORE_SRV", my_error.Error.Innererror.Application.Service_id)
		assert.Equal(t, "0001", my_error.Error.Innererror.Application.Service_version)
		assert.Equal(t, "1801B32D512B00C0E0066215B8D723B5", my_error.Error.Innererror.Transactionid)
		assert.Equal(t, "", my_error.Error.Innererror.Timestamp)
		assert.Equal(t, "", my_error.Error.Innererror.Error_Resolution.SAP_Transaction)
		assert.Equal(t, "See SAP Note 1797736 for error analysis (https://service.sap.com/sap/support/notes/1797736)", my_error.Error.Innererror.Error_Resolution.SAP_Note)
		assert.Equal(t, 3, len(my_error.Error.Innererror.Errordetails))
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[0].ContentID)
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[0].Propertyref)
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[0].Target)
		assert.Equal(t, "error", my_error.Error.Innererror.Errordetails[0].Severity)
		assert.Equal(t, false, my_error.Error.Innererror.Errordetails[0].Transition)
		assert.Equal(t, "/BUILD/CX_EXTERNAL", my_error.Error.Innererror.Errordetails[0].Code)
		assert.Equal(t, "Data Provider: SOME_PACKAGE and 0 other packages do not exist in system XXX", my_error.Error.Innererror.Errordetails[0].Message)
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[1].ContentID)
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[1].Propertyref)
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[1].Target)
		assert.Equal(t, "error", my_error.Error.Innererror.Errordetails[1].Severity)
		assert.Equal(t, false, my_error.Error.Innererror.Errordetails[1].Transition)
		assert.Equal(t, "/BUILD/CX_BUILD", my_error.Error.Innererror.Errordetails[1].Code)
		assert.Equal(t, "Reading SCs of Packages failed", my_error.Error.Innererror.Errordetails[1].Message)
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[2].ContentID)
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[2].Propertyref)
		assert.Equal(t, "", my_error.Error.Innererror.Errordetails[2].Target)
		assert.Equal(t, "error", my_error.Error.Innererror.Errordetails[2].Severity)
		assert.Equal(t, false, my_error.Error.Innererror.Errordetails[2].Transition)
		assert.Equal(t, "/IWBEP/CX_MGW_BUSI_EXCEPTION", my_error.Error.Innererror.Errordetails[2].Code)
		assert.Equal(t, "Reading SCs of Packages failed", my_error.Error.Innererror.Errordetails[2].Message)
	})
}

var validResponse = `{
    "error": {
        "code": "/IWBEP/CM_MGW_RT/022",
        "message": {
            "lang": "en",
            "value": "Data Provider: SOME_PACKAGE and 0 other packages do not exist in system XXX"
        },
        "innererror": {
            "application": {
                "component_id": "BC-UPG-ADDON",
                "service_namespace": "/BUILD/",
                "service_id": "CORE_SRV",
                "service_version": "0001"
            },
            "transactionid": "1801B32D512B00C0E0066215B8D723B5",
            "timestamp": "",
            "Error_Resolution": {
                "SAP_Transaction": "",
                "SAP_Note": "See SAP Note 1797736 for error analysis (https://service.sap.com/sap/support/notes/1797736)"
            },
            "errordetails": [{
                    "ContentID": "",
                    "code": "/BUILD/CX_EXTERNAL",
                    "message": "Data Provider: SOME_PACKAGE and 0 other packages do not exist in system XXX",
                    "propertyref": "",
                    "severity": "error",
                    "transition": false,
                    "target": ""
                }, {
                    "ContentID": "",
                    "code": "/BUILD/CX_BUILD",
                    "message": "Reading SCs of Packages failed",
                    "propertyref": "",
                    "severity": "error",
                    "transition": false,
                    "target": ""
                }, {
                    "ContentID": "",
                    "code": "/IWBEP/CX_MGW_BUSI_EXCEPTION",
                    "message": "Reading SCs of Packages failed",
                    "propertyref": "",
                    "severity": "error",
                    "transition": false,
                    "target": ""
                }
            ]
        }
    }
}`
