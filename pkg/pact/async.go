package pact

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/pact/async"
	"github.com/SAP/jenkins-library/pkg/pact/util"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// ValidateAsynch validates an async contract
func ValidateAsynch(pathToPactFolder, pathToAsyncFile string, fileUtils piperutils.FileUtils) error {
	// // Read AsyncAPI JSON Into Memory
	asyncAPISpec := &async.AsyncAPISpec{}
	if err := util.ReadAndUnmarshalFile(pathToAsyncFile, asyncAPISpec, fileUtils); err != nil {
		return err
	}

	// // Generate Async API Comparison Map
	asyncAPIMap, err := asyncAPISpec.ComparisonMap()
	if err != nil {
		return err
	}

	// Generate Pact Comparison Maps for all Pact files
	pactMaps, err := async.GeneratePactMaps(pathToPactFolder, fileUtils)
	if err != nil {
		return err
	}

	// comparisonMapPact := pactMaps["recommend-promotions-async-claims-backend-async.json"]
	// for topic, ruleMap := range comparisonMapPact.Topics {
	// 	fmt.Printf("Topic: %q has these rules:\n", topic)
	// 	for fieldName, rules := range ruleMap {
	// 		fmt.Printf("\nPact Field %s\n", fieldName)
	// 		fmt.Printf("%s: %v\n", "PayLoad Value", rules.PayloadValue)
	// 		// fmt.Print("")
	// 	}
	// }
	// Verify Async API against all associated Pact Contracts
	ok, err := asyncAPIMap.ValidateContracts(pactMaps)
	if err != nil {
		return fmt.Errorf("error occured when validating contracts: %w", err)
	}
	if !ok {
		return fmt.Errorf("some tests failed")
	}

	return nil
}
