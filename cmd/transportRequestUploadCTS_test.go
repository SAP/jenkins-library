package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/transportrequest"
	"github.com/stretchr/testify/assert"
	"testing"
)

type CTSUploadActionMock struct {
	Connection         transportrequest.CTSConnection
	Application        transportrequest.CTSApplication
	Node               transportrequest.CTSNode
	TransportRequestID string
	ConfigFile         string
	DeployUser         string

	thrown error
}

// WithConnection ...
func (action *CTSUploadActionMock) WithConnection(connection transportrequest.CTSConnection) {
	action.Connection = connection
}

// WithApplication ...
func (action *CTSUploadActionMock) WithApplication(app transportrequest.CTSApplication) {
	action.Application = app
}

// WithNodeProperties ...
func (action *CTSUploadActionMock) WithNodeProperties(node transportrequest.CTSNode) {
	action.Node = node
}

// WithTransportRequestID ...
func (action *CTSUploadActionMock) WithTransportRequestID(id string) {
	action.TransportRequestID = id
}

// WithConfigFile ...
func (action *CTSUploadActionMock) WithConfigFile(configFile string) {
	action.ConfigFile = configFile
}

// WithDeployUser ...
func (action *CTSUploadActionMock) WithDeployUser(deployUser string) {
	action.DeployUser = deployUser
}

func (action *CTSUploadActionMock) Perform(cmd command.ShellRunner) error {
	return action.thrown
}

type transportRequestUploadMockUtils struct {
	*mock.ShellMockRunner
	*ActionProvider
}

func newTransportRequestUploadCTSTestsUtils(thrown error) transportRequestUploadMockUtils {
	utils := transportRequestUploadMockUtils{
		ShellMockRunner: &mock.ShellMockRunner{},
		ActionProvider:  &ActionProvider{action: &CTSUploadActionMock{thrown: thrown}},
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

		utils := newTransportRequestUploadCTSTestsUtils(nil)

		// test
		err := runTransportRequestUploadCTS(&config, nil, utils)

		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, &CTSUploadActionMock{
				Connection: transportrequest.CTSConnection{
					Endpoint: "https://example.org:8000",
					Client:   "001",
					User:     "me",
					Password: "********",
				},
				Application: transportrequest.CTSApplication{
					Name: "myApp",
					Pack: "myPackage",
					Desc: "lorem ipsum",
				},
				Node: transportrequest.CTSNode{
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
			}, utils.GetAction())
		}
	})

	t.Run("error case", func(t *testing.T) {
		utils := newTransportRequestUploadCTSTestsUtils(fmt.Errorf("something went wrong"))

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

		err := runTransportRequestUploadCTS(&config, nil, utils)
		assert.EqualError(t, err, "something went wrong")
	})
}
