package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func npmExecuteTests(config npmExecuteTestsOptions, _ *telemetry.CustomData) {
	c := command.Command{}

	c.Stdout(log.Writer())
	c.Stderr(log.Writer())
	err := runNpmExecuteTests(&config, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("Step execution failed")
	}
}

func runNpmExecuteTests(config *npmExecuteTestsOptions, c command.ExecRunner) error {
	type AppURL struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	appURLs := make(map[string]AppURL)
	urlsRaw, ok := config.VaultMetadata["urls"].([]interface{})
	if ok {
		for _, urlRaw := range urlsRaw {
			urlMap := urlRaw.(map[string]interface{})
			url := urlMap["url"].(string)
			appURLs[url] = AppURL{
				URL:      url,
				Username: urlMap["username"].(string),
				Password: urlMap["password"].(string),
			}
		}
	}

	installCommandTokens := strings.Fields(config.InstallCommand)
	if err := c.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...); err != nil {
		return fmt.Errorf("failed to execute install command: %w", err)
	}

	isWdi5 := strings.Contains(config.RunScript, "wdi5")
	for _, appUrl := range appURLs {
		credentialsToEnv(appUrl.Username, appUrl.Password, isWdi5)
		err := runTestForUrl(appUrl.URL, config, c)
		if err != nil {
			return err
		}
	}

	username := config.AppSecrets["username"].(string)
	password := config.AppSecrets["password"].(string)
	credentialsToEnv(username, password, isWdi5)
	if err := runTestForUrl(config.BaseURL, config, c); err != nil {
		return err
	}
	return nil
}

func runTestForUrl(url string, config *npmExecuteTestsOptions, command command.ExecRunner) error {
	log.Entry().Infof("Running end to end tests for URL: %s", url)

	// Execute the npm script
	options := "--baseUrl=" + url
	runScriptTokens := strings.Fields(config.RunScript)
	if err := command.RunExecutable(runScriptTokens[0], append(runScriptTokens[1:], options)...); err != nil {
		return fmt.Errorf("failed to execute npm script: %w", err)
	}
	return nil
}

func credentialsToEnv(username, password string, wdi5 bool) {
	prefix := "e2e"
	if wdi5 {
		prefix = "wdi5"
	}
	os.Setenv(prefix+"_username", username)
	os.Setenv(prefix+"_password", password)
}
