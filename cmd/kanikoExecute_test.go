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
func (c *kanikoMockClient) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	if len(c.errorMessage) > 0 {
		return fmt.Errorf(c.errorMessage)
	}
	return nil
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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", execRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, execRunner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		assert.Equal(t, "/kaniko/executor", execRunner.Calls[1].Exec)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "myImage:tag"}, execRunner.Calls[1].Params)

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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))
		fileUtils.AddFile("/tmp/*-kanikoExecutetest/digest.txt", []byte(`sha256:468dd1253cc9f498fc600454bb8af96d880fec3f9f737e7057692adfe9f7d5b0`))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", execRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, execRunner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		assert.Equal(t, "/kaniko/executor", execRunner.Calls[1].Exec)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "myImage:tag", "--digest-file", "/tmp/*-kanikoExecutetest/digest.txt"}, execRunner.Calls[1].Params)

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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", execRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, execRunner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		assert.Equal(t, "/kaniko/executor", execRunner.Calls[1].Exec)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "my.registry.com:50000/myImage:1.2.3-a-x"}, execRunner.Calls[1].Params)

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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", execRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, execRunner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, string(c))

		assert.Equal(t, "/kaniko/executor", execRunner.Calls[1].Exec)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "my.other.registry.com:50000/myImage:3.2.1-a-x"}, execRunner.Calls[1].Params)

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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(``))
		fileUtils.FileReadErrors = map[string]error{"/kaniko/ssl/certs/ca-certificates.crt": fmt.Errorf("read error")}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoErrorf(t, err, "failed to update certificates: failed to load file '/kaniko/ssl/certs/ca-certificates.crt': read error")
	})

	t.Run("success case - no push, no docker config.json", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerBuildOptions:       "--skip-tls-verify-pull",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
		}

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)

		c, err := fileUtils.FileRead("/kaniko/.docker/config.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{}}`, string(c))

		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--no-push"}, execRunner.Calls[1].Params)
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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)
		cwd, _ := fileUtils.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "myImage:tag"}, execRunner.Calls[1].Params)
	})
	t.Run("success case - createBOM", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImage:              "myImage:tag",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
			CreateBOM:                   true,
		}

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("Syft/syft", []byte(`echo syft`))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "/kaniko/executor", execRunner.Calls[1].Exec)

		assert.Equal(t, "myImage:tag", commonPipelineEnvironment.container.imageNameTag)
		assert.Equal(t, "https://index.docker.io", commonPipelineEnvironment.container.registryURL)

		//Syft install and call
		assert.Contains(t, shellRunner.Calls[0], "mkdir Syft && tar -zxvf syftBinary.tar.gz -C Syft")
		assert.Contains(t, shellRunner.Calls[1], "Syft/syft index.docker.io/myImage:tag -o cyclonedx-xml=bom-docker-0.xml")
	})

	t.Run("success case - multi image build with root image", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
		}

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub1/Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub2/Dockerfile", []byte("some content"))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, nil, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, 3, len(execRunner.Calls))
		assert.Equal(t, "/kaniko/executor", execRunner.Calls[0].Exec)
		assert.Equal(t, "/kaniko/executor", execRunner.Calls[1].Exec)
		assert.Equal(t, "/kaniko/executor", execRunner.Calls[2].Exec)

		cwd, _ := fileUtils.Getwd()
		expectedParams := [][]string{
			{"--dockerfile", "Dockerfile", "--context", cwd, "--destination", "my.registry.com:50000/myImage:myTag"},
			{"--dockerfile", filepath.Join("sub1", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub1:myTag"},
			{"--dockerfile", filepath.Join("sub2", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub2:myTag"},
		}
		// need to go this way since we cannot count on the correct order
		for _, call := range execRunner.Calls {
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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub1/Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub2/Dockerfile", []byte("some content"))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, nil, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, 2, len(execRunner.Calls))
		assert.Equal(t, "/kaniko/executor", execRunner.Calls[0].Exec)
		assert.Equal(t, "/kaniko/executor", execRunner.Calls[1].Exec)

		cwd, _ := fileUtils.Getwd()
		expectedParams := [][]string{
			{"--dockerfile", filepath.Join("sub1", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub1:myTag"},
			{"--dockerfile", filepath.Join("sub2", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub2:myTag"},
		}
		// need to go this way since we cannot count on the correct order
		for _, call := range execRunner.Calls {
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

	t.Run("success case - multi image build with CreateBOM", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImageName:       "myImage",
			ContainerImageTag:        "myTag",
			ContainerRegistryURL:     "https://my.registry.com:50000",
			ContainerMultiImageBuild: true,
			DockerConfigJSON:         "path/to/docker/config.json",
			CreateBOM:                true,
		}
		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub1/Dockerfile", []byte("some content"))
		fileUtils.AddFile("sub2/Dockerfile", []byte("some content"))
		fileUtils.AddFile("Syft/syft", []byte(`echo syft`))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(execRunner.Calls))
		assert.Equal(t, "/kaniko/executor", execRunner.Calls[0].Exec)
		assert.Equal(t, "/kaniko/executor", execRunner.Calls[1].Exec)
		assert.Equal(t, "/kaniko/executor", execRunner.Calls[2].Exec)

		cwd, _ := fileUtils.Getwd()
		expectedParams := [][]string{
			{"--dockerfile", "Dockerfile", "--context", cwd, "--destination", "my.registry.com:50000/myImage:myTag"},
			{"--dockerfile", filepath.Join("sub1", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub1:myTag"},
			{"--dockerfile", filepath.Join("sub2", "Dockerfile"), "--context", cwd, "--destination", "my.registry.com:50000/myImage-sub2:myTag"},
		}
		// need to go this way since we cannot count on the correct order
		for _, call := range execRunner.Calls {
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

		//Syft install and call, can we do it better without 2 for loops?
		expectedShellCalls := []string{
			"mkdir Syft && tar -zxvf syftBinary.tar.gz -C Syft",
			"Syft/syft my.registry.com:50000/myImage:myTag -o cyclonedx-xml=bom-docker-0.xml",
			"Syft/syft my.registry.com:50000/myImage-sub1:myTag -o cyclonedx-xml=bom-docker-1.xml",
			"Syft/syft my.registry.com:50000/myImage-sub2:myTag -o cyclonedx-xml=bom-docker-2.xml",
		}
		for _, call := range shellRunner.Calls {
			found := false
			for _, expected := range expectedShellCalls {
				if strings.Contains(call, expected) {
					found = true
					break
				}
			}
			assert.True(t, found)
		}

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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths": {"dummyUrl": {"auth": "XXXXXXX"}}}`))
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", execRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, execRunner.Calls[0].Params)

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

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("/kaniko/ssl/certs/ca-certificates.crt", []byte(``))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", execRunner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, execRunner.Calls[0].Params)

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
		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}

		fileUtils := &mock.FilesMock{}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &cpe, execRunner, shellRunner, nil, fileUtils)

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
		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &cpe, execRunner, shellRunner, nil, fileUtils)

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
		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}

		execRunner.ShouldFailOnCommand = map[string]error{"/kaniko/executor": fmt.Errorf("execution failed")}

		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("Dockerfile", []byte("some content"))

		err := runKanikoExecute(config, &telemetry.CustomData{}, &cpe, execRunner, shellRunner, nil, fileUtils)

		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to build image")
	})

	t.Run("error case - Kaniko init failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
		}
		shellRunner := &mock.ShellMockRunner{}
		execRunner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{"rm": fmt.Errorf("rm failed")},
		}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to initialize Kaniko container: rm failed")
	})

	t.Run("error case - Kaniko execution failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{}
		shellRunner := &mock.ShellMockRunner{}
		execRunner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{"/kaniko/executor": fmt.Errorf("kaniko run failed")},
		}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

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
		shellRunner := &mock.ShellMockRunner{}
		execRunner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.FileReadErrors = map[string]error{"/kaniko/ssl/certs/ca-certificates.crt": fmt.Errorf("read error")}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to update certificates: failed to load file '/kaniko/ssl/certs/ca-certificates.crt': read error")
	})

	t.Run("error case - dockerconfig read failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			DockerConfigJSON: "path/to/docker/config.json",
		}
		shellRunner := &mock.ShellMockRunner{}
		execRunner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.FileReadErrors = map[string]error{"path/to/docker/config.json": fmt.Errorf("read error")}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to read existing docker config json at 'path/to/docker/config.json': read error")
	})

	t.Run("error case - dockerconfig write failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			DockerConfigJSON: "path/to/docker/config.json",
		}
		shellRunner := &mock.ShellMockRunner{}
		execRunner := &mock.ExecMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

		certClient := &kanikoMockClient{}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.FileWriteErrors = map[string]error{"/kaniko/.docker/config.json": fmt.Errorf("write error")}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to write file '/kaniko/.docker/config.json': write error")
	})

	t.Run("error cases - createBOM", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerImage:              "myImage:tag",
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CreateBOM:                   true,
		}

		execRunner := &mock.ExecMockRunner{}
		shellRunner := &mock.ShellMockRunner{}
		commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}
		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &mock.FilesMock{}
		fileUtils.AddFile("path/to/docker/config.json", []byte(`{"auths":{"custom":"test"}}`))
		fileUtils.AddFile("Syft/syft", []byte(`echo syft`))

		// Case 1 - Download of syft installation file failed
		certClient.errorMessage = "Download failed"
		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient, fileUtils)
		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), certClient.errorMessage)

		// Case 2 - Installation of syft failed, using new kanikoMockClient here
		certClient1 := &kanikoMockClient{}
		shellRunner.ShouldFailOnCommand = map[string]error{"mkdir Syft && tar -zxvf syftBinary.tar.gz -C Syft": fmt.Errorf("untar failed")}
		err = runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner, certClient1, fileUtils)
		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to untar syft")

		// Case 3 syft run failed, using new ShellMockRunner here
		shellRunner1 := &mock.ShellMockRunner{}
		shellRunner1.ShouldFailOnCommand = map[string]error{"Syft/syft index.docker.io/myImage:tag -o cyclonedx-xml=bom-docker-0.xml": fmt.Errorf("run failed")}
		err = runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, execRunner, shellRunner1, certClient1, fileUtils)
		assert.Error(t, err)
		assert.Contains(t, fmt.Sprint(err), "failed to generate SBOM")
	})
}
