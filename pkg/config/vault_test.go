//go:build unit
// +build unit

package config

import (
	"fmt"
	"os"
	"path"
	"strconv"
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
		vaultMock := &mocks.VaultClient{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		vaultData := map[string]string{secretName: "value1"}

		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Equal(t, "value1", stepConfig.Config[secretName])
	})

	t.Run("Load secret from Vault with path override", func(t *testing.T) {
		vaultMock := &mocks.VaultClient{}
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
		vaultMock := &mocks.VaultClient{}
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
		vaultMock := &mocks.VaultClient{}
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
		vaultMock := &mocks.VaultClient{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(nil, fmt.Errorf("test"))
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Len(t, stepConfig.Config, 1)
	})

	t.Run("Secret doesn't exist", func(t *testing.T) {
		vaultMock := &mocks.VaultClient{}
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
		vaultMock := &mocks.VaultClient{}
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
		vaultMock := &mocks.VaultClient{}
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
		vaultMock := &mocks.VaultClient{}
		stepConfig := StepConfig{Config: map[string]interface{}{}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecret", secretNameOverrideKey, secretName)}
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.Nil(t, stepConfig.Config[secretName])
		vaultMock.AssertNotCalled(t, "GetKvSecret", mock.AnythingOfType("string"))
	})
}

func TestVaultSecretFiles(t *testing.T) {
	const secretName = "testSecret"
	const secretNameOverrideKey = "mySecretVaultSecretName"
	t.Run("Test Vault Secret File Reference", func(t *testing.T) {
		vaultMock := &mocks.VaultClient{}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath": "team1",
		}}
		stepParams := []StepParameters{stepParam(secretName, "vaultSecretFile", secretNameOverrideKey, secretName)}
		vaultData := map[string]string{secretName: "value1"}
		vaultMock.On("GetKvSecret", path.Join("team1", secretName)).Return(vaultData, nil)
		resolveAllVaultReferences(&stepConfig, vaultMock, stepParams)
		assert.NotNil(t, stepConfig.Config[secretName])
		path := stepConfig.Config[secretName].(string)
		contentByte, err := os.ReadFile(path)
		assert.NoError(t, err)
		content := string(contentByte)
		assert.Equal(t, "value1", content)
	})

	os.RemoveAll(VaultSecretFileDirectory)
	VaultSecretFileDirectory = ""

	t.Run("Test temporary secret file cleanup", func(t *testing.T) {
		vaultMock := &mocks.VaultClient{}
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

func TestResolveVaultTestCredentialsWrapper(t *testing.T) {
	t.Parallel()
	t.Run("Default test credential prefix", func(t *testing.T) {
		t.Parallel()
		// init
		vaultMock := &mocks.VaultClient{}
		envPrefix := "PIPER_TESTCREDENTIAL_"
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":               "team1",
			"vaultTestCredentialPath": []interface{}{"appCredentials1", "appCredentials2"},
			"vaultTestCredentialKeys": []interface{}{[]interface{}{"appUser1", "appUserPw1"}, []interface{}{"appUser2", "appUserPw2"}},
		}}

		defer os.Unsetenv("PIPER_TESTCREDENTIAL_APPUSER1")
		defer os.Unsetenv("PIPER_TESTCREDENTIAL_APPUSERPW1")
		defer os.Unsetenv("PIPER_TESTCREDENTIAL_APPUSER2")
		defer os.Unsetenv("PIPER_TESTCREDENTIAL_APPUSERPW2")

		// mock
		vaultData1 := map[string]string{"appUser1": "test-user", "appUserPw1": "password1234"}
		vaultMock.On("GetKvSecret", "team1/appCredentials1").Return(vaultData1, nil)
		vaultData2 := map[string]string{"appUser2": "test-user", "appUserPw2": "password1234"}
		vaultMock.On("GetKvSecret", "team1/appCredentials2").Return(vaultData2, nil)

		// test
		resolveVaultTestCredentialsWrapper(&stepConfig, vaultMock)

		// assert
		for k, expectedValue := range vaultData1 {
			env := envPrefix + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, expectedValue, os.Getenv(env))
		}

		// assert
		for k, expectedValue := range vaultData2 {
			env := envPrefix + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, expectedValue, os.Getenv(env))
		}
	})

	t.Run("Multiple test credential prefixes", func(t *testing.T) {
		t.Parallel()
		// init
		vaultMock := &mocks.VaultClient{}
		envPrefixes := []interface{}{"TEST1_", "TEST2_"}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":                    "team1",
			"vaultTestCredentialPath":      []interface{}{"appCredentials1", "appCredentials2"},
			"vaultTestCredentialKeys":      []interface{}{[]interface{}{"appUser", "appUserPw"}, []interface{}{"appUser", "appUserPw"}},
			"vaultTestCredentialEnvPrefix": envPrefixes,
		}}

		defer os.Unsetenv("TEST1_APPUSER")
		defer os.Unsetenv("TEST1_APPUSERPW")
		defer os.Unsetenv("TEST2_APPUSER")
		defer os.Unsetenv("TEST2_APPUSERPW")

		// mock
		vaultData1 := map[string]string{"appUser": "test-user1", "appUserPw": "password1"}
		vaultMock.On("GetKvSecret", "team1/appCredentials1").Return(vaultData1, nil)
		vaultData2 := map[string]string{"appUser": "test-user2", "appUserPw": "password2"}
		vaultMock.On("GetKvSecret", "team1/appCredentials2").Return(vaultData2, nil)

		// test
		resolveVaultTestCredentialsWrapper(&stepConfig, vaultMock)

		// assert
		for k, expectedValue := range vaultData1 {
			env := envPrefixes[0].(string) + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, expectedValue, os.Getenv(env))
		}

		// assert
		for k, expectedValue := range vaultData2 {
			env := envPrefixes[1].(string) + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, expectedValue, os.Getenv(env))
		}
	})

	t.Run("Multiple custom general purpuse credential environment prefixes", func(t *testing.T) {
		t.Parallel()
		// init
		vaultMock := &mocks.VaultClient{}
		envPrefixes := []interface{}{"CUSTOM1_", "CUSTOM2_"}
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":                "team1",
			"vaultCredentialPath":      []interface{}{"appCredentials1", "appCredentials2"},
			"vaultCredentialKeys":      []interface{}{[]interface{}{"appUser", "appUserPw"}, []interface{}{"appUser", "appUserPw"}},
			"vaultCredentialEnvPrefix": envPrefixes,
		}}

		defer os.Unsetenv("CUSTOM1_APPUSER")
		defer os.Unsetenv("CUSTOM1_APPUSERPW")
		defer os.Unsetenv("CUSTOM2_APPUSER")
		defer os.Unsetenv("CUSTOM2_APPUSERPW")

		// mock
		vaultData1 := map[string]string{"appUser": "test-user1", "appUserPw": "password1"}
		vaultMock.On("GetKvSecret", "team1/appCredentials1").Return(vaultData1, nil)
		vaultData2 := map[string]string{"appUser": "test-user2", "appUserPw": "password2"}
		vaultMock.On("GetKvSecret", "team1/appCredentials2").Return(vaultData2, nil)

		// test
		resolveVaultCredentialsWrapper(&stepConfig, vaultMock)

		// assert
		for k, expectedValue := range vaultData1 {
			env := envPrefixes[0].(string) + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, expectedValue, os.Getenv(env))
		}

		// assert
		for k, expectedValue := range vaultData2 {
			env := envPrefixes[1].(string) + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, expectedValue, os.Getenv(env))
		}
	})

	// Test empty and non-empty custom general purpose credential prefix
	envPrefixes := []string{"CUSTOM_MYCRED1_", ""}
	for idx, envPrefix := range envPrefixes {
		tEnvPrefix := envPrefix
		// this variable is used to avoid race condition, because tests are running in parallel
		// env variable with default prefix is being created for each iteration and being set and unset asynchronously
		// race condition may occur while one function sets and tries to assert if it exists but the other unsets it before it
		stIdx := strconv.Itoa(idx)
		t.Run("Custom general purpose credential prefix along with fixed standard prefix", func(t *testing.T) {
			t.Parallel()
			// init
			vaultMock := &mocks.VaultClient{}
			standardEnvPrefix := "PIPER_VAULTCREDENTIAL_"
			stepConfig := StepConfig{Config: map[string]interface{}{
				"vaultPath":                "team1",
				"vaultCredentialPath":      "appCredentials3",
				"vaultCredentialKeys":      []interface{}{"appUser3" + stIdx, "appUserPw3" + stIdx},
				"vaultCredentialEnvPrefix": tEnvPrefix,
			}}

			defer os.Unsetenv(tEnvPrefix + "APPUSER3" + stIdx)
			defer os.Unsetenv(tEnvPrefix + "APPUSERPW3" + stIdx)
			defer os.Unsetenv("PIPER_VAULTCREDENTIAL_APPUSER3" + stIdx)
			defer os.Unsetenv("PIPER_VAULTCREDENTIAL_APPUSERPW3" + stIdx)

			// mock
			vaultData := map[string]string{"appUser3" + stIdx: "test-user", "appUserPw3" + stIdx: "password1234"}
			vaultMock.On("GetKvSecret", "team1/appCredentials3").Return(vaultData, nil)

			// test
			resolveVaultCredentialsWrapper(&stepConfig, vaultMock)

			// assert
			for k, expectedValue := range vaultData {
				env := tEnvPrefix + strings.ToUpper(k)
				assert.NotEmpty(t, os.Getenv(env))
				assert.Equal(t, expectedValue, os.Getenv(env))
				standardEnv := standardEnvPrefix + strings.ToUpper(k)
				assert.NotEmpty(t, os.Getenv(standardEnv))
				assert.Equal(t, expectedValue, os.Getenv(standardEnv))
			}
		})
	}
}

func TestResolveVaultTestCredentials(t *testing.T) {
	t.Parallel()
	t.Run("Default test credential prefix", func(t *testing.T) {
		t.Parallel()
		// init
		vaultMock := &mocks.VaultClient{}
		envPrefix := "PIPER_TESTCREDENTIAL_"
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":               "team1",
			"vaultTestCredentialPath": "appCredentials",
			"vaultTestCredentialKeys": []interface{}{"appUser4", "appUserPw4"},
		}}

		defer os.Unsetenv("PIPER_TESTCREDENTIAL_APPUSER4")
		defer os.Unsetenv("PIPER_TESTCREDENTIAL_APPUSERPW4")

		// mock
		vaultData := map[string]string{"appUser4": "test-user", "appUserPw4": "password1234"}
		vaultMock.On("GetKvSecret", "team1/appCredentials").Return(vaultData, nil)

		// test
		resolveVaultTestCredentials(&stepConfig, vaultMock)

		// assert
		for k, expectedValue := range vaultData {
			env := envPrefix + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, expectedValue, os.Getenv(env))
		}
	})

	// Test empty and non-empty custom general purpose credential prefix
	envPrefixes := []string{"CUSTOM_MYCRED_", ""}
	for idx, envPrefix := range envPrefixes {
		tEnvPrefix := envPrefix
		// this variable is used to avoid race condition, because tests are running in parallel
		// env variable with default prefix is being created for each iteration and being set and unset asynchronously
		// race condition may occur while one function sets and tries to assert if it exists but the other unsets it before it
		stIdx := strconv.Itoa(idx)
		t.Run("Custom general purpose credential prefix along with fixed standard prefix", func(t *testing.T) {
			t.Parallel()
			// init
			vaultMock := &mocks.VaultClient{}
			standardEnvPrefix := "PIPER_VAULTCREDENTIAL_"
			stepConfig := StepConfig{Config: map[string]interface{}{
				"vaultPath":                "team1",
				"vaultCredentialPath":      "appCredentials",
				"vaultCredentialKeys":      []interface{}{"appUser5" + stIdx, "appUserPw5" + stIdx},
				"vaultCredentialEnvPrefix": tEnvPrefix,
			}}

			defer os.Unsetenv(tEnvPrefix + "APPUSER5" + stIdx)
			defer os.Unsetenv(tEnvPrefix + "APPUSERPW5" + stIdx)
			defer os.Unsetenv("PIPER_VAULTCREDENTIAL_APPUSER5" + stIdx)
			defer os.Unsetenv("PIPER_VAULTCREDENTIAL_APPUSERPW5" + stIdx)

			// mock
			vaultData := map[string]string{"appUser5" + stIdx: "test-user", "appUserPw5" + stIdx: "password1234"}
			vaultMock.On("GetKvSecret", "team1/appCredentials").Return(vaultData, nil)

			// test
			resolveVaultCredentials(&stepConfig, vaultMock)

			// assert
			for k, expectedValue := range vaultData {
				env := tEnvPrefix + strings.ToUpper(k)
				assert.NotEmpty(t, os.Getenv(env))
				assert.Equal(t, expectedValue, os.Getenv(env))
				standardEnv := standardEnvPrefix + strings.ToUpper(k)
				assert.NotEmpty(t, os.Getenv(standardEnv))
				assert.Equal(t, expectedValue, os.Getenv(standardEnv))
			}
		})
	}

	t.Run("Custom test credential prefix", func(t *testing.T) {
		t.Parallel()
		// init
		vaultMock := &mocks.VaultClient{}
		envPrefix := "CUSTOM_CREDENTIAL_"
		stepConfig := StepConfig{Config: map[string]interface{}{
			"vaultPath":                    "team1",
			"vaultTestCredentialPath":      "appCredentials",
			"vaultTestCredentialKeys":      []interface{}{"appUser6", "appUserPw6"},
			"vaultTestCredentialEnvPrefix": envPrefix,
		}}

		defer os.Unsetenv("CUSTOM_CREDENTIAL_APPUSER6")
		defer os.Unsetenv("CUSTOM_CREDENTIAL_APPUSERPW6")

		// mock
		vaultData := map[string]string{"appUser6": "test-user", "appUserPw6": "password1234"}
		vaultMock.On("GetKvSecret", "team1/appCredentials").Return(vaultData, nil)

		// test
		resolveVaultTestCredentials(&stepConfig, vaultMock)

		// assert
		for k, expectedValue := range vaultData {
			env := envPrefix + strings.ToUpper(k)
			assert.NotEmpty(t, os.Getenv(env))
			assert.Equal(t, expectedValue, os.Getenv(env))
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
			if got := ConvertEnvVar(tt.args.s); got != tt.want {
				t.Errorf("convertEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}
