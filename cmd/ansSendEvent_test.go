package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const testTimestamp = 1651585103

func TestRunAnsSendEvent(t *testing.T) {
	tests := []struct {
		name       string
		config     ansSendEventOptions
		ansMock    ansMock
		wantErrMsg string
	}{
		{
			name:   "overwriting EventType",
			config: defaultEventOptions(),
		},
		{
			name:       "bad service key",
			config:     ansSendEventOptions{AnsServiceKey: `{"forgot": "closing", "bracket": "json"`},
			wantErrMsg: `error unmarshalling ANS serviceKey: unexpected end of JSON input`,
		},
		{
			name:       "invalid event json",
			config:     ansSendEventOptions{AnsServiceKey: goodServiceKey, Severity: "WRONG_SEVERITY"},
			wantErrMsg: `Severity must be one of [INFO NOTICE WARNING ERROR FATAL]: event JSON failed the validation`,
		},
		{
			name:       "fail to send",
			config:     defaultEventOptions(),
			ansMock:    ansMock{failToSend: true},
			wantErrMsg: `failed to send`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runAnsSendEvent(&tt.config, &tt.ansMock); tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "https://my.test.backend", tt.ansMock.testANS.URL)
				assert.Equal(t, defaultXsuaa(), tt.ansMock.testANS.XSUAA)
				assert.Equal(t, defaultEvent(), tt.ansMock.testEvent)
			}

		})
	}
}

func defaultEventOptions() ansSendEventOptions {
	return ansSendEventOptions{
		AnsServiceKey:    goodServiceKey,
		EventType:        "myEvent",
		Severity:         "INFO",
		Category:         "NOTIFICATION",
		Subject:          "testStep",
		Body:             "Call from Piper step: testStep",
		Priority:         123,
		Tags:             map[string]interface{}{"myNumber": 456},
		ResourceName:     "myResourceName",
		ResourceType:     "myResourceType",
		ResourceInstance: "myResourceInstance",
		ResourceTags:     map[string]interface{}{"myBoolean": true},
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
		Tags:           map[string]interface{}{"myNumber": 456},
		Resource: &ans.Resource{
			ResourceName:     "myResourceName",
			ResourceType:     "myResourceType",
			ResourceInstance: "myResourceInstance",
			Tags:             map[string]interface{}{"myBoolean": true},
		},
	}
}

func defaultXsuaa() xsuaa.XSUAA {
	return xsuaa.XSUAA{
		OAuthURL:     "https://my.test.oauth.provider",
		ClientID:     "myTestClientID",
		ClientSecret: "super secret",
	}
}

const goodServiceKey = `{
				"url": "https://my.test.backend",
				"client_id": "myTestClientID",
				"client_secret": "super secret",
				"oauth_url": "https://my.test.oauth.provider"
			   }`

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
