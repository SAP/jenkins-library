package piperutils

import (
	"errors"
	"os"
)

type Orchestrator int

const (
	AzureDevOps Orchestrator = iota
	GitHubActions
	Jenkins
	Travis
)

func (o Orchestrator) String() string {
	return [...]string{"AzureDevOps", "GitHubActions", "Travis", "Jenkins"}[o]
}

func DetectOrchestrator() (Orchestrator, error) {
	if isAzure() {
		return Orchestrator(AzureDevOps), nil
	} else if isGitHubActions() {
		return Orchestrator(GitHubActions), nil
	} else if isTravis() {
		return Orchestrator(Travis), nil
	} else if isJenkins() {
		return Orchestrator(Jenkins), nil
	} else {
		return -2, errors.New("could not detect orchestrator. Supported is: Azure DevOps, GitHub Actions, Travis, Jenkins")
	}
}

func areIndicatingEnvVarsSet(envVars []string) bool {
	for _, v := range envVars {
		_, exists := os.LookupEnv(v)
		if exists {
			return true
		}
	}

	return false
}

func isAzure() bool {
	envVars := []string{"AZURE_HTTP_USER_AGENT"}
	return areIndicatingEnvVarsSet(envVars)
}

func isGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	return areIndicatingEnvVarsSet(envVars)
}

func isTravis() bool {
	envVars := []string{"TRAVIS"}
	return areIndicatingEnvVarsSet(envVars)
}

func isJenkins() bool {
	envVars := []string{"JENKINS_HOME", "JENKINS_URL"}
	return areIndicatingEnvVarsSet(envVars)
}
