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
	"net/http/cookiejar"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

//todo:
//durch groovy skripte gehen -> hab ich was vergessen?
//aufräumen & logging
func abapEnvironmentAssembly(config abapEnvironmentAssemblyOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *abapEnvironmentAssemblyCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentAssembly(&config, telemetryData, &c, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentAssembly(config *abapEnvironmentAssemblyOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, commonPipelineEnvironment *abapEnvironmentAssemblyCommonPipelineEnvironment) error {
	// err := testDownload(config)
	err := startBuildAndLogs(config)
	// err := testGet(config)
	return err
}

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

//structs needed for json convertion

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

// resembling data model in backend

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

// import structure to post call

type inputForPost struct {
	phase  string
	values values
}

type values struct {
	Values []value `json:"results"`
}

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
	return fmt.Sprintf(
		`{ "phase": "%s",
		   "values": [ 
			   %s
		   ]
		   }`,
		in.phase,
		in.values.String())
}

// var cf = cloudfoundry.CFUtils{Exec: &command.Command{}}
// var cfReadServiceKey = cf.ReadServiceKeyAbapEnvironment

var getAbapCommunicationArrangement = abaputils.GetAbapCommunicationArrangementInfo

func (conn *connector) setConnectionDetails(options abapEnvironmentAssemblyOptions) error {
	// Mapping for options
	subOptions := abaputils.AbapEnvironmentOptions{}

	subOptions.CfAPIEndpoint = options.CfAPIEndpoint
	subOptions.CfServiceInstance = options.CfServiceInstance
	subOptions.CfServiceKeyName = options.CfServiceKeyName
	subOptions.CfOrg = options.CfOrg
	subOptions.CfSpace = options.CfSpace
	subOptions.Host = options.Host
	subOptions.Password = options.Password
	subOptions.Username = options.Username

	var c command.ExecRunner = &command.Command{}

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	// connectionDetails, errorGetInfo := abaputils.GetAbapCommunicationArrangementInfo(subOptions, c, "/sap/opu/odata/BUILD/CORE_SRV")
	connectionDetails, errorGetInfo := getAbapCommunicationArrangement(subOptions, c, "/sap/opu/odata/BUILD/CORE_SRV")
	if errorGetInfo != nil {
		log.Entry().WithError(errorGetInfo).Fatal("Parameters for the ABAP Connection not available")
	}

	conn.DownloadClient.SetOptions(piperhttp.ClientOptions{
		Username: connectionDetails.User,
		Password: connectionDetails.Password,
	})
	cookieJar, _ := cookiejar.New(nil)
	//TODO soll das benutzt werden?
	conn.Client.SetOptions(piperhttp.ClientOptions{
		// MaxRequestDuration: 180 * time.Second,
		Username:  connectionDetails.User,
		Password:  connectionDetails.Password,
		CookieJar: cookieJar,
	})
	conn.Baseurl = connectionDetails.URL
	return nil
}

//TODO das hatte ich fürs mocking benutzt -> checken wie ich das jetzt ersetze und dann löschen
// var cf = cloudfoundry.CFUtils{Exec: &command.Command{}}
// var cfReadServiceKey = cf.ReadServiceKeyAbapEnvironment

// func (conn *connector) setConnectionDetailsFromCF(config abapEnvironmentAssemblyOptions) error {
// 	cfconfig := cloudfoundry.ServiceKeyOptions{
// 		CfAPIEndpoint:     config.CfAPIEndpoint,
// 		CfOrg:             config.CfOrg,
// 		CfSpace:           config.CfSpace,
// 		Username:          config.Username,
// 		Password:          config.Password,
// 		CfServiceInstance: config.CfServiceInstance,
// 		CfServiceKey:      config.CfServiceKeyName,
// 	}
// 	if cfconfig.CfServiceInstance == "" || cfconfig.CfOrg == "" || cfconfig.CfAPIEndpoint == "" || cfconfig.CfSpace == "" || cfconfig.CfServiceKey == "" {
// 		return errors.New("Parameters missing. Please provide EITHER the Host of the ABAP server OR the Cloud Foundry ApiEndpoint, Organization, Space, Service Instance and a corresponding Service Key for the Communication Scenario SAP_COM_0582")
// 	}
// 	abapServiceKey, err := cfReadServiceKey(cfconfig, true)
// 	if err != nil {
// 		return fmt.Errorf("Reading Service Key failed: %w", err)
// 	}

// 	conn.DownloadClient.SetOptions(piperhttp.ClientOptions{
// 		Username: abapServiceKey.Abap.Username,
// 		Password: abapServiceKey.Abap.Password,
// 	})
// 	cookieJar, _ := cookiejar.New(nil)
// 	conn.Client.SetOptions(piperhttp.ClientOptions{
// 		Username:  abapServiceKey.Abap.Username,
// 		Password:  abapServiceKey.Abap.Password,
// 		CookieJar: cookieJar,
// 	})
// 	conn.Baseurl = abapServiceKey.URL
// 	return nil
// }

func (conn *connector) setupAttributes(inputclient piperhttp.Sender) {
	conn.Client = inputclient
	conn.Header = make(map[string][]string)
	conn.Header["Accept"] = []string{"application/json"}
	conn.Header["Content-Type"] = []string{"application/json"}
	conn.DownloadClient = &piperhttp.Client{}
	conn.DownloadClient.SetOptions(piperhttp.ClientOptions{TransportTimeout: 20 * time.Second})
}

func (conn *connector) getToken() error {
	conn.Header["X-CSRF-Token"] = []string{"Fetch"}
	response, err := conn.Client.SendRequest("HEAD", conn.Baseurl, nil, conn.Header, nil)
	if err != nil {
		return fmt.Errorf("Fetching Xcsrf-Token failed: %w", err)
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
			return nil, err
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Get failed %v", string(errorbody))

	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, err
}

func (conn connector) post(importBody string) ([]byte, error) {
	url := conn.Baseurl + "/builds"
	byteBody := bytes.NewBuffer([]byte(importBody))
	response, err := conn.Client.SendRequest("POST", url, byteBody, conn.Header, nil)
	if err != nil {
		if response == nil {
			return nil, err
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Post failed %v", string(errorbody))

	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, err
}

//TODO wohin speichern?
func (conn connector) download(appendum string, downloadPath string) error {
	url := conn.Baseurl + appendum
	// tmpFolder := createTempDir()
	// archive := filepath.Join(tmpFolder, path.Base(name))
	// archive := filepath.Join(".", path.Base(name))

	err := conn.DownloadClient.DownloadFile(url, downloadPath, nil, nil)
	return err
}

func (b *build) start(phase string, inputValues values) error {
	if err := b.getToken(); err != nil {
		return err
	}
	importBody := inputForPost{
		phase:  phase,
		values: inputValues,
	}.String()

	// TODO inline declaration
	var body []byte
	var err error
	body, err = b.connector.post(importBody)
	if err != nil {
		return err
	}

	var data jsonBuild
	json.Unmarshal(body, &data)
	b.BuildID = data.Build.BuildID
	b.RunState = data.Build.RunState
	b.ResultState = data.Build.ResultState
	b.Phase = data.Build.Phase
	b.Entitytype = data.Build.Entitytype
	b.Startedby = data.Build.Startedby
	b.StartedAt = data.Build.StartedAt
	b.FinishedAt = data.Build.FinishedAt
	return nil
}

func (b *build) get() error {
	appendum := "/builds('" + b.BuildID + "')"
	body, err := b.connector.get(appendum)
	if err != nil {
		return err
	}
	var data jsonBuild
	json.Unmarshal(body, &data)
	b.RunState = data.Build.RunState
	b.ResultState = data.Build.ResultState
	b.Phase = data.Build.Phase
	b.Entitytype = data.Build.Entitytype
	b.Startedby = data.Build.Startedby
	b.StartedAt = data.Build.StartedAt
	b.FinishedAt = data.Build.FinishedAt
	return nil
}

func (b *build) getTasks() error {
	if len(b.Tasks) == 0 {
		appendum := "/builds('" + b.BuildID + "')/tasks"
		body, err := b.connector.get(appendum)
		if err != nil {
			return err
		}
		var data jsonTasks
		json.Unmarshal(body, &data)
		b.Tasks = data.ResultTasks.Tasks
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
		var data jsonValues
		json.Unmarshal(body, &data)
		b.Values = data.ResultValues.Values
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

func (b *build) poll(maxRuntimeInMinutes time.Duration, pollIntervalsInSeconds time.Duration) (bool, error) {
	timeout := time.After(maxRuntimeInMinutes * time.Minute)
	ticker := time.Tick(pollIntervalsInSeconds * time.Second)
	for {
		select {
		case <-timeout:
			return false, errors.New("timed out")
		case <-ticker:
			b.get()
			ok, err := b.IsFinished()
			if ok {
				return true, err
			}
		}
	}
}

func (b *build) IsFinished() (bool, error) {
	switch b.RunState {
	case finished:
		return true, nil
	case failed:
		return true, errors.New("build failed")
	default:
		return false, nil
	}
}

func (t *task) getLogs() error {
	if len(t.Logs) == 0 {
		appendum := fmt.Sprint("/tasks(build_id='", t.BuildID, "',task_id=", t.TaskID, ")/logs")
		body, err := t.connector.get(appendum)
		if err != nil {
			return err
		}
		var data jsonLogs
		json.Unmarshal(body, &data)
		t.Logs = data.ResultLogs.Logs
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
		var data jsonResults
		json.Unmarshal(body, &data)
		t.Results = data.ResultResults.Results
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

// ############################## delete start ################################

func startBuildAndLogs(config *abapEnvironmentAssemblyOptions) error {

	conn := new(connector)
	conn.setupAttributes(&piperhttp.Client{})
	err := conn.setConnectionDetails(*config)
	if err != nil {
		return err
	}
	build1 := build{
		connector: *conn,
		// BuildID:   "ABIFNLDCSQPNVKWNGL2US3IB4Q",
	}
	valuesInput := values{
		Values: []value{
			// {
			// 	ValueID: "SWC",
			// 	Value:   config.SWC,
			// },
			{
				ValueID: "PACKAGES",
				Value:   "/BUILD/CORE",
			},
			{
				ValueID: "SOFTWARE_COMPONENT",
				Value:   config.SWC,
			},
			{
				ValueID: "CVERS",
				Value:   config.CVERS,
			},
			{
				ValueID: "NAMESPACE",
				Value:   config.Namespace,
			},
			{
				ValueID: "PREVIOUS_DELIVERY_COMMIT",
				Value:   config.PreviousDeliveryCommit,
			},
			{
				ValueID: "PACKAGE_NAME_" + config.PackageType,
				Value:   config.PackageName,
			},
		},
	}

	//TODO phase build_aoi etc testten
	// phase := "BUILD_" + config.PackageType
	phase := "test1"
	// if err := build1.start(phase, valuesInput); err != nil {
	// 	return err
	// }

	err = build1.start(phase, valuesInput)
	if err != nil {
		return err
	}
	build1.poll(15, 60)
	build1.getTasks()
	build1.getLogs()
	build1.getResults()
	result, err := build1.getResult("HT-6100.JPG")
	if err != nil {
		return err
	}
	fmt.Println(result.Mimetype)
	build1.printLogs()

	build1.getValues()

	for i, value := range build1.Values {
		fmt.Println(i)
		fmt.Println(value)
		fmt.Println("-----------------------------------")
	}

	return nil
}

func testDownload(config *abapEnvironmentAssemblyOptions) error {

	conn := new(connector)
	conn.setupAttributes(&piperhttp.Client{})
	err := conn.setConnectionDetails(*config)
	if err != nil {
		return err
	}
	b := build{
		connector: *conn,
		BuildID:   "ABIFNLDCSQPOVJOUDMJG2M37OU",
	}
	r, err := b.getResult("2times_hello")
	if err != nil {
		return err
	}

	envPath := filepath.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
	fmt.Println(envPath)
	downloadPath := filepath.Join(envPath, path.Base("SAR_XML"))
	fmt.Println(downloadPath)
	err = r.download(downloadPath)
	return err

}

func createTempDir() string {
	tmpFolder, err := ioutil.TempDir(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
	}
	return tmpFolder
}

func testGet(config *abapEnvironmentAssemblyOptions) error {

	conn := new(connector)
	conn.setupAttributes(&piperhttp.Client{})
	conn.setConnectionDetails(*config)
	b := build{
		connector: *conn,
		BuildID:   "ABIFNLDCSQPNVNE2XXXMMBC2KY",
	}
	err := b.get()
	if err != nil {
		return err
	}

	fmt.Println(b.RunState)
	fmt.Println(b.ResultState)

	err = b.printLogs()
	return err
}

// ############################ delete ende ##################
