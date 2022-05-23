package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRunAnsSendEvent(t *testing.T) {
	t.Parallel()

	defaultEvent := ans.Event{
		EventType: "Piper",
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
		Subject:        fmt.Sprint("testStep"),
		Body:           fmt.Sprintf("Call from Piper step: %s", "testStep"),
		EventTimestamp: 1651585103,
		Severity:       "INFO",
		Category:       "NOTIFICATION",
	}
	goodServiceKey := `{
				"url": "https://my.test.backend",
				"client_id": "myTestClientID",
				"client_secret": "super secret",
				"oauth_url": "https://my.test.oauth.provider"
			   }`

	log.Entry().Data["stepName"] = "testStep"

	t.Run("happy path - overwriting timestamp of event", func(t *testing.T) {
		t.Parallel()
		// init
		config := ansSendEventOptions{
			AnsServiceKey: goodServiceKey,
			EventJSON:     `{"eventTimestamp": 1651585103}`,
		}
		am := ansMock{}
		defer am.cleanup()

		// test
		err := runAnsSendEvent(&config, &am)

		// assert
		require.NoError(t, err)

		assert.Equal(t, "https://my.test.backend", am.testANS.URL)
		assert.Equal(t, "myTestClientID", am.testANS.XSUAA.ClientID)
		assert.Equal(t, "super secret", am.testANS.XSUAA.ClientSecret)
		assert.Equal(t, "https://my.test.oauth.provider", am.testANS.XSUAA.OAuthURL)

		assert.Equal(t, defaultEvent, am.testEvent)
	})

	t.Run("error - bad service key", func(t *testing.T) {
		t.Parallel()
		// init
		config := ansSendEventOptions{
			AnsServiceKey: `{
						"url": "https://my.test.backend",
						"client_id": "myTestClientID",
						"client_secret": "super secret",
						"oauth_url": "https://my.test.oauth.provider"`,
		}

		// test
		err := runAnsSendEvent(&config, &ansMock{})

		// assert
		assert.EqualError(t, err, "error unmarshalling ANS serviceKey: unexpected end of JSON input")
	})

	t.Run("error - bad event json", func(t *testing.T) {
		t.Parallel()
		// init
		config := ansSendEventOptions{
			AnsServiceKey: goodServiceKey,
			EventJSON:     `{"eventTimestamp": 1651585103`,
		}

		// test
		err := runAnsSendEvent(&config, &ansMock{})

		// assert
		assert.EqualError(t, err, "error unmarshalling ANS event from JSON string \"{\\\"eventTimestamp\\\": 1651585103\": unexpected end of JSON input")
	})

	t.Run("error - fail to send", func(t *testing.T) {
		t.Parallel()
		// init
		config := ansSendEventOptions{
			AnsServiceKey: goodServiceKey,
			EventJSON:     `{"eventTimestamp": 1651585103}`,
		}
		am := ansMock{failToSend: true}
		defer am.cleanup()

		// test
		err := runAnsSendEvent(&config, &am)

		// assert
		assert.EqualError(t, err, "failed to send")
	})
}

type ansMock struct {
	testANS    ans.ANS
	testEvent  ans.Event
	failToSend bool
}

func (am *ansMock) Send(event ans.Event) error {
	if am.failToSend {
		return fmt.Errorf("failed to send")
	}
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
