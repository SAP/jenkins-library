package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func npmExecuteEndToEndTests(config npmExecuteEndToEndTestsOptions, _ *telemetry.CustomData) {
	c := command.Command{}

	c.Stdout(log.Writer())
	c.Stderr(log.Writer())
	runNpmExecuteEndToEndTests(config, &c)
}

func runNpmExecuteEndToEndTests(config npmExecuteEndToEndTestsOptions, c command.ExecRunner) {

	log.Entry().Info("url value is %v", config.AppURLs[0])

	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().WithError(err).Warning("Cannot infer config from CI environment")
		return
	}
	env := provider.Branch()
	if config.OnlyRunInProductiveBranch && config.ProductiveBranch != env {
		log.Entry().Info("Skipping execution because it is configured to run only in the productive branch.")
		return
	}
	if config.Wdi5 {
		// install wdi5 and all required WebdriverIO peer dependencies
		// add a config file (wdio.conf.js) to your current working directory, using http://localhost:8080/index.html as baseUrl,
		// looking for tests in $ui5-app/webapp/test/**/* that follow the name pattern *.test.js
		// set an npm script named “wdi5” to run wdi5 so you can immediately do npm run wdi5
		if err := c.RunExecutable("npm", "init", "wdi5@latest"); err != nil {
			log.Entry().WithError(err).Fatal("Failed to install wdi5")
			return
		}
	}
	if len(config.AppURLs) > 0 {
		for _, appUrl := range config.AppURLs {
			url := appUrl["url"].(string)
			parameters := appUrl["parameters"].([]string)
			runEndToEndTestForUrl(url, parameters, config, c)
		}
		return
	}
	runEndToEndTestForUrl(config.BaseURL, []string{}, config, c)
}

func runEndToEndTestForUrl(url string, params []string, config npmExecuteEndToEndTestsOptions, command command.ExecRunner) {
	log.Entry().Infof("Running end to end tests for URL: %s", url)

	urlParam := "--baseUrl="
	if len(config.AppURLs) > 0 {
		urlParam = "--launchUrl="
	}

	// Prepare script options
	scriptOptions := []string{urlParam + config.BaseURL}
	if len(params) > 0 {
		scriptOptions = append(scriptOptions, params...)
	}
	if config.Wdi5 {
		if err := command.RunExecutable("npm", "run", "wdi5"); err != nil {
			log.Entry().WithError(err).Fatal("Failed to execute wdi5")
		}
		return
	}

	// Execute the npm script
	err := command.RunExecutable("npm", append([]string{config.RunScript}, scriptOptions...)...)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute end to end tests")
	}
}
