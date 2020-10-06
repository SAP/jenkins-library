package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	generator "github.com/SAP/jenkins-library/pkg/documentation/generator"
	"github.com/SAP/jenkins-library/pkg/generator/helper"
	"github.com/ghodss/yaml"
)

func main() {
	var metadataPath string
	var docTemplatePath string
	var customLibraryStepFile string

	flag.StringVar(&metadataPath, "metadataDir", "./resources/metadata", "The directory containing the step metadata. Default points to \\'resources/metadata\\'.")
	flag.StringVar(&docTemplatePath, "docuDir", "./documentation/docs/steps/", "The directory containing the docu stubs. Default points to \\'documentation/docs/steps/\\'.")
	flag.StringVar(&customLibraryStepFile, "customLibraryStepFile", "", "")

	flag.Parse()

	fmt.Println("using Metadata Directory:", metadataPath)
	fmt.Println("using Documentation Directory:", docTemplatePath)

	if len(customLibraryStepFile) > 0 {
		fmt.Println("Reading custom library step mapping..")
		content, err := ioutil.ReadFile(customLibraryStepFile)
		checkError(err)
		err = yaml.Unmarshal(content, &generator.CustomLibrarySteps)
		checkError(err)
		fmt.Println(generator.CustomLibrarySteps)
	}

	metadataFiles, err := helper.MetadataFiles(metadataPath)
	checkError(err)
	err = generator.GenerateStepDocumentation(metadataFiles, generator.DocuHelperData{
		DocTemplatePath:     docTemplatePath,
		OpenDocTemplateFile: openDocTemplateFile,
		DocFileWriter:       writeFile,
		OpenFile:            openFile,
	})
	checkError(err)
}

func openDocTemplateFile(docTemplateFilePath string) (io.ReadCloser, error) {
	//check if template exists otherwise print No Template found
	if _, err := os.Stat(docTemplateFilePath); os.IsNotExist(err) {
		err := fmt.Errorf("no template found: %v", docTemplateFilePath)
		return nil, err
	}

	return os.Open(docTemplateFilePath)
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func openFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)
		os.Exit(1)
	}
}
