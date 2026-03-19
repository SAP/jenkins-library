//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type buildkitMockClient struct {
	httpMethod     string
	httpStatusCode int
	urlsCalled     []string
	requestBody    io.Reader
	responseBody   string
	errorMessage   string
}

func (c *buildkitMockClient) SetOptions(opts piperhttp.ClientOptions) {}

func (c *buildkitMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlsCalled = append(c.urlsCalled, url)
	c.requestBody = body
	if len(c.errorMessage) > 0 {
		return nil, fmt.Errorf("%s", c.errorMessage)
	}
	return &http.Response{StatusCode: c.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(c.responseBody)))}, nil
}

func (c *buildkitMockClient) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	if len(c.errorMessage) > 0 {
		return fmt.Errorf("%s", c.errorMessage)
	}
	return nil
}

func TestRunBuildkitExecute(t *testing.T) {
	// required due to config resolution during build settings retrieval
	openFileBak := configOptions.OpenFile
	defer func() {
		configOptions.OpenFile = openFileBak
	}()
	configOptions.OpenFile = configOpenFileMock

	t.Run("success case - full image reference", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImage:    "my.registry.io/myimage:mytag",
			DockerfilePath:    "Dockerfile",
			DockerConfigJSON:  "path/to/docker/config.json",
			BuildSettingsInfo: `{"mavenExecuteBuild":[{"dockerImage":"maven"}]}`,
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, "https://my.registry.io", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "myimage:mytag", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myimage:mytag")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myimage")

		assert.Equal(t, 1, len(runner.Calls))
		assert.Equal(t, "docker", runner.Calls[0].Exec)
		assert.Contains(t, runner.Calls[0].Params, "buildx")
		assert.Contains(t, runner.Calls[0].Params, "build")
		assert.Contains(t, runner.Calls[0].Params, "--push")
		assert.Contains(t, runner.Calls[0].Params, "my.registry.io/myimage:mytag")
		assert.Contains(t, runner.Calls[0].Params, "--file")
		assert.Contains(t, runner.Calls[0].Params, "Dockerfile")

		// docker config written
		c, err := fileUtils.FileRead(buildkitDockerConfigPath)
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		// build settings info populated
		assert.Contains(t, commonPipelineEnvironment.custom.buildSettingsInfo, `"mavenExecuteBuild":[{"dockerImage":"maven"}]`)
		assert.Contains(t, commonPipelineEnvironment.custom.buildSettingsInfo, `"buildkitExecute"`)
	})

	t.Run("success case - image params (name + tag + registry)", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:   "myImage",
			ContainerImageTag:    "1.2.3-a+x",
			ContainerRegistryURL: "https://my.registry.com:50000",
			DockerfilePath:       "Dockerfile",
			DockerConfigJSON:     "path/to/docker/config.json",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, "https://my.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "myImage:1.2.3-a-x", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage:1.2.3-a-x")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage")

		assert.Equal(t, "docker", runner.Calls[0].Exec)
		assert.Contains(t, runner.Calls[0].Params, "--push")
		assert.Contains(t, runner.Calls[0].Params, "my.registry.com:50000/myImage:1.2.3-a-x")

		assert.Equal(t, "", commonPipelineEnvironment.container.imageDigest)
		assert.Empty(t, commonPipelineEnvironment.container.imageDigests)
	})

	t.Run("success case - destination in buildOptions (-t)", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			BuildOptions:   []string{"-t", "my.other.registry.com:50000/myImage:3.2.1-a-x"},
			DockerfilePath: "Dockerfile",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, "https://my.other.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "myImage:3.2.1-a-x", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage:3.2.1-a-x")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage")

		assert.Equal(t, "docker", runner.Calls[0].Exec)
		assert.Contains(t, runner.Calls[0].Params, "--push")
		// -t is passed through via buildOptions
		assert.Contains(t, runner.Calls[0].Params, "-t")
		assert.Contains(t, runner.Calls[0].Params, "my.other.registry.com:50000/myImage:3.2.1-a-x")

		assert.Equal(t, "", commonPipelineEnvironment.container.imageDigest)
		assert.Empty(t, commonPipelineEnvironment.container.imageDigests)
	})

	t.Run("success case - no push default", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			DockerfilePath: "Dockerfile",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(runner.Calls))
		assert.Equal(t, "docker", runner.Calls[0].Exec)
		assert.NotContains(t, runner.Calls[0].Params, "--push")

		// docker config written with default empty auths
		c, err := fileUtils.FileRead(buildkitDockerConfigPath)
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{}}`, string(c))
	})

	t.Run("success case - image digest capture", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImage:   "my.registry.io/myimage:mytag",
			DockerfilePath:   "Dockerfile",
			ReadImageDigest:  true,
			DockerConfigJSON: "path/to/docker/config.json",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/tmp/*-buildkitExecutetest/metadata.json", []byte(`{
			"buildx.build.provenance": {},
			"buildx.build.ref": "default/default/abc123",
			"containerimage.config.digest": "sha256:configdigest",
			"containerimage.digest": "sha256:468dd1253cc9f498fc600454bb8af96d880fec3f9f737e7057692adfe9f7d5b0"
		}`))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Contains(t, runner.Calls[0].Params, "--metadata-file")

		assert.Equal(t, "sha256:468dd1253cc9f498fc600454bb8af96d880fec3f9f737e7057692adfe9f7d5b0", commonPipelineEnvironment.container.imageDigest)
		assert.Equal(t, []string{"sha256:468dd1253cc9f498fc600454bb8af96d880fec3f9f737e7057692adfe9f7d5b0"}, commonPipelineEnvironment.container.imageDigests)
	})

	t.Run("success case - multi image build with root image", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub1/Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub2/Dockerfile", []byte("some content"))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, 3, len(runner.Calls))

		cwd, _ := fileUtils.Getwd()
		expectedParams := [][]string{
			{"buildx", "build", "--file", "Dockerfile", "-t", "my.registry.com:50000/myImage:myTag", "--push", cwd},
			{"buildx", "build", "--file", filepath.Join("sub1", "Dockerfile"), "-t", "my.registry.com:50000/myImage-sub1:myTag", "--push", cwd},
			{"buildx", "build", "--file", filepath.Join("sub2", "Dockerfile"), "-t", "my.registry.com:50000/myImage-sub2:myTag", "--push", cwd},
		}
		for _, call := range runner.Calls {
			assert.Equal(t, "docker", call.Exec)
			found := false
			for _, expected := range expectedParams {
				if strings.Join(call.Params, " ") == strings.Join(expected, " ") {
					found = true
					break
				}
			}
			assert.True(t, found, fmt.Sprintf("%v not found", call.Params))
		}

		assert.Equal(t, "https://my.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "myImage:myTag", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage-sub1")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage-sub2")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage-sub1:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage-sub2:myTag")
	})

	t.Run("success case - multi image build with excludes", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:               "myImage",
			ContainerImageTag:                "myTag",
			ContainerRegistryURL:             "https://my.registry.com:50000",
			ContainerMultiImageBuild:         true,
			ContainerMultiImageBuildExcludes: []string{"Dockerfile"},
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub1/Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub2/Dockerfile", []byte("some content"))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(runner.Calls))

		assert.Equal(t, "https://my.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage-sub1")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage-sub2")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage-sub1:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage-sub2:myTag")
	})

	t.Run("success case - multi context explicit build", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:   "myImage",
			ContainerImageTag:    "myTag",
			ContainerRegistryURL: "https://my.registry.com:50000",
			DockerfilePath:       "Dockerfile",
			MultipleImages: []map[string]interface{}{
				{
					"contextSubPath":     "/test1",
					"containerImageName": "myImageOne",
				},
				{
					"contextSubPath":     "/test2",
					"containerImageName": "myImageTwo",
					"containerImageTag":  "myTagTwo",
				},
			},
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(runner.Calls))

		cwd, _ := fileUtils.Getwd()
		expectedParams := [][]string{
			{"buildx", "build", "--file", "Dockerfile", "-t", "my.registry.com:50000/myImageOne:myTag", "--push", filepath.Join(cwd, "/test1")},
			{"buildx", "build", "--file", "Dockerfile", "-t", "my.registry.com:50000/myImageTwo:myTagTwo", "--push", filepath.Join(cwd, "/test2")},
		}
		for _, call := range runner.Calls {
			assert.Equal(t, "docker", call.Exec)
			found := false
			for _, expected := range expectedParams {
				if strings.Join(call.Params, " ") == strings.Join(expected, " ") {
					found = true
					break
				}
			}
			assert.True(t, found, fmt.Sprintf("%v not found", call.Params))
		}

		assert.Equal(t, "https://my.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "myImage:myTag", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImageOne")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImageTwo")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImageOne:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImageTwo:myTagTwo")
	})

	t.Run("success case - createBOM", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImage:   "myImage:tag",
			DockerfilePath:   "Dockerfile",
			DockerConfigJSON: "path/to/docker/config.json",
			CreateBOM:        true,
			SyftDownloadURL:  "http://test-syft-url.io",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		fakeArchive, err := fileUtils.CreateArchive(map[string][]byte{"syft": []byte("test")})
		assert.NoError(t, err)

		httpmock.RegisterResponder(http.MethodGet, "http://test-syft-url.io", httpmock.NewBytesResponder(http.StatusOK, fakeArchive))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		err = runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, client, fileUtils)
		assert.NoError(t, err)

		assert.Equal(t, "docker", runner.Calls[0].Exec)
		assert.Equal(t, "myImage:tag", commonPipelineEnvironment.container.imageNameTag)
		assert.Equal(t, "https://index.docker.io", commonPipelineEnvironment.container.registryURL)

		// syft was called
		assert.Equal(t, "/tmp/syfttest/syft", runner.Calls[1].Exec)
		assert.Equal(t, []string{"scan", "registry:index.docker.io/myImage:tag", "-o", "cyclonedx-xml@1.4=bom-docker-0.xml", "-q"}, runner.Calls[1].Params)
	})

	t.Run("success case - docker config with custom TLS certificates", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImage:            "my.registry.io/myimage:mytag",
			DockerfilePath:            "Dockerfile",
			DockerConfigJSON:          "path/to/docker/config.json",
			CustomTLSCertificateLinks: []string{"https://test.url/cert.crt"},
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		certClient := &buildkitMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile(buildkitTLSCertPath, []byte(``))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
	})

	t.Run("success case - registry mirrors (daemon.json written)", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImage:   "my.registry.io/myimage:mytag",
			DockerfilePath:   "Dockerfile",
			RegistryMirrors:  []string{"mirror.gcr.io", "192.168.0.1:5000"},
			DockerConfigJSON: "path/to/docker/config.json",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)

		daemonJSON, err := fileUtils.FileRead(buildkitDaemonJSONPath)
		assert.NoError(t, err)
		assert.Contains(t, string(daemonJSON), `"registry-mirrors"`)
		assert.Contains(t, string(daemonJSON), "mirror.gcr.io")
		assert.Contains(t, string(daemonJSON), "192.168.0.1:5000")
	})

	t.Run("success case - backward compatibility containerBuildOptions", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerBuildOptions: "--label test=true",
			ContainerImage:        "my.registry.io/myimage:mytag",
			DockerfilePath:        "Dockerfile",
			DockerConfigJSON:      "path/to/docker/config.json",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		telemetryData := telemetry.CustomData{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))

		err := runBuildkitExecute(config, &telemetryData, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, "--label test=true", telemetryData.ContainerBuildOptions)
		assert.Contains(t, runner.Calls[0].Params, "--label")
	})

	t.Run("success case - updating existing docker config json with additional credentials", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:        "myImage",
			ContainerImageTag:         "1.2.3-a+x",
			ContainerRegistryURL:      "https://my.registry.com:50000",
			DockerfilePath:            "Dockerfile",
			DockerConfigJSON:          "path/to/docker/config.json",
			ContainerRegistryUser:     "dummyUser",
			ContainerRegistryPassword: "dummyPassword",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths": {"dummyUrl": {"auth": "XXXXXXX"}}}`))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)

		c, err := fileUtils.FileRead(buildkitDockerConfigPath)
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"dummyUrl":{"auth":"XXXXXXX"},"https://my.registry.com:50000":{"auth":"ZHVtbXlVc2VyOmR1bW15UGFzc3dvcmQ="}}}`, string(c))
	})

	t.Run("success case - creating new docker config json with container credentials", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:        "myImage",
			ContainerImageTag:         "1.2.3-a+x",
			ContainerRegistryURL:      "https://my.registry.com:50000",
			DockerfilePath:            "Dockerfile",
			ContainerRegistryUser:     "dummyUser",
			ContainerRegistryPassword: "dummyPassword",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)

		c, err := fileUtils.FileRead(buildkitDockerConfigPath)
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"https://my.registry.com:50000":{"auth":"ZHVtbXlVc2VyOmR1bW15UGFzc3dvcmQ="}}}`, string(c))
	})

	t.Run("success case - multi image build with CreateBOM", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
			DockerConfigJSON:         "path/to/docker/config.json",
			CreateBOM:                true,
			SyftDownloadURL:          "http://test-syft-url.io",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub1/Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub2/Dockerfile", []byte("some content"))

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		fakeArchive, err := fileUtils.CreateArchive(map[string][]byte{"syft": []byte("test")})
		assert.NoError(t, err)

		httpmock.RegisterResponder(http.MethodGet, "http://test-syft-url.io", httpmock.NewBytesResponder(http.StatusOK, fakeArchive))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		err = runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, client, fileUtils)
		assert.NoError(t, err)

		// 3 docker builds + 3 syft scans
		assert.Equal(t, 6, len(runner.Calls))
		assert.Equal(t, "docker", runner.Calls[0].Exec)
		assert.Equal(t, "docker", runner.Calls[1].Exec)
		assert.Equal(t, "docker", runner.Calls[2].Exec)

		assert.Equal(t, "https://my.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "myImage:myTag", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage-sub1")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage-sub2")
	})

	t.Run("success case - multi context build with CreateBOM", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:   "myImage",
			ContainerImageTag:    "myTag",
			ContainerRegistryURL: "https://my.registry.com:50000",
			DockerConfigJSON:     "path/to/docker/config.json",
			DockerfilePath:       "Dockerfile",
			CreateBOM:            true,
			SyftDownloadURL:      "http://test-syft-url.io",
			MultipleImages: []map[string]interface{}{
				{
					"contextSubPath":     "/test1",
					"containerImageName": "myImageOne",
				},
				{
					"contextSubPath":     "/test2",
					"containerImageName": "myImageTwo",
					"containerImageTag":  "myTagTwo",
				},
			},
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		fakeArchive, err := fileUtils.CreateArchive(map[string][]byte{"syft": []byte("test")})
		assert.NoError(t, err)

		httpmock.RegisterResponder(http.MethodGet, "http://test-syft-url.io", httpmock.NewBytesResponder(http.StatusOK, fakeArchive))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		err = runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, client, fileUtils)
		assert.NoError(t, err)

		// 2 docker builds + 2 syft scans
		assert.Equal(t, 4, len(runner.Calls))
		assert.Equal(t, "docker", runner.Calls[0].Exec)
		assert.Equal(t, "docker", runner.Calls[1].Exec)

		assert.Equal(t, "https://my.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "myImage:myTag", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImageOne")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImageTwo")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImageOne:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImageTwo:myTagTwo")
	})

	t.Run("success case - createBuildArtifactsMetadata", func(t *testing.T) {
		// This test requires a real bom file on disk since filepath.Glob is used
		validBom := `<bom>
			<metadata>
				<component>
					<name>my.registry.io/dummyImage</name>
					<version>1.0.0</version>
				</component>
			</metadata>
		</bom>`
		bomFile, err := os.Create("bom-docker-0.xml")
		assert.NoError(t, err)
		defer bomFile.Close()
		defer os.Remove("bom-docker-0.xml")
		_, err = bomFile.WriteString(validBom)
		assert.NoError(t, err)

		imageNameTags := []string{"dummyImage:1.0.0"}
		cpe := buildkitExecuteCommonPipelineEnvironment{}

		err = buildkitCreateDockerBuildArtifactMetadata(imageNameTags, &cpe)

		assert.NoError(t, err)
	})

	t.Run("error case - container preparation failed", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerPreparationCommand: "failing-command arg1",
			DockerfilePath:              "Dockerfile",
		}
		runner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{
				"failing-command": fmt.Errorf("prep command failed"),
			},
		}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.EqualError(t, err, "failed to run container preparation command: prep command failed")
	})

	t.Run("error case - docker config read failed", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			DockerConfigJSON: "path/to/docker/config.json",
			DockerfilePath:   "Dockerfile",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.FileReadErrors = map[string]error{"path/to/docker/config.json": fmt.Errorf("read error")}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.EqualError(t, err, "failed to read existing docker config json at 'path/to/docker/config.json': read error")
	})

	t.Run("error case - docker config write failed", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			DockerConfigJSON: "path/to/docker/config.json",
			DockerfilePath:   "Dockerfile",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.FileWriteErrors = map[string]error{buildkitDockerConfigPath: fmt.Errorf("write error")}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.EqualError(t, err, fmt.Sprintf("failed to write file '%s': write error", buildkitDockerConfigPath))
	})

	t.Run("error case - BuildKit execution failed", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImage:   "my.registry.io/myimage:mytag",
			DockerfilePath:   "Dockerfile",
			DockerConfigJSON: "path/to/docker/config.json",
		}
		runner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{
				"docker": fmt.Errorf("buildx build failed"),
			},
		}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.EqualError(t, err, "execution of 'docker buildx build' failed: buildx build failed")
	})

	t.Run("error case - multi image build: no docker files", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
		}
		cpe := buildkitExecuteCommonPipelineEnvironment{}
		runner := &mock.ExecMockRunner{}
		fileUtils := &mock.FilesMock{}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &cpe, runner, nil, fileUtils)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to identify image list for multi image build")
	})

	t.Run("error case - multi image build: no docker files to process", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:               "myImage",
			ContainerImageTag:                "myTag",
			ContainerRegistryURL:             "https://my.registry.com:50000",
			ContainerMultiImageBuild:         true,
			ContainerMultiImageBuildExcludes: []string{"Dockerfile"},
		}
		cpe := buildkitExecuteCommonPipelineEnvironment{}
		runner := &mock.ExecMockRunner{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &cpe, runner, nil, fileUtils)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "no docker files to process, please check exclude list")
	})

	t.Run("error case - multi image build: build failed", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
		}
		cpe := buildkitExecuteCommonPipelineEnvironment{}
		runner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{"docker": fmt.Errorf("execution failed")},
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &cpe, runner, nil, fileUtils)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to build image")
	})

	t.Run("error case - cert update failed", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			CustomTLSCertificateLinks: []string{"https://test.url/cert.crt"},
			DockerfilePath:            "Dockerfile",
		}
		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := buildkitExecuteCommonPipelineEnvironment{}
		certClient := &buildkitMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.FileReadErrors = map[string]error{buildkitTLSCertPath: fmt.Errorf("read error")}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, fmt.Sprintf("failed to update certificates: failed to load file '%s': read error", buildkitTLSCertPath))
	})

	t.Run("error case - multi context build: no subcontext provided", func(t *testing.T) {
		config := &buildkitExecuteOptions{
			ContainerImageName:   "myImage",
			ContainerImageTag:    "myTag",
			ContainerRegistryURL: "https://my.registry.com:50000",
			MultipleImages: []map[string]interface{}{
				{"containerImageName": "myImageOne"},
				{"containerImageName": "myImageTwo"},
			},
		}
		cpe := buildkitExecuteCommonPipelineEnvironment{}
		runner := &mock.ExecMockRunner{}
		fileUtils := &mock.FilesMock{}

		err := runBuildkitExecute(config, &telemetry.CustomData{}, &cpe, runner, nil, fileUtils)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "multipleImages: empty contextSubPath")
	})
}

func TestExtractDigestFromMetadata(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("/tmp/metadata.json", []byte(`{
			"buildx.build.provenance": {},
			"buildx.build.ref": "default/default/abc123",
			"containerimage.config.digest": "sha256:configdigest",
			"containerimage.digest": "sha256:abc123def456"
		}`))

		digest, err := extractDigestFromMetadata("/tmp/metadata.json", fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, "sha256:abc123def456", digest)
	})

	t.Run("error - file not found", func(t *testing.T) {
		t.Parallel()
		fileUtils := &mock.FilesMock{}

		digest, err := extractDigestFromMetadata("/tmp/nonexistent.json", fileUtils)

		assert.Error(t, err)
		assert.Equal(t, "", digest)
	})

	t.Run("error - missing digest key", func(t *testing.T) {
		t.Parallel()
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("/tmp/metadata.json", []byte(`{"buildx.build.ref": "something"}`))

		digest, err := extractDigestFromMetadata("/tmp/metadata.json", fileUtils)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "'containerimage.digest' not found")
		assert.Equal(t, "", digest)
	})

	t.Run("error - invalid JSON", func(t *testing.T) {
		t.Parallel()
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("/tmp/metadata.json", []byte(`not json`))

		digest, err := extractDigestFromMetadata("/tmp/metadata.json", fileUtils)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing metadata JSON")
		assert.Equal(t, "", digest)
	})
}

func TestBuildkitHasDestination(t *testing.T) {
	t.Parallel()

	t.Run("has -t flag", func(t *testing.T) {
		t.Parallel()
		assert.True(t, buildkitHasDestination([]string{"-t", "registry/image:tag", "--some-other-flag"}))
	})

	t.Run("no -t flag", func(t *testing.T) {
		t.Parallel()
		assert.False(t, buildkitHasDestination([]string{"--some-flag", "value"}))
	})

	t.Run("empty options", func(t *testing.T) {
		t.Parallel()
		assert.False(t, buildkitHasDestination([]string{}))
	})
}

func TestBuildkitFindImageNameTagInPurl(t *testing.T) {
	t.Parallel()

	t.Run("exact match found", func(t *testing.T) {
		t.Parallel()
		result := buildkitFindImageNameTagInPurl([]string{"myImage:1.0.0", "anotherImage:2.0.0"}, "myImage:1.0.0")
		assert.Equal(t, "myImage:1.0.0", result)
	})

	t.Run("suffix match found", func(t *testing.T) {
		t.Parallel()
		result := buildkitFindImageNameTagInPurl([]string{"my.registry.com:50000/apps/myImage:1.0.0"}, "apps/myImage:1.0.0")
		assert.Equal(t, "my.registry.com:50000/apps/myImage:1.0.0", result)
	})

	t.Run("no match found", func(t *testing.T) {
		t.Parallel()
		result := buildkitFindImageNameTagInPurl([]string{"myImage:1.0.0"}, "nonExistentImage:3.0.0")
		assert.Equal(t, "", result)
	})

	t.Run("empty container image name tags", func(t *testing.T) {
		t.Parallel()
		result := buildkitFindImageNameTagInPurl([]string{}, "myImage:1.0.0")
		assert.Equal(t, "", result)
	})

	t.Run("multiple matches - exact match takes precedence", func(t *testing.T) {
		t.Parallel()
		result := buildkitFindImageNameTagInPurl([]string{"my.registry.com/myImage:1.0.0", "myImage:1.0.0"}, "myImage:1.0.0")
		assert.Equal(t, "myImage:1.0.0", result)
	})
}
