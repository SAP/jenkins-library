package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/transportrequest/cts"
	"github.com/go-git/go-git/v5"
	"os"
	//"github.com/go-git/go-git/v5/plumbing/object"
	pipergitutils "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/transportrequest"
)

type transportRequestUploadUtils interface {
	command.ShellRunner

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The transportRequestUploadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

// UploadAction ...
type UploadAction interface {
	Perform(command.ShellRunner) error
	WithConnection(cts.Connection)
	WithApplication(cts.Application)
	WithNodeProperties(cts.Node)
	WithTransportRequestID(string)
	WithConfigFile(string)
	WithDeployUser(string)
}

type transportRequestUploadCTSUtilsBundle struct {
	*command.Command

	// Embed more structs as necessary to implement methods or interfaces you add to transportRequestUploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// transportRequestUploadUtilsBundle and forward to the implementation of the dependency.
}

func newTransportRequestUploadCTSUtils() transportRequestUploadUtils {
	utils := transportRequestUploadCTSUtilsBundle{
		Command: &command.Command{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func transportRequestUploadCTS(config transportRequestUploadCTSOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newTransportRequestUploadCTSUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runTransportRequestUploadCTS(&config, &cts.UploadAction{}, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runTransportRequestUploadCTS(
	config *transportRequestUploadCTSOptions,
	action UploadAction,
	telemetryData *telemetry.CustomData,
	cmd command.ShellRunner) error {

	log.Entry().Debugf("Entering 'runTransportRequestUpload' with config: %v", config)

	transportRequestID := config.TransportRequestID

	if len(transportRequestID) == 0 {

		from := "origin/master"
		to := "HEAD"

		log.Entry().Infof("TransportRequestID not provided by configuration. Traversing commit history, range: '%s..%s'", from, to)
		workdir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error: $v\n", err)
			return err
		}
		fmt.Printf("Opening repo at '%s'\n", workdir)
		r, err := git.PlainOpen(workdir)
		if err != nil {
			fmt.Printf("Error: $v\n", err)
			return err
		}

		cIter, err := pipergitutils.LogRange(r, from, to)
		if err != nil {
			fmt.Printf("Error: $v\n", err)
			return err
		}
		ids, err := transportrequest.FindLabelsInCommits(cIter, "TransportRequest")
		if err != nil {
			fmt.Printf("Error: $v\n", err)
			return err
		}
		fmt.Printf("[MH] ids: %s\n", ids)

		if len(ids) > 1 {
			return fmt.Errorf("More than one transportRequestID found: %v", ids)
		}
		if len(ids) == 0 {
			return fmt.Errorf("No transportRequestID found.")
		}
		transportRequestID = ids[0]
		log.Entry().Infof("Transport request ID '%s' retrieved from commit history (range: '%s..%s')", transportRequestID, from, to)
	} else {
		log.Entry().Infof("Transport request ID '%s' explicitly provided by configuration", transportRequestID)
	}
	return nil

	action.WithConnection(cts.Connection{
		Endpoint: config.Endpoint,
		Client:   config.Client,
		User:     config.Username,
		Password: config.Password,
	})
	action.WithApplication(cts.Application{
		Name: config.ApplicationName,
		Pack: config.AbapPackage,
		Desc: config.Description,
	})
	action.WithNodeProperties(cts.Node{
		DeployDependencies: config.DeployToolDependencies,
		InstallOpts:        config.NpmInstallOpts,
	})

	action.WithTransportRequestID(transportRequestID)
	action.WithConfigFile(config.DeployConfigFile)
	action.WithDeployUser(config.OsDeployUser)

	return action.Perform(cmd)
}
