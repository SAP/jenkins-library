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
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type tmsUploadOptions struct {
	ServiceKey        string `json:"serviceKey,omitempty"`
	CustomDescription string `json:"customDescription,omitempty"`
	NamedUser         string `json:"namedUser,omitempty"`
	NodeName          string `json:"nodeName,omitempty"`
	MtaPath           string `json:"mtaPath,omitempty"`
	MtaVersion        string `json:"mtaVersion,omitempty"`
	Proxy             string `json:"proxy,omitempty"`
}

type tmsUploadInflux struct {
	step_data struct {
		fields struct {
			tms bool
		}
		tags struct {
		}
	}
}

func (i *tmsUploadInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       interface{}
	}{
		{valType: config.InfluxField, measurement: "step_data", name: "tms", value: i.step_data.fields.tms},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Influx environment")
	}
}

// TmsUploadCommand This step allows you to upload an MTA file (multi-target application archive) and multiple MTA extension descriptors into a TMS (SAP Cloud Transport Management service) landscape for further TMS-controlled distribution through a TMS-configured landscape.
func TmsUploadCommand() *cobra.Command {
	const STEP_NAME = "tmsUpload"

	metadata := tmsUploadMetadata()
	var stepConfig tmsUploadOptions
	var startTime time.Time
	var influx tmsUploadInflux
	var logCollector *log.CollectorHook

	var createTmsUploadCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step allows you to upload an MTA file (multi-target application archive) and multiple MTA extension descriptors into a TMS (SAP Cloud Transport Management service) landscape for further TMS-controlled distribution through a TMS-configured landscape.",
		Long: `This step allows you to upload an MTA file (multi-target application archive) and multiple MTA extension descriptors into a TMS (SAP Cloud Transport Management service) landscape for further TMS-controlled distribution through a TMS-configured landscape.

TMS lets you manage transports between SAP Business Technology Platform accounts in Neo and Cloud Foundry, such as from DEV to TEST and PROD accounts.
For more information, see [official documentation of SAP Cloud Transport Management service](https://help.sap.com/viewer/p/TRANSPORT_MANAGEMENT_SERVICE)

!!! note "Prerequisites"
* You have subscribed to and set up TMS, as described in [Initial Setup](https://help.sap.com/viewer/7f7160ec0d8546c6b3eab72fb5ad6fd8/Cloud/en-US/66fd7283c62f48adb23c56fb48c84a60.html), which includes the configuration of a node to be used for uploading an MTA file.
* A corresponding service key has been created, as described in [Set Up the Environment to Transport Content Archives directly in an Application](https://help.sap.com/viewer/7f7160ec0d8546c6b3eab72fb5ad6fd8/Cloud/en-US/8d9490792ed14f1bbf8a6ac08a6bca64.html). This service key (JSON) must be stored as a secret text within the Jenkins secure store.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				influx.persist(GeneralConfig.EnvRootPath, "influx")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunk.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunk.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			tmsUpload(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTmsUploadFlags(createTmsUploadCmd, &stepConfig)
	return createTmsUploadCmd
}

func addTmsUploadFlags(cmd *cobra.Command, stepConfig *tmsUploadOptions) {
	cmd.Flags().StringVar(&stepConfig.ServiceKey, "serviceKey", os.Getenv("PIPER_serviceKey"), "Temporal parameter for passing TMS service key.")
	cmd.Flags().StringVar(&stepConfig.CustomDescription, "customDescription", os.Getenv("PIPER_customDescription"), "Can be used as the description of a transport request. Will overwrite the default, which is corresponding Git commit-ID.")
	cmd.Flags().StringVar(&stepConfig.NamedUser, "namedUser", `Piper-Pipeline`, "Defines the named user to execute transport request with. The default value is 'Piper-Pipeline'.")
	cmd.Flags().StringVar(&stepConfig.NodeName, "nodeName", os.Getenv("PIPER_nodeName"), "Defines the name of the node to which the *.mtar file should be uploaded.")
	cmd.Flags().StringVar(&stepConfig.MtaPath, "mtaPath", os.Getenv("PIPER_mtaPath"), "Defines the relative path to *.mtar for the upload to the SAP Cloud Transport Management Service. If not specified, it will use the mtar file created in mtaBuild.")
	cmd.Flags().StringVar(&stepConfig.MtaVersion, "mtaVersion", `*`, "Defines the version of the MTA for which the MTA extension descriptor will be used. You can use an asterisk (*) to accept any MTA version, or use a specific version compliant with SemVer 2.0, e.g. 1.0.0 (see semver.org). If the parameter is not configured, an asterisk is used.")
	cmd.Flags().StringVar(&stepConfig.Proxy, "proxy", os.Getenv("PIPER_proxy"), "Proxy which should be used for the communication with the Transport Management Service Backend.")

	cmd.MarkFlagRequired("serviceKey")
	cmd.MarkFlagRequired("nodeName")
	cmd.MarkFlagRequired("mtaPath")
}

// retrieve step metadata
func tmsUploadMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "tmsUpload",
			Aliases:     []config.Alias{},
			Description: "This step allows you to upload an MTA file (multi-target application archive) and multiple MTA extension descriptors into a TMS (SAP Cloud Transport Management service) landscape for further TMS-controlled distribution through a TMS-configured landscape.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "credentialsId", Description: "Jenkins 'Secret text' credentials ID containing service key for SAP Cloud Transport Management service.", Type: "jenkins"},
				},
				Resources: []config.StepResources{
					{Name: "buildResult", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "serviceKey",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_serviceKey"),
					},
					{
						Name:        "customDescription",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_customDescription"),
					},
					{
						Name:        "namedUser",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `Piper-Pipeline`,
					},
					{
						Name:        "nodeName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_nodeName"),
					},
					{
						Name:        "mtaPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_mtaPath"),
					},
					{
						Name:        "mtaVersion",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `*`,
					},
					{
						Name:        "proxy",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_proxy"),
					},
				},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "influx",
						Type: "influx",
						Parameters: []map[string]interface{}{
							{"Name": "step_data"}, {"fields": []map[string]string{{"name": "tms"}}},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
