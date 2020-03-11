package nexus

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestAddArtifact(t *testing.T) {
	t.Run("Test valid artifact", func(t *testing.T) {
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
	})
	t.Run("Test missing ID", func(t *testing.T) {
		nexusUpload := Upload{}

		err := nexusUpload.AddArtifact(ArtifactDescription{
			ID:         "",
			Classifier: "",
			Type:       "pom",
			File:       "pom.xml",
		})

		assert.Error(t, err, "Expected to fail adding invalid artifact")
		assert.True(t, len(nexusUpload.artifacts) == 0)
	})
	t.Run("Test invalid ID", func(t *testing.T) {
		nexusUpload := Upload{}

		err := nexusUpload.AddArtifact(ArtifactDescription{
			ID:         "artifact/id",
			Classifier: "",
			Type:       "pom",
			File:       "pom.xml",
		})

		assert.Error(t, err, "Expected to fail adding invalid artifact")
		assert.True(t, len(nexusUpload.artifacts) == 0)
	})
	t.Run("Test missing type", func(t *testing.T) {
		nexusUpload := Upload{}

		err := nexusUpload.AddArtifact(ArtifactDescription{
			ID:         "artifact",
			Classifier: "",
			Type:       "",
			File:       "pom.xml",
		})

		assert.Error(t, err, "Expected to fail adding invalid artifact")
		assert.True(t, len(nexusUpload.artifacts) == 0)
	})
	t.Run("Test missing file", func(t *testing.T) {
		nexusUpload := Upload{}

		err := nexusUpload.AddArtifact(ArtifactDescription{
			ID:         "artifact",
			Classifier: "",
			Type:       "pom",
			File:       "",
		})

		assert.Error(t, err, "Expected to fail adding invalid artifact")
		assert.True(t, len(nexusUpload.artifacts) == 0)
	})
	t.Run("Test adding duplicate artifact is ignored", func(t *testing.T) {
		nexusUpload := Upload{}

		_ = nexusUpload.AddArtifact(ArtifactDescription{
			ID:         "blob",
			Classifier: "",
			Type:       "pom",
			File:       "pom.xml",
		})
		err := nexusUpload.AddArtifact(ArtifactDescription{
			ID:         "blob",
			Classifier: "",
			Type:       "pom",
			File:       "pom.xml",
		})
		assert.NoError(t, err, "Expected to succeed adding duplicate artifact")
		assert.True(t, len(nexusUpload.artifacts) == 1)
	})
}

func TestGetArtifacts(t *testing.T) {
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

func TestGetBaseURL(t *testing.T) {
	// Invalid parameters to getBaseURL() already tested via SetBaseURL() tests
	t.Run("Test base URL for nexus2 is sensible", func(t *testing.T) {
		baseURL, err := getBaseURL("localhost:8081/nexus", "nexus2", "maven-releases", "some.group.id")
		assert.NoError(t, err, "Expected getBaseURL() to succeed")
		assert.Equal(t, "localhost:8081/nexus/content/repositories/maven-releases/some/group/id/", baseURL)
	})
	t.Run("Test base URL for nexus3 is sensible", func(t *testing.T) {
		baseURL, err := getBaseURL("localhost:8081", "nexus3", "maven-releases", "some.group.id")
		assert.NoError(t, err, "Expected getBaseURL() to succeed")
		assert.Equal(t, "localhost:8081/repository/maven-releases/some/group/id/", baseURL)
	})
}

func TestSetBaseURL(t *testing.T) {
	t.Run("Test no host provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetBaseURL("", "nexus3", "maven-releases", "some.group.id")
		assert.Error(t, err, "Expected SetBaseURL() to fail (no host)")
	})
	t.Run("Test host wrongly includes protocol http://", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetBaseURL("htTp://localhost:8081", "nexus3", "maven-releases", "some.group.id")
		assert.Error(t, err, "Expected SetBaseURL() to fail (invalid host)")
	})
	t.Run("Test host wrongly includes protocol https://", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetBaseURL("htTpS://localhost:8081", "nexus3", "maven-releases", "some.group.id")
		assert.Error(t, err, "Expected SetBaseURL() to fail (invalid host)")
	})
	t.Run("Test invalid version provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetBaseURL("localhost:8081", "3", "maven-releases", "some.group.id")
		assert.Error(t, err, "Expected SetBaseURL() to fail (invalid nexus version)")
	})
	t.Run("Test no repository provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetBaseURL("localhost:8081", "nexus3", "", "some.group.id")
		assert.Error(t, err, "Expected SetBaseURL() to fail (no repository)")
	})
	t.Run("Test no group id provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetBaseURL("localhost:8081", "nexus3", "maven-releases", "")
		assert.Error(t, err, "Expected SetBaseURL() to fail (no groupID)")
	})
	t.Run("Test no nexus version provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetBaseURL("localhost:8081", "nexus1", "maven-releases", "some.group.id")
		assert.Error(t, err, "Expected SetBaseURL() to fail (unsupported nexus version)")
	})
	t.Run("Test unsupported nexus version provided", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetBaseURL("localhost:8081", "nexus1", "maven-releases", "some.group.id")
		assert.Error(t, err, "Expected SetBaseURL() to fail (unsupported nexus version)")
	})
}

func TestSetArtifactsVersion(t *testing.T) {
	t.Run("Test invalid artifact version", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetArtifactsVersion("")
		assert.Error(t, err, "Expected SetArtifactsVersion() to fail (empty version)")
	})
	t.Run("Test valid artifact version", func(t *testing.T) {
		nexusUpload := Upload{}
		err := nexusUpload.SetArtifactsVersion("1.0.0-SNAPSHOT")
		assert.NoError(t, err, "Expected SetArtifactsVersion() to succeed")
	})
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
		assert.EqualError(t, err, "no artifacts to upload, call AddArtifact() first")
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

func TestUploadArtifacts(t *testing.T) {
	t.Run("Test upload works", func(t *testing.T) {
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
	})
	t.Run("Test upload fails at md5 hash", func(t *testing.T) {
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
	})
	t.Run("Test upload fails at sha1 hash", func(t *testing.T) {
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
	})
	t.Run("Test upload fails at file", func(t *testing.T) {
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
	})
}

func TestDownloadMavenMetadata(t *testing.T) {
	buffer, err := downloadIntoBuffer(
		"http://localhost:8081/repository/maven-snapshots/com/sap/opensap/employee-browser-application/maven-metadata.xml",
		50*1024)
	assert.NoError(t, err)
	// XML file should be UTF-8 encoded, for this the conversion to string works.
	originalString := string(buffer[:])
	fmt.Print(originalString)

	metadata, err := xmlBufferToMavenMetadata(buffer)
	assert.NoError(t, err)

	fmt.Printf("last build: %v\n", metadata.Versioning.Snapshot.BuildNumber)

	buffer, err = mavenMetadataToXMLBuffer(metadata)
	newString := string(buffer[:])

	assert.Equal(t, originalString, newString, "expected conversion to be loss-less")
}
