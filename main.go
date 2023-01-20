package main

//go:generate go run pkg/generator/step-metadata.go --metadataDir=./resources/metadata/ --targetDir=./cmd/

import (
	"github.com/costae/jenkins-library/tree/dev/cmd"
)

func main() {
	cmd.Execute()
}
