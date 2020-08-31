package cmd

import (
	"bytes"
	"path"
	"path/filepath"
	"sort"
	"time"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentAssemblePackages(config abapEnvironmentAssemblePackagesOptions, telemetryData *telemetry.CustomData, cpe *abapEnvironmentAssemblePackagesCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}
	err := runAbapEnvironmentAssemblePackages(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

// *******************************************************************************************************************************
// ********************************************************** Step logic *********************************************************
// *******************************************************************************************************************************
func runAbapEnvironmentAssemblePackages(config *abapEnvironmentAssemblePackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender, cpe *abapEnvironmentAssemblePackagesCommonPipelineEnvironment) error {
	conn := new(connector)
	err := conn.init(config, com, client)
	if err != nil {
		return err
	}
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)
	builds, buildsAlreadyReleased, err := starting(addonDescriptor.Repositories, *conn)
	if err != nil {
		return err
	}
	err = polling(builds, time.Duration(config.MaxRuntimeInMinutes), 60)
	if err != nil {
		return err
	}
	err = checkIfFailedAndPrintLogs(builds)
	if err != nil {
		return err
	}
	reposBackToCPE, err := downloadSARXML(builds)
	if err != nil {
		return err
	}
	// also write the already released packages back to cpe
	for _, b := range buildsAlreadyReleased {
		reposBackToCPE = append(reposBackToCPE, b.repo)
	}
	addonDescriptor.Repositories = reposBackToCPE
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)

	return nil
}

func downloadSARXML(builds []buildWithRepository) ([]abaputils.Repository, error) {
	var reposBackToCPE []abaputils.Repository
	resultName := "SAR_XML"
	envPath := filepath.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment", "abap")
	for i, b := range builds {
		resultSARXML, err := b.build.getResult(resultName)
		if err != nil {
			return reposBackToCPE, err
		}
		sarPackage := resultSARXML.AdditionalInfo
		downloadPath := filepath.Join(envPath, path.Base(sarPackage))
		log.Entry().Infof("Downloading SAR file %s to %s", path.Base(sarPackage), downloadPath)
		err = resultSARXML.download(downloadPath)
		if err != nil {
			return reposBackToCPE, err
		}
		builds[i].repo.SarXMLFilePath = downloadPath
		reposBackToCPE = append(reposBackToCPE, builds[i].repo)
	}
	return reposBackToCPE, nil
}

func checkIfFailedAndPrintLogs(builds []buildWithRepository) error {
	var buildFailed bool = false
	for _, bR := range builds {
		b := bR.build
		if b.RunState == failed {
			log.Entry().Errorf("Assembly of %s failed", b.BuildID)
			buildFailed = true
		}
		b.printLogs()
	}
	if buildFailed {
		return errors.New("At least the assembly of one package failed")
	}
	return nil
}

func starting(repos []abaputils.Repository, conn connector) ([]buildWithRepository, []buildWithRepository, error) {
	var builds []buildWithRepository
	var buildsAlreadyReleased []buildWithRepository
	for _, repo := range repos {
		assemblyBuild := build{
			connector: conn,
		}
		buildRepo := buildWithRepository{
			build: assemblyBuild,
			repo:  repo,
		}
		if repo.Status == "P" {
			err := buildRepo.start()
			if err != nil {
				return builds, buildsAlreadyReleased, err
			}
			builds = append(builds, buildRepo)
		} else {
			log.Entry().Infof("Packages %s is already released. No need to run the assembly", repo.PackageName)
			buildsAlreadyReleased = append(buildsAlreadyReleased, buildRepo)
		}
	}
	return builds, buildsAlreadyReleased, nil
}

func polling(builds []buildWithRepository, maxRuntimeInMinutes time.Duration, pollIntervalsInSeconds time.Duration) error {
	timeout := time.After(maxRuntimeInMinutes * time.Minute)
	ticker := time.Tick(pollIntervalsInSeconds * time.Second)
	for {
		select {
		case <-timeout:
			return errors.New("Timed out")
		case <-ticker:
			var allFinished bool = true
			for i := range builds {
				if !builds[i].build.IsFinished() {
					builds[i].build.get()
					if !builds[i].build.IsFinished() {
						log.Entry().Infof("Assembly of %s is not yet finished, check again in %02d seconds", builds[i].repo.PackageName, pollIntervalsInSeconds)
						allFinished = false
					}
				}
			}
			if allFinished {
				return nil
			}
		}
	}
}

func (b *buildWithRepository) start() error {
	if b.repo.Name == "" || b.repo.Version == "" || b.repo.SpLevel == "" || b.repo.Namespace == "" || b.repo.PackageType == "" || b.repo.PackageName == "" {
		return errors.New("Parameters missing. Please provide software component name, version, sp-level, namespace, packagetype and packagename")
	}
	valuesInput := values{
		Values: []value{
			{
				ValueID: "SWC",
				Value:   b.repo.Name,
			},
			{
				ValueID: "CVERS",
				Value:   b.repo.Name + "." + b.repo.Version + "." + b.repo.SpLevel,
			},
			{
				ValueID: "NAMESPACE",
				Value:   b.repo.Namespace,
			},
			{
				ValueID: "PACKAGE_NAME_" + b.repo.PackageType,
				Value:   b.repo.PackageName,
			},
		},
	}
	if b.repo.PredecessorCommitID != "" {
		valuesInput.Values = append(valuesInput.Values,
			value{ValueID: "PREVIOUS_DELIVERY_COMMIT",
				Value: b.repo.PredecessorCommitID})
	}
	phase := "BUILD_" + b.repo.PackageType
	log.Entry().Infof("Starting assembly of package %s", b.repo.PackageName)
	return b.build.start(phase, valuesInput)
}

type buildWithRepository struct {
	build build
	repo  abaputils.Repository
}

// *******************************************************************************************************************************
// ************************************************************ REUSE ************************************************************
// *******************************************************************************************************************************

// *********************************************************************
// ******************************* Funcs *******************************
// *********************************************************************

// ******** technical communication settings ********

func (conn *connector) init(config *abapEnvironmentAssemblePackagesOptions, com abaputils.Communication, inputclient piperhttp.Sender) error {
	conn.Client = inputclient
	conn.Header = make(map[string][]string)
	conn.Header["Accept"] = []string{"application/json"}
	conn.Header["Content-Type"] = []string{"application/json"}
	conn.DownloadClient = &piperhttp.Client{}
	conn.DownloadClient.SetOptions(piperhttp.ClientOptions{TransportTimeout: 20 * time.Second})
	// Mapping for options
	subOptions := abaputils.AbapEnvironmentOptions{}
	subOptions.CfAPIEndpoint = config.CfAPIEndpoint
	subOptions.CfServiceInstance = config.CfServiceInstance
	subOptions.CfServiceKeyName = config.CfServiceKeyName
	subOptions.CfOrg = config.CfOrg
	subOptions.CfSpace = config.CfSpace
	subOptions.Host = config.Host
	subOptions.Password = config.Password
	subOptions.Username = config.Username

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, err := com.GetAbapCommunicationArrangementInfo(subOptions, "/sap/opu/odata/BUILD/CORE_SRV")
	if err != nil {
		return errors.Wrap(err, "Parameters for the ABAP Connection not available")
	}

	conn.DownloadClient.SetOptions(piperhttp.ClientOptions{
		Username: connectionDetails.User,
		Password: connectionDetails.Password,
	})
	cookieJar, _ := cookiejar.New(nil)
	conn.Client.SetOptions(piperhttp.ClientOptions{
		Username:  connectionDetails.User,
		Password:  connectionDetails.Password,
		CookieJar: cookieJar,
	})
	conn.Baseurl = connectionDetails.URL
	return nil
}

// ******** technical communication calls ********

func (conn *connector) getToken(appendum string) error {
	url := conn.Baseurl + appendum
	conn.Header["X-CSRF-Token"] = []string{"Fetch"}
	response, err := conn.Client.SendRequest("HEAD", url, nil, conn.Header, nil)
	if err != nil {
		if response == nil {
			return errors.Wrap(err, "Fetching X-CSRF-Token failed")
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errors.Wrapf(err, "Fetching X-CSRF-Token failed: %v", string(errorbody))

	}
	defer response.Body.Close()
	token := response.Header.Get("X-CSRF-Token")
	conn.Header["X-CSRF-Token"] = []string{token}
	return nil
}

func (conn connector) get(appendum string) ([]byte, error) {
	url := conn.Baseurl + appendum
	response, err := conn.Client.SendRequest("GET", url, nil, conn.Header, nil)
	if err != nil {
		if response == nil {
			return nil, errors.Wrap(err, "Get failed")
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Get failed: %v", string(errorbody))

	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, err
}

func (conn connector) post(appendum string, importBody string) ([]byte, error) {
	url := conn.Baseurl + appendum
	var response *http.Response
	var err error
	if importBody == "" {
		response, err = conn.Client.SendRequest("POST", url, nil, conn.Header, nil)
	} else {
		response, err = conn.Client.SendRequest("POST", url, bytes.NewBuffer([]byte(importBody)), conn.Header, nil)
	}
	if err != nil {
		if response == nil {
			return nil, errors.Wrap(err, "Post failed")
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Post failed: %v", string(errorbody))

	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, err
}

func (conn connector) download(appendum string, downloadPath string) error {
	url := conn.Baseurl + appendum
	err := conn.DownloadClient.DownloadFile(url, downloadPath, nil, nil)
	return err
}

// ******** BUILD logic ********

func (b *build) start(phase string, inputValues values) error {
	if err := b.getToken(""); err != nil {
		return err
	}
	importBody := inputForPost{
		phase:  phase,
		values: inputValues,
	}.String()

	body, err := b.connector.post("/builds", importBody)
	if err != nil {
		return err
	}

	var jBuild jsonBuild
	json.Unmarshal(body, &jBuild)
	b.BuildID = jBuild.Build.BuildID
	b.RunState = jBuild.Build.RunState
	b.ResultState = jBuild.Build.ResultState
	b.Phase = jBuild.Build.Phase
	b.Entitytype = jBuild.Build.Entitytype
	b.Startedby = jBuild.Build.Startedby
	b.StartedAt = jBuild.Build.StartedAt
	b.FinishedAt = jBuild.Build.FinishedAt
	return nil
}

func (b *build) get() error {
	appendum := "/builds('" + b.BuildID + "')"
	body, err := b.connector.get(appendum)
	if err != nil {
		return err
	}
	var jBuild jsonBuild
	json.Unmarshal(body, &jBuild)
	b.RunState = jBuild.Build.RunState
	b.ResultState = jBuild.Build.ResultState
	b.Phase = jBuild.Build.Phase
	b.Entitytype = jBuild.Build.Entitytype
	b.Startedby = jBuild.Build.Startedby
	b.StartedAt = jBuild.Build.StartedAt
	b.FinishedAt = jBuild.Build.FinishedAt
	return nil
}

func (b *build) getTasks() error {
	if len(b.Tasks) == 0 {
		appendum := "/builds('" + b.BuildID + "')/tasks"
		body, err := b.connector.get(appendum)
		if err != nil {
			return err
		}
		var jTasks jsonTasks
		json.Unmarshal(body, &jTasks)
		b.Tasks = jTasks.ResultTasks.Tasks
		sort.Slice(b.Tasks, func(i, j int) bool {
			return b.Tasks[i].TaskID < b.Tasks[j].TaskID
		})
		for i := range b.Tasks {
			b.Tasks[i].connector = b.connector
		}
	}
	return nil
}

func (b *build) getValues() error {
	if len(b.Values) == 0 {
		appendum := "/builds('" + b.BuildID + "')/values"
		body, err := b.connector.get(appendum)
		if err != nil {
			return err
		}
		var jValues jsonValues
		json.Unmarshal(body, &jValues)
		b.Values = jValues.ResultValues.Values
		for i := range b.Values {
			b.Values[i].connector = b.connector
		}
	}
	return nil
}

func (b *build) getLogs() error {
	if err := b.getTasks(); err != nil {
		return err
	}
	for i := range b.Tasks {
		if err := b.Tasks[i].getLogs(); err != nil {
			return err
		}
	}
	return nil
}

func (b *build) printLogs() error {
	if err := b.getTasks(); err != nil {
		return err
	}
	for i := range b.Tasks {
		if err := b.Tasks[i].printLogs(); err != nil {
			return err
		}
	}
	return nil
}

func (b *build) getResults() error {
	if err := b.getTasks(); err != nil {
		return err
	}
	for i := range b.Tasks {
		if err := b.Tasks[i].getResults(); err != nil {
			return err
		}
	}
	return nil
}

func (t *task) printLogs() error {
	if err := t.getLogs(); err != nil {
		return err
	}
	for _, logs := range t.Logs {
		logs.print()
	}
	return nil
}

func (b *build) getResult(name string) (result, error) {
	var Results []result
	var returnResult result
	if err := b.getResults(); err != nil {
		return returnResult, err
	}
	for _, task := range b.Tasks {
		for _, result := range task.Results {
			if result.Name == name {
				Results = append(Results, result)
			}
		}
	}
	switch len(Results) {
	case 0:
		return returnResult, errors.New("No result named " + name + " was found")
	case 1:
		return Results[0], nil
	default:
		return returnResult, errors.New("More than one result with the name " + name + " was found")
	}
}

func (b *build) IsFinished() bool {
	if b.RunState == finished || b.RunState == failed {
		return true
	}
	return false
}

func (t *task) getLogs() error {
	if len(t.Logs) == 0 {
		appendum := fmt.Sprint("/tasks(build_id='", t.BuildID, "',task_id=", t.TaskID, ")/logs")
		body, err := t.connector.get(appendum)
		if err != nil {
			return err
		}
		var jLogs jsonLogs
		json.Unmarshal(body, &jLogs)
		t.Logs = jLogs.ResultLogs.Logs
	}
	return nil
}

func (t *task) getResults() error {
	if len(t.Results) == 0 {
		appendum := fmt.Sprint("/tasks(build_id='", t.BuildID, "',task_id=", t.TaskID, ")/results")
		body, err := t.connector.get(appendum)
		if err != nil {
			return err
		}
		var jResults jsonResults
		json.Unmarshal(body, &jResults)
		t.Results = jResults.ResultResults.Results
		for i := range t.Results {
			t.Results[i].connector = t.connector
		}
	}
	return nil
}

func (result *result) download(downloadPath string) error {
	appendum := fmt.Sprint("/results(build_id='", result.BuildID, "',task_id=", result.TaskID, ",name='", result.Name, "')/$value")
	err := result.connector.download(appendum, downloadPath)
	return err
}

func (logging *logStruct) print() {
	switch logging.Msgty {
	case loginfo:
		log.Entry().WithField("Timestamp", logging.Timestamp).Info(logging.Logline)
	case logwarning:
		log.Entry().WithField("Timestamp", logging.Timestamp).Warn(logging.Logline)
	case logerror:
		log.Entry().WithField("Timestamp", logging.Timestamp).Error(logging.Logline)
	case logaborted:
		log.Entry().WithField("Timestamp", logging.Timestamp).Error(logging.Logline)
	default:
	}
}

// ******** parsing ********
func (v value) String() string {
	return fmt.Sprintf(
		`{ "value_id": "%s", "value": "%s" }`,
		v.ValueID,
		v.Value)
}

func (vs values) String() string {
	returnString := ""
	for _, value := range vs.Values {
		returnString = returnString + value.String() + ",\n"
	}
	returnString = returnString[:len(returnString)-2] //removes last ,
	return returnString
}

func (in inputForPost) String() string {
	return fmt.Sprintf(`{ "phase": "%s", "values": [%s]}`, in.phase, in.values.String())
}

// *********************************************************************
// ****************************** Structs ******************************
// *********************************************************************

type resultState string
type runState string
type msgty string

const (
	successful   resultState = "SUCCESSFUL"
	warning      resultState = "WARNING"
	erroneous    resultState = "ERRONEOUS"
	aborted      resultState = "ABORTED"
	initializing runState    = "INITIALIZING"
	accepted     runState    = "ACCEPTED"
	running      runState    = "RUNNING"
	finished     runState    = "FINISHED"
	failed       runState    = "FAILED"
	loginfo      msgty       = "I"
	logwarning   msgty       = "W"
	logerror     msgty       = "E"
	logaborted   msgty       = "A"
)

type connector struct {
	Client         piperhttp.Sender
	DownloadClient piperhttp.Downloader
	Header         map[string][]string
	Baseurl        string
}

//******** structs needed for json convertion ********

type jsonBuild struct {
	Build *build `json:"d"`
}

type jsonTasks struct {
	ResultTasks struct {
		Tasks []task `json:"results"`
	} `json:"d"`
}

type jsonLogs struct {
	ResultLogs struct {
		Logs []logStruct `json:"results"`
	} `json:"d"`
}

type jsonResults struct {
	ResultResults struct {
		Results []result `json:"results"`
	} `json:"d"`
}

type jsonValues struct {
	ResultValues struct {
		Values []value `json:"results"`
	} `json:"d"`
}

// ******** resembling data model in backend ********

type build struct {
	connector
	BuildID     string      `json:"build_id"`
	RunState    runState    `json:"run_state"`
	ResultState resultState `json:"result_state"`
	Phase       string      `json:"phase"`
	Entitytype  string      `json:"entitytype"`
	Startedby   string      `json:"startedby"`
	StartedAt   string      `json:"started_at"`
	FinishedAt  string      `json:"finished_at"`
	Tasks       []task
	Values      []value
}

type task struct {
	connector
	BuildID     string      `json:"build_id"`
	TaskID      int         `json:"task_id"`
	LogID       string      `json:"log_id"`
	PluginClass string      `json:"plugin_class"`
	StartedAt   string      `json:"started_at"`
	FinishedAt  string      `json:"finished_at"`
	ResultState resultState `json:"result_state"`
	Logs        []logStruct
	Results     []result
}

type logStruct struct {
	BuildID   string `json:"build_id"`
	TaskID    int    `json:"task_id"`
	LogID     string `json:"log_id"`
	Msgty     msgty  `json:"msgty"`
	Detlevel  string `json:"detlevel"`
	Logline   string `json:"log_line"`
	Timestamp string `json:"TIME_STMP"`
}

type result struct {
	connector
	BuildID        string `json:"build_id"`
	TaskID         int    `json:"task_id"`
	Name           string `json:"name"`
	AdditionalInfo string `json:"additional_info"`
	Mimetype       string `json:"mimetype"`
}

type value struct {
	connector
	BuildID string `json:"build_id"`
	ValueID string `json:"value_id"`
	Value   string `json:"value"`
}

// ******** import structure to post call ********

type inputForPost struct {
	phase  string
	values values
}

type values struct {
	Values []value `json:"results"`
}
