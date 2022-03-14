package cmd

import (
	"encoding/json"
	"fmt"
	"net/http/cookiejar"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentCreateTag(config abapEnvironmentCreateTagOptions, telemetryData *telemetry.CustomData) {

	c := command.Command{}

	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}

	err := runAbapEnvironmentCreateTag(&config, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCreateTag(config *abapEnvironmentCreateTagOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) error {

	connectionDetails, errorGetInfo := com.GetAbapCommunicationArrangementInfo(convertTagConfig(config), "")
	if errorGetInfo != nil {
		return errors.Wrap(errorGetInfo, "Parameters for the ABAP Connection not available")
	}

	// Configuring the HTTP Client and CookieJar
	cookieJar, errorCookieJar := cookiejar.New(nil)
	if errorCookieJar != nil {
		return errors.Wrap(errorCookieJar, "Could not create a Cookie Jar")
	}

	client.SetOptions(piperhttp.ClientOptions{
		MaxRequestDuration: 180 * time.Second,
		CookieJar:          cookieJar,
		Username:           connectionDetails.User,
		Password:           connectionDetails.Password,
	})

	backlog, errorPrepare := prepareBacklog(config)
	if errorPrepare != nil {
		return fmt.Errorf("Something failed during the tag creation: %w", errorPrepare)
	}

	createTags(backlog, telemetryData, connectionDetails, client)

	return nil
}

func createTags(backlog []CreateTagBacklog, telemetryData *telemetry.CustomData, con abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (err error) {

	connection := con
	connection.XCsrfToken = "fetch"
	resp, err := abaputils.GetHTTPResponse("HEAD", connection, nil, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Authentication on the ABAP system failed", con)
		return err
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", con.URL).Debug("Authentication on the ABAP system successful")
	con.XCsrfToken = resp.Header.Get("X-Csrf-Token")

	con.URL = con.URL + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Tags"

	for _, item := range backlog {
		err = createTagsForSingleItem(item, telemetryData, con, client)
	}
	return err
}

func createTagsForSingleItem(item CreateTagBacklog, telemetryData *telemetry.CustomData, con abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (err error) {

	for index, _ := range item.tags {
		err = createSingleTag(item, index, telemetryData, con, client)
	}
	return err
}

func createSingleTag(item CreateTagBacklog, index int, telemetryData *telemetry.CustomData, con abaputils.ConnectionDetailsHTTP, client piperhttp.Sender) (err error) {

	requestBodyStruct := CreateTagBody{repositoryName: item.repositoryName, commitID: item.commitID, tag: item.tags[index]}
	requestBodyJson, err := json.Marshal(&requestBodyStruct)
	if err != nil {
		return err
	}

	log.Entry().Debugf("Request body: %s", requestBodyJson)
	resp, err := abaputils.GetHTTPResponse("POST", con, requestBodyJson, client)
	if err != nil {
		err = abaputils.HandleHTTPError(resp, err, "Could not create tag "+requestBodyStruct.tag+" for repository "+requestBodyStruct.repositoryName+" with commitID "+requestBodyStruct.commitID, con)
		return err
	}
	defer resp.Body.Close()

	log.Entry().Info("Created tag " + requestBodyStruct.tag + " for repository " + requestBodyStruct.repositoryName + " with commitID " + requestBodyStruct.commitID)

	return err
}

func prepareBacklog(config *abapEnvironmentCreateTagOptions) (backlog []CreateTagBacklog, err error) {

	descriptor, err := abaputils.ReadAddonDescriptor(config.Repositories)
	if err != nil {
		return
	}
	repos := descriptor.Repositories

	for _, repo := range repos {

		backlogInstance := CreateTagBacklog{repositoryName: repo.Name, commitID: repo.CommitID}
		backlogInstance.tags = append(backlogInstance.tags, repo.VersionYAML)
		backlog = append(backlog, backlogInstance)

	}

	if config.RepositoryName != "" && config.CommitID != "" {
		backlog = append(backlog, CreateTagBacklog{repositoryName: config.RepositoryName, commitID: config.CommitID})
	}

	if config.CreateTagForAddonProductVersion {
		backlog = addTagToList(backlog, descriptor.AddonVersion)
	}
	if config.TagName != "" {
		backlog = addTagToList(backlog, config.TagName)
	}

	return
}

func addTagToList(backlog []CreateTagBacklog, tag string) []CreateTagBacklog {

	for _, item := range backlog {
		item.tags = append(item.tags, tag)
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

type CreateTagBacklog struct {
	repositoryName string
	commitID       string
	tags           []string
}

type CreateTagBody struct {
	repositoryName string `json:"sc_name"`
	commitID       string `json:"commit_id"`
	tag            string `json:"tag_name"`
}
