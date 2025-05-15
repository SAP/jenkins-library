//go:build !release
// +build !release

/*
** The Test Data is partly re-used by the steps in cmd folder, thus we Export them but remove the file from release build
 */

package aakaas

import abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"

const statusFilterCVEscaped string = "DeliveryStatus+eq+%27R%27"
const statusFilterPVEscaped string = "DeliveryStatus+eq+%27T%27+or+DeliveryStatus+eq+%27P%27"

var ResponseCheckPV = `{
    "d": {
        "Name": "/DRNMSPC/PRD01",
        "Version": "0003",
        "SpsLevel": "0002",
        "PatchLevel": "0001",
        "Vendor": "",
        "VendorType": ""
    }
}`

var ResponseCheckCVs = `{
    "d": {
        "Name": "/DRNMSPC/COMP01",
        "Version": "0001",
        "SpLevel": "0002",
        "PatchLevel": "0003",
        "Vendor": "",
        "VendorType": ""
    }
}`

var ResponseCheck = `{
	"d": {
		"ProductName": "/DRNMSPC/PRD01",
		"SemProductVersion": "2.0.0",
		"ProductVersion": "0002",
		"SpsLevel": "0000",
		"PatchLevel": "0000",
		"Vendor": "",
		"VendorType": "",
		"Content": {
			"results": [
				{
					"ProductName": "/DRNMSPC/PRD01",
					"SemProductVersion": "2.0.0",
					"ScName": "/DRNMSPC/COMP01",
					"SemScVersion": "2.0.0",
					"ScVersion": "0002",
					"SpLevel": "0000",
					"PatchLevel": "0000",
					"Vendor": "",
					"VendorType": ""
				},
				{
					"ProductName": "/DRNMSPC/PRD01",
					"SemProductVersion": "2.0.0",
					"ScName": "/DRNMSPC/COMP02",
					"SemScVersion": "1.0.0",
					"ScVersion": "0001",
					"SpLevel": "0000",
					"PatchLevel": "0000",
					"Vendor": "",
					"VendorType": ""
				}
			]
		}
	}
}`

var emptyResultBody = `{
    "d": {
        "results": []
    }
}`

var testDataAakaasCVGetReleaseExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechSpLevel+eq+%270000%27+and+TechPatchLevel+eq+%270000%27+and+%28+` + statusFilterCVEscaped + `+%29&%24format=json&%24orderby=TechRelease+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel%2CNamespace&%24top=1`,
	Body: `{
		"d": {
			"results": [
				{
					"Name": "DummyComp",
					"Version": "1.0.0",
					"TechRelease": "1",
					"TechSpLevel": "0000",
					"TechPatchLevel": "0000"
				}
			]
		}
	}`,
	StatusCode: 200,
}

var testDataAakaasCVGetReleaseNonExisting = abapbuild.MockData{
	Method:     `GET`,
	Url:        `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechSpLevel+eq+%270000%27+and+TechPatchLevel+eq+%270000%27+and+%28+` + statusFilterCVEscaped + `+%29&%24format=json&%24orderby=TechRelease+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel%2CNamespace&%24top=1`,
	Body:       emptyResultBody,
	StatusCode: 200,
}

var testDataAakaasCVGetSpLevelExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechRelease+eq+%271%27+and+TechPatchLevel+eq+%270000%27++and+%28+` + statusFilterCVEscaped + `+%29&%24format=json&%24orderby=TechSpLevel+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel%2CNamespace&%24top=1`,
	Body: `{
		"d": {
			"results": [
				{
					"Name": "DummyComp",
					"Version": "1.7.0",
					"TechRelease": "1",
					"TechSpLevel": "0007",
					"TechPatchLevel": "0000"
				}
			]
		}
	}`,
	StatusCode: 200,
}

var testDataAakaasCVGetSpLevelNonExisting = abapbuild.MockData{
	Method:     `GET`,
	Url:        `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechRelease+eq+%271%27+and+TechPatchLevel+eq+%270000%27++and+%28+` + statusFilterCVEscaped + `+%29&%24format=json&%24orderby=TechSpLevel+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel%2CNamespace&%24top=1`,
	Body:       emptyResultBody,
	StatusCode: 200,
}

var testDataAakaasCVGetPatchLevelExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechRelease+eq+%271%27+and+TechSpLevel+eq+%270003%27+and+%28+` + statusFilterCVEscaped + `+%29&%24format=json&%24orderby=TechPatchLevel+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel%2CNamespace&%24top=1`,
	Body: `{
		"d": {
			"results": [
				{
					"Name": "DummyComp",
					"Version": "1.3.46",
					"TechRelease": "1",
					"TechSpLevel": "0003",
					"TechPatchLevel": "0046"
				}
			]
		}
	}`,
	StatusCode: 200,
}

var testDataAakaasCVGetPatchLevelNonExisting = abapbuild.MockData{
	Method:     `GET`,
	Url:        `/odata/aas_ocs_package/xSSDAxC_Component_Version?%24filter=Name+eq+%27DummyComp%27+and+TechRelease+eq+%271%27+and+TechSpLevel+eq+%270003%27+and+%28+` + statusFilterCVEscaped + `+%29&%24format=json&%24orderby=TechPatchLevel+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel%2CNamespace&%24top=1`,
	Body:       emptyResultBody,
	StatusCode: 200,
}

var testDataAakaasPVGetReleaseExisting = abapbuild.MockData{
	Method: `GET`,
	Url:    `/odata/aas_ocs_package/xSSDAxC_Product_Version?%24filter=Name+eq+%27DummyProd%27+and+TechSpLevel+eq+%270000%27+and+TechPatchLevel+eq+%270000%27+and+%28+` + statusFilterPVEscaped + `+%29&%24format=json&%24orderby=TechRelease+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel%2CNamespace&%24top=1`,
	Body: `{
        "d": {
            "results": [
                {
                    "Name": "DummyProd",
                    "Version": "1.0.0",
                    "TechRelease": "0001",
                    "TechSpLevel": "0000",
                    "TechPatchLevel": "0000"
                }
            ]
        }
    }`,
	StatusCode: 200,
}

var testDataAakaasPVGetReleaseNonExisting = abapbuild.MockData{
	Method:     `GET`,
	Url:        `/odata/aas_ocs_package/xSSDAxC_Product_Version?%24filter=Name+eq+%27DummyProd%27+and+TechSpLevel+eq+%270000%27+and+TechPatchLevel+eq+%270000%27+and+%28+` + statusFilterPVEscaped + `+%29&%24format=json&%24orderby=TechRelease+desc&%24select=Name%2CVersion%2CTechRelease%2CTechSpLevel%2CTechPatchLevel%2CNamespace&%24top=1`,
	Body:       emptyResultBody,
	StatusCode: 200,
}
