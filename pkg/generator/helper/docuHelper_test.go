package helper

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
)

var expectedResultDocument string = "# testStep\n\n\t## Description \n\nLong Test description\n\n\t\n\t## Prerequisites\n\t\n\tnone\n\n\t\n\t\n\t## Parameters\n\n| name | mandatory | default |\n| ---- | --------- | ------- |\n | param0 | No | val0 | \n  | param1 | No | <nil> | \n  | param2 | Yes | <nil> | \n ## Details\n * ` param0 ` :  param0 description \n  * ` param1 ` :  param1 description \n  * ` param2 ` :  param1 description \n \n\t\n\t## We recommend to define values of step parameters via [config.yml file](../configuration.md). \n\nIn following sections of the config.yml the configuration is possible:\n\n| parameter | general | step/stage |\n|-----------|---------|------------|\n | param0 | X |  | \n  | param1 |  |  | \n  | param2 |  |  | \n \n\t\n\t## Side effects\n\t\n\tnone\n\t\n\t## Exceptions\n\t\n\tnone\n\t\n\t## Example\n\n\tnone\n"

func configMetaDataMock(name string) (io.ReadCloser, error) {
	meta1 := `metadata:
  name: testStep
  description: Test description
  longDescription: |
    Long Test description
spec:
  inputs:
    params:
      - name: param0
        type: string
        description: param0 description
        default: val0
        scope:
        - GENERAL
        - PARAMETERS
        mandatory: true
      - name: param1
        type: string
        description: param1 description
        scope:
        - PARAMETERS
      - name: param2
        type: string
        description: param1 description
        scope:
        - PARAMETERS
        mandatory: true
`
	var r string
	switch name {
	case "test.yaml":
		r = meta1
	default:
		r = ""
	}
	return ioutil.NopCloser(strings.NewReader(r)), nil
}

func configOpenDocTemplateFileMock(docTemplateFilePath string) (io.ReadCloser, error) {
	meta1 := `# ${docGenStepName}

	## ${docGenDescription}
	
	## Prerequisites
	
	none

	## ${docJenkinsPluginDependencies}
	
	## ${docGenParameters}
	
	## ${docGenConfiguration}
	
	## Side effects
	
	none
	
	## Exceptions
	
	none
	
	## Example

	none
`
	switch docTemplateFilePath {
	case "testStep.md":
		return ioutil.NopCloser(strings.NewReader(meta1)), nil
	default:
		return ioutil.NopCloser(strings.NewReader("")), fmt.Errorf("Wrong Path: %v", docTemplateFilePath)
	}
}

var resultDocumentContent string

func docFileWriterMock(docTemplateFilePath string, data []byte, perm os.FileMode) error {

	resultDocumentContent = string(data)
	switch docTemplateFilePath {
	case "testStep.md":
		return nil
	default:
		return fmt.Errorf("Wrong Path: %v", docTemplateFilePath)
	}
}

func TestGenerateStepDocumentationSuccess(t *testing.T) {
	var stepData config.StepData
	contentMetaData, _ := configMetaDataMock("test.yaml")
	stepData.ReadPipelineStepData(contentMetaData)

	generateStepDocumentation(stepData,DocuHelperData{true, "" ,configOpenDocTemplateFileMock , docFileWriterMock})

	t.Run("Docu Generation Success", func(t *testing.T) {
		assert.Equal(t, expectedResultDocument, resultDocumentContent)
	})
}

func TestGenerateStepDocumentationError(t *testing.T) {
	var stepData config.StepData
	contentMetaData, _ := configMetaDataMock("test.yaml")
	stepData.ReadPipelineStepData(contentMetaData)

	err := generateStepDocumentation(stepData, DocuHelperData{true, "Dummy" ,configOpenDocTemplateFileMock , docFileWriterMock})

	t.Run("Docu Generation Success", func(t *testing.T) {
		assert.Error(t, err, fmt.Sprintf("Error occured: %v\n", err))
	})
}
