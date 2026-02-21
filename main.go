package main

//go:generate go run cmd/step-metadata/step-metadata.go --metadataDir=./resources/metadata/ --targetDir=./cmd/

import (
	"github.com/SAP/jenkins-library/cmd"
)

func main() {
	cmd.Execute()
}
