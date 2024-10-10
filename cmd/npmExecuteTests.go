package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type parsedMetadata struct {
	GlobalUsername string
	GlobalPassword string
	URLs           []appUrl
}

type appUrl struct {
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

	installCommandTokens := strings.Fields(config.InstallCommand)
	if err := c.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...); err != nil {
		return fmt.Errorf("failed to execute install command: %w", err)
	}

	parsedMetadata, err := parseMetadata(config.VaultMetadata)
	if err != nil {
		return err
	}

	for _, app := range parsedMetadata.URLs {
		if err := runTestForUrl(app.URL, app.Username, app.Password, config, c); err != nil {
			return err
		}
	}

	if err := runTestForUrl(config.BaseURL, parsedMetadata.GlobalUsername, parsedMetadata.GlobalPassword, config, c); err != nil {
		return err
	}
	return nil
}

func runTestForUrl(url, username, password string, config *npmExecuteTestsOptions, command command.ExecRunner) error {
	log.Entry().Infof("Running end to end tests for URL: %s", url)

	credentialsToEnv(username, password, config.UsernameEnvVar, config.PasswordEnvVar, command)
	runScriptTokens := strings.Fields(config.RunCommand)
	if config.UrlOptionPrefix != "" {
		runScriptTokens = append(runScriptTokens, config.UrlOptionPrefix+url)
	}
	if err := command.RunExecutable(runScriptTokens[0], runScriptTokens[1:]...); err != nil {
		return fmt.Errorf("failed to execute npm script: %w", err)
	}

	// we need to reset the env vars as the next test might not have any credentials
	resetCredentials(config.UsernameEnvVar, config.PasswordEnvVar, command)
	return nil
}

func parseMetadata(metadata map[string]interface{}) (*parsedMetadata, error) {
	parsedMetadata := &parsedMetadata{
		URLs: []appUrl{},
	}

	if metadata != nil {
		if urls, ok := metadata["urls"].([]interface{}); ok {
			for _, url := range urls {
				urlMap, ok := url.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("failed to parse vault metadata: 'urls' entry is not a map")
				}

				app := appUrl{}
				if u, ok := urlMap["url"].(string); ok {
					app.URL = u
				} else {
					return nil, fmt.Errorf("failed to parse vault metadata: 'url' field is not a string")
				}

				if username, ok := urlMap["username"].(string); ok {
					app.Username = username
				}

				if password, ok := urlMap["password"].(string); ok {
					app.Password = password
				}

				parsedMetadata.URLs = append(parsedMetadata.URLs, app)
			}
		}

		if username, ok := metadata["username"].(string); ok {
			parsedMetadata.GlobalUsername = username
		}
		if password, ok := metadata["password"].(string); ok {
			parsedMetadata.GlobalPassword = password
		}
	}

	return parsedMetadata, nil
}

func credentialsToEnv(username, password, usernameEnv, passwordEnv string, c command.ExecRunner) {
	if username == "" || password == "" {
		log.Entry().Warnf("Missing credentials: username: %s, password: %s", username, password)
		return
	}
	c.SetEnv([]string{usernameEnv + "=" + username, passwordEnv + "=" + password})
}

func resetCredentials(usernameEnv, passwordEnv string, c command.ExecRunner) {
	c.SetEnv([]string{usernameEnv + "=", passwordEnv + "="})
}
