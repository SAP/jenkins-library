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
	found := false
	for _, v := range envVars {
		found = truthy(v)
	}
	return found
}

// Checks if var is set and neither empty nor false
func truthy(key string) bool {
	val, exists := os.LookupEnv(key)
	if !exists {
		return false
	}
	if val == "no" || val == "false" || val == "off" || val == "0" || len(val) == 0 {
		return false
	}

	return true
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
