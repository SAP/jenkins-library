package cmd

import (
	"encoding/json"
	"os"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	baseURLParam   = "--baseUrl="
	launchURLParam = "--launchUrl="
)

func npmExecuteEndToEndTests(config npmExecuteEndToEndTestsOptions, _ *telemetry.CustomData) {
	c := command.Command{}

	c.Stdout(log.Writer())
	c.Stderr(log.Writer())
	runNpmExecuteEndToEndTests(config, &c)
}

func runNpmExecuteEndToEndTests(config npmExecuteEndToEndTestsOptions, c command.ExecRunner) {
	type AppURL struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var appURLs []AppURL
	err := json.Unmarshal([]byte(config.AppURLsVault), &appURLs)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to unmarshal appURLsVault")
		return
	}

	log.Entry().Infof("vault value is %v, len is %d", appURLs, len(appURLs))

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

	for _, appUrl := range appURLs {
		credentialsToEnv(appUrl.Username, appUrl.Password, config.Wdi5)
		runEndToEndTestForUrl(appUrl.URL, config, c, launchURLParam)
	}

	for _, appUrl := range config.AppURLs {
		url := appUrl["url"].(string)
		runEndToEndTestForUrl(url, config, c, launchURLParam)
	}

	runEndToEndTestForUrl(config.BaseURL, config, c, baseURLParam)
}

func runEndToEndTestForUrl(url string, config npmExecuteEndToEndTestsOptions, command command.ExecRunner, urlParam string) {
	log.Entry().Infof("Running end to end tests for URL: %s", url)

	if config.Wdi5 {
		if err := command.RunExecutable("npm", "run", "wdi5"); err != nil {
			log.Entry().WithError(err).Fatal("Failed to execute wdi5")
		}
		return
	}
	// Install npm dependencies
	if err := command.RunExecutable("npm", "install"); err != nil {
		log.Entry().WithError(err).Fatal("Failed to install npm dependencies")
	}

	// Execute the npm script
	scriptOptions := []string{urlParam + url}
	if err := command.RunExecutable("npm", append([]string{"run", config.RunScript}, scriptOptions...)...); err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute end to end tests")
	}
}

func credentialsToEnv(username, password string, wdi5 bool) {
	prefix := "e2e"
	if wdi5 {
		prefix = "wdi5"
	}
	os.Setenv(prefix+"_username", username)
	os.Setenv(prefix+"_password", password)
}
