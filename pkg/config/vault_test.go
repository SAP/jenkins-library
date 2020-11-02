package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

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
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA")}
		vaultData := map[string]string{secretName: "value1"}

		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("Secrets are not overwritten", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
			secretName:      "preset value",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA")}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)

		assert.Equal(t, "preset value", stepConfig.Config[secretName])
	})

	t.Run("Error is passed through", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA")}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, fmt.Errorf("test"))
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Len(t, stepConfig.Config, 1)
	})

	t.Run("Secret doesn't exist", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA")}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Len(t, stepConfig.Config, 1)
	})

	t.Run("Search over multiple paths", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{
			stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA", "$(vaultBasePath)/pipelineB"),
		}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(nil, nil)
		vaultMock.On("GetKvSecret", "team1/pipelineB").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("Stop lookup when secret was found", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		stepParams := []StepParameters{
			stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA", "$(vaultBasePath)/pipelineB"),
		}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
		vaultMock.AssertNotCalled(t, "GetKvSecret", "team1/pipelineB")
	})

	t.Run("No BasePath is stepConfig.Configured", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA")}
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, nil, stepConfig.Config[secretName])
		vaultMock.AssertNotCalled(t, "GetKvSecret", mock.AnythingOfType("string"))
	})
}

func TestVaultSecretFiles(t *testing.T) {
	const secretName = "testSecret"
	t.Run("Test Vault Secret File Reference", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecretFile", "$(vaultPath)/pipelineA")}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.NotNil(t, stepConfig.Config[secretName])
		path := stepConfig.Config[secretName].(string)
		contentByte, err := ioutil.ReadFile(path)
		assert.NoError(t, err)
		content := string(contentByte)
		assert.Equal(t, content, "value1")
	})

	os.RemoveAll(VaultSecretFileDirectory)
	VaultSecretFileDirectory = ""

	t.Run("Test temporary secret file cleanup", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecretFile", "$(vaultPath)/pipelineA")}
		vaultData := map[string]string{secretName: "value1"}
		assert.NoDirExists(t, VaultSecretFileDirectory)
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.NotNil(t, stepConfig.Config[secretName])
		path := stepConfig.Config[secretName].(string)
		assert.DirExists(t, VaultSecretFileDirectory)
		assert.FileExists(t, path)
		RemoveVaultSecretFiles()
		assert.NoFileExists(t, path)
		assert.NoDirExists(t, VaultSecretFileDirectory)
	})
}

func stepParam(name string, refType string, refPaths ...string) StepParameters {
	return StepParameters{
		Name: name,
		ResourceRef: []ResourceReference{
			{
				Type:  refType,
				Paths: refPaths,
			},
		},
	}
}
