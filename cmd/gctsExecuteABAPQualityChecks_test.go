//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDiscoverServerSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Scope:      "localChangedObjects",
		Commit:     "0123456789abcdefghijkl",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("discover server successful", func(t *testing.T) {

		httpClient := httpMockGcts{
			StatusCode: 200,
			Header:     map[string][]string{"x-csrf-token": {"ZegUEgfa50R7ZfGGxOtx2A=="}},
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

func TestDiscoverServerFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost3.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Scope:      "localChangedObjects",
		Commit:     "0123456789abcdefghijkl",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 403, ResponseBody: `
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

		httpClient := httpMockGctsT{StatusCode: 200, ResponseBody: ``}

		header, err := discoverServer(&config, &httpClient)

		t.Run("check error", func(t *testing.T) {
			assert.EqualError(t, err, "discovery of the ABAP server failed: did not retrieve a HTTP response")
		})

		t.Run("check header", func(t *testing.T) {
			assert.Equal(t, (*http.Header)(nil), header)
		})

	})

	t.Run("discovery header is nil", func(t *testing.T) {

		httpClient := httpMockGctsT{
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

func TestGetLocalObjectsSuccess(t *testing.T) {

	t.Run("return multiple objects successfully", func(t *testing.T) {

		config := gctsExecuteABAPQualityChecksOptions{
			Host:       "http://testHost.com:50000",
			Client:     "000",
			Repository: "testRepo",
			Username:   "testUser",
			Password:   "testPassword",
			Commit:     "0123456789abcdefghijkl",
			Scope:      "localChangedObjects",
			Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
		}

		httpClient := httpMockGctsT{StatusCode: 200}

		object1 := repoObject{

			Object: "ZCL_GCTS",
			Type:   "CLAS",
		}
		object2 := repoObject{

			Object: "ZP_GCTS",
			Type:   "DEVC",
		}
		object3 := repoObject{

			Object: "ZIF_GCTS_API",
			Type:   "INTF",
		}
		var repoObjects []repoObject
		repoObjects = append(repoObjects, object1)
		repoObjects = append(repoObjects, object2)
		repoObjects = append(repoObjects, object3)

		objects, err := getLocalObjects(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/compareCommits?fromCommit=xyz987654321&toCommit=0123456789abcdefghijkl&sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})

		}

	})

	t.Run("no objects returned by http call", func(t *testing.T) {

		config := gctsExecuteABAPQualityChecksOptions{
			Host:       "http://testHost.com:50000",
			Client:     "000",
			Repository: "testRepo2",
			Username:   "testUser",
			Password:   "testPassword",
			Commit:     "0123456789abcdefghijkl",
			Scope:      "localChangedObjects",
			Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
		}

		var repoObjects []repoObject

		httpClient := httpMockGctsT{StatusCode: 200, ResponseBody: `{}`}

		objects, err := getLocalObjects(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2/compareCommits?fromCommit=xyz987654321&toCommit=0123456789abcdefghijkl&sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})
		}

	})
}

func TestGetLocalObjectsFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "localChangedObjects",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `
		{
			"exception": "No relation between system and repository"
		}
		`}

		_, err := getLocalObjects(&config, &httpClient)

		assert.EqualError(t, err, "get local changed objects failed: get history failed: a http error occurred")
	})
}

func TestGetRemoteObjectsSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "remoteChangedObjects",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("return multiple objects successfully", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		object1 := repoObject{

			Object: "ZCL_GCTS",
			Type:   "CLAS",
		}
		object2 := repoObject{

			Object: "ZP_GCTS",
			Type:   "DEVC",
		}
		object3 := repoObject{

			Object: "ZIF_GCTS_API",
			Type:   "INTF",
		}
		var repoObjects []repoObject
		repoObjects = append(repoObjects, object1)
		repoObjects = append(repoObjects, object2)
		repoObjects = append(repoObjects, object3)

		objects, err := getRemoteObjects(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/compareCommits?fromCommit=7845abaujztrw785&toCommit=0123456789abcdefghijkl&sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})

		}

	})

	t.Run("no objects returned by http call", func(t *testing.T) {

		config := gctsExecuteABAPQualityChecksOptions{
			Host:       "http://testHost.com:50000",
			Client:     "000",
			Repository: "testRepo2",
			Username:   "testUser",
			Password:   "testPassword",
			Commit:     "0123456789abcdefghijkl",
			Scope:      "remoteChangedObjects",
			Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
		}
		httpClient := httpMockGctsT{StatusCode: 200, ResponseBody: `{}`}
		var repoObjects []repoObject
		objects, err := getRemoteObjects(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2/compareCommits?fromCommit=7845abaujztrw785&toCommit=0123456789abcdefghijkl&sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})
		}

	})
}

func TestGetRemoteObjectsFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "remoteChangedObjects",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 500, ResponseBody: `
		{
			"exception": "No relation between system and repository"
		}
		`}

		_, err := getRemoteObjects(&config, &httpClient)

		assert.EqualError(t, err, "get remote changed objects failed: get repository history failed: a http error occurred")
	})
}

func TestGetLocalPackagesSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "localChangedPackages",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("return multiple objects successfully", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		object1 := repoObject{

			Object: "SGCTS",

			Type: "DEVC",
		}
		object2 := repoObject{

			Object: "SGCTS_2",
			Type:   "DEVC",
		}

		var repoObjects []repoObject
		repoObjects = append(repoObjects, object1)
		repoObjects = append(repoObjects, object2)

		objects, err := getLocalPackages(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/objects/INTF/ZIF_GCTS_API?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})

		}

	})

	t.Run("no objects returned by http call", func(t *testing.T) {

		config := gctsExecuteABAPQualityChecksOptions{
			Host:       "http://testHost.com:50000",
			Client:     "000",
			Repository: "testRepo2",
			Username:   "testUser",
			Password:   "testPassword",
			Commit:     "0123456789abcdefghijkl",
			Scope:      "localChangedObjects",
			Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
		}

		httpClient := httpMockGctsT{StatusCode: 200, ResponseBody: `{}`}
		var repoObjects []repoObject
		objects, err := getLocalPackages(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2/compareCommits?fromCommit=xyz987654321&toCommit=0123456789abcdefghijkl&sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})
		}

	})
}

func TestGetLocalPackagesFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "localChangedPackages",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}
	t.Run("http error occurred", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 500, ResponseBody: `
		{
			"exception": "No relation between system and repository"
		}
		`}

		_, err := getLocalPackages(&config, &httpClient)

		assert.EqualError(t, err, "get local changed objects failed: get history failed: a http error occurred")
	})
}

func TestGetRemotePackagesSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "remoteChangedPackages",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("return multiple objects successfully", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		object1 := repoObject{

			Object: "SGCTS",

			Type: "DEVC",
		}
		object2 := repoObject{

			Object: "SGCTS_2",
			Type:   "DEVC",
		}
		var repoObjects []repoObject
		repoObjects = append(repoObjects, object1)
		repoObjects = append(repoObjects, object2)

		objects, err := getRemotePackages(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/objects/INTF/ZIF_GCTS_API?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})

		}

	})

	t.Run("no objects returned by http call", func(t *testing.T) {

		config := gctsExecuteABAPQualityChecksOptions{
			Host:       "http://testHost.com:50000",
			Client:     "000",
			Repository: "testRepo2",
			Username:   "testUser",
			Password:   "testPassword",
			Commit:     "0123456789abcdefghijkl",
			Scope:      "remoteChangedPackages",
			Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
		}

		httpClient := httpMockGctsT{StatusCode: 200, ResponseBody: `{}`}
		var repoObjects []repoObject
		objects, err := getRemoteObjects(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2/compareCommits?fromCommit=7845abaujztrw785&toCommit=0123456789abcdefghijkl&sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})
		}

	})
}

func TestGetRemotePackagesFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "remoteChangedPackages",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}
	t.Run("http error occurred", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 500, ResponseBody: `
		{
			"exception": "No relation between system and repository"
		}
		`}

		_, err := getRemotePackages(&config, &httpClient)

		assert.EqualError(t, err, "get remote changed packages failed: get repository history failed: a http error occurred")
	})
}

func TestGetPackagesSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "packages",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("return multiple objects successfully", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		object1 := repoObject{

			Pgmid:       "R3TR",
			Object:      "ZP_GCTS",
			Type:        "DEVC",
			Description: "Package(ABAP Objects)",
		}

		var repoObjects []repoObject
		repoObjects = append(repoObjects, object1)

		objects, err := getPackages(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/objects?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})

		}

	})

	t.Run("no objects returned by http call", func(t *testing.T) {

		config := gctsExecuteABAPQualityChecksOptions{
			Host:       "http://testHost.com:50000",
			Client:     "000",
			Repository: "testRepo2",
			Username:   "testUser",
			Password:   "testPassword",
			Commit:     "0123456789abcdefghijkl",
			Scope:      "packages",
			Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
		}

		httpClient := httpMockGcts{StatusCode: 200, ResponseBody: `{}`}
		var repoObjects []repoObject
		objects, err := getPackages(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2/objects?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})
		}

	})
}

func TestGetPackagesFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo2",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "packages",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}
	t.Run("http error occurred", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 500, ResponseBody: `
		{
			"exception": "No relation between system and repository"
		}
		`}

		_, err := getPackages(&config, &httpClient)

		assert.EqualError(t, err, "get packages failed: could not get repository objects: a http error occurred")
	})
}

func TestGetRepositoryObjectsSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "repository",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("return multiple objects successfully", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		object1 := repoObject{
			Pgmid:       "R3TR",
			Object:      "ZCL_GCTS",
			Type:        "CLAS",
			Description: "Class (ABAP Objects)",
		}

		object3 := repoObject{
			Pgmid:       "R3TR",
			Object:      "ZIF_GCTS_API",
			Type:        "INTF",
			Description: "Interface (ABAP Objects)",
		}
		var repoObjects []repoObject
		repoObjects = append(repoObjects, object1)
		repoObjects = append(repoObjects, object3)

		objects, err := getRepositoryObjects(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/objects?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})

		}

	})

	t.Run("no objects returned by http call", func(t *testing.T) {

		config := gctsExecuteABAPQualityChecksOptions{
			Host:       "http://testHost.com:50000",
			Client:     "000",
			Repository: "testRepo2",
			Username:   "testUser",
			Password:   "testPassword",
			Commit:     "0123456789abcdefghijkl",
			Scope:      "repository",
			Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
		}

		httpClient := httpMockGctsT{StatusCode: 200, ResponseBody: `{}`}
		var repoObjects []repoObject
		objects, err := getRepositoryObjects(&config, &httpClient)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2/objects?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check package objects", func(t *testing.T) {
				assert.Equal(t, repoObjects, objects)
			})
		}

	})
}

func TestGetRepositoryObjectsFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:       "http://testHost.com:50000",
		Client:     "000",
		Repository: "testRepo",
		Username:   "testUser",
		Password:   "testPassword",
		Commit:     "0123456789abcdefghijkl",
		Scope:      "repository",
		Workspace:  "/var/jenkins_home/workspace/myFirstPipeline",
	}

	t.Run("http error occurred", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 500, ResponseBody: `
		{
			"exception": "No relation between system and repository"
		}
		`}

		_, err := getRepositoryObjects(&config, &httpClient)

		assert.EqualError(t, err, "could not get repository objects: a http error occurred")
	})
}

func TestExecuteAUnitTestSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:                 "http://testHost.com:50000",
		Client:               "000",
		Repository:           "testRepo",
		Username:             "testUser",
		Password:             "testPassword",
		Commit:               "0123456789abcdefghijkl",
		Scope:                "repository",
		Workspace:            "/var/jenkins_home/workspace/myFirstPipeline",
		AtcResultsFileName:   "ATCResults.xml",
		AUnitResultsFileName: "AUnitResults.xml",
	}

	t.Run("all unit tests were successful", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		object := repoObject{
			Pgmid:       "R3TR",
			Object:      "ZCL_GCTS",
			Type:        "CLAS",
			Description: "Clas Object",
		}

		var repoObjects []repoObject
		repoObjects = append(repoObjects, object)

		err := executeAUnitTest(&config, &httpClient, repoObjects)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

		}
	})

	t.Run("no unit tests found", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200, ResponseBody: `
		<?xml version="1.0" encoding="utf-8"?>
		<aunit:runResult xmlns:aunit="http://www.sap.com/adt/aunit">
				<alerts>
						<alert kind="noTestClasses" severity="tolerable">
								<title>The task definition does not refer to any test</title>
						</alert>
				</alerts>
		</aunit:runResult>
		`}
		object := repoObject{
			Pgmid:       "R3TR",
			Object:      "ZP_NON_EXISTANT",
			Type:        "CLAS",
			Description: "Clas Object",
		}

		var repoObjects []repoObject
		repoObjects = append(repoObjects, object)
		err := executeAUnitTest(&config, &httpClient, repoObjects)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

		}
	})
}

func TestExecuteAUnitTestFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:                 "http://testHost.com:50000",
		Client:               "000",
		Repository:           "testRepo",
		Username:             "testUser",
		Password:             "testPassword",
		Commit:               "0123456789abcdefghijkl",
		Scope:                "repository",
		Workspace:            "/var/jenkins_home/workspace/myFirstPipeline",
		AtcResultsFileName:   "ATCResults.xml",
		AUnitResultsFileName: "AUnitResults.xml",
	}

	var repoObjects []repoObject

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 403, ResponseBody: `
		CSRF token validation failed
		`}

		header := make(http.Header)
		header.Add("Accept", "application/atomsvc+xml")
		header.Add("x-csrf-token", "ZegUEgfa50R7ZfGGxOtx2A==")
		header.Add("saml2", "disabled")

		err := executeAUnitTest(&config, &httpClient, repoObjects)

		assert.EqualError(t, err, "execute of Aunit test has failed: run of unit tests failed: discovery of the ABAP server failed: a http error occurred")

	})
}

func TestExecuteATCCheckSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:               "http://testHost.com:50000",
		Client:             "000",
		Repository:         "testRepo",
		Username:           "testUser",
		Password:           "testPassword",
		Commit:             "0123456789abcdefghijkl",
		AtcVariant:         "DEFAULT_REMOTE_REF",
		Scope:              "repository",
		Workspace:          "/var/jenkins_home/workspace/myFirstPipeline",
		AtcResultsFileName: "ATCResults.xml",
	}

	header := make(http.Header)
	header.Add("Accept", "application/atomsvc+xml")
	header.Add("x-csrf-token", "ZegUEgfa50R7ZfGGxOtx2A==")
	header.Add("saml2", "disabled")

	t.Run("atc checks were found", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		object := repoObject{
			Pgmid:       "R3TR",
			Object:      "ZCL_GCTS",
			Type:        "CLAS",
			Description: "Clas Object",
		}

		var repoObjects []repoObject
		repoObjects = append(repoObjects, object)

		err := executeATCCheck(&config, &httpClient, repoObjects)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

		}
	})

	t.Run("no ATC Checks were found", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		config := gctsExecuteABAPQualityChecksOptions{
			Host:               "http://testHost.com:50000",
			Client:             "000",
			Repository:         "testRepo",
			Username:           "testUser",
			Password:           "testPassword",
			Commit:             "0123456789abcdefghijkl",
			AtcVariant:         "CUSTOM_REMOTE_REF",
			Scope:              "repository",
			Workspace:          "/var/jenkins_home/workspace/myFirstPipeline",
			AtcResultsFileName: "ATCResults.xml",
		}

		object := repoObject{
			Pgmid:       "R3TR",
			Object:      "ZP_NON_EXISTANT",
			Type:        "CLAS",
			Description: "Clas Object",
		}

		var repoObjects []repoObject
		repoObjects = append(repoObjects, object)
		err := executeATCCheck(&config, &httpClient, repoObjects)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/adt/atc/worklists/3E3D0764F95Z01ABCDHEF9C9F6B5C14P?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

		}
	})
}

func TestExecuteATCCheckFailure(t *testing.T) {

	object := repoObject{
		Pgmid:       "R3TR",
		Object:      "ZP_PIPER",
		Type:        "CLAS",
		Description: "Clas Object",
	}

	var repoObjects []repoObject
	repoObjects = append(repoObjects, object)

	config := gctsExecuteABAPQualityChecksOptions{
		Host:               "http://testHost.com:50000",
		Client:             "000",
		Repository:         "testRepo",
		Username:           "testUser",
		Password:           "testPassword",
		Commit:             "0123456789abcdefghijkl",
		Scope:              "repository",
		Workspace:          "/var/jenkins_home/workspace/myFirstPipeline",
		AtcResultsFileName: "ATCResults.xml",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGcts{StatusCode: 403, ResponseBody: `
		CSRF token validation failed
		`}

		header := make(http.Header)
		header.Add("Accept", "application/atomsvc+xml")
		header.Add("x-csrf-token", "ZegUEgfa50R7ZfGGxOtx2A==")
		header.Add("saml2", "disabled")

		err := executeATCCheck(&config, &httpClient, repoObjects)

		assert.EqualError(t, err, "execution of ATC Checks failed: get worklist failed: discovery of the ABAP server failed: a http error occurred")

	})
}

func TestParseAUnitResultSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:                 "http://testHost.com:50000",
		Client:               "000",
		Repository:           "testRepo",
		Username:             "testUser",
		Password:             "testPassword",
		Commit:               "0123456789abcdefghijkl",
		Scope:                "repository",
		Workspace:            "/var/jenkins_home/workspace/myFirstPipeline",
		AUnitResultsFileName: "AUnitResults.xml",
	}

	t.Run("unit test is successful", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		var xmlBody = []byte(`
		<?xml version="1.0" encoding="utf-8"?>
		<aunit:runResult xmlns:aunit="http://www.sap.com/adt/aunit">
				<program adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts" adtcore:type="CLAS/OC" adtcore:name="ZCL_GCTS" uriType="semantic" xmlns:adtcore="http://www.sap.com/adt/core">
						<testClasses>
								<testClass adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts/includes/testclasses#type=CLAS%2FOCL;name=LTCL_MASTER" adtcore:type="CLAS/OL" adtcore:name="LTCL_MASTER" uriType="semantic" navigationUri="/sap/bc/adt/oo/classes/zcl_gcts/includes/testclasses#type=CLAS%2FOCL;name=LTCL_MASTER" durationCategory="short" riskLevel="harmless">
										<testMethods>
												<testMethod adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts/includes/testclasses#type=CLAS%2FOLD;name=LTCL_MASTER%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20CHECK" adtcore:type="CLAS/OLI" adtcore:name="CHECK" executionTime="0" uriType="semantic" navigationUri="/sap/bc/adt/oo/classes/zcl_gcts/includes/testclasses#type=CLAS%2FOLD;name=LTCL_MASTER%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20CHECK" unit="s"/>
										</testMethods>
								</testClass>
						</testClasses>
				</program>
		</aunit:runResult>
		`)

		var resp *runResult
		xml.Unmarshal(xmlBody, &resp)

		parsedRes, err := parseUnitResult(&config, &httpClient, resp)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check file name", func(t *testing.T) {
				assert.Equal(t, "/var/jenkins_home/workspace/myFirstPipeline//objects/CLAS/ZCL_GCTS/CINC ZCL_GCTS======================CCAU.abap", parsedRes.File[0].Name)
			})

		}
	})

	t.Run("unit test failed", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		var xmlBody = []byte(`<?xml version="1.0" encoding="utf-8"?>
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
		</aunit:runResult>`)

		var resp *runResult
		xml.Unmarshal(xmlBody, &resp)

		parsedRes, err := parseUnitResult(&config, &httpClient, resp)

		if assert.NoError(t, err) {

			t.Run("check url", func(t *testing.T) {
				assert.Equal(t, "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo?sap-client=000", httpClient.URL)
			})

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "GET", httpClient.Method)
			})

			t.Run("check file name", func(t *testing.T) {
				assert.Equal(t, "/var/jenkins_home/workspace/myFirstPipeline//objects/CLAS/ZCL_GCTS_PIPER_DEMO/CINC ZCL_GCTS_PIPER_DEMO===========CCAU.abap", parsedRes.File[0].Name)
			})

			t.Run("check line number", func(t *testing.T) {
				assert.Equal(t, "21", parsedRes.File[0].Error[0].Line)
			})

			t.Run("check severity", func(t *testing.T) {
				assert.Equal(t, "error", parsedRes.File[0].Error[0].Severity)
			})

			t.Run("check source", func(t *testing.T) {
				assert.Equal(t, "LTCL_MASTER/CHECK", parsedRes.File[0].Error[0].Source)
			})

			t.Run("check message", func(t *testing.T) {
				assert.Equal(t, " Different Values: Expected [Hello] Actual [Hello2] Test 'LTCL_MASTER->CHECK' in Main Program 'ZCL_GCTS_PIPER_DEMO===========CP'.", parsedRes.File[0].Error[0].Message)
			})

		}
	})

}

func TestParseAUnitResultFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:                 "http://testHost.com:50000",
		Client:               "000",
		Repository:           "testRepo",
		Username:             "testUser",
		Password:             "testPassword",
		Commit:               "0123456789abcdefghijkl",
		AtcVariant:           "DEFAULT_REMOTE_REF",
		Scope:                "repository",
		Workspace:            "/var/jenkins_home/workspace/myFirstPipeline",
		AUnitResultsFileName: "AUnitResults.xml",
	}

	t.Run("parser fails", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 403}

		var xmlBody = []byte(`<?xml version="1.0" encoding="utf-8"?>
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
		</aunit:runResult>`)

		var resp *runResult
		xml.Unmarshal(xmlBody, &resp)

		parsedRes, err := parseUnitResult(&config, &httpClient, resp)

		if assert.Error(t, err) {

			t.Run("check method", func(t *testing.T) {
				assert.Equal(t, "parse AUnit Result failed: get file name has failed: could not check readable source format: could not get repository layout: a http error occurred", err.Error())
			})

			assert.NotEmpty(t, parsedRes, "results are not empty")

		}
	})

}

func TestParseATCCheckResultSuccess(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:               "http://testHost.com:50000",
		Client:             "000",
		Repository:         "testRepo",
		Username:           "testUser",
		Password:           "testPassword",
		Commit:             "0123456789abcdefghijkl",
		AtcVariant:         "DEFAULT_REMOTE_REF",
		Scope:              "repository",
		Workspace:          "/var/jenkins_home/workspace/myFirstPipeline",
		AtcResultsFileName: "ATCResults.xml",
	}

	t.Run("atc found", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 200}

		var xmlBody = []byte(`<?xml version="1.0" encoding="utf-8"?>
	<atcworklist:worklist atcworklist:id="248A076493C01EEC8FA9CEAED527BD53" atcworklist:timestamp="2021-11-04T09:08:18Z" atcworklist:usedObjectSet="99999999999999999999999999999999" atcworklist:objectSetIsComplete="true" xmlns:atcworklist="http://www.sap.com/adt/atc/worklist">
		<atcworklist:objectSets>
			<atcworklist:objectSet atcworklist:name="00000000000000000000000000000000" atcworklist:title="All Objects" atcworklist:kind="ALL"/>
			<atcworklist:objectSet atcworklist:name="99999999999999999999999999999999" atcworklist:title="Last Check Run" atcworklist:kind="LAST_RUN"/>
		</atcworklist:objectSets>
		<atcworklist:objects>
			<atcobject:object adtcore:uri="/sap/bc/adt/atc/objects/R3TR/CLAS/ZCL_GCTS" adtcore:type="CLAS" adtcore:name="ZCL_GCTS" adtcore:packageName="ZPL_GCTS" atcobject:author="testUser" xmlns:atcobject="http://www.sap.com/adt/atc/object" xmlns:adtcore="http://www.sap.com/adt/core">
				<atcobject:findings>
					<atcfinding:finding adtcore:uri="/sap/bc/adt/atc/findings/itemid/248A076493C01EEC8FA9D609860AFD93/index/233439" atcfinding:location="/sap/bc/adt/oo/classes/zcl_gcts/source/main#type=CLAS%2FOSI;name=ZCL_GCTS;start=20" atcfinding:processor="testUser" atcfinding:lastChangedBy="testUser" atcfinding:priority="1" atcfinding:checkId="78A08159CD2A822100535FBEB655BDB8" atcfinding:checkTitle="Package Check (Remote-enabled)" atcfinding:messageId="USEM" atcfinding:messageTitle="Package Violation - Error" atcfinding:exemptionApproval="" atcfinding:exemptionKind="" atcfinding:checksum="553596936" atcfinding:quickfixInfo="atc:248A076493C01EEC8FA9D609860AFD93,233439" xmlns:atcfinding="http://www.sap.com/adt/atc/finding">
						<atom:link href="/sap/bc/adt/documentation/atc/documents/itemid/248A076493C01EEC8FA9D609860AFD93/index/233439" rel="http://www.sap.com/adt/relations/documentation" type="text/html" xmlns:atom="http://www.w3.org/2005/Atom"/>
						<atcfinding:quickfixes atcfinding:manual="false" atcfinding:automatic="false" atcfinding:pseudo="false"/>
					</atcfinding:finding>
				</atcobject:findings>
			</atcobject:object>
		</atcworklist:objects>
	</atcworklist:worklist>`)

		var resp *worklist

		xml.Unmarshal(xmlBody, &resp)

		parsedRes, err := parseATCCheckResult(&config, &httpClient, resp)

		if assert.NoError(t, err) {

			t.Run("check file name", func(t *testing.T) {
				assert.Equal(t, "/var/jenkins_home/workspace/myFirstPipeline//objects/CLAS/ZCL_GCTS/CPRI ZCL_GCTS.abap", parsedRes.File[0].Name)
			})

			t.Run("check line number", func(t *testing.T) {
				assert.Equal(t, "20", parsedRes.File[0].Error[0].Line)
			})

			t.Run("check severity", func(t *testing.T) {
				assert.Equal(t, "error", parsedRes.File[0].Error[0].Severity)
			})

			t.Run("check source", func(t *testing.T) {
				assert.Equal(t, "ZCL_GCTS", parsedRes.File[0].Error[0].Source)
			})

			t.Run("check message", func(t *testing.T) {
				assert.Equal(t, "Package Check (Remote-enabled) Package Violation - Error", parsedRes.File[0].Error[0].Message)
			})

			assert.NotEmpty(t, parsedRes, "results are not empty")

		}
	})

	t.Run("no ATC Checks were found", func(t *testing.T) {

		config := gctsExecuteABAPQualityChecksOptions{
			Host:               "http://testHost.com:50000",
			Client:             "000",
			Repository:         "testRepo",
			Username:           "testUser",
			Password:           "testPassword",
			Commit:             "0123456789abcdefghijkl",
			AtcVariant:         "DEFAULT_REMOTE_REF",
			Scope:              "repository",
			Workspace:          "/var/jenkins_home/workspace/myFirstPipeline",
			AtcResultsFileName: "ATCResults.xml",
		}

		httpClient := httpMockGctsT{StatusCode: 200}

		var xmlBody = []byte(`<?xml version="1.0" encoding="utf-8"?>
		<atcworklist:worklist atcworklist:id="248A076493C01EEC8FA9CEAED527BD53" atcworklist:timestamp="2021-11-08T08:44:40Z" atcworklist:usedObjectSet="99999999999999999999999999999999" atcworklist:objectSetIsComplete="true" xmlns:atcworklist="http://www.sap.com/adt/atc/worklist">
    		<atcworklist:objectSets>
        		<atcworklist:objectSet atcworklist:name="00000000000000000000000000000000" atcworklist:title="All Objects" atcworklist:kind="ALL"/>
        		<atcworklist:objectSet atcworklist:name="99999999999999999999999999999999" atcworklist:title="Last Check Run" atcworklist:kind="LAST_RUN"/>
    		</atcworklist:objectSets>
    		<atcworklist:objects/>
		</atcworklist:worklist>`)
		var resp *worklist
		xml.Unmarshal(xmlBody, &resp)
		parsedRes, err := parseATCCheckResult(&config, &httpClient, resp)

		if assert.NoError(t, err) {

			assert.Equal(t, parsedRes.Version, "1.0")

		}
	})

}

func TestParseATCCheckResultFailure(t *testing.T) {

	config := gctsExecuteABAPQualityChecksOptions{
		Host:               "http://testHost.com:50000",
		Client:             "000",
		Repository:         "testRepo",
		Username:           "testUser",
		Password:           "testPassword",
		Commit:             "0123456789abcdefghijkl",
		AtcVariant:         "DEFAULT_REMOTE_REF",
		Scope:              "repsoitory",
		Workspace:          "/var/jenkins_home/workspace/myFirstPipeline",
		AtcResultsFileName: "ATCResults.xml",
	}

	t.Run("a http error occurred", func(t *testing.T) {

		httpClient := httpMockGctsT{StatusCode: 403}

		var xmlBody = []byte(`<?xml version="1.0" encoding="utf-8"?>
	<atcworklist:worklist atcworklist:id="248A076493C01EEC8FA9CEAED527BD53" atcworklist:timestamp="2021-11-04T09:08:18Z" atcworklist:usedObjectSet="99999999999999999999999999999999" atcworklist:objectSetIsComplete="true" xmlns:atcworklist="http://www.sap.com/adt/atc/worklist">
		<atcworklist:objectSets>
			<atcworklist:objectSet atcworklist:name="00000000000000000000000000000000" atcworklist:title="All Objects" atcworklist:kind="ALL"/>
			<atcworklist:objectSet atcworklist:name="99999999999999999999999999999999" atcworklist:title="Last Check Run" atcworklist:kind="LAST_RUN"/>
		</atcworklist:objectSets>
		<atcworklist:objects>
			<atcobject:object adtcore:uri="/sap/bc/adt/atc/objects/R3TR/CLAS/ZCL_GCTS" adtcore:type="CLAS" adtcore:name="ZCL_GCTS" adtcore:packageName="ZPL_GCTS" atcobject:author="testUser" xmlns:atcobject="http://www.sap.com/adt/atc/object" xmlns:adtcore="http://www.sap.com/adt/core">
				<atcobject:findings>
					<atcfinding:finding adtcore:uri="/sap/bc/adt/atc/findings/itemid/248A076493C01EEC8FA9D609860AFD93/index/233439" atcfinding:location="/sap/bc/adt/oo/classes/zcl_gcts/source/main#type=CLAS%2FOSI;name=ZCL_GCTS;start=20" atcfinding:processor="testUser" atcfinding:lastChangedBy="testUser" atcfinding:priority="1" atcfinding:checkId="78A08159CD2A822100535FBEB655BDB8" atcfinding:checkTitle="Package Check (Remote-enabled)" atcfinding:messageId="USEM" atcfinding:messageTitle="Package Violation - Error" atcfinding:exemptionApproval="" atcfinding:exemptionKind="" atcfinding:checksum="553596936" atcfinding:quickfixInfo="atc:248A076493C01EEC8FA9D609860AFD93,233439" xmlns:atcfinding="http://www.sap.com/adt/atc/finding">
						<atom:link href="/sap/bc/adt/documentation/atc/documents/itemid/248A076493C01EEC8FA9D609860AFD93/index/233439" rel="http://www.sap.com/adt/relations/documentation" type="text/html" xmlns:atom="http://www.w3.org/2005/Atom"/>
						<atcfinding:quickfixes atcfinding:manual="false" atcfinding:automatic="false" atcfinding:pseudo="false"/>
					</atcfinding:finding>
				</atcobject:findings>
			</atcobject:object>
		</atcworklist:objects>
	</atcworklist:worklist>`)

		var resp *worklist
		xml.Unmarshal(xmlBody, &resp)
		parsedRes, err := parseATCCheckResult(&config, &httpClient, resp)

		assert.EqualError(t, err, "conversion of ATC check results to CheckStyle has failed: get file name has failed: could not check readable source format: could not get repository layout: a http error occurred")
		assert.NotEmpty(t, parsedRes)
	})

}

type httpMockGctsT struct {
	Method       string                  // is set during test execution
	URL          string                  // is set before test execution
	Header       map[string][]string     // is set before test execution
	ResponseBody string                  // is set before test execution
	Options      piperhttp.ClientOptions // is set during test
	StatusCode   int                     // is set during test
}

func (c *httpMockGctsT) SetOptions(options piperhttp.ClientOptions) {
	c.Options = options
}

func (c *httpMockGctsT) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {

	c.Method = method
	c.URL = url

	switch url {

	case "http://testHost.com:50000/sap/bc/adt/core/discovery?sap-client=000":

		c.Header = map[string][]string{"X-Csrf-Token": {"ZegUEgfa50R7ZfGGxOtx2A=="}}

		c.ResponseBody = `
		<?xml version="1.0" encoding="utf-8"?>
				<app:service xmlns:app="http://www.w3.org/2007/app" xmlns:atom="http://www.w3.org/2005/Atom"/>
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo?sap-client=000":

		c.ResponseBody = `
		{
			"result": {

					"rid": "testRepo",
					"name": "testRepo",
					"role": "PROVIDED",
					"type": "GIT",
					"vsid": "vSID",
					"privateFlag": "false",
					"url": "http://github.com/testRepo",
					"createdBy": "testUser",
					"createdDate": "02/02/2022",
					"objects": 3,
					"currentCommit": "xyz987654321"
				}
		}
		`
	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2?sap-client=000":

		c.ResponseBody = `
			{
				"result": {

						"rid": "testRepo2",
						"name": "testRepo2",
						"role": "PROVIDED",
						"type": "GIT",
						"vsid": "vSID",
						"privateFlag": "false",
						"url": "http://github.com/testRepo2",
						"createdBy": "testUser",
						"createdDate": "02/02/2022",
						"objects": 3,
						"currentCommit": "xyz987654321"
					}

			}
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/getCommit?sap-client=000":

		c.ResponseBody = `
			{
				"commits": [
					{

						"id": "0123456789abcdefghijkl"

					},
					{

						"id": "7845abaujztrw785"
					},
					{

						"id": "45poiztr785423"
					}
				]
			}
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2/getCommit?sap-client=000":

		c.ResponseBody = `
			{
				"commits": [
					{

						"id": "0123456789abcdefghijkl"

					},
					{

						"id": "7845abaujztrw785"
					},
					{

						"id": "45poiztr785423"
					}
				]
			}
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/getHistory?sap-client=000":

		c.ResponseBody = `
			{
				"result": [
					{

							"rid": "testRepo",
							"checkoutTime": 20220216233655,
							"fromCommit": "xyz987654321",
							"toCommit": "0123456789abcdefghijkl",
							"caller": "USER",
							"type": "PULL"
					},
					{
						    "rid": "testRepo",
						    "checkoutTime": 20220216233788,
						    "fromCommit": "ghi98765432",
						    "toCommit": "xyz987654321",
						    "caller": "USER",
						    "type": "PULL"
					}
				]
			}
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo2/getHistory?sap-client=000":

		c.ResponseBody = `
				{
					"result": [
						{

								"rid": "testRepo",
								"checkoutTime": 20220216233655,
								"fromCommit": "xyz987654321",
								"toCommit": "0123456789abcdefghijkl",
								"caller": "USER",
								"type": "PULL"
						},
						{
								"rid": "testRepo",
								"checkoutTime": 20220216233788,
								"fromCommit": "ghi98765432",
								"toCommit": "xyz987654321",
								"caller": "USER",
								"type": "PULL"
						}
					]
				}
				`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/compareCommits?fromCommit=xyz987654321&toCommit=0123456789abcdefghijkl&sap-client=000":

		c.ResponseBody = `
			{
				"objects": [
					{

							"name": "ZCL_GCTS",
							"type": "CLAS",
							"action": "Class (ABAP Objects)"
					},
					{

							"name": "ZP_GCTS",
							"type": "DEVC",
							"action": "Package(ABAP Objects)"
					},
					{

							"name": "ZIF_GCTS_API",
							"type": "INTF",
							"action": "Interface (ABAP Objects)"
					}
				]
			}
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/compareCommits?fromCommit=7845abaujztrw785&toCommit=0123456789abcdefghijkl&sap-client=000":

		c.ResponseBody = `
			{
				"objects": [
					{

							"name": "ZCL_GCTS",
							"type": "CLAS",
							"action": "Class (ABAP Objects)"
					},
					{

							"name": "ZP_GCTS",
							"type": "DEVC",
							"action": "Package(ABAP Objects)"
					},
					{

							"name": "ZIF_GCTS_API",
							"type": "INTF",
							"action": "Interface (ABAP Objects)"
					}
				]
			}
			`
	case "http://testHost.com:50000/sap/bc/cts_abapvcs/objects/CLAS/ZCL_GCTS?sap-client=000":

		c.ResponseBody = `

			 {

					"pgmid": "R3TR",
					"object": "CLAS",
					"objName": "ZCL_GCTS",
					"srcsystem": "src",
					"author": "HUGO",
					"devclass": "SGCTS"
				}
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/objects/INTF/ZIF_GCTS_API?sap-client=000":

		c.ResponseBody = `

			 {

					"pgmid": "R3TR",
					"object": "INTF",
					"objName": "ZIF_GCTS_API",
					"srcsystem": "src",
					"author": "HUGO",
					"devclass": "SGCTS_2"
				}
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/objects/DEVC/ZP_GCTS?sap-client=000":

		c.ResponseBody = `

			{

				   "pgmid": "R3TR",
				   "object": "DEVC",
				   "objName": "ZP_GCTS",
				   "srcsystem": "src",
				   "author": "HUGO",
				   "devclass": "SGCTS"
			   }
		   `

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/layout?sap-client=000":

		c.ResponseBody =
			`{
				"layout": {
					"formatVersion": 5,
					"format": "json",
					"objectStorage": "plain",
					"metaInformation": ".gctsmetadata/",
					"tableContent": "true",
					"subdirectory": "src/",
					"readableSource": "false"
				}
			}
			`

	case "http://testHost.com:50000/sap/bc/cts_abapvcs/repository/testRepo/objects?sap-client=000":

		c.ResponseBody = `
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
							"description": "Package(ABAP Objects)"
					},
					{
							"pgmid": "R3TR",
							"object": "ZIF_GCTS_API",
							"type": "INTF",
							"description": "Interface (ABAP Objects)"
					}
				]
			}
			`

	case "http://testHost.com:50000/sap/bc/adt/abapunit/testruns?sap-client=000":

		c.Header = map[string][]string{"Accept": {"application/xml"}}
		c.Header = map[string][]string{"x-csrf-token": {"ZegUEgfa50R7ZfGGxOtx2A=="}}
		c.Header = map[string][]string{"Content-Type": {"application/vnd.sap.adt.abapunit.testruns.result.v1+xml"}}

		c.ResponseBody = `
		<?xml version="1.0" encoding="utf-8"?>
		<aunit:runResult xmlns:aunit="http://www.sap.com/adt/aunit">
				<program adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts_piper_demo" adtcore:type="CLAS/OC" adtcore:name="ZCL_GCTS" uriType="semantic" xmlns:adtcore="http://www.sap.com/adt/core">
						<testClasses>
								<testClass adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts/includes/testclasses#type=CLAS%2FOCL;name=LTCL_MASTER" adtcore:type="CLAS/OL" adtcore:name="LTCL_MASTER" uriType="semantic" navigationUri="/sap/bc/adt/oo/classes/zcl_gcts/includes/testclasses#type=CLAS%2FOCL;name=LTCL_MASTER" durationCategory="short" riskLevel="harmless">
										<testMethods>
												<testMethod adtcore:uri="/sap/bc/adt/oo/classes/zcl_gcts/includes/testclasses#type=CLAS%2FOLD;name=LTCL_MASTER%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20CHECK" adtcore:type="CLAS/OLI" adtcore:name="CHECK" executionTime="0" uriType="semantic" navigationUri="/sap/bc/adt/oo/classes/zcl_gcts/includes/testclasses#type=CLAS%2FOLD;name=LTCL_MASTER%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20CHECK" unit="s"/>
										</testMethods>
								</testClass>
						</testClasses>
				</program>
		</aunit:runResult>
		`

	case "http://testHost.com:50000/sap/bc/adt/atc/worklists?checkVariant=DEFAULT_REMOTE_REF&sap-client=000":

		c.Header = map[string][]string{"Location": {"/atc/worklists/worklistId/123Z076495C01ABCDEF9C9F6B257OD70"}}

	case "http://testHost.com:50000/sap/bc/adt/atc/worklists?checkVariant=CUSTOM_REMOTE_REF&sap-client=000":

		c.Header = map[string][]string{"Location": {"/atc/worklists/worklistId/3E3D0764F95Z01ABCDHEF9C9F6B5C14P"}}

	case "http://testHost.com:50000/sap/bc/adt/atc/runs?worklistId=123Z076495C01ABCDEF9C9F6B257OD70?sap-client=000":

		c.ResponseBody =
			`<?xml version="1.0" encoding="utf-8"?>
			<atcworklist:worklistRun xmlns:atcworklist="http://www.sap.com/adt/atc/worklist">
				<atcworklist:worklistId>123Z076495C01ABCDEF9C9F6B257OD70</atcworklist:worklistId>
				<atcworklist:worklistTimestamp>2021-11-29T14:46:46Z</atcworklist:worklistTimestamp>
				<atcworklist:infos>
					<atcinfo:info xmlns:atcinfo="http://www.sap.com/adt/atc/info">
						<atcinfo:type>FINDING_STATS</atcinfo:type>
						<atcinfo:description>0,0,1</atcinfo:description>
					</atcinfo:info>
				</atcworklist:infos>
			</atcworklist:worklistRun>`

	case "http://testHost.com:50000/sap/bc/adt/atc/worklists/3E3D0764F95Z01ABCDHEF9C9F6B5C14P?sap-client=000":
		c.ResponseBody =
			`<?xml version="1.0" encoding="utf-8"?>
			<atcworklist:worklist atcworklist:id="42010AEF3CA51EDC94AC4683B035E12D" atcworklist:timestamp="2021-11-29T22:18:58Z" atcworklist:usedObjectSet="99999999999999999999999999999999" atcworklist:objectSetIsComplete="true" xmlns:atcworklist="http://www.sap.com/adt/atc/worklist">
				<atcworklist:objectSets>
					<atcworklist:objectSet atcworklist:name="00000000000000000000000000000000" atcworklist:title="All Objects" atcworklist:kind="ALL"/>
					<atcworklist:objectSet atcworklist:name="99999999999999999999999999999999" atcworklist:title="Last Check Run" atcworklist:kind="LAST_RUN"/>
				</atcworklist:objectSets>
				<atcworklist:objects/>
			</atcworklist:worklist>`

	case "http://testHost.com:50000/sap/bc/adt/atc/worklists/123Z076495C01ABCDEF9C9F6B257OD70?sap-client=000":
		c.ResponseBody =
			`<?xml version="1.0" encoding="utf-8"?>
<atcworklist:worklist atcworklist:id="248A076493C01EEC8FA9CEAED527BD53" atcworklist:timestamp="2021-11-04T09:08:18Z" atcworklist:usedObjectSet="99999999999999999999999999999999" atcworklist:objectSetIsComplete="true" xmlns:atcworklist="http://www.sap.com/adt/atc/worklist">
	<atcworklist:objectSets>
		<atcworklist:objectSet atcworklist:name="00000000000000000000000000000000" atcworklist:title="All Objects" atcworklist:kind="ALL"/>
		<atcworklist:objectSet atcworklist:name="99999999999999999999999999999999" atcworklist:title="Last Check Run" atcworklist:kind="LAST_RUN"/>
	</atcworklist:objectSets>
	<atcworklist:objects>
		<atcobject:object adtcore:uri="/sap/bc/adt/atc/objects/R3TR/CLAS/ZCL_GCTS" adtcore:type="CLAS" adtcore:name="ZCL_GCTS" adtcore:packageName="ZPL_GCTS" atcobject:author="testUser" xmlns:atcobject="http://www.sap.com/adt/atc/object" xmlns:adtcore="http://www.sap.com/adt/core">
			<atcobject:findings>
				<atcfinding:finding adtcore:uri="/sap/bc/adt/atc/findings/itemid/248A076493C01EEC8FA9D609860AFD93/index/233439" atcfinding:location="/sap/bc/adt/oo/classes/zcl_gcts/source/main#type=CLAS%2FOSI;name=ZCL_GCTS;start=20" atcfinding:processor="testUser" atcfinding:lastChangedBy="testUser" atcfinding:priority="1" atcfinding:checkId="78A08159CD2A822100535FBEB655BDB8" atcfinding:checkTitle="Package Check (Remote-enabled)" atcfinding:messageId="USEM" atcfinding:messageTitle="Package Violation - Error" atcfinding:exemptionApproval="" atcfinding:exemptionKind="" atcfinding:checksum="553596936" atcfinding:quickfixInfo="atc:248A076493C01EEC8FA9D609860AFD93,233439" xmlns:atcfinding="http://www.sap.com/adt/atc/finding">
					<atom:link href="/sap/bc/adt/documentation/atc/documents/itemid/248A076493C01EEC8FA9D609860AFD93/index/233439" rel="http://www.sap.com/adt/relations/documentation" type="text/html" xmlns:atom="http://www.w3.org/2005/Atom"/>
					<atcfinding:quickfixes atcfinding:manual="false" atcfinding:automatic="false" atcfinding:pseudo="false"/>
				</atcfinding:finding>
			</atcobject:findings>
		</atcobject:object>
	</atcworklist:objects>
</atcworklist:worklist>`
	}

	if r != nil {
		_, err := io.ReadAll(r)

		if err != nil {
			return nil, err
		}
	}

	res := http.Response{
		StatusCode: c.StatusCode,
		Header:     c.Header,
		Body:       io.NopCloser(bytes.NewReader([]byte(c.ResponseBody))),
	}

	if c.StatusCode >= 400 {
		return &res, errors.New("a http error occurred")
	}

	return &res, nil
}
