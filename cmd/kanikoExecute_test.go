package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

type kanikoMockClient struct {
	httpMethod     string
	httpStatusCode int
	urlsCalled     []string
	requestBody    io.Reader
	responseBody   string
	errorMessage   string
}

func (c *kanikoMockClient) SetOptions(opts piperhttp.ClientOptions) {}

func (c *kanikoMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlsCalled = append(c.urlsCalled, url)
	c.requestBody = body
	if len(c.errorMessage) > 0 {
		return nil, fmt.Errorf(c.errorMessage)
	}
	return &http.Response{StatusCode: c.httpStatusCode, Body: ioutil.NopCloser(bytes.NewReader([]byte(c.responseBody)))}, nil
}

func TestRunKanikoExecute(t *testing.T) {

	// required due to config resolution during build settings retrieval
	// ToDo: proper mocking
	openFileBak := configOptions.openFile
	defer func() {
		configOptions.openFile = openFileBak
	}()

	configOptions.openFile = configOpenFileMock

	t.Run("success case", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull"},
			ContainerImage:              "myImage:tag",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
			BuildSettingsInfo:           `{"mavenExecuteBuild":[{"dockerImage":"maven"}]}`,
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", runner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, runner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		assert.Equal(t, "/kaniko/executor", runner.Calls[1].Exec)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "myImage:tag"}, runner.Calls[1].Params)

		assert.Contains(t, commonPipelineEnvironment.custom.buildSettingsInfo, `"mavenExecuteBuild":[{"dockerImage":"maven"}]`)
		assert.Contains(t, commonPipelineEnvironment.custom.buildSettingsInfo, `"kanikoExecute":[{"dockerImage":"gcr.io/kaniko-project/executor:debug"}]`)

		assert.Equal(t, "myImage:tag", commonPipelineEnvironment.container.imageNameTag)
		assert.Equal(t, "https://index.docker.io", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, []string{"myImage"}, commonPipelineEnvironment.container.imageNames)
		assert.Equal(t, []string{"myImage:tag"}, commonPipelineEnvironment.container.imageNameTags)

		assert.Equal(t, "", commonPipelineEnvironment.container.imageDigest)
		assert.Empty(t, commonPipelineEnvironment.container.imageDigests)
	})

	t.Run("success case - pass image digest to cpe if activated", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull"},
			ContainerImage:              "myImage:tag",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
			BuildSettingsInfo:           `{"mavenExecuteBuild":[{"dockerImage":"maven"}]}`,
			ReadImageDigest:             true,
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))
		fileUtils.AddFile("/tmp/*-kanikoExecutetest/digest.txt", []byte(`sha256:468dd1253cc9f498fc600454bb8af96d880fec3f9f737e7057692adfe9f7d5b0`))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", runner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, runner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		assert.Equal(t, "/kaniko/executor", runner.Calls[1].Exec)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "myImage:tag", "--digest-file", "/tmp/*-kanikoExecutetest/digest.txt"}, runner.Calls[1].Params)

		assert.Contains(t, commonPipelineEnvironment.custom.buildSettingsInfo, `"mavenExecuteBuild":[{"dockerImage":"maven"}]`)
		assert.Contains(t, commonPipelineEnvironment.custom.buildSettingsInfo, `"kanikoExecute":[{"dockerImage":"gcr.io/kaniko-project/executor:debug"}]`)

		assert.Equal(t, "myImage:tag", commonPipelineEnvironment.container.imageNameTag)
		assert.Equal(t, "https://index.docker.io", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, []string{"myImage"}, commonPipelineEnvironment.container.imageNames)
		assert.Equal(t, []string{"myImage:tag"}, commonPipelineEnvironment.container.imageNameTags)

		assert.Equal(t, "sha256:468dd1253cc9f498fc600454bb8af96d880fec3f9f737e7057692adfe9f7d5b0", commonPipelineEnvironment.container.imageDigest)
		assert.Equal(t, []string{"sha256:468dd1253cc9f498fc600454bb8af96d880fec3f9f737e7057692adfe9f7d5b0"}, commonPipelineEnvironment.container.imageDigests)
	})

	t.Run("success case - image params", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull"},
			ContainerImageName:          "myImage",
			ContainerImageTag:           "1.2.3-a+x",
			ContainerRegistryURL:        "https://my.registry.com:50000",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", runner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, runner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		assert.Equal(t, "/kaniko/executor", runner.Calls[1].Exec)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "my.registry.com:50000/myImage:1.2.3-a-x"}, runner.Calls[1].Params)

		assert.Equal(t, "myImage:1.2.3-a-x", commonPipelineEnvironment.container.imageNameTag)
		assert.Equal(t, "https://my.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, []string{"myImage"}, commonPipelineEnvironment.container.imageNames)
		assert.Equal(t, []string{"myImage:1.2.3-a-x"}, commonPipelineEnvironment.container.imageNameTags)

		assert.Equal(t, "", commonPipelineEnvironment.container.imageDigest)
		assert.Empty(t, commonPipelineEnvironment.container.imageDigests)
	})

	t.Run("success case - image params with custom destination", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull", "--destination", "my.other.registry.com:50000/myImage:3.2.1-a-x"},
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", runner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, runner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		assert.Equal(t, "/kaniko/executor", runner.Calls[1].Exec)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "my.other.registry.com:50000/myImage:3.2.1-a-x"}, runner.Calls[1].Params)

		assert.Equal(t, "myImage:3.2.1-a-x", commonPipelineEnvironment.container.imageNameTag)
		assert.Equal(t, "https://my.other.registry.com:50000", commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, []string{"myImage"}, commonPipelineEnvironment.container.imageNames)
		assert.Equal(t, []string{"myImage:3.2.1-a-x"}, commonPipelineEnvironment.container.imageNameTags)

		assert.Equal(t, "", commonPipelineEnvironment.container.imageDigest)
		assert.Empty(t, []string{}, commonPipelineEnvironment.container.imageDigests)
	})

	t.Run("no error case - when cert update skipped", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull"},
			ContainerImageName:          "myImage",
			ContainerImageTag:           "1.2.3-a+x",
			ContainerRegistryURL:        "https://my.registry.com:50000",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(``))
		fileUtils.FileReadErrors = map[string]error{"/kaniko/ssl/certs/ca-certificates.crt": fmt.Errorf("read error")}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoErrorf(t, err, "failed to update certificates: failed to load file '/kaniko/ssl/certs/ca-certificates.crt': read error")
	})

	t.Run("success case - no push, no docker config.json", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerBuildOptions:       "--skip-tls-verify-pull",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{}}`, string(c))

		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--no-push"}, runner.Calls[1].Params)
	})

	t.Run("success case - backward compatibility", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerBuildOptions:       "--skip-tls-verify-pull",
			ContainerImage:              "myImage:tag",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "myImage:tag"}, runner.Calls[1].Params)
	})

	t.Run("success case - multi image build with root image", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub1/Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub2/Dockerfile", []byte("some content"))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, 3, len(runner.Calls))
		assert.Equal(t, "/kaniko/executor", runner.Calls[0].Exec)
		assert.Equal(t, "/kaniko/executor", runner.Calls[1].Exec)
		assert.Equal(t, "/kaniko/executor", runner.Calls[2].Exec)

		cwd, _ := fileUtils.Getwd()
		expectedParams := [][]string{
			{"--dockerfile", "Dockerfile", "--context", cwd, "--destination", "my.registry.com:50000/myImage:myTag"},
			{"--dockerfile", filepath.Join("sub1", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub1:myTag"},
			{"--dockerfile", filepath.Join("sub2", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub2:myTag"},
		}
		// need to go this way since we cannot count on the correct order
		for _, call := range runner.Calls {
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

		assert.Equal(t, "", commonPipelineEnvironment.container.imageDigest)
		assert.Empty(t, commonPipelineEnvironment.container.imageDigests)
	})

	t.Run("success case - multi image build excluding root image", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImageName:               "myImage",
			ContainerImageTag:                "myTag",
			ContainerRegistryURL:             "https://my.registry.com:50000",
			ContainerMultiImageBuild:         true,
			ContainerMultiImageBuildExcludes: []string{"Dockerfile"},
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub1/Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub2/Dockerfile", []byte("some content"))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, nil, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, 2, len(runner.Calls))
		assert.Equal(t, "/kaniko/executor", runner.Calls[0].Exec)
		assert.Equal(t, "/kaniko/executor", runner.Calls[1].Exec)

		cwd, _ := fileUtils.Getwd()
		expectedParams := [][]string{
			{"--dockerfile", filepath.Join("sub1", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub1:myTag"},
			{"--dockerfile", filepath.Join("sub2", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub2:myTag"},
		}
		// need to go this way since we cannot count on the correct order
		for _, call := range runner.Calls {
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
		assert.Equal(t, "", commonPipelineEnvironment.container.imageNameTag)
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage-sub1")
		assert.Contains(t, commonPipelineEnvironment.container.imageNames, "myImage-sub2")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage-sub1:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.imageNameTags, "myImage-sub2:myTag")
	})

	t.Run("success case - updating an existing docker config json with addtional credentials", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull"},
			ContainerImageName:          "myImage",
			ContainerImageTag:           "1.2.3-a+x",
			ContainerRegistryURL:        "https://my.registry.com:50000",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
			ContainerRegistryUser:       "dummyUser",
			ContainerRegistryPassword:   "dummyPassword",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths": {"dummyUrl": {"auth": "XXXXXXX"}}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", runner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, runner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"dummyUrl":{"auth":"XXXXXXX"},"https://my.registry.com:50000":{"auth":"ZHVtbXlVc2VyOmR1bW15UGFzc3dvcmQ="}}}`, string(c))
	})

	t.Run("success case - creating new docker config json with provided container credentials", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull"},
			ContainerImageName:          "myImage",
			ContainerImageTag:           "1.2.3-a+x",
			ContainerRegistryURL:        "https://my.registry.com:50000",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			ContainerRegistryUser:       "dummyUser",
			ContainerRegistryPassword:   "dummyPassword",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths": {"dummyUrl": {"auth": "XXXXXXX"}}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", runner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, runner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"https://my.registry.com:50000":{"auth":"ZHVtbXlVc2VyOmR1bW15UGFzc3dvcmQ="}}}`, string(c))
	})

	t.Run("error case - multi image build: no docker files", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
		}

		cpe := kanikoExecuteCommonPipelineEnvironment{}
		runner := &mock.ExecMockRunner{}

		fileUtils := &mock.FilesMock{}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &cpe, runner, nil, fileUtils)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to identify image list for multi image build")
	})

	t.Run("error case - multi image build: no docker files to process", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImageName:               "myImage",
			ContainerImageTag:                "myTag",
			ContainerRegistryURL:             "https://my.registry.com:50000",
			ContainerMultiImageBuild:         true,
			ContainerMultiImageBuildExcludes: []string{"Dockerfile"},
		}

		cpe := kanikoExecuteCommonPipelineEnvironment{}
		runner := &mock.ExecMockRunner{}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &cpe, runner, nil, fileUtils)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "no docker files to process, please check exclude list")
	})

	t.Run("error case - multi image build: build failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
		}

		cpe := kanikoExecuteCommonPipelineEnvironment{}
		runner := &mock.ExecMockRunner{}
		runner.ShouldFailOnCommand = map[string]error{"/kaniko/executor": fmt.Errorf("execution failed")}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &cpe, runner, nil, fileUtils)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to build image")
	})

	t.Run("error case - Kaniko init failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
		}

		runner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{"rm": fmt.Errorf("rm failed")},
		}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to initialize Kaniko container: rm failed")
	})

	t.Run("error case - Kaniko execution failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{}

		runner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{"/kaniko/executor": fmt.Errorf("kaniko run failed")},
		}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "execution of '/kaniko/executor' failed: kaniko run failed")
	})

	t.Run("error case - cert update failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull"},
			ContainerImageName:          "myImage",
			ContainerImageTag:           "1.2.3-a+x",
			ContainerRegistryURL:        "https://my.registry.com:50000",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.FileReadErrors = map[string]error{"/kaniko/ssl/certs/ca-certificates.crt": fmt.Errorf("read error")}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to update certificates: failed to load file '/kaniko/ssl/certs/ca-certificates.crt': read error")
	})

	t.Run("error case - dockerconfig read failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			DockerConfigJSON: "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.FileReadErrors = map[string]error{"path/to/docker/config.json": fmt.Errorf("read error")}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to read existing docker config json at 'path/to/docker/config.json': read error")
	})

	t.Run("error case - dockerconfig write failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			DockerConfigJSON: "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.FileWriteErrors = map[string]error{"/kaniko/.docker/config.json": fmt.Errorf("write error")}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to write file '/kaniko/.docker/config.json': write error")
	})

}
