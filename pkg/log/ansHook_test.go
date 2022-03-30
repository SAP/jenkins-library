package log

import (
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const testCorrelationID = "1234"

func TestANSHook_Levels(t *testing.T) {
	hook := NewANSHook("", "", "", "")
	assert.Equal(t, []logrus.Level{logrus.InfoLevel, logrus.DebugLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel},
		hook.Levels())
}

func TestNewANSHook(t *testing.T) {
	testClient := ans.ANS{
		XSUAA: xsuaa.XSUAA{
			OAuthURL:     "https://my.test.oauth.provider",
			ClientID:     "myTestClientID",
			ClientSecret: "super secret",
		},
		URL: "https://my.test.backend",
	}
	testServiceKeyJSON := `{
					"url": "https://my.test.backend",
					"client_id": "myTestClientID",
					"client_secret": "super secret",
					"oauth_url": "https://my.test.oauth.provider"
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
		want                     ANSHook
	}{
		{
			name: "Straight forward test",
			args: args{
				serviceKey:    testServiceKeyJSON,
				correlationID: testCorrelationID,
			},
			want: ANSHook{
				client: testClient,
				event:  defaultEvent(),
			},
		},
		{
			name: "No service key = no client",
			args: args{
				correlationID: testCorrelationID,
			},
			want: ANSHook{
				client: ans.ANS{},
				event:  defaultEvent(),
			},
		},
		{
			name: "With event template as file",
			args: args{
				serviceKey:    testServiceKeyJSON,
				correlationID: testCorrelationID,
			},
			eventTemplateFileContent: `{"priority":123}`,
			want: ANSHook{
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
			want: ANSHook{
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
			want: ANSHook{
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

			got := NewANSHook(tt.args.serviceKey, tt.args.correlationID, testEventTemplateFilePath, tt.args.eventTemplate)
			assert.Equal(t, tt.want, got, "new ANSHook not as expected")
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
					Level:   logrus.InfoLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "my log message",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
			wantEvent: ans.Event{
				EventType:      "Piper",
				EventTimestamp: time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC).Unix(),
				Severity:       "INFO",
				Category:       "NOTIFICATION",
				Subject:        "testStep",
				Body:           "my log message",
				Resource: &ans.Resource{
					ResourceType: "Piper",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "stepName": "testStep", "logLevel": "info"},
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
					ResourceType: "Piper",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "stepName": "testStep", "error": "an error occurred!", "logLevel": "error"},
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
					ResourceType: "Piper",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "stepName": "testStep", "logLevel": "fatal"},
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
					Level:   logrus.InfoLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "my log message",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
			wantEvent: ans.Event{
				EventType:      "My event type",
				EventTimestamp: time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC).Unix(),
				Severity:       "INFO",
				Category:       "NOTIFICATION",
				Subject:        "My subject line",
				Body:           "my log message",
				Resource: &ans.Resource{
					ResourceType: "Piper",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": testCorrelationID, "stepName": "testStep", "logLevel": "info", "Some": 1.0, "Additional": "a string", "Tags": true},
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
					ResourceType: "Piper",
					ResourceName: "Pipeline",
				},
				Tags: map[string]interface{}{"ans:correlationId": "1234", "stepName": "testStep", "logLevel": "warning"},
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
					Level:   logrus.InfoLevel,
					Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
					Message: "   ",
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
		Tags:      map[string]interface{}{"ans:correlationId": testCorrelationID},
		Resource: &ans.Resource{
			ResourceType: "Piper",
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
