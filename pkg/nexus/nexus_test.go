package nexus

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestAddArtifactValid(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{
		ID:         "artifact.id",
		Classifier: "",
		Type:       "pom",
		File:       "pom.xml",
	})

	assert.NoError(t, err, "Expected to add valid artifact")
	assert.True(t, len(nexusUpload.artifacts) == 1)

	assert.True(t, nexusUpload.artifacts[0].ID == "artifact.id")
	assert.True(t, nexusUpload.artifacts[0].Classifier == "")
	assert.True(t, nexusUpload.artifacts[0].Type == "pom")
	assert.True(t, nexusUpload.artifacts[0].File == "pom.xml")
}

func TestAddArtifactMissingID(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{
		ID:         "",
		Classifier: "",
		Type:       "pom",
		File:       "pom.xml",
	})

	assert.Error(t, err, "Expected to fail adding invalid artifact")
	assert.True(t, len(nexusUpload.artifacts) == 0)
}

func TestAddDuplicateArtifact(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{
		ID:         "blob",
		Classifier: "",
		Type:       "pom",
		File:       "pom.xml",
	})
	err = nexusUpload.AddArtifact(ArtifactDescription{
		ID:         "blob",
		Classifier: "",
		Type:       "pom",
		File:       "pom.xml",
	})
	assert.NoError(t, err, "Expected to succeed adding duplicate artifact")
	assert.True(t, len(nexusUpload.artifacts) == 1)
}

func TestAddArtifactsFromValidJSON(t *testing.T) {
	nexusUpload := Upload{}

	artifactJSON := `
		[
			{
				"artifactId" : "my-app",
				"classifier" : "",
				"file"       : "app.jar",
				"type"       : "jar"
			},
			{
				"artifactId" : "my-app",
				"file"       : "pom.xml",
				"type"       : "pom"
			}
		]`

	err := nexusUpload.AddArtifactsFromJSON(artifactJSON)
	assert.NoError(t, err, "Expected to succeed adding 2 artifacts from valid JSON")
	assert.Equal(t, 2, len(nexusUpload.artifacts))
}

func TestAddNoArtifactsFromPartiallyValidJSON(t *testing.T) {
	nexusUpload := Upload{}

	artifactJSON := `
		[
			{
				"blah" : "blub",
			},
			{
				"artifactId" : "my-app",
				"file"       : "pom.xml",
				"type"       : "pom"
			}
		]`

	err := nexusUpload.AddArtifactsFromJSON(artifactJSON)
	assert.Error(t, err, "Expected to fail adding artifacts from partially valid JSON")
	assert.Equal(t, 0, len(nexusUpload.artifacts))
}

func TestAddArtifactsFromJSONWithWrongTypes(t *testing.T) {
	nexusUpload := Upload{}

	artifactJSON := `
		[
			{
				"artifactId" : "my-app",
				"file"       : 1,
				"type"       : true
			}
		]`

	err := nexusUpload.AddArtifactsFromJSON(artifactJSON)
	assert.Error(t, err, "Expected to fail adding artifacts from JSON with wrong types")
	assert.Equal(t, 0, len(nexusUpload.artifacts))
}

func TestAddNoArtifactsFromInvalidJSON(t *testing.T) {
	nexusUpload := Upload{}

	artifactJSON := "grmpf"

	err := nexusUpload.AddArtifactsFromJSON(artifactJSON)
	assert.Error(t, err, "Expected to fail adding artifacts from invalid JSON")
	assert.Equal(t, 0, len(nexusUpload.artifacts))
}

func TestUpload_AddArtifactsFromJSON(t *testing.T) {
	json := `[{"artifactId":"myapp-pom","classifier":"myapp-1.0","type":"pom","file":"pom.xml"}]`
	nexusUpload := Upload{}
	if err := nexusUpload.AddArtifactsFromJSON(json); err != nil {
		t.Errorf("AddArtifactsFromJSON() error = %v", err)
	}
	assert.Equal(t, []ArtifactDescription{{
		ID:         "myapp-pom",
		Classifier: "myapp-1.0",
		Type:       "pom",
		File:       "pom.xml",
	}}, nexusUpload.artifacts)
}

func TestArtifactsNotDirectlyAccessible(t *testing.T) {
	nexusUpload := Upload{}

	err := nexusUpload.AddArtifact(ArtifactDescription{
		ID:         "artifact.id",
		Classifier: "",
		Type:       "pom",
		File:       "pom.xml",
	})
	assert.NoError(t, err, "Expected to succeed adding valid artifact")

	artifacts := nexusUpload.GetArtifacts()
	// Overwrite array entry in the returned array...
	artifacts[0] = ArtifactDescription{
		ID:         "another.id",
		Classifier: "",
		Type:       "pom",
		File:       "pom.xml",
	}
	// ... but expect the entry in nexusUpload object to be unchanged
	assert.True(t, nexusUpload.artifacts[0].ID == "artifact.id")
}

func TestSensibleBaseURLNexus2(t *testing.T) {
	baseURL, err := getBaseURL("localhost:8081/nexus", "nexus2", "maven-releases", "some.group.id")
	assert.NoError(t, err, "Expected getBaseURL() to succeed")
	assert.Equal(t, "localhost:8081/nexus/content/repositories/maven-releases/some/group/id/", baseURL)
}

func TestSensibleBaseURLNexus3(t *testing.T) {
	baseURL, err := getBaseURL("localhost:8081", "nexus3", "maven-releases", "some.group.id")
	assert.NoError(t, err, "Expected getBaseURL() to succeed")
	assert.Equal(t, "localhost:8081/repository/maven-releases/some/group/id/", baseURL)
}

func TestBaseURLUnsupportedNexusVersion(t *testing.T) {
	baseURL, err := getBaseURL("localhost:8081", "nexus1", "maven-releases", "some.group.id")
	assert.Error(t, err, "Expected getBaseURL() to fail")
	assert.Equal(t, "", baseURL)
}

func TestSetBaseURLParamChecking(t *testing.T) {
	nexusUpload := Upload{}
	err := nexusUpload.SetBaseURL("", "nexus3", "maven-releases", "some.group.id")
	assert.Error(t, err, "Expected SetBaseURL() to fail (no host)")
	err = nexusUpload.SetBaseURL("localhost:8081", "3", "maven-releases", "some.group.id")
	assert.Error(t, err, "Expected SetBaseURL() to fail (invalid nexus version)")
	err = nexusUpload.SetBaseURL("localhost:8081", "nexus3", "", "some.group.id")
	assert.Error(t, err, "Expected SetBaseURL() to fail (no repository)")
	err = nexusUpload.SetBaseURL("localhost:8081", "nexus3", "maven-releases", "")
	assert.Error(t, err, "Expected SetBaseURL() to fail (no groupID)")
}

func TestSetInvalidArtifactsVersion(t *testing.T) {
	nexusUpload := Upload{}
	err := nexusUpload.SetArtifactsVersion("")
	assert.Error(t, err, "Expected SetArtifactsVersion() to fail (empty version)")
}

func TestSetValidArtifactsVersion(t *testing.T) {
	nexusUpload := Upload{}
	err := nexusUpload.SetArtifactsVersion("1.0.0-SNAPSHOT")
	assert.NoError(t, err, "Expected SetArtifactsVersion() to succeed")
}

type simpleHttpMock struct {
	responseStatus string
	responseError  error
}

func (m *simpleHttpMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return &http.Response{Status: m.responseStatus}, m.responseError
}

func (m *simpleHttpMock) SetOptions(options piperhttp.ClientOptions) {
}

func TestUploadNoInit(t *testing.T) {
	var mockedHttp = simpleHttpMock{
		responseStatus: "200 OK",
		responseError:  nil,
	}

	t.Run("Expect that upload fails without base-URL", func(t *testing.T) {
		nexusUpload := Upload{}

		err := nexusUpload.uploadArtifacts(&mockedHttp)
		assert.EqualError(t, err, "the nexus.Upload needs to be configured by calling SetBaseURL() first")
	})

	t.Run("Expect that upload fails without version", func(t *testing.T) {
		nexusUpload := Upload{}
		_ = nexusUpload.SetBaseURL("localhost:8081", "nexus3", "maven-releases", "my.group.id")

		err := nexusUpload.uploadArtifacts(&mockedHttp)
		assert.EqualError(t, err, "the nexus.Upload needs to be configured by calling SetArtifactsVersion() first")
	})

	t.Run("Expect that upload fails without artifacts", func(t *testing.T) {
		nexusUpload := Upload{}
		_ = nexusUpload.SetBaseURL("localhost:8081", "nexus3", "maven-releases", "my.group.id")
		_ = nexusUpload.SetArtifactsVersion("1.0")

		err := nexusUpload.uploadArtifacts(&mockedHttp)
		assert.EqualError(t, err, "no artifacts to upload, call AddArtifact() or AddArtifactsFromJSON() first")
	})
}

type request struct {
	method string
	url    string
}

type requestReply struct {
	response string
	err      error
}

type httpMock struct {
	username       string
	password       string
	requestIndex   int
	requestReplies []requestReply
	requests       []request
}

func (m *httpMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	// store the request
	m.requests = append(m.requests, request{method: method, url: url})

	// Return the configured response for this request's index
	response := m.requestReplies[m.requestIndex].response
	err := m.requestReplies[m.requestIndex].err

	m.requestIndex++

	return &http.Response{Status: response}, err
}

func (m *httpMock) SetOptions(options piperhttp.ClientOptions) {
	m.username = options.Username
	m.password = options.Password
}

func (m *httpMock) appendReply(reply requestReply) {
	m.requestReplies = append(m.requestReplies, reply)
}

func createConfiguredNexusUpload() Upload {
	nexusUpload := Upload{}
	_ = nexusUpload.SetBaseURL("localhost:8081", "nexus3", "maven-releases", "my.group.id")
	_ = nexusUpload.SetArtifactsVersion("1.0")
	_ = nexusUpload.AddArtifact(ArtifactDescription{ID: "artifact.id", Classifier: "", Type: "pom", File: "../../pom.xml"})
	return nexusUpload
}

func TestUploadWorks(t *testing.T) {
	var mockedHttp = httpMock{}
	// There will be three requests, md5, sha1 and the file itself
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})

	nexusUpload := createConfiguredNexusUpload()

	err := nexusUpload.uploadArtifacts(&mockedHttp)
	assert.NoError(t, err, "Expected that uploading the artifact works")

	assert.Equal(t, 3, mockedHttp.requestIndex, "Expected 3 HTTP requests")

	assert.Equal(t, http.MethodPut, mockedHttp.requests[0].method)
	assert.Equal(t, http.MethodPut, mockedHttp.requests[1].method)
	assert.Equal(t, http.MethodPut, mockedHttp.requests[2].method)

	assert.Equal(t, "http://localhost:8081/repository/maven-releases/my/group/id/artifact.id/1.0/artifact.id-1.0.pom.md5",
		mockedHttp.requests[0].url)
	assert.Equal(t, "http://localhost:8081/repository/maven-releases/my/group/id/artifact.id/1.0/artifact.id-1.0.pom.sha1",
		mockedHttp.requests[1].url)
	assert.Equal(t, "http://localhost:8081/repository/maven-releases/my/group/id/artifact.id/1.0/artifact.id-1.0.pom",
		mockedHttp.requests[2].url)
}

func TestUploadFailsAtMd5(t *testing.T) {
	var mockedHttp = httpMock{}
	// There will be three requests, md5, sha1 and the file itself
	mockedHttp.appendReply(requestReply{response: "404 Error", err: fmt.Errorf("failed")})
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})

	nexusUpload := createConfiguredNexusUpload()

	err := nexusUpload.uploadArtifacts(&mockedHttp)
	assert.Error(t, err, "Expected that uploading the artifact failed")

	assert.Equal(t, 1, mockedHttp.requestIndex, "Expected only one HTTP requests")
	assert.Equal(t, 1, len(nexusUpload.artifacts), "Expected the artifact to be still present in the nexusUpload")
}

func TestUploadFailsAtSha1(t *testing.T) {
	var mockedHttp = httpMock{}
	// There will be three requests, md5, sha1 and the file itself
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})
	mockedHttp.appendReply(requestReply{response: "404 Error", err: fmt.Errorf("failed")})
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})

	nexusUpload := createConfiguredNexusUpload()

	err := nexusUpload.uploadArtifacts(&mockedHttp)
	assert.Error(t, err, "Expected that uploading the artifact failed")

	assert.Equal(t, 2, mockedHttp.requestIndex, "Expected only two HTTP requests")
	assert.Equal(t, 1, len(nexusUpload.artifacts), "Expected the artifact to be still present in the nexusUpload")
}

func TestUploadFailsAtFile(t *testing.T) {
	var mockedHttp = httpMock{}
	// There will be three requests, md5, sha1 and the file itself
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})
	mockedHttp.appendReply(requestReply{response: "200 OK", err: nil})
	mockedHttp.appendReply(requestReply{response: "404 Error", err: fmt.Errorf("failed")})

	nexusUpload := createConfiguredNexusUpload()

	err := nexusUpload.uploadArtifacts(&mockedHttp)
	assert.Error(t, err, "Expected that uploading the artifact failed")

	assert.Equal(t, 3, mockedHttp.requestIndex, "Expected only three HTTP requests")
	assert.Equal(t, 1, len(nexusUpload.artifacts), "Expected the artifact to be still present in the nexusUpload")
}
