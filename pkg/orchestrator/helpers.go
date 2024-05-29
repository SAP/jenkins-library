package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"os"
)

// envVarsAreSet verifies if any envvar from the list has nona non-empty, non-false value
func envVarsAreSet(envVars []string) bool {
	for _, v := range envVars {
		if envVarIsTrue(v) {
			return true
		}
	}
	return false
}

// envVarIsTrue verifies if the variable is set and has a non-empty, non-false value.
func envVarIsTrue(key string) bool {
	val, exists := os.LookupEnv(key)
	if !exists {
		return false
	}
	if len(val) == 0 || val == "no" || val == "false" || val == "off" || val == "0" {
		return false
	}

	return true
}

// Wrapper function to read env variable and set default value
func getEnv(key, fallback string) string {
	if value, found := os.LookupEnv(key); found {
		log.Entry().Debugf("For: %s, found: %s", key, value)
		return value
	}

	log.Entry().Debugf("Could not read env variable %v using fallback value %v", key, fallback)
	return fallback
}
