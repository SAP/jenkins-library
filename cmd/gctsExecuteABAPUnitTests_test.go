package cmd

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverySuccess(t *testing.T) {

	config := gctsExecuteABAPUnitTestsOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("discovery successfull", func(t *testing.T) {

		httpClient := httpMockGcts{
			StatusCode: 200,
			Header:     map[string][]string{"x-csrf-token": []string{"ZegUEgfa50R7ZfGGxOtx2A=="}},
			ResponseBody: `
				<?xml version="1.0" encoding="utf-8"?>
				<app:service xmlns:app="http://www.w3.org/2007/app" xmlns:atom="http://www.w3.org/2005/Atom"/>
			`}

		header, err := discoverServer(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/adt/core/discovery?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check header", func(t *testing.T) {
				assert.Equal(t, http.Header{"x-csrf-token": []string{"ZegUEgfa50R7ZfGGxOtx2A=="}}, *header)
			})

		}

	})
}

func TestDiscoveryFailure(t *testing.T) {

	config := gctsExecuteABAPUnitTestsOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 403, ResponseBody: `
		<html>
		<head>
			<meta http-equiv="content-type" content="text/html; charset=windows-1252">
			<title>Service cannot be reached</title>
			<style>
				body {
					background: #ffffff;
					text-align: center;
					width: 100%;
					height: 100%;
					overflow: hidden;
				}
			</style>
		</head>

		<body>
			<div class="content">
				<div class="valigned">
					<p class="centerText"><span class="errorTextHeader"> 403 Forbidden </span></p>
				</div>
			</div>
			<div class="footer">
				<div class="footerLeft"></div>
					<div class="footerRight">
						<p class="bottomText"><span class="biggerBottomText">&copy;</span>2020 SAP SE, All rights reserved.</p>
					</div>
				</div>
		</body>
		</html>
		`}

		header, err := discoverServer(&config, &httpClient)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "discovery of the ABAP server failed: a http error occurred")
		})

		t.Run("check header", func(t *testing.T) {
			assert.Equal(t, (*http.Header)(nil), header)
		})

	})

	t.Run("discovery response is nil", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: ``}

		header, err := discoverServer(&config, &httpClient)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "discovery of the ABAP server failed: did not retrieve a HTTP response")
		})

		t.Run("check header", func(t *testing.T) {
			assert.Equal(t, (*http.Header)(nil), header)
		})

	})

	t.Run("discovery header is nil", func(t *testing.T) {

		httpClient := httpMockGcts{
			StatusCode: 200,
			Header:     nil,
			ResponseBody: `
				<?xml version="1.0" encoding="utf-8"?>
				<app:service xmlns:app="http://www.w3.org/2007/app" xmlns:atom="http://www.w3.org/2005/Atom"/>
			`}

		header, err := discoverServer(&config, &httpClient)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "discovery of the ABAP server failed: did not retrieve a HTTP response")
		})

		t.Run("check header", func(t *testing.T) {
			assert.Equal(t, (*http.Header)(nil), header)
		})

	})
}

func TestGetPackageListSuccess(t *testing.T) {

	config := gctsExecuteABAPUnitTestsOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("return multiple objects sucessfully", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `
		{
			"objects": [
				{
						"pgmid": "R3TR",
						"object": "ZCL_GCTS",
						"type": "CLAS",
						"description": "Class (ABAP Objects)"
				},
				{
						"pgmid": "R3TR",
						"object": "ZP_GCTS",
						"type": "DEVC",
						"description": "Package"
				},
				{
						"pgmid": "R3TR",
						"object": "ZP_PIPER",
						"type": "DEVC",
						"description": "Package"
				}
			]
		}
		`}

		objects, err := getPackageList(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/getObjects?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, []string{"ZP_GCTS", "ZP_PIPER"}, objects)
			})

		}

	})

	t.Run("no objects returned by http call", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `{}`}

		objects, err := getPackageList(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/getObjects?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, []string{}, objects)
			})
		}

	})
}

func TestGetPackageListFailure(t *testing.T) {

	config := gctsExecuteABAPUnitTestsOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("http error occured", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `
		{
			"exception": "No relation between system and repository"
		}
		`}

		_, err := getPackageList(&config, &httpClient)

		assert.EqualError(t, err, "getting repository object/package list failed: a http error occurred")
	})
}

func TestExecuteTestsForPackageSuccess(t *testing.T) {

	config := gctsExecuteABAPUnitTestsOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	header := make(http.Header)
	header.Add("Accept", "application/atomsvc+xml")
	header.Add("x-csrf-token", "ZegUEgfa50R7ZfGGxOtx2A==")
	header.Add("saml2", "disabled")

	t.Run("all unit tests were successfull", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `
		<?xml version="1.0" encoding="utf-8"?>
		<aunit:runResult xmlns:aunit="http://www.sap.com/adt/aunit">
				<program adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo" adtcore:type="CLAS/OC" adtcore:name="ZCL_GCTS_PIPER_DEMO" uriType="semantic" xmlns:adtcore="http://www.sap.com/adt/core">
						<testClasses>
								<testClass adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#type=CLAS%2FOCL;name=LTCL_MASTER" adtcore:type="CLAS/OL" adtcore:name="LTCL_MASTER" uriType="semantic" navigationUri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#type=CLAS%2FOCL;name=LTCL_MASTER" durationCategory="short" riskLevel="harmless">
										<testMethods>
												<testMethod adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#type=CLAS%2FOLD;name=LTCL_MASTER%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20CHECK" adtcore:type="CLAS/OLI" adtcore:name="CHECK" executionTime="0" uriType="semantic" navigationUri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#type=CLAS%2FOLD;name=LTCL_MASTER%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20CHECK" unit="s"/>
										</testMethods>
								</testClass>
						</testClasses>
				</program>
		</aunit:runResult>
		`}

		err := executeTestsForPackage(&config, &httpClient, header, "ZP_PIPER")

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/adt/abapunit/testruns?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})

	t.Run("no unit tests found", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `
		<?xml version="1.0" encoding="utf-8"?>
		<aunit:runResult xmlns:aunit="http://www.sap.com/adt/aunit">
				<alerts>
						<alert kind="noTestClasses" severity="tolerable">
								<title>The task definition does not refer to any test</title>
						</alert>
				</alerts>
		</aunit:runResult>
		`}

		err := executeTestsForPackage(&config, &httpClient, header, "ZP_NON_EXISTANT")

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/adt/abapunit/testruns?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "POST", httpClient.Method)
			})
		}
	})
}

func TestExecuteTestsForPackageFailure(t *testing.T) {

	config := gctsExecuteABAPUnitTestsOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
	}

	t.Run("some unit tests failed", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `
		<?xml version="1.0" encoding="utf-8"?>
		<aunit:runResult xmlns:aunit="http://www.sap.com/adt/aunit">
				<program adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo" adtcore:type="CLAS/OC" adtcore:name="ZCL_GCTS_PIPER_DEMO" uriType="semantic" xmlns:adtcore="http://www.sap.com/adt/core">
						<testClasses>
								<testClass adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#type=CLAS%2FOCL;name=LTCL_MASTER" adtcore:type="CLAS/OL" adtcore:name="LTCL_MASTER" uriType="semantic" navigationUri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#type=CLAS%2FOCL;name=LTCL_MASTER" durationCategory="short" riskLevel="harmless">
										<testMethods>
												<testMethod adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#type=CLAS%2FOLD;name=LTCL_MASTER%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20CHECK" adtcore:type="CLAS/OLI" adtcore:name="CHECK" executionTime="0" uriType="semantic" navigationUri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#type=CLAS%2FOLD;name=LTCL_MASTER%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20CHECK" unit="s">
														<alerts>
																<alert kind="failedAssertion" severity="critical">
																		<title>Critical Assertion Error: 'Check: ASSERT_EQUALS'</title>
																		<details>
																				<detail text="Different Values:">
																						<details>
																								<detail text="Expected [Hello] Actual [Hello2]"/>
																						</details>
																				</detail>
																				<detail text="Test 'LTCL_MASTER-&gt;CHECK' in Main Program 'ZCL_GCTS_PIPER_DEMO===========CP'."/>
																		</details>
																		<stack>
																				<stackEntry adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo/includes/testclasses#start=21,0" adtcore:type="CLAS/OCN/testclasses" adtcore:name="ZCL_GCTS_PIPER_DEMO" adtcore:description="Include: &lt;ZCL_GCTS_PIPER_DEMO===========CCAU&gt; Line: &lt;21&gt; (CHECK)"/>
																		</stack>
																</alert>
														</alerts>
												</testMethod>
										</testMethods>
								</testClass>
						</testClasses>
				</program>
		</aunit:runResult>
		`}

		header := make(http.Header)
		header.Add("Accept", "application/atomsvc+xml")
		header.Add("x-csrf-token", "ZegUEgfa50R7ZfGGxOtx2A==")
		header.Add("saml2", "disabled")

		err := executeTestsForPackage(&config, &httpClient, header, "ZP_PIPER")

		assert.EqualError(t, err, "some unit tests failed")
	})

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 403, ResponseBody: `
		CSRF token validation failed
		`}

		header := make(http.Header)
		header.Add("Accept", "application/atomsvc+xml")
		header.Add("x-csrf-token", "ZegUEgfa50R7ZfGGxOtx2A==")
		header.Add("saml2", "disabled")

		err := executeTestsForPackage(&config, &httpClient, header, "ZP_PIPER")

		assert.EqualError(t, err, "execution of unit tests failed: a http error occurred")
	})
}
