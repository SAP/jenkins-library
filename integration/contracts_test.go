package integration

import (
	"io/ioutil"
	"os"
	"github.com/stretchr/testify/assert"
	"github.com/SAP/jenkins-library/pkg/generator/helper"
	"testing"
)

func TestCommandContract(t *testing.T) {
	assert.Equal(t, "", "")
}

func TestGenerator(t *testing.T) {
	var docTemplatePath string
	var isGenerateDocu bool

	dir, err := ioutil.TempDir("", "")
	defer os.RemoveAll(dir) // clean up
	assert.NoError(t, err, "Error when creating temp dir")

	metadataPath := "./resources/metadata"
	docuHelperData := piperOsGenerator.DocuHelperData{isGenerateDocu, docTemplatePath, openDocTemplate, docFileWriter}
	stepHelperData := piperOsGenerator.StepHelperData{openMetaFile, fileWriter, "piperOsCmd"}

	metadataFiles, err := piperOsGenerator.MetadataFiles(metadataPath)
	checkError(err)

	err = piperOsGenerator.ProcessMetaFiles(metadataFiles, stepHelperData, docuHelperData)
	checkError(err)

	cmd := exec.Command("go", "fmt", "./cmd")
	err = cmd.Run()
	checkError(err)
}
