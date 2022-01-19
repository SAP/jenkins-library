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
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", ".", "--skip-tls-verify-pull", "--destination", "myImage:tag", "--ignore-path", "/busybox"}, runner.Calls[1].Params)

		assert.Contains(t, commonPipelineEnvironment.custom.buildSettingsInfo, `"mavenExecuteBuild":[{"dockerImage":"maven"}]`)
		assert.Contains(t, commonPipelineEnvironment.custom.buildSettingsInfo, `"kanikoExecute":[{"dockerImage":"gcr.io/kaniko-project/executor:debug"}]`)
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
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", ".", "--skip-tls-verify-pull", "--destination", "my.registry.com:50000/myImage:1.2.3-a-x", "--ignore-path", "/busybox"}, runner.Calls[1].Params)

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

		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", ".", "--skip-tls-verify-pull", "--no-push", "--ignore-path", "/busybox"}, runner.Calls[1].Params)
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
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", ".", "--skip-tls-verify-pull", "--destination", "myImage:tag", "--ignore-path", "/busybox"}, runner.Calls[1].Params)
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

		expectedParams := [][]string{
			{"--dockerfile", "Dockerfile", "--context", ".", "--destination", "my.registry.com:50000/myImage:myTag", "--ignore-path", "/busybox"},
			{"--dockerfile", filepath.Join("sub1", "Dockerfile"), "--context", "sub1", "--destination", "my.registry.com:50000/myImage-sub1:myTag", "--ignore-path", "/busybox"},
			{"--dockerfile", filepath.Join("sub2", "Dockerfile"), "--context", "sub2", "--destination", "my.registry.com:50000/myImage-sub2:myTag", "--ignore-path", "/busybox"},
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
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNames, "myImage")
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNames, "myImage-sub1")
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNames, "myImage-sub2")
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNameTags, "myImage:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNameTags, "myImage-sub1:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNameTags, "myImage-sub2:myTag")
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

		expectedParams := [][]string{
			{"--dockerfile", filepath.Join("sub1", "Dockerfile"), "--context", "sub1", "--destination", "my.registry.com:50000/myImage-sub1:myTag", "--ignore-path", "/busybox"},
			{"--dockerfile", filepath.Join("sub2", "Dockerfile"), "--context", "sub2", "--destination", "my.registry.com:50000/myImage-sub2:myTag", "--ignore-path", "/busybox"},
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
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNames, "myImage-sub1")
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNames, "myImage-sub2")
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNameTags, "myImage-sub1:myTag")
		assert.Contains(t, commonPipelineEnvironment.container.multiImageNameTags, "myImage-sub2:myTag")
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

		assert.EqualError(t, err, "failed to read file 'path/to/docker/config.json': read error")
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
