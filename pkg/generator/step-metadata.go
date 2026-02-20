package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/SAP/jenkins-library/pkg/generator/helper"
)

func main() {
	var metadataPath string
	var targetDir string

	flag.StringVar(&metadataPath, "metadataDir", "./resources/metadata", "The directory containing the step metadata. Default points to \\'resources/metadata\\'.")
	flag.StringVar(&targetDir, "targetDir", "./cmd", "The target directory for the generated commands.")
	flag.Parse()

	fmt.Printf("metadataDir: %v\n, targetDir: %v\n", metadataPath, targetDir)

	metadataFiles, err := helper.MetadataFiles(metadataPath)
	if err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}
	if err = helper.ProcessMetaFiles(metadataFiles, targetDir, helper.StepHelperData{
		OpenFile:     openMetaFile,
		WriteFile:    fileWriter,
		ExportPrefix: "",
	}); err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}
}
func openMetaFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func fileWriter(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}
