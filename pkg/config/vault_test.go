package config

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config/mocks"
	"github.com/stretchr/testify/assert"
)

func TestVaultConfigLoad(t *testing.T) {
	const secretName = "testSecret"
	t.Parallel()
	t.Run("Load secret from vault", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"testBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "testBasePath", "vaultSecret", "pipelineA")}
		vaultData := map[string]string{secretName: "value1"}

		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		config, err := getVaultConfig(vaultMock, stepConfig, stepParams)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "value1", config[secretName])
	})

	t.Run("Secrets are not overwritten", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"testBasePath": "team1",
			secretName:     "preset value",
		}}
		stepParams := []StepParameters{stepParam(secretName, "testBasePath", "vaultSecret", "pipelineA")}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		config, err := getVaultConfig(vaultMock, stepConfig, stepParams)

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.NotContains(t, config, secretName)
	})

	t.Run("Error is passed through", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"testBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "testBasePath", "vaultSecret", "pipelineA")}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, fmt.Errorf("test"))
		config, err := getVaultConfig(vaultMock, stepConfig, stepParams)
		assert.Nil(t, config)
		assert.EqualError(t, err, "test")
	})

	t.Run("Secret doesn't exist", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"testBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "testBasePath", "vaultSecret", "pipelineA")}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, nil)
		config, err := getVaultConfig(vaultMock, stepConfig, stepParams)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Len(t, config, 0)
	})
}

func stepParam(name, refName, refType, refPath string) StepParameters {
	return StepParameters{
		Name: name,
		ResourceRef: []ResourceReference{
			ResourceReference{
				Name: refName,
				Type: refType,
				Path: refPath,
			},
		},
	}
}
