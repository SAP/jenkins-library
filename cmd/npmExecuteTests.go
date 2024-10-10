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
	type AppUnderTest struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	apps := []AppUnderTest{}
	urlsRaw, ok := config.VaultMetadata["urls"].([]interface{})
	if ok {
		for _, urlRaw := range urlsRaw {
			urlMap := urlRaw.(map[string]interface{})
			app := AppUnderTest{
				URL:      urlMap["url"].(string),
				Username: urlMap["username"].(string),
				Password: urlMap["password"].(string),
			}
			apps = append(apps, app)
		}
	}

	if len(config.EnvVars) > 0 {
		c.SetEnv(config.EnvVars)
	}

	if len(config.Paths) > 0 {
		path := fmt.Sprintf("PATH=%s:%s", os.Getenv("PATH"), strings.Join(config.Paths, ":"))
		c.SetEnv([]string{path})
	}

	installCommandTokens := strings.Fields(config.InstallCommand)
	if err := c.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...); err != nil {
		return fmt.Errorf("failed to execute install command: %w", err)
	}

	for _, app := range apps {
		credentialsToEnv(app.Username, app.Password, config.UsernameEnvVar, config.PasswordEnvVar, c)
		err := runTestForUrl(app.URL, config, c)
		if err != nil {
			return err
		}
	}

	username := config.VaultMetadata["username"].(string)
	password := config.VaultMetadata["password"].(string)
	credentialsToEnv(username, password, config.UsernameEnvVar, config.PasswordEnvVar, c)
	if err := runTestForUrl(config.BaseURL, config, c); err != nil {
		return err
	}
	return nil
}

func runTestForUrl(url string, config *npmExecuteTestsOptions, command command.ExecRunner) error {
	log.Entry().Infof("Running end to end tests for URL: %s", url)

	runScriptTokens := strings.Fields(config.RunCommand)
	if config.UrlOptionPrefix != "" {
		runScriptTokens = append(runScriptTokens, config.UrlOptionPrefix+url)
	}
	if err := command.RunExecutable(runScriptTokens[0], runScriptTokens[1:]...); err != nil {
		return fmt.Errorf("failed to execute npm script: %w", err)
	}
	return nil
}

func credentialsToEnv(username, password, usernameEnv, passwordEnv string, c command.ExecRunner) {
	if username == "" || password == "" {
		return
	}
	c.SetEnv([]string{usernameEnv + "=" + username, passwordEnv + "=" + password})
}
