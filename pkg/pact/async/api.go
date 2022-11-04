package async

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/SAP/jenkins-library/pkg/pact/validate"
)

// ComparionMap creates a comparison map from the calling asyncSpec
func (a *AsyncAPISpec) ComparisonMap() (*validate.ComparisonMap, error) {
	matchingRuleMap := &validate.ComparisonMap{Topics: map[string]map[string]*validate.Field{}}
	payLoadMap := &validate.ComparisonMap{Topics: map[string]map[string]*validate.Field{}}
	for channelName, channel := range a.Channels {
		channelTitle := channel.Publish.Message.Title

		matchingRuleMap.Topics[channelName] = map[string]*validate.Field{}
		a.generateMatchingRuleMap(channelName, channelTitle, "", matchingRuleMap)

		payLoadMap.Topics[channelName] = map[string]*validate.Field{}
		generatePayloadMap("", channelName, a.Components.Schemas[channelTitle].Examples, payLoadMap)
	}

	payLoadCount := 0
	for topic, ruleMap := range payLoadMap.Topics {
		matchingRule, ok := matchingRuleMap.Topics[topic]
		if !ok {
			fmt.Printf("[ WARN ] Missing topic")
			continue
		}
		for fieldName := range ruleMap {
			_, ok := matchingRule[fieldName]
			if !ok {
				fmt.Println("[ ERROR ] No matching rule associated payload field in asyncapidoc.json. Unable to verify contract.")
				fmt.Println("Every field listed in the example payload under components schema must have a matching rule associated with it.")
				fmt.Printf("\t- Topic: '%s'\n", topic)
				fmt.Printf("\t- Field: '%s'\n", fieldName)
				fmt.Println("\t- For more information refer to https://wiki.wdf.sap.corp/wiki/display/Eureka/Contract+Test+with+PACT")
				continue
			}
			payLoadCount++

			// fmt.Print("")
		}
	}

	matchingCount := 0
	for topic, ruleMap := range matchingRuleMap.Topics {
		payloadField, ok := payLoadMap.Topics[topic]
		if !ok {
			fmt.Printf("[ WARN ] Missing topic")
			continue
		}
		for fieldName := range ruleMap {
			payloadField, ok := payloadField[fieldName]
			if !ok {
				fmt.Println("[ ERROR ] No payload value associated with matching rule in asyncapidoc.json. Unable to verify contract.")
				fmt.Println("Every field listed under components schema must have an example payload value associated with it.")
				fmt.Printf("\t- Topic: '%s'\n", topic)
				fmt.Printf("\t- Field: '%s'\n", fieldName)
				fmt.Println("\t- For more information refer to https://wiki.wdf.sap.corp/wiki/display/Eureka/Contract+Test+with+PACT")
				continue
			}
			matchingRuleMap.Topics[topic][fieldName].PayloadValue = payloadField.PayloadValue
			matchingCount++
		}
	}

	// fmt.Printf("Payload Count: %d\n", payLoadCount)
	// fmt.Printf("Matching Count: %d\n", matchingCount)

	return matchingRuleMap, nil
}

//	generateMap creates comparison map used for async validation.
func (a *AsyncAPISpec) generateMatchingRuleMap(channelName, channelTitle, parentKey string, compareMap *validate.ComparisonMap) {

	for key, property := range a.Components.Schemas[channelTitle].Properties {
		// Builds key if field belongs to a nested object
		if parentKey != "" {
			key = fmt.Sprintf("%s.%s", parentKey, key)
		}

		// Reached end of nested object and can extract details into comparison map
		if property.Type != "" && property.Type != "array" {
			compareMap.Topics[channelName][key] = &validate.Field{MatchingRule: property.Type}

			// Format field used when validating dates
			if property.Format != "" {
				compareMap.Topics[channelName][key].Format = property.Format
			}

		}

		// Field references another object so we recursively drill into it
		if property.Ref != "" {
			compareMap.Topics[channelName][key] = &validate.Field{MatchingRule: "map"}
			refSplit := strings.Split(property.Ref, "/") // "#/components/schemas/ApprovalEO" -> ApprovalEO
			title := refSplit[len(refSplit)-1]
			a.generateMatchingRuleMap(channelName, title, key, compareMap)
		}

		// Field references another object so we recursively drill into it
		if property.Type == "array" {
			compareMap.Topics[channelName][key] = &validate.Field{MatchingRule: property.Type}
			// Check if its array of object or not
			if property.Items.Ref != "" {
				refSplit := strings.Split(property.Items.Ref, "/") // "#/components/schemas/ApprovalEO" -> ApprovalEO
				title := refSplit[len(refSplit)-1]
				a.generateMatchingRuleMap(channelName, title, key, compareMap)
			}
		}
	}
}

func generatePayloadMap(parentKey, channelName string, currentContents map[string]interface{}, compareMap *validate.ComparisonMap) {
	for key, value := range currentContents {
		if parentKey != "" {
			key = fmt.Sprintf("%s.%s", parentKey, key)
		}
		compareMap.Topics[channelName][key] = &validate.Field{PayloadValue: value}

		if reflect.TypeOf(value) == reflect.TypeOf(([]interface{}{})) && len(value.([]interface{})) > 0 && reflect.TypeOf(value.([]interface{})[0]) == reflect.TypeOf(map[string]interface{}{}) {
			currentContents = value.([]interface{})[0].(map[string]interface{})
			generatePayloadMap(key, channelName, currentContents, compareMap)

		} else if reflect.TypeOf(value) == reflect.TypeOf(map[string]interface{}{}) {
			currentContents = value.(map[string]interface{})
			generatePayloadMap(key, channelName, currentContents, compareMap)
		}
	}
}
