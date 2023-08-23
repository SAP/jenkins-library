package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/cookiejar"
	"strings"
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

	if err := runAbapEnvironmentCreateTag(&config, telemetryData, &autils, &client); err != nil {
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

	return createTags(backlog, telemetryData, connectionDetails, client, com)
}

func createTags(backlog []CreateTagBacklog, telemetryData *telemetry.CustomData, con abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, com abaputils.Communication) (err error) {

	connection := con
	connection.XCsrfToken = "fetch"
	connection.URL = con.URL + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Tags"
	resp, err := abaputils.GetHTTPResponse("HEAD", connection, nil, client)
	if err != nil {
		return abaputils.HandleHTTPError(resp, err, "Authentication on the ABAP system failed", con)
	}
	defer resp.Body.Close()

	log.Entry().WithField("StatusCode", resp.Status).WithField("ABAP Endpoint", connection.URL).Debug("Authentication on the ABAP system successful")
	connection.XCsrfToken = resp.Header.Get("X-Csrf-Token")

	errorOccurred := false
	for _, item := range backlog {
		err = createTagsForSingleItem(item, telemetryData, connection, client, com)
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

func createTagsForSingleItem(item CreateTagBacklog, telemetryData *telemetry.CustomData, con abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, com abaputils.Communication) (err error) {

	errorOccurred := false
	for index := range item.tags {
		err = createSingleTag(item, index, telemetryData, con, client, com)
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

func createSingleTag(item CreateTagBacklog, index int, telemetryData *telemetry.CustomData, con abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, com abaputils.Communication) (err error) {

	requestBodyStruct := CreateTagBody{RepositoryName: item.repositoryName, CommitID: item.commitID, Tag: item.tags[index].tagName, Description: item.tags[index].tagDescription}
	requestBodyJson, err := json.Marshal(&requestBodyStruct)
	if err != nil {
		return err
	}

	log.Entry().Debugf("Request body: %s", requestBodyJson)
	resp, err := abaputils.GetHTTPResponse("POST", con, requestBodyJson, client)
	if err != nil {
		errorMessage := "Could not create tag " + requestBodyStruct.Tag + " for repository " + requestBodyStruct.RepositoryName + " with commitID " + requestBodyStruct.CommitID
		err = abaputils.HandleHTTPError(resp, err, errorMessage, con)
		return err
	}
	defer resp.Body.Close()

	// Parse response
	var createTagResponse CreateTagResponse
	var abapResp map[string]*json.RawMessage
	bodyText, _ := io.ReadAll(resp.Body)

	if err = json.Unmarshal(bodyText, &abapResp); err != nil {
		return err
	}
	if err = json.Unmarshal(*abapResp["d"], &createTagResponse); err != nil {
		return err
	}

	con.URL = con.Host + "/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull(guid'" + createTagResponse.UUID + "')"
	err = checkStatus(con, client, com)

	if err == nil {
		log.Entry().Info("Created tag " + requestBodyStruct.Tag + " for repository " + requestBodyStruct.RepositoryName + " with commitID " + requestBodyStruct.CommitID)
	} else {
		log.Entry().Error("NOT created: Tag " + requestBodyStruct.Tag + " for repository " + requestBodyStruct.RepositoryName + " with commitID " + requestBodyStruct.CommitID)
	}

	return err
}

func checkStatus(con abaputils.ConnectionDetailsHTTP, client piperhttp.Sender, com abaputils.Communication) (err error) {
	var status string
	pollIntervall := com.GetPollIntervall()
	count := 0
	for {
		count += 1
		entity, _, err := abaputils.GetStatus("Could not create Tag", con, client)
		if err != nil {
			return err
		}
		status = entity.Status
		if status != "R" {
			if status == "E" {
				err = errors.New("Could not create Tag")
			}
			return err
		}
		if count >= 200 {
			return errors.New("Could not create Tag (Timeout)")
		}
		time.Sleep(pollIntervall)
	}
}

func prepareBacklog(config *abapEnvironmentCreateTagOptions) (backlog []CreateTagBacklog, err error) {

	if config.Repositories != "" && config.RepositoryName != "" {
		return nil, errors.New("Configuring the parameter repositories and the parameter repositoryName at the same time is not allowed")
	}

	if config.RepositoryName != "" && config.CommitID != "" {
		backlog = append(backlog, CreateTagBacklog{repositoryName: config.RepositoryName, commitID: config.CommitID})
	}

	if config.Repositories != "" {
		descriptor, err := abaputils.ReadAddonDescriptor(config.Repositories) //config.Repositories should contain a file name
		if err != nil {
			return nil, err
		}
		for _, repo := range descriptor.Repositories {
			backlogInstance := CreateTagBacklog{repositoryName: repo.Name, commitID: repo.CommitID}
			if config.GenerateTagForAddonComponentVersion && repo.VersionYAML != "" {
				tag := Tag{tagName: "v" + repo.VersionYAML, tagDescription: "Generated by the ABAP Environment Pipeline"}
				backlogInstance.tags = append(backlogInstance.tags, tag)
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

func addTagToList(backlog []CreateTagBacklog, tag string, description string) []CreateTagBacklog {

	for i, item := range backlog {
		tag := Tag{tagName: tag, tagDescription: description}
		backlog[i].tags = append(item.tags, tag)
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
	tags           []Tag
}

type Tag struct {
	tagName        string
	tagDescription string
}

type CreateTagBody struct {
	RepositoryName string `json:"sc_name"`
	CommitID       string `json:"commit_id"`
	Tag            string `json:"tag_name"`
	Description    string `json:"tag_description"`
}

type CreateTagResponse struct {
	UUID string `json:"uuid"`
}
