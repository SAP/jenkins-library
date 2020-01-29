package cmd

import (
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/spf13/cobra"
)

type mtaBuildOptions struct {
	BuildTarget string `json:"buildTarget,omitempty"`
}

type mtaBuildCommonPipelineEnvironment struct {
	mtarFilePath string
}

func (p *mtaBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    string
	}{
		{category: "", name: "mtarFilePath", value: p.mtarFilePath},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		os.Exit(1)
	}
}

var myMtaBuildOptions mtaBuildOptions

// MtaBuildCommand Performs an mta build
func MtaBuildCommand() *cobra.Command {
	metadata := mtaBuildMetadata()
	var commonPipelineEnvironment mtaBuildCommonPipelineEnvironment

	var createMtaBuildCmd = &cobra.Command{
		Use:   "mtaBuild",
		Short: "Performs an mta build",
		Long:  `Performs an mta build`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("mtaBuild")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "mtaBuild", &myMtaBuildOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			handler := func() {
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
			}
			log.DeferExitHandler(handler)
			defer handler()
			return mtaBuild(myMtaBuildOptions, &commonPipelineEnvironment)
		},
	}

	addMtaBuildFlags(createMtaBuildCmd)
	return createMtaBuildCmd
}

func addMtaBuildFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myMtaBuildOptions.BuildTarget, "buildTarget", os.Getenv("PIPER_buildTarget"), "Lorem ipsum")

}

// retrieve step metadata
func mtaBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "buildTarget",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
