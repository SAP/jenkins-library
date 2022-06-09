package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestANSHook_Levels(t *testing.T) {

	//	hook := &ANSHook{client: defaultClient(), eventTemplate: defaultEvent()}
	registrationUtil := defaultRegistrationUtil()

	t.Run("good", func(t *testing.T) {
		t.Run("default hook levels", func(t *testing.T) {
			registerANSHookIfConfigured(testCorrelationID, registrationUtil)
			assert.Equal(t, []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel}, registrationUtil.Hook.Levels())
		})
	})
}

func TestANSHook_setupEventTemplate(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		t.Run("setup event without customer template", func(t *testing.T) {
			event, _ := setupEventTemplate("", defaultCorrelationID())
			assert.Equal(t, defaultEvent(), event, "unexpected event data")
		})
		t.Run("setup event from default customer template", func(t *testing.T) {
			event, _ := setupEventTemplate(customerEventString(), defaultCorrelationID())
			assert.Equal(t, defaultEvent(), event, "unexpected event data")
		})
		t.Run("setup event with category", func(t *testing.T) {
			event, _ := setupEventTemplate(customerEventString(map[string]interface{}{"Category": "ALERT"}), defaultCorrelationID())
			assert.Equal(t, "", event.Category, "unexpected category data")
		})
		t.Run("setup event with severity", func(t *testing.T) {
			event, _ := setupEventTemplate(customerEventString(map[string]interface{}{"Severity": "WARNING"}), defaultCorrelationID())
			assert.Equal(t, "", event.Severity, "unexpected severity data")
		})
		t.Run("setup event with invalid category", func(t *testing.T) {
			event, _ := setupEventTemplate(customerEventString(map[string]interface{}{"Category": "invalid"}), defaultCorrelationID())
			assert.Equal(t, "", event.Category, "unexpected category data")
		})
		t.Run("setup event with priority", func(t *testing.T) {
			event, _ := setupEventTemplate(customerEventString(map[string]interface{}{"Priority": "1"}), defaultCorrelationID())
			assert.Equal(t, 1, event.Priority, "unexpected priority data")
		})
		t.Run("setup event with omitted priority 0", func(t *testing.T) {
			event, err := setupEventTemplate(customerEventString(map[string]interface{}{"Priority": "0"}), defaultCorrelationID())
			assert.Equal(t, nil, err, "priority 0 must not fail")
			assert.Equal(t, 0, event.Priority, "unexpected priority data")
		})
	})

	t.Run("bad", func(t *testing.T) {
		t.Run("setup event with invalid priority", func(t *testing.T) {
			_, err := setupEventTemplate(customerEventString(map[string]interface{}{"Priority": "-1"}), defaultCorrelationID())
			assert.Contains(t, err.Error(), "Priority must be 1 or greater", "unexpected error text")
		})
		t.Run("setup event with invalid variable name", func(t *testing.T) {
			_, err := setupEventTemplate(customerEventString(map[string]interface{}{"Invalid": "invalid"}), defaultCorrelationID())
			assert.Contains(t, err.Error(), "could not be unmarshalled", "unexpected error text")
		})
	})
}

func TestANSHook_newANSHook(t *testing.T) {
	t.Parallel()
	type args struct {
		serviceKey    string
		eventTemplate string
	}
	tests := []struct {
		name                     string
		args                     args
		eventTemplateFileContent string
		checkErr                 error
		wantEvent                ans.Event
		wantErrMsg               string
	}{
		{
			name:      "Straight forward test",
			args:      args{serviceKey: defaultServiceKeyJSON},
			wantEvent: defaultEvent(),
		},
		{
			name:       "No service key yields error",
			wantErrMsg: "cannot initialize SAP Alert Notification Service due to faulty serviceKey json: error unmarshalling ANS serviceKey: unexpected end of JSON input",
		},
		{
			name:       "Fails on check error",
			args:       args{serviceKey: defaultServiceKeyJSON},
			wantErrMsg: "check http request to SAP Alert Notification Service failed; not setting up the ANS hook: check failed",
			checkErr:   fmt.Errorf("check failed"),
		},
		{
			name:                     "With event template as file",
			args:                     args{serviceKey: defaultServiceKeyJSON},
			eventTemplateFileContent: `{"priority":123}`,
			wantEvent:                mergeEvents(t, defaultEvent(), ans.Event{Priority: 123}),
		},
		{
			name:      "With event template as string",
			args:      args{serviceKey: defaultServiceKeyJSON, eventTemplate: `{"priority":123}`},
			wantEvent: mergeEvents(t, defaultEvent(), ans.Event{Priority: 123}),
		},
		{
			name:                     "With event template from two sources, string overwrites file",
			args:                     args{serviceKey: defaultServiceKeyJSON, eventTemplate: `{"priority":789}`},
			eventTemplateFileContent: `{"priority":123}`,
			wantEvent:                mergeEvents(t, defaultEvent(), ans.Event{Priority: 789}),
		},
		{
			name:       "Fails on validation error",
			args:       args{serviceKey: defaultServiceKeyJSON, eventTemplate: `{"priority":-1}`},
			wantErrMsg: "did not initialize SAP Alert Notification Service due to faulty event template json: Priority must be 1 or greater: event JSON failed the validation",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var testEventTemplateFilePath string
			if len(tt.eventTemplateFileContent) > 0 {
				testEventTemplateFilePath = writeTempFile(t, tt.eventTemplateFileContent)
				defer os.Remove(testEventTemplateFilePath)
			}

			registrationUtil := defaultRegistrationUtil(map[string]interface{}{"CheckErr": tt.checkErr})

			//			clientMock := ansMock{checkErr: tt.checkErr}
			if err := registerANSHookIfConfigured(testCorrelationID, registrationUtil); err != nil {
				assert.EqualError(t, err, tt.wantErrMsg, "Error mismatch")
			} else {
				assert.Equal(t, tt.wantErrMsg, "", "There was an error expected")
				//				assert.Equal(t, defaultANSClient(), clientMock.a, "new ANSHook not as expected")
				//				assert.Equal(t, tt.wantEvent, got.eventTemplate, "new ANSHook not as expected")
			}
		})
	}
}

func TestANSHook_Fire(t *testing.T) {
	SetErrorCategory(ErrorTest)
	defer SetErrorCategory(ErrorUndefined)
	type fields struct {
		levels       []logrus.Level
		defaultEvent ans.Event
		firing       bool
	}
	tests := []struct {
		name       string
		fields     fields
		entryArgs  []*logrus.Entry
		wantEvent  ans.Event
		wantErrMsg string
	}{
		{
			name:      "Straight forward test",
			fields:    fields{defaultEvent: defaultEvent()},
			entryArgs: []*logrus.Entry{defaultLogrusEntry()},
			wantEvent: defaultResultingEvent(),
		},
		{
			name: "Event already set",
			fields: fields{
				defaultEvent: mergeEvents(t, defaultEvent(), ans.Event{
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
			fields: fields{defaultEvent: defaultEvent()},
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
			fields: fields{defaultEvent: defaultEvent()},
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
			name:       "Should not fire twice",
			fields:     fields{firing: true, defaultEvent: defaultEvent()},
			entryArgs:  []*logrus.Entry{defaultLogrusEntry()},
			wantErrMsg: "ANS hook has already been fired",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registrationUtil := defaultRegistrationUtil()

			ansHook := &ANSHook{
				client:        registrationUtil,
				eventTemplate: tt.fields.defaultEvent,
				firing:        tt.fields.firing,
			}
			for _, entryArg := range tt.entryArgs {
				originalLogLevel := entryArg.Level
				if err := ansHook.Fire(entryArg); err != nil {
					assert.EqualError(t, err, tt.wantErrMsg)
				}
				assert.Equal(t, originalLogLevel.String(), entryArg.Level.String(), "Entry error level has been altered")
			}
			assert.Equal(t, tt.wantEvent, registrationUtil.Event, "Event is not as expected.")
		})
	}
}

const testCorrelationID = "1234"
const defaultServiceKeyJSON = `{"url": "https://my.test.backend", "client_id": "myTestClientID", "client_secret": "super secret", "oauth_url": "https://my.test.oauth.provider"}`

var defaultTime = time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC)

func defaultCorrelationID() string {
	return testCorrelationID
}

func customerEventString(params ...interface{}) string {
	event := defaultEvent()

	additionalFields := make(map[string]interface{})

	if len(params) > 0 {
		for i := 0; i < len(params); i++ {
			switch params[i].(type) {
			case map[string]interface{}:
				{
					m := params[i].(map[string]interface{})
					for key, value := range m {
						obj := reflect.Indirect(reflect.ValueOf(&event))
						if field := obj.FieldByName(key); field != (reflect.Value{}) {
							switch field.Kind() {
							case reflect.String:
								field.SetString(value.(string))
							case reflect.Int:
								switch value.(type) {
								case string:
									v, _ := strconv.Atoi(value.(string))
									field.SetInt(int64(v))
								case int:
									field.SetInt(int64((value).(int)))
								}
							}
						} else {
							additionalFields[key] = value
						}
					}
				}
			}
		}
	}

	marshaled, err := json.Marshal(event)
	if err != nil {
		panic(fmt.Sprintf("cannot marshal customer event: %v", err))
	}

	if len(additionalFields) > 0 {
		closingBraceIdx := bytes.LastIndexByte(marshaled, '}')
		for key, value := range additionalFields {
			var entry string
			switch value.(type) {
			default:
				panic(fmt.Sprintf("invalid key value type: %v", key))
			case string:
				entry = `, "` + key + `": "` + value.(string) + `"`
			case int:
				entry = `, "` + key + `": "` + strconv.Itoa(value.(int)) + `"`
			}

			add := []byte(entry)
			marshaled = append(marshaled[:closingBraceIdx], add...)
		}
		marshaled = append(marshaled, '}')
	}

	return string(marshaled)
}

type RegistrationUtilMock struct {
	ans.Client
	Event      ans.Event
	ServiceKey ans.ServiceKey
	SendErr    error
	CheckErr   error
	Hook       *ANSHook
}

func (m *RegistrationUtilMock) Send(event ans.Event) error {
	m.Event = event
	return m.SendErr
}

func (m *RegistrationUtilMock) CheckCorrectSetup() error {
	return m.CheckErr
}

func (m *RegistrationUtilMock) SetServiceKey(serviceKey ans.ServiceKey) {
	m.ServiceKey = serviceKey
}

func (m *RegistrationUtilMock) registerHook(hook *ANSHook) {
	m.Hook = hook
}

func defaultRegistrationUtil(params ...interface{}) *RegistrationUtilMock {

	mock := RegistrationUtilMock{}
	if len(params) > 0 {
		for i := 0; i < len(params); i++ {
			pokeObject(&mock, params[i])
		}
	}
	return &mock
}

func pokeObject(obj interface{}, param interface{}) map[string]interface{} {

	additionalFields := make(map[string]interface{})

	switch param.(type) {
	case map[string]interface{}:
		{
			m := param.(map[string]interface{})
			v := reflect.ValueOf(obj)
			iv := reflect.Indirect(v)
			fmt.Printf("Obj ValueOf:  v:%v, v.i:%v, kind:%v, type:%v\n", v, v.Interface(), v.Kind(), v.Type())
			fmt.Printf("Obj Indirect: v:%v, v.i:%v, kind:%v, type:%v\n", iv, iv.Interface(), iv.Kind(), v.Type())

			if v.Kind() == reflect.Ptr {
				v = v.Elem()
				fmt.Printf("Obj Elem:     v:%v, v.i:%v, kind:%v, type:%v\n", v, v.Interface(), v.Kind(), v.Type())
			}
			for key, value := range m {
				f := v.FieldByName(key)
				fmt.Printf("Field(%v):    f:%v, kind:%v, type:%v\n", key, f, f.Kind(), v.Type())

				if f != (reflect.Value{}) {
					switch f.Kind() {
					case reflect.String:
						f.SetString(value.(string))
					case reflect.Int:
						switch value.(type) {
						case string:
							v, _ := strconv.Atoi(value.(string))
							f.SetInt(int64(v))
						case int:
							f.SetInt(int64((value).(int)))
						}
					case reflect.Interface:
						n := reflect.New(reflect.TypeOf(value))
						n.Elem().Set(reflect.ValueOf(value))
						f.Set(n)
					}
				} else {
					additionalFields[key] = value
				}
			}
		}
	}
	return additionalFields
}

func defaultEvent() ans.Event {
	event := ans.Event{
		EventType: "Piper",
		Tags:      map[string]interface{}{"ans:correlationId": testCorrelationID, "ans:sourceEventId": testCorrelationID},
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
	}
	return event
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
		Tags: map[string]interface{}{"ans:correlationId": "1234", "ans:sourceEventId": "1234", "stepName": "testStep", "logLevel": "warning", "errorCategory": "test"},
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

func defaultANSClient() *ans.ANS {
	return &ans.ANS{
		XSUAA: xsuaa.XSUAA{
			OAuthURL:     "https://my.test.oauth.provider",
			ClientID:     "myTestClientID",
			ClientSecret: "super secret",
		},
		URL: "https://my.test.backend",
	}
}

func writeTempFile(t *testing.T, fileContent string) (fileName string) {
	var err error
	testEventTemplateFile, err := os.CreateTemp("", "event_template_*.json")
	require.NoError(t, err, "File creation failed!")
	defer testEventTemplateFile.Close()
	data := []byte(fileContent)
	_, err = testEventTemplateFile.Write(data)
	require.NoError(t, err, "Could not write test data to test file!")
	return testEventTemplateFile.Name()
}

func mergeEvents(t *testing.T, event1, event2 ans.Event) ans.Event {
	event2JSON, err := json.Marshal(event2)
	require.NoError(t, err)
	err = event1.MergeWithJSON(event2JSON)
	require.NoError(t, err)
	return event1
}

type aansMock struct {
	event      ans.Event
	serviceKey ans.ServiceKey
	sendErr    error
	checkErr   error
}

func (am *aansMock) Send(event ans.Event) error {
	am.event = event
	return am.sendErr
}

func (am *aansMock) CheckCorrectSetup() error {
	return am.checkErr
}

func (am *aansMock) SetServiceKey(serviceKey ans.ServiceKey) {
	am.serviceKey = serviceKey
}
