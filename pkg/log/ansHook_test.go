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

func TestANSHook_Levels(t *testing.T) {
	hook, _ := newANSHook(ans.Configuration{}, "", &ansMock{})
	assert.Equal(t, []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel},
		hook.Levels())
}

func TestANSHook_newANSHook(t *testing.T) {
	t.Parallel()
	testClient := &ans.ANS{
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
			t.Parallel()
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
			clientMock := &ansMock{}
			defer clientMock.cleanup()
			got, err := newANSHook(ansConfig, tt.args.correlationID, clientMock)

			if tt.wantErr {
				assert.Error(t, err, "An error was expected here")
			} else {
				require.NoError(t, err, "No error was expected here")
				assert.Equal(t, tt.wantHook.client, clientMock.testANS, "new ANSHook not as expected")
				assert.Equal(t, tt.wantHook.event, got.event, "new ANSHook not as expected")
			}
		})
	}
}

func TestANSHook_Fire(t *testing.T) {
	t.Parallel()
	type fields struct {
		levels []logrus.Level
		event  ans.Event
		firing bool
	}
	tests := []struct {
		name      string
		fields    fields
		entryArgs []*logrus.Entry
		wantEvent ans.Event
	}{
		{
			name:      "Straight forward test",
			fields:    fields{event: defaultEvent()},
			entryArgs: []*logrus.Entry{defaultLogrusEntry()},
			wantEvent: defaultResultingEvent(),
		},
		{
			name: "Event already set",
			fields: fields{
				event: mergeEvents(t, defaultEvent(), ans.Event{
					EventType: "My event type",
					Subject:   "My subject line",
					Tags:      map[string]interface{}{"Some": 1.0, "Additional": "a string", "Tags": true},
				}),
			},
			entryArgs: []*logrus.Entry{defaultLogrusEntry()},
			wantEvent: mergeEvents(t, defaultResultingEvent(), ans.Event{
				EventType: "My event type",
				Subject:   "My subject line",
				Tags:      map[string]interface{}{"Some": 1.0, "Additional": "a string", "Tags": true},
			}),
		},
		{
			name:   "Log entries should not affect each other",
			fields: fields{event: defaultEvent()},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.ErrorLevel,
					Time:    defaultTime.Add(1234),
					Message: "first log message",
					Data:    map[string]interface{}{"stepName": "testStep", "this entry": "should only be part of this event"},
				},
				defaultLogrusEntry(),
			},
			wantEvent: defaultResultingEvent(),
		},
		{
			name:   "White space messages should not send",
			fields: fields{event: defaultEvent()},
			entryArgs: []*logrus.Entry{
				{
					Level:   logrus.ErrorLevel,
					Time:    defaultTime,
					Message: "   ",
					Data:    map[string]interface{}{"stepName": "testStep"},
				},
			},
		},
		{
			name:      "Should not fire twice",
			fields:    fields{firing: true, event: defaultEvent()},
			entryArgs: []*logrus.Entry{defaultLogrusEntry()},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientMock := ansMock{}
			ansHook := &ANSHook{
				client: &clientMock,
				event:  tt.fields.event,
				firing: tt.fields.firing,
			}
			defer clientMock.cleanup()
			for _, entryArg := range tt.entryArgs {
				originalLogLevel := entryArg.Level
				ansHook.Fire(entryArg)
				assert.Equal(t, originalLogLevel.String(), entryArg.Level.String(), "Entry error level has been altered")
			}
			assert.Equal(t, tt.wantEvent, clientMock.testEvent, "Event is not as expected.")
		})
	}
}

const testCorrelationID = "1234"

var defaultTime = time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC)

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
func defaultResultingEvent() ans.Event {
	return ans.Event{
		EventType:      "Piper",
		EventTimestamp: defaultTime.Unix(),
		Severity:       "WARNING",
		Category:       "ALERT",
		Subject:        "testStep",
		Body:           "my log message",
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
		Tags: map[string]interface{}{"ans:correlationId": "1234", "ans:sourceEventId": "1234", "stepName": "testStep", "logLevel": "warning"},
	}
}

func defaultLogrusEntry() *logrus.Entry {
	return &logrus.Entry{
		Level:   logrus.WarnLevel,
		Time:    defaultTime,
		Message: "my log message",
		Data:    map[string]interface{}{"stepName": "testStep"},
	}
}

func mergeEvents(t *testing.T, event1, event2 ans.Event) ans.Event {
	event2JSON, err := json.Marshal(event2)
	require.NoError(t, err)
	err = event1.MergeWithJSON(event2JSON)
	require.NoError(t, err)
	return event1
}

type ansMock struct {
	testANS   *ans.ANS
	testEvent ans.Event
}

func (am *ansMock) Send(event ans.Event) error {
	am.testEvent = event
	return nil
}

func (am *ansMock) CheckCorrectSetup() error {
	return nil
}

func (am *ansMock) SetOptions(serviceKey ans.ServiceKey) {
	a := &ans.ANS{}
	a.SetOptions(serviceKey)
	am.testANS = a
}

func (am *ansMock) cleanup() {
	am.testANS = &ans.ANS{}
}
