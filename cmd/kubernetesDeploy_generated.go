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
)

type kubernetesDeployOptions struct {
	AdditionalParameters       []string               `json:"additionalParameters,omitempty"`
	APIServer                  string                 `json:"apiServer,omitempty"`
	AppTemplate                string                 `json:"appTemplate,omitempty"`
	ChartPath                  string                 `json:"chartPath,omitempty"`
	ContainerRegistryPassword  string                 `json:"containerRegistryPassword,omitempty"`
	ContainerImageName         string                 `json:"containerImageName,omitempty"`
	ContainerImageTag          string                 `json:"containerImageTag,omitempty"`
	ContainerRegistryURL       string                 `json:"containerRegistryUrl,omitempty"`
	ContainerRegistryUser      string                 `json:"containerRegistryUser,omitempty"`
	ContainerRegistrySecret    string                 `json:"containerRegistrySecret,omitempty"`
	CreateDockerRegistrySecret bool                   `json:"createDockerRegistrySecret,omitempty"`
	DeploymentName             string                 `json:"deploymentName,omitempty"`
	DeployTool                 string                 `json:"deployTool,omitempty" validate:"possible-values=kubectl helm helm3"`
	ForceUpdates               bool                   `json:"forceUpdates,omitempty"`
	HelmDeployWaitSeconds      int                    `json:"helmDeployWaitSeconds,omitempty"`
	HelmTestWaitSeconds        int                    `json:"helmTestWaitSeconds,omitempty"`
	HelmValues                 []string               `json:"helmValues,omitempty"`
	ValuesMapping              map[string]interface{} `json:"valuesMapping,omitempty"`
	GithubToken                string                 `json:"githubToken,omitempty"`
	Image                      string                 `json:"image,omitempty"`
	ImageNames                 []string               `json:"imageNames,omitempty"`
	ImageNameTags              []string               `json:"imageNameTags,omitempty"`
	ImageDigests               []string               `json:"imageDigests,omitempty"`
	IngressHosts               []string               `json:"ingressHosts,omitempty"`
	KeepFailedDeployments      bool                   `json:"keepFailedDeployments,omitempty"`
	RunHelmTests               bool                   `json:"runHelmTests,omitempty"`
	ShowTestLogs               bool                   `json:"showTestLogs,omitempty"`
	KubeConfig                 string                 `json:"kubeConfig,omitempty"`
	KubeContext                string                 `json:"kubeContext,omitempty"`
	KubeToken                  string                 `json:"kubeToken,omitempty"`
	Namespace                  string                 `json:"namespace,omitempty"`
	TillerNamespace            string                 `json:"tillerNamespace,omitempty"`
	DockerConfigJSON           string                 `json:"dockerConfigJSON,omitempty"`
	DeployCommand              string                 `json:"deployCommand,omitempty" validate:"possible-values=apply replace"`
	SetupScript                string                 `json:"setupScript,omitempty"`
	VerificationScript         string                 `json:"verificationScript,omitempty"`
	TeardownScript             string                 `json:"teardownScript,omitempty"`
}

// KubernetesDeployCommand Deployment to Kubernetes test or production namespace within the specified Kubernetes cluster.
func KubernetesDeployCommand() *cobra.Command {
	const STEP_NAME = "kubernetesDeploy"

	metadata := kubernetesDeployMetadata()
	var stepConfig kubernetesDeployOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createKubernetesDeployCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Deployment to Kubernetes test or production namespace within the specified Kubernetes cluster.",
		Long: `Deployment to Kubernetes test or production namespace within the specified Kubernetes cluster.

!!! note "Deployment supports multiple deployment tools"
    Currently the following are supported:

    * [Helm](https://helm.sh/) command line tool and [Helm Charts](https://docs.helm.sh/developing_charts/#charts).
    * [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) and ` + "`" + `kubectl apply` + "`" + ` command.

## Helm
Following helm command will be executed by default:

` + "`" + `` + "`" + `` + "`" + `
helm upgrade <deploymentName> <chartPath> --install --force --namespace <namespace> --wait --timeout <helmDeployWaitSeconds> --set "image.repository=<yourRegistry>/<yourImageName>,image.tag=<yourImageTag>,secret.dockerconfigjson=<dockerSecret>,ingress.hosts[0]=<ingressHosts[0]>,,ingress.hosts[1]=<ingressHosts[1]>,...
` + "`" + `` + "`" + `` + "`" + `

* ` + "`" + `yourRegistry` + "`" + ` will be retrieved from ` + "`" + `containerRegistryUrl` + "`" + `
* ` + "`" + `yourImageName` + "`" + `, ` + "`" + `yourImageTag` + "`" + ` will be retrieved from ` + "`" + `image` + "`" + `
* ` + "`" + `dockerSecret` + "`" + ` will be calculated with a call to ` + "`" + `kubectl create secret generic <containerRegistrySecret> --from-file=.dockerconfigjson=<dockerConfigJson> --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json` + "`" + ``,
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
			log.RegisterSecret(stepConfig.ContainerRegistryPassword)
			log.RegisterSecret(stepConfig.ContainerRegistryUser)
			log.RegisterSecret(stepConfig.GithubToken)
			log.RegisterSecret(stepConfig.KubeConfig)
			log.RegisterSecret(stepConfig.KubeToken)
			log.RegisterSecret(stepConfig.DockerConfigJSON)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
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
		Run: func(_ *cobra.Command, _ []string) {
			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			kubernetesDeploy(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addKubernetesDeployFlags(createKubernetesDeployCmd, &stepConfig)
	return createKubernetesDeployCmd
}

func addKubernetesDeployFlags(cmd *cobra.Command, stepConfig *kubernetesDeployOptions) {
	cmd.Flags().StringSliceVar(&stepConfig.AdditionalParameters, "additionalParameters", []string{}, "Defines additional parameters for \"helm install\" or \"kubectl apply\" command.")
	cmd.Flags().StringVar(&stepConfig.APIServer, "apiServer", os.Getenv("PIPER_apiServer"), "Defines the Url of the API Server of the Kubernetes cluster.")
	cmd.Flags().StringVar(&stepConfig.AppTemplate, "appTemplate", os.Getenv("PIPER_appTemplate"), "Defines the filename for the kubernetes app template (e.g. k8s_apptemplate.yaml).")
	cmd.Flags().StringVar(&stepConfig.ChartPath, "chartPath", os.Getenv("PIPER_chartPath"), "Defines the chart path for deployments using helm. It is a mandatory parameter when `deployTool:helm` or `deployTool:helm3`.")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryPassword, "containerRegistryPassword", os.Getenv("PIPER_containerRegistryPassword"), "Password for container registry access - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&stepConfig.ContainerImageName, "containerImageName", os.Getenv("PIPER_containerImageName"), "Name of the container which will be built - will be used together with `containerImageTag` instead of parameter `containerImage`")
	cmd.Flags().StringVar(&stepConfig.ContainerImageTag, "containerImageTag", os.Getenv("PIPER_containerImageTag"), "Tag of the container which will be built - will be used together with `containerImageName` instead of parameter `containerImage`")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryURL, "containerRegistryUrl", os.Getenv("PIPER_containerRegistryUrl"), "http(s) url of the Container registry where the image to deploy is located.")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryUser, "containerRegistryUser", os.Getenv("PIPER_containerRegistryUser"), "Username for container registry access - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistrySecret, "containerRegistrySecret", `regsecret`, "Name of the container registry secret used for pulling containers from the registry.")
	cmd.Flags().BoolVar(&stepConfig.CreateDockerRegistrySecret, "createDockerRegistrySecret", false, "Only for `deployTool:kubectl`: Toggle to turn on `containerRegistrySecret` creation.")
	cmd.Flags().StringVar(&stepConfig.DeploymentName, "deploymentName", os.Getenv("PIPER_deploymentName"), "Defines the name of the deployment. It is a mandatory parameter when `deployTool:helm` or `deployTool:helm3`.")
	cmd.Flags().StringVar(&stepConfig.DeployTool, "deployTool", `kubectl`, "Defines the tool which should be used for deployment.")
	cmd.Flags().BoolVar(&stepConfig.ForceUpdates, "forceUpdates", true, "Adds `--force` flag to a helm resource update command or to a kubectl replace command")
	cmd.Flags().IntVar(&stepConfig.HelmDeployWaitSeconds, "helmDeployWaitSeconds", 300, "Number of seconds before helm deploy returns.")
	cmd.Flags().IntVar(&stepConfig.HelmTestWaitSeconds, "helmTestWaitSeconds", 0, "Time to wait for any individual Kubernetes operation (like Jobs for hooks) . this param gets translated to `--timeout param` for helm cli during helm test. Default value is 5m0s (5 min and 0 seconds) if not set via parameter. See https://helm.sh/docs/helm/helm_test/#options for further details")
	cmd.Flags().StringSliceVar(&stepConfig.HelmValues, "helmValues", []string{}, "List of helm values as YAML file reference or URL (as per helm parameter description for `-f` / `--values`)")

	cmd.Flags().StringVar(&stepConfig.GithubToken, "githubToken", os.Getenv("PIPER_githubToken"), "GitHub personal access token as per https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line")
	cmd.Flags().StringVar(&stepConfig.Image, "image", os.Getenv("PIPER_image"), "Full name of the image to be deployed.")
	cmd.Flags().StringSliceVar(&stepConfig.ImageNames, "imageNames", []string{}, "List of names of the images to be deployed.")
	cmd.Flags().StringSliceVar(&stepConfig.ImageNameTags, "imageNameTags", []string{}, "List of full names (registry and tag) of the images to be deployed.")
	cmd.Flags().StringSliceVar(&stepConfig.ImageDigests, "imageDigests", []string{}, "List of image digests of the images to be deployed, in the format `sha256:<hash>`. If provided, image digests will be appended to the image tag, e.g. `<repository>/<name>:<tag>@<digest>`")
	cmd.Flags().StringSliceVar(&stepConfig.IngressHosts, "ingressHosts", []string{}, "(Deprecated) List of ingress hosts to be exposed via helm deployment.")
	cmd.Flags().BoolVar(&stepConfig.KeepFailedDeployments, "keepFailedDeployments", false, "Defines whether a failed deployment will be purged")
	cmd.Flags().BoolVar(&stepConfig.RunHelmTests, "runHelmTests", false, "Defines whether or not to run helm tests against the recently deployed release")
	cmd.Flags().BoolVar(&stepConfig.ShowTestLogs, "showTestLogs", false, "Defines whether to print the pod logs after running helm tests")
	cmd.Flags().StringVar(&stepConfig.KubeConfig, "kubeConfig", os.Getenv("PIPER_kubeConfig"), "Defines the path to the \"kubeconfig\" file.")
	cmd.Flags().StringVar(&stepConfig.KubeContext, "kubeContext", os.Getenv("PIPER_kubeContext"), "Defines the context to use from the \"kubeconfig\" file.")
	cmd.Flags().StringVar(&stepConfig.KubeToken, "kubeToken", os.Getenv("PIPER_kubeToken"), "Contains the id_token used by kubectl for authentication. Consider using kubeConfig parameter instead.")
	cmd.Flags().StringVar(&stepConfig.Namespace, "namespace", `default`, "Defines the target Kubernetes namespace for the deployment.")
	cmd.Flags().StringVar(&stepConfig.TillerNamespace, "tillerNamespace", os.Getenv("PIPER_tillerNamespace"), "Defines optional tiller namespace for deployments using helm.")
	cmd.Flags().StringVar(&stepConfig.DockerConfigJSON, "dockerConfigJSON", `.pipeline/docker/config.json`, "Path to the file `.docker/config.json` - this is typically provided by your CI/CD system. You can find more details about the Docker credentials in the [Docker documentation](https://docs.docker.com/engine/reference/commandline/login/).")
	cmd.Flags().StringVar(&stepConfig.DeployCommand, "deployCommand", `apply`, "Only for `deployTool: kubectl`: defines the command `apply` or `replace`. The default is `apply`.")
	cmd.Flags().StringVar(&stepConfig.SetupScript, "setupScript", os.Getenv("PIPER_setupScript"), "HTTP location of setup script")
	cmd.Flags().StringVar(&stepConfig.VerificationScript, "verificationScript", os.Getenv("PIPER_verificationScript"), "HTTP location of verification script")
	cmd.Flags().StringVar(&stepConfig.TeardownScript, "teardownScript", os.Getenv("PIPER_teardownScript"), "HTTP location of teardown script")

	cmd.MarkFlagRequired("containerRegistryUrl")
	cmd.MarkFlagRequired("deployTool")
	cmd.Flags().MarkDeprecated("image", "This parameter is deprecated, please use [containerImageName](#containerimagename) and [containerImageTag](#containerimagetag)")
}

// retrieve step metadata
func kubernetesDeployMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "kubernetesDeploy",
			Aliases:     []config.Alias{{Name: "deployToKubernetes", Deprecated: true}},
			Description: "Deployment to Kubernetes test or production namespace within the specified Kubernetes cluster.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "kubeConfigFileCredentialsId", Description: "Jenkins 'Secret file' credentials ID containing kubeconfig file. Details can be found in the [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/).", Type: "jenkins", Aliases: []config.Alias{{Name: "kubeCredentialsId", Deprecated: true}}},
					{Name: "kubeTokenCredentialsId", Description: "Jenkins 'Secret text' credentials ID containing token to authenticate to Kubernetes. This is an alternative way to using a kubeconfig file. Details can be found in the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/).", Type: "jenkins", Aliases: []config.Alias{{Name: "k8sTokenCredentialsId", Deprecated: true}}},
					{Name: "dockerCredentialsId", Type: "jenkins"},
					{Name: "dockerConfigJsonCredentialsId", Description: "Jenkins 'Secret file' credentials ID containing Docker config.json (with registry credential(s)).", Type: "jenkins"},
					{Name: "githubTokenCredentialsId", Description: "Jenkins credentials ID containing the github token.", Type: "jenkins"},
				},
				Resources: []config.StepResources{
					{Name: "deployDescriptor", Type: "stash"},
					{Name: "downloadedArtifact", Type: "stash"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "additionalParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "helmDeploymentParameters"}},
						Default:     []string{},
					},
					{
						Name:        "apiServer",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "k8sAPIServer"}},
						Default:     os.Getenv("PIPER_apiServer"),
					},
					{
						Name:        "appTemplate",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "k8sAppTemplate"}},
						Default:     os.Getenv("PIPER_appTemplate"),
					},
					{
						Name: "chartPath",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/localHelmChartPath",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "helmChartPath"}},
						Default:   os.Getenv("PIPER_chartPath"),
					},
					{
						Name: "containerRegistryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "dockerCredentialsId",
								Param: "password",
								Type:  "secret",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "container/repositoryPassword",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_containerRegistryPassword"),
					},
					{
						Name:        "containerImageName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "dockerImageName"}},
						Default:     os.Getenv("PIPER_containerImageName"),
					},
					{
						Name: "containerImageTag",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "artifactVersion"}},
						Default:   os.Getenv("PIPER_containerImageTag"),
					},
					{
						Name: "containerRegistryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/registryUrl",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "dockerRegistryUrl"}},
						Default:   os.Getenv("PIPER_containerRegistryUrl"),
					},
					{
						Name: "containerRegistryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "dockerCredentialsId",
								Param: "username",
								Type:  "secret",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "container/repositoryUsername",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUsername",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_containerRegistryUser"),
					},
					{
						Name:        "containerRegistrySecret",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `regsecret`,
					},
					{
						Name:        "createDockerRegistrySecret",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "deploymentName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "helmDeploymentName"}},
						Default:     os.Getenv("PIPER_deploymentName"),
					},
					{
						Name:        "deployTool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     `kubectl`,
					},
					{
						Name:        "forceUpdates",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "force"}},
						Default:     true,
					},
					{
						Name:        "helmDeployWaitSeconds",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     300,
					},
					{
						Name:        "helmTestWaitSeconds",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     0,
					},
					{
						Name:        "helmValues",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "valuesMapping",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "map[string]interface{}",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name: "githubToken",
						ResourceRef: []config.ResourceReference{
							{
								Name: "githubTokenCredentialsId",
								Type: "secret",
							},

							{
								Name:    "githubVaultSecretName",
								Type:    "vaultSecret",
								Default: "github",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "access_token"}},
						Default:   os.Getenv("PIPER_githubToken"),
					},
					{
						Name: "image",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/imageNameTag",
							},
						},
						Scope:              []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:               "string",
						Mandatory:          false,
						Aliases:            []config.Alias{{Name: "deployImage"}},
						Default:            os.Getenv("PIPER_image"),
						DeprecationMessage: "This parameter is deprecated, please use [containerImageName](#containerimagename) and [containerImageTag](#containerimagetag)",
					},
					{
						Name: "imageNames",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/imageNames",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "[]string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   []string{},
					},
					{
						Name: "imageNameTags",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/imageNameTags",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "[]string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   []string{},
					},
					{
						Name: "imageDigests",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/imageDigests",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "[]string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   []string{},
					},
					{
						Name:        "ingressHosts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "keepFailedDeployments",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "runHelmTests",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "showTestLogs",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "kubeConfig",
						ResourceRef: []config.ResourceReference{
							{
								Name: "kubeConfigFileCredentialsId",
								Type: "secret",
							},

							{
								Name:    "kubeConfigFileVaultSecretName",
								Type:    "vaultSecretFile",
								Default: "kube-config",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_kubeConfig"),
					},
					{
						Name:        "kubeContext",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_kubeContext"),
					},
					{
						Name: "kubeToken",
						ResourceRef: []config.ResourceReference{
							{
								Name: "kubeTokenCredentialsId",
								Type: "secret",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_kubeToken"),
					},
					{
						Name:        "namespace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "helmDeploymentNamespace"}, {Name: "k8sDeploymentNamespace"}},
						Default:     `default`,
					},
					{
						Name:        "tillerNamespace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "helmTillerNamespace"}},
						Default:     os.Getenv("PIPER_tillerNamespace"),
					},
					{
						Name: "dockerConfigJSON",
						ResourceRef: []config.ResourceReference{
							{
								Name: "dockerConfigJsonCredentialsId",
								Type: "secret",
							},

							{
								Name:    "dockerConfigFileVaultSecretName",
								Type:    "vaultSecretFile",
								Default: "docker-config",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   `.pipeline/docker/config.json`,
					},
					{
						Name:        "deployCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `apply`,
					},
					{
						Name:        "setupScript",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_setupScript"),
					},
					{
						Name:        "verificationScript",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_verificationScript"),
					},
					{
						Name:        "teardownScript",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_teardownScript"),
					},
				},
			},
			Containers: []config.Container{
				{Image: "dtzar/helm-kubectl:3", WorkingDir: "/config", Options: []config.Option{{Name: "-u", Value: "0"}}, Conditions: []config.Condition{{ConditionRef: "strings-equal", Params: []config.Param{{Name: "deployTool", Value: "helm3"}}}}},
				{Image: "dtzar/helm-kubectl:2.17.0", WorkingDir: "/config", Options: []config.Option{{Name: "-u", Value: "0"}}, Conditions: []config.Condition{{ConditionRef: "strings-equal", Params: []config.Param{{Name: "deployTool", Value: "helm"}}}}},
				{Image: "dtzar/helm-kubectl:2.17.0", WorkingDir: "/config", Options: []config.Option{{Name: "-u", Value: "0"}}, Conditions: []config.Condition{{ConditionRef: "strings-equal", Params: []config.Param{{Name: "deployTool", Value: "kubectl"}}}}},
			},
		},
	}
	return theMetaData
}
