package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

type kanikoFileMock struct {
	fileReadContent  map[string]string
	fileReadErr      map[string]error
	fileWriteContent map[string]string
	fileWriteErr     map[string]error
}

func (f *kanikoFileMock) FileExists(path string) (bool, error) {
	return true, nil
}

func (f *kanikoFileMock) Copy(src, dest string) (int64, error) {
	return 0, nil
}

func (f *kanikoFileMock) FileRead(path string) ([]byte, error) {
	if f.fileReadErr[path] != nil {
		return []byte{}, f.fileReadErr[path]
	}
	return []byte(f.fileReadContent[path]), nil
}

func (f *kanikoFileMock) FileWrite(path string, content []byte, perm os.FileMode) error {
	if f.fileWriteErr[path] != nil {
		return f.fileWriteErr[path]
	}
	f.fileWriteContent[path] = string(content)
	return nil
}

func (f *kanikoFileMock) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (f *kanikoFileMock) Chmod(path string, mode os.FileMode) error {
	return fmt.Errorf("not implemented. func is only present in order to fullfil the interface contract. Needs to be ajusted in case it gets used.")
}

func (f *kanikoFileMock) Abs(path string) (string, error) {
	return "", fmt.Errorf("not implemented. func is only present in order to fullfil the interface contract. Needs to be ajusted in case it gets used.")
}

func (f *kanikoFileMock) Glob(pattern string) (matches []string, err error) {
	return nil, fmt.Errorf("not implemented. func is only present in order to fullfil the interface contract. Needs to be ajusted in case it gets used.")
}

func TestRunKanikoExecute(t *testing.T) {

	commonPipelineEnvironment := kanikoExecuteCommonPipelineEnvironment{}

	t.Run("success case", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			BuildOptions:                []string{"--skip-tls-verify-pull"},
			ContainerImage:              "myImage:tag",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
			DockerConfigJSON:            "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &kanikoFileMock{
			fileReadContent:  map[string]string{"path/to/docker/config.json": `{"auths":{"custom":"test"}}`},
			fileWriteContent: map[string]string{},
		}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", runner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, runner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, fileUtils.fileWriteContent["/kaniko/.docker/config.json"])

		assert.Equal(t, "/kaniko/executor", runner.Calls[1].Exec)
		cwd, _ := os.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "myImage:tag"}, runner.Calls[1].Params)

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

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &kanikoFileMock{
			fileReadContent:  map[string]string{"path/to/docker/config.json": `{"auths":{"custom":"test"}}`},
			fileWriteContent: map[string]string{},
		}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, "rm", runner.Calls[0].Exec)
		assert.Equal(t, []string{"-f", "/kaniko/.docker/config.json"}, runner.Calls[0].Params)

		assert.Equal(t, config.CustomTLSCertificateLinks, certClient.urlsCalled)
		assert.Equal(t, `{"auths":{"custom":"test"}}`, fileUtils.fileWriteContent["/kaniko/.docker/config.json"])

		assert.Equal(t, "/kaniko/executor", runner.Calls[1].Exec)
		cwd, _ := os.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "my.registry.com:50000/myImage:1.2.3-a-x"}, runner.Calls[1].Params)

	})

	t.Run("success case - no push, no docker config.json", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerBuildOptions:       "--skip-tls-verify-pull",
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
			CustomTLSCertificateLinks:   []string{"https://test.url/cert.crt"},
			DockerfilePath:              "Dockerfile",
		}

		runner := &mock.ExecMockRunner{}

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &kanikoFileMock{
			fileWriteContent: map[string]string{},
		}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)

		assert.Equal(t, `{"auths":{}}`, fileUtils.fileWriteContent["/kaniko/.docker/config.json"])

		cwd, _ := os.Getwd()
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

		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &kanikoFileMock{
			fileReadContent:  map[string]string{"path/to/docker/config.json": `{"auths":{"custom":"test"}}`},
			fileWriteContent: map[string]string{},
		}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.NoError(t, err)
		cwd, _ := os.Getwd()
		assert.Equal(t, []string{"--dockerfile", "Dockerfile", "--context", cwd, "--skip-tls-verify-pull", "--destination", "myImage:tag"}, runner.Calls[1].Params)
	})

	t.Run("error case - Kaniko init failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			ContainerPreparationCommand: "rm -f /kaniko/.docker/config.json",
		}

		runner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{"rm": fmt.Errorf("rm failed")},
		}

		certClient := &kanikoMockClient{}
		fileUtils := &kanikoFileMock{}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to initialize Kaniko container: rm failed")
	})

	t.Run("error case - Kaniko execution failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{}

		runner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{"/kaniko/executor": fmt.Errorf("kaniko run failed")},
		}

		certClient := &kanikoMockClient{}
		fileUtils := &kanikoFileMock{
			fileWriteContent: map[string]string{},
		}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "execution of '/kaniko/executor' failed: kaniko run failed")
	})

	t.Run("error case - cert update failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{}

		runner := &mock.ExecMockRunner{}

		certClient := &kanikoMockClient{}
		fileUtils := &kanikoFileMock{
			fileWriteContent: map[string]string{},
			fileReadErr:      map[string]error{"/kaniko/ssl/certs/ca-certificates.crt": fmt.Errorf("read error")},
		}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to update certificates: failed to load file '/kaniko/ssl/certs/ca-certificates.crt': read error")
	})

	t.Run("error case - dockerconfig read failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			DockerConfigJSON: "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}

		certClient := &kanikoMockClient{}
		fileUtils := &kanikoFileMock{
			fileWriteContent: map[string]string{},
			fileReadErr:      map[string]error{"path/to/docker/config.json": fmt.Errorf("read error")},
		}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to read file 'path/to/docker/config.json': read error")
	})

	t.Run("error case - dockerconfig write failed", func(t *testing.T) {
		config := &kanikoExecuteOptions{
			DockerConfigJSON: "path/to/docker/config.json",
		}

		runner := &mock.ExecMockRunner{}

		certClient := &kanikoMockClient{}
		fileUtils := &kanikoFileMock{
			fileWriteContent: map[string]string{},
			fileWriteErr:     map[string]error{"/kaniko/.docker/config.json": fmt.Errorf("write error")},
		}

		err := runKanikoExecute(config, &telemetry.CustomData{}, &commonPipelineEnvironment, runner, certClient, fileUtils)

		assert.EqualError(t, err, "failed to write file '/kaniko/.docker/config.json': write error")
	})

}

func TestCertificateUpdate(t *testing.T) {
	certLinks := []string{"https://my.first/cert.crt", "https://my.second/cert.crt"}

	t.Run("success case", func(t *testing.T) {
		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &kanikoFileMock{
			fileReadContent:  map[string]string{"/kaniko/ssl/certs/ca-certificates.crt": "initial cert\n"},
			fileWriteContent: map[string]string{},
		}

		err := certificateUpdate(certLinks, certClient, fileUtils)

		assert.NoError(t, err)
		assert.Equal(t, certLinks, certClient.urlsCalled)
		assert.Equal(t, "initial cert\ntestCert\ntestCert\n", fileUtils.fileWriteContent["/kaniko/ssl/certs/ca-certificates.crt"])
	})

	t.Run("error case - read certs", func(t *testing.T) {
		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &kanikoFileMock{
			fileReadErr: map[string]error{"/kaniko/ssl/certs/ca-certificates.crt": fmt.Errorf("read error")},
		}

		err := certificateUpdate(certLinks, certClient, fileUtils)
		assert.EqualError(t, err, "failed to load file '/kaniko/ssl/certs/ca-certificates.crt': read error")
	})

	t.Run("error case - write certs", func(t *testing.T) {
		certClient := &kanikoMockClient{
			responseBody: "testCert",
		}
		fileUtils := &kanikoFileMock{
			fileReadContent: map[string]string{"/kaniko/ssl/certs/ca-certificates.crt": "initial cert\n"},
			fileWriteErr:    map[string]error{"/kaniko/ssl/certs/ca-certificates.crt": fmt.Errorf("write error")},
		}

		err := certificateUpdate(certLinks, certClient, fileUtils)
		assert.EqualError(t, err, "failed to update file '/kaniko/ssl/certs/ca-certificates.crt': write error")
	})

	t.Run("error case - get cert via http", func(t *testing.T) {
		certClient := &kanikoMockClient{
			responseBody: "testCert",
			errorMessage: "http error",
		}
		fileUtils := &kanikoFileMock{
			fileReadContent:  map[string]string{"/kaniko/ssl/certs/ca-certificates.crt": "initial cert\n"},
			fileWriteContent: map[string]string{},
		}

		err := certificateUpdate(certLinks, certClient, fileUtils)
		assert.EqualError(t, err, "failed to load certificate from url: http error")
	})

}
