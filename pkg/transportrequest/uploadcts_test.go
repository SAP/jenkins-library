package transportrequest

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUploadCTS(t *testing.T) {

	filesMock := mock.FilesMock{}
	files = &filesMock
	defer func() { files = piperutils.Files{} }()

	t.Run("npm install command tests", func(t *testing.T) {
		cmd := mock.ShellMockRunner{}
		action := CTSUploadAction{
			Connection:  CTSConnection{Endpoint: "", Client: "", User: "me", Password: "******"},
			Application: CTSApplication{Pack: "", Name: "", Desc: ""},
			Node: CTSNode{
				DeployDependencies: []string{"@sap/my-dep"},
				InstallOpts:        []string{"--verbose", "--registry", "https://registry.example.org"},
			},
			TransportRequestID: "12345678",
			ConfigFile:         "ui5-deploy.yaml",
			DeployUser:         "node",
		}
		err := action.Perform(&cmd)
		if assert.NoError(t, err) {
			assert.Contains(
				t,
				cmd.Calls[0],
				"npm install --global --verbose --registry https://registry.example.org @sap/my-dep",
			)
			assert.Contains(
				t,
				cmd.Calls[0],
				"su node",
			)
		}
	})

	t.Run("deploy command tests", func(t *testing.T) {
		t.Run("all possible values provided", func(t *testing.T) {
			cmd := mock.ShellMockRunner{}
			action := CTSUploadAction{
				Connection:  CTSConnection{Endpoint: "https://example.org:8080/cts", Client: "001", User: "me", Password: "******"},
				Application: CTSApplication{Pack: "abapPackage", Name: "appName", Desc: "the Desc"},
				Node: CTSNode{
					DeployDependencies: []string{},
					InstallOpts:        []string{},
				},
				TransportRequestID: "12345678",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}

			err := action.Perform(&cmd)
			if assert.NoError(t, err) {
				assert.Contains(
					t,
					cmd.Calls[0],
					"fiori deploy -f -y --username ABAP_USER --password ABAP_PASSWORD -e \"the Desc\" --noConfig --url https://example.org:8080/cts --client 001 -t 12345678 -p abapPackage --name appName",
				)
				assert.Equal(t, []string{"ABAP_USER=me", "ABAP_PASSWORD=******"}, cmd.Env)
			}
		})

		t.Run("all possible values omitted", func(t *testing.T) {
			// In this case the values are expected inside the fiori deploy config file
			cmd := mock.ShellMockRunner{}
			action := CTSUploadAction{
				Connection:  CTSConnection{Endpoint: "", Client: "", User: "me", Password: "******"},
				Application: CTSApplication{Pack: "", Name: "", Desc: ""},
				Node: CTSNode{
					DeployDependencies: []string{},
					InstallOpts:        []string{},
				},
				TransportRequestID: "12345678",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}
			err := action.Perform(&cmd)
			if assert.NoError(t, err) {
				assert.Contains(
					t,
					cmd.Calls[0],
					"fiori deploy -f -y --username ABAP_USER --password ABAP_PASSWORD -e \"Deployed with Piper based on SAP Fiori tools\" --noConfig -t 12345678",
				)
				assert.Equal(t, []string{"ABAP_USER=me", "ABAP_PASSWORD=******"}, cmd.Env)
			}
		})
	})

	t.Run("config file releated tests", func(t *testing.T) {
		connection := CTSConnection{Endpoint: "", Client: "", User: "me", Password: "******"}
		app := CTSApplication{Pack: "", Name: "", Desc: ""}
		node := CTSNode{
			DeployDependencies: []string{},
			InstallOpts:        []string{},
		}
		t.Run("default config file exists", func(t *testing.T) {
			filesMock := mock.FilesMock{}
			filesMock.AddFile("ui5-deploy.yaml", []byte{})
			files = &filesMock
			defer func() { files = piperutils.Files{} }()
			cmd := mock.ShellMockRunner{}
			action := CTSUploadAction{
				Connection:         connection,
				Application:        app,
				Node:               node,
				TransportRequestID: "12345678",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}
			err := action.Perform(&cmd)
			if assert.NoError(t, err) {
				assert.Contains(t, cmd.Calls[0], "-c \"ui5-deploy.yaml\"")
			}
		})
		t.Run("Config file exists", func(t *testing.T) {
			filesMock := mock.FilesMock{}
			filesMock.AddFile("my-ui5-deploy.yaml", []byte{})
			files = &filesMock
			defer func() { files = piperutils.Files{} }()
			cmd := mock.ShellMockRunner{}
			action := CTSUploadAction{
				Connection:         connection,
				Application:        app,
				Node:               node,
				TransportRequestID: "12345678",
				ConfigFile:         "my-ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}

			err := action.Perform(&cmd)
			if assert.NoError(t, err) {
				assert.Contains(t, cmd.Calls[0], "-c \"my-ui5-deploy.yaml\"")
			}
		})
		t.Run("Config file missing", func(t *testing.T) {
			cmd := mock.ShellMockRunner{}
			action := CTSUploadAction{
				Connection:         connection,
				Application:        app,
				Node:               node,
				TransportRequestID: "12345678",
				ConfigFile:         "my-ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}
			err := action.Perform(&cmd)
			assert.EqualError(t, err, "Configured deploy config file 'my-ui5-deploy.yaml' does not exists")
		})
	})
}
