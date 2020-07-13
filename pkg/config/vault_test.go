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
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "pipelineA")}
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
			"vaultBasePath": "team1",
			secretName:      "preset value",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "pipelineA")}
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
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "pipelineA")}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, fmt.Errorf("test"))
		config, err := getVaultConfig(vaultMock, stepConfig, stepParams)
		assert.Nil(t, config)
		assert.EqualError(t, err, "test")
	})

	t.Run("Secret doesn't exist", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "pipelineA")}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, nil)
		config, err := getVaultConfig(vaultMock, stepConfig, stepParams)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Len(t, config, 0)
	})

	t.Run("Search over multiple references", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{
			stepParam(secretName, "vaultSecret", "pipelineA"),
			stepParam(secretName, "vaultSecret", "pipelineB"),
		}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, nil)
		vaultMock.On("GetKvSecret", "team1/pipelineB").Return(vaultData, nil)
		config, err := getVaultConfig(vaultMock, stepConfig, stepParams)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "value1", config[secretName])
	})

	t.Run("No BasePath is configured", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "pipelineA")}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "pipelineA").Return(vaultData, nil)
		config, err := getVaultConfig(vaultMock, stepConfig, stepParams)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "value1", config[secretName])
	})
}

func stepParam(name, refType, refPath string) StepParameters {
	return StepParameters{
		Name: name,
		ResourceRef: []ResourceReference{
			ResourceReference{
				Type:  refType,
				Paths: []string{refPath},
			},
		},
	}
}
