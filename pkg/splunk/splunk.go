package splunk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

// Splunk SplunkHook provides a logrus hook which enables error logging to splunk platform.
// This is helpful in order to provide better monitoring and alerting on errors
// as well as the given error details can help to find the root cause of bugs.
type Splunk struct {
	levels        []logrus.Level
	tags          map[string]string
	splunkClient  piperhttp.Client
	correlationID string
	hostName      string
	splunkDsn     string
	splunkIndex   string

	// boolean which forces to send all logs on error or none at all
	sendLogs bool

	// How large a batch of messages can be
	postMessagesBatchSize int
}

func (s *Splunk) Initialize(correlationID, dsn, token, index string, sendLogs bool) error {
	log.Entry().Debugf("Initializing Splunk with DSN %v", dsn)

	if !strings.HasPrefix(token, "Splunk ") {
		token = "Splunk " + token
	}

	log.RegisterSecret(token)
	client := piperhttp.Client{}

	client.SetOptions(piperhttp.ClientOptions{
		MaxRequestDuration:        10 * time.Second,
		Token:                     token,
		TransportSkipVerification: true,
		MaxRetries:                1,
	})

	hostName, err := os.Hostname()
	if err != nil {
		log.Entry().WithError(err).Debug("Could not get hostName.")
		hostName = "n/a"
	}
	s.hostName = hostName
	s.splunkClient = client
	s.splunkDsn = dsn
	s.splunkIndex = index
	s.correlationID = correlationID
	s.postMessagesBatchSize = 5000
	s.sendLogs = sendLogs

	return nil
}

func (s *Splunk) Send(telemetryData telemetry.Data, logCollector *log.CollectorHook) error {
	// Sends telemetry and or additionally logging data to Splunk
	preparedTelemetryData := s.prepareTelemetry(telemetryData)
	messagesLen := len(logCollector.Messages)
	// TODO: Logic for errorCategory (undefined, service, infrastructure)
	if telemetryData.ErrorCode == "0" || (telemetryData.ErrorCode == "1" && !s.sendLogs) {
		// Either Successful run, we only send the telemetry data, no logging information
		// OR Failure run, and we do not want to send the logs
		err := s.tryPostMessages(preparedTelemetryData, []log.Message{})
		if err != nil {
			return errors.Wrap(err, "error while sending logs")
		}
		return nil
	} else {
		// ErrorCode indicates an error in the step, so we want to send all the logs with telemetry
		for i := 0; i < messagesLen; i += s.postMessagesBatchSize {
			upperBound := i + s.postMessagesBatchSize
			if upperBound > messagesLen {
				upperBound = messagesLen
			}
			err := s.tryPostMessages(preparedTelemetryData, logCollector.Messages[i:upperBound])
			if err != nil {
				return errors.Wrap(err, "error while sending logs")
			}
		}
	}
	return nil
}

func readCommonPipelineEnvironment(filePath string) string {

	// TODO: Dependent on a groovy step, which creates the folder.
	contentFile, err := os.ReadFile(".pipeline/commonPipelineEnvironment/" + filePath)
	if err != nil {
		log.Entry().Debugf("Could not read %v file. %v", filePath, err)
		contentFile = []byte("N/A")
	}
	return string(contentFile)
}

func (s *Splunk) prepareTelemetry(telemetryData telemetry.Data) MonitoringData {
	var errorMessage string

	if telemetryData.CustomData.ErrorCode != "0" {
		if fatalErrorDetail := log.GetFatalErrorDetail(); fatalErrorDetail != nil {
			var errorDetail map[string]any
			if err := json.Unmarshal(fatalErrorDetail, &errorDetail); err == nil {
				var parts []string

				if messageVal, exists := errorDetail["message"]; exists && messageVal != nil {
					parts = append(parts, fmt.Sprintf("%v", messageVal))
				}

				if errorVal, exists := errorDetail["error"]; exists && errorVal != nil {
					parts = append(parts, fmt.Sprintf("%v", errorVal))
				}

				if len(parts) > 0 {
					errorMessage = strings.Join(parts, " ")
				}
			}
		}
	}

	monitoringData := MonitoringData{
		PipelineUrlHash: telemetryData.PipelineURLHash,
		BuildUrlHash:    telemetryData.BuildURLHash,
		Orchestrator:    telemetryData.Orchestrator,
		TemplateName:    telemetryData.TemplateName,
		PiperCommitHash: telemetryData.PiperCommitHash,
		StageName:       telemetryData.StageName,
		StepName:        telemetryData.BaseData.StepName,
		ExitCode:        telemetryData.CustomData.ErrorCode,
		Duration:        telemetryData.CustomData.Duration,
		ErrorCode:       telemetryData.CustomData.ErrorCode,
		ErrorCategory:   telemetryData.CustomData.ErrorCategory,
		ErrorMessage:    errorMessage,
		CorrelationID:   s.correlationID,
		CommitHash:      readCommonPipelineEnvironment("git/headCommitId"),
		Branch:          readCommonPipelineEnvironment("git/branch"),
		GitOwner:        readCommonPipelineEnvironment("git/organization"),
		GitRepository:   readCommonPipelineEnvironment("git/repository"),
	}
	monitoringJson, err := json.Marshal(monitoringData)
	if err != nil {
		log.Entry().Error("could not marshal monitoring data")
		log.Entry().Debugf("Step monitoring data: {n/a}")
	} else {
		// log step monitoring data, changes here need to change the regex in the internal piper lib
		log.Entry().Debugf("Step monitoring data:%v", string(monitoringJson))
	}

	return monitoringData
}

func (s *Splunk) SendPipelineStatus(pipelineTelemetryData map[string]interface{}, logFile *[]byte) error {
	// Sends telemetry and or additionally logging data to Splunk

	readLogFile := string(*logFile)
	splitted := strings.Split(readLogFile, "\n")
	messagesLen := len(splitted)

	log.Entry().Debugf("Sending pipeline telemetry data to Splunk: %v", pipelineTelemetryData)
	s.postTelemetry(pipelineTelemetryData)

	if s.sendLogs {
		log.Entry().Debugf("Sending %v messages to Splunk.", messagesLen)
		for i := 0; i < messagesLen; i += s.postMessagesBatchSize {
			upperBound := i + s.postMessagesBatchSize
			if upperBound > messagesLen {
				upperBound = messagesLen
			}
			err := s.postLogFile(pipelineTelemetryData, splitted[i:upperBound])
			if err != nil {
				return errors.Wrap(err, "error while sending logs")
			}
		}
	}
	return nil
}

func (s *Splunk) postTelemetry(telemetryData map[string]interface{}) error {
	if telemetryData == nil {
		telemetryData = map[string]interface{}{"Empty": "No telemetry available."}
	}
	details := DetailsTelemetry{
		Host:       s.hostName,
		SourceType: "piper:pipeline:telemetry",
		Index:      s.splunkIndex,
		Event:      telemetryData,
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return errors.Wrap(err, "error while marshalling Splunk message details")
	}
	prettyPayload, err := json.MarshalIndent(details, "", "    ")
	if err != nil {
		log.Entry().WithError(err).Warn("Failed to generate pretty payload json")
		prettyPayload = nil
	}
	log.Entry().Debugf("Sending the follwing payload to Splunk HEC: %s", string(prettyPayload))

	if err != nil {
		return errors.Wrap(err, "error while marshalling Splunk message details")
	}

	resp, err := s.splunkClient.SendRequest(http.MethodPost, s.splunkDsn, bytes.NewBuffer(payload), nil, nil)

	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			// log it to stdout
			rdr := io.LimitReader(resp.Body, 1000)
			body, errRead := io.ReadAll(rdr)
			log.Entry().Infof("%v: Splunk logging failed - %v", resp.Status, string(body))
			if errRead != nil {
				return errors.Wrap(errRead, "Error reading response body from Splunk.")
			}
			return errors.Wrapf(err, "%v: Splunk logging failed - %v", resp.Status, string(body))
		}
	}

	if err != nil {
		return errors.Wrap(err, "error sending the requests to Splunk")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			errors.Wrap(err, "closing response body failed")
		}
	}()

	return nil
}

func (s *Splunk) postLogFile(telemetryData map[string]interface{}, messages []string) error {

	var logfileEvents []string
	for _, message := range messages {
		logMessage := LogFileEvent{
			Event:      message,
			Host:       s.hostName,
			Source:     s.correlationID,
			SourceType: "piper:pipeline:logfile",
			Index:      s.splunkIndex,
		}
		marshalledLogMessage, err := json.Marshal(logMessage)
		if err != nil {
			return errors.Wrap(err, "error while marshalling Splunk messages")
		}
		logfileEvents = append(logfileEvents, string(marshalledLogMessage))
	}
	// creates payload {"event":"this is a sample event ", "Host":"myHost", "Source":"mySource", "SourceType":"valueA", "Index":"valueB"}{"event":"this is a sample event ", "Host":"myHost", "Source":"mySource", "SourceType":"valueA", "Index":"valueB"}..
	strout := strings.Join(logfileEvents, ",")
	payload := strings.NewReader(strout)

	resp, err := s.splunkClient.SendRequest(http.MethodPost, s.splunkDsn, payload, nil, nil)

	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			// log it to stdout
			rdr := io.LimitReader(resp.Body, 1000)
			body, errRead := io.ReadAll(rdr)
			log.Entry().Infof("%v: Splunk logging failed - %v", resp.Status, string(body))
			if errRead != nil {
				return errors.Wrap(errRead, "Error reading response body from Splunk.")
			}
			return errors.Wrapf(err, "%v: Splunk logging failed - %v", resp.Status, string(body))
		}
	}

	if err != nil {
		return errors.Wrap(err, "error sending the requests to Splunk")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			errors.Wrap(err, "closing response body failed")
		}
	}()

	return nil
}

func (s *Splunk) tryPostMessages(telemetryData MonitoringData, messages []log.Message) error {

	event := Event{
		Messages:  messages,
		Telemetry: telemetryData,
	}
	details := Details{
		Host:       s.hostName,
		SourceType: "_json",
		Index:      s.splunkIndex,
		Event:      event,
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return errors.Wrap(err, "error while marshalling Splunk message details")
	}

	resp, err := s.splunkClient.SendRequest(http.MethodPost, s.splunkDsn, bytes.NewBuffer(payload), nil, nil)

	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			// log it to stdout
			rdr := io.LimitReader(resp.Body, 1000)
			body, errRead := io.ReadAll(rdr)
			log.Entry().Infof("%v: Splunk logging failed - %v", resp.Status, string(body))
			if errRead != nil {
				return errors.Wrap(errRead, "Error reading response body from Splunk.")
			}
			return errors.Wrapf(err, "%v: Splunk logging failed - %v", resp.Status, string(body))
		}
	}

	if err != nil {
		return errors.Wrap(err, "error sending the requests to Splunk")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			errors.Wrap(err, "closing response body failed")
		}
	}()

	return nil
}
