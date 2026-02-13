//go:build integration
// +build integration

package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/generator"
	"github.com/stretchr/testify/assert"
)

func TestCommandContract(t *testing.T) {
	assert.Equal(t, "", "")
}

// Test provided by consumer: SAP InnerSource project
// Changes to the test require peer review by core-team members involved in the project.
func TestGenerator(t *testing.T) {
	dir := t.TempDir()

	metadata := `metadata:
  name: test
  description: testDescription
  longDescription: testLongDescription
  spec:
    inputs:
      secrets:
        - name: secret
          description: secretDescription
          type: jenkins
      params:
        - name: testParam
          aliases:
            - name: testAlias
          type: string
          description: The name of the Checkmarx project to scan into
          mandatory: true
          scope:
            - PARAMETERS
            - STAGES
            - STEPS
          resourceRef:
            - name: commonPipelineEnvironment
              param: test/test
      outputs:
        resources:
          - name: influx
            type: influx
            params:
              - name: test_influx
                fields:
                  - name: testfield
          - name: commonPipelineEnvironment
            type: piperEnvironment
            params:
              - name: test_cpe
`

	os.WriteFile(filepath.Join(dir, "test.yaml"), []byte(metadata), 0755)

	openMetaFile := func(name string) (io.ReadCloser, error) { return os.Open(name) }
	fileWriter := func(filename string, data []byte, perm os.FileMode) error { return nil }

	stepHelperData := generator.StepHelperData{OpenFile: openMetaFile, WriteFile: fileWriter, ExportPrefix: "piperOsCmd"}

	metadataFiles, err := generator.MetadataFiles(dir)
	assert.NoError(t, err)

	err = generator.ProcessMetaFiles(metadataFiles, "./cmd", stepHelperData)
	assert.NoError(t, err)
}
