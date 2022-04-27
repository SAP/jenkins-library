package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunAnsSendEvent(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := ansSendEventOptions{
			AnsServiceKey: `{
						"url": "https://my.test.backend",
						"client_id": "myTestClientID",
						"client_secret": "super secret",
						"oauth_url": "https://my.test.oauth.provider"
					}`,
			EventJSON: 	`{
						"subject": "test"
					}`,
		}

		// test
		am := &ansMock{}
		defer am.cleanup()
		err := runAnsSendEvent(&config, am)

		assert.Equal(t, "https://my.test.backend", am.testANS.URL)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := ansSendEventOptions{}

		// test
		err := runAnsSendEvent(&config, &ansMock{})

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}

type ansMock struct{
	testANS ans.ANS
}

func (am ansMock) Send(event ans.Event) error {
	return nil
}

func (am ansMock) CheckCorrectSetup() error {
	return fmt.Errorf("not implemented")
}

func (am *ansMock) SetOptions(serviceKey ans.ServiceKey) {
	am.testANS.SetOptions(serviceKey)
}

func (am *ansMock) cleanup() {
	am.testANS = ans.ANS{}
}
