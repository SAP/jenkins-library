package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/SAP/jenkins-library/pkg/generator/helper"
)

func main() {

	metadataPath := "./resources/metadata"

	metadataFiles, err := helper.MetadataFiles(metadataPath)
	checkError(err)

	err = helper.ProcessMetaFiles(metadataFiles, openMetaFile, fileWriter, "")
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
