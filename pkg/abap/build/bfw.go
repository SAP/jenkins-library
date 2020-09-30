package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/SAP/jenkins-library/pkg/log"
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
	Failed     RunState = "FAILED"
	loginfo    msgty    = "I"
	logwarning msgty    = "W"
	logerror   msgty    = "E"
	logaborted msgty    = "A"
)

//******** structs needed for json convertion ********

type jsonBuild struct {
	Build *Build `json:"d"`
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
	json.Unmarshal(body, &jBuild)
	b.BuildID = jBuild.Build.BuildID
	b.RunState = jBuild.Build.RunState
	b.ResultState = jBuild.Build.ResultState
	b.Phase = jBuild.Build.Phase
	b.Entitytype = jBuild.Build.Entitytype
	b.Startedby = jBuild.Build.Startedby
	b.StartedAt = jBuild.Build.StartedAt
	b.FinishedAt = jBuild.Build.FinishedAt
	fmt.Println("Hello")
	return nil
}

// Get : Get all Build tasks
func (b *Build) Get() error {
	appendum := "/builds('" + b.BuildID + "')"
	body, err := b.Connector.Get(appendum)
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

func (b *Build) getTasks() error {
	if len(b.Tasks) == 0 {
		appendum := "/builds('" + b.BuildID + "')/tasks"
		body, err := b.Connector.Get(appendum)
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
			b.Tasks[i].connector = b.Connector
		}
	}
	return nil
}

func (b *Build) getValues() error {
	if len(b.Values) == 0 {
		appendum := "/builds('" + b.BuildID + "')/values"
		body, err := b.Connector.Get(appendum)
		if err != nil {
			return err
		}
		var jValues jsonValues
		json.Unmarshal(body, &jValues)
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

func (b *Build) getResults() error {
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
func (b *Build) GetResult(name string) (Result, error) {
	var Results []Result
	var returnResult Result
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

// IsFinished : Returns Build run state
func (b *Build) IsFinished() bool {
	if b.RunState == Finished || b.RunState == Failed {
		return true
	}
	return false
}

func (t *task) getLogs() error {
	if len(t.Logs) == 0 {
		appendum := fmt.Sprint("/tasks(build_id='", t.BuildID, "',task_id=", t.TaskID, ")/logs")
		body, err := t.connector.Get(appendum)
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
		body, err := t.connector.Get(appendum)
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

// Download : Provides the atrefact of build step
func (result *Result) Download(downloadPath string) error {
	appendum := fmt.Sprint("/results(build_id='", result.BuildID, "',task_id=", result.TaskID, ",name='", result.Name, "')/$value")
	err := result.connector.Download(appendum, downloadPath)
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
