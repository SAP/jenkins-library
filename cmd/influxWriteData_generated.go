// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	// "go.opentelemetry.io/otel/propagation"
)

type influxWriteDataOptions struct {
	ServerURL    string `json:"serverUrl,omitempty"`
	AuthToken    string `json:"authToken,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	Organization string `json:"organization,omitempty"`
	DataMap      string `json:"dataMap,omitempty"`
	DataMapTags  string `json:"dataMapTags,omitempty"`
}

// InfluxWriteDataCommand Writes metrics to influxdb
func InfluxWriteDataCommand() *cobra.Command {
	const STEP_NAME = "influxWriteData"

	metadata := influxWriteDataMetadata()
	var stepConfig influxWriteDataOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createInfluxWriteDataCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Writes metrics to influxdb",
		Long:  `In this step, the metrics are written to the timeseries database [InfluxDB](https://www.influxdata.com/time-series-platform/influxdb/)`,
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
			log.RegisterSecret(stepConfig.AuthToken)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 || len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := cmd.Root().Context()
			// propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
			// extractedCarrier := propagation.MapCarrier(GeneralConfig.OtelCarrier)
			// ctx = propagator.Extract(ctx, extractedCarrier)
			log.Entry().Infof("OtelCarrier from step: %v", GeneralConfig.OtelCarrier)
			tracer := telemetry.GetTracer(ctx)
			_, span := tracer.Start(ctx, "piper.step.run")
			span.SetAttributes(attribute.String("piper.step.name", STEP_NAME))

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				defer span.End()
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.Dsn,
						GeneralConfig.HookConfig.SplunkConfig.Token,
						GeneralConfig.HookConfig.SplunkConfig.Index,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblToken,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblIndex,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME, GeneralConfig.HookConfig.PendoConfig.Token)
			influxWriteData(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addInfluxWriteDataFlags(createInfluxWriteDataCmd, &stepConfig)
	return createInfluxWriteDataCmd
}

func addInfluxWriteDataFlags(cmd *cobra.Command, stepConfig *influxWriteDataOptions) {
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", os.Getenv("PIPER_serverUrl"), "Base URL to the InfluxDB server")
	cmd.Flags().StringVar(&stepConfig.AuthToken, "authToken", os.Getenv("PIPER_authToken"), "Token to authenticate to the Influxdb")
	cmd.Flags().StringVar(&stepConfig.Bucket, "bucket", `piper`, "Name of database (1.8) or bucket (2.0)")
	cmd.Flags().StringVar(&stepConfig.Organization, "organization", os.Getenv("PIPER_organization"), "Name of influx organization. Only for Influxdb 2.0")
	cmd.Flags().StringVar(&stepConfig.DataMap, "dataMap", os.Getenv("PIPER_dataMap"), "Map of fields for each measurements. It has to be a JSON string. For example: {'series_1':{'field_a':11,'field_b':12},'series_2':{'field_c':21,'field_d':22}}")
	cmd.Flags().StringVar(&stepConfig.DataMapTags, "dataMapTags", os.Getenv("PIPER_dataMapTags"), "Map of tags for each measurements. It has to be a JSON string. For example: {'series_1':{'tag_a':'a','tag_b':'b'},'series_2':{'tag_c':'c','tag_d':'d'}}")

	cmd.MarkFlagRequired("serverUrl")
	cmd.MarkFlagRequired("authToken")
	cmd.MarkFlagRequired("dataMap")
}

// retrieve step metadata
func influxWriteDataMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "influxWriteData",
			Aliases:     []config.Alias{},
			Description: "Writes metrics to influxdb",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "influxAuthTokenId", Description: "Influxdb token for authentication to the InfluxDB. In 1.8 version use 'username:password' instead.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_serverUrl"),
					},
					{
						Name: "authToken",
						ResourceRef: []config.ResourceReference{
							{
								Name: "influxAuthTokenId",
								Type: "secret",
							},

							{
								Name:    "influxVaultSecretName",
								Type:    "vaultSecret",
								Default: "influxdb",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_authToken"),
					},
					{
						Name:        "bucket",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `piper`,
					},
					{
						Name:        "organization",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_organization"),
					},
					{
						Name:        "dataMap",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_dataMap"),
					},
					{
						Name:        "dataMapTags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_dataMapTags"),
					},
				},
			},
		},
	}
	return theMetaData
}
