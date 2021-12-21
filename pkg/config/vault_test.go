package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/SAP/jenkins-library/pkg/config/mocks"
	"github.com/stretchr/testify/assert"
)

func TestVaultConfigLoad(t *testing.T) {
	const secretName = "testSecret"
	const secretNameOverrideKey = "mySecretVaultSecretName"
	t.Parallel()
	t.Run("Load secret from vault", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		vaultData := map[string]string{secretName: "value1"}

		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("Load secret from vault with path override", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":           "team1",
			secretNameOverrideKey: "overrideSecretName",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		vaultData := map[string]string{secretName: "value1"}

		vaultMock.On("GetKvSecret", path.Join("team1", "overrideSecretName")).Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("Secrets are not overwritten", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":             "team1",
			secretName:              "preset value",
			"vaultDisableOverwrite": true,
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)

		assert.Equal(t, "preset value", stepConfig.Config[secretName])
	})

	t.Run("Secrets can be overwritten", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
			secretName:  "preset value",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)

		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("Error is passed through", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(nil, fmt.Errorf("test"))
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Len(t, stepConfig.Config, 1)
	})

	t.Run("Secret doesn't exist", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(nil, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Len(t, stepConfig.Config, 1)
	})

	t.Run("Alias names should be considered", func(t *testing.T) {
		aliasName := "alias"
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		param := stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)
		addAlias(&param, aliasName)
		stepParams := []StepParameters{param}
		vaultData := map[string]string{aliasName: "value1"}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("Search over multiple paths", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultBasePath": "team2",
			"vaultPath":     "team1",
		}}
		stepParams := []StepParameters{
			stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName),
		}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(nil, nil)
		vaultMock.On("GetKvSecret", path.Join("team2/GROUP-SECRETS", secretName)).Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("No BasePath is stepConfig.Configured", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, nil, stepConfig.Config[secretName])
		vaultMock.AssertNotCalled(t, "GetKvSecret", mock.AnythingOfType("string"))
	})
}

func TestVaultSecretFiles(t *testing.T) {
	const secretName = "testSecret"
	const secretNameOverrideKey = "mySecretVaultSecretName"
	t.Run("Test Vault Secret File Reference", func(t *testing.T) {
		vaultMock := &mocks.VaultMock{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecretFile", secretNameOverrideKey, secretName)}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(vaultData, nil)
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
		stepParams := []StepParameters{stepParam(secretName, "vaultSecretFile", secretNameOverrideKey, secretName)}
		vaultData := map[string]string{secretName: "value1"}
		assert.NoDirExists(t, VaultSecretFileDirectory)
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(vaultData, nil)
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

	config.mixinVaultConfig(nil, general, steps)

	assert.Contains(t, config.Config, "vaultServerUrl")
	assert.Equal(t, vaultServerUrl, config.Config["vaultServerUrl"])
	assert.Contains(t, config.Config, "vaultPath")
	assert.Equal(t, vaultPath, config.Config["vaultPath"])
	assert.NotContains(t, config.Config, "unknownConfig")

}

func stepParam(name, refType, vaultSecretNameProperty, defaultSecretNameName string) StepParameters {
	return StepParameters{
		Name:    name,
		Aliases: []Alias{},
		ResourceRef: []ResourceReference{
			{
				Type:    refType,
				Name:    vaultSecretNameProperty,
				Default: defaultSecretNameName,
			},
		},
	}
}

func addAlias(param *StepParameters, aliasName string) {
	alias := Alias{Name: aliasName}
	param.Aliases = append(param.Aliases, alias)
}

func TestResolveVaultTestCredentials(t *testing.T) {
	t.Parallel()
	t.Run("Default test credential prefix", func(t *testing.T) {
		t.Parallel()
		// init
		vaultMock := &mocks.VaultMock{}
		envPrefix := "PIPER_TESTCREDENTIAL_"
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":               "team1",
			"vaultTestCredentialPath": "appCredentials",
			"vaultTestCredentialKeys": []interface{}{"appUser", "appUserPw"},
		}}

		defer os.Unsetenv("PIPER_TESTCREDENTIAL_APPUSER")
		defer os.Unsetenv("PIPER_TESTCREDENTIAL_APPUSERPW")

		// mock
		vaultData := map[string]string{"appUser": "test-user", "appUserPw": "password1234"}
		vaultMock.On("GetKvSecret", "team1/appCredentials").Return(vaultData, nil)

		// test
		resolveVaultTestCredentials(&stepConfig, vaultMock)

		// assert
		for k, v := range vaultData {
			env := envPrefix + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, os.Getenv(env), v)
		}
	})

	t.Run("Custom general purpose credential prefix along with fixed standard prefix", func(t *testing.T) {
		t.Parallel()
		// init
		vaultMock := &mocks.VaultMock{}
		envPrefix := "CUSTOM_MYCRED_"
		standardEnvPrefix := "PIPER_VAULTCREDENTIAL_"
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":                "team1",
			"vaultCredentialPath":      "appCredentials",
			"vaultCredentialKeys":      []interface{}{"appUser", "appUserPw"},
			"vaultCredentialEnvPrefix": envPrefix,
		}}

		defer os.Unsetenv("CUSTOM_MYCRED_APPUSER")
		defer os.Unsetenv("CUSTOM_MYCRED_APPUSERPW")
		defer os.Unsetenv("PIPER_VAULTCREDENTIAL_APPUSER")
		defer os.Unsetenv("PIPER_VAULTCREDENTIAL_APPUSERPW")

		// mock
		vaultData := map[string]string{"appUser": "test-user", "appUserPw": "password1234"}
		vaultMock.On("GetKvSecret", "team1/appCredentials").Return(vaultData, nil)

		// test
		resolveVaultCredentials(&stepConfig, vaultMock)

		// assert
		for k, v := range vaultData {
			env := envPrefix + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, os.Getenv(env), v)
			standardEnv := standardEnvPrefix + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(standardEnv))
			assert.Equal(t, os.Getenv(standardEnv), v)
		}
	})

	t.Run("Custom test credential prefix", func(t *testing.T) {
		t.Parallel()
		// init
		vaultMock := &mocks.VaultMock{}
		envPrefix := "CUSTOM_CREDENTIAL_"
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":                    "team1",
			"vaultTestCredentialPath":      "appCredentials",
			"vaultTestCredentialKeys":      []interface{}{"appUser", "appUserPw"},
			"vaultTestCredentialEnvPrefix": envPrefix,
		}}

		defer os.Unsetenv("CUSTOM_CREDENTIAL_APPUSER")
		defer os.Unsetenv("CUSTOM_CREDENTIAL_APPUSERPW")

		// mock
		vaultData := map[string]string{"appUser": "test-user", "appUserPw": "password1234"}
		vaultMock.On("GetKvSecret", "team1/appCredentials").Return(vaultData, nil)

		// test
		resolveVaultTestCredentials(&stepConfig, vaultMock)

		// assert
		for k, v := range vaultData {
			env := envPrefix + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, os.Getenv(env), v)
		}
	})
}

func Test_convertEnvVar(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{""},
			want: "",
		},
		{
			name: "alphanumerical string",
			args: args{"myApp1"},
			want: "MYAPP1",
		},
		{
			name: "string with hyphen",
			args: args{"my_App-1"},
			want: "MY_APP_1",
		},
		{
			name: "string with special characters",
			args: args{"my_App?-(1]"},
			want: "MY_APP_1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertEnvVar(tt.args.s); got != tt.want {
				t.Errorf("convertEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}
