package validate

import (
	"fmt"
	"reflect"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/pact/util"
)

const failureMessage = `{{range $topicName, $fieldFailureMap := .Failures }} {{if $fieldFailureMap}}
- Topic: '{{$topicName}}' {{range $field, $failures := $fieldFailureMap }}

   - ERROR validating field: '{{$field}}' {{ range $failure := $failures }}
      {{ $failure }}{{end}}{{end}}{{end}}{{end}}`

// FailureData defines a structure to contain individual failures which may occur when checking contracts
type FailureData struct {
	Failures map[string]map[string][]string
}

// ValidateContracts iterates over all pact contracts passed in as an argument and verifies asyncAPI against each contract
func (c *ComparisonMap) ValidateContracts(pactsMap map[string]*ComparisonMap) (bool, error) {
	success := true
	for contractName, pactMap := range pactsMap {
		if tplData, ok := c.ValidateContract(pactMap); !ok {
			success = false
			log.Entry().Warningf("FAILED CONTRACT - %s", contractName)
			val, err := util.ExecuteTemplate(failureMessage, tplData)
			if err != nil {
				return false, err
			}
			log.Entry().Warning(val)
		} else {
			log.Entry().Infof("VERIFIED CONTRACT - %s", contractName)

		}
	}
	return success, nil
}

// ValidateContract validates an asyncAPI doc against a pact contract
func (c *ComparisonMap) ValidateContract(pactMap *ComparisonMap) (*FailureData, bool) {

	success := true
	failureTemplate := FailureData{Failures: map[string]map[string][]string{}}

	for topic, pactFieldMap := range pactMap.Topics {
		failureTemplate.Failures[topic] = map[string][]string{}
		asyncFieldMap, ok := c.Topics[topic]

		if !ok {
			failureTemplate.Failures[topic]["*"] = []string{fmt.Sprintf("topic '%v' does not exist in async file", topic)}
			success = false
			continue
		}

		for pactFieldName, pactField := range pactFieldMap {
			asyncField, ok := asyncFieldMap[pactFieldName]
			if !ok {
				success = false
				failureTemplate.Failures[topic][pactFieldName] = append(failureTemplate.Failures[topic][pactFieldName], fmt.Sprintf("fieldName '%v' does not exist in async file", pactFieldName))
				continue
			}

			consumerMatchingRule := fmt.Sprintf("consumer matching rule listed as %s", pactField.MatchingRule)
			consumerPayloadValue := fmt.Sprintf("consumer payload value listed as %v", pactField.PayloadValue)
			var consumerPayloadType string
			if pactField.PayloadValue != nil {
				consumerPayloadType = fmt.Sprintf("consumer payload is of type %s", reflect.TypeOf(pactField.PayloadValue).String())
			}
			
			// consumerDateFormat := fmt.Sprintf("consumer date format is %s", pactField.Date)

			providerMatchingRule := fmt.Sprintf("provider matching rule listed as %s", asyncField.MatchingRule)
			providerPayloadValue := fmt.Sprintf("provider payload value listed as %v", asyncField.PayloadValue)
			var providerPayloadType string
			if asyncField.PayloadValue != nil {
				providerPayloadType = fmt.Sprintf("provider payload is of type %s", reflect.TypeOf(asyncField.PayloadValue).String())
			}
			providerFormat := fmt.Sprintf("provider format is listed as %s", asyncField.Format)

			switch pactField.MatchingRule {
			case "timestamp":
				fallthrough
			case "date":
				if asyncField.MatchingRule != "string" {
					success = false
					failureTemplate.Failures[topic][pactFieldName] = appendErrorMessages(failureTemplate.Failures[topic][pactFieldName], consumerMatchingRule, providerMatchingRule,
						consumerPayloadValue, providerPayloadValue)
				}

				if asyncField.Format != "" && asyncField.Format != "date-time" && asyncField.Format != "date" {
					success = false
					failureTemplate.Failures[topic][pactFieldName] = appendErrorMessages(failureTemplate.Failures[topic][pactFieldName], consumerMatchingRule, providerFormat)
				}
			case "decimal":
				if asyncField.MatchingRule != "number" {
					success = false
					failureTemplate.Failures[topic][pactFieldName] = appendErrorMessages(failureTemplate.Failures[topic][pactFieldName], consumerMatchingRule, providerMatchingRule,
						consumerPayloadValue, providerPayloadValue)
				}
			case "integer":
				if asyncField.MatchingRule != "integer" {
					success = false
					failureTemplate.Failures[topic][pactFieldName] = appendErrorMessages(failureTemplate.Failures[topic][pactFieldName], consumerMatchingRule, providerMatchingRule,
						consumerPayloadValue, providerPayloadValue)
				}
			case "number":
				if asyncField.MatchingRule != "integer" && asyncField.MatchingRule != "number" {
					success = false
					failureTemplate.Failures[topic][pactFieldName] = appendErrorMessages(failureTemplate.Failures[topic][pactFieldName], consumerMatchingRule, providerMatchingRule,
						consumerPayloadValue, providerPayloadValue)
				}
			default: // Defaults to matching rule 'type'
				if reflect.TypeOf(pactField.PayloadValue) != reflect.TypeOf(asyncField.PayloadValue) {
					success = false
					failureTemplate.Failures[topic][pactFieldName] = appendErrorMessages(failureTemplate.Failures[topic][pactFieldName], consumerMatchingRule, consumerPayloadType,
						providerMatchingRule, providerPayloadType)
				}
			}

		}
	}

	return &failureTemplate, success
}

func appendErrorMessages(messages []string, messagesToAppend ...string) []string {
	messages = append(messages, messagesToAppend...)
	return messages
}
