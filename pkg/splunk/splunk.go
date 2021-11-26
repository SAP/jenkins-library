package splunk

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
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
	splunkDsn     string
	splunkIndex   string

	// boolean which forces to send all logs on error or none at all
	sendLogs bool

	// How big can be batch of messages
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
		MaxRequestDuration:        5 * time.Second,
		Token:                     token,
		TransportSkipVerification: true,
		MaxRetries:                -1,
	})

	s.splunkClient = client
	s.splunkDsn = dsn
	s.splunkIndex = index
	s.correlationID = correlationID
	s.postMessagesBatchSize = 20000
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
	contentFile, err := ioutil.ReadFile(".pipeline/commonPipelineEnvironment/" + filePath)
	if err != nil {
		log.Entry().Warnf("Could not read %v file. %v", filePath, err)
		contentFile = []byte("N/A")
	}
	return string(contentFile)
}

func (s *Splunk) prepareTelemetry(telemetryData telemetry.Data) MonitoringData {

	monitoringData := MonitoringData{
		PipelineUrlHash: telemetryData.PipelineURLHash,
		BuildUrlHash:    telemetryData.BuildURLHash,
		StageName:       telemetryData.StageName,
		StepName:        telemetryData.BaseData.StepName,
		ExitCode:        telemetryData.CustomData.ErrorCode,
		Duration:        telemetryData.CustomData.Duration,
		ErrorCode:       telemetryData.CustomData.ErrorCode,
		ErrorCategory:   telemetryData.CustomData.ErrorCategory,
		CorrelationID:   s.correlationID,
		CommitHash:      readCommonPipelineEnvironment("git/headCommitId"),
		Branch:          readCommonPipelineEnvironment("git/branch"),
		GitOwner:        readCommonPipelineEnvironment("github/owner"),
		GitRepository:   readCommonPipelineEnvironment("github/repository"),
	}
	monitoringJson, err := json.Marshal(monitoringData)
	if err != nil {
		log.Entry().Error("could not marshal monitoring data")
		log.Entry().Infof("Step monitoring data: {n/a}")
	} else {
		// log step monitoring data, changes here need to change the regex in the internal piper lib
		log.Entry().Infof("Step monitoring data:%v", string(monitoringJson))
	}

	return monitoringData
}

func (s *Splunk) SendPipelineStatus(pipelineTelemetryData map[string]interface{}, logFile *[]byte) error {
	// Sends telemetry and or additionally logging data to Splunk

	readLogFile := string(*logFile)
	splitted := strings.Split(readLogFile, "\n")
	messagesLen := len(splitted)

	log.Entry().Debugf("Sending %v messages to Splunk.", messagesLen)
	log.Entry().Debugf("Sending pipeline telemetry data to Splunk: %v", pipelineTelemetryData)
	s.postTelemetry(pipelineTelemetryData)

	if s.sendLogs {
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
		Host:       s.correlationID,
		SourceType: "_json",
		Index:      s.splunkIndex,
		Event:      telemetryData,
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return errors.Wrap(err, "error while marshalling Splunk message details")
	}
	log.Entry().Debugf("Sending the follwing payload to Splunk HEC: %v", string(payload))
	resp, err := s.splunkClient.SendRequest(http.MethodPost, s.splunkDsn, bytes.NewBuffer(payload), nil, nil)

	if resp != nil {
		if resp.StatusCode != http.StatusOK {
			// log it to stdout
			rdr := io.LimitReader(resp.Body, 1000)
			body, errRead := ioutil.ReadAll(rdr)
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

	event := LogFileEvents{
		Messages:  messages,
		Telemetry: telemetryData,
	}
	details := DetailsLog{
		Host:       s.correlationID,
		SourceType: "txt",
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
			body, errRead := ioutil.ReadAll(rdr)
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
		Host:       s.correlationID,
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
			body, errRead := ioutil.ReadAll(rdr)
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
