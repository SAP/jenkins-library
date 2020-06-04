package cloudfoundry

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
)

var _readFile = ioutil.ReadFile
var _stat = os.Stat
var _writeFile = ioutil.WriteFile
var _traverse = traverse

// Substitute ...
func Substitute(ymlFile, replacementsFile string) (bool, error) {

	bIn, err := _readFile(ymlFile)
	if err != nil {
		return false, err
	}

	inDecoder := yaml.NewDecoder(bytes.NewReader(bIn))

	buf := new(bytes.Buffer)
	outEncoder := yaml.NewEncoder(buf)

	var updated bool

	for {

		mIn := make(map[string]interface{})

		decodeErr := inDecoder.Decode(&mIn)

		if decodeErr != nil {
			if decodeErr == io.EOF {
				break
			}

			return false, decodeErr
		}

		mReplacements, err := getReplacements(replacementsFile)
		if err != nil {
			return false, err
		}

		out, _updated, err := _traverse(mIn, mReplacements)

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

		err = _writeFile(ymlFile, buf.Bytes(), fInfo.Mode())
		if err != nil {
			return false, err
		}
	}

	return updated, nil
}

func traverse(node interface{}, replacements map[string]interface{}) (interface{}, bool, error) {
	switch t := node.(type) {
	case string:
		return handleString(t, replacements)
	case bool:
		return t, false, nil
	case int:
		return t, false, nil
	case map[string]interface{}:
		return handleMap(t, replacements)
	case []interface{}:
		return handleSlice(t, replacements)
	default:
		return nil, false, fmt.Errorf("Unkown type received: '%v' (%v)", reflect.TypeOf(node), node)
	}
}

func handleString(value string, replacements map[string]interface{}) (interface{}, bool, error) {

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

func getParameterValue(name string, replacements map[string]interface{}) interface{} {

	r := replacements[name]
	log.Entry().Infof("Value '%v' resolved for parameter '%s'", r, name)
	return r
}

func isFullMatch(value string, matches [][][]byte) bool {
	return strings.HasPrefix(value, "((") && strings.HasSuffix(value, "))") && len(matches) == 1 && len(matches[0]) == 1
}

func handleSlice(t []interface{}, replacements map[string]interface{}) ([]interface{}, bool, error) {
	tNode := make([]interface{}, 0)
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

func handleMap(t map[string]interface{}, replacements map[string]interface{}) (map[string]interface{}, bool, error) {
	tNode := make(map[string]interface{})
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

func getReplacements(replacementsFile string) (map[string]interface{}, error) {

	bReplacements, err := _readFile(replacementsFile)
	if err != nil {
		return nil, err
	}

	mReplacements := make(map[string]interface{})
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

	return mReplacements, nil
}
