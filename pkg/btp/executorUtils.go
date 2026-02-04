package btp

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	errorBlockRegex       = regexp.MustCompile(`\{[\s]*?"error"\s*:\s*".*?"[\s\S]*?"description"\s*:\s*".*?"[\s\S]*?\}`)
	multipleBindingsRegex = regexp.MustCompile(`(?i)found multiple service bindings with the name`)
	bindingExistsRegex    = regexp.MustCompile(`(?i)binding with same name exists for instance`)
	instanceNotFoundRegex = regexp.MustCompile(`(?i)could not find such (?:service )?instance`)
	bindingNotFoundRegex  = regexp.MustCompile(`(?i)could not find such (?:service )?binding`)
	instanceExistsRegex   = regexp.MustCompile(`(?i)instance with same name exists for the current tenant`)
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

	matches := errorBlockRegex.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return "", errors.New("no Error block found")
	}

	// Last match, first capturing group
	lastMatch := matches[len(matches)-1][0]
	return lastMatch, nil
}

func mapErrorMessageToCode(message string) string {
	if multipleBindingsRegex.MatchString(message) {
		return "MULTIPLE_BINDINGS_FOUND"
	} else if bindingExistsRegex.MatchString(message) {
		return "BINDING_ALREADY_EXISTS"
	} else if instanceNotFoundRegex.MatchString(message) {
		return "SERVICE_INSTANCE_NOT_FOUND"
	} else if bindingNotFoundRegex.MatchString(message) {
		return "SERVICE_BINDING_NOT_FOUND"
	} else if instanceExistsRegex.MatchString(message) {
		return "INSTANCE_ALREADY_EXISTS"
	} else {
		return ""
	}
}
