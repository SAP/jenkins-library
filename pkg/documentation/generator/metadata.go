package generator

import (
	"fmt"
	"reflect"

	"github.com/SAP/jenkins-library/pkg/config"
)

// readStepMetadata reads and parses the provided step metadata file
func readStepMetadata(metadataFilePath string, docuHelperData DocuHelperData) config.StepData {
	stepMetadata := config.StepData{}
	metadataFile, err := docuHelperData.OpenFile(metadataFilePath)
	checkError(err)
	defer metadataFile.Close()
	fmt.Printf("Reading metadata file: %v\n", metadataFilePath)
	err = stepMetadata.ReadPipelineStepData(metadataFile)
	checkError(err)
	return stepMetadata
}

// adjustDefaultValues corrects the Default value according to the Type.
func adjustDefaultValues(stepMetadata *config.StepData) {
	for key, parameter := range stepMetadata.Spec.Inputs.Parameters {
		var typedDefault interface{} = nil
		if parameter.Type == "bool" {
			typedDefault = false
		}
		if parameter.Default != nil ||
			parameter.Default == typedDefault {
			continue
		}
		fmt.Printf("Changing default value to '%v' for parameter '%s', was '%v'.\n", typedDefault, parameter.Name, parameter.Default)
		stepMetadata.Spec.Inputs.Parameters[key].Default = typedDefault
	}
}

// adjustMandatoryFlags corrects the Mandatory flag on each parameter if a non-empty default value is provided
func adjustMandatoryFlags(stepMetadata *config.StepData) {
	for key, parameter := range stepMetadata.Spec.Inputs.Parameters {
		if parameter.Mandatory {
			if parameter.Default == nil ||
				parameter.Default == "" ||
				parameter.Type == "[]string" && len(parameter.Default.([]string)) == 0 {
				continue
			}
			fmt.Printf("Changing mandatory flag to '%v' for parameter '%s', default value available '%v'.\n", false, parameter.Name, parameter.Default)
			stepMetadata.Spec.Inputs.Parameters[key].Mandatory = false
		}
	}
}

// applyCustomDefaultValues applies custom default values from the passed config
func applyCustomDefaultValues(stepMetadata *config.StepData, stepConfiguration config.StepConfig) {
	for key, parameter := range stepMetadata.Spec.Inputs.Parameters {
		configValue := stepConfiguration.Config[parameter.Name]
		if len(parameter.Conditions) != 0 {
			fmt.Printf("Skipping custom default values for '%s' as the parameter depends on other parameter values.\n", parameter.Name)
			continue
		}
		if configValue != nil && configValue != "" {
			switch parameter.Type {
			case "[]string":
				if reflect.DeepEqual(parameter.Default, configValue) {
					continue
				}
				fmt.Printf("Applying custom default value '%v' for parameter '%s', was '%v'.\n", configValue, parameter.Name, parameter.Default)
				stepMetadata.Spec.Inputs.Parameters[key].Default = configValue
			default:
				if parameter.Default != configValue {
					fmt.Printf("Applying custom default value '%v' for parameter '%s', was '%v'.\n", configValue, parameter.Name, parameter.Default)
					stepMetadata.Spec.Inputs.Parameters[key].Default = configValue
				}
			}
		}
	}
}
