package cmd

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

const scanCommand = "sonar-scanner"

func sonarExecuteScan(options sonarExecuteScanOptions) error {
	c := command.Command{}
	// reroute command output to loging framework
	// also log stdout as Karma reports into it
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	runSonar(options, &c)
	return nil
}

func runSonar(options sonarExecuteScanOptions, command execRunner) {

	arguments := []string{}

	if len(options.Organization) > 0 {
		arguments = append(arguments, "sonar.organization="+options.Organization)
	}

	if len(options.ProjectVersion) > 0 {
		arguments = append(arguments, "sonar.projectVersion="+options.ProjectVersion)
	}

	//if(configuration.options instanceof String)
	//configuration.options = [].plus(configuration.options)

	//config.options.add("sonar.login=$SONAR_TOKEN")
	//config.options.add("sonar.host.url=$SONAR_HOST")

	if len(options.LegacyPRHandling) > 0 { //&& options.LegacyPRHandling {
		// see https://docs.sonarqube.org/display/PLUG/GitHub+Plugin
		arguments = append(arguments, "sonar.analysis.mode=preview")
		arguments = append(arguments, "sonar.github.oauth=$GITHUB_TOKEN")
		arguments = append(arguments, "sonar.github.pullRequest=${env.CHANGE_ID}")
		arguments = append(arguments, "sonar.github.repository=${config.githubOrg}/${config.githubRepo}")
		arguments = append(arguments, "sonar.analysis.mode=preview")
		if len(options.GithubAPIURL) > 0 {
			arguments = append(arguments, "sonar.github.endpoint={{ options.githubApiUrl }}")
		}
		if len(options.DisableInlineComments) > 0 {
			arguments = append(arguments, "sonar.github.disableInlineComments={{ options.disableInlineComments }}")
		}
	} else {
		// see https://sonarcloud.io/documentation/analysis/pull-request/
		arguments = append(arguments, "sonar.pullrequest.key={{ env.CHANGE_ID }}")
		arguments = append(arguments, "sonar.pullrequest.base={{ env.CHANGE_TARGET }}")
		arguments = append(arguments, "sonar.pullrequest.branch={{ env.CHANGE_BRANCH }}")
		arguments = append(arguments, "sonar.pullrequest.provider={{ options.pullRequestProvider }}")
		/*if options.PullRequestProvider == "GitHub" {
			arguments = append(arguments, "sonar.pullrequest.github.repository={{ options.githubOrg }}/{{ optiojns.githubRepo }}")
		} else {
			log.Entry().Fatal("Pull-Request provider '{{ options.pullRequestProvider }}' is not supported!")
		}*/
	}

	scan(scanCommand, arguments, command)
}

func scan(scanCommand string, options []string, command execRunner) {
	for idx, element := range options {
		element = strings.TrimSpace(element)
		if !strings.HasPrefix(element, "-D") {
			element = "-D" + element
		}
		options[idx] = element
	}
	log.Entry().
		WithField("command", scanCommand+" "+strings.Join(options, " ")).
		Debug("executing sonar scan command")

	err := command.RunExecutable(scanCommand, strings.Join(options, " "))
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("failed to execute scan command")
	}
}
