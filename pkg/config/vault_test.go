package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
			"vaultBasePath":         "team1",
			secretName:              "preset value",
			"vaultDisableOverwrite": true,
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA")}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)

		assert.Equal(t, "preset value", stepConfig.Config[secretName])
	})

	t.Run("Secrets can be overwritten", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
			secretName:      "preset value",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA")}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)

		assert.Equal(t, "value1", stepConfig.Config[secretName])
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

	t.Run("Alias names should be considered", func(t *testing.T) {
		aliasName := "alias"
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team1",
		}}
		param := stepParam(secretName, "vaultSecret", "$(vaultBasePath)/pipelineA")
		addAlias(&param, aliasName)
		stepParams := []StepParameters{param}
		vaultData := map[string]string{aliasName: "value1"}
		vaultMock.On("GetKvSecret", "team1/pipelineA").Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
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

func TestMixinVault(t *testing.T) {
	vaultServerUrl := "https://testServer"
	vaultPath := "testPath"
	config := StepConfig{
		Config:     map[string]interface{}{},
		HookConfig: nil,
	}
	general := map[string]interface{}{
		"vaultPath": vaultPath,
	}
	steps := map[string]interface{}{
		"vaultServerUrl": vaultServerUrl,
		"unknownConfig":  "test",
	}

	config.mixinVaultConfig(general, steps)

	assert.Contains(t, config.Config, "vaultServerUrl")
	assert.Equal(t, vaultServerUrl, config.Config["vaultServerUrl"])
	assert.Contains(t, config.Config, "vaultPath")
	assert.Equal(t, vaultPath, config.Config["vaultPath"])
	assert.NotContains(t, config.Config, "unknownConfig")

}

func stepParam(name string, refType string, refPaths ...string) StepParameters {
	return StepParameters{
		Name:    name,
		Aliases: []Alias{},
		ResourceRef: []ResourceReference{
			{
				Type:  refType,
				Paths: refPaths,
			},
		},
	}
}

func addAlias(param *StepParameters, aliasName string) {
	alias := Alias{Name: aliasName}
	param.Aliases = append(param.Aliases, alias)
}

func Test_resolveVaultTestCredentials(t *testing.T) {
	vaultMock := &mocks.VaultMock{}
	envPrefix := "PIPER_TESTCREDENTIAL_"

	stepConfig := StepConfig{Config: map[string]interface{}{
		"vaultPath":               "team1",
		"vaultTestCredentialPath": "appCredentials",
		"vaultTestCredentialKeys": []string{"appUser", "appUserPw"},
	}}

	vaultData := map[string]string{"appUser": "test-user", "appUserPw": "password1234"}
	vaultMock.On("GetKvSecret", "team1/appCredentials").Return(vaultData, nil)

	resolveVaultTestCredentials(&stepConfig, vaultMock)
	for k, v := range vaultData {
		env := envPrefix + k
		assert.NotEmpty(t, os.Getenv(env))
		assert.Equal(t, os.Getenv(env), v)
	}

	// assert.NotNil(t, stepConfig.Config["appUser"])
	// assert.NotNil(t, stepConfig.Config["appUserPw"])
	// assert.Equal(t, stepConfig.Config["appUser"], "test-user")
	// assert.Equal(t, stepConfig.Config["appUserPw"],  )

	// type args struct {
	// 	config *StepConfig
	// 	client vaultClient
	// }
	// tests := []struct {
	// 	name string
	// 	args args
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		resolveVaultTestCredentials(tt.args.config, tt.args.client)
	// 	})
	// }
}
