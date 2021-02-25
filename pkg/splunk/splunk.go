package splunk

import (
	"bytes"
	"encoding/json"
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// SplunkHook provides a logrus hook which enables error logging to splunk platform.
// This is helpful in order to provide better monitoring and alerting on errors
// as well as the given error details can help to find the root cause of bugs.
type Splunk struct {
	levels        []logrus.Level
	tags          map[string]string
	splunkClient  piperhttp.Client
	correlationID string
	splunkDsn     string
	splunkToken   string
	splunkIndex   string

	// boolean which forces to send all logs on error or none at all
	sendLogs bool

	// How big can be batch of messages
	postMessagesBatchSize int
}

var SplunkClient *Splunk

func Initialize(correlationID, dsn, token, index string, sendLogs bool) error {
	log.Entry().Debugf("Initializing Splunk with DSN %v", dsn)

	client := piperhttp.Client{}

	client.SetOptions(piperhttp.ClientOptions{
		MaxRequestDuration:        5 * time.Second,
		Token:                     token,
		TransportSkipVerification: true,
	})

	SplunkClient = &Splunk{
		splunkClient:          client,
		splunkDsn:             dsn,
		splunkToken:           token,
		splunkIndex:           index,
		correlationID:         correlationID,
		postMessagesBatchSize: 1000,
		sendLogs:              sendLogs,
	}
	return nil
}

func Send(customTelemetryData *telemetry.CustomData, logCollector *log.CollectorHook) error {
	// Sends telemetry and or additionally logging data to Splunk
	telemetryData := prepareTelemetry(*customTelemetryData)
	messagesLen := len(logCollector.Messages)
	// TODO: Logic for errorCategory (undefined, service, infrastructure)
	if telemetryData.ErrorCode == "0" {
		// Successful run, we only send the telemetry data, no logging information
		//fmt.Println("Successful step, sending only telemetry data!")
		err := tryPostMessages(telemetryData, []log.Message{})
		if err != nil {
			log.Entry().WithError(err).WithField("module", "logger/splunk").Warn("Error while sending logs")
			//logrus.Error(fmt.Errorf("Failed to send telemetry data '%s'", string(json.Marshal(telemetryData))))
		}
		return nil
	} else if telemetryData.ErrorCode == "1" && !SplunkClient.sendLogs {
		// Failure run and we do not want to send the logs
		err := tryPostMessages(telemetryData, []log.Message{})
		if err != nil {
			log.Entry().WithError(err).WithField("module", "logger/splunk").Warn("Error while sending logs")
		}
		return nil
	} else {
		// ErrorCode indicates an error in the step, so we want to send all the logs with telemetry
		//fmt.Println("Logged Messages:", messagesLen, ", Max batch size:", SplunkClient.postMessagesBatchSize, ", Sending in ", int(math.Ceil(float64(messagesLen)/float64(SplunkClient.postMessagesBatchSize))), "batches.")
		for i := 0; i < messagesLen; i += SplunkClient.postMessagesBatchSize {
			upperBound := i + SplunkClient.postMessagesBatchSize
			if upperBound > messagesLen {
				upperBound = messagesLen
			}
			err := tryPostMessages(telemetryData, logCollector.Messages[i:upperBound])
			if err != nil {
				log.Entry().WithError(err).WithField("module", "logger/splunk").Warn("Error while sending logs")
			}
		}
	}
	return nil
}

// Data definition for monitoring
type MonitoringData struct {
	PipelineUrlHash string `json:"PipelineUrlHash,omitempty"`
	BuildUrlHash    string `json:"BuildUrlHash,omitempty"`
	StageName       string `json:"StageName"`
	StepName        string `json:"StepName,omitempty"`
	ExitCode        string `json:"ExitCode,omitempty"`
	Duration        string `json:"Duration,omitempty"`
	ErrorCode       string `json:"ErrorCode,omitempty"`
	ErrorCategory   string `json:"ErrorCategory,omitempty"`
	CorrelationID   string `json:"CorrelationID"`
	CommitHash      string `json:"CommitHash,omitempty"`
	Branch          string `json:"Branch,omitempty"`
}

func prepareTelemetry(customTelemetryData telemetry.CustomData) MonitoringData {
	tData := telemetry.GetData(&customTelemetryData)

	// TODO: Dependent on a groovy step, which creates the folder.
	// TODO: go git step here to read infos from git folder later add this to telemetry package?
	content, err := ioutil.ReadFile(".pipeline/commonPipelineEnvironment/git/commitId")
	if err != nil {
		logrus.Warning("Could not read commitId file.", err)
		content = []byte("N/A")
	}
	commitHash := string(content)

	content_branch, err := ioutil.ReadFile(".pipeline/commonPipelineEnvironment/git/branch")
	if err != nil {
		logrus.Warning("Could not read branch file.", err)
		content_branch = []byte("N/A")
	}
	branch := string(content_branch)

	// TODO: Return different telemetry information also for DORA Metrics.
	return MonitoringData{
		PipelineUrlHash: tData.PipelineURLHash,
		BuildUrlHash:    tData.BuildURLHash,
		StageName:       tData.StageName,
		StepName:        tData.BaseData.StepName,
		ExitCode:        tData.CustomData.ErrorCode,
		Duration:        tData.CustomData.Duration,
		ErrorCode:       tData.CustomData.ErrorCode,
		ErrorCategory:   tData.CustomData.ErrorCategory,
		CorrelationID:   SplunkClient.correlationID,
		CommitHash:      commitHash,
		Branch:          branch,
	}
}

type Event struct {
	Messages  []log.Message  `json:"messages,omitempty"`  // messages
	Telemetry MonitoringData `json:"telemetry,omitempty"` // telemetryData
}
type Details struct {
	Host       string `json:"host"`                 // hostname
	Source     string `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Event      Event  `json:"event,omitempty"`      // throw any useful key/val pairs here}
}

func tryPostMessages(telemetryData MonitoringData, messages []log.Message) error {

	event := Event{
		Messages:  messages,
		Telemetry: telemetryData,
	}
	details := Details{
		Host:       SplunkClient.correlationID,
		SourceType: "_json",
		Index:      SplunkClient.splunkIndex,
		Event:      event,
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}

	resp, err := SplunkClient.splunkClient.SendRequest(http.MethodPost, SplunkClient.splunkDsn, bytes.NewBuffer(payload), nil, nil)

	if err != nil {
		return err
	}
	defer func() {
		resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		rdr := io.LimitReader(resp.Body, 1000)
		body, err := ioutil.ReadAll(rdr)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s: Splunk logging failed - %s", resp.Status, string(body))
	}
	return nil
}
