package yaml

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"gopkg.in/yaml.v3"
)

type fUtils interface {
	FileRead(name string) ([]byte, error)
	FileWrite(name string, data []byte, mode os.FileMode) error
}

var _fileUtils fUtils

var _stat = os.Stat
var _traverse = traverse

// Substitute ...
func Substitute(ymlFile string, replacements map[string]any, replacementsFiles []string) (bool, error) {

	if _fileUtils == nil {
		_fileUtils = piperutils.Files{}
	}

	bIn, err := _fileUtils.FileRead(ymlFile)
	if err != nil {
		return false, err
	}

	inDecoder := yaml.NewDecoder(bytes.NewReader(bIn))

	buf := new(bytes.Buffer)
	outEncoder := yaml.NewEncoder(buf)

	var updated bool

	mergedReplacements, err := getReplacements(replacements, replacementsFiles)
	if err != nil {
		return false, err
	}

	for {

		mIn := make(map[string]any)

		decodeErr := inDecoder.Decode(&mIn)

		if decodeErr != nil {
			if decodeErr == io.EOF {
				break
			}
			return false, decodeErr
		}

		if err != nil {
			return false, err
		}

		out, _updated, err := _traverse(mIn, mergedReplacements)
		if err != nil {
			return false, err
		}

		updated = _updated || updated

		err = outEncoder.Encode(out)
	}

	if updated {

		fInfo, err := _stat(ymlFile)
		if err != nil {
			return false, err
		}

		err = _fileUtils.FileWrite(ymlFile, buf.Bytes(), fInfo.Mode())
		if err != nil {
			return false, err
		}
	}

	return updated, nil
}

func traverse(node any, replacements map[string]any) (any, bool, error) {
	switch t := node.(type) {
	case string:
		return handleString(t, replacements)
	case bool:
		return t, false, nil
	case int:
		return t, false, nil
	case map[string]any:
		return handleMap(t, replacements)
	case map[any]any:
		m, err := keysToString(t)
		if err != nil {
			return nil, false, err
		}
		return handleMap(m, replacements)
	case []any:
		return handleSlice(t, replacements)
	default:
		return nil, false, fmt.Errorf("Unknown type received: '%v' (%v)", reflect.TypeOf(node), node)
	}
}

func keysToString(m map[any]any) (map[string]any, error) {
	result := map[string]any{}
	for key, val := range m {
		if k, ok := key.(string); ok {
			result[k] = val
		} else {
			return nil, fmt.Errorf("Cannot downcast'%v' to string. Type: %v)", reflect.TypeOf(key), key)
		}
	}
	return result, nil
}

func handleString(value string, replacements map[string]any) (any, bool, error) {

	trimmed := strings.TrimSpace(value)
	re := regexp.MustCompile(`\(\(.*?\)\)`)
	matches := re.FindAllSubmatch([]byte(trimmed), -1)
	fullMatch := isFullMatch(trimmed, matches)
	if fullMatch {
		log.Entry().Infof("FullMatchFound: %v", value)
		parameterName := getParameterName(matches[0][0])
		parameterValue := getParameterValue(parameterName, replacements)
		if parameterValue == nil {
			return nil, false, fmt.Errorf("No value available for parameters '%s', replacements: %v", parameterName, replacements)
		}
		log.Entry().Infof("FullMatchFound: '%s', replacing with '%v'", parameterName, parameterValue)
		return parameterValue, true, nil
	}
	// we have to scan for multiple variables
	// we return always a string
	updated := false
	for i, match := range matches {
		parameterName := getParameterName(match[0])
		log.Entry().Infof("XPartial match found: (%d) %v, %v", i, parameterName, value)
		parameterValue := getParameterValue(parameterName, replacements)
		if parameterValue == nil {
			return nil, false, fmt.Errorf("No value available for parameter '%s', replacements: %v", parameterName, replacements)
		}

		var conversion string
		switch t := parameterValue.(type) {
		case string:
			conversion = "%s"
		case bool:
			conversion = "%t"
		case int:
			conversion = "%d"
		case float64:
			conversion = "%g" // exponent as need, only required digits
		default:
			return nil, false, fmt.Errorf("Unsupported datatype found during travseral of yaml file: '%v', type: '%v'", parameterValue, reflect.TypeOf(t))
		}
		valueAsString := fmt.Sprintf(conversion, parameterValue)
		log.Entry().Infof("Value as String: %v: '%v'", parameterName, valueAsString)
		value = strings.Replace(value, "(("+parameterName+"))", valueAsString, -1)
		updated = true
		log.Entry().Infof("PartialMatchFound (%d): '%v', replaced with : '%s'", i, parameterName, valueAsString)
	}

	return value, updated, nil
}

func getParameterName(b []byte) string {
	pName := string(b)
	log.Entry().Infof("ParameterName is: '%s'", pName)
	return strings.Replace(strings.Replace(string(b), "((", "", 1), "))", "", 1)
}

func getParameterValue(name string, replacements map[string]any) any {

	r := replacements[name]
	log.Entry().Infof("Value '%v' resolved for parameter '%s'", r, name)
	return r
}

func isFullMatch(value string, matches [][][]byte) bool {
	return strings.HasPrefix(value, "((") && strings.HasSuffix(value, "))") && len(matches) == 1 && len(matches[0]) == 1
}

func handleSlice(t []any, replacements map[string]any) ([]any, bool, error) {
	tNode := make([]any, 0)
	updated := false
	for _, e := range t {
		if val, _updated, err := traverse(e, replacements); err == nil {
			updated = updated || _updated
			tNode = append(tNode, val)

		} else {
			return nil, false, err
		}
	}
	return tNode, updated, nil
}

func handleMap(t map[string]any, replacements map[string]any) (map[string]any, bool, error) {
	tNode := make(map[string]any)
	updated := false
	for key, value := range t {
		if val, _updated, err := traverse(value, replacements); err == nil {
			updated = updated || _updated
			tNode[key] = val
		} else {
			return nil, false, err
		}
	}
	return tNode, updated, nil
}

func getReplacements(replacements map[string]any, replacementsFiles []string) (map[string]any, error) {

	mReplacements := make(map[string]any)

	for _, replacementsFile := range replacementsFiles {
		bReplacements, err := _fileUtils.FileRead(replacementsFile)
		if err != nil {
			return nil, err
		}

		replacementsDecoder := yaml.NewDecoder(bytes.NewReader(bReplacements))

		for {
			decodeErr := replacementsDecoder.Decode(&mReplacements)

			if decodeErr != nil {
				if decodeErr == io.EOF {
					break
				}
				return nil, decodeErr
			}
		}
	}

	// the parameters from the map has a higher precedence,
	// hence we merge after resolving parameters from the files
	maps.Copy(mReplacements, replacements)
	return mReplacements, nil
}
