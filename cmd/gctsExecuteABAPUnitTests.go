package cmd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

var atcFailure, aUnitFailure bool

func gctsExecuteABAPUnitTests(config gctsExecuteABAPUnitTestsOptions, telemetryData *telemetry.CustomData) {

	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	httpClient := &piperhttp.Client{}
	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runGctsExecuteABAPUnitTests(&config, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	if aUnitFailure || atcFailure {

		log.Entry().Fatal("unit test(s) has failed and/or ATC issue(s) found! Check Statistic Analysis Warning for more information!")

	}

}

func runGctsExecuteABAPUnitTests(config *gctsExecuteABAPUnitTestsOptions, httpClient piperhttp.Sender) error {

	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return errors.Wrap(cookieErr, "creating a cookie jar failed")
	}

	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}

	httpClient.SetOptions(clientOptions)

	log.Entry().Infof("start of gCTSExecuteABAPUnitTests step with configuration values: %v", config)

	var objects []repoObject
	var err error

	log.Entry().Infof("scope:", config.Scope)

	switch strings.ToLower(config.Scope) {
	case "local_objects":
		objects, err = getLocalObjects(config, httpClient)
	case "remote_objects":
		objects, err = getRemoteObjects(config, httpClient)
	case "local_packages":
		objects, err = getLocalPackages(config, httpClient)
	case "remote_packages":
		objects, err = getRemotePackages(config, httpClient)
	case "repository":
		objects, err = getRepositoryObjects(config, httpClient)
	case "packages":
		objects, err = getPackages(config, httpClient)
	default:
		objects, err = getLocalObjects(config, httpClient)
	}

	if err != nil {
		log.Entry().WithError(cookieErr).Fatal("failure in getting objects")
	}

	log.Entry().Infof("changed objects:", objects)

	if config.AUnitTest {

		err := executeAUnitTest(config, httpClient, objects)

		if err != nil {
			log.Entry().WithError(err)

		}

		log.Entry().Info("AUnit test run completed successfully. If there are any results from the run, the results are saved in .xml file")

	}

	if config.ATCCheck {

		err = executeATCCheck(config, httpClient, objects)

		if err != nil {
			log.Entry().WithError(err).Fatal("execute ATC Check failed")
		}

		log.Entry().Info("ATCCheck test run completed successfully. If there are any results from the run, the results are saved in .xml file")

	}

	return nil

}

func getLocalObjects(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) ([]repoObject, error) {

	var localObjects []repoObject
	var localObject repoObject

	repository, err := getRepository(config, client)
	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting local changed objects failed")
	}

	resp, err := getObjectDifference(config, repository.Result.CurrentCommit, config.CommitID, client)
	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting local changed objects failed")
	}

	for _, object := range resp.Objects {
		localObject.Object = object.Name
		localObject.Type = object.Type
		localObjects = append(localObjects, localObject)
	}

	return localObjects, nil
}

func getRemoteObjects(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) ([]repoObject, error) {

	var remoteObjects []repoObject
	var remoteObject repoObject
	var lastCommit string

	commitList, err := getCommitList(config, client)

	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting remote changed objects failed")
	}

	for i, commit := range commitList.Commits {
		if commit.ID == config.CommitID {
			lastCommit = commitList.Commits[i+1].ID
			break
		}
	}
	if lastCommit == "" {
		return []repoObject{}, errors.New("last remote commit was not found")

	}

	resp, err := getObjectDifference(config, lastCommit, config.CommitID, client)

	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting remote changed objects failed")
	}

	for _, object := range resp.Objects {
		remoteObject.Object = object.Name
		remoteObject.Type = object.Type
		remoteObjects = append(remoteObjects, remoteObject)
	}

	return remoteObjects, nil
}

func getLocalPackages(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) ([]repoObject, error) {

	var localPackages []repoObject
	var localPackage repoObject

	repository, err := getRepository(config, client)
	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting local packages failed")
	}

	resp, err := getObjectDifference(config, repository.Result.CurrentCommit, config.CommitID, client)

	log.Entry().Info("object delta", resp.Objects)

	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting local packages failed")

	}

	myPackages := map[string]bool{}
	for _, object := range resp.Objects {
		objInfo, err := getObjectInfo(config, client, object.Name, object.Type)
		if err != nil {
			return []repoObject{}, errors.Wrap(err, "getting local packages failed")
		}
		if myPackages[objInfo.Devclass] {

		} else {
			myPackages[objInfo.Devclass] = true
			localPackage.Object = objInfo.Devclass
			localPackage.Type = "DEVC"
			localPackages = append(localPackages, localPackage)
		}

	}

	return localPackages, nil
}

func getRemotePackages(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) ([]repoObject, error) {

	var remotePackages []repoObject
	var remotePackage repoObject
	var lastCommit string

	commitList, err := getCommitList(config, client)

	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting remote packages failed")
	}

	for i, commit := range commitList.Commits {
		if commit.ID == config.CommitID {
			lastCommit = commitList.Commits[i+1].ID
			break
		}
	}

	if lastCommit == "" {
		return []repoObject{}, errors.Wrap(err, "last remote commit was not found")

	}

	resp, err := getObjectDifference(config, lastCommit, config.CommitID, client)
	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting remote packages failed")
	}
	myPackages := map[string]bool{}
	for _, object := range resp.Objects {
		objInfo, err := getObjectInfo(config, client, object.Name, object.Type)
		if err != nil {
			return []repoObject{}, errors.Wrap(err, "last remote commit was not found")
		}
		if myPackages[objInfo.Devclass] {

		} else {
			myPackages[objInfo.Devclass] = true
			remotePackage.Object = objInfo.Devclass
			remotePackage.Type = "DEVC"
			remotePackages = append(remotePackages, remotePackage)
		}

	}

	return remotePackages, nil
}

func getRepositoryObjects(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) ([]repoObject, error) {

	var repositoryObjects objectsResponse

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/objects?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return []repoObject{}, errors.Wrap(httpErr, "could not get repository objects")
	} else if resp == nil {
		return []repoObject{}, errors.New("could not get repository objects: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &repositoryObjects)
	if parsingErr != nil {
		return []repoObject{}, errors.Errorf("%v", parsingErr)
	}

	return repositoryObjects.Objects, nil
}

func getPackages(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) ([]repoObject, error) {

	var packages []repoObject

	resp, err := getRepositoryObjects(config, client)
	if err != nil {
		return []repoObject{}, errors.Wrap(err, "getting packages failed")
	}

	for _, object := range resp {

		if object.Type == "DEVC" {
			packages = append(packages, object)
		}

	}
	return packages, nil
}

func discoverServer(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) (*http.Header, error) {

	url := config.Host +
		"/sap/bc/adt/core/discovery?sap-client=" + config.Client

	header := make(http.Header)
	header.Add("Accept", "application/atomsvc+xml")
	header.Add("x-csrf-token", "fetch")
	header.Add("saml2", "disabled")

	disc, httpErr := client.SendRequest("GET", url, nil, header, nil)

	defer func() {
		if disc != nil && disc.Body != nil {
			disc.Body.Close()
		}
	}()

	if httpErr != nil {
		return nil, errors.Wrap(httpErr, "discovery of the ABAP server failed")
	} else if disc == nil || disc.Header == nil {
		return nil, errors.New("discovery of the ABAP server failed: did not retrieve a HTTP response")
	}

	return &disc.Header, nil
}

func executeAUnitTest(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, objects []repoObject) error {

	log.Entry().Info("execution of unit test has started")

	var innerXml string
	var result runResult

	for _, object := range objects {

		switch object.Type {
		case "CLAS":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/oo/classes/` + object.Object + `"/>`
		case "DEVC":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/packages/` + object.Object + `"/>`

		}

	}

	var xmlBody = []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<aunit:runConfiguration xmlns:aunit="http://www.sap.com/adt/aunit">
			<external>
				<coverage active="false"/>
			</external>
			<options>
				<uriType value="semantic"/>
				<testDeterminationStrategy appendAssignedTestsPreview="true" assignedTests="false" sameProgram="true"/>
				<testRiskLevels critical="true" dangerous="true" harmless="true"/>
				<testDurations long="true" medium="true" short="true"/>
				<withNavigationUri enabled="false"/>
			</options>
			<adtcore:objectSets xmlns:adtcore="http://www.sap.com/adt/core">
			<objectSet kind="inclusive">
		<adtcore:objectReferences>` +
		innerXml +
		`</adtcore:objectReferences>
			</objectSet>
			</adtcore:objectSets>
		</aunit:runConfiguration>`)

	resp, err := runAUnitTest(config, client, xmlBody)
	if err != nil {
		return errors.Wrap(err, "execute of Aunit test has failed")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyXML(resp, &result)
	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
	}

	parsedRes, err := parseAUnitResult(config, client, &result)

	if parsedRes.Version != "" {
		return err

	}

	log.Entry().Info("execution of unit test finished.")
	return nil
}

func runAUnitTest(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, xml []byte) (response *http.Response, err error) {

	url := config.Host +
		"/sap/bc/adt/abapunit/testruns?sap-client=" + config.Client

	discHeader, discError := discoverServer(config, client)

	if discError != nil {
		return response, errors.Wrap(discError, "run of unit tests failed")
	}

	if discHeader.Get("X-Csrf-Token") == "" {

		return response, errors.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	header.Add("Accept", "application/xml")
	header.Add("Content-Type", "application/vnd.sap.adt.abapunit.testruns.result.v1+xml")

	response, httpErr := client.SendRequest("POST", url, bytes.NewBuffer(xml), header, nil)

	/*	defer func() {
			if response != nil && response.Body != nil {
				response.Body.Close()
			}
		}()
	*/
	if httpErr != nil {
		return response, errors.Wrap(httpErr, "run of unit tests failed")
	} else if response == nil {
		return response, errors.New("run of unit tests failed: did not retrieve a HTTP response")
	}

	return response, nil
}

func parseAUnitResult(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, aUnitRunResult *runResult) (parsedResult Checkstyle, error error) {

	log.Entry().Info("parsing unit test result to checkstyle started...")

	var targetDir, FileName string
	var aUnitFile checkstyleFile
	var aUnitError unitError

	repoLayout, err := getRepositoryLayout(config, client)
	if err != nil {
		log.Entry().Error(err)
	}

	repository, err := getRepository(config, client)
	if err != nil {
		log.Entry().Error(err)
	}

	for _, config := range repository.Result.Config {
		if config.Key == "VCS_TARGET_DIR" {
			targetDir = config.Value
		}
	}

	parsedResult.Version = "1.0"
	regexLine := regexp.MustCompile(`.start=\d*`)

	for _, program := range aUnitRunResult.Program {

		objectType := program.Type[0:4]
		objectName := program.Name

		//syntax error use case
		if program.Alerts.Alert.HasSyntaxErrors == "true" {
			aUnitFailure = true
			aUnitError.Source = objectName
			aUnitError.Severity = "error"
			aUnitError.Message = html.UnescapeString(program.Alerts.Alert.Title + " " + program.Alerts.Alert.Details.Detail.AttrText)
			line := regexLine.FindString(program.Alerts.Alert.Stack.StackEntry.URI)
			aUnitError.Line = line[len(`.start=`):]
			aUnitFile.Error = append(aUnitFile.Error, aUnitError)
			aUnitError = unitError{}
		}

		for _, testClass := range program.TestClasses.TestClass {

			for _, testMethod := range testClass.TestMethods.TestMethod {

				aUnitError.Source = testMethod.Name

				if len(testMethod.Alerts.Alert) > 0 {

					for _, testalert := range testMethod.Alerts.Alert {

						switch testalert.Severity {
						case "fatal":
							aUnitFailure = true
							aUnitError.Severity = "error"
						case "critical":
							aUnitFailure = true
							aUnitError.Severity = "error"
						case "tolerable":
							aUnitError.Severity = "warning"
						default:
							aUnitError.Severity = "info"

						}

						for _, detail := range testalert.Details.Detail {
							aUnitError.Message = aUnitError.Message + " " + detail.AttrText
							for _, subdetail := range detail.Details.Detail {

								aUnitError.Message = html.UnescapeString(aUnitError.Message + " " + subdetail.AttrText)
							}

						}
						line := regexLine.FindString(testalert.Stack.StackEntry.URI)
						aUnitError.Line = line[len(`.start=`):]

					}
					aUnitFile.Error = append(aUnitFile.Error, aUnitError)
					aUnitError = unitError{}

				} else {
					log.Entry().Info(testMethod.Name, " - unit test was successful")

				}

			}
		}
		if repoLayout.Layout.ReadableSource == "true" || repoLayout.Layout.ReadableSource == "only" || repoLayout.Layout.ReadableSource == "all" {

			FileName = strings.ToLower(objectName) + "." + strings.ToLower(objectType) + "." + "testclasses.abap"

		} else {

			FileName = "CINC " + objectName + "============CCAU.abap"

		}

		aUnitFile.Name = config.Workspace + "/" + targetDir + "/objects/" + strings.ToUpper(objectType) + "/" + strings.ToUpper(objectName) + "/" + FileName
		parsedResult.File = append(parsedResult.File, aUnitFile)
		aUnitFile = checkstyleFile{}

	}

	body, _ := xml.Marshal(parsedResult)

	writeErr := ioutil.WriteFile("UnitTestResults", body, 0644)

	if writeErr != nil {
		log.Entry().Error("file UnitTestResults could not be created")
		return parsedResult, fmt.Errorf("handling unit test results failed: %w", writeErr)
	}

	log.Entry().Info("parsing of unit test result to checkstyle has finished.")

	return parsedResult, nil

}

func executeATCCheck(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, objects []repoObject) (error error) {

	log.Entry().Info("excecution of ATC checks has started")

	var innerXml string
	var result Worklist

	for _, object := range objects {

		switch object.Type {

		case "CLAS":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/oo/classes/` + object.Object + `"/>`
		case "INTF":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/oo/interfaces/` + object.Object + `"/>`
		case "DEVC":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/packages/` + object.Object + `"/>`
		case "FUGR":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/functions/groups/` + object.Object + `/source/main"/>`
		/*case "FUNC":
		objectInfo, objectErr := resolvePackageForObject(config, client, object.Object, object.Type)
		if objectErr == nil{
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/functions/groups/` + objectInfo.Devclass + `/fmodules/` + object.Object + `/source/main"/>`
		} */
		case "TABU":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/ddic/tables/` + object.Object + `"/>`
		case "DTEL":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/ddic/dataelements/` + object.Object + `"/>`
		case "DOMA":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/ddic/domains/` + object.Object + `"/>`
		case "MSAG":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/messageclass/` + object.Object + `"/>`
		case "PROG":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/programs/programs/` + object.Object + `/source/main"/>`
		default:
			log.Entry().Warning("Object Type" + object.Type + "is not supported!")

		}

	}

	var xmlBody = []byte(`<?xml version="1.0" encoding="UTF-8"?>
	<atc:run xmlns:atc="http://www.sap.com/adt/atc" 
	maximumVerdicts="100">
			<objectSets xmlns:adtcore="http://www.sap.com/adt/core">
			<objectSet kind="inclusive">
 		<adtcore:objectReferences>` + innerXml +
		`</adtcore:objectReferences>
			</objectSet>
			</objectSets>
				</atc:run>`)

	worklist, err := getWorklist(config, client)
	if err != nil {
		return errors.Wrap(err, "execution of ATC Checks failed")
	}

	err = startATCRun(config, client, xmlBody, worklist)

	if err != nil {
		return errors.Wrap(err, "execution of ATC Checks failed")
	}

	resp, err := getATCRun(config, client, worklist)

	if err != nil {
		return errors.Wrap(err, "execution of ATC Checks failed")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyXML(resp, &result)
	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
	}

	atcRes, err := parseATCCheckResult(config, client, &result)

	if atcRes.Version != "" {
		return err

	}

	log.Entry().Info("excecution of ATC checks finished.")
	return nil
}

func startATCRun(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, xml []byte, worklistID string) (err error) {
	log.Entry().Info("start ATC run")
	discHeader, discError := discoverServer(config, client)
	if discError != nil {
		return errors.Wrap(discError, "start of ATC run failed")
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return errors.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	header.Add("Accept", "application/xml")

	url := config.Host +
		"/sap/bc/adt/atc/runs?worklistId=" + worklistID + "?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("POST", url, bytes.NewBuffer(xml), header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return errors.Wrap(httpErr, "start of ATC run failed")
	} else if resp == nil {
		return errors.New("start of ATC run failed: did not retrieve a HTTP response")
	}

	return nil

}

func getATCRun(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, worklistID string) (response *http.Response, err error) {
	log.Entry().Info("get ATC run")
	discHeader, discError := discoverServer(config, client)
	if discError != nil {
		return response, errors.Wrap(discError, "get ATC run failed")
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return response, errors.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	header.Add("Accept", "application/xml")

	url := config.Host +
		"/sap/bc/adt/atc/worklists/" + worklistID + "?sap-client=" + config.Client

	header.Add("Accept", "application/atc.worklist.v1+xml")

	resp, httpErr := client.SendRequest("GET", url, nil, header, nil)

	if httpErr != nil {
		return response, errors.Wrap(httpErr, "get ATC run failed")
	} else if resp == nil {
		return response, errors.New("get ATC run failed: did not retrieve a HTTP response")
	}

	return resp, nil

}

func getWorklist(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) (worklistID string, error error) {

	log.Entry().Info("get Worklist")

	url := config.Host +
		"/sap/bc/adt/atc/worklists?checkVariant=DEFAULT_REMOTE_REF?sap-client=" + config.Client
	discHeader, discError := discoverServer(config, client)

	if discError != nil {
		return worklistID, errors.Wrap(discError, "get worklist failed")
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return worklistID, errors.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	header.Add("Accept", "*/*")

	resp, httpErr := client.SendRequest("POST", url, nil, header, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return worklistID, errors.Wrap(httpErr, "get worklist failed")
	} else if resp == nil {
		return worklistID, errors.New("get worklist failed: did not retrieve a HTTP response")
	}
	location := resp.Header["Location"][0]
	locationSlice := strings.Split(location, "/")
	worklistID = locationSlice[len(locationSlice)-1]

	return worklistID, nil
}

func parseATCCheckResult(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, response *Worklist) (atcResults Checkstyle, error error) {

	log.Entry().Info("conversion of ATC check results to CheckStyle has started")

	var atcFile checkstyleFile
	var fileName, targetDir, subObject string
	var aTCUnitError unitError
	var readableSourceFormat bool
	//var linepointer int

	repoLayout, err := getRepositoryLayout(config, client)
	if err != nil {
		return atcResults, errors.Wrap(err, "could not get repository layout")
	}

	repository, err := getRepository(config, client)
	if err != nil {
		return atcResults, errors.Wrap(err, "getting repository information failed")
	}
	for _, value := range repository.Result.Config {
		if value.Key == "VCS_TARGET_DIR" {
			targetDir = value.Value
		}
	}

	if repoLayout.Layout.ReadableSource == "true" || repoLayout.Layout.ReadableSource == "only" || repoLayout.Layout.ReadableSource == "all" {
		readableSourceFormat = true

	} else {
		readableSourceFormat = false

	}

	atcResults.Version = "1.0"

	for _, object := range response.Objects.Object {

		objectType := object.Type
		objectName := object.Name

		for _, atcworklist := range object.Findings.Finding {

			path, err := url.PathUnescape(atcworklist.Location)

			if err != nil {
				return atcResults, errors.Wrap(err, "conversion of ATC check results to CheckStyle has failed")

			}

			if len(atcworklist.Atcfinding) > 0 {

				priority, err := strconv.Atoi(atcworklist.Priority)

				if err != nil {
					return atcResults, errors.Wrap(err, "conversion of ATC check results to CheckStyle has failed")

				}

				switch priority {
				case 1:
					atcFailure = true
					aTCUnitError.Severity = "error"
				case 2:
					atcFailure = true
					aTCUnitError.Severity = "high"
				case 3:
					aTCUnitError.Severity = "normal"
				default:
					aTCUnitError.Severity = "low"
				}

				if aTCUnitError.Line == "" {

					aTCUnitError.Line, err = findLine(path, readableSourceFormat, objectName, objectType, config.Workspace, targetDir)

					if err != nil {
						return atcResults, errors.Wrap(err, "conversion of ATC check results to CheckStyle has failed")

					}

				}

				if subObject != "" {
					aTCUnitError.Source = objectName + "/" + strings.ToUpper(subObject)
				} else {
					aTCUnitError.Source = objectName
				}

				aTCUnitError.Message = html.UnescapeString(atcworklist.CheckTitle + " " + atcworklist.MessageTitle)
				fileName, err = getFileName(path, readableSourceFormat, objectName)
				if err != nil {
					return atcResults, errors.Wrap(err, "conversion of ATC check results to CheckStyle has failed")

				}
				atcFile.Error = append(atcFile.Error, aTCUnitError)
				aTCUnitError = unitError{}
			}

			if atcFile.Error[0].Message != "" && fileName != "" {

				atcFile.Name = config.Workspace + "/" + targetDir + "/objects/" + strings.ToUpper(objectType) + "/" + strings.ToUpper(objectName) + "/" + fileName
				atcResults.File = append(atcResults.File, atcFile)
				atcFile = checkstyleFile{}
				fileName = ""
			}

		}
	}

	atcBody, _ := xml.Marshal(atcResults)

	writeErr := ioutil.WriteFile("ATCResults", atcBody, 0644)

	if writeErr != nil {
		log.Entry().Error("ATCResults could not be created")
		return atcResults, fmt.Errorf("handling atc results failed: %w", writeErr)
	}
	log.Entry().Info("parsing ATC check results to CheckStyle has finished.")
	return atcResults, writeErr
}

func constructPath(workspace string, path string, targetDir string, objectName string, objectType string) (filePath string, error error) {

	fileName, err := getFileName(path, true, objectName)
	if err != nil {
		return filePath, errors.Wrap(err, "path could not be constructed")

	}

	filePath = workspace + "/" + targetDir + "/objects/" + strings.ToUpper(objectType) + "/" + strings.ToUpper(objectName) + "/" + fileName
	return filePath, nil

}

func findLine(path string, readableSource bool, objName string, objectType string, workspace string, targetDir string) (line string, error error) {
	var linepointer int
	if readableSource {
		if strings.Contains(path, "CLAS/OSO") || strings.Contains(path, "CLAS/OM") || strings.Contains(path, "CLAS/OSI") {

			filePath, err := constructPath(workspace, path, targetDir, objName, objectType)
			if err != nil {
				return line, errors.Wrap(err, "find Line has failed")

			}
			rawfilecontent, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Println("File reading error", err)
				return
			}
			filecontent := string(rawfilecontent)

			splittedfilecontent := strings.Split(filecontent, "\n")
			for line, linecontent := range splittedfilecontent {
				regexSubObj := regexp.MustCompile(`includes\/\w*`)
				subObject := regexSubObj.FindString(path)
				subObject = subObject[len(`includes/`):]
				if strings.Contains(linecontent, "protected section.") || strings.Contains(linecontent, "method"+subObject) || strings.Contains(linecontent, "private section.") {
					linepointer = line
					break
				}

			}

			regexLine := regexp.MustCompile(`.start=\d*`)
			linestring := regexLine.FindString(path)
			if linestring != "" {

				lineint, err := strconv.Atoi(linestring[len(`.start=`):])
				if err == nil {
					line = strconv.Itoa(linepointer + lineint)

				}

			}

		} else {
			regexLine := regexp.MustCompile(`.start=\d*`)
			linestring := regexLine.FindString(path)
			if linestring != "" {
				line = linestring[len(`.start=`):]

			}

		}
	} else {

		regexLine := regexp.MustCompile(`.start=\d*`)
		linestring := regexLine.FindString(path)
		if linestring != "" {
			line = linestring[len(`.start=`):]

		}

	}

	return line, nil
}
func getFileName(path string, readableSource bool, objName string) (fileName string, error error) {

	path, err := url.PathUnescape(path)

	if err != nil {
		return fileName, errors.Wrap(err, "get file name has failed")

	}

	//  INTERFACES
	regexInterface := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/interfaces\/\w*`)
	intf := regexInterface.FindString(path)
	if intf != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".intf.abap"
		} else {
			fileName = "REPS " + strings.ToUpper(objName) + "====IU.abap"
		}

	}
	// CLASSES DEFINITIONS
	regexClasDef := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/includes\/definitions\/`)
	clasDef := regexClasDef.FindString(path)
	if clasDef != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + "." + ".clas.definitions.abap"
		} else {
			fileName = "CINC " + objName + "=======CCDEF.abap"
		}

	}

	// CLASSES IMPLEMENTATIONS
	regexClasImpl := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/includes\/implementations\/`)
	clasImpl := regexClasImpl.FindString(path)
	if clasImpl != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + "." + ".clas.implementations.abap"
		} else {
			fileName = "CINC " + objName + "=======CCIMP.abap"
		}

	}

	// CLASSES MACROS
	regexClasMacro := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/includes\/macros\/`)
	clasMacro := regexClasMacro.FindString(path)
	if clasMacro != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + "." + ".clas.macros.abap"
		} else {
			fileName = "CINC " + objName + "=======CCMAC.abap"
		}

	}

	// TEST CLASSES
	regexTestClass := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/includes\/testclasses\/`)
	testClass := regexTestClass.FindString(path)
	if testClass != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + "." + ".clas.testclasses.abap"
		} else {
			fileName = "CINC " + objName + "=======CCAU.abap"
		}

	}

	// CLASS PROTECTED
	regexClasProtected := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/source\/main#type=CLAS\/OSO`)
	classProtected := regexClasProtected.FindString(path)
	if classProtected != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + "." + ".clas.abap"
		} else {
			fileName = "CPRO " + objName + ".abap"
		}

	}

	// CLASS PRIVATE
	regexClasPrivate := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/source\/main#type=CLAS\/OSI`)
	classPrivate := regexClasPrivate.FindString(path)
	if classPrivate != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + "." + ".clas.abap"
		} else {
			fileName = "CPRI " + objName + ".abap"
		}

	}

	// CLASS METHOD
	regexClasMethod := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/source\/main#type=CLAS\/OM`)
	classMethod := regexClasMethod.FindString(path)
	if classMethod != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.global.abap"
		} else {
			fileName = "METH " + objName + ".abap"
		}

	}

	// CLASS PUBLIC
	regexClasPublic := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/source\/main#start`)
	classPublic := regexClasPublic.FindString(path)
	if classPublic != "" {

		if readableSource {

			fileName = strings.ToLower(objName) + "." + ".clas.abap"
		} else {
			fileName = "CPUB " + objName + ".abap"
		}

	}

	// FUNCTION INCLUDE
	regexFuncIncl := regexp.MustCompile(`\/sap\/bc\/adt\/functions/groups\/\w*\/includes/\w*`)

	funcIncl := regexFuncIncl.FindString(path)
	if funcIncl != "" {

		regexSubObj := regexp.MustCompile(`includes\/\w*`)
		subObject := regexSubObj.FindString(path)
		subObject = subObject[len(`includes/`):]

		if readableSource {

			fileName = strings.ToLower(objName) + ".fugr." + strings.ToLower(subObject) + ".reps.abap"
		} else {
			fileName = "REPS " + subObject + ".abap"
		}

	}

	// FUNCTION MODULE
	regexFuncMod := regexp.MustCompile(`\/sap\/bc\/adt\/functions/groups\/\w*\/fmodules/\w*`)
	funcMod := regexFuncMod.FindString(path)
	if funcMod != "" {

		regexSubObj := regexp.MustCompile(`includes\/\w*`)
		subObject := regexSubObj.FindString(path)
		subObject = subObject[len(`includes/`):]

		if readableSource {

			fileName = strings.ToLower(subObject) + ".func.abap"
		} else {
			fileName = "FUNC " + subObject + ".abap"
		}

	}

	return fileName, nil

}

func getRepository(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) (repositoryResponseBody, error) {
	var repositoryResponse repositoryResponseBody
	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"?sap-client=" + config.Client
	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return repositoryResponseBody{}, errors.Wrap(httpErr, "could not get repository")
	} else if resp == nil {
		return repositoryResponseBody{}, errors.New("could not get repository: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &repositoryResponse)
	if parsingErr != nil {
		return repositoryResponseBody{}, errors.Errorf("%v", parsingErr)
	}

	return repositoryResponse, nil

}

func getRepositoryLayout(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) (repositoryLayout, error) {

	var repoLayoutResponse repositoryLayout
	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/layout?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return repositoryLayout{}, errors.Wrap(httpErr, "could not get repository layout")
	} else if resp == nil {
		return repositoryLayout{}, errors.New("could not get repository layout: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &repoLayoutResponse)
	if parsingErr != nil {
		return repositoryLayout{}, errors.Errorf("%v", parsingErr)
	}

	return repoLayoutResponse, nil
}

func getRepositoryHistory(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) (getRepoHistoryResponseBody, error) {

	var historyResponse getRepoHistoryResponseBody

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/getHistory?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return getRepoHistoryResponseBody{}, errors.Wrap(httpErr, "getting repository history failed")
	} else if resp == nil {
		return getRepoHistoryResponseBody{}, errors.New("getting repository history failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &historyResponse)
	if parsingErr != nil {
		return getRepoHistoryResponseBody{}, errors.Errorf("%v", parsingErr)
	}

	return historyResponse, nil
}

func getCommitList(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender) (commitsResponseBody, error) {

	var commitResponse commitsResponseBody
	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/getCommit?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return commitsResponseBody{}, errors.Wrap(httpErr, "getting repository history failed")
	} else if resp == nil {
		return commitsResponseBody{}, errors.New("getting repository history failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &commitResponse)
	if parsingErr != nil {
		return commitsResponseBody{}, errors.Errorf("%v", parsingErr)
	}

	return commitResponse, nil
}

func getObjectDifference(config *gctsExecuteABAPUnitTestsOptions, fromCommit string, toCommit string, client piperhttp.Sender) (objectsResponseBody, error) {
	var objectResponse objectsResponseBody
	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/compareCommits?fromCommit=" + fromCommit + "&toCommit=" + toCommit + "&sap-client=" + config.Client

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return objectsResponseBody{}, errors.Wrap(httpErr, "getting object difference failed")
	} else if resp == nil {
		return objectsResponseBody{}, errors.New("getting object difference failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &objectResponse)
	if parsingErr != nil {
		return objectsResponseBody{}, errors.Errorf("%v", parsingErr)
	}
	return objectResponse, nil
}

func getObjectInfo(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, objectName string, objectType string) (objectMetaInfo, error) {

	var objectMetInfoResponse objectMetaInfo
	url := config.Host +
		"/sap/bc/cts_abapvcs/objects/" + objectType + "/" + objectName +
		"?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return objectMetaInfo{}, errors.Wrap(httpErr, "resolve package failed")
	} else if resp == nil {
		return objectMetaInfo{}, errors.New("resolve package failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &objectMetInfoResponse)
	if parsingErr != nil {
		return objectMetaInfo{}, errors.Errorf("%v", parsingErr)
	}
	return objectMetInfoResponse, nil

}

/****************************************************************************************************
*******************************************************************************************************/

func executeUnitTestV2(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, objects []repoObject) error {

	var maxTimeOut int64

	const defaultMaxTimeOut = 10000

	//	if config.MaxTimeOut != 0 {
	//		maxTimeOut = int64(config.MaxTimeOut)
	//	} else {
	maxTimeOut = defaultMaxTimeOut
	//	}

	runId, err := executeTestRun(config, client, objects)

	if err != nil {
		return errors.Wrap(err, "execution of unit tests failed")
	}
	initialTime := time.Now().Unix()
	for {

		statusResponse, err := getRunStatus(config, client, runId)

		if err != nil {
			return errors.Wrap(err, "execution of unit tests failed")
		}

		currentTime := time.Now().Unix()
		timeDuration := currentTime - initialTime
		log.Entry().
			Info("Status", statusResponse.Progress.Status)

		if statusResponse.Progress.Status == "FINISHED" || timeDuration > maxTimeOut {
			break

		}
	}
	log.Entry().
		Info("Get Unit Test Result")

	testResults, err := getTestResults(config, client, runId)

	log.Entry().
		Info("Test Result", testResults)

	if testResults.Failures != "0" || testResults.Errors != "0" {

		return errors.Wrap(err, "execution of unit tests failed")

	}

	return nil
}

func executeATCV2(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, objects []repoObject) error {

	var maxTimeOut int64
	var ATCStatus ATCRun
	var ATCId string
	const defaultMaxTimeOut = 10000
	//	if config.MaxTimeOut != 0 {
	//		maxTimeOut = int64(config.MaxTimeOut)
	//	} else {
	maxTimeOut = defaultMaxTimeOut
	//	}

	runId, err := startATCRunV(config, client, objects)

	if err != nil {
		return errors.Wrap(err, "execution of atc failed")
	}

	initialTime := time.Now().Unix()
	for {
		log.Entry().
			Info("Start Check ATC Status")
		ATCStatus, err = checkATCStatus(config, client, runId)

		if err != nil {
			return errors.Wrap(err, "execution of unit tests failed")
		}

		currentTime := time.Now().Unix()
		timeDuration := currentTime - initialTime
		log.Entry().
			Info("Time duration", timeDuration)
		log.Entry().
			Info("ATC Status", ATCStatus.Status)
		if ATCStatus.Status == "Completed" || ATCStatus.Status == "Not Created" || ATCStatus.Status == "" || timeDuration > maxTimeOut {
			break

		}

	}

	if len(ATCStatus.Link) > 0 {
		location := ATCStatus.Link[0].Key
		locationSlice := strings.Split(location, "/")
		ATCId = locationSlice[len(locationSlice)-1]
		log.Entry().
			Info("Start ATC Result")
		err := getATCResult(config, client, ATCId)
		if err != nil {
			return errors.Wrap(err, "execution of unit tests failed")
		}
	} else {

		return fmt.Errorf("could not get any response from ATC poll: %w", errors.New("status from ATC run is empty. Either it's not an ABAP system or ATC run hasn't started"))
	}

	return nil

}

func runUnitTests(config *gctsExecuteABAPUnitTestsOptions, httpClient piperhttp.Sender) error {

	cookieJar, cookieErr := cookiejar.New(nil)
	if cookieErr != nil {
		return errors.Wrap(cookieErr, "execution of unit tests failed")
	}
	clientOptions := piperhttp.ClientOptions{
		CookieJar: cookieJar,
		Username:  config.Username,
		Password:  config.Password,
	}
	httpClient.SetOptions(clientOptions)

	var repoObjects []repoObject
	var getPackageErr error

	/*	if config.Scope == "CHANGED" {

			repoObjects, getPackageErr = getChangedObjects(config, httpClient)

		} else if config.Scope == "PACKAGE" {

			repoObjects, getPackageErr = getPackageObjects(config, httpClient)

		} else if config.Scope == "REPOSITORY" {

			repoObjects, getPackageErr = getRepositoryObjects(config, httpClient)

		}
	*/
	if getPackageErr != nil {
		return errors.Wrap(getPackageErr, "execution of unit tests failed")
	}

	discHeader, discError := discoverServer(config, httpClient)

	if discError != nil {
		return errors.Wrap(discError, "execution of unit tests failed")
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return errors.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	//header.Add("Accept", "application/xml")
	header.Add("Accept", "application/vnd.sap.adt.api.abapunit.run.v1+xml")
	//header.Add("Content-Type", "application/vnd.sap.adt.abapunit.testruns.result.v1+xml")
	header.Add("Content-Type", "application/vnd.sap.adt.api.abapunit.run.v1+xml")

	/*	executeTestsErr := executeTestsForObject(config, httpClient, header, repoObjects)
		if executeTestsErr != nil {
			return errors.Wrap(executeTestsErr, "execution of unit tests failed")
		}
	*/
	runID, startATCErr := startATCRunV(config, httpClient, repoObjects)
	if startATCErr != nil {
		return errors.Wrap(startATCErr, "execution of unit tests failed")
	}

	status, checkATCStatusErr := checkATCStatus(config, httpClient, runID)

	if checkATCStatusErr != nil {
		return errors.Wrap(checkATCStatusErr, "execution of unit tests failed")
	}

	getATCResultErr := getATCResult(config, httpClient, status.Status)

	if getATCResultErr != nil {
		return errors.Wrap(getATCResultErr, "execution of unit tests failed")
	}

	/*for _, object := range repoObjects {
		executeTestsErr := executeTestsForObject(config, httpClient, header, object.Object, object.Type)
		//executeTestsErr := executeTestsForPackage(config, httpClient, header, object)

		if executeTestsErr != nil {
			return errors.Wrap(executeTestsErr, "execution of unit tests failed")
		}
	}
	*/
	log.Entry().
		WithField("repository", config.Repository).
		Info("all unit tests were successful")

	return nil
}

func startATCRunV(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, objects []repoObject) (runId string, error error) {

	discHeader, discError := discoverServer(config, client)

	var xmlBody []byte

	if discError != nil {
		return runId, errors.Wrap(discError, "execution of unit tests failed")
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return runId, errors.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	header.Add("Accept", "application/vnd.sap.atc.run.parameters.v1+xml")
	header.Add("Content-Type", "application/vnd.sap.atc.run.parameters.v1+xml")

	var innerXml string
	for _, object := range objects {

		if object.Type == "DEVC" {

			innerXml = innerXml + `<obj:package value="` + object.Object + `" includeSubpackages="true"/>`
		}
	}

	//	if config.CheckVariant != "" {

	//		xmlBody = []byte(`<?xml version="1.0" encoding="UTF-8"?><atc:runparameters xmlns:atc="http://www.sap.com/adt/atc"
	//                         xmlns:obj="http://www.sap.com/adt/objectset" checkVariant="` + config.CheckVariant +
	//			`"> <obj:objectSet><obj:packages>` + innerXml + `</obj:packages></obj:objectSet></atc:runparameters>`)

	//	} else {

	xmlBody = []byte(`<?xml version="1.0" encoding="UTF-8"?><atc:runparameters xmlns:atc="http://www.sap.com/adt/atc"
                         xmlns:obj="http://www.sap.com/adt/objectset"><obj:objectSet><obj:packages>` + innerXml + `</obj:packages></obj:objectSet></atc:runparameters>`)

	//	}

	url := config.Host + "/sap/bc/adt/api/atc/runs?clientWait=false"

	resp, httpErr := client.SendRequest("POST", url, bytes.NewBuffer(xmlBody), header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return runId, errors.Wrap(httpErr, "execution of ATC Checks failed")
	} else if resp == nil {
		return runId, errors.New("execution of ATC Checks failed: did not retrieve a HTTP response")
	}

	location := resp.Header["Location"][0]
	locationSlice := strings.Split(location, "/")
	runId = locationSlice[len(locationSlice)-1]

	return runId, nil

}

func checkATCStatus(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, runId string) (status ATCRun, err error) {

	url := config.Host +
		"/sap/bc/adt/atc/runs/" + runId + "?sap-client=" + config.Client

	header := make(http.Header)

	header.Add("Accept", "application/vnd.sap.atc.run.v1+xml")

	resp, httpErr := client.SendRequest("GET", url, nil, header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return status, errors.Wrap(httpErr, "execution of ATC checks failed")
	} else if resp == nil {
		return status, errors.New("execution of ATC checks failed: did not retrieve a HTTP response")
	}
	//	statusResponse := new(ATCRun)
	parsingErr := piperhttp.ParseHTTPResponseBodyXML(resp, &status)
	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
	}

	return status, nil

}

func getATCResult(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, resultID string) (err error) {

	url := config.Host +
		"/sap/bc/adt/api/atc/results/" + resultID + "?sap-client=" + config.Client

	header := make(http.Header)

	header.Add("Accept", "application/vnd.sap.atc.checkstyle.v1+xml")

	resp, httpErr := client.SendRequest("GET", url, nil, header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return errors.Wrap(httpErr, "execution of ATC checks failed")
	} else if resp == nil {
		return errors.New("execution of ATC checks failed: did not retrieve a HTTP response")
	}

	var body []byte
	const atcResultFileName = "ATCResults"
	if httpErr == nil {
		body, err = ioutil.ReadAll(resp.Body)
	}
	if err == nil {
		defer resp.Body.Close()
		err = parseATCResponseResult(body, atcResultFileName)
	}
	if err != nil {
		return fmt.Errorf("handling ATC result failed: %w", err)
	}
	err = ioutil.WriteFile(atcResultFileName, body, 0644)
	return nil

}

func parseATCResponseResult(body []byte, atcResultFileName string) (err error) {
	if len(body) == 0 {
		return fmt.Errorf("parsing ATC result failed: %w", errors.New("body is empty, can't parse empty body"))
	}

	parsedXML := new(ATCFiles)
	xml.Unmarshal([]byte(body), &parsedXML)
	if len(parsedXML.Files) == 0 {
		log.Entry().Info("there were no results from this run, most likely the checked Package are empty or contain no ATC findings")
	}

	err = ioutil.WriteFile(atcResultFileName, body, 0644)
	if err == nil {
		var reports []piperutils.Path
		reports = append(reports, piperutils.Path{Target: atcResultFileName, Name: "ATC Results", Mandatory: true})
		piperutils.PersistReportsAndLinks("gctsExecuteABAPUnitTests", "", reports, nil)
		for _, s := range parsedXML.Files {
			for _, t := range s.ATCErrors {
				log.Entry().Error("Error in file " + s.Key + ": " + t.Key)
			}
		}
	}
	if err != nil {
		return fmt.Errorf("writing results to XML file failed: %w", err)
	}
	return nil
}

func executeTestRun(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, objects []repoObject) (runId string, error error) {

	var objectXml string
	var packageXml string
	packageXml = `<osl:set xsi:type="packageSet">`
	objectXml = `<osl:set xsi:type="flatObjectSet">`
	for _, object := range objects {

		if object.Type == "DEVC" {
			packageXml = packageXml + `<osl:package name="` + object.Object + `" includeSubpackages="true"/>"`
		} else {
			objectXml = objectXml + `<osl:object name="` + object.Object + `" type="` + object.Type + `"/>`
		}
	}
	packageXml = packageXml + `</osl:set>`
	objectXml = objectXml + `</osl:set>`
	var xmlBody = []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<aunit:run title="My Run" context="AIE Integration Test" xmlns:aunit="http://www.sap.com/adt/api/aunit">
		  <aunit:options>
			<aunit:measurements/>
			<aunit:scope ownTests="true" foreignTests="true"/>
			<aunit:riskLevel harmless="true" dangerous="true" critical="true"/>
			<aunit:duration short="true" medium="true" long="true"/>
		  </aunit:options>
		  <osl:objectSet xsi:type="unionSet" xmlns:osl="http://www.sap.com/api/osl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">` + packageXml + objectXml +
		`</osl:objectSet>
		</aunit:run>`)

	discHeader, discError := discoverServer(config, client)

	if discError != nil {
		return runId, errors.Wrap(discError, "execution of unit tests failed")
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return runId, errors.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))

	header.Add("Accept", "application/vnd.sap.adt.api.abapunit.run.v1+xml")
	header.Add("Content-Type", "application/vnd.sap.adt.api.abapunit.run.v1+xml")

	url := config.Host +
		"/sap/bc/adt/api/abapunit/runs?sap-client=" + config.Client

	resp, httpErr := client.SendRequest("POST", url, bytes.NewBuffer(xmlBody), header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return runId, errors.Wrap(httpErr, "execution of unit tests failed")
	} else if resp == nil {
		return runId, errors.New("execution of unit tests failed: did not retrieve a HTTP response")
	}

	location := resp.Header["Location"][0]
	locationSlice := strings.Split(location, "/")
	runId = locationSlice[len(locationSlice)-1]

	return runId, nil
}

func getRunStatus(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, runId string) (status run, error error) {

	url := config.Host +
		"/sap/bc/adt/api/abapunit/runs/" + runId + "?sap-client=" + config.Client

	header := make(http.Header)

	header.Add("Accept", "application/vnd.sap.adt.api.abapunit.run-status.v1+xml")

	resp, httpErr := client.SendRequest("GET", url, nil, header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return status, errors.Wrap(httpErr, "execution of unit tests failed")
	} else if resp == nil {
		return status, errors.New("execution of unit tests failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyXML(resp, &status)
	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
	}

	return status, nil
}

func getTestResults(config *gctsExecuteABAPUnitTestsOptions, client piperhttp.Sender, runId string) (results Testsuites, error error) {

	var response Testsuites
	var UnitTestResults Checkstyle
	var File checkstyleFile
	var UnitError unitError
	var FileName string
	var VcsTargetDir string
	var Source string
	var objectType string
	var objectName string

	url := config.Host +
		"/sap/bc/adt/api/abapunit/results/" + runId + "?sap-client=" + config.Client

	header := make(http.Header)

	header.Add("Accept", "application/vnd.sap.adt.api.junit.run-result.v1+xml")

	resp, httpErr := client.SendRequest("GET", url, nil, header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return results, errors.Wrap(httpErr, "execution of unit tests failed")
	} else if resp == nil {
		return results, errors.New("execution of unit tests failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyXML(resp, &response)
	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
	}

	layout, layouterr := getRepositoryLayout(config, client)
	if layouterr != nil {
		return response, fmt.Errorf("getting repository layout failed: %w", layouterr)
	}

	repository, repoErr := getRepository(config, client)
	if repoErr != nil {
		return response, fmt.Errorf("getting repository information failed: %w", repoErr)
	}
	for _, value := range repository.Result.Config {
		if value.Key == "VCS_TARGET_DIR" {
			VcsTargetDir = value.Value
		}
	}

	for tskey, _ := range response.Testsuite {

		tests, testerr := strconv.Atoi(response.Testsuite[tskey].Tests)
		if testerr != nil {
			log.Entry().Warning(testerr)
		}
		if tests != 0 {

			/*		failures, falirerr := strconv.Atoi(response.Testsuite[key].Failures)
					if falirerr != nil {
						log.Entry().Warning(falirerr)

					}

					} */
			for tckey, _ := range response.Testsuite[tskey].Testcase {
				asserts, assertserr := strconv.Atoi(response.Testsuite[tskey].Testcase[tckey].Asserts)
				if assertserr != nil {
					log.Entry().Warning(assertserr)
				}

				if asserts == 0 {

					UnitError.Source = response.Testsuite[tskey].Testcase[tckey].Name
					UnitError.Severity = "low"
					UnitError.Message = "test case is successful"
					UnitError.Line = ""
					File.Error = append(File.Error, UnitError)
					regexObjectName := regexp.MustCompile(`:[a-zA-Z0-9_]*-`)
					regexObjectType := regexp.MustCompile(`.[a-zA-Z]*:`)
					preobjectName := regexObjectName.FindString(response.Testsuite[tskey].Testcase[tckey].Classname)
					preobjectType := regexObjectType.FindString(response.Testsuite[tskey].Testcase[tckey].Classname)
					objectType = preobjectType[1 : len(preobjectType)-1]
					objectName = preobjectName[1 : len(preobjectName)-1]

				} else {
					UnitError.Source = response.Testsuite[tskey].Testcase[tckey].Name
					UnitError.Severity = "error"

					UnitError.Message = html.UnescapeString(response.Testsuite[tskey].Testcase[tckey].Failure.Text)

					regexLine := regexp.MustCompile(`Line: <\d*>`)
					//	re2 := regexp.MustCompile(`\d+`)

					linestring := regexLine.FindString(UnitError.Message)
					UnitError.Line = linestring[7 : len(linestring)-1]
					File.Error = append(File.Error, UnitError)
					regexObjectName := regexp.MustCompile(`:[a-zA-Z0-9_]*-`)
					regexObjectType := regexp.MustCompile(`.[a-zA-Z]*:`)
					preobjectName := regexObjectName.FindString(response.Testsuite[tskey].Testcase[tckey].Classname)
					preobjectType := regexObjectType.FindString(response.Testsuite[tskey].Testcase[tckey].Classname)
					objectType = preobjectType[1 : len(preobjectType)-1]
					objectName = preobjectName[1 : len(preobjectName)-1]

				}
			}
		} else {

			log.Entry().Warning("No Unit Tests were found!")
		}

		//	if failures != 0 {

		//		for i := 0; i < failures; i++ {

		if layout.Layout.ReadableSource == "true" || layout.Layout.ReadableSource == "only" || layout.Layout.ReadableSource == "all" {

			FileName = objectName + "." + objectType + "." + "testclasses.abap"

		} else {

			FileName = "CINC " + objectName + "============CCAU.abap"

		}

		if layout.Layout.Subdirectory != "" {

			Source = layout.Layout.Subdirectory

		} else if VcsTargetDir != "" {

			Source = VcsTargetDir

		}

		File.Name = config.Workspace + "/" + Source + "/objects/" + strings.ToUpper(objectType) + "/" + strings.ToUpper(objectName) + "/" + FileName
		UnitTestResults.File = append(UnitTestResults.File, File)
		File = checkstyleFile{}
		UnitError = unitError{}

	}

	//UnitError.Line = re2.FindString(linestring)

	UnitTestResults.Version = "1.0"

	const UnitTestFileName = "UnitTestResults"

	body, _ := xml.Marshal(UnitTestResults)

	err := ioutil.WriteFile(UnitTestFileName, body, 0644)

	if err != nil {
		return response, fmt.Errorf("handling unit test results failed: %w", err)
	}

	return response, nil
}

type ATCRun struct {
	XMLName xml.Name  `xml:"run"`
	Status  string    `xml:"status,attr"`
	Link    []ATCLink `xml:"link"`
}

//Link of XML object

type Worklist struct {
	XMLName             xml.Name `xml:"worklist"`
	Text                string   `xml:",chardata"`
	ID                  string   `xml:"id,attr"`
	Timestamp           string   `xml:"timestamp,attr"`
	UsedObjectSet       string   `xml:"usedObjectSet,attr"`
	ObjectSetIsComplete string   `xml:"objectSetIsComplete,attr"`
	Atcworklist         string   `xml:"atcworklist,attr"`
	ObjectSets          struct {
		Text      string `xml:",chardata"`
		ObjectSet []struct {
			Text  string `xml:",chardata"`
			Name  string `xml:"name,attr"`
			Title string `xml:"title,attr"`
			Kind  string `xml:"kind,attr"`
		} `xml:"objectSet"`
	} `xml:"objectSets"`
	Objects struct {
		Text   string `xml:",chardata"`
		Object []struct {
			Text        string `xml:",chardata"`
			URI         string `xml:"uri,attr"`
			Type        string `xml:"type,attr"`
			Name        string `xml:"name,attr"`
			PackageName string `xml:"packageName,attr"`
			Author      string `xml:"author,attr"`
			Atcobject   string `xml:"atcobject,attr"`
			Adtcore     string `xml:"adtcore,attr"`
			Findings    struct {
				Text    string `xml:",chardata"`
				Finding []struct {
					Text              string `xml:",chardata"`
					URI               string `xml:"uri,attr"`
					Location          string `xml:"location,attr"`
					Processor         string `xml:"processor,attr"`
					LastChangedBy     string `xml:"lastChangedBy,attr"`
					Priority          string `xml:"priority,attr"`
					CheckId           string `xml:"checkId,attr"`
					CheckTitle        string `xml:"checkTitle,attr"`
					MessageId         string `xml:"messageId,attr"`
					MessageTitle      string `xml:"messageTitle,attr"`
					ExemptionApproval string `xml:"exemptionApproval,attr"`
					ExemptionKind     string `xml:"exemptionKind,attr"`
					Checksum          string `xml:"checksum,attr"`
					QuickfixInfo      string `xml:"quickfixInfo,attr"`
					Atcfinding        string `xml:"atcfinding,attr"`
					Link              struct {
						Text string `xml:",chardata"`
						Href string `xml:"href,attr"`
						Rel  string `xml:"rel,attr"`
						Type string `xml:"type,attr"`
						Atom string `xml:"atom,attr"`
					} `xml:"link"`
					Quickfixes struct {
						Text      string `xml:",chardata"`
						Manual    string `xml:"manual,attr"`
						Automatic string `xml:"automatic,attr"`
						Pseudo    string `xml:"pseudo,attr"`
					} `xml:"quickfixes"`
				} `xml:"finding"`
			} `xml:"findings"`
		} `xml:"object"`
	} `xml:"objects"`
}

type ATCLink struct {
	Key   string `xml:"href,attr"`
	Value string `xml:",chardata"`
}
type gctsException struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Code        int    `json:"code"`
}

type gctsLogs struct {
	Time     int    `json:"time"`
	User     string `json:"user"`
	Section  string `json:"section"`
	Action   string `json:"action"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Code     string `json:"code"`
}

type runResult struct {
	XMLName xml.Name `xml:"runResult"`
	Text    string   `xml:",chardata"`
	Aunit   string   `xml:"aunit,attr"`
	Program []struct {
		Text    string `xml:",chardata"`
		URI     string `xml:"uri,attr"`
		Type    string `xml:"type,attr"`
		Name    string `xml:"name,attr"`
		URIType string `xml:"uriType,attr"`
		Adtcore string `xml:"adtcore,attr"`
		Alerts  struct {
			Text  string `xml:",chardata"`
			Alert struct {
				Text            string `xml:",chardata"`
				HasSyntaxErrors string `xml:"hasSyntaxErrors,attr"`
				Kind            string `xml:"kind,attr"`
				Severity        string `xml:"severity,attr"`
				Title           string `xml:"title"`
				Details         struct {
					Text   string `xml:",chardata"`
					Detail struct {
						Text     string `xml:",chardata"`
						AttrText string `xml:"text,attr"`
					} `xml:"detail"`
				} `xml:"details"`
				Stack struct {
					Text       string `xml:",chardata"`
					StackEntry struct {
						Text        string `xml:",chardata"`
						URI         string `xml:"uri,attr"`
						Description string `xml:"description,attr"`
					} `xml:"stackEntry"`
				} `xml:"stack"`
			} `xml:"alert"`
		} `xml:"alerts"`

		TestClasses struct {
			Text      string `xml:",chardata"`
			TestClass []struct {
				Text             string `xml:",chardata"`
				URI              string `xml:"uri,attr"`
				Type             string `xml:"type,attr"`
				Name             string `xml:"name,attr"`
				URIType          string `xml:"uriType,attr"`
				NavigationURI    string `xml:"navigationUri,attr"`
				DurationCategory string `xml:"durationCategory,attr"`
				RiskLevel        string `xml:"riskLevel,attr"`
				TestMethods      struct {
					Text       string `xml:",chardata"`
					TestMethod []struct {
						Text          string `xml:",chardata"`
						URI           string `xml:"uri,attr"`
						Type          string `xml:"type,attr"`
						Name          string `xml:"name,attr"`
						ExecutionTime string `xml:"executionTime,attr"`
						URIType       string `xml:"uriType,attr"`
						NavigationURI string `xml:"navigationUri,attr"`
						Unit          string `xml:"unit,attr"`
						Alerts        struct {
							Text  string `xml:",chardata"`
							Alert []struct {
								Text     string `xml:",chardata"`
								Kind     string `xml:"kind,attr"`
								Severity string `xml:"severity,attr"`
								Title    string `xml:"title"`
								Details  struct {
									Text   string `xml:",chardata"`
									Detail []struct {
										Text     string `xml:",chardata"`
										AttrText string `xml:"text,attr"`
										Details  struct {
											Text   string `xml:",chardata"`
											Detail []struct {
												Text     string `xml:",chardata"`
												AttrText string `xml:"text,attr"`
											} `xml:"detail"`
										} `xml:"details"`
									} `xml:"detail"`
								} `xml:"details"`
								Stack struct {
									Text       string `xml:",chardata"`
									StackEntry struct {
										Text        string `xml:",chardata"`
										URI         string `xml:"uri,attr"`
										Type        string `xml:"type,attr"`
										Name        string `xml:"name,attr"`
										Description string `xml:"description,attr"`
									} `xml:"stackEntry"`
								} `xml:"stack"`
							} `xml:"alert"`
						} `xml:"alerts"`
					} `xml:"testMethod"`
				} `xml:"testMethods"`
			} `xml:"testClass"`
		} `xml:"testClasses"`
	} `xml:"program"`
}

type run struct {
	XMLName  xml.Name `xml:"run"`
	Text     string   `xml:",chardata"`
	Title    string   `xml:"title,attr"`
	Context  string   `xml:"context,attr"`
	Aunit    string   `xml:"aunit,attr"`
	Progress struct {
		Text       string `xml:",chardata"`
		Status     string `xml:"status,attr"`
		Percentage string `xml:"percentage,attr"`
	} `xml:"progress"`
	ExecutedBy struct {
		Text string `xml:",chardata"`
		User string `xml:"user,attr"`
	} `xml:"executedBy"`
	Time struct {
		Text    string `xml:",chardata"`
		Started string `xml:"started,attr"`
		Ended   string `xml:"ended,attr"`
	} `xml:"time"`
	Link struct {
		Text  string `xml:",chardata"`
		Href  string `xml:"href,attr"`
		Rel   string `xml:"rel,attr"`
		Type  string `xml:"type,attr"`
		Title string `xml:"title,attr"`
		Atom  string `xml:"atom,attr"`
	} `xml:"link"`
}

type bodyRun struct {
	XMLName xml.Name `xml:"bodyRun"`
	Text    string   `xml:",chardata"`
	Title   string   `xml:"title,attr"`
	Context string   `xml:"context,attr"`
	Aunit   string   `xml:"aunit,attr"`
	Options struct {
		Text         string `xml:",chardata"`
		Measurements string `xml:"measurements"`
		Scope        struct {
			Text         string `xml:",chardata"`
			OwnTests     string `xml:"ownTests,attr"`
			ForeignTests string `xml:"foreignTests,attr"`
		} `xml:"scope"`
		RiskLevel struct {
			Text      string `xml:",chardata"`
			Harmless  string `xml:"harmless,attr"`
			Dangerous string `xml:"dangerous,attr"`
			Critical  string `xml:"critical,attr"`
		} `xml:"riskLevel"`
		Duration struct {
			Text   string `xml:",chardata"`
			Short  string `xml:"short,attr"`
			Medium string `xml:"medium,attr"`
			Long   string `xml:"long,attr"`
		} `xml:"duration"`
	} `xml:"options"`
	ObjectSet struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
		Osl  string `xml:"osl,attr"`
		Xsi  string `xml:"xsi,attr"`
		Set  struct {
			Text   string `xml:",chardata"`
			Type   string `xml:"type,attr"`
			Object []struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
				Type string `xml:"type,attr"`
			} `xml:"object"`
		} `xml:"set"`
	} `xml:"objectSet"`
}

type objects struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Action string `json:"action"`
}

type commits struct {
	ID string `json:"id"`
}

type objectsResponseBody struct {
	Objects   []objects     `json:"objects"`
	Log       []gctsLogs    `json:"log"`
	Exception gctsException `json:"exception"`
	ErrorLogs []gctsLogs    `json:"errorLog"`
}

type commitsResponseBody struct {
	Commits   []commits     `json:"commits"`
	ErrorLog  []gctsLogs    `json:"errorLog"`
	Log       []gctsLogs    `json:"log"`
	Exception gctsException `json:"exception"`
}

type objectMetaInfo struct {
	Pgmid     string `json:"pgmid"`
	Object    string `json:"object"`
	ObjName   string `json:"objName"`
	Srcsystem string `json:"srcsystem"`
	Author    string `json:"author"`
	Devclass  string `json:"devclass"`
}
type repoObject struct {
	Pgmid       string `json:"pgmid"`
	Object      string `json:"object"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type repoconfig struct {
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	Cprivate   string  `json:"cprivate"`
	Cprotected string  `json:"cprotected"`
	Cvisible   string  `json:"cvisible"`
	Category   string  `json:"category"`
	Scope      string  `json:"scope"`
	ChangedAt  float64 `json:"changeAt"`
	ChangedBy  string  `json:"changedBy"`
}

type result struct {
	Rid           string       `json:"rid"`
	Name          string       `json:"name"`
	Role          string       `json:"role"`
	Type          string       `json:"type"`
	Vsid          string       `json:"vsid"`
	PrivateFlag   string       `json:"privateFlag"`
	Status        string       `json:"status"`
	Branch        string       `json:"branch"`
	Url           string       `json:"url"`
	CreatedBy     string       `json:"createdBy"`
	CreatedDate   string       `json:"createdDate"`
	Config        []repoconfig `json:"config"`
	Objects       int          `json:"objects"`
	CurrentCommit string       `json:"currentCommit"`
}

type objectsResponse struct {
	Objects   []repoObject  `json:"objects"`
	Log       []gctsLogs    `json:"log"`
	Exception gctsException `json:"exception"`
	ErrorLogs []gctsLogs    `json:"errorLog"`
}

type layout struct {
	FormatVersion   int    `json:"formatVersion"`
	Format          string `json:"format"`
	ObjectStorage   string `json:"objectStorage"`
	MetaInformation string `json:"metaInformation"`
	TableContent    string `json:"tableContent"`
	Subdirectory    string `json:"subdirectory"`
	ReadableSource  string `json:"readableSource"`
	KeepClient      string `json:"keepClient"`
}

type repositoryLayout struct {
	Layout    layout     `json:"layout"`
	Log       []gctsLogs `json:"log"`
	Exception string     `json:"exception"`
	ErrorLogs []gctsLogs `json:"errorLog"`
}

type repositoryResponseBody struct {
	Result    result        `json:"result"`
	Exception gctsException `json:"exception"`
}

type Testsuites struct {
	XMLName    xml.Name `xml:"testsuites"`
	Text       string   `xml:",chardata"`
	Title      string   `xml:"title,attr"`
	System     string   `xml:"system,attr"`
	Client     string   `xml:"client,attr"`
	ExecutedBy string   `xml:"executedBy,attr"`
	Time       string   `xml:"time,attr"`
	Timestamp  string   `xml:"timestamp,attr"`
	Failures   string   `xml:"failures,attr"`
	Errors     string   `xml:"errors,attr"`
	Skipped    string   `xml:"skipped,attr"`
	Asserts    string   `xml:"asserts,attr"`
	Tests      string   `xml:"tests,attr"`
	Testsuite  []struct {
		Text      string `xml:",chardata"`
		Name      string `xml:"name,attr"`
		Tests     string `xml:"tests,attr"`
		Failures  string `xml:"failures,attr"`
		Errors    string `xml:"errors,attr"`
		Skipped   string `xml:"skipped,attr"`
		Asserts   string `xml:"asserts,attr"`
		Package   string `xml:"package,attr"`
		Timestamp string `xml:"timestamp,attr"`
		Time      string `xml:"time,attr"`
		Hostname  string `xml:"hostname,attr"`
		Testcase  []struct {
			Text      string `xml:",chardata"`
			Classname string `xml:"classname,attr"`
			Name      string `xml:"name,attr"`
			Time      string `xml:"time,attr"`
			Asserts   string `xml:"asserts,attr"`
			Failure   struct {
				Text    string `xml:",chardata"`
				Message string `xml:"message,attr"`
				Type    string `xml:"type,attr"`
			} `xml:"failure"`
		} `xml:"testcase"`
	} `xml:"testsuite"`
}

type ATCFiles struct {
	XMLName xml.Name `xml:"checkstyle"`
	Files   []File   `xml:"file"`
}

type unitError struct {
	Text     string `xml:",chardata"`
	Message  string `xml:"message,attr"`
	Source   string `xml:"source,attr"`
	Line     string `xml:"line,attr"`
	Severity string `xml:"severity,attr"`
}

type checkstyleFile struct {
	Text  string      `xml:",chardata"`
	Name  string      `xml:"name,attr"`
	Error []unitError `xml:"error"`
}

type Checkstyle struct {
	XMLName xml.Name         `xml:"checkstyle"`
	Text    string           `xml:",chardata"`
	Version string           `xml:"version,attr"`
	File    []checkstyleFile `xml:"file"`
}
