// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type transportRequestUploadSOLMANOptions struct {
	Endpoint              string   `json:"endpoint,omitempty"`
	Username              string   `json:"username,omitempty"`
	Password              string   `json:"password,omitempty"`
	ApplicationID         string   `json:"applicationId,omitempty"`
	ChangeDocumentID      string   `json:"changeDocumentId,omitempty"`
	TransportRequestID    string   `json:"transportRequestId,omitempty"`
	FilePath              string   `json:"filePath,omitempty"`
	CmClientOpts          []string `json:"cmClientOpts,omitempty"`
	GitFrom               string   `json:"gitFrom,omitempty"`
	GitTo                 string   `json:"gitTo,omitempty"`
	ChangeDocumentLabel   string   `json:"changeDocumentLabel,omitempty"`
	TransportRequestLabel string   `json:"transportRequestLabel,omitempty"`
}

type transportRequestUploadSOLMANCommonPipelineEnvironment struct {
	custom struct {
		changeDocumentID   string
		transportRequestID string
	}
}

func (p *transportRequestUploadSOLMANCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "changeDocumentId", value: p.custom.changeDocumentID},
		{category: "custom", name: "transportRequestId", value: p.custom.transportRequestID},
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
		log.Entry().Fatal("failed to persist Piper environment")
	}
}

// TransportRequestUploadSOLMANCommand Uploads a specified file into a given transport via Solution Manager
func TransportRequestUploadSOLMANCommand() *cobra.Command {
	const STEP_NAME = "transportRequestUploadSOLMAN"

	metadata := transportRequestUploadSOLMANMetadata()
	var stepConfig transportRequestUploadSOLMANOptions
	var startTime time.Time
	var commonPipelineEnvironment transportRequestUploadSOLMANCommonPipelineEnvironment

	var createTransportRequestUploadSOLMANCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Uploads a specified file into a given transport via Solution Manager",
		Long: `Uploads the file specified by 'filePath' into the given transport request 'transportRequestId' via Solution Manager.
The mandatory 'changeDocumentId' points to the associate change document item.
The 'applicationId' specifies how the file needs to be handled on server side.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			transportRequestUploadSOLMAN(stepConfig, &telemetryData, &commonPipelineEnvironment)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTransportRequestUploadSOLMANFlags(createTransportRequestUploadSOLMANCmd, &stepConfig)
	return createTransportRequestUploadSOLMANCmd
}

func addTransportRequestUploadSOLMANFlags(cmd *cobra.Command, stepConfig *transportRequestUploadSOLMANOptions) {
	cmd.Flags().StringVar(&stepConfig.Endpoint, "endpoint", os.Getenv("PIPER_endpoint"), "Service endpoint")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "Service user for uploading to the Solution Manager")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Service user password for uploading to the Solution Manager")
	cmd.Flags().StringVar(&stepConfig.ApplicationID, "applicationId", os.Getenv("PIPER_applicationId"), "Id of the application. Specifies how the file needs to be handled on server side")
	cmd.Flags().StringVar(&stepConfig.ChangeDocumentID, "changeDocumentId", os.Getenv("PIPER_changeDocumentId"), "Id of the change document the file is uploaded to")
	cmd.Flags().StringVar(&stepConfig.TransportRequestID, "transportRequestId", os.Getenv("PIPER_transportRequestId"), "Id of the transport request the file is uploaded to")
	cmd.Flags().StringVar(&stepConfig.FilePath, "filePath", os.Getenv("PIPER_filePath"), "Name/Path of the file which should be uploaded")
	cmd.Flags().StringSliceVar(&stepConfig.CmClientOpts, "cmClientOpts", []string{}, "Additional options handed over to the cm client")
	cmd.Flags().StringVar(&stepConfig.GitFrom, "gitFrom", `origin/master`, "GIT starting point for retrieving the change document and transport request id.")
	cmd.Flags().StringVar(&stepConfig.GitTo, "gitTo", `HEAD`, "GIT ending point for retrieving the change document and transport request id. ")
	cmd.Flags().StringVar(&stepConfig.ChangeDocumentLabel, "changeDocumentLabel", `ChangeDocument`, "Pattern used for identifying lines holding the change document id. The GIT commit log messages are scanned for this label.")
	cmd.Flags().StringVar(&stepConfig.TransportRequestLabel, "transportRequestLabel", `TransportRequest`, "Pattern used for identifying lines holding the transport request id. The GIT commit log messages are scanned for this label.")

	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("applicationId")
	cmd.MarkFlagRequired("filePath")
	cmd.MarkFlagRequired("cmClientOpts")
}

// retrieve step metadata
func transportRequestUploadSOLMANMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "transportRequestUploadSOLMAN",
			Aliases:     []config.Alias{{Name: "transportRequestUploadFile", Deprecated: false}},
			Description: "Uploads a specified file into a given transport via Solution Manager",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "endpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "changeManagement/endpoint"}},
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "uploadCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "uploadCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:        "applicationId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name: "changeDocumentId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/changeDocumentId",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name: "transportRequestId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/transportRequestId",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name: "filePath",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "mtarFilePath",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:        "cmClientOpts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEP", "GENERAL"},
						Type:        "[]string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "clientOpts"}, {Name: "changeManagement/clientOpts"}},
					},
					{
						Name:        "gitFrom",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEP", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/git/from"}},
					},
					{
						Name:        "gitTo",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEP", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/git/to"}},
					},
					{
						Name:        "changeDocumentLabel",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEP", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/changeDocumentLabel"}},
					},
					{
						Name:        "transportRequestLabel",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEP", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/transportRequestLabel"}},
					},
				},
			},
			Containers: []config.Container{
				{Name: "cmclient", Image: "ppiper/cm-client"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"Name": "custom/changeDocumentId"},
							{"Name": "custom/transportRequestId"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
