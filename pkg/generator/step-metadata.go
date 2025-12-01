package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/SAP/jenkins-library/pkg/generator/helper"
)

func main() {
	var metadataPath string
	var targetDir string

	flag.StringVar(&metadataPath, "metadataDir", "./resources/metadata", "The directory containing the step metadata. Default points to 'resources/metadata'.")
	flag.StringVar(&targetDir, "targetDir", "./cmd", "The target directory for the generated commands.")
	flag.Parse()

	fmt.Printf("metadataDir: %v\ntargetDir: %v\n", metadataPath, targetDir)

	metadataFiles, err := helper.MetadataFiles(metadataPath)
	if err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}
	if err = helper.ProcessMetaFiles(metadataFiles, targetDir, helper.StepHelperData{
		OpenFile:  openMetaFile,
		WriteFile: fileWriter,
	}); err != nil {
		log.Fatalf("Error occurred: %v\n", err)
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
	if err = cmd.Run(); err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}
}
func openMetaFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func fileWriter(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
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
	return os.WriteFile(filename, data, perm)
}
