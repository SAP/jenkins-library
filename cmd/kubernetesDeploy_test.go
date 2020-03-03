package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunKubernetesDeploy(t *testing.T) {

	t.Run("test helm", func(t *testing.T) {
		opts := kubernetesDeployOptions{
			ContainerRegistryURL:      "https://my.registry:55555",
			ContainerRegistryUser:     "registryUser",
			ContainerRegistryPassword: "********",
			ChartPath:                 "path/to/chart",
			DeploymentName:            "deploymentName",
			DeployTool:                "helm",
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
				"kubectl --insecure-skip-tls-verify=true create secret docker-registry regsecret --docker-server=my.registry:55555 --docker-username=registryUser --docker-password=******** --dry-run=true --output=json": dockerConfigJSON,
			},
		}

		var stdout bytes.Buffer

		runKubernetesDeploy(opts, &e, &stdout)

		assert.Equal(t, "helm", e.Calls[0].Exec, "Wrong init command")
		assert.Equal(t, []string{"init", "--client-only"}, e.Calls[0].Params, "Wrong init parameters")

		assert.Equal(t, "kubectl", e.Calls[1].Exec, "Wrong secret creation command")
		assert.Equal(t, []string{"--insecure-skip-tls-verify=true", "create", "secret", "docker-registry", "regsecret", "--docker-server=my.registry:55555", "--docker-username=registryUser", "--docker-password=********", "--dry-run=true", "--output=json"}, e.Calls[1].Params, "Wrong secret creation parameters")

		assert.Equal(t, "helm", e.Calls[2].Exec, "Wrong upgrade command")
		assert.Equal(t, []string{
			"upgrade",
			"deploymentName",
			"path/to/chart",
			"--install",
			"--force",
			"--namespace",
			"deploymentNamespace",
			"--wait",
			"--timeout",
			"400",
			"--set",
			"image.repository=my.registry:55555/path/to/Image,image.tag=latest,secret.dockerconfigjson=ThisIsOurBase64EncodedSecret==,ingress.hosts[0]=ingress.host1,ingress.hosts[1]=ingress.host2",
			"--kube-context",
			"testCluster",
			"--testParam",
			"testValue",
		}, e.Calls[2].Params, "Wrong upgrade parameters")
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
