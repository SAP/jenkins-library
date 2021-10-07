package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestRunKubernetesDeploy(t *testing.T) {

	t.Run("test helm", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
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
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		e := mock.ExecMockRunner{
			StdoutReturn: map[string]string{
				`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: dockerConfigJSON,
			},
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "helm", e.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, e.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", e.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, e.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", e.Calls[2].Exec, "Wrong upgrade command")
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
		}, e.Calls[2].Params, "Wrong upgrade parameters")
	})

	t.Run("test helm - with containerImageName and containerImageTag instead of image", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
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
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		e := mock.ExecMockRunner{
			StdoutReturn: map[string]string{
				`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: dockerConfigJSON,
			},
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "helm", e.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, e.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", e.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, e.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", e.Calls[2].Exec, "Wrong upgrade command")
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
		}, e.Calls[2].Params, "Wrong upgrade parameters")
	})

	t.Run("test helm - docker config.json path passed as parameter", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:    "https://my.registry:55555",
			DockerConfigJSON:        "/path/to/.docker/config.json",
			ContainerRegistrySecret: "testSecret",
			ChartPath:               "path/to/chart",
			DeploymentName:          "deploymentName",
			DeployTool:              "helm",
			ForceUpdates:            true,
			HelmDeployWaitSeconds:   400,
			IngressHosts:            []string{"ingress.host1", "ingress.host2"},
			Image:                   "path/to/Image:latest",
			AdditionalParameters:    []string{"--testParam", "testValue"},
			KubeContext:             "testCluster",
			Namespace:               "deploymentNamespace",
		}

		k8sSecretSpec := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`
		e := mock.ExecMockRunner{
			StdoutReturn: map[string]string{
				`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json generic testSecret --from-file=.dockerconfigjson=/path/to/.docker/config.json --type=kubernetes.io/dockerconfigjson`: k8sSecretSpec,
			},
		}

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &e, &stdout)
		assert.NoError(t, err)

		assert.Equal(t, "helm", e.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, e.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", e.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{
			"create",
			"secret",
			"--insecure-skip-tls-verify=true",
			"--dry-run=true",
			"--output=json",
			"generic",
			"testSecret",
			"--from-file=.dockerconfigjson=/path/to/.docker/config.json",
			`--type=kubernetes.io/dockerconfigjson`,
		}, e.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", e.Calls[2].Exec, "Wrong upgrade command")
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
		}, e.Calls[2].Params, "Wrong upgrade parameters")
	})

	t.Run("test helm -- keep failed deployment", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
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
		}

		k8sSecretSpec := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		e := mock.ExecMockRunner{
			StdoutReturn: map[string]string{
				`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: k8sSecretSpec,
			},
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "helm", e.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, e.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", e.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, e.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", e.Calls[2].Exec, "Wrong upgrade command")
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
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, e.Calls[2].Params, "Wrong upgrade parameters")
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
		e := mock.ExecMockRunner{}

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &e, &stdout)
		assert.EqualError(t, err, "Image information not given. Please either set containerImageName, containerImageTag, and containerRegistryURL, or set image and containerRegistryURL.")
	})

	t.Run("test helm v3", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
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
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		e := mock.ExecMockRunner{
			StdoutReturn: map[string]string{
				`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: dockerConfigJSON,
			},
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "kubectl", e.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, e.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", e.Calls[1].Exec, "Wrong upgrade command")
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
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, e.Calls[1].Params, "Wrong upgrade parameters")
	})

	t.Run("test helm v3 - with containerImageName and containerImageTag instead of image", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
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
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		e := mock.ExecMockRunner{
			StdoutReturn: map[string]string{
				`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: dockerConfigJSON,
			},
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "kubectl", e.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, e.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", e.Calls[1].Exec, "Wrong upgrade command")
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
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, e.Calls[1].Params, "Wrong upgrade parameters")
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
		e := mock.ExecMockRunner{}

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &e, &stdout)
		assert.EqualError(t, err, "Image information not given. Please either set containerImageName, containerImageTag, and containerRegistryURL, or set image and containerRegistryURL.")
	})

	t.Run("test helm v3 - keep failed deployments", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
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
		}

		dockerConfigJSON := `{"kind": "Secret","data":{".dockerconfigjson": "ThisIsOurBase64EncodedSecret=="}}`

		e := mock.ExecMockRunner{
			StdoutReturn: map[string]string{
				`kubectl create secret --insecure-skip-tls-verify=true --dry-run=true --output=json docker-registry testSecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=\*\*\*\*\*\*\*\*`: dockerConfigJSON,
			},
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "kubectl", e.Calls[0].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"create", "secret", "--insecure-skip-tls-verify=true", "--dry-run=true", "--output=json", "docker-registry", "testSecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********"}, e.Calls[0].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", e.Calls[1].Exec, "Wrong upgrade command")
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
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.name=testSecret,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, e.Calls[1].Params, "Wrong upgrade parameters")
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
		e := mock.ExecMockRunner{}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, 1, len(e.Calls), "Wrong number of upgrade commands")
		assert.Equal(t, "helm", e.Calls[0].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,imagePullSecrets[0].name=testSecret",
			"--force",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, e.Calls[0].Params, "Wrong upgrade parameters")
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
		e := mock.ExecMockRunner{}

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &e, &stdout)
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
		e := mock.ExecMockRunner{}

		var stdout bytes.Buffer

		err := runKubernetesDeploy(opts, &e, &stdout)
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
		e := mock.ExecMockRunner{}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--namespace",
			"deploymentNamespace",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,imagePullSecrets[0].name=testSecret",
			"--wait",
			"--timeout",
			"400s",
			"--atomic",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, e.Calls[0].Params, "Wrong upgrade parameters")
	})

	t.Run("test kubectl - create secret/kubeconfig", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		defer os.RemoveAll(dir) // clean up
		assert.NoError(t, err, "Error when creating temp dir")

		opts := kubernetesDeployOptions{
			AppTemplate:                filepath.Join(dir, "test.yaml"),
			ContainerRegistryURL:       "https://my.registry:55555",
			ContainerRegistryUser:      "registryUser",
			ContainerRegistryPassword:  "********",
			ContainerRegistrySecret:    "regSecret",
			CreateDockerRegistrySecret: true,
			DeployTool:                 "kubectl",
			Image:                      "path/to/Image:latest",
			AdditionalParameters:       []string{"--testParam", "testValue"},
			KubeConfig:                 "This is my kubeconfig",
			KubeContext:                "testCluster",
			Namespace:                  "deploymentNamespace",
		}

		kubeYaml := `kind: Deployment
metadata:
spec:
  spec:
    image: <image-name>`

		ioutil.WriteFile(opts.AppTemplate, []byte(kubeYaml), 0755)

		e := mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{
				"kubectl --insecure-skip-tls-verify=true --namespace=deploymentNamespace --context=testCluster get secret regSecret": fmt.Errorf("secret not found"),
			},
		}
		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, e.Env, []string{"KUBECONFIG=This is my kubeconfig"})

		assert.Equal(t, "kubectl", e.Calls[0].Exec, "Wrong secret lookup command")
		assert.Equal(t, []string{
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			fmt.Sprintf("--context=%v", opts.KubeContext),
			"get",
			"secret",
			opts.ContainerRegistrySecret,
		}, e.Calls[0].Params, "kubectl parameters incorrect")

		assert.Equal(t, "kubectl", e.Calls[1].Exec, "Wrong secret create command")
		assert.Equal(t, []string{
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			fmt.Sprintf("--context=%v", opts.KubeContext),
			"create",
			"secret",
			"docker-registry",
			opts.ContainerRegistrySecret,
			"--docker-server=my.registry:55555",
			fmt.Sprintf("--docker-username=%v", opts.ContainerRegistryUser),
			fmt.Sprintf("--docker-password=%v", opts.ContainerRegistryPassword),
		}, e.Calls[1].Params, "kubectl parameters incorrect")

		assert.Equal(t, "kubectl", e.Calls[2].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			fmt.Sprintf("--context=%v", opts.KubeContext),
			"apply",
			"--filename",
			opts.AppTemplate,
			"--testParam",
			"testValue",
		}, e.Calls[2].Params, "kubectl parameters incorrect")

		appTemplate, err := ioutil.ReadFile(opts.AppTemplate)
		assert.Contains(t, string(appTemplate), "my.registry:55555/path/to/Image:latest")
	})

	t.Run("test kubectl - create secret from docker config.json", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		defer os.RemoveAll(dir) // clean up
		assert.NoError(t, err, "Error when creating temp dir")

		opts := kubernetesDeployOptions{
			AppTemplate:                filepath.Join(dir, "test.yaml"),
			DockerConfigJSON:           "/path/to/.docker/config.json",
			ContainerRegistryURL:       "https://my.registry:55555",
			ContainerRegistryUser:      "registryUser",
			ContainerRegistryPassword:  "********",
			ContainerRegistrySecret:    "regSecret",
			CreateDockerRegistrySecret: true,
			DeployTool:                 "kubectl",
			Image:                      "path/to/Image:latest",
			AdditionalParameters:       []string{"--testParam", "testValue"},
			KubeConfig:                 "This is my kubeconfig",
			KubeContext:                "testCluster",
			Namespace:                  "deploymentNamespace",
		}

		kubeYaml := `kind: Deployment
metadata:
spec:
  spec:
    image: <image-name>`

		ioutil.WriteFile(opts.AppTemplate, []byte(kubeYaml), 0755)

		e := mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{
				"kubectl --insecure-skip-tls-verify=true --namespace=deploymentNamespace --context=testCluster get secret regSecret": fmt.Errorf("secret not found"),
			},
		}
		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, e.Env, []string{"KUBECONFIG=This is my kubeconfig"})

		assert.Equal(t, []string{
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			fmt.Sprintf("--context=%v", opts.KubeContext),
			"create",
			"secret",
			"generic",
			opts.ContainerRegistrySecret,
			fmt.Sprintf("--from-file=.dockerconfigjson=%v", opts.DockerConfigJSON),
			`--type=kubernetes.io/dockerconfigjson`,
		}, e.Calls[1].Params, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - lookup secret/kubeconfig", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		defer os.RemoveAll(dir) // clean up
		assert.NoError(t, err, "Error when creating temp dir")

		opts := kubernetesDeployOptions{
			AppTemplate:                filepath.Join(dir, "test.yaml"),
			ContainerRegistryURL:       "https://my.registry:55555",
			ContainerRegistryUser:      "registryUser",
			ContainerRegistryPassword:  "********",
			ContainerRegistrySecret:    "regSecret",
			CreateDockerRegistrySecret: true,
			DeployTool:                 "kubectl",
			Image:                      "path/to/Image:latest",
			KubeConfig:                 "This is my kubeconfig",
			Namespace:                  "deploymentNamespace",
		}

		ioutil.WriteFile(opts.AppTemplate, []byte("testYaml"), 0755)

		e := mock.ExecMockRunner{}

		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "kubectl", e.Calls[0].Exec, "Wrong secret lookup command")
		assert.Equal(t, []string{
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"get",
			"secret",
			opts.ContainerRegistrySecret,
		}, e.Calls[0].Params, "kubectl parameters incorrect")

		assert.Equal(t, "kubectl", e.Calls[1].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"apply",
			"--filename",
			opts.AppTemplate,
		}, e.Calls[1].Params, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - token only", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		defer os.RemoveAll(dir) // clean up
		assert.NoError(t, err, "Error when creating temp dir")

		opts := kubernetesDeployOptions{
			APIServer:                 "https://my.api.server",
			AppTemplate:               filepath.Join(dir, "test.yaml"),
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
			ContainerRegistrySecret:   "regSecret",
			DeployTool:                "kubectl",
			Image:                     "path/to/Image:latest",
			KubeToken:                 "testToken",
			Namespace:                 "deploymentNamespace",
		}

		ioutil.WriteFile(opts.AppTemplate, []byte("testYaml"), 0755)

		e := mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{},
		}
		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "kubectl", e.Calls[0].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			fmt.Sprintf("--server=%v", opts.APIServer),
			fmt.Sprintf("--token=%v", opts.KubeToken),
			"apply",
			"--filename",
			opts.AppTemplate,
		}, e.Calls[0].Params, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - with containerImageName and containerImageTag instead of image", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		defer os.RemoveAll(dir) // clean up
		assert.NoError(t, err, "Error when creating temp dir")

		opts := kubernetesDeployOptions{
			APIServer:                 "https://my.api.server",
			AppTemplate:               filepath.Join(dir, "test.yaml"),
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
			ContainerRegistrySecret:   "regSecret",
			DeployTool:                "kubectl",
			ContainerImageTag:         "latest",
			ContainerImageName:        "path/to/Image",
			KubeConfig:                "This is my kubeconfig",
			Namespace:                 "deploymentNamespace",
		}

		ioutil.WriteFile(opts.AppTemplate, []byte("testYaml"), 0755)

		e := mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{},
		}
		var stdout bytes.Buffer
		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "kubectl", e.Calls[0].Exec, "Wrong apply command")
		assert.Equal(t, []string{
			"--insecure-skip-tls-verify=true",
			fmt.Sprintf("--namespace=%v", opts.Namespace),
			"apply",
			"--filename",
			opts.AppTemplate,
		}, e.Calls[0].Params, "kubectl parameters incorrect")
	})

	t.Run("test kubectl - fails without image information", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		defer os.RemoveAll(dir) // clean up
		assert.NoError(t, err, "Error when creating temp dir")

		opts := kubernetesDeployOptions{
			APIServer:                 "https://my.api.server",
			AppTemplate:               filepath.Join(dir, "test.yaml"),
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
			ContainerRegistrySecret:   "regSecret",
			DeployTool:                "kubectl",
			KubeConfig:                "This is my kubeconfig",
			Namespace:                 "deploymentNamespace",
		}

		ioutil.WriteFile(opts.AppTemplate, []byte("testYaml"), 0755)
		e := mock.ExecMockRunner{}

		var stdout bytes.Buffer

		err = runKubernetesDeploy(opts, &e, &stdout)
		assert.EqualError(t, err, "Image information not given. Please either set containerImageName, containerImageTag, and containerRegistryURL, or set image and containerRegistryURL.")
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
