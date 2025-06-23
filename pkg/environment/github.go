package environment

import "os"

// IsGitHubActions detects if we're running in GitHub Actions environment
func IsGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	for _, v := range envVars {
		if val, exists := os.LookupEnv(v); exists && val == "true" {
			return true
		}
	}
	return false
}
