package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type vaultUrl struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

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
	if len(config.Envs) > 0 {
		c.SetEnv(config.Envs)
	}

	if len(config.Paths) > 0 {
		path := fmt.Sprintf("PATH=%s:%s", os.Getenv("PATH"), strings.Join(config.Paths, ":"))
		c.SetEnv([]string{path})
	}

	if config.WorkingDirectory != "" {
		if err := os.Chdir(config.WorkingDirectory); err != nil {
			return fmt.Errorf("failed to change directory: %w", err)
		}
	}

	installCommandTokens := strings.Fields(config.InstallCommand)
	if err := c.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...); err != nil {
		return fmt.Errorf("failed to execute install command: %w", err)
	}

	parsedURLs, err := parseURLs(config.URLs)
	if err != nil {
		return err
	}

	for _, app := range parsedURLs {
		if err := runTestForUrl(app.URL, app.Username, app.Password, config, c); err != nil {
			return err
		}
	}

	if err := runTestForUrl(config.BaseURL, config.Username, config.Password, config, c); err != nil {
		return err
	}
	return nil
}

func runTestForUrl(url, username, password string, config *npmExecuteTestsOptions, command command.ExecRunner) error {
	credentialsToEnv(username, password, config.UsernameEnvVar, config.PasswordEnvVar, command)
	// we need to reset the env vars as the next test might not have any credentials
	defer resetCredentials(config.UsernameEnvVar, config.PasswordEnvVar, command)

	runScriptTokens := strings.Fields(config.RunCommand)
	if config.UrlOptionPrefix != "" {
		runScriptTokens = append(runScriptTokens, config.UrlOptionPrefix+url)
	}
	if err := command.RunExecutable(runScriptTokens[0], runScriptTokens[1:]...); err != nil {
		return fmt.Errorf("failed to execute npm script: %w", err)
	}

	return nil
}

func parseURLs(urls []map[string]interface{}) ([]vaultUrl, error) {
	parsedUrls := []vaultUrl{}

	for _, url := range urls {
		parsedUrl := vaultUrl{}
		urlStr, ok := url["url"].(string)
		if !ok {
			return nil, fmt.Errorf("url field is not a string")
		}
		parsedUrl.URL = urlStr
		if username, ok := url["username"].(string); ok {
			parsedUrl.Username = username
		}

		if password, ok := url["password"].(string); ok {
			parsedUrl.Password = password
		}
		parsedUrls = append(parsedUrls, parsedUrl)
	}
	return parsedUrls, nil
}

func credentialsToEnv(username, password, usernameEnv, passwordEnv string, c command.ExecRunner) {
	if username == "" || password == "" {
		return
	}
	c.SetEnv([]string{usernameEnv + "=" + username, passwordEnv + "=" + password})
}

func resetCredentials(usernameEnv, passwordEnv string, c command.ExecRunner) {
	c.SetEnv([]string{usernameEnv + "=", passwordEnv + "="})
}
