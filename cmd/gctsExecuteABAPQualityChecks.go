package cmd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/gcts"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

var atcFailure, aUnitFailure bool

func gctsExecuteABAPQualityChecks(config gctsExecuteABAPQualityChecksOptions, telemetryData *telemetry.CustomData) {

	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	httpClient := &piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := rungctsExecuteABAPQualityChecks(&config, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	if aUnitFailure || atcFailure {

		log.Entry().Fatal("step execution failed")

	}

}

func rungctsExecuteABAPQualityChecks(config *gctsExecuteABAPQualityChecksOptions, httpClient piperhttp.Sender) error {

	const localChangedObjects = "localchangedobjects"
	const remoteChangedObjects = "remotechangedobjects"
	const localChangedPackages = "localchangedpackages"
	const remoteChangedPackages = "remotechangedpackages"
	const repository = "repository"
	const packages = "packages"

	clientOptions, err := gcts.NewHttpClientOptions(config.Username, config.Password, config.Proxy, config.SkipSSLVerification)
	if err != nil {
		return err
	}
	httpClient.SetOptions(clientOptions)

	log.Entry().Infof("start of gctsExecuteABAPQualityChecks step with configuration values: %v", config)

	var objects []repoObject

	log.Entry().Info("scope:", config.Scope)

	switch strings.ToLower(config.Scope) {
	case localChangedObjects:
		objects, err = getLocalObjects(config, httpClient)
	case remoteChangedObjects:
		objects, err = getRemoteObjects(config, httpClient)
	case localChangedPackages:
		objects, err = getLocalPackages(config, httpClient)
	case remoteChangedPackages:
		objects, err = getRemotePackages(config, httpClient)
	case repository:
		objects, err = getRepositoryObjects(config, httpClient)
	case packages:
		objects, err = getPackages(config, httpClient)
	default:
		log.Entry().Info("the specified scope does not exists, the default one will be used:" + repository)
		objects, err = getRepositoryObjects(config, httpClient)
	}

	if err != nil {
		log.Entry().WithError(err).Fatal("failure in get objects")
	}

	if objects == nil {
		log.Entry().Warning("no object delta was found, therefore the step execution will stop")
		return nil

	}

	log.Entry().Infof("objects to be checked:")
	for _, object := range objects {
		log.Entry().Info(object.Type, " ", object.Object)
	}

	if config.AUnitTest {

		// wrapper for execution of AUnit Test
		err := executeAUnitTest(config, httpClient, objects)

		if err != nil {
			log.Entry().WithError(err)

		}

		if aUnitFailure {

			log.Entry().Error("unit test(s) has/have failed! Check " + config.AUnitResultsFileName + " for more information! If you have enabled Warnings-Next-Generation Plugin, you can see the issues there!")

		} else {

			log.Entry().Info("AUnit test run completed successfully. If there are any results from the run, the results are saved in " + config.AUnitResultsFileName)

		}
	}

	if config.AtcCheck {

		// wrapper for execution of ATCChecks
		err = executeATCCheck(config, httpClient, objects)

		if err != nil {
			log.Entry().WithError(err).Fatal("execute ATC Check failed")
		}

		if atcFailure {

			log.Entry().Error(" ATC issue(s) found! Check " + config.AtcResultsFileName + " for more information! If you have enabled Warnings-Next-Generation Plugin, you can see the issues there!")

		} else {

			log.Entry().Info("ATCCheck test run completed successfully. If there are any results from the run, the results are saved in " + config.AtcResultsFileName)

		}

	}

	return nil

}

func getLocalObjects(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) ([]repoObject, error) {

	var localObjects []repoObject
	var localObject repoObject

	log.Entry().Info("get local changed objects started")

	if config.Commit == "" {

		return []repoObject{}, fmt.Errorf("For scope: localChangedObjects you need to specify a commit")

	}

	history, err := getHistory(config, client)
	if err != nil {
		return []repoObject{}, fmt.Errorf("get local changed objects failed: %w", err)
	}

	if len(history.Result) == 0 {

		return []repoObject{}, fmt.Errorf("no activities (from commit - to commit) were found: %w", err)
	}

	fromCommit := history.Result[0].FromCommit
	log.Entry().Info("from Commit: ", fromCommit)
	toCommit := history.Result[0].ToCommit
	log.Entry().Info("to Commit: ", toCommit)

	// object delta between FromCommit and ToCommit retrieved from Activities Tab in gCTS
	resp, err := getObjectDifference(config, fromCommit, toCommit, client)
	if err != nil {
		return []repoObject{}, fmt.Errorf("get local changed objects failed: %w", err)
	}

	for _, object := range resp.Objects {
		localObject.Object = object.Name
		localObject.Type = object.Type
		localObjects = append(localObjects, localObject)
	}

	log.Entry().Info("get local changed objects finished")

	return localObjects, nil
}

func getRemoteObjects(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) ([]repoObject, error) {

	var remoteObjects []repoObject
	var remoteObject repoObject
	var currentRemoteCommit string

	log.Entry().Info("get remote changed objects started")

	if config.Commit == "" {

		return []repoObject{}, fmt.Errorf("For scope: remoteChangedObjects you need to specify a commit")

	}

	commitList, err := getCommitList(config, client)

	if err != nil {
		return []repoObject{}, fmt.Errorf("get remote changed objects failed: %w", err)
	}

	for i, commit := range commitList.Commits {
		if commit.ID == config.Commit {
			currentRemoteCommit = commitList.Commits[i+1].ID
			break
		}
	}
	if currentRemoteCommit == "" {
		return []repoObject{}, errors.New("current remote commit was not found")

	}
	log.Entry().Info("current commit in the remote repository: ", currentRemoteCommit)
	// object delta between the commit that triggered the pipeline and the current commit in the remote repository
	resp, err := getObjectDifference(config, currentRemoteCommit, config.Commit, client)

	if err != nil {
		return []repoObject{}, fmt.Errorf("get remote changed objects failed: %w", err)
	}

	for _, object := range resp.Objects {
		remoteObject.Object = object.Name
		remoteObject.Type = object.Type
		remoteObjects = append(remoteObjects, remoteObject)
	}

	log.Entry().Info("get remote changed objects finished")

	return remoteObjects, nil
}

func getLocalPackages(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) ([]repoObject, error) {

	var localPackages []repoObject
	var localPackage repoObject

	log.Entry().Info("get local changed packages started")

	if config.Commit == "" {

		return []repoObject{}, fmt.Errorf("For scope: localChangedPackages you need to specify a commit")

	}

	history, err := getHistory(config, client)
	if err != nil {
		return []repoObject{}, fmt.Errorf("get local changed objects failed: %w", err)
	}

	if len(history.Result) == 0 {

		return []repoObject{}, fmt.Errorf("no activities (from commit - to commit) were found: %w", err)
	}

	fromCommit := history.Result[0].FromCommit
	log.Entry().Info("from Commit: ", fromCommit)
	toCommit := history.Result[0].ToCommit
	log.Entry().Info("to Commit: ", toCommit)

	// object delta between FromCommit and ToCommit retrieved from Activities Tab in gCTS
	resp, err := getObjectDifference(config, fromCommit, config.Commit, client)

	if err != nil {
		return []repoObject{}, fmt.Errorf("get local changed packages failed: %w", err)

	}

	myPackages := map[string]bool{}

	// objects are resolved into packages(DEVC)
	for _, object := range resp.Objects {
		objInfo, err := getObjectInfo(config, client, object.Name, object.Type)
		if err != nil {
			return []repoObject{}, fmt.Errorf("get local changed packages failed: %w", err)
		}
		if myPackages[objInfo.Devclass] {

		} else {
			myPackages[objInfo.Devclass] = true
			localPackage.Object = objInfo.Devclass
			localPackage.Type = "DEVC"
			localPackages = append(localPackages, localPackage)
		}

	}

	log.Entry().Info("get local changed packages finished")
	return localPackages, nil
}

func getRemotePackages(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) ([]repoObject, error) {

	var remotePackages []repoObject
	var remotePackage repoObject
	var currentRemoteCommit string

	log.Entry().Info("get remote changed packages started")

	if config.Commit == "" {

		return []repoObject{}, fmt.Errorf("For scope: remoteChangedPackages you need to specify a commit")

	}

	commitList, err := getCommitList(config, client)

	if err != nil {
		return []repoObject{}, fmt.Errorf("get remote changed packages failed: %w", err)
	}

	for i, commit := range commitList.Commits {
		if commit.ID == config.Commit {
			currentRemoteCommit = commitList.Commits[i+1].ID
			break
		}
	}

	if currentRemoteCommit == "" {
		return []repoObject{}, fmt.Errorf("current remote commit was not found: %w", err)

	}
	log.Entry().Info("current commit in the remote repository: ", currentRemoteCommit)
	//object delta between the commit that triggered the pipeline and the current commit in the remote repository
	resp, err := getObjectDifference(config, currentRemoteCommit, config.Commit, client)
	if err != nil {
		return []repoObject{}, fmt.Errorf("get remote changed packages failed: %w", err)
	}

	myPackages := map[string]bool{}
	// objects are resolved into packages(DEVC)
	for _, object := range resp.Objects {
		objInfo, err := getObjectInfo(config, client, object.Name, object.Type)
		if err != nil {
			return []repoObject{}, fmt.Errorf("get remote changed packages failed: %w", err)
		}
		if myPackages[objInfo.Devclass] {

		} else {
			myPackages[objInfo.Devclass] = true
			remotePackage.Object = objInfo.Devclass
			remotePackage.Type = "DEVC"
			remotePackages = append(remotePackages, remotePackage)
		}

	}
	log.Entry().Info("get remote changed packages finished")
	return remotePackages, nil
}

func getRepositoryObjects(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) ([]repoObject, error) {

	log.Entry().Info("get repository objects started")

	var repoResp repoObjectResponse

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/objects?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return nil, urlErr
	}

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return []repoObject{}, fmt.Errorf("could not get repository objects: %w", httpErr)
	} else if resp == nil {
		return []repoObject{}, errors.New("could not get repository objects: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &repoResp)
	if parsingErr != nil {
		return []repoObject{}, fmt.Errorf("%v", parsingErr)
	}

	var repositoryObjects []repoObject

	// remove object type DEVC, because it is already included in scope packages
	// also if you run ATC Checks for DEVC together with other object types, ATC checks will run only for DEVC
	for _, object := range repoResp.Objects {

		if object.Type != "DEVC" {
			repositoryObjects = append(repositoryObjects, object)
		}

	}

	log.Entry().Info("get repository objects finished")

	// all objects that are part of the local repository
	return repositoryObjects, nil
}

func getPackages(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) ([]repoObject, error) {

	var packages []repoObject

	log.Entry().Info("get packages started")

	var repoResp repoObjectResponse

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/objects?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return nil, urlErr
	}

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return []repoObject{}, fmt.Errorf("get packages failed: could not get repository objects: %w", httpErr)
	} else if resp == nil {
		return []repoObject{}, errors.New("get packages failed: could not get repository objects: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &repoResp)
	if parsingErr != nil {
		return []repoObject{}, fmt.Errorf("%v", parsingErr)
	}
	// chose only DEVC from repository objects
	for _, object := range repoResp.Objects {

		if object.Type == "DEVC" {
			packages = append(packages, object)
		}

	}

	log.Entry().Info("get packages finished")
	return packages, nil
}

func discoverServer(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) (*http.Header, error) {

	url := config.Host +
		"/sap/bc/adt/core/discovery?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return nil, urlErr
	}

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
		return nil, fmt.Errorf("discovery of the ABAP server failed: %w", httpErr)
	} else if disc == nil || disc.Header == nil {
		return nil, errors.New("discovery of the ABAP server failed: did not retrieve a HTTP response")
	}

	return &disc.Header, nil
}

func executeAUnitTest(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, objects []repoObject) error {

	log.Entry().Info("execute ABAP Unit Test started")

	var innerXml string
	var result runResult

	for _, object := range objects {

		switch object.Type {
		case "CLAS":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/oo/classes/` + url.QueryEscape(object.Object) + `"/>`
		case "DEVC":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/repository/informationsystem/virtualfolders?selection=package%3a` + url.QueryEscape(object.Object) + `"/>`

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
		return fmt.Errorf("execute of Aunit test has failed: %w", err)
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyXML(resp, &result)
	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
		return nil
	}

	parsedRes, err := parseUnitResult(config, client, &result)

	if err != nil {
		log.Entry().Warning(err)
		return nil
	}

	log.Entry().Info("execute ABAP Unit Test finished.", parsedRes.Text)

	return nil
}

func runAUnitTest(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, xml []byte) (response *http.Response, err error) {

	log.Entry().Info("run ABAP Unit Test started")
	url := config.Host +
		"/sap/bc/adt/abapunit/testruns?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return nil, urlErr
	}

	discHeader, discError := discoverServer(config, client)

	if discError != nil {
		return response, fmt.Errorf("run of unit tests failed: %w", discError)
	}

	if discHeader.Get("X-Csrf-Token") == "" {

		return response, fmt.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	header.Add("Accept", "application/xml")
	header.Add("Content-Type", "application/vnd.sap.adt.abapunit.testruns.result.v1+xml")

	response, httpErr := client.SendRequest("POST", url, bytes.NewBuffer(xml), header, nil)

	if httpErr != nil {
		return response, fmt.Errorf("run of unit tests failed: %w", httpErr)
	} else if response == nil {
		return response, errors.New("run of unit tests failed: did not retrieve a HTTP response")
	}

	log.Entry().Info("run ABAP Unit Test finished")
	return response, nil
}

func parseUnitResult(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, aUnitRunResult *runResult) (parsedResult checkstyle, err error) {

	log.Entry().Info("parse ABAP Unit Result started")

	var fileName string
	var aUnitFile file
	var aUnitError checkstyleError

	parsedResult.Version = "1.0"

	for _, program := range aUnitRunResult.Program {

		objectType := program.Type[0:4]
		objectName := program.Name

		//syntax error in unit test or class
		if program.Alerts.Alert.HasSyntaxErrors == "true" {

			aUnitFailure = true
			aUnitError.Source = objectName
			aUnitError.Severity = "error"
			log.Entry().Info("severity: ", aUnitError.Severity)
			aUnitError.Message = html.UnescapeString(program.Alerts.Alert.Title + " " + program.Alerts.Alert.Details.Detail.AttrText)
			log.Entry().Info("message: ", aUnitError.Message)
			aUnitError.Line, err = findLine(config, client, program.Alerts.Alert.Stack.StackEntry.URI, objectName, objectType)
			log.Entry().Error("line: ", aUnitError.Line)
			if err != nil {
				return parsedResult, fmt.Errorf("parse AUnit Result failed: %w", err)

			}
			fileName, err = getFileName(config, client, program.Alerts.Alert.Stack.StackEntry.URI, objectName)
			log.Entry().Error("file path: ", aUnitError.Line)
			if err != nil {
				return parsedResult, fmt.Errorf("parse AUnit Result failed: %w", err)

			}

			aUnitFile.Error = append(aUnitFile.Error, aUnitError)
			aUnitError = checkstyleError{}
			log.Entry().Error("there is a syntax error", aUnitFile)
		}

		for _, testClass := range program.TestClasses.TestClass {

			for _, testMethod := range testClass.TestMethods.TestMethod {

				aUnitError.Source = testClass.Name + "/" + testMethod.Name

				// unit test failure
				if len(testMethod.Alerts.Alert) > 0 {

					for _, testalert := range testMethod.Alerts.Alert {

						switch testalert.Severity {
						case "fatal":
							log.Entry().Error("unit test " + aUnitError.Source + " has failed with severity fatal")
							aUnitFailure = true
							aUnitError.Severity = "error"
						case "critical":
							log.Entry().Error("unit test " + aUnitError.Source + " has failed with severity critical")
							aUnitFailure = true
							aUnitError.Severity = "error"
						case "tolerable":
							log.Entry().Warning("unit test " + aUnitError.Source + "  has failed with severity warning")
							aUnitError.Severity = "warning"
						default:
							aUnitError.Severity = "info"

						}

						//unit test message is spread in different elements
						for _, detail := range testalert.Details.Detail {
							aUnitError.Message = aUnitError.Message + " " + detail.AttrText
							for _, subdetail := range detail.Details.Detail {

								aUnitError.Message = html.UnescapeString(aUnitError.Message + " " + subdetail.AttrText)
								log.Entry().Info("message: ", aUnitError.Message)
							}

						}

						aUnitError.Line, err = findLine(config, client, testalert.Stack.StackEntry.URI, objectName, objectType)
						log.Entry().Info("line: ", aUnitError.Line)
						if err != nil {

							log.Entry().Warning(err)

						}

					}

					aUnitFile.Error = append(aUnitFile.Error, aUnitError)
					aUnitError = checkstyleError{}

				} else {

					log.Entry().Info("unit test:", aUnitError.Source, "- was successful")

				}

			}

			fileName, err = getFileName(config, client, testClass.URI, objectName)
			if err != nil {
				return parsedResult, fmt.Errorf("parse AUnit Result failed: %w", err)

			}
		}

		aUnitFile.Name, err = constructPath(config, client, fileName, objectName, objectType)
		log.Entry().Error("file path: ", aUnitFile.Name)
		if err != nil {

			return parsedResult, fmt.Errorf("parse AUnit Result failed: %w", err)
		}
		parsedResult.File = append(parsedResult.File, aUnitFile)
		aUnitFile = file{}

	}

	body, _ := xml.Marshal(parsedResult)

	writeErr := os.WriteFile(config.AUnitResultsFileName, body, 0644)

	if writeErr != nil {
		log.Entry().Error("file AUnitResults.xml could not be created")
		return parsedResult, fmt.Errorf("handling unit test results failed: %w", writeErr)
	}

	log.Entry().Info("parse ABAP Unit Result finished")
	return parsedResult, nil

}

func executeATCCheck(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, objects []repoObject) (error error) {

	log.Entry().Info("execute ATC Check started")

	var innerXml string
	var result worklist

	for _, object := range objects {

		switch object.Type {

		case "CLAS":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/oo/classes/` + url.QueryEscape(object.Object) + `"/>`
		case "INTF":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/oo/interfaces/` + object.Object + `"/>`
		case "DEVC":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/repository/informationsystem/virtualfolders?selection=package%3a` + url.QueryEscape(object.Object) + `"/>`
		case "FUGR":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/functions/groups/` + object.Object + `/source/main"/>`
		case "TABL":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/ddic/tables/` + object.Object + `/source/main"/>`
		case "DTEL":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/ddic/dataelements/` + object.Object + `"/>`
		case "DOMA":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/ddic/domains/` + object.Object + `"/>`
		case "MSAG":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/messageclass/` + object.Object + `"/>`
		case "PROG":
			innerXml = innerXml + `<adtcore:objectReference adtcore:uri="/sap/bc/adt/programs/programs/` + object.Object + `/source/main"/>`
		default:
			log.Entry().Warning("object Type " + object.Type + " is not supported!")

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
		return fmt.Errorf("execution of ATC Checks failed: %w", err)
	}

	err = startATCRun(config, client, xmlBody, worklist)

	if err != nil {
		return fmt.Errorf("execution of ATC Checks failed: %w", err)
	}

	resp, err := getATCRun(config, client, worklist)

	if err != nil {
		return fmt.Errorf("execution of ATC Checks failed: %w", err)
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyXML(resp, &result)
	if parsingErr != nil {
		log.Entry().Warning(parsingErr)
		return nil
	}

	atcRes, err := parseATCCheckResult(config, client, &result)

	if err != nil {
		log.Entry().Error(err)
		return fmt.Errorf("execution of ATC Checks failed: %w", err)
	}

	log.Entry().Info("execute ATC Checks finished.", atcRes.Text)

	return nil

}
func startATCRun(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, xml []byte, worklistID string) (err error) {

	log.Entry().Info("ATC Run started")

	discHeader, discError := discoverServer(config, client)
	if discError != nil {
		return fmt.Errorf("start of ATC run failed: %w", discError)
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return fmt.Errorf("could not retrieve x-csrf-token from server")
	}

	header := make(http.Header)
	header.Add("x-csrf-token", discHeader.Get("X-Csrf-Token"))
	header.Add("Accept", "application/xml")

	url := config.Host +
		"/sap/bc/adt/atc/runs?worklistId=" + worklistID + "&sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return urlErr
	}

	resp, httpErr := client.SendRequest("POST", url, bytes.NewBuffer(xml), header, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return fmt.Errorf("start of ATC run failed: %w", httpErr)
	} else if resp == nil {
		return errors.New("start of ATC run failed: did not retrieve a HTTP response")
	}

	log.Entry().Info("ATC Run finished")

	return nil

}

func getATCRun(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, worklistID string) (response *http.Response, err error) {

	log.Entry().Info("get ATC Run Results started")

	header := make(http.Header)

	url := config.Host +
		"/sap/bc/adt/atc/worklists/" + worklistID + "?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return nil, urlErr
	}

	header.Add("Accept", "application/atc.worklist.v1+xml")

	resp, httpErr := client.SendRequest("GET", url, nil, header, nil)

	if httpErr != nil {
		return response, fmt.Errorf("get ATC run failed: %w", httpErr)
	} else if resp == nil {
		return response, errors.New("get ATC run failed: did not retrieve a HTTP response")
	}
	log.Entry().Info("get ATC Run Results finished")
	return resp, nil

}

func getWorklist(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) (worklistID string, error error) {

	url := config.Host +
		"/sap/bc/adt/atc/worklists?checkVariant=" + config.AtcVariant + "&sap-client=" + config.Client
	discHeader, discError := discoverServer(config, client)

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return worklistID, urlErr
	}

	if discError != nil {
		return worklistID, fmt.Errorf("get worklist failed: %w", discError)
	}

	if discHeader.Get("X-Csrf-Token") == "" {
		return worklistID, fmt.Errorf("could not retrieve x-csrf-token from server")
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
		return worklistID, fmt.Errorf("get worklist failed: %w", httpErr)
	} else if resp == nil {
		return worklistID, errors.New("get worklist failed: did not retrieve a HTTP response")
	}
	location := resp.Header["Location"][0]
	locationSlice := strings.Split(location, "/")
	worklistID = locationSlice[len(locationSlice)-1]
	log.Entry().Info("worklist id for ATC check: ", worklistID)

	return worklistID, nil
}

func parseATCCheckResult(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, response *worklist) (atcResults checkstyle, error error) {

	log.Entry().Info("parse ATC Check Result started")

	var atcFile file
	var subObject string
	var aTCUnitError checkstyleError

	atcResults.Version = "1.0"

	for _, object := range response.Objects.Object {

		objectType := object.Type
		objectName := object.Name

		for _, atcworklist := range object.Findings.Finding {

			log.Entry().Info("there is atc finding for object type: ", objectType+" object name: "+objectName)

			path, err := url.PathUnescape(atcworklist.Location)

			if err != nil {
				return atcResults, fmt.Errorf("conversion of ATC check results to CheckStyle has failed: %w", err)

			}

			if len(atcworklist.Atcfinding) > 0 {

				priority, err := strconv.Atoi(atcworklist.Priority)

				if err != nil {
					return atcResults, fmt.Errorf("conversion of ATC check results to CheckStyle has failed: %w", err)

				}

				switch priority {
				case 1:
					atcFailure = true
					aTCUnitError.Severity = "error"
					log.Entry().Error("atc issue with priority: 1 ")
				case 2:
					atcFailure = true
					aTCUnitError.Severity = "error"
					log.Entry().Error("atc issue with priority: 2 ")
				case 3:
					aTCUnitError.Severity = "warning"
					log.Entry().Warning("atc issue with priority: 3 ")
				default:
					aTCUnitError.Severity = "info"
					log.Entry().Info("atc issue with low priority ")
				}

				log.Entry().Error("severity: ", aTCUnitError.Severity)

				if aTCUnitError.Line == "" {

					aTCUnitError.Line, err = findLine(config, client, path, objectName, objectType)
					log.Entry().Info("line: ", aTCUnitError.Line)

					if err != nil {
						log.Entry().Info(path)
						log.Entry().Warning(err)

					}

				}

				if subObject != "" {
					aTCUnitError.Source = objectName + "/" + strings.ToUpper(subObject)
				} else {
					aTCUnitError.Source = objectName
				}

				aTCUnitError.Message = html.UnescapeString(atcworklist.CheckTitle + " " + atcworklist.MessageTitle)
				log.Entry().Info("message: ", aTCUnitError.Message)
				atcFile.Error = append(atcFile.Error, aTCUnitError)
				aTCUnitError = checkstyleError{}
			}

			if atcFile.Error[0].Message != "" {

				fileName, err := getFileName(config, client, path, objectName)

				if err != nil {
					return atcResults, fmt.Errorf("conversion of ATC check results to CheckStyle has failed: %w", err)
				}

				atcFile.Name, err = constructPath(config, client, fileName, objectName, objectType)
				log.Entry().Info("file path: ", atcFile.Name)
				if err != nil {
					return atcResults, fmt.Errorf("conversion of ATC check results to CheckStyle has failed: %w", err)
				}
				atcResults.File = append(atcResults.File, atcFile)
				atcFile = file{}

			}

		}
	}

	atcBody, _ := xml.Marshal(atcResults)

	writeErr := os.WriteFile(config.AtcResultsFileName, atcBody, 0644)

	if writeErr != nil {
		log.Entry().Error("ATCResults.xml could not be created")
		return atcResults, fmt.Errorf("handling atc results failed: %w", writeErr)
	}
	log.Entry().Info("parsing ATC check results to CheckStyle has finished.")
	return atcResults, writeErr
}

func constructPath(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, fileName string, objectName string, objectType string) (filePath string, error error) {

	targetDir, err := getTargetDir(config, client)
	if err != nil {
		return filePath, fmt.Errorf("path could not be constructed: %w", err)

	}

	filePath = config.Workspace + "/" + targetDir + "/objects/" + strings.ToUpper(objectType) + "/" + strings.ToUpper(objectName) + "/" + fileName
	return filePath, nil

}

func findLine(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, path string, objectName string, objectType string) (line string, error error) {

	regexLine := regexp.MustCompile(`.start=\d*`)
	regexMethod := regexp.MustCompile(`.name=[a-zA-Z0-9_-]*;`)

	readableSource, err := checkReadableSource(config, client)

	if err != nil {

		return line, fmt.Errorf("could not find line in source code: %w", err)

	}

	fileName, err := getFileName(config, client, path, objectName)

	if err != nil {

		return line, err

	}

	filePath, err := constructPath(config, client, fileName, objectName, objectType)
	if err != nil {
		return line, fmt.Errorf(objectType+"/"+objectName+"could not find line in source code", err)

	}

	var absLine int
	if readableSource {

		// the error line that we get from UnitTest Run or ATC Check is not aligned for the readable source, we need to calculated it
		rawfile, err := os.ReadFile(filePath)

		if err != nil {

			return line, fmt.Errorf("could not find object in the workspace of your CI/CD tool : %w", err)
		}

		file := string(rawfile)

		splittedfile := strings.Split(file, "\n")

		// CLAS/OSO - is unique identifier for protection section in CLAS
		if strings.Contains(path, "CLAS/OSO") {

			for l, line := range splittedfile {

				if strings.Contains(line, "protected section.") {
					absLine = l
					break
				}

			}

			// CLAS/OM - is unique identifier for method section in CLAS
		} else if strings.Contains(path, "CLAS/OM") {

			methodName := regexMethod.FindString(path)

			if methodName != "" {
				methodName = methodName[len(`.name=`) : len(methodName)-1]

			}

			for line, linecontent := range splittedfile {

				if strings.Contains(linecontent, "method"+" "+methodName) {
					absLine = line
					break
				}

			}

			// CLAS/OSI - is unique identifier for private section in CLAS
		} else if strings.Contains(path, "CLAS/OSI") {

			for line, linecontent := range splittedfile {

				if strings.Contains(linecontent, "private section.") {
					absLine = line
					break
				}

			}

		}

		errLine := regexLine.FindString(path)

		if errLine != "" {

			errLine, err := strconv.Atoi(errLine[len(`.start=`):])
			if err == nil {
				line = strconv.Itoa(absLine + errLine)

			}

		}

	} else {
		// classic format
		errLine := regexLine.FindString(path)
		if errLine != "" {
			line = errLine[len(`.start=`):]

		}

	}

	return line, nil
}
func getFileName(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, path string, objName string) (fileName string, error error) {

	readableSource, err := checkReadableSource(config, client)
	if err != nil {
		return fileName, fmt.Errorf("get file name has failed: %w", err)

	}

	path, err = url.PathUnescape(path)

	var fileExtension string
	fileExtensionLength := 30 - len(objName)
	for range fileExtensionLength {
		fileExtension += "="
	}

	if err != nil {
		return fileName, fmt.Errorf("get file name has failed: %w", err)

	}

	//  INTERFACES
	regexInterface := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/interfaces\/\w*`)
	intf := regexInterface.FindString(path)
	if intf != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".intf.abap"
		} else {
			fileName = "REPS " + strings.ToUpper(objName) + fileExtension + "IU.abap"
		}

	}
	// CLASSES DEFINITIONS
	regexClasDef := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/includes\/definitions\/`)
	clasDef := regexClasDef.FindString(path)
	if clasDef != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.definitions.abap"
		} else {
			fileName = "CINC " + objName + fileExtension + "CCDEF.abap"
		}

	}

	// CLASSES IMPLEMENTATIONS
	regexClasImpl := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/includes\/implementations\/`)
	clasImpl := regexClasImpl.FindString(path)
	if clasImpl != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.implementations.abap"
		} else {
			fileName = "CINC " + objName + fileExtension + "CCIMP.abap"
		}

	}

	// CLASSES MACROS
	regexClasMacro := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/includes\/macros\/`)
	clasMacro := regexClasMacro.FindString(path)
	if clasMacro != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.macros.abap"
		} else {
			fileName = "CINC " + objName + fileExtension + "CCMAC.abap"
		}

	}

	// TEST CLASSES
	regexTestClass := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*#?\/?\w*\/?testclass`)
	testClass := regexTestClass.FindString(path)
	if testClass != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.testclasses.abap"
		} else {
			fileName = "CINC " + objName + fileExtension + "CCAU.abap"
		}

	}

	// CLASS PROTECTED
	regexClasProtected := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/source\/main#type=CLAS\/OSO`)
	classProtected := regexClasProtected.FindString(path)
	if classProtected != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.abap"
		} else {
			fileName = "CPRO " + objName + ".abap"
		}

	}

	// CLASS PRIVATE
	regexClasPrivate := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/source\/main#type=CLAS\/OSI`)
	classPrivate := regexClasPrivate.FindString(path)
	if classPrivate != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.abap"
		} else {
			fileName = "CPRI " + objName + ".abap"
		}

	}

	// CLASS METHOD
	regexClasMethod := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/source\/main#type=CLAS\/OM`)
	classMethod := regexClasMethod.FindString(path)
	if classMethod != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.abap"
		} else {

			regexmethodName := regexp.MustCompile(`name=\w*`)
			methodName := regexmethodName.FindString(path)

			fileName = "METH " + methodName[len(`name=`):] + ".abap"
		}

	}

	// CLASS PUBLIC
	regexClasPublic := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/\w*\/source\/main#start`)
	classPublic := regexClasPublic.FindString(path)
	if classPublic != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.abap"
		} else {
			fileName = "CPUB " + objName + ".abap"
		}

	}

	// FUNCTION INCLUDE
	regexFuncIncl := regexp.MustCompile(`\/sap\/bc\/adt\/functions\/groups\/\w*\/includes/\w*`)

	funcIncl := regexFuncIncl.FindString(path)
	if funcIncl != "" && fileName == "" {

		regexSubObj := regexp.MustCompile(`includes\/\w*`)
		subObject := regexSubObj.FindString(path)
		subObject = subObject[len(`includes/`):]

		if readableSource {

			fileName = strings.ToLower(objName) + ".fugr." + strings.ToLower(subObject) + ".reps.abap"
		} else {
			fileName = "REPS " + strings.ToUpper(subObject) + ".abap"
		}

	}

	// FUNCTION GROUP
	regexFuncGr := regexp.MustCompile(`\/sap\/bc\/adt\/functions\/groups\/\w*\/source\/main`)

	funcGr := regexFuncGr.FindString(path)
	if funcGr != "" && fileName == "" {

		if readableSource {

			fileName = strings.ToLower(objName) + ".fugr.sapl" + strings.ToLower(objName) + ".reps.abap"
		} else {
			fileName = "REPS SAPL" + objName + ".abap"
		}

	}

	// FUNCTION MODULE
	regexFuncMod := regexp.MustCompile(`\/sap\/bc\/adt\/functions\/groups\/\w*\/fmodules/\w*`)
	funcMod := regexFuncMod.FindString(path)
	if funcMod != "" && fileName == "" {

		regexSubObj := regexp.MustCompile(`includes\/\w*`)
		subObject := regexSubObj.FindString(path)
		subObject = subObject[len(`includes/`):]

		if readableSource {

			fileName = strings.ToLower(subObject) + ".func.abap"
		} else {
			fileName = "FUNC " + subObject + ".abap"
		}

	}
	// CLAS
	regexClas := regexp.MustCompile(`\/sap\/bc\/adt\/oo\/classes\/` + strings.ToLower(objName))
	clas := regexClas.FindString(path)
	if clas != "" && fileName == "" {
		if readableSource {

			fileName = strings.ToLower(objName) + ".clas.abap"
		} else {

			fileName = "CPUB " + objName + ".abap"
		}

	}

	// PROGRAM
	regexProg := regexp.MustCompile(`\/sap\/bc\/adt\/programs\/programs\/` + strings.ToLower(objName))
	prog := regexProg.FindString(path)
	if prog != "" && fileName == "" {

		fileName = "REPS " + objName + ".abap"

	}

	// TABLES
	regexTab := regexp.MustCompile(`\/sap\/bc\/adt\/ddic\/tables\/` + strings.ToLower(objName))
	tab := regexTab.FindString(path)
	if tab != "" && fileName == "" {

		fileName = "TABL " + objName + ".asx.json"

	}

	return fileName, nil

}

func getTargetDir(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) (string, error) {

	var targetDir string

	repository, err := getRepo(config, client)

	if err != nil {
		return targetDir, err
	}

	for _, config := range repository.Result.Config {
		if config.Key == "VCS_TARGET_DIR" {
			targetDir = config.Value
		}
	}

	return targetDir, nil

}

func checkReadableSource(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) (readableSource bool, error error) {

	repoLayout, err := getRepositoryLayout(config, client)
	if err != nil {
		return readableSource, fmt.Errorf("could not check readable source format: %w", err)
	}

	if repoLayout.Layout.ReadableSource == "true" || repoLayout.Layout.ReadableSource == "only" || repoLayout.Layout.ReadableSource == "all" {

		readableSource = true

	} else {

		readableSource = false

	}

	return readableSource, nil
}

func getRepo(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) (repositoryResponse, error) {

	var repositoryResp repositoryResponse
	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return repositoryResp, urlErr
	}

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return repositoryResponse{}, fmt.Errorf("could not get repository: %w", httpErr)
	} else if resp == nil {
		return repositoryResponse{}, errors.New("could not get repository: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &repositoryResp)
	if parsingErr != nil {
		return repositoryResponse{}, fmt.Errorf("%v", parsingErr)
	}

	return repositoryResp, nil

}

func getRepositoryLayout(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) (layoutResponse, error) {

	var repoLayoutResponse layoutResponse
	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/layout?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return repoLayoutResponse, urlErr
	}

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return layoutResponse{}, fmt.Errorf("could not get repository layout: %w", httpErr)
	} else if resp == nil {
		return layoutResponse{}, errors.New("could not get repository layout: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &repoLayoutResponse)
	if parsingErr != nil {
		return layoutResponse{}, fmt.Errorf("%v", parsingErr)
	}

	return repoLayoutResponse, nil
}

func getCommitList(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) (commitResponse, error) {

	var commitResp commitResponse
	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/getCommit?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)
	if urlErr != nil {

		return commitResp, urlErr
	}

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return commitResponse{}, fmt.Errorf("get repository history failed: %w", httpErr)
	} else if resp == nil {
		return commitResponse{}, errors.New("get repository history failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &commitResp)
	if parsingErr != nil {
		return commitResponse{}, fmt.Errorf("%v", parsingErr)
	}

	return commitResp, nil
}

func getObjectDifference(config *gctsExecuteABAPQualityChecksOptions, fromCommit string, toCommit string, client piperhttp.Sender) (objectsResponse, error) {
	var objectResponse objectsResponse

	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository +
		"/compareCommits?fromCommit=" + fromCommit + "&toCommit=" + toCommit + "&sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return objectResponse, urlErr
	}

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return objectsResponse{}, fmt.Errorf("get object difference failed: %w", httpErr)
	} else if resp == nil {
		return objectsResponse{}, errors.New("get object difference failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &objectResponse)
	if parsingErr != nil {
		return objectsResponse{}, fmt.Errorf("%v", parsingErr)
	}
	log.Entry().Info("get object differences: ", objectResponse.Objects)
	return objectResponse, nil
}

func getObjectInfo(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender, objectName string, objectType string) (objectInfo, error) {

	var objectMetInfoResponse objectInfo
	url := config.Host +
		"/sap/bc/cts_abapvcs/objects/" + objectType + "/" + objectName +
		"?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return objectMetInfoResponse, urlErr
	}

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if httpErr != nil {
		return objectInfo{}, fmt.Errorf("resolve package failed: %w", httpErr)
	} else if resp == nil {
		return objectInfo{}, errors.New("resolve package failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &objectMetInfoResponse)
	if parsingErr != nil {
		return objectInfo{}, fmt.Errorf("%v", parsingErr)
	}
	return objectMetInfoResponse, nil

}

func getHistory(config *gctsExecuteABAPQualityChecksOptions, client piperhttp.Sender) (historyResponse, error) {

	var historyResp historyResponse
	url := config.Host +
		"/sap/bc/cts_abapvcs/repository/" + config.Repository + "/getHistory?sap-client=" + config.Client

	url, urlErr := addQueryToURL(url, config.QueryParameters)

	if urlErr != nil {

		return historyResp, urlErr
	}

	resp, httpErr := client.SendRequest("GET", url, nil, nil, nil)

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if httpErr != nil {
		return historyResponse{}, fmt.Errorf("get history failed: %w", httpErr)
	} else if resp == nil {
		return historyResponse{}, errors.New("get history failed: did not retrieve a HTTP response")
	}

	parsingErr := piperhttp.ParseHTTPResponseBodyJSON(resp, &historyResp)
	if parsingErr != nil {
		return historyResponse{}, fmt.Errorf("%v", parsingErr)
	}

	return historyResp, nil
}

type worklist struct {
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

type commit struct {
	ID string `json:"id"`
}

type commitResponse struct {
	Commits   []commit      `json:"commits"`
	ErrorLog  []gctsLogs    `json:"errorLog"`
	Log       []gctsLogs    `json:"log"`
	Exception gctsException `json:"exception"`
}

type objectInfo struct {
	Pgmid     string `json:"pgmid"`
	Object    string `json:"object"`
	ObjName   string `json:"objName"`
	Srcsystem string `json:"srcsystem"`
	Author    string `json:"author"`
	Devclass  string `json:"devclass"`
}

type repoConfig struct {
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

type repository struct {
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
	Config        []repoConfig `json:"config"`
	Objects       any          `json:"objects"`
	CurrentCommit string       `json:"currentCommit"`
}

type repositoryResponse struct {
	Result    repository    `json:"result"`
	Exception gctsException `json:"exception"`
}

type objects struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Action string `json:"action"`
}
type objectsResponse struct {
	Objects   []objects     `json:"objects"`
	Log       []gctsLogs    `json:"log"`
	Exception gctsException `json:"exception"`
	ErrorLogs []gctsLogs    `json:"errorLog"`
}

type repoObject struct {
	Pgmid       string `json:"pgmid"`
	Object      string `json:"object"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type repoObjectResponse struct {
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

type layoutResponse struct {
	Layout    layout     `json:"layout"`
	Log       []gctsLogs `json:"log"`
	Exception string     `json:"exception"`
	ErrorLogs []gctsLogs `json:"errorLog"`
}

type history struct {
	Rid          string `json:"rid"`
	CheckoutTime int    `json:"checkoutTime"`
	FromCommit   string `json:"fromCommit"`
	ToCommit     string `json:"toCommit"`
	Caller       string `json:"caller"`
	Type         string `json:"type"`
}

type historyResponse struct {
	Result    []history `xml:"result"`
	Exception string    `json:"exception"`
}

type checkstyleError struct {
	Text     string `xml:",chardata"`
	Message  string `xml:"message,attr"`
	Source   string `xml:"source,attr"`
	Line     string `xml:"line,attr"`
	Severity string `xml:"severity,attr"`
}

type file struct {
	Text  string            `xml:",chardata"`
	Name  string            `xml:"name,attr"`
	Error []checkstyleError `xml:"error"`
}

type checkstyle struct {
	XMLName xml.Name `xml:"checkstyle"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	File    []file   `xml:"file"`
}
