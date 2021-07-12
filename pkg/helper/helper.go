package helper

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/http"
)

// GithubCredentialsConfig defines the credentials for the github authentication
type GithubCredentialsConfig struct {
	Username string
	Token    string
}

// GithubCredentials stores github enterprise auth
var GithubCredentials *GithubCredentialsConfig

// GetFileContent reads and unmarshals any file locally or from web
func GetFileContent(file string, githubAuth *GithubCredentialsConfig) (string, error) {
	fileReadCloser, err := OpenFile(file, githubAuth)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open file file %v", file)
	}
	defer fileReadCloser.Close()

	fileData, err := ioutil.ReadAll(fileReadCloser)
	if err != nil {
		return "", errors.Wrapf(err, "failed to load file %v", file)
	}
	return string(fileData), nil
}

// OpenFile opens file from web if path is a link
func OpenFile(name string, githubAuth *GithubCredentialsConfig) (io.ReadCloser, error) {
	if !strings.HasPrefix(name, "http://") && !strings.HasPrefix(name, "https://") {
		return os.Open(name)
	}
	client := http.Client{}
	if strings.Contains(name, "wdf.sap.corp") && githubAuth != nil {
		options := &http.ClientOptions{Username: githubAuth.Username, Password: githubAuth.Token}
		client.SetOptions(*options)
	}
	response, err := client.SendRequest("GET", name, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return response.Body, nil
}
