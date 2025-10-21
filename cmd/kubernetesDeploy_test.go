//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type kubernetesDeployMockUtils struct {
	shouldFail bool
	*mock.FilesMock
	*mock.ExecMockRunner
	*mock.HttpClientMock
}

func newKubernetesDeployMockUtils() kubernetesDeployMockUtils {
	utils := kubernetesDeployMockUtils{
		shouldFail:     false,
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func TestRunKubernetesDeploy(t *testing.T) {

	t.Run("test helm", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm",
			ForceUpdates:              true,
			RenderSubchartNotes:       true,
			HelmDeployWaitSeconds:     400,
			IngressHosts:              []string{"ingress.host1", "ingress.host2"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		telemetryData := &telemetry.CustomData{}

		runKubernetesDeploy(opts, telemetryData, mockUtils, &stdout)

		assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, mockUtils.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"create", "secret", "generic", "testSecret", "--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson", "--insecure-skip-tls-verify=true", "--dry-run=client", "--output=json"},
			mockUtils.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2",
			"--force",
			"--wait",
			"--timeout",
			"400",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--render-subchart-notes",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[2].Params, "Wrong upgrade parameters")

		assert.Equal(t, &telemetry.CustomData{
			DeployTool: "helm",
		}, telemetryData)
	})

	t.Run("test helm - with containerImageName and containerImageTag instead of image", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			IngressHosts:              []string{"ingress.host1", "ingress.host2"},
			ContainerImageTag:         "latest",
			ContainerImageName:        "path/to/Image",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, mockUtils.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong upgrade command")

		assert.Contains(t, mockUtils.Calls[2].Params, "image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2", "Wrong upgrade parameters")
	})

	t.Run("test helm - docker config.json path passed as parameter", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			DockerConfigJSON:          "/path/to/.docker/config.json",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			IngressHosts:              []string{"ingress.host1", "ingress.host2"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			InsecureSkipTLSVerify:     true,
		}

		k8sSecretSpec := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("/path/to/.docker/config.json", []byte(`{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`))
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=/path/to/.docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: k8sSecretSpec,
		}

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.NoError(t, err)

		assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, mockUtils.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=/path/to/.docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2",
			"--force",
			"--wait",
			"--timeout",
			"400",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[2].Params, "Wrong upgrade parameters")
	})

	t.Run("test helm -- keep failed deployment", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			IngressHosts:              []string{"ingress.host1", "ingress.host2"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			KeepFailedDeployments:     true,
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		k8sSecretSpec := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: k8sSecretSpec,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, mockUtils.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2",
			"--force",
			"--wait",
			"--timeout",
			"400",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[2].Params, "Wrong upgrade parameters")
	})

	t.Run("test helm - fails without image information", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "testSecret",
			ChartPath:               "path/to/chart",
			DeploymentName:          "deploymentName",
			DeployTool:              "helm",
			ForceUpdates:            true,
			HelmDeployWaitSeconds:   400,
			IngressHosts:            []string{},
			AdditionalParameters:    []string{"--testParam", "testValue"},
			KubeContext:             "testCluster",
			Namespace:               "deploymentNamespace",
		}
		mockUtils := newKubernetesDeployMockUtils()

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.EqualError(t, err, "failed to process deployment values: image information not given - please either set image or containerImageName and containerImageTag")
	})

	t.Run("test helm - insecureSkipTLSVerify is false", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm",
			ForceUpdates:              true,
			RenderSubchartNotes:       true,
			HelmDeployWaitSeconds:     400,
			IngressHosts:              []string{"ingress.host1", "ingress.host2"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=false --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		telemetryData := &telemetry.CustomData{}

		runKubernetesDeploy(opts, telemetryData, mockUtils, &stdout)

		assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, mockUtils.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"create", "secret", "generic", "testSecret", "--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson", "--insecure-skip-tls-verify=false", "--dry-run=client", "--output=json"},
			mockUtils.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2",
			"--force",
			"--wait",
			"--timeout",
			"400",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--render-subchart-notes",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[2].Params, "Wrong upgrade parameters")

		assert.Equal(t, &telemetry.CustomData{
			DeployTool: "helm",
		}, telemetryData)
	})

	t.Run("test helm v3", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		telemetryData := &telemetry.CustomData{}
		runKubernetesDeploy(opts, telemetryData, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--values",
			"values1.yaml",
			"--values",
			"values2.yaml",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[1].Params, "Wrong upgrade parameters")

		assert.Equal(t, &telemetry.CustomData{
			DeployTool: "helm3",
		}, telemetryData)
	})

	t.Run("test helm v3 - runs helm tests", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			RunHelmTests:              true,
			HelmTestWaitSeconds:       400,
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--values",
			"values1.yaml",
			"--values",
			"values2.yaml",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[1].Params, "Wrong upgrade parameters")

		assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong test command")
		assert.Equal(t, []string{
			"test",
			"deploymentName",
			"--namespace",
			"deploymentNamespace",
			"--kube-context",
			"testCluster",
			"--timeout",
			"400s",
		}, mockUtils.Calls[2].Params, "Wrong test parameters")
	})

	t.Run("test helm v3 - runs helm tests with logs", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			RunHelmTests:              true,
			ShowTestLogs:              true,
			HelmTestWaitSeconds:       400,
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--values",
			"values1.yaml",
			"--values",
			"values2.yaml",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[1].Params, "Wrong upgrade parameters")

		assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong test command")
		assert.Equal(t, []string{
			"test",
			"deploymentName",
			"--namespace",
			"deploymentNamespace",
			"--kube-context",
			"testCluster",
			"--timeout",
			"400s",
			"--logs",
		}, mockUtils.Calls[2].Params, "Wrong test parameters")
	})

	t.Run("test helm v3 - should not run helm tests", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			RunHelmTests:              false,
			ShowTestLogs:              true,
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--values",
			"values1.yaml",
			"--values",
			"values2.yaml",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[1].Params, "Wrong upgrade parameters")

		assert.Equal(t, 2, len(mockUtils.Calls), "Too many helm calls")
	})

	t.Run("test helm v3 - with containerImageName and containerImageTag instead of image", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			ContainerImageName:        "path/to/Image",
			ContainerImageTag:         "latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")

		assert.Contains(t, mockUtils.Calls[1].Params, "image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret", "Wrong upgrade parameters")

	})

	t.Run("test helm v3 - with multiple images", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			ImageNames:                []string{"myImage", "myImage.sub1", "myImage.sub2"},
			ImageNameTags:             []string{"myImage:myTag", "myImage-sub1:myTag", "myImage-sub2:myTag"},
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		require.NoError(t, runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout))

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")

		assert.Contains(t, mockUtils.Calls[1].Params, `image.myImage.repository=my.registry:55555/myImage,image.myImage.tag=myTag,myImage.image.repository=my.registry:55555/myImage,myImage.image.tag=myTag,myImage.image.repository=my.registry:55555/myImage,myImage.image.tag=myTag,image.myImage_sub1.repository=my.registry:55555/myImage-sub1,image.myImage_sub1.tag=myTag,myImage_sub1.image.repository=my.registry:55555/myImage-sub1,myImage_sub1.image.tag=myTag,myImage_sub1.image.repository=my.registry:55555/myImage-sub1,myImage_sub1.image.tag=myTag,image.myImage_sub2.repository=my.registry:55555/myImage-sub2,image.myImage_sub2.tag=myTag,myImage_sub2.image.repository=my.registry:55555/myImage-sub2,myImage_sub2.image.tag=myTag,myImage_sub2.image.repository=my.registry:55555/myImage-sub2,myImage_sub2.image.tag=myTag,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret`, "Wrong upgrade parameters")

	})

	t.Run("test helm v3 - with one image containing - in name", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			ImageNames:                []string{"my-Image"},
			ImageNameTags:             []string{"my-Image:myTag"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		require.NoError(t, runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout))

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
		assert.Contains(t, mockUtils.Calls[1].Params[11], "my-Image.image.tag=myTag")
		assert.Contains(t, mockUtils.Calls[1].Params[11], "my-Image.image.repository=")
	})

	t.Run("test helm v3 - with one image in  multiple images array", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			ImageNames:                []string{"myImage"},
			ImageNameTags:             []string{"myImage:myTag"},
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		require.NoError(t, runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout))

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")

		assert.Contains(t, mockUtils.Calls[1].Params, `image.myImage.repository=my.registry:55555/myImage,image.myImage.tag=myTag,myImage.image.repository=my.registry:55555/myImage,myImage.image.tag=myTag,myImage.image.repository=my.registry:55555/myImage,myImage.image.tag=myTag,image.repository=my.registry:55555/myImage,image.tag=myTag,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret`, "Wrong upgrade parameters")

	})

	t.Run("test helm v3 - with multiple images - missing ImageNameTags", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			ImageNames:                []string{"myImage", "myImage.sub1", "myImage.sub2"},
			ImageNameTags:             []string{"myImage:myTag"},
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.EqualError(t, err, "failed to process deployment values: number of imageNames and imageNameTags must be equal")
	})

	t.Run("test helm v3 - with multiple images and valuesMapping", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			ValuesMapping: map[string]interface{}{
				"subchart.image.registry": "image.myImage.repository",
				"subchart.image.tag":      "image.myImage.tag",
			},
			ImageNames:            []string{"myImage", "myImage.sub1", "myImage.sub2"},
			ImageNameTags:         []string{"myImage:myTag", "myImage-sub1:myTag", "myImage-sub2:myTag"},
			AdditionalParameters:  []string{"--testParam", "testValue"},
			KubeContext:           "testCluster",
			Namespace:             "deploymentNamespace",
			DockerConfigJSON:      ".pipeline/docker/config.json",
			InsecureSkipTLSVerify: true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		require.NoError(t, runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout))

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
		assert.Equal(t, len(mockUtils.Calls[1].Params), 21, "Unexpected upgrade command")
		pos := 11
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "image.myImage.repository=my.registry:55555/myImage", "Missing update parameter")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "image.myImage.tag=myTag", "Wrong upgrade parameters")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "image.myImage_sub1.repository=my.registry:55555/myImage-sub1", "Missing update parameter")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "image.myImage_sub1.tag=myTag", "Missing update parameter")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "image.myImage_sub2.repository=my.registry:55555/myImage-sub2", "Missing update parameter")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "image.myImage_sub2.tag=myTag")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==", "Missing update parameter")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "imagePullSecrets[0].name=testSecret", "Missing update parameter")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "subchart.image.registry=my.registry:55555/myImage", "Missing update parameter")
		assert.Contains(t, mockUtils.Calls[1].Params[pos], "subchart.image.tag=myTag", "Missing update parameter")
	})

	t.Run("test helm v3 - with multiple images and incorrect valuesMapping", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			ValuesMapping: map[string]interface{}{
				"subchart.image.registry": false,
			},
			ImageNames:            []string{"myImage", "myImage.sub1", "myImage.sub2"},
			ImageNameTags:         []string{"myImage:myTag", "myImage-sub1:myTag", "myImage-sub2:myTag"},
			AdditionalParameters:  []string{"--testParam", "testValue"},
			KubeContext:           "testCluster",
			Namespace:             "deploymentNamespace",
			DockerConfigJSON:      ".pipeline/docker/config.json",
			InsecureSkipTLSVerify: true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		require.Error(t, runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout), "invalid path 'false' is used for valueMapping, only strings are supported")

	})

	t.Run("test helm3 - fails without image information", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "testSecret",
			ChartPath:               "path/to/chart",
			DeploymentName:          "deploymentName",
			DeployTool:              "helm3",
			ForceUpdates:            true,
			HelmDeployWaitSeconds:   400,
			IngressHosts:            []string{},
			AdditionalParameters:    []string{"--testParam", "testValue"},
			KubeContext:             "testCluster",
			Namespace:               "deploymentNamespace",
		}
		mockUtils := newKubernetesDeployMockUtils()

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.EqualError(t, err, "failed to process deployment values: image information not given - please either set image or containerImageName and containerImageTag")
	})

	t.Run("test helm v3 - keep failed deployments", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			KeepFailedDeployments:     true,
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     true,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--values",
			"values1.yaml",
			"--values",
			"values2.yaml",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[1].Params, "Wrong upgrade parameters")
	})

	t.Run("test helm v3 - no container credentials", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:    "https://my.registry:55555",
			ChartPath:               "path/to/chart",
			ContainerRegistrySecret: "testSecret",
			DeploymentName:          "deploymentName",
			DeployTool:              "helm3",
			ForceUpdates:            true,
			HelmDeployWaitSeconds:   400,
			IngressHosts:            []string{},
			Image:                   "path/to/Image:latest",
			AdditionalParameters:    []string{"--testParam", "testValue"},
			KubeContext:             "testCluster",
			Namespace:               "deploymentNamespace",
		}
		mockUtils := newKubernetesDeployMockUtils()

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, 1, len(mockUtils.Calls), "Wrong number of upgrade commands")
		assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[0].Params, "Wrong upgrade parameters")
	})

	t.Run("test helm v3 - insecureSkipTLSVerify is false", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "dummy",
			ContainerRegistrySecret:   "testSecret",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm3",
			ForceUpdates:              true,
			HelmDeployWaitSeconds:     400,
			HelmValues:                []string{"values1.yaml", "values2.yaml"},
			Image:                     "path/to/Image:latest",
			AdditionalParameters:      []string{"--testParam", "testValue"},
			KubeContext:               "testCluster",
			Namespace:                 "deploymentNamespace",
			DockerConfigJSON:          ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:     false,
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=false --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		telemetryData := &telemetry.CustomData{}
		runKubernetesDeploy(opts, telemetryData, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=false",
			"--dry-run=client",
			"--output=json"},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--values",
			"values1.yaml",
			"--values",
			"values2.yaml",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[1].Params, "Wrong upgrade parameters")

		assert.Equal(t, &telemetry.CustomData{
			DeployTool: "helm3",
		}, telemetryData)
	})

	t.Run("test helm - use extensions", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:    "https://my.registry:55555",
			ChartPath:               "path/to/chart",
			ContainerRegistrySecret: "testSecret",
			DeploymentName:          "deploymentName",
			DeployTool:              "helm3",
			IngressHosts:            []string{},
			Image:                   "path/to/Image:latest",
			KubeContext:             "testCluster",
			Namespace:               "deploymentNamespace",
			GithubToken:             "testGHToken",
			SetupScript:             "https://github.com/my/test/setup_script.sh",
			VerificationScript:      "https://github.com/my/test/verification_script.sh",
			TeardownScript:          "https://github.com/my/test/teardown_script.sh",
		}
		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.HttpClientMock = &mock.HttpClientMock{HTTPFileUtils: mockUtils.FilesMock}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, 4, len(mockUtils.Calls))
		assert.Equal(t, ".pipeline/setup_script.sh", mockUtils.Calls[0].Exec)
		assert.Equal(t, ".pipeline/verification_script.sh", mockUtils.Calls[2].Exec)
		assert.Equal(t, ".pipeline/teardown_script.sh", mockUtils.Calls[3].Exec)
	})

	t.Run("test helm v3 - fails without chart path", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "testSecret",
			DeploymentName:          "deploymentName",
			DeployTool:              "helm3",
			ForceUpdates:            true,
			HelmDeployWaitSeconds:   400,
			IngressHosts:            []string{},
			Image:                   "path/to/Image:latest",
			AdditionalParameters:    []string{"--testParam", "testValue"},
			KubeContext:             "testCluster",
			Namespace:               "deploymentNamespace",
		}
		mockUtils := newKubernetesDeployMockUtils()

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.EqualError(t, err, "chart path has not been set, please configure chartPath parameter")
	})

	t.Run("test helm v3 - fails without deployment name", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "testSecret",
			ChartPath:               "path/to/chart",
			DeployTool:              "helm3",
			ForceUpdates:            true,
			HelmDeployWaitSeconds:   400,
			IngressHosts:            []string{},
			Image:                   "path/to/Image:latest",
			AdditionalParameters:    []string{"--testParam", "testValue"},
			KubeContext:             "testCluster",
			Namespace:               "deploymentNamespace",
		}
		mockUtils := newKubernetesDeployMockUtils()

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.EqualError(t, err, "deployment name has not been set, please configure deploymentName parameter")
	})

	t.Run("test helm v3 - no force", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:    "https://my.registry:55555",
			ChartPath:               "path/to/chart",
			ContainerRegistrySecret: "testSecret",
			DeploymentName:          "deploymentName",
			DeployTool:              "helm3",
			HelmDeployWaitSeconds:   400,
			IngressHosts:            []string{},
			Image:                   "path/to/Image:latest",
			AdditionalParameters:    []string{"--testParam", "testValue"},
			KubeContext:             "testCluster",
			Namespace:               "deploymentNamespace",
		}
		mockUtils := newKubernetesDeployMockUtils()

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,image.path/to/Image.repository=my.registry:55555/path/to/Image,image.path/to/Image.tag=latest,imagePullSecrets[0].name=testSecret",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[0].Params, "Wrong upgrade parameters")
	})

	t.Run("test kubectl - create secret from docker config.json", func(t *testing.T) {

		opts := kubernetesDeployOptions{
			AppTemplate:                "path/to/test.yaml",
			ContainerRegistryURL:       "https://my.registry:55555",
			ContainerRegistryUser:      "registryUser",
			ContainerRegistryPassword:  "dummy",
			ContainerRegistrySecret:    "regSecret",
			CreateDockerRegistrySecret: true,
			DeployTool:                 "kubectl",
			Image:                      "path/to/Image:latest",
			AdditionalParameters:       []string{"--testParam", "testValue"},
			KubeConfig:                 "This is my kubeconfig",
			KubeContext:                "testCluster",
			Namespace:                  "deploymentNamespace",
			DeployCommand:              "apply",
			DockerConfigJSON:           ".pipeline/docker/config.json",
			InsecureSkipTLSVerify:      true,
		}

		kubeYaml := `kind: Deployment
		metadata:
		spec:
		 spec:
		   image: <image-name>`

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile(opts.AppTemplate, []byte(kubeYaml))

		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic regSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json --namespace=deploymentNamespace`: dockerConfigJSON,
		}

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, mockUtils.Env, []string{"KUBECONFIG=This is my kubeconfig"})

		assert.Equal(t, []string{
			"create",
			"secret",
			"generic",
			"regSecret",
			"--from-file=.dockerconfigjson=.pipeline/docker/config.json",
			"--type=kubernetes.io/dockerconfigjson",
			"--insecure-skip-tls-verify=true",
			"--dry-run=client",
			"--output=json",
			"--namespace=deploymentNamespace",
			"--insecure-skip-tls-verify=true",
			"--context=testCluster",
		},
			mockUtils.Calls[0].Params, "Wrong secret creation parameters")

		assert.Containsf(t, mockUtils.Calls[1].Params, "apply", "Wrong secret creation parameters")
		assert.Containsf(t, mockUtils.Calls[1].Params, "-f", "Wrong secret creation parameters")
	})

	t.Run("test kubectl - token only", func(t *testing.T) {

		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "path/to/test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			Image:                   "path/to/Image:latest",
			KubeToken:               "testToken",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
			InsecureSkipTLSVerify:   true,
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile(opts.AppTemplate, []byte("testYaml"))
		mockUtils.ShouldFailOnCommand = map[string]error{}

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--server=%v", opts.APIServer),
			fmt.Sprintf("--token=%v", opts.KubeToken),
			"apply",
			"--filename",
			opts.AppTemplate,
		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - with containerImageName and containerImageTag instead of image", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			ContainerImageTag:       "latest",
			ContainerImageName:      "path/to/Image",
			KubeConfig:              "This is my kubeconfig",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("test.yaml", []byte("image: <image-name>"))

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")

		appTemplateFileContents, err := mockUtils.FileRead(opts.AppTemplate)
		assert.NoError(t, err)
		assert.Contains(t, string(appTemplateFileContents), "image: my.registry:55555/path/to/Image:latest", "kubectl parameters incorrect")
	})

	t.Run("test kubectl - with containerImageName and containerImageTag instead of image using go template", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			ContainerImageTag:       "latest",
			ContainerImageName:      "path/to/Image",
			KubeConfig:              "This is my kubeconfig",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("test.yaml", []byte("image: {{ .Values.image.repository }}:{{ .Values.image.tag }}"))

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")

		appTemplateFileContents, err := mockUtils.FileRead(opts.AppTemplate)
		assert.NoError(t, err)
		assert.Contains(t, string(appTemplateFileContents), "image: my.registry:55555/path/to/Image:latest", "kubectl parameters incorrect")
	})

	t.Run("test kubectl - with multiple images using go template", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			KubeConfig:              "This is my kubeconfig",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
			ValuesMapping: map[string]interface{}{
				"subchart.image.repository": "image.myImage.repository",
				"subchart.image.tag":        "image.myImage.tag",
			},
			ImageNames:    []string{"myImage", "myImage-sub1", "myImage-sub2"},
			ImageNameTags: []string{"myImage:myTag", "myImage-sub1:myTag", "myImage-sub2:myTag"},
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("test.yaml", []byte(`image: {{ .Values.image.myImage.repository }}:{{ .Values.image.myImage.tag }}
image2: {{ .Values.subchart.image.repository }}:{{ .Values.subchart.image.tag }}
image3: {{ .Values.image.myImage_sub1.repository }}:{{ .Values.image.myImage_sub1.tag }}`))

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")

		appTemplateFileContents, err := mockUtils.FileRead(opts.AppTemplate)
		assert.NoError(t, err)
		assert.Contains(t, string(appTemplateFileContents), "image: my.registry:55555/myImage:myTag\nimage2: my.registry:55555/myImage:myTag\nimage3: my.registry:55555/myImage-sub1:myTag", "kubectl parameters incorrect")
	})

	t.Run("test kubectl - with multiple images and digests", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			KubeConfig:              "This is my kubeconfig",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
			ValuesMapping: map[string]interface{}{
				"subchart.image.repository": "image.myImage.repository",
				"subchart.image.tag":        "image.myImage.tag",
			},
			ImageNames:    []string{"myImage", "myImage-sub1", "myImage-sub2"},
			ImageNameTags: []string{"myImage:myTag", "myImage-sub1:myTag", "myImage-sub2:myTag"},
			ImageDigests:  []string{"sha256:111", "sha256:222", "sha256:333"},
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("test.yaml", []byte(`image: {{ .Values.image.myImage.repository }}:{{ .Values.image.myImage.tag }}
image2: {{ .Values.subchart.image.repository }}:{{ .Values.subchart.image.tag }}
image3: {{ .Values.image.myImage_sub1.repository }}:{{ .Values.image.myImage_sub1.tag }}
image4: {{ .Values.image.myImage_sub2.repository }}:{{ .Values.image.myImage_sub2.tag }}`))

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")

		appTemplateFileContents, err := mockUtils.FileRead(opts.AppTemplate)
		assert.NoError(t, err)
		assert.Contains(t, string(appTemplateFileContents), `image: my.registry:55555/myImage:myTag@sha256:111
image2: my.registry:55555/myImage:myTag@sha256:111
image3: my.registry:55555/myImage-sub1:myTag@sha256:222
image4: my.registry:55555/myImage-sub2:myTag@sha256:333`, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - fail with multiple images using placeholder", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			KubeConfig:              "This is my kubeconfig",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
			ImageNames:              []string{"myImage", "myImage-sub1", "myImage-sub2"},
			ImageNameTags:           []string{"myImage:myTag", "myImage-sub1:myTag", "myImage-sub2:myTag"},
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("test.yaml", []byte("image: <image-name>"))

		var stdout bytes.Buffer
		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.EqualError(t, err, "multi-image replacement not supported for single image placeholder")
	})

	t.Run("test kubectl - fails without image information", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			KubeConfig:              "This is my kubeconfig",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("test.yaml", []byte("testYaml"))

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.EqualError(t, err, "failed to process deployment values: image information not given - please either set image or containerImageName and containerImageTag")
	})

	t.Run("test kubectl - use replace deploy command", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			AppTemplate:                "test.yaml",
			ContainerRegistryURL:       "https://my.registry:55555",
			ContainerRegistrySecret:    "regSecret",
			CreateDockerRegistrySecret: true,
			DeployTool:                 "kubectl",
			Image:                      "path/to/Image:latest",
			AdditionalParameters:       []string{"--testParam", "testValue"},
			KubeConfig:                 "This is my kubeconfig",
			KubeContext:                "testCluster",
			Namespace:                  "deploymentNamespace",
			DeployCommand:              "replace",
			InsecureSkipTLSVerify:      true,
		}

		kubeYaml := `kind: Deployment
	metadata:
	spec:
	  spec:
	    image: <image-name>`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("test.yaml", []byte(kubeYaml))

		var stdout bytes.Buffer
		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.NoError(t, err, "Command should not fail")

		assert.Equal(t, mockUtils.Env, []string{"KUBECONFIG=This is my kubeconfig"})

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong replace command")
		assert.Equal(t, []string{
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--context=%v", opts.KubeContext),
			"replace",
			"--filename",
			opts.AppTemplate,
			"--testParam",
			"testValue",
		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")

		appTemplate, err := mockUtils.FileRead(opts.AppTemplate)
		assert.Contains(t, string(appTemplate), "my.registry:55555/path/to/Image:latest")
	})

	t.Run("test kubectl - use replace --force deploy command", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			AppTemplate:                "test.yaml",
			ContainerRegistryURL:       "https://my.registry:55555",
			ContainerRegistrySecret:    "regSecret",
			CreateDockerRegistrySecret: true,
			DeployTool:                 "kubectl",
			Image:                      "path/to/Image:latest",
			AdditionalParameters:       []string{"--testParam", "testValue"},
			KubeConfig:                 "This is my kubeconfig",
			KubeContext:                "testCluster",
			Namespace:                  "deploymentNamespace",
			DeployCommand:              "replace",
			ForceUpdates:               true,
			InsecureSkipTLSVerify:      true,
		}

		kubeYaml := `kind: Deployment
	metadata:
	spec:
	  spec:
	    image: <image-name>`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile("test.yaml", []byte(kubeYaml))

		var stdout bytes.Buffer
		err := runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)
		assert.NoError(t, err, "Command should not fail")

		assert.Equal(t, mockUtils.Env, []string{"KUBECONFIG=This is my kubeconfig"})

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong replace command")
		assert.Equal(t, []string{
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--context=%v", opts.KubeContext),
			"replace",
			"--filename",
			opts.AppTemplate,
			"--force",
			"--testParam",
			"testValue",
		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")

		appTemplate, err := mockUtils.FileRead(opts.AppTemplate)
		assert.Contains(t, string(appTemplate), "my.registry:55555/path/to/Image:latest")
	})

	t.Run("test kubectl - insecureSkipTLSVerify is false", func(t *testing.T) {

		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "path/to/test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			Image:                   "path/to/Image:latest",
			KubeToken:               "testToken",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
			InsecureSkipTLSVerify:   false,
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile(opts.AppTemplate, []byte("testYaml"))
		mockUtils.ShouldFailOnCommand = map[string]error{}

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"--insecure-skip-tls-verify=false",
			fmt.Sprintf("--server=%v", opts.APIServer),
			fmt.Sprintf("--token=%v", opts.KubeToken),
			"apply",
			"--filename",
			opts.AppTemplate,
		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - insecureSkipTLSVerify is false with custom CA", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "path/to/test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			Image:                   "path/to/Image:latest",
			KubeToken:               "testToken",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
			CACertificate:           "path/to/ca.crt",
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile(opts.AppTemplate, []byte("testYaml"))
		mockUtils.ShouldFailOnCommand = map[string]error{}

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			fmt.Sprintf("--certificate-authority=%v", opts.CACertificate),
			"--insecure-skip-tls-verify=false",
			fmt.Sprintf("--server=%v", opts.APIServer),
			fmt.Sprintf("--token=%v", opts.KubeToken),
			"apply",
			"--filename",
			opts.AppTemplate,
		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - insecureSkipTLSVerify is false without custom CA", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "path/to/test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			Image:                   "path/to/Image:latest",
			KubeToken:               "testToken",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
			InsecureSkipTLSVerify:   false,
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile(opts.AppTemplate, []byte("testYaml"))
		mockUtils.ShouldFailOnCommand = map[string]error{}

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"--insecure-skip-tls-verify=false",
			fmt.Sprintf("--server=%v", opts.APIServer),
			fmt.Sprintf("--token=%v", opts.KubeToken),
			"apply",
			"--filename",
			opts.AppTemplate,
		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - insecureSkipTLSVerify is true with custom CA", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			APIServer:               "https://my.api.server",
			AppTemplate:             "path/to/test.yaml",
			ContainerRegistryURL:    "https://my.registry:55555",
			ContainerRegistrySecret: "regSecret",
			DeployTool:              "kubectl",
			Image:                   "path/to/Image:latest",
			KubeToken:               "testToken",
			Namespace:               "deploymentNamespace",
			DeployCommand:           "apply",
			InsecureSkipTLSVerify:   true,
			CACertificate:           "path/to/ca.crt",
		}

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.AddFile(opts.AppTemplate, []byte("testYaml"))
		mockUtils.ShouldFailOnCommand = map[string]error{}

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &telemetry.CustomData{}, mockUtils, &stdout)

		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--server=%v", opts.APIServer),
			fmt.Sprintf("--token=%v", opts.KubeToken),
			"apply",
			"--filename",
			opts.AppTemplate,
		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")
	})

}

func TestSplitRegistryURL(t *testing.T) {
	tt := []struct {
		in          string
		outProtocol string
		outRegistry string
		outError    error
	}{
		{in: "https://my.registry.com", outProtocol: "https", outRegistry: "my.registry.com", outError: nil},
		{in: "https://", outProtocol: "", outRegistry: "", outError: fmt.Errorf("Failed to split registry url 'https://'")},
		{in: "my.registry.com", outProtocol: "", outRegistry: "", outError: fmt.Errorf("Failed to split registry url 'my.registry.com'")},
		{in: "", outProtocol: "", outRegistry: "", outError: fmt.Errorf("Failed to split registry url ''")},
		{in: "https://https://my.registry.com", outProtocol: "", outRegistry: "", outError: fmt.Errorf("Failed to split registry url 'https://https://my.registry.com'")},
	}

	for _, test := range tt {
		p, r, err := splitRegistryURL(test.in)
		assert.Equal(t, test.outProtocol, p, "Protocol value unexpected")
		assert.Equal(t, test.outRegistry, r, "Registry value unexpected")
		assert.Equal(t, test.outError, err, "Error value not as expected")
	}

}

func TestSplitImageName(t *testing.T) {
	tt := []struct {
		in       string
		outImage string
		outTag   string
		outError error
	}{
		{in: "", outImage: "", outTag: "", outError: fmt.Errorf("Failed to split image name ''")},
		{in: "path/to/image", outImage: "path/to/image", outTag: "", outError: nil},
		{in: "path/to/image:tag", outImage: "path/to/image", outTag: "tag", outError: nil},
		{in: "https://my.registry.com/path/to/image:tag", outImage: "", outTag: "", outError: fmt.Errorf("Failed to split image name 'https://my.registry.com/path/to/image:tag'")},
	}
	for _, test := range tt {
		i, tag, err := splitFullImageName(test.in)
		assert.Equal(t, test.outImage, i, "Image value unexpected")
		assert.Equal(t, test.outTag, tag, "Tag value unexpected")
		assert.Equal(t, test.outError, err, "Error value not as expected")
	}
}
