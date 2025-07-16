package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/generator/helper"
	"io"
	"os"
	"os/exec"
)

func main() {
	var metadataFile, moduleName, metadataPath, targetDir string
	flag.StringVar(&metadataFile, "metadataFile", "", "Single metadata file used to generate code for a step.")
	flag.StringVar(&moduleName, "moduleName", "", "Name of the module created with go mod init.")
	flag.StringVar(&metadataPath, "metadataDir", "./resources/metadata", "The directory containing the step metadata. Default points to \\'resources/metadata\\'.")
	flag.StringVar(&targetDir, "targetDir", "./cmd", "The target directory for the generated commands.")
	flag.Parse()
	fmt.Printf("metadataFile: %v\nmoduleName: %v\nmetadataDir: %v\ntargetDir: %v\n", metadataFile, moduleName, metadataPath, targetDir)

	var metadataFiles []string
	var err error
	if metadataFile != "" {
		fmt.Printf("Using single metadata file: %v\n", metadataFile)
		metadataFiles = []string{metadataFile}
		fmt.Println("Setting target directory to './cmd' as only one step is being generated.")
		targetDir = "./cmd"
	} else {
		fmt.Printf("Using metadata directory: %v\n", metadataPath)
		metadataFiles, err = helper.MetadataFiles(metadataPath)
		if err != nil {
			fmt.Printf("Error occurred: %v\n", err)
			os.Exit(1)
		}
	}

	err = processMetaFiles(metadataFiles, moduleName, targetDir, stepHelperData{
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
