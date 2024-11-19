package cts

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type fileUtils interface {
	FileExists(string) (bool, error)
}

var files fileUtils = piperutils.Files{}

// Connection Everything wee need for connecting to CTS
type Connection struct {
	// The endpoint in for form <protocol>://<host>:<port>, no path
	Endpoint string
	// The ABAP client, like e.g. "001"
	Client   string
	User     string
	Password string
}

// Application The details of the application
type Application struct {
	// Name of the application
	Name string
	// The ABAP package
	Pack string
	// A description. Only taken into account for initial upload, not
	// in case of a re-deployment.
	Desc string
}

// Node The details for configuring the node image
type Node struct {
	// The dependencies which are installed on a basic node image in order
	// to enable it for fiori deployment. If left empty we assume the
	// provided base image has already everything installed.
	DeployDependencies []string
	// Additional options for the npm install command. Useful e.g.
	// for providing additional registries or for triggering verbose mode
	InstallOpts []string
}

// UploadAction Collects all the properties we need for the deployment
type UploadAction struct {
	Connection         Connection
	Application        Application
	Node               Node
	TransportRequestID string
	ConfigFile         string
	DeployUser         string
}

const (
	abapUserKey           = "ABAP_USER"
	abapPasswordKey       = "ABAP_PASSWORD"
	defaultConfigFileName = "ui5-deploy.yaml"
	pattern               = "^(/[A-Za-z0-9_]{3,8}/)?[A-Za-z0-9_]+$"
)

// WithConnection ...
func (action *UploadAction) WithConnection(connection Connection) {
	action.Connection = connection
}

// WithApplication ...
func (action *UploadAction) WithApplication(app Application) {
	action.Application = app
}

// WithNodeProperties ...
func (action *UploadAction) WithNodeProperties(node Node) {
	action.Node = node
}

// WithTransportRequestID ...
func (action *UploadAction) WithTransportRequestID(id string) {
	action.TransportRequestID = id
}

// WithConfigFile ...
func (action *UploadAction) WithConfigFile(configFile string) {
	action.ConfigFile = configFile
}

// WithDeployUser ...
func (action *UploadAction) WithDeployUser(deployUser string) {
	action.DeployUser = deployUser
}

// Perform Performs the upload
func (action *UploadAction) Perform(command command.ShellRunner) error {

	command.AppendEnv(
		[]string{
			fmt.Sprintf("%s=%s", abapUserKey, action.Connection.User),
			fmt.Sprintf("%s=%s", abapPasswordKey, action.Connection.Password),
		})

	cmd := []string{"#!/bin/sh -e"}

	noInstall := len(action.Node.DeployDependencies) == 0
	if !noInstall {
		cmd = append(cmd, "echo \"Current user is '$(whoami)'\"")
		cmd = append(cmd, getPrepareFioriEnvironmentStatement(action.Node.DeployDependencies, action.Node.InstallOpts))
		cmd = append(cmd, getSwitchUserStatement(action.DeployUser))
	} else {
		log.Entry().Info("No deploy dependencies provided. Skipping npm install call. Assuming current docker image already contains the dependencies for performing the deployment.")
	}

	deployStatement, err := getFioriDeployStatement(action.TransportRequestID, action.ConfigFile, action.Application, action.Connection)
	if err != nil {
		return err
	}

	cmd = append(cmd, deployStatement)

	return command.RunShell("/bin/sh", strings.Join(cmd, "\n"))
}

func getPrepareFioriEnvironmentStatement(deps []string, npmInstallOpts []string) string {
	cmd := []string{
		"npm",
		"install",
		"--global",
	}
	cmd = append(cmd, npmInstallOpts...)
	cmd = append(cmd, deps...)
	return strings.Join(cmd, " ")
}

func getFioriDeployStatement(
	transportRequestID string,
	configFile string,
	app Application,
	cts Connection,
) (string, error) {
	desc := app.Desc
	if len(desc) == 0 {
		desc = "Deployed with Piper based on SAP Fiori tools"
	}

	useConfigFileOptionInCommandInvocation, useNoConfigFileOptionInCommandInvocation, err := handleConfigFileOptions(configFile)
	if err != nil {
		return "", err
	}
	cmd := []string{
		"fiori",
		"deploy",
		"--failfast", // provide return code != 0 in case of any failure
		"--yes",      // autoconfirm --> no need to press 'y' key in order to confirm the params and trigger the deployment
		"--username", abapUserKey,
		"--password", abapPasswordKey,
		"--description", fmt.Sprintf("\"%s\"", desc),
	}

	if useNoConfigFileOptionInCommandInvocation {
		cmd = append(cmd, "--noConfig") // no config file, but we will provide our parameters
	}
	if useConfigFileOptionInCommandInvocation {
		cmd = append(cmd, "--config", fmt.Sprintf("\"%s\"", configFile))
	}
	if len(cts.Endpoint) > 0 {
		log.Entry().Debugf("Endpoint '%s' used from piper config", cts.Endpoint)
		cmd = append(cmd, "--url", cts.Endpoint)
	} else {
		log.Entry().Debug("No endpoint found in piper config.")
	}
	if len(cts.Client) > 0 {
		log.Entry().Debugf("Client '%s' used from piper config", cts.Client)
		cmd = append(cmd, "--client", cts.Client)
	} else {
		log.Entry().Debug("No client found in piper config.")
	}
	if len(transportRequestID) > 0 {
		log.Entry().Debugf("TransportRequestID '%s' used from piper config", transportRequestID)
		cmd = append(cmd, "--transport", transportRequestID)
	} else {
		log.Entry().Debug("No transportRequestID found in piper config.")
	}
	if len(app.Pack) > 0 {
		log.Entry().Debugf("application package '%s' used from piper config", app.Pack)
		cmd = append(cmd, "--package", app.Pack)
	} else {
		log.Entry().Debug("No application package found in piper config.")
	}
	if len(app.Name) > 0 {
		re := regexp.MustCompile(pattern)
		if !re.MatchString(app.Name) {
			return "", fmt.Errorf("application name '%s' contains spaces or special characters or invalid namespace prefix and is not according to the regex '%s'.", app.Name, pattern)
		}
		log.Entry().Debugf("application name '%s' used from piper config", app.Name)
		cmd = append(cmd, "--name", app.Name)
	} else {
		log.Entry().Debug("No application name found in piper config.")
	}

	return strings.Join(cmd, " "), nil
}

func getSwitchUserStatement(user string) string {
	return fmt.Sprintf("su %s", user)
}

func handleConfigFileOptions(path string) (useConfigFileOptionInCommandInvocation, useNoConfigFileOptionInCommandInvoction bool, err error) {

	exists := false
	if len(path) == 0 {
		exists, err = files.FileExists(defaultConfigFileName)
		if err != nil {
			return
		}
		useConfigFileOptionInCommandInvocation = false
		useNoConfigFileOptionInCommandInvoction = !exists
		return
	}
	exists, err = files.FileExists(path)
	if err != nil {
		return
	}
	if exists {
		useConfigFileOptionInCommandInvocation = true
		useNoConfigFileOptionInCommandInvoction = false
	} else {
		if path != defaultConfigFileName {
			err = fmt.Errorf("Configured deploy config file '%s' does not exists", path)
			return
		}
		// in this case this is most likely provided by the piper default config and
		// it was not explicitly configured. Hence we assume not having a config file
		useConfigFileOptionInCommandInvocation = false
		useNoConfigFileOptionInCommandInvoction = true
	}
	return
}
