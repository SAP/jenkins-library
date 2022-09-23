package pact

import (
	"fmt"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecPactPublish(t *testing.T) {
	t.Parallel()

	t.Run("success - default", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{
			GitRepo: "testRepo",
			Utils: mockUtils,
			PathToPactsFolder: "pacts/",
		}
		mockUtils.AddFile("pacts/pact1.json", []byte(`{"consumer":{"name":"testRepo-http"},"provider":{"name":"testRepo-http"}}`))
		mockUtils.AddFile("pacts/pact1.json", []byte(`{"consumer":{"name":"testRepo-async"},"provider":{"name":"testRepo-async"}}`))
		err := pConfig.ExecPactPublish()
		assert.NoError(t, err)
		assert.Equal(t, "pact", mockUtils.Calls[0].Exec)
	})

	t.Run("failure - pact file parsing", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{
			GitRepo: "testRepo",
			Utils: mockUtils,
			PathToPactsFolder: "pacts/",
		}
		mockUtils.AddFile("pacts/pact1.json", []byte(`{"`))
		err := pConfig.ExecPactPublish()
		assert.Contains(t, fmt.Sprint(err), "failed to parse pact file:")
	})

	t.Run("failure - invalid naming of contract", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{
			GitRepo: "testRepo",
			Utils: mockUtils,
			PathToPactsFolder: "pacts/",
		}
		mockUtils.AddFile("pacts/pact1.json", []byte(`{"consumer":{"name":"testRepo"},"provider":{"name":"testRepo-http"}}`))
		err := pConfig.ExecPactPublish()
		assert.EqualError(t, err, "pact contract does not follow the correct naming conventions: pacts/pact1.json")
	})

	t.Run("failure - no files", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{
			GitRepo: "testRepo",
			Utils: mockUtils,
			PathToPactsFolder: "pacts/",
		}
		err := pConfig.ExecPactPublish()
		assert.Contains(t, fmt.Sprint(err), "no pact files found")
	})

	t.Run("failure - publishing", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{
			GitRepo: "testRepo",
			Utils: mockUtils,
			PathToPactsFolder: "pacts/",
		}
		mockUtils.AddFile("pacts/pact1.json", []byte(`{"consumer":{"name":"testRepo-http"},"provider":{"name":"testRepo-http"}}`))
		mockUtils.AddFile("pacts/pact1.json", []byte(`{"consumer":{"name":"testRepo-async"},"provider":{"name":"testRepo-async"}}`))
		mockUtils.LookPathError = fmt.Errorf("lookPath error")
		err := pConfig.ExecPactPublish()
		assert.Contains(t, fmt.Sprint(err), "lookPath error")
	})

	t.Run("failure - reporting", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{
			GitRepo: "testRepo",
			Utils: mockUtils,
			PathToPactsFolder: "pacts/",
		}
		mockUtils.AddFile("pacts/pact1.json", []byte(`{"consumer":{"name":"testRepo-http"},"provider":{"name":"testRepo-http"}}`))
		mockUtils.AddFile("pacts/pact1.json", []byte(`{"consumer":{"name":"testRepo-async"},"provider":{"name":"testRepo-async"}}`))
		mockUtils.FileWriteErrors = map[string]error{"pactPublishReport.json": fmt.Errorf("write failure")}
		err := pConfig.ExecPactPublish()
		assert.Contains(t, fmt.Sprint(err), "error saving report")
	})
}

func TestPublishPact(t *testing.T) {
	t.Parallel()

	t.Run("success - publising", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{
			GitCommit: "theCommit",
			GitSourceBranch: "main",
		}
		c := NewPactBrokerClient("testhost", "testuser", "testpassword")
		buf := bytes.NewBuffer([]byte{})
		
		err := c.PublishPact(&pConfig, "path/to/contract", mockUtils, buf)
		assert.NoError(t, err)
		assert.Equal(t, "pact", mockUtils.Calls[0].Exec)
		expectedParams := []string {
			"publish",
			"path/to/contract",
			"--broker-username=testuser",
			"--broker-password=testpassword",
			"--broker-base-url=https://testhost",
			"--consumer-app-version=theCommit",
			"--tag=main",
		}
		assert.Equal(t, expectedParams, mockUtils.Calls[0].Params)
	})


	t.Run("success - already published", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{		}
		c := NewPactBrokerClient("testhost", "testuser", "testpassword")
		buf := bytes.NewBuffer([]byte{})
		mockUtils.ShouldFailOnCommand = map[string]error{"pact": fmt.Errorf("pact error")}
		mockUtils.StdoutReturn = map[string]string{"pact": "Each pact must be published with a unique consumer version number."}

		err := c.PublishPact(&pConfig, "path/to/contract", mockUtils, buf)
		assert.NoError(t, err)
	})

	t.Run("failure - executable not found", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{		}
		c := NewPactBrokerClient("testhost", "testuser", "testpassword")
		buf := bytes.NewBuffer([]byte{})
		mockUtils.LookPathError = fmt.Errorf("lookPath error")
		
		err := c.PublishPact(&pConfig, "path/to/contract", mockUtils, buf)
		assert.EqualError(t, err, "failed to find pact executable 'pact': lookPath error")
	})

	t.Run("failure - publishing", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		pConfig := PublishConfig{		}
		c := NewPactBrokerClient("testhost", "testuser", "testpassword")
		buf := bytes.NewBuffer([]byte{})
		mockUtils.ShouldFailOnCommand = map[string]error{"pact": fmt.Errorf("pact error")}
		
		err := c.PublishPact(&pConfig, "path/to/contract", mockUtils, buf)
		assert.EqualError(t, err, "pact error")
	})
}

func TestEnforceNaming(t *testing.T) {
	tt := []struct{
		name string
		gitRepo string
		consumerName string
		providerName string
		expected bool
	}{
		{name: "success - http", gitRepo: "testRepo", consumerName: "testRepo-http", providerName: "provider-http", expected: true},
		{name: "success - async", gitRepo: "testRepo", consumerName: "testRepo-async", providerName: "provider-async", expected: true},
		{name: "failure - http consumer", gitRepo: "testRepo", consumerName: "testRepo", providerName: "provider-http", expected: false},
		{name: "failure - http provider", gitRepo: "testRepo", consumerName: "testRepo-http", providerName: "provider", expected: false},
		{name: "failure - async consumer", gitRepo: "testRepo", consumerName: "testRepo", providerName: "provider-async", expected: false},
		{name: "failure - async provider", gitRepo: "testRepo", consumerName: "testRepo-async", providerName: "provider", expected: false},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T){
			assert.Equal(t, test.expected, enforceNaming(test.gitRepo, test.consumerName, test.providerName))
		})
	}
}