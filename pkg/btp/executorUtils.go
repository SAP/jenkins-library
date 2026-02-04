package btp

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func GetErrorInfos(value string) (BTPErrorData, string, error) {
	var errorBlock, err = extractLastErrorBlock(value)

	if errorBlock != "" && err == nil {
		// Try to extract more specific error information
		res, err := GetJSON(errorBlock)
		if err == nil {
			errorData := BTPErrorData{}

			err := json.Unmarshal([]byte(res), &errorData)
			if err != nil {
				return errorData, "", err
			}

			errorMessageCode := mapErrorMessageToCode(errorData.Description)

			return errorData, errorMessageCode, nil
		}
	}
	return BTPErrorData{}, "", errors.New("no Error block found")
}

func extractLastErrorBlock(value string) (string, error) {
	var responseMaps = strings.Split(value, "Response mapping")
	var input = responseMaps[len(responseMaps)-1]

	var re = regexp.MustCompile(`\{[\s]*?"error"\s*:\s*".*?"[\s\S]*?"description"\s*:\s*".*?"[\s\S]*?\}`)

	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return "", errors.New("no Error block found")
	}

	// Last match, first capturing group
	lastMatch := matches[len(matches)-1][0]
	return lastMatch, nil
}

func mapErrorMessageToCode(message string) string {
	if regexp.MustCompile(`(?i)Found multiple service bindings with the name`).MatchString(message) {
		return "MULTIPLE_BINDINGS_FOUND"
	} else if regexp.MustCompile(`(?i)binding with same name exists for instance`).MatchString(message) {
		return "BINDING_ALREADY_EXISTS"
	} else if regexp.MustCompile(`(?i)Could not find such (service)? instance`).MatchString(message) {
		return "SERVICE_INSTANCE_NOT_FOUND"
	} else if regexp.MustCompile(`(?i)Could not find such (service)? binding`).MatchString(message) {
		return "SERVICE_BINDING_NOT_FOUND"
	} else if regexp.MustCompile(`(?i)instance with same name exists for the current tenant`).MatchString(message) {
		return "INSTANCE_ALREADY_EXISTS"
	} else {
		return ""
	}
}
