package btp

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	errorBlockRegex = regexp.MustCompile(`\{[\s]*?"error"\s*:\s*".*?"[\s\S]*?"description"\s*:\s*".*?"[\s\S]*?\}`)
)

func GetErrorInfos(value string) (BTPErrorData, error) {
	var errorBlock, err = extractLastErrorBlock(value)

	if errorBlock != "" && err == nil {
		// Try to extract more specific error information
		res, err := GetJSON(errorBlock)
		if err == nil {
			errorData := BTPErrorData{}

			err := json.Unmarshal([]byte(res), &errorData)
			if err != nil {
				return errorData, err
			}

			return errorData, nil
		}
	}
	return BTPErrorData{}, errors.New("no Error block found")
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
