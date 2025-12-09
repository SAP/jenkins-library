package cmd

import (
	"fmt"
	"reflect"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentCheckoutBranch(options abapEnvironmentCheckoutBranchOptions, _ *telemetry.CustomData) {

	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	apiManager := abaputils.SoftwareComponentApiManager{
		Client:        &piperhttp.Client{},
		PollIntervall: 5 * time.Second,
	}

	var reports []piperutils.Path

	logOutputManager := abaputils.LogOutputManager{
		LogOutput:    options.LogOutput,
		PiperStep:    "checkoutBranch",
		FileNameStep: "checkoutBranch",
		StepReports:  reports,
	}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentCheckoutBranch(&options, &autils, &apiManager, &logOutputManager)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCheckoutBranch(options *abapEnvironmentCheckoutBranchOptions, com abaputils.Communication, apiManager abaputils.SoftwareComponentApiManagerInterface, logOutputManager *abaputils.LogOutputManager) (err error) {

	// Mapping for options
	subOptions := convertCheckoutConfig(options)

	//  Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, errorGetInfo := com.GetAbapCommunicationArrangementInfo(subOptions, "")
	if errorGetInfo != nil {
		log.Entry().WithError(errorGetInfo).Fatal("Parameters for the ABAP Connection not available")
	}
	connectionDetails.CertificateNames = options.CertificateNames

	repositories := []abaputils.Repository{}
	err = checkCheckoutBranchRepositoryConfiguration(*options)
	if err != nil {
		return errors.Wrap(err, "Configuration is not consistent")
	}
	repositories, err = abaputils.GetRepositories(&abaputils.RepositoriesConfig{BranchName: options.BranchName, RepositoryName: options.RepositoryName, Repositories: options.Repositories}, true)
	if err != nil {
		return errors.Wrap(err, "Could not read repositories")
	}

	err = checkoutBranches(repositories, connectionDetails, apiManager, logOutputManager)
	if err != nil {
		return fmt.Errorf("Something failed during the checkout: %w", err)
	}

	// Persist log archive
	abaputils.PersistArchiveLogsForPiperStep(logOutputManager)

	log.Entry().Infof("-------------------------")
	log.Entry().Info("All branches were checked out successfully")
	return nil
}

func checkoutBranches(repositories []abaputils.Repository, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, apiManager abaputils.SoftwareComponentApiManagerInterface, logOutputManager *abaputils.LogOutputManager) (err error) {
	log.Entry().Infof("Start switching %v branches", len(repositories))
	for _, repo := range repositories {
		err = handleCheckout(repo, checkoutConnectionDetails, apiManager, logOutputManager)
		if err != nil {
			break
		}
	}
	return err
}

func checkCheckoutBranchRepositoryConfiguration(options abapEnvironmentCheckoutBranchOptions) error {
	if options.Repositories == "" && options.RepositoryName == "" && options.BranchName == "" {
		return errors.New("You have not specified any repository or branch configuration to be checked out in the ABAP Environment System. Please make sure that you specified the repositories with their branches that should be checked out either in a dedicated file or via the parameters 'repositoryName' and 'branchName'. For more information please read the user documentation")
	}
	if options.Repositories != "" && options.RepositoryName != "" && options.BranchName != "" {
		log.Entry().Info("It seems like you have specified repositories directly via the configuration parameters 'repositoryName' and 'branchName' as well as in the dedicated repositories configuration file. Please note that in this case both configurations will be handled and checked out.")
	}
	if options.Repositories != "" && ((options.RepositoryName == "") != (options.BranchName == "")) {
		log.Entry().Info("It seems like you have specified a dedicated repository configuration file but also a wrong configuration for the parameters 'repositoryName' and 'branchName' to be checked out.")
		if options.RepositoryName != "" {
			log.Entry().Info("Please also add the value for the branchName parameter or remove the repositoryName parameter.")
		} else {
			log.Entry().Info("Please also add the value for the repositoryName parameter or remove the branchName parameter.")
		}
	}
	return nil
}

func handleCheckout(repo abaputils.Repository, checkoutConnectionDetails abaputils.ConnectionDetailsHTTP, apiManager abaputils.SoftwareComponentApiManagerInterface, logOutputManager *abaputils.LogOutputManager) (err error) {

	if reflect.DeepEqual(abaputils.Repository{}, repo) {
		return fmt.Errorf("Failed to read repository configuration: %w", errors.New("Error in configuration, most likely you have entered empty or wrong configuration values. Please make sure that you have correctly specified the branches in the repositories to be checked out"))
	}
	startCheckoutLogs(repo.Branch, repo.Name)

	api, errGetAPI := apiManager.GetAPI(checkoutConnectionDetails, repo)
	if errGetAPI != nil {
		return errors.Wrap(errGetAPI, "Could not initialize the connection to the system")
	}

	err = api.CheckoutBranch()
	if err != nil {
		return fmt.Errorf("Failed to trigger Checkout: %w", errors.New("Checkout of "+repo.Branch+" for software component "+repo.Name+" failed on the ABAP System"))
	}

	// set correct filename for archive file
	logOutputManager.FileNameStep = "checkoutBranch"
	// Polling the status of the repository import on the ABAP Environment system
	status, errorPollEntity := abaputils.PollEntity(api, apiManager.GetPollIntervall(), logOutputManager)
	if errorPollEntity != nil {
		return fmt.Errorf("Failed to poll Checkout: %w", errors.New("Status of checkout action on repository"+repo.Name+" failed on the ABAP System"))
	}
	const abapStatusCheckoutFail = "E"
	if status == abapStatusCheckoutFail {
		return fmt.Errorf("Checkout failed: %w", errors.New("Checkout of branch "+repo.Branch+" failed on the ABAP System"))
	}
	finishCheckoutLogs(repo.Branch, repo.Name)

	return err
}

func startCheckoutLogs(branchName string, repositoryName string) {
	log.Entry().Infof("Starting to switch branch to branch '%v' on repository '%v'", branchName, repositoryName)
	log.Entry().Infof("-------------------------")
	log.Entry().Info("Start checkout branch: " + branchName)
	log.Entry().Infof("-------------------------")
}

func finishCheckoutLogs(branchName string, repositoryName string) {
	log.Entry().Infof("-------------------------")
	log.Entry().Infof("Checkout of branch %v on repository %v was successful", branchName, repositoryName)
	log.Entry().Infof("-------------------------")
}

func convertCheckoutConfig(config *abapEnvironmentCheckoutBranchOptions) abaputils.AbapEnvironmentOptions {
	subOptions := abaputils.AbapEnvironmentOptions{}

	subOptions.CfAPIEndpoint = config.CfAPIEndpoint
	subOptions.CfServiceInstance = config.CfServiceInstance
	subOptions.CfServiceKeyName = config.CfServiceKeyName
	subOptions.CfOrg = config.CfOrg
	subOptions.CfSpace = config.CfSpace
	subOptions.Host = config.Host
	subOptions.Password = config.Password
	subOptions.Username = config.Username
	return subOptions
}
