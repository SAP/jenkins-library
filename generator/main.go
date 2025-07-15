package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/SAP/jenkins-library/generator/helper"
)

func main() {
	metadataFile := *flag.String("metadataFile", "", "Single metadata file used to generate code for a step.")
	metadataPath := *flag.String("metadataDir", "./resources/metadata", "The directory containing the step metadata. Default points to \\'resources/metadata\\'.")
	targetDir := *flag.String("targetDir", "./cmd", "The target directory for the generated commands.")
	flag.Parse()
	fmt.Printf("metadataFile: %v\nmetadataDir: %v\ntargetDir: %v\n", metadataFile, metadataPath, targetDir)

	var metadataFiles []string
	var err error
	if metadataFile != "" {
		fmt.Printf("Using single metadata file: %v\n", metadataFile)
		metadataFiles = []string{metadataFile}
	} else {
		fmt.Printf("Using metadata directory: %v\n", metadataPath)
		metadataFiles, err = helper.MetadataFiles(metadataPath)
		if err != nil {
			fmt.Printf("Error occurred: %v\n", err)
			os.Exit(1)
		}
	}

	err = helper.ProcessMetaFiles(metadataFiles, targetDir, helper.StepHelperData{
		OpenFile:     openMetaFile,
		WriteFile:    os.WriteFile,
		ExportPrefix: "",
	})
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running go fmt %v\n", targetDir)
	cmd := exec.Command("go", "fmt", targetDir)
	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	done := make(chan struct{})
	scanner := bufio.NewScanner(r)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		done <- struct{}{}

	}()
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)
		os.Exit(1)
	}
}

func openMetaFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}
