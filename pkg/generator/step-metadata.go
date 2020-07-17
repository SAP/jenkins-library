package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/SAP/jenkins-library/pkg/generator/helper"
)

func main() {
	var metadataPath string
	var targetDir string
	var docTemplatePath string
	var isGenerateDocu bool

	flag.StringVar(&metadataPath, "metadataDir", "./resources/metadata", "The directory containing the step metadata. Default points to \\'resources/metadata\\'.")
	flag.StringVar(&targetDir, "targetDir", "./cmd", "The target directory for the generated commands.")
	flag.StringVar(&docTemplatePath, "docuDir", "./documentation/docs/steps/", "The directory containing the docu stubs. Default points to \\'documentation/docs/steps/\\'.")
	flag.BoolVar(&isGenerateDocu, "docuGen", false, "Boolean to generate Documentation or Step-MetaData. Default is false.")
	flag.Parse()

	fmt.Printf("metadataDir: %v\n, targetDir: %v\n, docuDir: %v\n, genDocu: %v\n", metadataPath, targetDir, docTemplatePath, isGenerateDocu)

	metadataFiles, err := helper.MetadataFiles(metadataPath)
	checkError(err)
	docuHelperData := helper.DocuHelperData{isGenerateDocu, docTemplatePath, openDocTemplate, docFileWriter}
	stepHelperData := helper.StepHelperData{openMetaFile, fileWriter, ""}
	err = helper.ProcessMetaFiles(metadataFiles, targetDir, stepHelperData, docuHelperData)
	checkError(err)

	fmt.Printf("Running go fmt %v\n", targetDir)
	cmd := exec.Command("go", "fmt", targetDir)
	err = cmd.Run()
	checkError(err)

}
func openMetaFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func fileWriter(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error occured: %v\n", err)
		os.Exit(1)
	}
}

func openDocTemplate(docTemplateFilePath string) (io.ReadCloser, error) {

	//check if template exists otherwise print No Template found
	if _, err := os.Stat(docTemplateFilePath); os.IsNotExist(err) {
		err := fmt.Errorf("no template found: %v", docTemplateFilePath)
		return nil, err
	}

	return os.Open(docTemplateFilePath)
}

func docFileWriter(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}
