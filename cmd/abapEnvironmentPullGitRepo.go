package cmd

import (
	"fmt"
	"time"

	"errors"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapEnvironmentPullGitRepo(options abapEnvironmentPullGitRepoOptions, _ *telemetry.CustomData) {

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
		PiperStep:    "pull",
		FileNameStep: "pull",
		StepReports:  reports,
	}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentPullGitRepo(&options, &autils, &apiManager, &logOutputManager)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentPullGitRepo(options *abapEnvironmentPullGitRepoOptions, com abaputils.Communication, apiManager abaputils.SoftwareComponentApiManagerInterface, logOutputManager *abaputils.LogOutputManager) (err error) {

	subOptions := convertPullConfig(options)

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, err := com.GetAbapCommunicationArrangementInfo(subOptions, "")
	if err != nil {
		return fmt.Errorf("Parameters for the ABAP Connection not available: %w", err)
	}
	connectionDetails.CertificateNames = options.CertificateNames

	var repositories []abaputils.Repository
	err = checkPullRepositoryConfiguration(*options)
	if err != nil {
		return err
	}

	repositories, err = abaputils.GetRepositories(&abaputils.RepositoriesConfig{RepositoryNames: options.RepositoryNames, Repositories: options.Repositories, RepositoryName: options.RepositoryName, CommitID: options.CommitID}, false)
	handleIgnoreCommit(repositories, options.IgnoreCommit)
	if err != nil {
		return err
	}

	err = pullRepositories(repositories, connectionDetails, apiManager, logOutputManager)

	// Persist log archive
	abaputils.PersistArchiveLogsForPiperStep(logOutputManager)

	return err

}

func pullRepositories(repositories []abaputils.Repository, pullConnectionDetails abaputils.ConnectionDetailsHTTP, apiManager abaputils.SoftwareComponentApiManagerInterface, logOutputManager *abaputils.LogOutputManager) (err error) {
	log.Entry().Infof("Start pulling %v repositories", len(repositories))
	for _, repo := range repositories {
		err = handlePull(repo, pullConnectionDetails, apiManager, logOutputManager)
		if err != nil {
			break
		}
	}
	if err == nil {
		finishPullLogs()
	}
	return err
}

func handlePull(repo abaputils.Repository, con abaputils.ConnectionDetailsHTTP, apiManager abaputils.SoftwareComponentApiManagerInterface, logOutputManager *abaputils.LogOutputManager) (err error) {

	logString := repo.GetPullLogString()
	errorString := "Pull of the " + logString + " failed on the ABAP system"

	abaputils.AddDefaultDashedLine(1)
	log.Entry().Info("Start pulling the " + logString)
	abaputils.AddDefaultDashedLine(1)

	api, errGetAPI := apiManager.GetAPI(con, repo)
	if errGetAPI != nil {
		return fmt.Errorf("Could not initialize the connection to the system: %w", errGetAPI)
	}

	err = api.Pull()
	if err != nil {
		return fmt.Errorf("%s: %w", errorString, err)
	}

	// set correct filename for archive file
	logOutputManager.FileNameStep = "pull"
	// Polling the status of the repository import on the ABAP Environment system
	status, errorPollEntity := abaputils.PollEntity(api, apiManager.GetPollIntervall(), logOutputManager)
	if errorPollEntity != nil {
		return fmt.Errorf("%s: %w", errorString, errorPollEntity)
	}
	if status == "E" {
		return errors.New(errorString)
	}
	log.Entry().Info(repo.Name + " was pulled successfully")
	return err
}

func checkPullRepositoryConfiguration(options abapEnvironmentPullGitRepoOptions) error {

	if (len(options.RepositoryNames) > 0 && options.Repositories != "") || (len(options.RepositoryNames) > 0 && options.RepositoryName != "") || (options.RepositoryName != "" && options.Repositories != "") {
		return fmt.Errorf("Checking configuration failed: %w", errors.New("Only one of the paramters `RepositoryName`,`RepositoryNames` or `Repositories` may be configured at the same time"))
	}
	if len(options.RepositoryNames) == 0 && options.Repositories == "" && options.RepositoryName == "" {
		return fmt.Errorf("Checking configuration failed: %w", errors.New("You have not specified any repository configuration to be pulled into the ABAP Environment System. Please make sure that you specified the repositories that should be pulled either in a dedicated file or via the parameter 'repositoryNames'. For more information please read the User documentation"))
	}
	return nil
}

func finishPullLogs() {
	abaputils.AddDefaultDashedLine(1)
	log.Entry().Info("All repositories were pulled successfully")
}

func convertPullConfig(config *abapEnvironmentPullGitRepoOptions) abaputils.AbapEnvironmentOptions {
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

func handleIgnoreCommit(repositories []abaputils.Repository, ignoreCommit bool) {
	for i := range repositories {
		if ignoreCommit {
			repositories[i].CommitID = ""
		}
	}
}
