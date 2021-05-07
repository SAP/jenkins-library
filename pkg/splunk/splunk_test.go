package splunk

import (
	"encoding/json"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/jarcoal/httpmock"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestInitialize(t *testing.T) {
	type args struct {
		correlationID string
		dsn           string
		token         string
		index         string
		sendLogs      bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Testing initialize splunk",
			args{
				correlationID: "correlationID",
				dsn:           "https://splunkURL.sap/services/collector",
				token:         "SECRET-TOKEN",
				index:         "test-index",
				sendLogs:      false,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Initialize(tt.args.correlationID, tt.args.dsn, tt.args.token, tt.args.index, tt.args.sendLogs); (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestSend(t *testing.T) {

	type args struct {
		customTelemetryData *telemetry.CustomData
		logCollector        *log.CollectorHook
		sendLogs            bool
		maxBatchSize        int
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		payloadLength int
		logLength     int // length of log per payload
	}{
		{name: "Testing Success Step - Send Telemetry Only",
			args: args{
				customTelemetryData: &telemetry.CustomData{
					Duration:      "100",
					ErrorCode:     "0",
					ErrorCategory: "DEBUG",
				},
				logCollector: &log.CollectorHook{CorrelationID: "DEBUG",
					Messages: []log.Message{
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 0",
						},
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 1",
						},
					}},
				sendLogs: false,
			},
			wantErr:       false,
			payloadLength: 1,
			logLength:     0,
		},
		{name: "Testing Success Step - Send Telemetry Only Although sendLogs Active",
			args: args{
				customTelemetryData: &telemetry.CustomData{
					Duration:      "100",
					ErrorCode:     "0",
					ErrorCategory: "DEBUG",
				},
				logCollector: &log.CollectorHook{CorrelationID: "DEBUG",
					Messages: []log.Message{
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 0",
						},
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 1",
						},
					}},
				sendLogs: true,
			},
			wantErr:       false,
			payloadLength: 1,
			logLength:     0,
		},
		{name: "Testing Failure Step - Send Telemetry Only",
			args: args{
				customTelemetryData: &telemetry.CustomData{
					Duration:  "100",
					ErrorCode: "1",
				},
				logCollector: &log.CollectorHook{CorrelationID: "DEBUG",
					Messages: []log.Message{
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 0",
						},
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 1",
						},
					}},
				sendLogs:     false,
				maxBatchSize: 1000,
			},
			wantErr:       false,
			payloadLength: 1,
			logLength:     0,
		},
		{name: "Testing Failure Step - Send Telemetry and Logs",
			args: args{
				customTelemetryData: &telemetry.CustomData{
					Duration:  "100",
					ErrorCode: "1",
				},
				logCollector: &log.CollectorHook{CorrelationID: "DEBUG",
					Messages: []log.Message{
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 0",
						},
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 1",
						},
					}},
				sendLogs:     true,
				maxBatchSize: 1000,
			},
			wantErr:       false,
			payloadLength: 1,
			logLength:     2,
		},
		{name: "Testing len(maxBatchSize)==len(logMessages)",
			args: args{
				customTelemetryData: &telemetry.CustomData{
					Duration:  "100",
					ErrorCode: "1",
				},
				logCollector: &log.CollectorHook{CorrelationID: "DEBUG",
					Messages: []log.Message{
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 0",
						},
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 1",
						},
					}},
				sendLogs:     true,
				maxBatchSize: 2,
			},
			wantErr:       false,
			payloadLength: 1,
			logLength:     2,
		},
		{name: "Testing len(maxBatchSize)<len(logMessages)",
			args: args{
				customTelemetryData: &telemetry.CustomData{
					Duration:  "100",
					ErrorCode: "1",
				},
				logCollector: &log.CollectorHook{CorrelationID: "DEBUG",
					Messages: []log.Message{
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 0",
						},
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 1",
						},
					}},
				sendLogs:     true,
				maxBatchSize: 1,
			},
			wantErr:       false,
			payloadLength: 2,
			logLength:     1, // equal to maxBatchSize
		},
		{name: "Testing len(maxBatchSize)>len(logMessages)",
			args: args{
				customTelemetryData: &telemetry.CustomData{
					Duration:  "100",
					ErrorCode: "1",
				},
				logCollector: &log.CollectorHook{CorrelationID: "DEBUG",
					Messages: []log.Message{
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 0",
						},
						{
							Time:    time.Time{},
							Level:   0,
							Message: "DEBUG",
							Data:    "DEBUG 1",
						},
					}},
				sendLogs:     true,
				maxBatchSize: 1000,
			},
			wantErr:       false,
			payloadLength: 1,
			logLength:     2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			fakeUrl := "https://splunk.example.com/services/collector"
			// Our database of received payloads
			var payloads []Details
			httpmock.RegisterResponder("POST", fakeUrl,
				func(req *http.Request) (*http.Response, error) {
					splunkMessage := Details{}
					if err := json.NewDecoder(req.Body).Decode(&splunkMessage); err != nil {
						return httpmock.NewStringResponse(400, ""), nil
					}

					defer req.Body.Close()
					payloads = append(payloads, splunkMessage)

					resp, err := httpmock.NewJsonResponse(200, splunkMessage)
					if err != nil {
						return httpmock.NewStringResponse(500, ""), nil
					}
					return resp, nil
				},
			)

			client := piperhttp.Client{}
			client.SetOptions(piperhttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Token:                     "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true,
			})

			SplunkClient = &Splunk{
				splunkClient:          client,
				splunkDsn:             fakeUrl,
				splunkIndex:           "index",
				correlationID:         "DEBUG",
				postMessagesBatchSize: tt.args.maxBatchSize,
				sendLogs:              tt.args.sendLogs,
			}
			if err := Send(tt.args.customTelemetryData, tt.args.logCollector); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(payloads) != tt.payloadLength {
				t.Errorf("Send() error, wanted %v payloads, got %v.", tt.payloadLength, len(payloads))
			}

			// The case if more than one payload is present is covered in the if statement above.
			if len(payloads[0].Event.Messages) != tt.logLength {
				t.Errorf("Send() error, wanted %v event messages, got %v.", tt.logLength, len(payloads[0].Event.Messages))
			}
			SplunkClient = nil
		})
	}
}

func Test_prepareTelemetry(t *testing.T) {
	type args struct {
		customTelemetryData telemetry.CustomData
	}
	tests := []struct {
		name string
		args args
		want MonitoringData
	}{
		{name: "Testing prepare telemetry information",
			args: args{
				customTelemetryData: telemetry.CustomData{
					Duration:      "1234",
					ErrorCode:     "0",
					ErrorCategory: "Undefined",
				},
			},
			want: MonitoringData{
				PipelineUrlHash: "",
				BuildUrlHash:    "",
				StageName:       "",
				StepName:        "",
				ExitCode:        "0",
				Duration:        "1234",
				ErrorCode:       "0",
				ErrorCategory:   "Undefined",
				CorrelationID:   "Correlation-Test",
				CommitHash:      "N/A",
				Branch:          "N/A",
				GitOwner:        "N/A",
				GitRepository:   "N/A",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize("Correlation-Test", "splunkUrl", "TOKEN", "index", false)
			if err != nil {
				t.Errorf("Error Initalizing Splunk. %v", err)
			}
			if got := prepareTelemetry(tt.args.customTelemetryData); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareTelemetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tryPostMessages(t *testing.T) {
	type args struct {
		telemetryData MonitoringData
		messages      []log.Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test HTTP Success",
			args: args{
				telemetryData: MonitoringData{
					PipelineUrlHash: "1234",
					BuildUrlHash:    "5678",
					StageName:       "deploy",
					StepName:        "cloudFoundryDeploy",
					ExitCode:        "0",
					Duration:        "12345678",
					ErrorCode:       "0",
					ErrorCategory:   "undefined",
					CorrelationID:   "123",
					CommitHash:      "a6bc",
					Branch:          "prod",
					GitOwner:        "N/A",
					GitRepository:   "N/A",
				},
				messages: []log.Message{},
			},
			wantErr: false,
		},
		{
			name: "Test HTTP Failure",
			args: args{
				telemetryData: MonitoringData{
					PipelineUrlHash: "1234",
					BuildUrlHash:    "5678",
					StageName:       "deploy",
					StepName:        "cloudFoundryDeploy",
					ExitCode:        "0",
					Duration:        "12345678",
					ErrorCode:       "0",
					ErrorCategory:   "undefined",
					CorrelationID:   "123",
					CommitHash:      "a6bc",
					Branch:          "prod",
					GitOwner:        "N/A",
					GitRepository:   "N/A",
				},
				messages: []log.Message{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			fakeUrl := "https://splunk.example.com/services/collector"
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", fakeUrl,
				func(req *http.Request) (*http.Response, error) {
					if tt.wantErr == true {
						return &http.Response{
							Status:           "400",
							StatusCode:       400,
							Proto:            "",
							ProtoMajor:       0,
							ProtoMinor:       0,
							Header:           nil,
							Body:             nil,
							ContentLength:    0,
							TransferEncoding: nil,
							Close:            false,
							Uncompressed:     false,
							Trailer:          nil,
							Request:          req,
							TLS:              nil,
						}, nil
					}
					return httpmock.NewStringResponse(200, ""), nil
				},
			)
			client := piperhttp.Client{}
			client.SetOptions(piperhttp.ClientOptions{
				MaxRequestDuration:        5 * time.Second,
				Token:                     "TOKEN",
				TransportSkipVerification: true,
				UseDefaultTransport:       true,
			})
			SplunkClient = &Splunk{
				splunkClient:          client,
				splunkDsn:             fakeUrl,
				splunkIndex:           "index",
				correlationID:         "DEBUG",
				postMessagesBatchSize: 1000,
			}
			if err := tryPostMessages(tt.args.telemetryData, tt.args.messages); (err != nil) != tt.wantErr {
				t.Errorf("tryPostMessages() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_readPipelineEnvironment(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		createFile bool
	}{
		{
			name:       "Test read pipelineEnvironment files not available",
			result:     "N/A",
			createFile: false,
		},
		{
			name:       "Test read pipelineEnvironment files available",
			result:     "master",
			createFile: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.createFile {

				// creating temporarily folders
				path := ".pipeline/commonPipelineEnvironment/"
				err := os.MkdirAll(path, os.ModePerm)
				if err != nil {
					t.Errorf("Could not create .pipeline/ folders: %v", err)
				}

				err = os.Mkdir(path+"git/", os.ModePerm)
				if err != nil {
					t.Errorf("Could not create git folder: %v", err)
				}

				// creating temporarily files with dummy content
				branch := []byte("master")
				err = ioutil.WriteFile(path+"git/branch", branch, 0644)
				if err != nil {
					t.Errorf("Could not create branch file: %v", err)
				}

			}
			result := readCommonPipelineEnvironment("git/branch")
			if result != tt.result {
				t.Errorf("readCommonPipelineEnvironment() got = %v, want %v", result, tt.result)
			}

			if tt.createFile {
				// deletes temp files
				err := os.RemoveAll(".pipeline")
				if err != nil {
					t.Errorf("Could not delete .pipeline folder: %v", err)
				}

			}
		})
	}
}
