package log

import (
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const testCorrelationID = "1234"

func TestANSHook_Levels(t *testing.T) {
	hook := NewANSHook("", "", "")
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
		name string
		args args
		want ANSHook
	}{
		{
			name: "Straight forward test",
			args: args{
				serviceKey:    testServiceKeyJSON,
				correlationID: testCorrelationID,
			},
			want: ANSHook{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         defaultEvent(),
			},
		},
		{
			name: "No service key = no client",
			args: args{
				correlationID: testCorrelationID,
			},
			want: ANSHook{
				correlationID: testCorrelationID,
				client:        ans.ANS{},
				event:         defaultEvent(),
			},
		},
		{
			name: "With event template",
			args: args{
				serviceKey:    testServiceKeyJSON,
				correlationID: testCorrelationID,
				eventTemplate: `{"priority":123}`,
			},
			want: ANSHook{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         mergeEvents(t, defaultEvent(), ans.Event{Priority: 123}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewANSHook(tt.args.serviceKey, tt.args.correlationID, tt.args.eventTemplate)
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
		entryArg  *logrus.Entry
		wantEvent ans.Event
	}{
		{
			name: "Straight forward test",
			fields: fields{
				correlationID: testCorrelationID,
				client:        testClient,
				event:         defaultEvent(),
			},
			entryArg: &logrus.Entry{
				Level:   logrus.InfoLevel,
				Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
				Message: "my log message",
				Data:    map[string]interface{}{"stepName": "testStep"},
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
			entryArg: &logrus.Entry{
				Level:   logrus.InfoLevel,
				Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
				Message: "my log message",
				Data:    map[string]interface{}{"stepName": "testStep", "error": "an error occurred!"},
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
			entryArg: &logrus.Entry{
				Level:   logrus.InfoLevel,
				Time:    time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC),
				Message: "my log message",
				Data:    map[string]interface{}{"stepName": "testStep"},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLogLevel := tt.entryArg.Level
			ansHook := &ANSHook{
				correlationID: tt.fields.correlationID,
				client:        tt.fields.client,
				event:         tt.fields.event,
			}
			defer func() { testEvent = ans.Event{} }()
			ansHook.Fire(tt.entryArg)
			assert.Equal(t, tt.wantEvent, testEvent, "Event is not as expected.")
			assert.Equal(t, originalLogLevel.String(), tt.entryArg.Level.String(), "Entry error level has been altered")
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
