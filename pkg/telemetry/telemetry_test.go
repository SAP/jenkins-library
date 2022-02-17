package telemetry

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestTelemetry_Initialize(t *testing.T) {
	type fields struct {
		baseData             BaseData
		baseMetaData         BaseMetaData
		data                 Data
		provider             orchestrator.OrchestratorSpecificConfigProviding
		disabled             bool
		client               *piperhttp.Client
		CustomReportingDsn   string
		CustomReportingToken string
		customClient         *piperhttp.Client
		BaseURL              string
		Endpoint             string
		SiteID               string
	}
	type args struct {
		telemetryDisabled bool
		stepName          string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *piperhttp.Client
	}{
		{
			name:   "telemetry disabled",
			fields: fields{},
			args: args{
				telemetryDisabled: true,
				stepName:          "test",
			},
			want: nil,
		},
		{
			name:   "telemetry enabled",
			fields: fields{},
			args: args{
				telemetryDisabled: false,
				stepName:          "test",
			},
			want: &piperhttp.Client{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telemetryClient := &Telemetry{}
			telemetryClient.Initialize(tt.args.telemetryDisabled, tt.args.stepName)
			// assert
			assert.NotEqual(t, tt.want, telemetryClient.client)
			assert.Equal(t, tt.args.stepName, telemetryClient.baseData.StepName)
		})
	}
}

func TestTelemetry_Send(t *testing.T) {
	type fields struct {
		baseData             BaseData
		baseMetaData         BaseMetaData
		data                 Data
		provider             orchestrator.OrchestratorSpecificConfigProviding
		disabled             bool
		client               *piperhttp.Client
		CustomReportingDsn   string
		CustomReportingToken string
		BaseURL              string
		Endpoint             string
		SiteID               string
	}
	tests := []struct {
		name     string
		fields   fields
		swaCalls int
	}{
		{
			name: "Telemetry disabled, reporting disabled",
			fields: fields{
				disabled: true,
			},
			swaCalls: 0,
		},
		{
			name: "Telemetry enabled",
			fields: fields{
				disabled: false,
			},
			swaCalls: 1,
		},
		{
			name: "Telemetry disabled",
			fields: fields{
				disabled: true,
			},
			swaCalls: 0,
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Reset()
			telemetryClient := &Telemetry{disabled: tt.fields.disabled}
			telemetryClient.Initialize(tt.fields.disabled, tt.name)
			telemetryClient.CustomReportingDsn = tt.fields.CustomReportingDsn
			if telemetryClient.client == nil {
				telemetryClient.client = &piperhttp.Client{}
			}

			url := telemetryClient.BaseURL + telemetryClient.Endpoint

			telemetryClient.client.SetOptions(piperhttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Token:                     "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true,
				MaxRetries:                -1,
			})

			if tt.fields.CustomReportingDsn != "" {
				telemetryClient.customClient = &piperhttp.Client{}
				telemetryClient.customClient.SetOptions(piperhttp.ClientOptions{
					MaxRequestDuration:        5 * time.Second,
					Token:                     "TOKEN",
					TransportSkipVerification: true,
					UseDefaultTransport:       true, // Needed for mocking
					MaxRetries:                -1,
				})
			}

			httpmock.RegisterResponder(http.MethodGet, url,
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewStringResponse(200, "Ok"), nil
				},
			)
			httpmock.RegisterResponder(http.MethodPost, telemetryClient.CustomReportingDsn,
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewStringResponse(200, "Ok"), nil
				},
			)

			// test
			telemetryClient.SetData(&CustomData{})
			telemetryClient.Send()

			// assert
			info := httpmock.GetCallCountInfo()

			if got := info["GET "+url]; !assert.Equal(t, got, tt.swaCalls) {
				t.Errorf("Send() = swa calls %v, wanted %v", got, tt.swaCalls)
			}

		})
	}
	defer httpmock.DeactivateAndReset()
}

func TestSetData(t *testing.T) {
	type args struct {
		customData *CustomData
	}
	tests := []struct {
		name string
		args args
		want Data
	}{
		{
			name: "Test",
			args: args{customData: &CustomData{
				Duration:        "100",
				ErrorCode:       "0",
				ErrorCategory:   "Undefined",
				PiperCommitHash: "abcd12345",
			},
			},
			want: Data{
				BaseData: BaseData{
					URL:             "",
					ActionName:      "",
					EventType:       "",
					StepName:        "TestCreateDataObject",
					SiteID:          "",
					PipelineURLHash: "",
					BuildURLHash:    "",
					Orchestrator:    "Unknown",
				},
				BaseMetaData: BaseMetaData{
					StepNameLabel:        "stepName",
					StageNameLabel:       "stageName",
					PipelineURLHashLabel: "pipelineUrlHash",
					BuildURLHashLabel:    "buildUrlHash",
					DurationLabel:        "duration",
					ExitCodeLabel:        "exitCode",
					ErrorCategoryLabel:   "errorCategory",
					OrchestratorLabel:    "orchestrator",
					PiperCommitHashLabel: "piperCommitHash",
				},
				CustomData: CustomData{
					Duration:        "100",
					ErrorCode:       "0",
					ErrorCategory:   "Undefined",
					PiperCommitHash: "abcd12345",
					Custom1Label:    "",
					Custom2Label:    "",
					Custom3Label:    "",
					Custom4Label:    "",
					Custom5Label:    "",
					Custom1:         "",
					Custom2:         "",
					Custom3:         "",
					Custom4:         "",
					Custom5:         "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telemetryClient := Telemetry{}
			telemetryClient.Initialize(false, "TestCreateDataObject")
			telemetryClient.baseData = BaseData{
				URL:             "",
				ActionName:      "",
				EventType:       "",
				StepName:        "TestCreateDataObject",
				SiteID:          "",
				PipelineURLHash: "",
				BuildURLHash:    "",
				Orchestrator:    "Unknown",
			}
			telemetryClient.baseMetaData = baseMetaData
			telemetryClient.SetData(tt.args.customData)
			fmt.Println(telemetryClient.data)
			fmt.Println(tt.want)
			if !reflect.DeepEqual(telemetryClient.data, tt.want) {
				t.Errorf("CreateDataObject() t.data= %v, want %v", telemetryClient.data, tt.want)
			}
		})
	}
}

func TestTelemetry_logStepTelemetryData(t *testing.T) {

	os.Setenv("JENKINS_URL", "FOO BAR BAZ")
	os.Setenv("BUILD_URL", "jaas.com/foo/bar/main/42")
	os.Setenv("BRANCH_NAME", "main")
	os.Setenv("GIT_COMMIT", "abcdef42713")
	os.Setenv("GIT_URL", "github.com/foo/bar")

	provider, _ := orchestrator.NewOrchestratorSpecificConfigProvider()

	type fields struct {
		data     Data
		provider orchestrator.OrchestratorSpecificConfigProviding
	}
	tests := []struct {
		name       string
		fields     fields
		fatalError logrus.Fields
		logOutput  string // TODO
	}{
		{
			name: "logging with error, no fatalError set",
			fields: fields{
				data: Data{
					BaseData:     BaseData{},
					BaseMetaData: BaseMetaData{},
					CustomData:   CustomData{ErrorCode: "1"},
				},
				provider: provider,
			},
		},
		{
			name: "logging with error, fatal error set",
			fields: fields{
				data: Data{
					BaseData:     BaseData{},
					BaseMetaData: BaseMetaData{},
					CustomData:   CustomData{ErrorCode: "1"},
				},
				provider: provider,
			},
			fatalError: logrus.Fields{
				"message":       "Some error happened",
				"error":         "Oh snap!",
				"category":      "undefined",
				"result":        "failure",
				"correlationId": "test",
				"time":          "0000-00-00 00:00:00.000",
			},
		},
		{
			name: "logging without error",
			fields: fields{
				data: Data{
					CustomData: CustomData{ErrorCode: "0"},
				},
				provider: &orchestrator.JenkinsConfigProvider{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, hook := test.NewNullLogger()
			log.RegisterHook(hook)
			defer resetEnv(os.Environ())
			os.Clearenv()
			telemetry := &Telemetry{
				data:     tt.fields.data,
				provider: tt.fields.provider,
			}
			var expected string
			if tt.fatalError != nil {
				errDetails, _ := json.Marshal(&tt.fatalError)
				log.SetFatalErrorDetail(errDetails)
				expected = "Step telemetry data:{\"PipelineURLHash\":\"\",\"BuildURLHash\":\"\",\"StageName\":\"\",\"StepName\":\"\",\"ErrorCode\":\"" + tt.fields.data.ErrorCode + "\",\"Duration\":\"\",\"ErrorCategory\":\"\",\"CorrelationID\":\"n/a\",\"CommitHash\":\"n/a\",\"Branch\":\"n/a\",\"GitOwner\":\"n/a\",\"GitRepository\":\"n/a\",\"ErrorDetail\":" + string(errDetails) + "}"

			} else {
				expected = "Step telemetry data:{\"PipelineURLHash\":\"\",\"BuildURLHash\":\"\",\"StageName\":\"\",\"StepName\":\"\",\"ErrorCode\":\"" + tt.fields.data.ErrorCode + "\",\"Duration\":\"\",\"ErrorCategory\":\"\",\"CorrelationID\":\"n/a\",\"CommitHash\":\"n/a\",\"Branch\":\"n/a\",\"GitOwner\":\"n/a\",\"GitRepository\":\"n/a\",\"ErrorDetail\":null}"
			}
			telemetry.logStepTelemetryData()
			assert.Equal(t, expected, hook.LastEntry().Message)
			hook.Reset()
		})
	}
}

func resetEnv(e []string) {
	for _, val := range e {
		tmp := strings.Split(val, "=")
		os.Setenv(tmp[0], tmp[1])
	}
}
