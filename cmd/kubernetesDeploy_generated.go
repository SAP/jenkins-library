package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/spf13/cobra"
)

type kubernetesDeployOptions struct {
	AdditionalParameters       []string `json:"additionalParameters,omitempty"`
	APIServer                  string   `json:"apiServer,omitempty"`
	AppTemplate                string   `json:"appTemplate,omitempty"`
	ChartPath                  string   `json:"chartPath,omitempty"`
	ContainerRegistryPassword  string   `json:"containerRegistryPassword,omitempty"`
	ContainerRegistryURL       string   `json:"containerRegistryUrl,omitempty"`
	ContainerRegistryUser      string   `json:"containerRegistryUser,omitempty"`
	ContainerRegistrySecret    string   `json:"containerRegistrySecret,omitempty"`
	CreateDockerRegistrySecret bool     `json:"createDockerRegistrySecret,omitempty"`
	DeploymentName             string   `json:"deploymentName,omitempty"`
	DeployTool                 string   `json:"deployTool,omitempty"`
	HelmDeployWaitSeconds      int      `json:"helmDeployWaitSeconds,omitempty"`
	Image                      string   `json:"image,omitempty"`
	IngressHosts               []string `json:"ingressHosts,omitempty"`
	KubeConfig                 string   `json:"kubeConfig,omitempty"`
	KubeContext                string   `json:"kubeContext,omitempty"`
	KubeToken                  string   `json:"kubeToken,omitempty"`
	Namespace                  string   `json:"namespace,omitempty"`
	TillerNamespace            string   `json:"tillerNamespace,omitempty"`
}

var myKubernetesDeployOptions kubernetesDeployOptions

// KubernetesDeployCommand Deployment to Kubernetes test or production namespace within the specified Kubernetes cluster.
func KubernetesDeployCommand() *cobra.Command {
	metadata := kubernetesDeployMetadata()

	var createKubernetesDeployCmd = &cobra.Command{
		Use:   "kubernetesDeploy",
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
* ` + "`" + `dockerSecret` + "`" + ` will be calculated with a call to ` + "`" + `kubectl create secret docker-registry regsecret --docker-server=<yourRegistry> --docker-username=<containerRegistryUser> --docker-password=<containerRegistryPassword> --dry-run=true --output=json'` + "`" + ``,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("kubernetesDeploy")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "kubernetesDeploy", &myKubernetesDeployOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			telemetry.Initialize(GeneralConfig.NoTelemetry, "kubernetesDeploy")
			telemetry.Send(&telemetry.CustomData{})
			return kubernetesDeploy(myKubernetesDeployOptions)
		},
	}

	addKubernetesDeployFlags(createKubernetesDeployCmd)
	return createKubernetesDeployCmd
}

func addKubernetesDeployFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&myKubernetesDeployOptions.AdditionalParameters, "additionalParameters", []string{}, "Defines additional parameters for \"helm install\" or \"kubectl apply\" command.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.APIServer, "apiServer", os.Getenv("PIPER_apiServer"), "Defines the Url of the API Server of the Kubernetes cluster.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.AppTemplate, "appTemplate", os.Getenv("PIPER_appTemplate"), "Defines the filename for the kubernetes app template (e.g. k8s_apptemplate.yaml)")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.ChartPath, "chartPath", os.Getenv("PIPER_chartPath"), "Defines the chart path for deployments using helm.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.ContainerRegistryPassword, "containerRegistryPassword", os.Getenv("PIPER_containerRegistryPassword"), "Password for container registry access - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.ContainerRegistryURL, "containerRegistryUrl", os.Getenv("PIPER_containerRegistryUrl"), "http(s) url of the Container registry.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.ContainerRegistryUser, "containerRegistryUser", os.Getenv("PIPER_containerRegistryUser"), "Username for container registry access - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.ContainerRegistrySecret, "containerRegistrySecret", "regsecret", "Name of the container registry secret used for pulling containers from the registry.")
	cmd.Flags().BoolVar(&myKubernetesDeployOptions.CreateDockerRegistrySecret, "createDockerRegistrySecret", false, "Toggle to turn on Regsecret creation with a \"deployTool:kubectl\" deployment.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.DeploymentName, "deploymentName", os.Getenv("PIPER_deploymentName"), "Defines the name of the deployment.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.DeployTool, "deployTool", "kubectl", "Defines the tool which should be used for deployment.")
	cmd.Flags().IntVar(&myKubernetesDeployOptions.HelmDeployWaitSeconds, "helmDeployWaitSeconds", 300, "Number of seconds before helm deploy returns.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.Image, "image", os.Getenv("PIPER_image"), "Full name of the image to be deployed.")
	cmd.Flags().StringSliceVar(&myKubernetesDeployOptions.IngressHosts, "ingressHosts", []string{}, "List of ingress hosts to be exposed via helm deployment.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.KubeConfig, "kubeConfig", os.Getenv("PIPER_kubeConfig"), "Defines the path to the \"kubeconfig\" file.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.KubeContext, "kubeContext", os.Getenv("PIPER_kubeContext"), "Defines the context to use from the \"kubeconfig\" file.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.KubeToken, "kubeToken", os.Getenv("PIPER_kubeToken"), "Contains the id_token used by kubectl for authentication. Consider using kubeConfig parameter instead.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.Namespace, "namespace", "default", "Defines the target Kubernetes namespace for the deployment.")
	cmd.Flags().StringVar(&myKubernetesDeployOptions.TillerNamespace, "tillerNamespace", os.Getenv("PIPER_tillerNamespace"), "Defines optional tiller namespace for deployments using helm.")

	cmd.MarkFlagRequired("chartPath")
	cmd.MarkFlagRequired("containerRegistryUrl")
	cmd.MarkFlagRequired("deploymentName")
	cmd.MarkFlagRequired("deployTool")
	cmd.MarkFlagRequired("image")
}

// retrieve step metadata
func kubernetesDeployMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "additionalParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "helmDeploymentParameters"}},
					},
					{
						Name:        "apiServer",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "k8sAPIServer"}},
					},
					{
						Name:        "appTemplate",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "k8sAppTemplate"}},
					},
					{
						Name:        "chartPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "helmChartPath"}},
					},
					{
						Name:        "containerRegistryPassword",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "containerRegistryUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "dockerRegistryUrl"}},
					},
					{
						Name:        "containerRegistryUser",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "containerRegistrySecret",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "createDockerRegistrySecret",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "deploymentName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "helmDeploymentName"}},
					},
					{
						Name:        "deployTool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "helmDeployWaitSeconds",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "image",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "deployImage"}},
					},
					{
						Name:        "ingressHosts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "kubeConfig",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "kubeContext",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "kubeToken",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "namespace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "helmDeploymentNamespace"}, {Name: "k8sDeploymentNamespace"}},
					},
					{
						Name:        "tillerNamespace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "helmTillerNamespace"}},
					},
				},
			},
		},
	}
	return theMetaData
}
