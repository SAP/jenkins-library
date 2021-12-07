package build

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

// RunState : Current Status of the Build
type RunState string
type resultState string
type msgty string

const (
	successful resultState = "SUCCESSFUL"
	warning    resultState = "WARNING"
	erroneous  resultState = "ERRONEOUS"
	aborted    resultState = "ABORTED"
	// Initializing : Build Framework prepared
	Initializing RunState = "INITIALIZING"
	// Accepted : Build Framework triggered
	Accepted RunState = "ACCEPTED"
	// Running : Build Framework performs build
	Running RunState = "RUNNING"
	// Finished : Build Framework ended successful
	Finished RunState = "FINISHED"
	// Failed : Build Framework endded with error
	Failed          RunState = "FAILED"
	loginfo         msgty    = "I"
	logwarning      msgty    = "W"
	logerror        msgty    = "E"
	logaborted      msgty    = "A"
	dummyResultName string   = "Dummy"
)

//******** structs needed for json convertion ********
type jsonBuild struct {
	Build struct {
		BuildID     string      `json:"build_id"`
		RunState    RunState    `json:"run_state"`
		ResultState resultState `json:"result_state"`
		Phase       string      `json:"phase"`
		Entitytype  string      `json:"entitytype"`
		Startedby   string      `json:"startedby"`
		StartedAt   string      `json:"started_at"`
		FinishedAt  string      `json:"finished_at"`
	} `json:"d"`
}

type jsonTasks struct {
	ResultTasks struct {
		Tasks []jsonTask `json:"results"`
	} `json:"d"`
}

type jsonTask struct {
	BuildID     string      `json:"build_id"`
	TaskID      int         `json:"task_id"`
	LogID       string      `json:"log_id"`
	PluginClass string      `json:"plugin_class"`
	StartedAt   string      `json:"started_at"`
	FinishedAt  string      `json:"finished_at"`
	ResultState resultState `json:"result_state"`
}

type jsonLogs struct {
	ResultLogs struct {
		Logs []logStruct `json:"results"`
	} `json:"d"`
}

type jsonResults struct {
	ResultResults struct {
		Results []Result `json:"results"`
	} `json:"d"`
}

type jsonValues struct {
	ResultValues struct {
		Values []Value `json:"results"`
	} `json:"d"`
}

// ******** resembling data model in backend ********

// Build : Information for all data comming from Build Framework
type Build struct {
	Connector   Connector
	BuildID     string      `json:"build_id"`
	RunState    RunState    `json:"run_state"`
	ResultState resultState `json:"result_state"`
	Phase       string      `json:"phase"`
	Entitytype  string      `json:"entitytype"`
	Startedby   string      `json:"startedby"`
	StartedAt   string      `json:"started_at"`
	FinishedAt  string      `json:"finished_at"`
	Tasks       []task
	Values      []Value
}

type task struct {
	connector   Connector
	BuildID     string      `json:"build_id"`
	TaskID      int         `json:"task_id"`
	LogID       string      `json:"log_id"`
	PluginClass string      `json:"plugin_class"`
	StartedAt   string      `json:"started_at"`
	FinishedAt  string      `json:"finished_at"`
	ResultState resultState `json:"result_state"`
	Logs        []logStruct
	Results     []Result
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

// Result : Artefact from Build Framework step
type Result struct {
	connector      Connector
	BuildID        string `json:"build_id"`
	TaskID         int    `json:"task_id"`
	Name           string `json:"name"`
	AdditionalInfo string `json:"additional_info"`
	Mimetype       string `json:"mimetype"`
	SavedFilename  string
	DownloadPath   string
}

// Value : Returns Build Runtime Value
type Value struct {
	connector Connector
	BuildID   string `json:"build_id"`
	ValueID   string `json:"value_id"`
	Value     string `json:"value"`
}

// Values : Returns Build Runtime Values
type Values struct {
	Values []Value `json:"results"`
}

type inputForPost struct {
	phase  string
	values Values
}

// *********************************************************************
// ******************************* Funcs *******************************
// *********************************************************************

// Start : Starts the Build Framework
func (b *Build) Start(phase string, inputValues Values) error {
	if err := b.Connector.GetToken(""); err != nil {
		return err
	}
	importBody := inputForPost{
		phase:  phase,
		values: inputValues,
	}.String()

	body, err := b.Connector.Post("/builds", importBody)
	if err != nil {
		return err
	}

	var jBuild jsonBuild
	if err := json.Unmarshal(body, &jBuild); err != nil {
		return errors.Wrap(err, "Unexpected buildFrameWork response: "+string(body))
	}
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

// Poll : waits for the build framework to be finished
func (b *Build) Poll() error {
	timeout := time.After(b.Connector.MaxRuntime)
	ticker := time.Tick(b.Connector.PollingInterval)
	for {
		select {
		case <-timeout:
			return errors.Errorf("Timed out: (max Runtime %v reached)", b.Connector.MaxRuntime)
		case <-ticker:
			b.Get()
			if !b.IsFinished() {
				log.Entry().Infof("Build is not yet finished, check again in %s", b.Connector.PollingInterval)
			} else {
				return nil
			}
		}
	}
}

// EvaluteIfBuildSuccessful : Checks the finale state of the build framework
func (b *Build) EvaluteIfBuildSuccessful(treatWarningsAsError bool) error {
	if b.RunState == Failed {
		return errors.Errorf("Build ended with runState failed")
	}
	if treatWarningsAsError && b.ResultState == warning {
		return errors.Errorf("Build ended with resultState warning, setting to failed as configured")
	}
	if (b.ResultState == aborted) || (b.ResultState == erroneous) {
		return errors.Errorf("Build ended with resultState %s", b.ResultState)
	}
	return nil
}

// Get : Get all Build tasks
func (b *Build) Get() error {
	appendum := "/builds('" + url.QueryEscape(b.BuildID) + "')"
	body, err := b.Connector.Get(appendum)
	if err != nil {
		return err
	}
	var jBuild jsonBuild
	if err := json.Unmarshal(body, &jBuild); err != nil {
		return errors.Wrap(err, "Unexpected buildFrameWork response: "+string(body))
	}
	b.RunState = jBuild.Build.RunState
	b.ResultState = jBuild.Build.ResultState
	b.Phase = jBuild.Build.Phase
	b.Entitytype = jBuild.Build.Entitytype
	b.Startedby = jBuild.Build.Startedby
	b.StartedAt = jBuild.Build.StartedAt
	b.FinishedAt = jBuild.Build.FinishedAt
	return nil
}

func (b *Build) getTasks() error {
	if len(b.Tasks) == 0 {
		appendum := "/builds('" + url.QueryEscape(b.BuildID) + "')/tasks"
		body, err := b.Connector.Get(appendum)
		if err != nil {
			return err
		}
		b.Tasks, err = unmarshalTasks(body, b.Connector)
		if err != nil {
			return err
		}
		sort.Slice(b.Tasks, func(i, j int) bool {
			return b.Tasks[i].TaskID < b.Tasks[j].TaskID
		})
	}
	return nil
}

// GetValues : Gets all Build values
func (b *Build) GetValues() error {
	if len(b.Values) == 0 {
		appendum := "/builds('" + url.QueryEscape(b.BuildID) + "')/values"
		body, err := b.Connector.Get(appendum)
		if err != nil {
			return err
		}
		var jValues jsonValues
		if err := json.Unmarshal(body, &jValues); err != nil {
			return errors.Wrap(err, "Unexpected buildFrameWork response: "+string(body))
		}
		b.Values = jValues.ResultValues.Values
		for i := range b.Values {
			b.Values[i].connector = b.Connector
		}
	}
	return nil
}

func (b *Build) getLogs() error {
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

// PrintLogs : Returns the Build logs
func (b *Build) PrintLogs() error {
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

// GetResults : Gets all Build results
func (b *Build) GetResults() error {
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

// GetResult : Returns the last Build artefact created from build step
func (b *Build) GetResult(name string) (*Result, error) {
	var Results []*Result
	var returnResult Result
	if err := b.GetResults(); err != nil {
		return &returnResult, err
	}
	for i_task := range b.Tasks {
		for i_result := range b.Tasks[i_task].Results {
			if b.Tasks[i_task].Results[i_result].Name == name {
				Results = append(Results, &b.Tasks[i_task].Results[i_result])
			}
		}
	}
	switch len(Results) {
	case 0:
		return &returnResult, errors.New("No result named " + name + " was found")
	case 1:
		return Results[0], nil
	default:
		return &returnResult, errors.New("More than one result with the name " + name + " was found")
	}
}

// IsFinished : Returns Build run state
func (b *Build) IsFinished() bool {
	if b.RunState == Finished || b.RunState == Failed {
		return true
	}
	return false
}

func (t *task) getLogs() error {
	if len(t.Logs) == 0 {
		appendum := fmt.Sprint("/tasks(build_id='", url.QueryEscape(t.BuildID), "',task_id=", t.TaskID, ")/logs")
		body, err := t.connector.Get(appendum)
		if err != nil {
			return err
		}
		var jLogs jsonLogs
		if err := json.Unmarshal(body, &jLogs); err != nil {
			return errors.Wrap(err, "Unexpected buildFrameWork response: "+string(body))
		}
		t.Logs = jLogs.ResultLogs.Logs
	}
	return nil
}

func (t *task) getResults() error {
	if len(t.Results) == 0 {
		appendum := fmt.Sprint("/tasks(build_id='", url.QueryEscape(t.BuildID), "',task_id=", t.TaskID, ")/results")
		body, err := t.connector.Get(appendum)
		if err != nil {
			return err
		}
		var jResults jsonResults
		if err := json.Unmarshal(body, &jResults); err != nil {
			return errors.Wrap(err, "Unexpected buildFrameWork response: "+string(body))
		}
		t.Results = jResults.ResultResults.Results
		for i := range t.Results {
			t.Results[i].connector = t.connector
		}
		if len(t.Results) == 0 {
			//prevent 2nd GET request - no new results will occure...
			t.Results = append(t.Results, Result{Name: dummyResultName})
		}
	}
	return nil
}

// DownloadAllResults : Downloads all build artefacts, saves it to basePath and the filenames can be modified with the filenamePrefix
func (b *Build) DownloadAllResults(basePath string, filenamePrefix string) error {
	if err := b.GetResults(); err != nil {
		return err
	}
	for i_task := range b.Tasks {
		//in case there was no result, there is only one entry with dummyResultName, obviously we don't want to download this
		if b.Tasks[i_task].Results[0].Name != dummyResultName {
			for i_result := range b.Tasks[i_task].Results {
				if err := b.Tasks[i_task].Results[i_result].DownloadWithFilenamePrefixAndTargetDirectory(basePath, filenamePrefix); err != nil {
					return errors.Wrapf(err, "Error during the download of file %s", b.Tasks[i_task].Results[i_result].Name)
				}
			}
		}
	}
	return nil
}

// DownloadResults : Download results which are specified in filenames
func (b *Build) DownloadResults(filenames []string, basePath string, filenamePrefix string) error {
	for _, name := range filenames {
		result, err := b.GetResult(name)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "Problems finding the file %s, please check your config whether this file is really a result file", name)
		}
		if err := result.DownloadWithFilenamePrefixAndTargetDirectory(basePath, filenamePrefix); err != nil {
			return errors.Wrapf(err, "Error during the download of file %s", name)
		}
	}
	return nil
}

// PublishAllDownloadedResults : publishes all build artefacts which were downloaded before
func (b *Build) PublishAllDownloadedResults(stepname string, publish Publish) {
	var filesToPublish []piperutils.Path
	for i_task := range b.Tasks {
		for i_result := range b.Tasks[i_task].Results {
			if b.Tasks[i_task].Results[i_result].wasDownloaded() {
				filesToPublish = append(filesToPublish, piperutils.Path{Target: b.Tasks[i_task].Results[i_result].DownloadPath,
					Name: b.Tasks[i_task].Results[i_result].SavedFilename, Mandatory: true})
			}
		}
	}
	if len(filesToPublish) > 0 {
		publish.PersistReportsAndLinks(stepname, "", filesToPublish, nil)
	}
}

// PublishDownloadedResults : Publishes build artefacts specified in filenames
func (b *Build) PublishDownloadedResults(stepname string, filenames []string, publish Publish) error {
	var filesToPublish []piperutils.Path
	for i := range filenames {
		result, err := b.GetResult(filenames[i])
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "Problems finding the file %s, please check your config whether this file is really a result file", filenames[i])
		}
		if result.wasDownloaded() {
			filesToPublish = append(filesToPublish, piperutils.Path{Target: result.DownloadPath, Name: result.SavedFilename, Mandatory: true})
		} else {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Errorf("Trying to publish the file %s which was not downloaded", result.Name)
		}
	}
	if len(filesToPublish) > 0 {
		publish.PersistReportsAndLinks(stepname, "", filesToPublish, nil)
	}
	return nil
}

// Download : Provides the atrefact of build step
func (result *Result) Download(downloadPath string) error {
	appendum := fmt.Sprint("/results(build_id='", url.QueryEscape(result.BuildID), "',task_id=", result.TaskID, ",name='", url.QueryEscape(result.Name), "')/$value")
	err := result.connector.Download(appendum, downloadPath)
	return err
}

// DownloadWithFilenamePrefixAndTargetDirectory : downloads build artefact, saves it to basePath and the filename can be modified with the filenamePrefix
func (result *Result) DownloadWithFilenamePrefixAndTargetDirectory(basePath string, filenamePrefix string) error {
	basePath, err := result.resolveParamter(basePath)
	if err != nil {
		return errors.Wrapf(err, "Could not resolve parameter %s for the target directory", basePath)
	}
	filenamePrefix, err = result.resolveParamter(filenamePrefix)
	if err != nil {
		return errors.Wrapf(err, "Could not resolve parameter %s for the filename prefix", filenamePrefix)
	}
	appendum := fmt.Sprint("/results(build_id='", result.BuildID, "',task_id=", result.TaskID, ",name='", result.Name, "')/$value")
	filename := filenamePrefix + result.Name
	downloadPath := filepath.Join(path.Base(basePath), path.Base(filename))
	if err := result.connector.Download(appendum, downloadPath); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return errors.Wrapf(err, "Could not download %s", result.Name)
	}
	result.SavedFilename = filename
	result.DownloadPath = downloadPath
	log.Entry().Infof("Saved file %s as %s to %s", result.Name, result.SavedFilename, result.DownloadPath)
	return nil
}

func (result *Result) resolveParamter(parameter string) (string, error) {
	if len(parameter) == 0 {
		return parameter, nil
	}
	if (string(parameter[0]) == "{") && string(parameter[len(parameter)-1]) == "}" {
		trimmedParam := strings.ToLower(parameter[1 : len(parameter)-1])
		switch trimmedParam {
		case "buildid":
			return result.BuildID, nil
		case "taskid":
			return strconv.Itoa(result.TaskID), nil
		default:
			log.SetErrorCategory(log.ErrorConfiguration)
			return "", errors.Errorf("Unknown parameter %s", parameter)
		}
	} else {
		return parameter, nil
	}
}

func (result *Result) wasDownloaded() bool {
	if len(result.DownloadPath) > 0 && len(result.SavedFilename) > 0 {
		return true
	} else {
		return false
	}
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
func (v Value) String() string {
	return fmt.Sprintf(
		`{ "value_id": "%s", "value": "%s" }`,
		v.ValueID,
		v.Value)
}

func (vs Values) String() string {
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

//******** unmarshal function  ************
func unmarshalTasks(body []byte, connector Connector) ([]task, error) {

	var tasks []task
	var append_task task
	var jTasks jsonTasks
	if err := json.Unmarshal(body, &jTasks); err != nil {
		return tasks, errors.Wrap(err, "Unexpected buildFrameWork response: "+string(body))
	}
	for _, jTask := range jTasks.ResultTasks.Tasks {
		append_task.connector = connector
		append_task.BuildID = jTask.BuildID
		append_task.TaskID = jTask.TaskID
		append_task.LogID = jTask.LogID
		append_task.PluginClass = jTask.PluginClass
		append_task.StartedAt = jTask.StartedAt
		append_task.FinishedAt = jTask.FinishedAt
		append_task.ResultState = jTask.ResultState
		tasks = append(tasks, append_task)
	}
	return tasks, nil
}

// *****************publish *******************************
type Publish interface {
	PersistReportsAndLinks(stepName, workspace string, reports, links []piperutils.Path)
}

func PersistReportsAndLinks(stepName, workspace string, reports, links []piperutils.Path) {
	piperutils.PersistReportsAndLinks(stepName, workspace, reports, links)
}
