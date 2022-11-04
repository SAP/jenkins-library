package async

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/pact/util"
	"github.com/SAP/jenkins-library/pkg/pact/validate"
)

func (s *AsyncPactSpec) ComparisonMap() (*validate.ComparisonMap, error) {

	matchingRuleMap := &validate.ComparisonMap{Topics: map[string]map[string]*validate.Field{}}
	payLoadMap := &validate.ComparisonMap{Topics: map[string]map[string]*validate.Field{}}

	for _, message := range s.Messages {
		// TODO:- Topic and matching rule -> payload need to be enforced prior to being published
		// These checks will catch cases where the pacts were published prior to enforcement being enabled
		if message.MetaData.Topic == "" {
			fmt.Println("[ ERROR ] Pact Message is not properly annotated. Topic metadata is missing.")
			fmt.Printf("Please notify a developer working on %s that their pact contract needs to be updated.\n", s.Consumer)
			fmt.Printf("\t- Contract is between consumer '%s' and provider '%s'\n", s.Consumer, s.Provider)
			fmt.Println("\t- For more information refer to https://wiki.wdf.sap.corp/wiki/display/Eureka/Contract+Test+with+PACT")
			// Should cause failure here os.Exit(1).
			continue
		}

		if message.MatchingRules == nil {
			fmt.Printf("[ ERROR ] Pact does not contain matching rules. Unable to verify contract without matching rules\n")
			fmt.Printf("Please notify a developer working on %s that their pact contract needs to be updated.\n", s.Consumer)
			fmt.Printf("\t- Contract is between consumer '%s' and provider '%s'\n", s.Consumer, s.Provider)
			fmt.Println("\t- For more information refer to https://wiki.wdf.sap.corp/wiki/display/Eureka/Contract+Test+with+PACT")
			// Should cause failure here os.Exit(1)
			continue
		}
		matchingRuleMap.Topics[message.MetaData.Topic] = map[string]*validate.Field{}
		generateMatchersMap(message.MetaData.Topic, message.MatchingRules.Body, matchingRuleMap)

		payLoadMap.Topics[message.MetaData.Topic] = map[string]*validate.Field{}
		s.generatePayloadMap("", message.MetaData.Topic, message.Contents, payLoadMap)

	}

	for topic, ruleMap := range payLoadMap.Topics {
		matchingRule, ok := matchingRuleMap.Topics[topic]
		if !ok {
			fmt.Printf("[ WARN ] Missing topic")
			continue
		}
		for fieldName, payLoadMap := range ruleMap {
			_, ok := matchingRule[fieldName]
			if !ok {
				if reflect.TypeOf(payLoadMap.PayloadValue) != reflect.TypeOf([]interface{}{}) && reflect.TypeOf(payLoadMap.PayloadValue) != reflect.TypeOf(map[string]interface{}{}) {
					fmt.Println("[ ERROR ] No matching rule associated payload field in PACT contract. Unable to verify contract.")
					fmt.Println("Every field listed in the example payload must have a matching rule associated with it.")
					fmt.Printf("Please notify a developer working on %s that their pact contract needs to be updated.\n", s.Consumer)
					fmt.Printf("\t- Contract is between consumer '%s' and provider '%s'\n", s.Consumer, s.Provider)
					fmt.Printf("\t- Topic: '%s'\n", topic)
					fmt.Printf("\t- Field: '%s'\n", fieldName)
					fmt.Println("\t- For more information refer to https://wiki.wdf.sap.corp/wiki/display/Eureka/Contract+Test+with+PACT")
				}

				continue
			}

		}
	}

	for topic, ruleMap := range matchingRuleMap.Topics {
		payloadField, ok := payLoadMap.Topics[topic]
		if !ok {
			fmt.Printf("[ WARN ] Missing topic")
			continue
		}
		for fieldName := range ruleMap {
			payloadField, ok := payloadField[fieldName]
			if !ok {
				fmt.Println("[ ERROR ] No payload value assocaiated with matching rule in PACT contract. Unable to verify contract.")
				fmt.Println("Every field listed under matching rules must have an example payload value associated with it.")
				fmt.Printf("Please notify a developer working on %s that their pact contract needs to be updated.\n", s.Consumer)
				fmt.Printf("\t- Contract is between consumer '%s' and provider '%s'\n", s.Consumer, s.Provider)
				fmt.Printf("\t- Topic: '%s'\n", topic)
				fmt.Printf("\t- Field: '%s'\n", fieldName)
				fmt.Println("\t- For more information refer to https://wiki.wdf.sap.corp/wiki/display/Eureka/Contract+Test+with+PACT")
				continue
			}
			matchingRuleMap.Topics[topic][fieldName].PayloadValue = payloadField.PayloadValue
			// If payloadFeild.PayloadValue is a slice make sure the matching rule isn't specifying the type inside the slice. This has led to issues
			// across different pact frameworks. 
			if reflect.TypeOf(payloadField.PayloadValue) == reflect.TypeOf([]interface{}{}) {
				matchingRuleMap.Topics[topic][fieldName].MatchingRule = "array"
			}

		}
	}

	return matchingRuleMap, nil
}

// ComparisonMap creates a comparison map based off the AsyncPactSpec.
func generateMatchersMap(topic string, matchingRules map[string]MatchingRule, compareMap *validate.ComparisonMap) {

	// Used to strip all array indices from key in order to make final key consistent across languages
	r := regexp.MustCompile(`\[[0-9*]\]`)

	compareMap.Topics[topic] = map[string]*validate.Field{}

	for field, rule := range matchingRules {
		// Clean key
		trimmedField := strings.TrimPrefix(field, "$.")     // $.claimingAccount.[0].fullName -> claimingAccount[0].fullName
		trimmedField = r.ReplaceAllString(trimmedField, "") // claimingAccount[0].fullName -> claimingAccount.fullName

		compareMap.Topics[topic][trimmedField] = &validate.Field{MatchingRule: rule.Matchers[0].Match}

		if matchingRules[field].Matchers[0].Date != "" {
			compareMap.Topics[topic][trimmedField].Date = rule.Matchers[0].Date
		}

	}

}

func (s *AsyncPactSpec) generatePayloadMap(parentKey, topic string, currentContents map[string]interface{}, compareMap *validate.ComparisonMap) {
	for key, value := range currentContents {
		if parentKey != "" {
			key = fmt.Sprintf("%s.%s", parentKey, key)
		}

		compareMap.Topics[topic][key] = &validate.Field{PayloadValue: value}

		if reflect.TypeOf(value) == reflect.TypeOf([]interface{}{}) && len(value.([]interface{})) > 0 && reflect.TypeOf(value.([]interface{})[0]) == reflect.TypeOf(map[string]interface{}{}) {
			currentContents = value.([]interface{})[0].(map[string]interface{})
			s.generatePayloadMap(key, topic, currentContents, compareMap)
		} else if reflect.TypeOf(value) == reflect.TypeOf(map[string]interface{}{}) {
			currentContents = value.(map[string]interface{})
			s.generatePayloadMap(key, topic, currentContents, compareMap)
		}
	}
}

// GeneratePactMaps will accept in as an argument the path to a folder containing pact contracts.
//  For each pact contract a ComparisonMap will be created. The function will return a map containing
//  the key value pair: pactFileName => comparionMap
func GeneratePactMaps(pathToPactFolder string, fileUtils piperutils.FileUtils) (map[string]*validate.ComparisonMap, error) {

	var pactMaps = make(map[string]*validate.ComparisonMap)

	// Open Pact directory
	pactFiles, err := os.ReadDir(pathToPactFolder)
	if err != nil {
		log.Fatal(err)
	}

	// Read in each pact file into memory and generate comparison map
	for _, pactFile := range pactFiles {
		pactContractSpec := &AsyncPactSpec{}
		if err = util.ReadAndUnmarshalFile(pathToPactFolder+pactFile.Name(), pactContractSpec, fileUtils); err != nil {
			log.Fatal(err)
		}

		// Generate Pact comparison map
		pactComparisonMap, err := pactContractSpec.ComparisonMap()
		if err != nil {
			log.Fatal(err)
		}

		// Add comparison map to returnMap [pactFile] => pactComparisonMap
		pactMaps[pactFile.Name()] = pactComparisonMap
	}

	return pactMaps, nil
}
