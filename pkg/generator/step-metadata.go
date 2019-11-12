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
	var docTemplatePath string
	var isGenerateDocu bool

	flag.StringVar(&docTemplatePath, "docuDir", "./documentation/docs/steps/", "The directory containing the docu stubs. Default points to \\'documentation/docs/steps.\\'")
	flag.BoolVar(&isGenerateDocu, "docuGen", false, "Boolean to generate Documentation or Step-MetaData. Default is false")
	flag.Parse()

	fmt.Printf("docuDir: %v, genDocu: %v \n", docTemplatePath, isGenerateDocu)

	metadataPath := "./resources/metadata"

	metadataFiles, err := helper.MetadataFiles(metadataPath)
	checkError(err)

	err = helper.ProcessMetaFiles(metadataFiles, openMetaFile, fileWriter, "", isGenerateDocu, docTemplatePath)
	checkError(err)

	cmd := exec.Command("go", "fmt", "./cmd")
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
