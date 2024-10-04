package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentCreateTag(config abapEnvironmentCreateTagOptions, _ *telemetry.CustomData) {

	c := command.Command{}

	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	apiManager := abaputils.SoftwareComponentApiManager{
		Client:        &piperhttp.Client{},
		PollIntervall: 5 * time.Second,
	}

	if err := runAbapEnvironmentCreateTag(&config, &autils, &apiManager); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCreateTag(config *abapEnvironmentCreateTagOptions, com abaputils.Communication, apiManager abaputils.SoftwareComponentApiManagerInterface) error {

	connectionDetails, errorGetInfo := com.GetAbapCommunicationArrangementInfo(convertTagConfig(config), "")
	if errorGetInfo != nil {
		return errors.Wrap(errorGetInfo, "Parameters for the ABAP Connection not available")
	}
	connectionDetails.CertificateNames = config.CertificateNames

	backlog, errorPrepare := prepareBacklog(config)
	if errorPrepare != nil {
		return fmt.Errorf("Something failed during the tag creation: %w", errorPrepare)
	}

	return createTags(backlog, connectionDetails, apiManager)
}

func createTags(backlog []abaputils.CreateTagBacklog, con abaputils.ConnectionDetailsHTTP, apiManager abaputils.SoftwareComponentApiManagerInterface) (err error) {

	errorOccurred := false
	for _, item := range backlog {
		err = createTagsForSingleItem(item, con, apiManager)
		if err != nil {
			errorOccurred = true
		}
	}

	if errorOccurred {
		message := "At least one tag has not been created"
		log.Entry().Errorf(message)
		return errors.New(message)
	}
	return nil

}

func createTagsForSingleItem(item abaputils.CreateTagBacklog, con abaputils.ConnectionDetailsHTTP, apiManager abaputils.SoftwareComponentApiManagerInterface) (err error) {

	errorOccurred := false
	for index := range item.Tags {
		err = createSingleTag(item, index, con, apiManager)
		if err != nil {
			errorOccurred = true
		}
	}
	if errorOccurred {
		message := "At least one tag has not been created"
		err = errors.New(message)
	}
	return err
}

func createSingleTag(item abaputils.CreateTagBacklog, index int, con abaputils.ConnectionDetailsHTTP, apiManager abaputils.SoftwareComponentApiManagerInterface) (err error) {

	api, errGetAPI := apiManager.GetAPI(con, abaputils.Repository{Name: item.RepositoryName, CommitID: item.CommitID})
	if errGetAPI != nil {
		return errors.Wrap(errGetAPI, "Could not initialize the connection to the system")
	}

	createTagError := api.CreateTag(item.Tags[index])
	if createTagError != nil {
		return errors.Wrapf(err, "Creation of Tag failed on the ABAP system")
	}

	logOutputManager := abaputils.LogOutputManager{
		LogOutput:    "STANDARD",
		PiperStep:    "createTag",
		FileNameStep: "createTag",
		StepReports:  nil,
	}

	status, errorPollEntity := abaputils.PollEntity(api, apiManager.GetPollIntervall(), &logOutputManager)

	if errorPollEntity == nil && status == "S" {
		log.Entry().Info("Created tag " + item.Tags[index].TagName + " for repository " + item.RepositoryName + " with commitID " + item.CommitID)
	} else {
		log.Entry().Error("NOT created: Tag " + item.Tags[index].TagName + " for repository " + item.RepositoryName + " with commitID " + item.CommitID)
		err = errors.New("Creation of Tag failed on the ABAP system")
	}

	return err
}

func prepareBacklog(config *abapEnvironmentCreateTagOptions) (backlog []abaputils.CreateTagBacklog, err error) {

	if config.Repositories != "" && config.RepositoryName != "" {
		return nil, errors.New("Configuring the parameter repositories and the parameter repositoryName at the same time is not allowed")
	}

	if config.RepositoryName != "" && config.CommitID != "" {
		backlog = append(backlog, abaputils.CreateTagBacklog{RepositoryName: config.RepositoryName, CommitID: config.CommitID})
	}

	if config.Repositories != "" {
		descriptor, err := abaputils.ReadAddonDescriptor(config.Repositories) //config.Repositories should contain a file name
		if err != nil {
			return nil, err
		}
		for _, repo := range descriptor.Repositories {
			backlogInstance := abaputils.CreateTagBacklog{RepositoryName: repo.Name, CommitID: repo.CommitID}
			if config.GenerateTagForAddonComponentVersion && repo.VersionYAML != "" {
				tag := abaputils.Tag{TagName: "v" + repo.VersionYAML, TagDescription: "Generated by the ABAP Environment Pipeline"}
				backlogInstance.Tags = append(backlogInstance.Tags, tag)
			}
			backlog = append(backlog, backlogInstance)
		}
		if config.GenerateTagForAddonProductVersion {
			if descriptor.AddonProduct != "" && descriptor.AddonVersionYAML != "" {
				addonProductDash := strings.Replace(descriptor.AddonProduct, "/", "-", 2)
				backlog = addTagToList(backlog, addonProductDash+"-"+descriptor.AddonVersionYAML, "Generated by the ABAP Environment Pipeline")
			} else {
				log.Entry().WithField("generateTagForAddonProductVersion", config.GenerateTagForAddonProductVersion).WithField("AddonProduct", descriptor.AddonProduct).WithField("AddonVersion", descriptor.AddonVersionYAML).Infof("Not all required values are provided to create an addon product version tag")
			}
		}
	}
	if config.TagName != "" {
		backlog = addTagToList(backlog, config.TagName, config.TagDescription)
	}
	return backlog, nil
}

func addTagToList(backlog []abaputils.CreateTagBacklog, tag string, description string) []abaputils.CreateTagBacklog {

	for i, item := range backlog {
		tag := abaputils.Tag{TagName: tag, TagDescription: description}
		backlog[i].Tags = append(item.Tags, tag)
	}
	return backlog
}

func convertTagConfig(config *abapEnvironmentCreateTagOptions) abaputils.AbapEnvironmentOptions {
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
