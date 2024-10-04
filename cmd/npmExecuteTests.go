package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
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
	var appURLs []AppURL
	err := json.Unmarshal([]byte(config.AppURLs), &appURLs)
	if err != nil {
		return fmt.Errorf("failed to parse app URLs: %w", err)
	}

	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		return fmt.Errorf("failed to get orchestrator config provider: %w", err)
	}

	env := provider.Branch()
	if config.OnlyRunInProductiveBranch && config.ProductiveBranch != env {
		log.Entry().Info("Skipping execution because it is configured to run only in the productive branch.")
		return nil
	}

	installCommandTokens := strings.Fields(config.InstallCommand)
	if err := c.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...); err != nil {
		return fmt.Errorf("failed to execute install command: %w", err)
	}

	for _, appUrl := range appURLs {
		credentialsToEnv(appUrl.Username, appUrl.Password, config.Wdi5)
		err := runTestForUrl(appUrl.URL, config, c)
		if err != nil {
			return err
		}
	}

	if err := runTestForUrl(config.BaseURL, config, c); err != nil {
		return err
	}
	return nil
}

func runTestForUrl(url string, config *npmExecuteTestsOptions, command command.ExecRunner) error {
	log.Entry().Infof("Running end to end tests for URL: %s", url)

	if config.Wdi5 {
		// install wdi5 and all required WebdriverIO peer dependencies
		// add a config file (wdio.conf.js) to your current working directory, using http://localhost:8080/index.html as baseUrl,
		// looking for tests in $ui5-app/webapp/test/**/* that follow the name pattern *.test.js
		// set an npm script named “wdi5” to run wdi5 so you can immediately do npm run wdi5
		if err := command.RunExecutable("npm", "init", "wdi5@latest", "--baseUrl", url); err != nil {
			return fmt.Errorf("failed to install wdi5: %w", err)
		}
		if err := command.RunExecutable("npm", "run", "wdi5"); err != nil {
			return fmt.Errorf("failed to execute wdi5: %w", err)
		}
		return nil
	}

	// Execute the npm script
	options := "--baseUrl_" + url
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
