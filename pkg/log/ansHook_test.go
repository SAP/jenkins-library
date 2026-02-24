//go:build unit
// +build unit

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestANSHook_Levels(t *testing.T) {

	registrationUtil := createRegUtil()

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

func TestANSHook_registerANSHook(t *testing.T) {

	os.Setenv("PIPER_ansHookServiceKey", defaultServiceKeyJSON)

	t.Run("good", func(t *testing.T) {
		t.Run("No service key skips registration", func(t *testing.T) {
			util := createRegUtil()
			os.Setenv("PIPER_ansHookServiceKey", "")
			assert.Nil(t, registerANSHookIfConfigured(testCorrelationID, util), "registration did not nil")
			assert.Nil(t, util.Hook, "registration registered hook")
			os.Setenv("PIPER_ansHookServiceKey", defaultServiceKeyJSON)
		})
		t.Run("Default registration registers hook and secret", func(t *testing.T) {
			util := createRegUtil()
			assert.Nil(t, registerANSHookIfConfigured(testCorrelationID, util), "registration did not nil")
			assert.NotNil(t, util.Hook, "registration didnt register hook")
			assert.NotNil(t, util.Secret, "registration didnt register secret")
		})
		t.Run("Registration with default template", func(t *testing.T) {
			util := createRegUtil()
			os.Setenv("PIPER_ansEventTemplate", customerEventString())
			assert.Nil(t, registerANSHookIfConfigured(testCorrelationID, util), "registration did not return nil")
			assert.Equal(t, customerEvent(), util.Hook.eventTemplate, "unexpected event template data")
			os.Setenv("PIPER_ansEventTemplate", "")
		})
		t.Run("Registration with customized template", func(t *testing.T) {
			util := createRegUtil()
			os.Setenv("PIPER_ansEventTemplate", customerEventString(map[string]interface{}{"Priority": "123"}))
			assert.Nil(t, registerANSHookIfConfigured(testCorrelationID, util), "registration did not return nil")
			assert.Equal(t, 123, util.Hook.eventTemplate.Priority, "unexpected event template data")
			os.Setenv("PIPER_ansEventTemplate", "")
		})
	})

	t.Run("bad", func(t *testing.T) {
		t.Run("Fails on check error", func(t *testing.T) {
			util := createRegUtil(map[string]interface{}{"CheckErr": fmt.Errorf("check failed")})
			err := registerANSHookIfConfigured(testCorrelationID, util)
			assert.Contains(t, err.Error(), "check failed", "unexpected error text")
		})

		t.Run("Fails on validation error", func(t *testing.T) {
			os.Setenv("PIPER_ansEventTemplate", customerEventString(map[string]interface{}{"Priority": "-1"}))
			err := registerANSHookIfConfigured(testCorrelationID, createRegUtil())
			assert.Contains(t, err.Error(), "Priority must be 1 or greater", "unexpected error text")
			os.Setenv("PIPER_ansEventTemplate", "")
		})

	})
	os.Setenv("PIPER_ansHookServiceKey", "")
}

func TestANSHook_Fire(t *testing.T) {
	registrationUtil := createRegUtil()
	ansHook := &ANSHook{
		client: registrationUtil,
	}

	t.Run("Straight forward test", func(t *testing.T) {
		ansHook.eventTemplate = defaultEvent()
		require.NoError(t, ansHook.Fire(defaultLogrusEntry()), "error is not nil")
		assert.Equal(t, defaultResultingEvent(), registrationUtil.Event, "error category tag is not as expected")
		registrationUtil.clearEventTemplate()
	})
	t.Run("Set error category", func(t *testing.T) {
		SetErrorCategory(ErrorTest)
		ansHook.eventTemplate = defaultEvent()
		require.NoError(t, ansHook.Fire(defaultLogrusEntry()), "error is not nil")
		assert.Equal(t, "test", registrationUtil.Event.Tags["cicd:errorCategory"], "error category tag is not as expected")
		SetErrorCategory(ErrorUndefined)
		registrationUtil.clearEventTemplate()
	})
	t.Run("Event already set", func(t *testing.T) {
		alreadySetEvent := ans.Event{EventType: "My event type", Subject: "My subject line", Tags: map[string]interface{}{"Some": 1.0, "Additional": "a string", "Tags": true}}
		ansHook.eventTemplate = mergeEvents(t, defaultEvent(), alreadySetEvent)
		require.NoError(t, ansHook.Fire(defaultLogrusEntry()), "error is not nil")
		assert.Equal(t, mergeEvents(t, defaultResultingEvent(), alreadySetEvent), registrationUtil.Event, "event is not as expected")
		registrationUtil.clearEventTemplate()
	})
	t.Run("Log entries should not affect each other", func(t *testing.T) {
		ansHook.eventTemplate = defaultEvent()
		SetErrorCategory(ErrorTest)
		require.NoError(t, ansHook.Fire(defaultLogrusEntry()), "error is not nil")
		assert.Equal(t, "test", registrationUtil.Event.Tags["cicd:errorCategory"], "error category tag is not as expected")
		SetErrorCategory(ErrorUndefined)
		require.NoError(t, ansHook.Fire(defaultLogrusEntry()), "error is not nil")
		assert.Nil(t, registrationUtil.Event.Tags["cicd:errorCategory"], "error category tag is not nil")
		registrationUtil.clearEventTemplate()
	})
	t.Run("White space messages should not send", func(t *testing.T) {
		ansHook.eventTemplate = defaultEvent()
		entryWithSpaceMessage := defaultLogrusEntry()
		entryWithSpaceMessage.Message = "   "
		require.NoError(t, ansHook.Fire(entryWithSpaceMessage), "error is not nil")
		assert.Equal(t, ans.Event{}, registrationUtil.Event, "event is not empty")
	})
	t.Run("Should not fire twice", func(t *testing.T) {
		ansHook.eventTemplate = defaultEvent()
		ansHook.firing = true
		require.EqualError(t, ansHook.Fire(defaultLogrusEntry()), "ANS hook has already been fired", "error message is not as expected")
		ansHook.firing = false
	})
	t.Run("No stepName set", func(t *testing.T) {
		ansHook.eventTemplate = defaultEvent()
		logrusEntryWithoutStepName := defaultLogrusEntry()
		logrusEntryWithoutStepName.Data = map[string]interface{}{}
		require.NoError(t, ansHook.Fire(logrusEntryWithoutStepName), "error is not nil")
		assert.Equal(t, "n/a", registrationUtil.Event.Tags["cicd:stepName"], "event step name tag is not as expected.")
		assert.Equal(t, "Step 'n/a' sends 'WARNING'", registrationUtil.Event.Subject, "event subject is not as expected")
		registrationUtil.clearEventTemplate()
	})
}

const testCorrelationID = "1234"
const defaultServiceKeyJSON = `{"url": "https://my.test.backend", "client_id": "myTestClientID", "client_secret": "super secret", "oauth_url": "https://my.test.oauth.provider"}`

var defaultTime = time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC)

func defaultCorrelationID() string {
	return testCorrelationID
}

func merge(base, overlay map[string]interface{}) map[string]interface{} {

	result := map[string]interface{}{}

	if base == nil {
		base = map[string]interface{}{}
	}

	for key, value := range base {
		result[key] = value
	}

	for key, value := range overlay {
		if val, ok := value.(map[string]interface{}); ok {
			if valBaseKey, ok := base[key].(map[string]interface{}); !ok {
				result[key] = merge(map[string]interface{}{}, val)
			} else {
				result[key] = merge(valBaseKey, val)
			}
		} else {
			result[key] = value
		}
	}
	return result
}

func customerEvent(params ...interface{}) ans.Event {
	event := ans.Event{}
	json.Unmarshal([]byte(customerEventString(params)), &event)
	return event
}

func customerEventString(params ...interface{}) string {
	event := defaultEvent()

	additionalFields := make(map[string]interface{})
	if len(params) > 0 {
		for i := 0; i < len(params); i++ {
			additionalFields = merge(additionalFields, pokeObject(&event, params[i]))
		}
	}

	//  create json string from Event
	marshaled, err := json.Marshal(event)
	if err != nil {
		panic(fmt.Sprintf("cannot marshal customer event: %v", err))
	}

	// add non Event members to json string
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

type registrationUtilMock struct {
	ans.Client
	Event      ans.Event
	ServiceKey ans.ServiceKey
	SendErr    error
	CheckErr   error
	Hook       *ANSHook
	Secret     string
}

func (m *registrationUtilMock) Send(event ans.Event) error {
	m.Event = event
	return m.SendErr
}

func (m *registrationUtilMock) CheckCorrectSetup() error {
	return m.CheckErr
}

func (m *registrationUtilMock) SetServiceKey(serviceKey ans.ServiceKey) {
	m.ServiceKey = serviceKey

}
func (m *registrationUtilMock) registerHook(hook *ANSHook) {
	m.Hook = hook
}

func (m *registrationUtilMock) registerSecret(secret string) {
	m.Secret = secret
}

func (m *registrationUtilMock) clearEventTemplate() {
	m.Event = ans.Event{}
}

func createRegUtil(params ...interface{}) *registrationUtilMock {

	mock := registrationUtilMock{}
	if len(params) > 0 {
		for i := 0; i < len(params); i++ {
			pokeObject(&mock, params[i])
		}
	}
	return &mock
}

func pokeObject(obj interface{}, param interface{}) map[string]interface{} {

	additionalFields := make(map[string]interface{})

	switch t := param.(type) {
	case map[string]interface{}:
		{
			m := param.(map[string]interface{})
			v := reflect.ValueOf(obj)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			for key, value := range m {
				f := v.FieldByName(key)

				if f != (reflect.Value{}) {
					switch f.Kind() {
					case reflect.String:
						f.SetString(value.(string))
					case reflect.Int, reflect.Int64:
						switch t := value.(type) {
						case string:
							v, _ := strconv.Atoi(value.(string))
							f.SetInt(int64(v))
						case int:
							f.SetInt(int64((value).(int)))
						case int64:
							f.SetInt(value.(int64))
						default:
							panic(fmt.Sprintf("unsupported value type: %v of key:%v value:%v\n", t, key, value))
						}
					case reflect.Map:
						switch value.(type) {
						case map[string]string, map[string]interface{}:
							if value != nil {
								val := reflect.ValueOf(value)
								f.Set(val)
							} else {
								f.Set(reflect.Zero(f.Type()))
							}
						}
					case reflect.Interface:
						if value != nil {
							val := reflect.ValueOf(value)
							f.Set(val)
						} else {
							f.Set(reflect.Zero(f.Type()))
						}
					default:
						panic(fmt.Sprintf("unsupported field type: %v of key:%v value:%v\n", f.Kind(), key, value))
					}
				} else {
					additionalFields[key] = value
				}
			}
		}
	case []interface{}:
		p := param.([]interface{})
		for i := 0; i < len(p); i++ {
			pokeObject(obj, p[i])
		}
	default:
		panic(fmt.Sprintf("unsupported paramter type: %v", t))
	}
	return additionalFields
}

func defaultEvent() ans.Event {
	event := ans.Event{
		EventType: "Piper",
		Tags: map[string]interface{}{
			"ans:correlationId": testCorrelationID,
			"ans:sourceEventId": testCorrelationID,
		},
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
	}
	return event
}

func defaultResultingEvent() ans.Event {
	return customerEvent(map[string]interface{}{
		"EventTimestamp": defaultTime.Unix(),
		"Severity":       "WARNING",
		"Category":       "ALERT",
		"Subject":        "Step 'testStep' sends 'WARNING'",
		"Body":           "my log message",
		"Tags": map[string]interface{}{
			"ans:correlationId": "1234",
			"ans:sourceEventId": "1234",
			"cicd:stepName":     "testStep",
			"cicd:logLevel":     "warning",
		},
	})
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
