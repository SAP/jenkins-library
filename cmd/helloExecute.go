package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type helloExecuteUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	GetDockerImageValue(stepName string) (string, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The helloExecuteUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type helloExecuteUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to helloExecuteUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// helloExecuteUtilsBundle and forward to the implementation of the dependency.
}

func (h *helloExecuteUtilsBundle) GetDockerImageValue(stepName string) (string, error) {
	return GetDockerImageValue(stepName)
}

func newHelloExecuteUtils() helloExecuteUtils {
	utils := helloExecuteUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func helloExecute(config helloExecuteOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newHelloExecuteUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runHelloExecute(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runHelloExecute(config *helloExecuteOptions, telemetryData *telemetry.CustomData, utils helloExecuteUtils) error {
	log.Entry().Infof("Greeting user: %s", config.HelloUsername)

	// Retrieve the docker image for this step
	stepName := "helloExecute"
	dockerImage, err := utils.GetDockerImageValue(stepName)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("failed to retrieve docker image: %w", err)
	}

	log.Entry().Debugf("Using docker image: %s", dockerImage)

	// Execute the greeting command in docker container
	greetingCommand := fmt.Sprintf("echo 'Hello %s!'", config.HelloUsername)
	log.Entry().Infof("Executing greeting: %s", greetingCommand)

	// Build docker run command
	err = utils.RunExecutable("docker", "run", "--rm", dockerImage, "sh", "-c", greetingCommand)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return fmt.Errorf("failed to execute greeting command in docker: %w", err)
	}

	log.Entry().Info("Greeting executed successfully")
	return nil
}
