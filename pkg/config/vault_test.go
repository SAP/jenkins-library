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
		err := addVaultCredentials(&stepConfig, vaultMock, stepParams)
		assert.NoError(t, err)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
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
		err := addVaultCredentials(&stepConfig, vaultMock, stepParams)

		assert.NoError(t, err)
		assert.Equal(t, "preset value", stepConfig.Config[secretName])
	})

	t.Run("Error is passed through", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "pipelineA")}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, fmt.Errorf("test"))
		err := addVaultCredentials(&stepConfig, vaultMock, stepParams)
		assert.Len(t, stepConfig.Config, 1)
		assert.EqualError(t, err, "test")
	})

	t.Run("Secret doesn't exist", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "pipelineA")}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, nil)
		err := addVaultCredentials(&stepConfig, vaultMock, stepParams)
		assert.NoError(t, err)
		assert.Len(t, stepConfig.Config, 1)
	})

	t.Run("Search over multiple paths", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{
			stepParam(secretName, "vaultSecret", "pipelineA", "pipelineB"),
		}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, nil)
		vaultMock.On("GetKvSecret", "team1/pipelineB").Return(vaultData, nil)
		err := addVaultCredentials(&stepConfig, vaultMock, stepParams)
		assert.NoError(t, err)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("No BasePath is stepConfig.Configured", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "pipelineA")}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "pipelineA").Return(vaultData, nil)
		err := addVaultCredentials(&stepConfig, vaultMock, stepParams)
		assert.NoError(t, err)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

}

func stepParam(name string, refType string, refPaths ...string) StepParameters {
	return StepParameters{
		Name: name,
		ResourceRef: []ResourceReference{
			ResourceReference{
				Type:  refType,
				Paths: refPaths,
			},
		},
	}
}
