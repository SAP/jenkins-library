//go:build unit
// +build unit

package cts

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
)

func TestUploadCTS(t *testing.T) {

	fMock := &mock.FilesMock{}
	files = fMock
	defer func() { files = piperutils.Files{} }()

	t.Run("npm install command tests", func(t *testing.T) {
		cmd := mock.ShellMockRunner{}
		action := UploadAction{
			Connection:  Connection{Endpoint: "", Client: "", User: "me", Password: "******"},
			Application: Application{Pack: "", Name: "", Desc: ""},
			Node: Node{
				DeployDependencies: []string{"@sap/my-dep"},
				InstallOpts:        []string{"--verbose", "--registry", "https://registry.example.org"},
			},
			TransportRequestID: "12345678",
			ConfigFile:         "ui5-deploy.yaml",
			DeployUser:         "node",
		}
		err := action.Perform(&cmd)
		if assert.NoError(t, err) {
			assert.Regexp(
				t,
				"(?m)^npm install --global --verbose --registry https://registry.example.org @sap/my-dep$",
				cmd.Calls[0],
				"Expected npm install command not found",
			)
			assert.Regexp(
				t,
				"(?m)^su node$",
				cmd.Calls[0],
				"Expected switch user statement not found",
			)
		}
	})

	t.Run("deploy command tests", func(t *testing.T) {
		t.Run("all possible values provided", func(t *testing.T) {
			cmd := mock.ShellMockRunner{}
			action := UploadAction{
				Connection:  Connection{Endpoint: "https://example.org:8080/cts", Client: "001", User: "me", Password: "******"},
				Application: Application{Pack: "abapPackage", Name: "/0ABCD/appName", Desc: "the Desc"},
				Node: Node{
					DeployDependencies: []string{},
					InstallOpts:        []string{},
				},
				TransportRequestID: "12345678",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}

			err := action.Perform(&cmd)
			if assert.NoError(t, err) {
				assert.Regexp(
					t,
					"(?m)^fiori deploy --failfast --yes --username ABAP_USER --password ABAP_PASSWORD --description \"the Desc\" --noConfig --url https://example.org:8080/cts --client 001 --transport 12345678 --package abapPackage --name /0ABCD/appName",
					cmd.Calls[0],
					"Expected fiori deploy command not found",
				)
				assert.Equal(t, []string{"ABAP_USER=me", "ABAP_PASSWORD=******"}, cmd.Env)
			}
		})

		t.Run("all possible values omitted", func(t *testing.T) {
			// In this case the values are expected inside the fiori deploy config file
			cmd := mock.ShellMockRunner{}
			action := UploadAction{
				Connection:  Connection{Endpoint: "", Client: "", User: "me", Password: "******"},
				Application: Application{Pack: "", Name: "", Desc: ""},
				Node: Node{
					DeployDependencies: []string{},
					InstallOpts:        []string{},
				},
				TransportRequestID: "12345678",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}
			err := action.Perform(&cmd)

			if assert.NoError(t, err) {
				assert.Regexp(
					t,
					"(?m)^fiori deploy --failfast --yes --username ABAP_USER --password ABAP_PASSWORD --description \"Deployed with Piper based on SAP Fiori tools\" --noConfig --transport 12345678$",
					cmd.Calls[0],
					"Expected fiori deploy command not found",
				)
				assert.Equal(t, []string{"ABAP_USER=me", "ABAP_PASSWORD=******"}, cmd.Env)
			}
		})

		t.Run("fail in case of invalid app name", func(t *testing.T) {
			cmd := mock.ShellMockRunner{}
			action := UploadAction{
				Connection:  Connection{Endpoint: "https://example.org:8080/cts", Client: "001", User: "me", Password: "******"},
				Application: Application{Pack: "abapPackage", Name: "/AB/app1", Desc: "the Desc"},
				Node: Node{
					DeployDependencies: []string{},
					InstallOpts:        []string{},
				},
				TransportRequestID: "12345678",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}

			err := action.Perform(&cmd)
			expectedErrorMessge := "application name '/AB/app1' contains spaces or special characters or invalid namespace prefix and is not according to the regex '^(/[A-Za-z0-9_]{3,8}/)?[A-Za-z0-9_]+$'."

			assert.EqualErrorf(t, err, expectedErrorMessge, "invalid app name")
		})
	})

	t.Run("config file releated tests", func(t *testing.T) {
		connection := Connection{Endpoint: "", Client: "", User: "me", Password: "******"}
		app := Application{Pack: "", Name: "", Desc: ""}
		node := Node{
			DeployDependencies: []string{},
			InstallOpts:        []string{},
		}
		t.Run("default config file exists", func(t *testing.T) {
			filesMock := mock.FilesMock{}
			filesMock.AddFile("ui5-deploy.yaml", []byte{})
			files = &filesMock
			defer func() { files = fMock }()
			cmd := mock.ShellMockRunner{}
			action := UploadAction{
				Connection:         connection,
				Application:        app,
				Node:               node,
				TransportRequestID: "12345678",
				ConfigFile:         "ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}
			err := action.Perform(&cmd)
			if assert.NoError(t, err) {
				assert.Contains(t, cmd.Calls[0], " --config \"ui5-deploy.yaml\" ")
			}
		})
		t.Run("Config file exists", func(t *testing.T) {
			filesMock := mock.FilesMock{}
			filesMock.AddFile("my-ui5-deploy.yaml", []byte{})
			files = &filesMock
			defer func() { files = fMock }()
			cmd := mock.ShellMockRunner{}
			action := UploadAction{
				Connection:         connection,
				Application:        app,
				Node:               node,
				TransportRequestID: "12345678",
				ConfigFile:         "my-ui5-deploy.yaml",
				DeployUser:         "doesNotMatterInThisCase",
			}

			err := action.Perform(&cmd)
			if assert.NoError(t, err) {
				assert.Contains(t, cmd.Calls[0], " --config \"my-ui5-deploy.yaml\" ")
			}
		})
		t.Run("Config file missing", func(t *testing.T) {
			cmd := mock.ShellMockRunner{}
			action := UploadAction{
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
