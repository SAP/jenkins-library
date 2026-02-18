//go:build unit

package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
	transportrequest "github.com/SAP/jenkins-library/pkg/transportrequest/cts"
	"github.com/stretchr/testify/assert"
)

type UploadActionMock struct {
	Connection         transportrequest.Connection
	Application        transportrequest.Application
	Node               transportrequest.Node
	TransportRequestID string
	ConfigFile         string
	DeployUser         string
	thrown             error
}

// WithConnection ...
func (action *UploadActionMock) WithConnection(connection transportrequest.Connection) {
	action.Connection = connection
}

// WithApplication ...
func (action *UploadActionMock) WithApplication(app transportrequest.Application) {
	action.Application = app
}

// WithNodeProperties ...
func (action *UploadActionMock) WithNodeProperties(node transportrequest.Node) {
	action.Node = node
}

// WithTransportRequestID ...
func (action *UploadActionMock) WithTransportRequestID(id string) {
	action.TransportRequestID = id
}

// WithConfigFile ...
func (action *UploadActionMock) WithConfigFile(configFile string) {
	action.ConfigFile = configFile
}

// WithDeployUser ...
func (action *UploadActionMock) WithDeployUser(deployUser string) {
	action.DeployUser = deployUser
}

func (action *UploadActionMock) Perform(cmd command.ShellRunner) error {
	return action.thrown
}

type transportRequestUploadMockUtils struct {
	*mock.ShellMockRunner
}

func newTransportRequestUploadCTSTestsUtils() transportRequestUploadMockUtils {
	utils := transportRequestUploadMockUtils{
		ShellMockRunner: &mock.ShellMockRunner{},
	}
	return utils
}

func TestRunTransportRequestUploadCTS(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		// init
		config := transportRequestUploadCTSOptions{
			Endpoint:               "https://example.org:8000",
			Client:                 "001",
			Username:               "me",
			Password:               "********",
			ApplicationName:        "myApp",
			AbapPackage:            "myPackage",
			Description:            "lorem ipsum",
			TransportRequestID:     "XXXK123456",
			OsDeployUser:           "node",            // default provided in config
			DeployConfigFile:       "ui5-deploy.yaml", // default provided in config
			DeployToolDependencies: []string{"@ui5/cli", "@sap/ux-ui5-tooling"},
			NpmInstallOpts:         []string{"--verbose", "--registry", "https://registry.example.org/"},
		}

		actionMock := &UploadActionMock{thrown: nil}
		cpe := &transportRequestUploadCTSCommonPipelineEnvironment{}
		// test
		err := runTransportRequestUploadCTS(&config, actionMock, nil, newTransportRequestUploadCTSTestsUtils(), cpe)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, &UploadActionMock{
				Connection: transportrequest.Connection{
					Endpoint: "https://example.org:8000",
					Client:   "001",
					User:     "me",
					Password: "********",
				},
				Application: transportrequest.Application{
					Name: "myApp",
					Pack: "myPackage",
					Desc: "lorem ipsum",
				},
				Node: transportrequest.Node{
					DeployDependencies: []string{
						"@ui5/cli",
						"@sap/ux-ui5-tooling",
					},
					InstallOpts: []string{
						"--verbose",
						"--registry",
						"https://registry.example.org/",
					},
				},
				TransportRequestID: "XXXK123456",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "node",
			}, actionMock)
		}
	})

	t.Run("error case", func(t *testing.T) {

		config := transportRequestUploadCTSOptions{
			Endpoint:               "https://example.org:8000",
			Client:                 "001",
			Username:               "me",
			Password:               "********",
			ApplicationName:        "myApp",
			AbapPackage:            "myPackage",
			Description:            "lorem ipsum",
			TransportRequestID:     "XXXK123456",
			OsDeployUser:           "node",            // default provided in config
			DeployConfigFile:       "ui5-deploy.yaml", // default provided in config
			DeployToolDependencies: []string{"@ui5/cli", "@sap/ux-ui5-tooling"},
			NpmInstallOpts:         []string{"--verbose", "--registry", "https://registry.example.org/"},
		}
		cpe := &transportRequestUploadCTSCommonPipelineEnvironment{}

		err := runTransportRequestUploadCTS(
			&config,
			&UploadActionMock{thrown: fmt.Errorf("something went wrong")},
			nil,
			newTransportRequestUploadCTSTestsUtils(),
			cpe)
		assert.EqualError(t, err, "something went wrong")
	})
}
