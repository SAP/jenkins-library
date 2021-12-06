package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	generator "github.com/SAP/jenkins-library/pkg/documentation/generator"
	"github.com/SAP/jenkins-library/pkg/generator/helper"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/ghodss/yaml"
)

type sliceFlags struct {
	list []string
}

func (f *sliceFlags) String() string {
	return ""
}

func (f *sliceFlags) Set(value string) error {
	f.list = append(f.list, value)
	return nil
}

func main() {
	// flags for step documentation
	var metadataPath string
	var docTemplatePath string
	var customLibraryStepFile string
	var customDefaultFiles sliceFlags
	var includeAzure bool
	flag.StringVar(&metadataPath, "metadataDir", "./resources/metadata", "The directory containing the step metadata. Default points to \\'resources/metadata\\'.")
	flag.StringVar(&docTemplatePath, "docuDir", "./documentation/docs/steps/", "The directory containing the docu stubs. Default points to \\'documentation/docs/steps/\\'.")
	flag.StringVar(&customLibraryStepFile, "customLibraryStepFile", "", "")
	flag.Var(&customDefaultFiles, "customDefaultFile", "Path to a custom default configuration file.")
	flag.BoolVar(&includeAzure, "includeAzure", false, "Include Azure-specifics in step documentation.")

	// flags for stage documentation
	var generateStageConfig bool
	var stageMetadataPath string
	var stageTargetPath string
	var relativeStepsPath string
	flag.BoolVar(&generateStageConfig, "generateStageConfig", false, "Create stage documentation instead of step documentation.")
	flag.StringVar(&stageMetadataPath, "stageMetadataPath", "./resources/com.sap.piper/pipeline/stageDefaults.yml", "The file containing the stage metadata. Default points to \\'./resources/com.sap.piper/pipeline/stageDefaults.yml\\'.")
	flag.StringVar(&stageTargetPath, "stageTargetPath", "./documentation/docs/stages/", "The target path for the generated stage documentation. Default points to \\'./documentation/docs/stages/\\'.")
	flag.StringVar(&relativeStepsPath, "relativeStepsPath", "../../steps", "The relative path from stages to steps")

	flag.Parse()

	if generateStageConfig {
		// generating stage documentation
		fmt.Println("Generating STAGE documentation")
		fmt.Println("using Metadata:", stageMetadataPath)
		fmt.Println("using stage target directory:", stageTargetPath)
		fmt.Println("using relative steps path:", relativeStepsPath)

		utils := &piperutils.Files{}
		err := generator.GenerateStageDocumentation(stageMetadataPath, stageTargetPath, relativeStepsPath, utils)
		checkError(err)

	} else {
		// generating step documentation
		fmt.Println("Generating STEP documentation")
		fmt.Println("using Metadata Directory:", metadataPath)
		fmt.Println("using Documentation Directory:", docTemplatePath)
		fmt.Println("using Custom Default Files:", strings.Join(customDefaultFiles.list, ", "))

		if len(customLibraryStepFile) > 0 {
			fmt.Println("Reading custom library step mapping..")
			content, err := ioutil.ReadFile(customLibraryStepFile)
			checkError(err)
			err = yaml.Unmarshal(content, &generator.CustomLibrarySteps)
			checkError(err)
			fmt.Println(generator.CustomLibrarySteps)
		}

		metadataFiles, err := helper.MetadataFiles(metadataPath)
		checkError(err)
		err = generator.GenerateStepDocumentation(metadataFiles, customDefaultFiles.list, generator.DocuHelperData{
			DocTemplatePath:     docTemplatePath,
			OpenDocTemplateFile: openDocTemplateFile,
			DocFileWriter:       writeFile,
			OpenFile:            openFile,
		}, includeAzure)
		checkError(err)
	}
}

func openDocTemplateFile(docTemplateFilePath string) (io.ReadCloser, error) {
	//check if template exists otherwise print No Template found
	if _, err := os.Stat(docTemplateFilePath); os.IsNotExist(err) {
		err := fmt.Errorf("no template found: %v", docTemplateFilePath)
		return nil, err
	}

	return os.Open(docTemplateFilePath)
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func openFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)
		os.Exit(1)
	}
}
