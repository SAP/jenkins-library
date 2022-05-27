package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const testTimestamp = 1651585103

func TestRunAnsSendEvent(t *testing.T) {
	t.Parallel()

	log.Entry().Data["stepName"] = "testStep"

	tests := []struct {
		name       string
		config     ansSendEventOptions
		ansMock    ansMock
		wantErrMsg string
	}{
		{
			name:   "overwriting EventType",
			config: ansSendEventOptions{AnsServiceKey: goodServiceKey, EventJSON: goodEventJSON},
		},
		{
			name:       "bad service key",
			config:     ansSendEventOptions{AnsServiceKey: `{"forgot": "closing", "bracket": "json"`},
			wantErrMsg: `error unmarshalling ANS serviceKey: unexpected end of JSON input`,
		},
		{
			name:       "bad event json",
			config:     ansSendEventOptions{AnsServiceKey: goodServiceKey, EventJSON: `faulty JSON`},
			wantErrMsg: `error unmarshalling ANS event from JSON string "faulty JSON": invalid character 'u' in literal false (expecting 'l')`,
		},
		{
			name:       "unknown field in json",
			config:     ansSendEventOptions{AnsServiceKey: goodServiceKey, EventJSON: `{"unknown": "yields error"}`},
			wantErrMsg: `error unmarshalling ANS event from JSON string "{\"unknown\": \"yields error\"}": json: unknown field "unknown"`,
		},
		{
			name:       "invalid event json",
			config:     ansSendEventOptions{AnsServiceKey: goodServiceKey, EventJSON: `{"severity": "WRONG_SEVERITY"}`},
			wantErrMsg: `Severity must be one of [INFO NOTICE WARNING ERROR FATAL]: event JSON failed the validation`,
		},
		{
			name:       "fail to send",
			config:     ansSendEventOptions{AnsServiceKey: goodServiceKey, EventJSON: goodEventJSON},
			ansMock:    ansMock{failToSend: true},
			wantErrMsg: `failed to send`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer tt.ansMock.cleanup()
			if err := runAnsSendEvent(&tt.config, &tt.ansMock); tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "https://my.test.backend", tt.ansMock.testANS.URL)
				assert.Equal(t, "myTestClientID", tt.ansMock.testANS.XSUAA.ClientID)
				assert.Equal(t, "super secret", tt.ansMock.testANS.XSUAA.ClientSecret)
				assert.Equal(t, "https://my.test.oauth.provider", tt.ansMock.testANS.XSUAA.OAuthURL)
				assert.Equal(t, defaultEvent, tt.ansMock.testEvent)
			}

		})
	}
}

func defaultEvent() ans.Event {
	return ans.Event{
		EventType:      "myEvent",
		EventTimestamp: testTimestamp,
		Severity:       "INFO",
		Category:       "NOTIFICATION",
		Subject:        "testStep",
		Body:           "Call from Piper step: testStep",
		Priority:       123,
		Resource: &ans.Resource{
			ResourceName:     "myResourceName",
			ResourceType:     "myResourceType",
			ResourceInstance: "myResourceInstance",
		},
	}
}

const (
	goodServiceKey = `{
				"url": "https://my.test.backend",
				"client_id": "myTestClientID",
				"client_secret": "super secret",
				"oauth_url": "https://my.test.oauth.provider"
			   }`
	goodEventJSON = `{"eventType": "myEvent"}`
)

type ansMock struct {
	testANS    ans.ANS
	testEvent  ans.Event
	failToSend bool
}

func (am *ansMock) Send(event ans.Event) error {
	if am.failToSend {
		return fmt.Errorf("failed to send")
	}
	event.EventTimestamp = testTimestamp
	am.testEvent = event
	return nil
}

func (am ansMock) CheckCorrectSetup() error {
	return fmt.Errorf("not implemented")
}

func (am *ansMock) SetServiceKey(serviceKey ans.ServiceKey) {
	am.testANS.SetServiceKey(serviceKey)
}

func (am *ansMock) cleanup() {
	am = &ansMock{}
}
