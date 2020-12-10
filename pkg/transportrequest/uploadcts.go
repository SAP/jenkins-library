package transportrequest

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"strings"
)

type fileUtils interface {
	FileExists(string) (bool, error)
}

var files fileUtils = piperutils.Files{}

// CTSConnection ...
type CTSConnection struct {
	Endpoint string
	Client   string
	User     string
	Password string
}

// CTSApplication ...
type CTSApplication struct {
	Name string
	Pack string
	Desc string
}

// CTSNode ...
type CTSNode struct {
	DeployDependencies []string
	InstallOpts        []string
}

// CTSUploadAction ...
type CTSUploadAction struct {
	Connection         CTSConnection
	Application        CTSApplication
	Node               CTSNode
	TransportRequestID string
	ConfigFile         string
	DeployUser         string
}

const (
	abapUserKey           = "ABAP_USER"
	abapPasswordKey       = "ABAP_PASSWORD"
	defaultConfigFileName = "ui5-deploy.yaml"
)

// WithConnection ...
func (action *CTSUploadAction) WithConnection(connection CTSConnection) {
	action.Connection = connection
}

// WithApplication ...
func (action *CTSUploadAction) WithApplication(app CTSApplication) {
	action.Application = app
}

// WithNodeProperties ...
func (action *CTSUploadAction) WithNodeProperties(node CTSNode) {
	action.Node = node
}

// WithTransportRequestID ...
func (action *CTSUploadAction) WithTransportRequestID(id string) {
	action.TransportRequestID = id
}

// WithConfigFile ...
func (action *CTSUploadAction) WithConfigFile(configFile string) {
	action.ConfigFile = configFile
}

// WithDeployUser ...
func (action *CTSUploadAction) WithDeployUser(deployUser string) {
	action.DeployUser = deployUser
}

// Perform ...
func (action *CTSUploadAction) Perform(command command.ShellRunner) error {

	command.AppendEnv(
		[]string{
			fmt.Sprintf("%s=%s", abapUserKey, action.Connection.User),
			fmt.Sprintf("%s=%s", abapPasswordKey, action.Connection.Password),
		})

	cmd := []string{"/bin/bash -e"}

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

	return command.RunShell("/bin/bash", strings.Join(cmd, "\n"))
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
	app CTSApplication,
	cts CTSConnection,
) (string, error) {
	desc := app.Desc
	if len(desc) == 0 {
		desc = "Deployed with Piper based on SAP Fiori tools"
	}

	useConfigFile, noConfig, err := handleConfigFile(configFile)
	if err != nil {
		return "", err
	}
	cmd := []string{
		"fiori",
		"deploy",
		"-f", // failfast --> provide return code != 0 in case of any failure
		"-y", // autoconfirm --> no need to press 'y' key in order to confirm the params and trigger the deployment
		"--username", abapUserKey,
		"--password", abapPasswordKey,
		"-e", fmt.Sprintf("\"%s\"", desc),
	}

	if noConfig {
		cmd = append(cmd, "--noConfig") // no config file, but we will provide our parameters
	}
	if useConfigFile {
		cmd = append(cmd, "-c", fmt.Sprintf("\"%s\"", configFile))
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
		cmd = append(cmd, "-t", transportRequestID)
	} else {
		log.Entry().Debug("No transportRequestID found in piper config.")
	}
	if len(app.Pack) > 0 {
		log.Entry().Debugf("application package '%s' used from piper config", app.Pack)
		cmd = append(cmd, "-p", app.Pack)
	} else {
		log.Entry().Debug("No application package found in piper config.")
	}
	if len(app.Name) > 0 {
		log.Entry().Debugf("application name '%s' used from piper config", app.Name)
		cmd = append(cmd, "-n", app.Name)
	} else {
		log.Entry().Debug("No application name found in piper config.")
	}

	return strings.Join(cmd, " "), nil
}

func getSwitchUserStatement(user string) string {
	return fmt.Sprintf("su %s", user)
}

func handleConfigFile(path string) (bool, bool, error) {

	useConfigFile := true
	noConfig := false

	if len(path) == 0 {
		useConfigFile = false
		exists, err := files.FileExists(defaultConfigFileName)
		if err != nil {
			return false, false, err
		}
		noConfig = !exists
	} else {
		exists, err := files.FileExists(path)
		if err != nil {
			return false, false, err
		}
		if exists {
			useConfigFile = true
			noConfig = false
		} else {
			if path == defaultConfigFileName {
				// in this case this is most likely provided by the piper default config and
				// it was not explicitly configured. Hence we assume not having a config file
				useConfigFile = false
				noConfig = true
			} else {
				err = fmt.Errorf("Configured deploy config file '%s' does not exists", path)
				return false, false, err
			}
		}
	}
	return useConfigFile, noConfig, nil
}
