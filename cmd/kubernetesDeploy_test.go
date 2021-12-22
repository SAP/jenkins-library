package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

type kubernetesDeployMockUtils struct {
	shouldFail     bool
	requestedUrls  []string
	requestedFiles []string
	*mock.FilesMock
	*mock.ExecMockRunner
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
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, mockUtils, &stdout)

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
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2",
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
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		mockUtils := newKubernetesDeployMockUtils()
		mockUtils.StdoutReturn = map[string]string{
			`kubectl create secret generic testSecret --from-file=.dockerconfigjson=.pipeline/docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: dockerConfigJSON,
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, mockUtils, &stdout)

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

		assert.Contains(t, mockUtils.Calls[2].Params, "image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2", "Wrong upgrade parameters")
	})

	// t.Run("test helm - docker config.json path passed as parameter", func(t *testing.T) {
	// 	opts := kubernetesDeployOptions{
	// 		ContainerRegistryURL:    "https://my.registry:55555",
	// 		DockerConfigJSON:        "/path/to/.docker/config.json",
	// 		ContainerRegistrySecret: "testSecret",
	// 		ChartPath:               "path/to/chart",
	// 		DeploymentName:          "deploymentName",
	// 		DeployTool:              "helm",
	// 		ForceUpdates:            true,
	// 		HelmDeployWaitSeconds:   400,
	// 		IngressHosts:            []string{"ingress.host1", "ingress.host2"},
	// 		Image:                   "path/to/Image:latest",
	// 		AdditionalParameters:    []string{"--testParam", "testValue"},
	// 		KubeContext:             "testCluster",
	// 		Namespace:               "deploymentNamespace",
	// 	}

	// 	k8sSecretSpec := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

	// 	mockUtils := newKubernetesDeployMockUtils()
	// 	mockUtils.AddFile("/path/to/.docker/config.json", []byte("ThisIsOurBase64EncodedSecret=="))
	// 	mockUtils.StdoutReturn = map[string]string{
	// 		`kubectl create secret generic testSecret --from-file=.dockerconfigjson=/path/to/.docker/config.json --type=kubernetes.io/dockerconfigjson --insecure-skip-tls-verify=true --dry-run=client --output=json`: k8sSecretSpec,
	// 	}

	// 	var stdout bytes.Buffer

	// 	err := runKubernetesDeploy(opts, mockUtils, &stdout)
	// 	assert.NoError(t, err)

	// 	assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong init command")
	// 	assert.Equal(t, []string{"init", "--client-only"}, mockUtils.Calls[0].Params, "Wrong init parameters")

	// 	assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong secret creation command")
	// 	assert.Equal(t, []string{
	// 		"create",
	// 		"secret",
	// 		"generic",
	// 		"testSecret",
	// 		"--from-file=.dockerconfigjson=/path/to/.docker/config.json",
	// 		"--type=kubernetes.io/dockerconfigjson",
	// 		"--insecure-skip-tls-verify=true",
	// 		"--dry-run=client",
	// 		"--output=json"},
	// 		mockUtils.Calls[1].Params, "Wrong secret creation parameters")

	// 	assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong upgrade command")
	// 	assert.Equal(t, []string{
	// 		"upgrade",
	// 		"deploymentName",
	// 		"path/to/chart",
	// 		"--install",
	// 		"--namespace",
	// 		"deploymentNamespace",
	// 		"--set",
	// 		"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2",
	// 		"--force",
	// 		"--wait",
	// 		"--timeout",
	// 		"400",
	// 		"--atomic",
	// 		"--kube-context",
	// 		"testCluster",
	// 		"--testParam",
	// 		"testValue",
	// 	}, mockUtils.Calls[2].Params, "Wrong upgrade parameters")
	// })

	// 	t.Run("test helm -- keep failed deployment", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:      "https://my.registry:55555",
	// 			ContainerRegistryUser:     "registryUser",
	// 			ContainerRegistryPassword: "********",
	// 			ContainerRegistrySecret:   "testSecret",
	// 			ChartPath:                 "path/to/chart",
	// 			DeploymentName:            "deploymentName",
	// 			DeployTool:                "helm",
	// 			ForceUpdates:              true,
	// 			HelmDeployWaitSeconds:     400,
	// 			IngressHosts:              []string{"ingress.host1", "ingress.host2"},
	// 			Image:                     "path/to/Image:latest",
	// 			AdditionalParameters:      []string{"--testParam", "testValue"},
	// 			KubeContext:               "testCluster",
	// 			Namespace:                 "deploymentNamespace",
	// 			KeepFailedDeployments:     true,
	// 		}

	// 		k8sSecretSpec := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.StdoutReturn = map[string]string{
	// 			`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: k8sSecretSpec,
	// 		}

	// 		var stdout bytes.Buffer

	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong init command")
	// 		assert.Equal(t, []string{"init", "--client-only"}, mockUtils.Calls[0].Params, "Wrong init parameters")

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong secret creation command")
	// 		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, mockUtils.Calls[1].Params, "Wrong secret creation parameters")

	// 		assert.Equal(t, "helm", mockUtils.Calls[2].Exec, "Wrong upgrade command")
	// 		assert.Equal(t, []string{
	// 			"upgrade",
	// 			"deploymentName",
	// 			"path/to/chart",
	// 			"--install",
	// 			"--namespace",
	// 			"deploymentNamespace",
	// 			"--set",
	// 			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2",
	// 			"--force",
	// 			"--wait",
	// 			"--timeout",
	// 			"400",
	// 			"--kube-context",
	// 			"testCluster",
	// 			"--testParam",
	// 			"testValue",
	// 		}, mockUtils.Calls[2].Params, "Wrong upgrade parameters")
	// 	})

	// 	t.Run("test helm - fails without image information", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:    "https://my.registry:55555",
	// 			ContainerRegistrySecret: "testSecret",
	// 			ChartPath:               "path/to/chart",
	// 			DeploymentName:          "deploymentName",
	// 			DeployTool:              "helm",
	// 			ForceUpdates:            true,
	// 			HelmDeployWaitSeconds:   400,
	// 			IngressHosts:            []string{},
	// 			AdditionalParameters:    []string{"--testParam", "testValue"},
	// 			KubeContext:             "testCluster",
	// 			Namespace:               "deploymentNamespace",
	// 		}
	// 		mockUtils := newKubernetesDeployMockUtils()

	// 		var stdout bytes.Buffer

	// 		err := runKubernetesDeploy(opts, mockUtils, &stdout)
	// 		assert.EqualError(t, err, "image information not given - please either set image or containerImageName and containerImageTag")
	// 	})

	// 	t.Run("test helm v3", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:      "https://my.registry:55555",
	// 			ContainerRegistryUser:     "registryUser",
	// 			ContainerRegistryPassword: "********",
	// 			ContainerRegistrySecret:   "testSecret",
	// 			ChartPath:                 "path/to/chart",
	// 			DeploymentName:            "deploymentName",
	// 			DeployTool:                "helm3",
	// 			ForceUpdates:              true,
	// 			HelmDeployWaitSeconds:     400,
	// 			HelmValues:                []string{"values1.yaml", "values2.yaml"},
	// 			Image:                     "path/to/Image:latest",
	// 			AdditionalParameters:      []string{"--testParam", "testValue"},
	// 			KubeContext:               "testCluster",
	// 			Namespace:                 "deploymentNamespace",
	// 		}

	// 		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.StdoutReturn = map[string]string{
	// 			`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: dockerConfigJSON,
	// 		}

	// 		var stdout bytes.Buffer

	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
	// 		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, mockUtils.Calls[0].Params, "Wrong secret creation parameters")

	// 		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
	// 		assert.Equal(t, []string{
	// 			"upgrade",
	// 			"deploymentName",
	// 			"path/to/chart",
	// 			"--values",
	// 			"values1.yaml",
	// 			"--values",
	// 			"values2.yaml",
	// 			"--install",
	// 			"--namespace",
	// 			"deploymentNamespace",
	// 			"--set",
	// 			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
	// 			"--force",
	// 			"--wait",
	// 			"--timeout",
	// 			"400s",
	// 			"--atomic",
	// 			"--kube-context",
	// 			"testCluster",
	// 			"--testParam",
	// 			"testValue",
	// 		}, mockUtils.Calls[1].Params, "Wrong upgrade parameters")
	// 	})

	// 	t.Run("test helm v3 - with containerImageName and containerImageTag instead of image", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:      "https://my.registry:55555",
	// 			ContainerRegistryUser:     "registryUser",
	// 			ContainerRegistryPassword: "********",
	// 			ContainerRegistrySecret:   "testSecret",
	// 			ChartPath:                 "path/to/chart",
	// 			DeploymentName:            "deploymentName",
	// 			DeployTool:                "helm3",
	// 			ForceUpdates:              true,
	// 			HelmDeployWaitSeconds:     400,
	// 			HelmValues:                []string{"values1.yaml", "values2.yaml"},
	// 			ContainerImageName:        "path/to/Image",
	// 			ContainerImageTag:         "latest",
	// 			AdditionalParameters:      []string{"--testParam", "testValue"},
	// 			KubeContext:               "testCluster",
	// 			Namespace:                 "deploymentNamespace",
	// 		}

	// 		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.StdoutReturn = map[string]string{
	// 			`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: dockerConfigJSON,
	// 		}

	// 		var stdout bytes.Buffer

	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
	// 		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, mockUtils.Calls[0].Params, "Wrong secret creation parameters")

	// 		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")

	// 		assert.Contains(t, mockUtils.Calls[1].Params, "image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret", "Wrong upgrade parameters")

	// 	})

	// 	t.Run("test helm3 - fails without image information", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:    "https://my.registry:55555",
	// 			ContainerRegistrySecret: "testSecret",
	// 			ChartPath:               "path/to/chart",
	// 			DeploymentName:          "deploymentName",
	// 			DeployTool:              "helm3",
	// 			ForceUpdates:            true,
	// 			HelmDeployWaitSeconds:   400,
	// 			IngressHosts:            []string{},
	// 			AdditionalParameters:    []string{"--testParam", "testValue"},
	// 			KubeContext:             "testCluster",
	// 			Namespace:               "deploymentNamespace",
	// 		}
	// 		mockUtils := newKubernetesDeployMockUtils()

	// 		var stdout bytes.Buffer

	// 		err := runKubernetesDeploy(opts, mockUtils, &stdout)
	// 		assert.EqualError(t, err, "image information not given - please either set image or containerImageName and containerImageTag")
	// 	})

	// 	t.Run("test helm v3 - keep failed deployments", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:      "https://my.registry:55555",
	// 			ContainerRegistryUser:     "registryUser",
	// 			ContainerRegistryPassword: "********",
	// 			ContainerRegistrySecret:   "testSecret",
	// 			ChartPath:                 "path/to/chart",
	// 			DeploymentName:            "deploymentName",
	// 			DeployTool:                "helm3",
	// 			ForceUpdates:              true,
	// 			HelmDeployWaitSeconds:     400,
	// 			HelmValues:                []string{"values1.yaml", "values2.yaml"},
	// 			Image:                     "path/to/Image:latest",
	// 			AdditionalParameters:      []string{"--testParam", "testValue"},
	// 			KubeContext:               "testCluster",
	// 			Namespace:                 "deploymentNamespace",
	// 			KeepFailedDeployments:     true,
	// 		}

	// 		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.StdoutReturn = map[string]string{
	// 			`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: dockerConfigJSON,
	// 		}

	// 		var stdout bytes.Buffer

	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret creation command")
	// 		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, mockUtils.Calls[0].Params, "Wrong secret creation parameters")

	// 		assert.Equal(t, "helm", mockUtils.Calls[1].Exec, "Wrong upgrade command")
	// 		assert.Equal(t, []string{
	// 			"upgrade",
	// 			"deploymentName",
	// 			"path/to/chart",
	// 			"--values",
	// 			"values1.yaml",
	// 			"--values",
	// 			"values2.yaml",
	// 			"--install",
	// 			"--namespace",
	// 			"deploymentNamespace",
	// 			"--set",
	// 			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
	// 			"--force",
	// 			"--wait",
	// 			"--timeout",
	// 			"400s",
	// 			"--kube-context",
	// 			"testCluster",
	// 			"--testParam",
	// 			"testValue",
	// 		}, mockUtils.Calls[1].Params, "Wrong upgrade parameters")
	// 	})

	// 	t.Run("test helm v3 - no container credentials", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:    "https://my.registry:55555",
	// 			ChartPath:               "path/to/chart",
	// 			ContainerRegistrySecret: "testSecret",
	// 			DeploymentName:          "deploymentName",
	// 			DeployTool:              "helm3",
	// 			ForceUpdates:            true,
	// 			HelmDeployWaitSeconds:   400,
	// 			IngressHosts:            []string{},
	// 			Image:                   "path/to/Image:latest",
	// 			AdditionalParameters:    []string{"--testParam", "testValue"},
	// 			KubeContext:             "testCluster",
	// 			Namespace:               "deploymentNamespace",
	// 		}
	// 		mockUtils := newKubernetesDeployMockUtils()

	// 		var stdout bytes.Buffer

	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, 1, len(mockUtils.Calls), "Wrong number of upgrade commands")
	// 		assert.Equal(t, "helm", mockUtils.Calls[0].Exec, "Wrong upgrade command")
	// 		assert.Equal(t, []string{
	// 			"upgrade",
	// 			"deploymentName",
	// 			"path/to/chart",
	// 			"--install",
	// 			"--namespace",
	// 			"deploymentNamespace",
	// 			"--set",
	// 			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,imagePullSecrets[0].name=testSecret",
	// 			"--force",
	// 			"--wait",
	// 			"--timeout",
	// 			"400s",
	// 			"--atomic",
	// 			"--kube-context",
	// 			"testCluster",
	// 			"--testParam",
	// 			"testValue",
	// 		}, mockUtils.Calls[0].Params, "Wrong upgrade parameters")
	// 	})

	// 	t.Run("test helm v3 - fails without chart path", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:    "https://my.registry:55555",
	// 			ContainerRegistrySecret: "testSecret",
	// 			DeploymentName:          "deploymentName",
	// 			DeployTool:              "helm3",
	// 			ForceUpdates:            true,
	// 			HelmDeployWaitSeconds:   400,
	// 			IngressHosts:            []string{},
	// 			Image:                   "path/to/Image:latest",
	// 			AdditionalParameters:    []string{"--testParam", "testValue"},
	// 			KubeContext:             "testCluster",
	// 			Namespace:               "deploymentNamespace",
	// 		}
	// 		mockUtils := newKubernetesDeployMockUtils()

	// 		var stdout bytes.Buffer

	// 		err := runKubernetesDeploy(opts, mockUtils, &stdout)
	// 		assert.EqualError(t, err, "chart path has not been set, please configure chartPath parameter")
	// 	})

	// 	t.Run("test helm v3 - fails without deployment name", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:    "https://my.registry:55555",
	// 			ContainerRegistrySecret: "testSecret",
	// 			ChartPath:               "path/to/chart",
	// 			DeployTool:              "helm3",
	// 			ForceUpdates:            true,
	// 			HelmDeployWaitSeconds:   400,
	// 			IngressHosts:            []string{},
	// 			Image:                   "path/to/Image:latest",
	// 			AdditionalParameters:    []string{"--testParam", "testValue"},
	// 			KubeContext:             "testCluster",
	// 			Namespace:               "deploymentNamespace",
	// 		}
	// 		mockUtils := newKubernetesDeployMockUtils()

	// 		var stdout bytes.Buffer

	// 		err := runKubernetesDeploy(opts, mockUtils, &stdout)
	// 		assert.EqualError(t, err, "deployment name has not been set, please configure deploymentName parameter")
	// 	})

	// 	t.Run("test helm v3 - no force", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			ContainerRegistryURL:    "https://my.registry:55555",
	// 			ChartPath:               "path/to/chart",
	// 			ContainerRegistrySecret: "testSecret",
	// 			DeploymentName:          "deploymentName",
	// 			DeployTool:              "helm3",
	// 			HelmDeployWaitSeconds:   400,
	// 			IngressHosts:            []string{},
	// 			Image:                   "path/to/Image:latest",
	// 			AdditionalParameters:    []string{"--testParam", "testValue"},
	// 			KubeContext:             "testCluster",
	// 			Namespace:               "deploymentNamespace",
	// 		}
	// 		mockUtils := newKubernetesDeployMockUtils()

	// 		var stdout bytes.Buffer

	// 		runKubernetesDeploy(opts, mockUtils, &stdout)
	// 		assert.Equal(t, []string{
	// 			"upgrade",
	// 			"deploymentName",
	// 			"path/to/chart",
	// 			"--install",
	// 			"--namespace",
	// 			"deploymentNamespace",
	// 			"--set",
	// 			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,imagePullSecrets[0].name=testSecret",
	// 			"--wait",
	// 			"--timeout",
	// 			"400s",
	// 			"--atomic",
	// 			"--kube-context",
	// 			"testCluster",
	// 			"--testParam",
	// 			"testValue",
	// 		}, mockUtils.Calls[0].Params, "Wrong upgrade parameters")
	// 	})

	// 	t.Run("test kubectl - create secret/kubeconfig", func(t *testing.T) {

	// 		opts := kubernetesDeployOptions{
	// 			AppTemplate:                "path/to/test.yaml",
	// 			ContainerRegistryURL:       "https://my.registry:55555",
	// 			ContainerRegistryUser:      "registryUser",
	// 			ContainerRegistryPassword:  "********",
	// 			ContainerRegistrySecret:    "regSecret",
	// 			CreateDockerRegistrySecret: true,
	// 			DeployTool:                 "kubectl",
	// 			Image:                      "path/to/Image:latest",
	// 			AdditionalParameters:       []string{"--testParam", "testValue"},
	// 			KubeConfig:                 "This is my kubeconfig",
	// 			KubeContext:                "testCluster",
	// 			Namespace:                  "deploymentNamespace",
	// 			DeployCommand:              "apply",
	// 		}

	// 		kubeYaml := `kind: Deployment
	// metadata:
	// spec:
	//   spec:
	//     image: <image-name>`

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.AddFile(opts.AppTemplate, []byte(kubeYaml))

	// 		mockUtils.ShouldFailOnCommand = map[string]error{
	// 			"kubectl --insecure-skip-tls-verify=true --namespace=deploymentNamespace --context=testCluster get secret regSecret": fmt.Errorf("secret not found"),
	// 		}
	// 		var stdout bytes.Buffer
	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, mockUtils.Env, []string{"KUBECONFIG=This is my kubeconfig"})

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret lookup command")
	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			fmt.Sprintf("--context=%v", opts.KubeContext),
	// 			"get",
	// 			"secret",
	// 			opts.ContainerRegistrySecret,
	// 		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong secret create command")
	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			fmt.Sprintf("--context=%v", opts.KubeContext),
	// 			"create",
	// 			"secret",
	// 			"docker-registry",
	// 			opts.ContainerRegistrySecret,
	// 			"--docker-server=my.registry:55555",
	// 			fmt.Sprintf("--docker-username=%v", opts.ContainerRegistryUser),
	// 			fmt.Sprintf("--docker-password=%v", opts.ContainerRegistryPassword),
	// 		}, mockUtils.Calls[1].Params, "kubectl parameters incorrect")

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[2].Exec, "Wrong apply command")
	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			fmt.Sprintf("--context=%v", opts.KubeContext),
	// 			"apply",
	// 			"--filename",
	// 			opts.AppTemplate,
	// 			"--testParam",
	// 			"testValue",
	// 		}, mockUtils.Calls[2].Params, "kubectl parameters incorrect")

	// 		appTemplate, _ := mockUtils.FileRead(opts.AppTemplate)
	// 		assert.Contains(t, string(appTemplate), "my.registry:55555/path/to/Image:latest")
	// 	})

	// 	t.Run("test kubectl - create secret from docker config.json", func(t *testing.T) {

	// 		opts := kubernetesDeployOptions{
	// 			AppTemplate:                "path/to/test.yaml",
	// 			DockerConfigJSON:           "/path/to/.docker/config.json",
	// 			ContainerRegistryURL:       "https://my.registry:55555",
	// 			ContainerRegistryUser:      "registryUser",
	// 			ContainerRegistryPassword:  "********",
	// 			ContainerRegistrySecret:    "regSecret",
	// 			CreateDockerRegistrySecret: true,
	// 			DeployTool:                 "kubectl",
	// 			Image:                      "path/to/Image:latest",
	// 			AdditionalParameters:       []string{"--testParam", "testValue"},
	// 			KubeConfig:                 "This is my kubeconfig",
	// 			KubeContext:                "testCluster",
	// 			Namespace:                  "deploymentNamespace",
	// 			DeployCommand:              "apply",
	// 		}

	// 		kubeYaml := `kind: Deployment
	// metadata:
	// spec:
	//   spec:
	//     image: <image-name>`

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.AddFile(opts.AppTemplate, []byte(kubeYaml))
	// 		mockUtils.ShouldFailOnCommand = map[string]error{
	// 			"kubectl --insecure-skip-tls-verify=true --namespace=deploymentNamespace --context=testCluster get secret regSecret": fmt.Errorf("secret not found"),
	// 		}
	// 		var stdout bytes.Buffer
	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, mockUtils.Env, []string{"KUBECONFIG=This is my kubeconfig"})

	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			fmt.Sprintf("--context=%v", opts.KubeContext),
	// 			"create",
	// 			"secret",
	// 			"generic",
	// 			opts.ContainerRegistrySecret,
	// 			fmt.Sprintf("--from-file=.dockerconfigjson=%v", opts.DockerConfigJSON),
	// 			`--type=kubernetes.io/dockerconfigjson`,
	// 		}, mockUtils.Calls[1].Params, "kubectl parameters incorrect")
	// 	})

	// 	t.Run("test kubectl - lookup secret/kubeconfig", func(t *testing.T) {

	// 		opts := kubernetesDeployOptions{
	// 			AppTemplate:                "path/to/test.yaml",
	// 			ContainerRegistryURL:       "https://my.registry:55555",
	// 			ContainerRegistryUser:      "registryUser",
	// 			ContainerRegistryPassword:  "********",
	// 			ContainerRegistrySecret:    "regSecret",
	// 			CreateDockerRegistrySecret: true,
	// 			DeployTool:                 "kubectl",
	// 			Image:                      "path/to/Image:latest",
	// 			KubeConfig:                 "This is my kubeconfig",
	// 			Namespace:                  "deploymentNamespace",
	// 			DeployCommand:              "apply",
	// 		}

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.AddFile(opts.AppTemplate, []byte("testYaml"))

	// 		var stdout bytes.Buffer
	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong secret lookup command")
	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			"get",
	// 			"secret",
	// 			opts.ContainerRegistrySecret,
	// 		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong apply command")
	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			"apply",
	// 			"--filename",
	// 			opts.AppTemplate,
	// 		}, mockUtils.Calls[1].Params, "kubectl parameters incorrect")
	// 	})

	// 	t.Run("test kubectl - token only", func(t *testing.T) {

	// 		opts := kubernetesDeployOptions{
	// 			APIServer:                 "https://my.api.server",
	// 			AppTemplate:               "path/to/test.yaml",
	// 			ContainerRegistryURL:      "https://my.registry:55555",
	// 			ContainerRegistryUser:     "registryUser",
	// 			ContainerRegistryPassword: "********",
	// 			ContainerRegistrySecret:   "regSecret",
	// 			DeployTool:                "kubectl",
	// 			Image:                     "path/to/Image:latest",
	// 			KubeToken:                 "testToken",
	// 			Namespace:                 "deploymentNamespace",
	// 			DeployCommand:             "apply",
	// 		}

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.AddFile(opts.AppTemplate, []byte("testYaml"))
	// 		mockUtils.ShouldFailOnCommand = map[string]error{}

	// 		var stdout bytes.Buffer
	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")
	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			fmt.Sprintf("--server=%v", opts.APIServer),
	// 			fmt.Sprintf("--token=%v", opts.KubeToken),
	// 			"apply",
	// 			"--filename",
	// 			opts.AppTemplate,
	// 		}, mockUtils.Calls[0].Params, "kubectl parameters incorrect")
	// 	})

	// 	t.Run("test kubectl - with containerImageName and containerImageTag instead of image", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			APIServer:                 "https://my.api.server",
	// 			AppTemplate:               "test.yaml",
	// 			ContainerRegistryURL:      "https://my.registry:55555",
	// 			ContainerRegistryUser:     "registryUser",
	// 			ContainerRegistryPassword: "********",
	// 			ContainerRegistrySecret:   "regSecret",
	// 			DeployTool:                "kubectl",
	// 			ContainerImageTag:         "latest",
	// 			ContainerImageName:        "path/to/Image",
	// 			KubeConfig:                "This is my kubeconfig",
	// 			Namespace:                 "deploymentNamespace",
	// 			DeployCommand:             "apply",
	// 		}

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.AddFile("test.yaml", []byte("image: <image-name>"))

	// 		var stdout bytes.Buffer
	// 		runKubernetesDeploy(opts, mockUtils, &stdout)

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[0].Exec, "Wrong apply command")

	// 		appTemplateFileContents, err := mockUtils.FileRead(opts.AppTemplate)
	// 		assert.NoError(t, err)
	// 		assert.Contains(t, string(appTemplateFileContents), "image: my.registry:55555/path/to/Image:latest", "kubectl parameters incorrect")
	// 	})

	// 	t.Run("test kubectl - fails without image information", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			APIServer:                 "https://my.api.server",
	// 			AppTemplate:               "test.yaml",
	// 			ContainerRegistryURL:      "https://my.registry:55555",
	// 			ContainerRegistryUser:     "registryUser",
	// 			ContainerRegistryPassword: "********",
	// 			ContainerRegistrySecret:   "regSecret",
	// 			DeployTool:                "kubectl",
	// 			KubeConfig:                "This is my kubeconfig",
	// 			Namespace:                 "deploymentNamespace",
	// 			DeployCommand:             "apply",
	// 		}

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.AddFile("test.yaml", []byte("testYaml"))

	// 		var stdout bytes.Buffer

	// 		err := runKubernetesDeploy(opts, mockUtils, &stdout)
	// 		assert.EqualError(t, err, "image information not given - please either set image or containerImageName and containerImageTag")
	// 	})

	// 	t.Run("test kubectl - use replace deploy command", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			AppTemplate:                "test.yaml",
	// 			ContainerRegistryURL:       "https://my.registry:55555",
	// 			ContainerRegistryUser:      "registryUser",
	// 			ContainerRegistryPassword:  "********",
	// 			ContainerRegistrySecret:    "regSecret",
	// 			CreateDockerRegistrySecret: true,
	// 			DeployTool:                 "kubectl",
	// 			Image:                      "path/to/Image:latest",
	// 			AdditionalParameters:       []string{"--testParam", "testValue"},
	// 			KubeConfig:                 "This is my kubeconfig",
	// 			KubeContext:                "testCluster",
	// 			Namespace:                  "deploymentNamespace",
	// 			DeployCommand:              "replace",
	// 		}

	// 		kubeYaml := `kind: Deployment
	// metadata:
	// spec:
	//   spec:
	//     image: <image-name>`

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.AddFile("test.yaml", []byte(kubeYaml))

	// 		var stdout bytes.Buffer
	// 		err := runKubernetesDeploy(opts, mockUtils, &stdout)
	// 		assert.NoError(t, err, "Command should not fail")

	// 		assert.Equal(t, mockUtils.Env, []string{"KUBECONFIG=This is my kubeconfig"})

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong replace command")
	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			fmt.Sprintf("--context=%v", opts.KubeContext),
	// 			"replace",
	// 			"--filename",
	// 			opts.AppTemplate,
	// 			"--testParam",
	// 			"testValue",
	// 		}, mockUtils.Calls[1].Params, "kubectl parameters incorrect")

	// 		appTemplate, err := mockUtils.FileRead(opts.AppTemplate)
	// 		assert.Contains(t, string(appTemplate), "my.registry:55555/path/to/Image:latest")
	// 	})

	// 	t.Run("test kubectl - use replace --force deploy command", func(t *testing.T) {
	// 		opts := kubernetesDeployOptions{
	// 			AppTemplate:                "test.yaml",
	// 			ContainerRegistryURL:       "https://my.registry:55555",
	// 			ContainerRegistryUser:      "registryUser",
	// 			ContainerRegistryPassword:  "********",
	// 			ContainerRegistrySecret:    "regSecret",
	// 			CreateDockerRegistrySecret: true,
	// 			DeployTool:                 "kubectl",
	// 			Image:                      "path/to/Image:latest",
	// 			AdditionalParameters:       []string{"--testParam", "testValue"},
	// 			KubeConfig:                 "This is my kubeconfig",
	// 			KubeContext:                "testCluster",
	// 			Namespace:                  "deploymentNamespace",
	// 			DeployCommand:              "replace",
	// 			ForceUpdates:               true,
	// 		}

	// 		kubeYaml := `kind: Deployment
	// metadata:
	// spec:
	//   spec:
	//     image: <image-name>`

	// 		mockUtils := newKubernetesDeployMockUtils()
	// 		mockUtils.AddFile("test.yaml", []byte(kubeYaml))

	// 		var stdout bytes.Buffer
	// 		err := runKubernetesDeploy(opts, mockUtils, &stdout)
	// 		assert.NoError(t, err, "Command should not fail")

	// 		assert.Equal(t, mockUtils.Env, []string{"KUBECONFIG=This is my kubeconfig"})

	// 		assert.Equal(t, "kubectl", mockUtils.Calls[1].Exec, "Wrong replace command")
	// 		assert.Equal(t, []string{
	// 			"--insecure-skip-tls-verify=true",
	// 			fmt.Sprintf("--namespace=%v", opts.Namespace),
	// 			fmt.Sprintf("--context=%v", opts.KubeContext),
	// 			"replace",
	// 			"--filename",
	// 			opts.AppTemplate,
	// 			"--force",
	// 			"--testParam",
	// 			"testValue",
	// 		}, mockUtils.Calls[1].Params, "kubectl parameters incorrect")

	// 		appTemplate, err := mockUtils.FileRead(opts.AppTemplate)
	// 		assert.Contains(t, string(appTemplate), "my.registry:55555/path/to/Image:latest")
	// 	})

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
