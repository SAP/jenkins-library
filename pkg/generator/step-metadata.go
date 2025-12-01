package main

import (
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"

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
		WriteFile: formatAndWriteFile,
	}); err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}
}

func openMetaFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// formatAndWriteFile formats Go files using go/format before writing
func formatAndWriteFile(filename string, data []byte, perm os.FileMode) error {
	// Only format .go files
	if filepath.Ext(filename) == ".go" {
		formatted, err := format.Source(data)
		if err != nil {
			// If formatting fails, log the error but write the unformatted content
			// This prevents generation from failing due to syntax errors in templates
			fmt.Printf("Warning: failed to format %s: %v\n", filename, err)
			return os.WriteFile(filename, data, perm)
		}
		return os.WriteFile(filename, formatted, perm)
	}
	// Non-Go files are written as-is
	return os.WriteFile(filename, data, perm)
}
