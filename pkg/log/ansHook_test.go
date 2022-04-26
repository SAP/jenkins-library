package log

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const testCorrelationID = "1234"

func TestANSHook_Levels(t *testing.T) {
	hook, _ := NewANSHook(ans.Configuration{}, "")
	assert.Equal(t, []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel},
		hook.Levels())
}

func TestNewANSHook(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"access_token":"1234"}`))
	}))
	defer mockServer.Close()
	testClient := ans.ANS{
		XSUAA: xsuaa.XSUAA{
			OAuthURL:     mockServer.URL,
			ClientID:     "myTestClientID",
			ClientSecret: "super secret",
		},
		URL: mockServer.URL,
	}
	testServiceKeyJSON := `{
					"url": "` + mockServer.URL + `",
					"client_id": "myTestClientID",
					"client_secret": "super secret",
					"oauth_url": "` + mockServer.URL + `"
				}`
	type args struct {
		serviceKey    string
		correlationID string
		eventTemplate string
	}
	tests := []struct {
		name                     string
		args                     args
		eventTemplateFileContent string
		wantHook                 ANSHook
		wantErr                  bool
	}{
		{
			name: "Straight forward test",
			args: args{
				serviceKey:    testServiceKeyJSON,
				correlationID: testCorrelationID,
			},
			wantHook: ANSHook{
				client: testClient,
				event:  defaultEvent(),
			},
		},
		{
			name: "No service key yields error",
			args: args{
				correlationID: testCorrelationID,
			},
			wantErr: true,
		},
		{
			name: "With event template as file",
			args: args{
				serviceKey:    testServiceKeyJSON,
				correlationID: testCorrelationID,
			},
			eventTemplateFileContent: `{"priority":123}`,
			wantHook: ANSHook{
				client: testClient,
				event:  mergeEvents(t, defaultEvent(), ans.Event{Priority: 123}),
			},
		},
		{
			name: "With event template as string",
			args: args{
				serviceKey:    testServiceKeyJSON,
				correlationID: testCorrelationID,
				eventTemplate: `{"priority":123}`,
			},
			wantHook: ANSHook{
				client: testClient,
				event:  mergeEvents(t, defaultEvent(), ans.Event{Priority: 123}),
			},
		},
		{
			name: "With event template from two sources, string overwrites file",
			args: args{
				serviceKey:    testServiceKeyJSON,
				correlationID: testCorrelationID,
				eventTemplate: `{"priority":789}`,
			},
			eventTemplateFileContent: `{"priority":123}`,
			wantHook: ANSHook{
				client: testClient,
				event:  mergeEvents(t, defaultEvent(), ans.Event{Priority: 789}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testEventTemplateFilePath string
			if len(tt.eventTemplateFileContent) > 0 {
				var err error
				testEventTemplateFile, err := os.CreateTemp("", "event_template_*.json")
				require.NoError(t, err, "File creation failed!")
				defer testEventTemplateFile.Close()
				defer os.Remove(testEventTemplateFile.Name())
				data := []byte(tt.eventTemplateFileContent)
				_, err = testEventTemplateFile.Write(data)
				require.NoError(t, err, "Could not write test data to test file!")
				testEventTemplateFilePath = testEventTemplateFile.Name()
			}

			ansConfig := ans.Configuration{
				ServiceKey:            tt.args.serviceKey,
				EventTemplateFilePath: testEventTemplateFilePath,
				EventTemplate:         tt.args.eventTemplate,
			}
			got, err := NewANSHook(ansConfig, tt.args.correlationID)
			if tt.wantErr {
				assert.Error(t, err, "An error was expected here")
			} else {
				require.NoError(t, err, "No error was expected here")
				assert.Equal(t, tt.wantHook, got, "new ANSHook not as expected")
			}
		})
	}
}

func TestANSHook_Fire(t *testing.T) {
	testClient := ansMock{}
	type fields struct {
		correlationID string
		client        ansMock
		levels        []logrus.Level
		event         ans.Event
	}
	tests := []struct {
		name      string
		fields    fields
		entryArgs []*logrus.Entry
		wantEvent ans.Event
	}{
		{
			name: "Straight forward test",
			fields: fields{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         defaultEvent(),
			},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.WarnLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "my log message",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
			wantEvent: ans.Event{
				EventType:      "Piper",
				EventTimestamp: time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC).Unix(),
				Severity:       "WARNING",
				Category:       "ALERT",
				Subject:        "testStep",
				Body:           "my log message",
				Resource: &ans.Resource{
					ResourceType: "Pipeline",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "ans:sourceEventId": "1234", "stepName": "testStep", "logLevel": "warning"},
			},
		},
		{
			name: "If error key set in data, severity should be error",
			fields: fields{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         defaultEvent(),
			},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.InfoLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "my log message",
					Data:    map[string]interface{}{"stepName": "testStep", "error": "an error occurred!"},
				},
			},
			wantEvent: ans.Event{
				EventType:      "Piper",
				EventTimestamp: time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC).Unix(),
				Severity:       "ERROR",
				Category:       "EXCEPTION",
				Subject:        "testStep",
				Body:           "my log message",
				Resource: &ans.Resource{
					ResourceType: "Pipeline",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "ans:sourceEventId": "1234", "stepName": "testStep", "error": "an error occurred!", "logLevel": "error"},
			},
		},
		{
			name: "If message is fatal error, severity should be fatal",
			fields: fields{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         defaultEvent(),
			},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.InfoLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "fatal error: an error occurred",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
			wantEvent: ans.Event{
				EventType:      "Piper",
				EventTimestamp: time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC).Unix(),
				Severity:       "FATAL",
				Category:       "EXCEPTION",
				Subject:        "testStep",
				Body:           "fatal error: an error occurred",
				Resource: &ans.Resource{
					ResourceType: "Pipeline",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "ans:sourceEventId": "1234", "stepName": "testStep", "logLevel": "fatal"},
			},
		},
		{
			name: "Event already set",
			fields: fields{
				correlationID: testCorrelationID,
				client:        testClient,
				event: mergeEvents(t, defaultEvent(), ans.Event{
					EventType: "My event type",
					Subject:   "My subject line",
					Tags:      map[string]interface{}{"Some": 1.0, "Additional": "a string", "Tags": true},
				}),
			},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.WarnLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "my log message",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
			wantEvent: ans.Event{
				EventType:      "My event type",
				EventTimestamp: time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC).Unix(),
				Severity:       "WARNING",
				Category:       "ALERT",
				Subject:        "My subject line",
				Body:           "my log message",
				Resource: &ans.Resource{
					ResourceType: "Pipeline",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "ans:sourceEventId": "1234", "stepName": "testStep", "logLevel": "warning", "Some": 1.0, "Additional": "a string", "Tags": true},
			},
		},
		{
			name: "Log entries should not affect each other",
			fields: fields{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         defaultEvent(),
			},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.InfoLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "my log message",
					Data:    map[string]interface{}{"stepName": "testStep", "this entry": "should only be part of this event"},
				},
				{
					Level:   logrus.WarnLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 8, time.UTC),
					Message: "another message",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
			wantEvent: ans.Event{
				EventType:      "Piper",
				EventTimestamp: time.Date(2001, 2, 3, 4, 5, 6, 8, time.UTC).Unix(),
				Severity:       "WARNING",
				Category:       "ALERT",
				Subject:        "testStep",
				Body:           "another message",
				Resource: &ans.Resource{
					ResourceType: "Pipeline",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "ans:sourceEventId": "1234", "stepName": "testStep", "logLevel": "warning"},
			},
		},
		{
			name: "White space messages should not send",
			fields: fields{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         defaultEvent(),
			},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.ErrorLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "   ",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
		},
		{
			name: "INFO severity should not be sent",
			fields: fields{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         defaultEvent(),
			},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.InfoLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "this is not an error",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ansHook := &ANSHook{
				client: tt.fields.client,
				event:  tt.fields.event,
			}
			defer func() { testEvent = ans.Event{} }()
			for _, entryArg := range tt.entryArgs {
				originalLogLevel := entryArg.Level
				ansHook.Fire(entryArg)
				assert.Equal(t, originalLogLevel.String(), entryArg.Level.String(), "Entry error level has been altered")
			}
			assert.Equal(t, tt.wantEvent, testEvent, "Event is not as expected.")
		})
	}
}

func defaultEvent() ans.Event {
	return ans.Event{
		EventType: "Piper",
		Tags:      map[string]interface{}{"ans:correlationId": testCorrelationID, "ans:sourceEventId": testCorrelationID},
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
	}
}

func mergeEvents(t *testing.T, event1, event2 ans.Event) ans.Event {
	event2JSON, err := json.Marshal(event2)
	require.NoError(t, err)
	err = event1.MergeWithJSON(event2JSON)
	require.NoError(t, err)
	return event1
}

type ansMock struct{}

var testEvent ans.Event

func (ans ansMock) Send(event ans.Event) error {
	testEvent = event
	return nil
}

func (ans ansMock) CheckCorrectSetup() error {
	return fmt.Errorf("not implemented")
}
