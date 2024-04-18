//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	piperconf "github.com/SAP/jenkins-library/pkg/config"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const imageRegistry = "some-registry"

func newCnbBuildTestsUtils() cnbutils.MockUtils {
	imageStub := func(imageRef, target string) (v1.Image, error) {
		fakeImage := &fake.FakeImage{}
		var imageConfig v1.Config
		switch imageRef {
		case "pre-test":
			imageConfig = v1.Config{
				Labels: map[string]string{
					"io.buildpacks.buildpackage.metadata": "{\"id\": \"pre-testbuildpack\", \"version\": \"0.0.1\"}",
				},
			}
		case "post-test":
			imageConfig = v1.Config{
				Labels: map[string]string{
					"io.buildpacks.buildpackage.metadata": "{\"id\": \"post-testbuildpack\", \"version\": \"0.0.1\"}",
				},
			}
		default:
			imageConfig = v1.Config{
				Labels: map[string]string{
					"io.buildpacks.buildpackage.metadata": "{\"id\": \"testbuildpack\", \"version\": \"0.0.1\"}",
				},
			}
		}

		fakeImage.ConfigFileReturns(&v1.ConfigFile{
			Config: imageConfig,
		}, nil)

		return fakeImage, nil
	}

	utils := cnbutils.MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		DownloadMock: &mock.DownloadMock{
			ImageContentStub: imageStub,
			ImageInfoStub: func(imageRef string) (v1.Image, error) {
				return imageStub(imageRef, "")
			},
		},
	}

	utils.AddFile("/cnb/order.toml", []byte(`[[order]]
    [[order.group]]
      id = "buildpacks/java"
      version = "1.8.0"
[[order]]
    [[order.group]]
      id = "buildpacks/nodejs"
      version = "1.6.0"`))
	utils.AddFile("/layers/report.toml", []byte(`[build]
[image]
tags = ["localhost:5000/not-found:0.0.1"]
digest = "sha256:52eac630560210e5ae13eb10797c4246d6f02d425f32b9430ca00bde697c79ec"
manifest-size = 2388`))
	return utils
}

func addBuilderFiles(utils *cnbutils.MockUtils) {
	utils.FilesMock.AddFile(creatorPath, []byte(`xyz`))
}

func assertLifecycleCalls(t *testing.T, runner *mock.ExecMockRunner, callNo int) {
	require.GreaterOrEqual(t, len(runner.Calls), callNo)
	assert.Equal(t, creatorPath, runner.Calls[callNo-1].Exec)
	for _, arg := range []string{"-no-color", "-buildpacks", "/cnb/buildpacks", "-order", "/cnb/order.toml", "-platform", "/tmp/platform"} {
		assert.Contains(t, runner.Calls[callNo-1].Params, arg)
	}
}

func assetBuildEnv(t *testing.T, utils cnbutils.MockUtils, key, value string) bool {
	env, err := utils.FilesMock.ReadFile(filepath.Join("/tmp/platform/env/", key))
	if !assert.NoError(t, err) {
		return false
	}
	return assert.Equal(t, value, string(env))
}

func TestRunCnbBuild(t *testing.T) {
	configOptions.OpenFile = piperconf.OpenPiperFile

	t.Setenv("CNB_USER_ID", "1000")
	t.Setenv("CNB_GROUP_ID", "1000")

	t.Run("prefers direct configuration", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			DockerConfigJSON:     "/path/to/config.json",
			RunImage:             "my-run-image",
			DefaultProcess:       "my-process",
		}

		projectToml := `[project]
		id = "io.buildpacks.my-app"
		`

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddFile("project.toml", []byte(projectToml))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner, 2)
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", imageRegistry, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[1].Params, "-run-image")
		assert.Contains(t, runner.Calls[1].Params, "my-run-image")
		assert.Contains(t, runner.Calls[1].Params, "-process-type")
		assert.Contains(t, runner.Calls[1].Params, "my-process")
		assert.Equal(t, config.ContainerRegistryURL, commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:0.0.1", commonPipelineEnvironment.container.imageNameTag)
		assert.Equal(t, `{"cnbBuild":[{"dockerImage":"paketobuildpacks/builder-jammy-base:latest"}]}`, commonPipelineEnvironment.custom.buildSettingsInfo)
	})

	t.Run("prefers project descriptor", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			DockerConfigJSON:     "/path/to/config.json",
			ProjectDescriptor:    "project.toml",
		}

		projectToml := `[project]
		id = "io.buildpacks.my-app"
		`

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddFile("project.toml", []byte(projectToml))
		addBuilderFiles(&utils)

		telemetryData := &telemetry.CustomData{}
		err := callCnbBuild(&config, telemetryData, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner, 2)
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", imageRegistry, "io-buildpacks-my-app", config.ContainerImageTag))
		assert.Equal(t, config.ContainerRegistryURL, commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "io-buildpacks-my-app:0.0.1", commonPipelineEnvironment.container.imageNameTag)

		assert.Equal(t, "sha256:52eac630560210e5ae13eb10797c4246d6f02d425f32b9430ca00bde697c79ec", commonPipelineEnvironment.container.imageDigest)
		assert.Contains(t, commonPipelineEnvironment.container.imageDigests, "sha256:52eac630560210e5ae13eb10797c4246d6f02d425f32b9430ca00bde697c79ec")
	})

	t.Run("success case (registry with https)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner, 2)
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", imageRegistry, config.ContainerImageName, config.ContainerImageTag))
		assert.Equal(t, config.ContainerRegistryURL, commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:0.0.1", commonPipelineEnvironment.container.imageNameTag)
	})

	t.Run("success case (registry without https)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assertLifecycleCalls(t, runner, 2)
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
		assert.Equal(t, fmt.Sprintf("https://%s", config.ContainerRegistryURL), commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:0.0.1", commonPipelineEnvironment.container.imageNameTag)
	})

	t.Run("success case (custom buildpacks and custom env variables with expand, renaming docker conf file, additional tag)", func(t *testing.T) {
		t.Setenv("BAR", "BAZZ")
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/test.json",
			Buildpacks:           []string{"test"},
			ExpandBuildEnvVars:   true,
			BuildEnvVars: map[string]interface{}{
				"FOO": "${BAR}",
			},
			AdditionalTags: []string{"latest"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, creatorPath, runner.Calls[1].Exec)
		assert.Contains(t, runner.Calls[1].Params, "/tmp/buildpacks")
		assert.Contains(t, runner.Calls[1].Params, "/tmp/buildpacks/order.toml")
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:latest", config.ContainerRegistryURL, config.ContainerImageName))

		copiedFileExists, _ := utils.FileExists("/tmp/config.json")
		assert.True(t, copiedFileExists)

		assetBuildEnv(t, utils, "FOO", "BAZZ")
	})

	t.Run("success case (custom buildpacks, pre and post buildpacks and custom env variables, renaming docker conf file, additional tag)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/test.json",
			PreBuildpacks:        []string{"pre-test"},
			PostBuildpacks:       []string{"post-test"},
			Buildpacks:           []string{"test"},
			ExpandBuildEnvVars:   false,
			BuildEnvVars: map[string]interface{}{
				"FOO": "${BAR}",
			},
			AdditionalTags: []string{"latest"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, creatorPath, runner.Calls[1].Exec)
		assert.Contains(t, runner.Calls[1].Params, "/tmp/buildpacks")
		assert.Contains(t, runner.Calls[1].Params, "/tmp/buildpacks/order.toml")
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:latest", config.ContainerRegistryURL, config.ContainerImageName))

		copiedFileExists, _ := utils.FileExists("/tmp/config.json")
		assert.True(t, copiedFileExists)

		assetBuildEnv(t, utils, "FOO", "${BAR}")
	})

	t.Run("success case (custom pre and post buildpacks and custom env variables, renaming docker conf file, additional tag)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/test.json",
			PostBuildpacks:       []string{"post-test"},
			PreBuildpacks:        []string{"pre-test"},
			BuildEnvVars: map[string]interface{}{
				"FOO": "BAR",
			},
			AdditionalTags: []string{"latest"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Equal(t, creatorPath, runner.Calls[1].Exec)
		assert.Contains(t, runner.Calls[1].Params, "/tmp/buildpacks")
		assert.Contains(t, runner.Calls[1].Params, "/tmp/buildpacks/order.toml")
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:latest", config.ContainerRegistryURL, config.ContainerImageName))

		copiedFileExists, _ := utils.FileExists("/tmp/config.json")
		assert.True(t, copiedFileExists)
	})

	t.Run("success case (customTlsCertificates)", func(t *testing.T) {
		t.Parallel()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://test-cert.com/cert.crt", httpmock.NewStringResponder(200, "testCert"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		caCertsFile := "/etc/ssl/certs/ca-certificates.crt"
		caCertsTmpFile := "/tmp/ca-certificates.crt"
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:        "my-image",
			ContainerImageTag:         "0.0.1",
			ContainerRegistryURL:      registry,
			DockerConfigJSON:          "/path/to/config.json",
			CustomTLSCertificateLinks: []string{"https://test-cert.com/cert.crt", "https://test-cert.com/cert.crt"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(caCertsFile, []byte("test\n"))
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, client)
		require.NoError(t, err)

		result, err := utils.FilesMock.FileRead(caCertsTmpFile)
		require.NoError(t, err)
		assert.Equal(t, "test\ntestCert\ntestCert\n", string(result))

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Contains(t, runner.Env, fmt.Sprintf("SSL_CERT_FILE=%s", caCertsTmpFile))
		assertLifecycleCalls(t, runner, 2)
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
	})

	t.Run("success case (additionalTags)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/config.json",
			AdditionalTags:       []string{"3", "3.1", "3.1", "3.1.5"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		require.NoError(t, err)

		runner := utils.ExecMockRunner
		assertLifecycleCalls(t, runner, 2)
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, config.ContainerImageTag))
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:3", config.ContainerRegistryURL, config.ContainerImageName))
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:3.1", config.ContainerRegistryURL, config.ContainerImageName))
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:3.1.5", config.ContainerRegistryURL, config.ContainerImageName))
	})

	t.Run("success case: build environment variables", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			ProjectDescriptor:    "project.toml",
			BuildEnvVars: map[string]interface{}{
				"OPTIONS_KEY": "OPTIONS_VALUE",
				"OVERWRITE":   "this should win",
			},
		}

		projectToml := `[project]
		id = "io.buildpacks.my-app"

		[[build.env]]
		name="PROJECT_DESCRIPTOR_KEY"
		value="PROJECT_DESCRIPTOR_VALUE"

		[[build.env]]
		name="OVERWRITE"
		value="this should be overwritten"
		`

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile("project.toml", []byte(projectToml))
		addBuilderFiles(&utils)

		telemetryData := telemetry.CustomData{}
		err := callCnbBuild(&config, &telemetryData, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		assertLifecycleCalls(t, utils.ExecMockRunner, 2)

		assetBuildEnv(t, utils, "OPTIONS_KEY", "OPTIONS_VALUE")
		assetBuildEnv(t, utils, "PROJECT_DESCRIPTOR_KEY", "PROJECT_DESCRIPTOR_VALUE")
		assetBuildEnv(t, utils, "OVERWRITE", "this should win")
	})

	t.Run("pom.xml exists (symlink for the target folder)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.CurrentDir = "/jenkins"
		utils.FilesMock.AddDir("/jenkins")
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddFile("/workspace/pom.xml", []byte("test"))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		require.NoError(t, err)

		runner := utils.ExecMockRunner
		assertLifecycleCalls(t, runner, 2)

		assert.True(t, utils.FilesMock.HasCreatedSymlink("/jenkins/target", "/workspace/target"))
	})

	t.Run("no pom.xml exists (no symlink for the target folder)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.CurrentDir = "/jenkins"
		utils.FilesMock.AddDir("/jenkins")
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		require.NoError(t, err)

		runner := utils.ExecMockRunner
		assertLifecycleCalls(t, runner, 2)

		assert.False(t, utils.FilesMock.HasCreatedSymlink("/jenkins/target", "/workspace/target"))
	})

	t.Run("error case: Invalid DockerConfigJSON file", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			ContainerImageName:   "my-image",
			DockerConfigJSON:     "/path/to/config.json",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":"dXNlcjpwYXNz"}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to generate CNB_REGISTRY_AUTH: json: cannot unmarshal string into Go struct field ConfigFile.auths of type types.AuthConfig")
	})

	t.Run("error case: DockerConfigJSON file not there (config.json)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			ContainerImageName:   "my-image",
			DockerConfigJSON:     "not-there/config.json",
		}

		utils := newCnbBuildTestsUtils()
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to create/rename DockerConfigJSON file: cannot copy 'not-there/config.json': file does not exist")
	})

	t.Run("error case: DockerConfigJSON file not there (not config.json)", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageTag:    "0.0.1",
			ContainerRegistryURL: imageRegistry,
			ContainerImageName:   "my-image",
			DockerConfigJSON:     "not-there",
		}

		utils := newCnbBuildTestsUtils()
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to create/rename DockerConfigJSON file: cannot copy 'not-there': file does not exist")
	})

	t.Run("error case: dockerImage is not a valid builder", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{}

		utils := newCnbBuildTestsUtils()

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		assert.EqualError(t, err, "the provided dockerImage is not a valid builder: binary '/cnb/lifecycle/creator' not found")
	})

	t.Run("error case: builder image does not contain tls certificates", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:        "my-image",
			ContainerImageTag:         "0.0.1",
			ContainerRegistryURL:      imageRegistry,
			DockerConfigJSON:          "/path/to/config.json",
			Buildpacks:                []string{"test"},
			CustomTLSCertificateLinks: []string{"http://example.com/certs.pem"},
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		err := callCnbBuild(&config, &telemetry.CustomData{}, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		assert.EqualError(t, err, "failed to copy certificates: cannot copy '/etc/ssl/certs/ca-certificates.crt': file does not exist")
	})

	t.Run("success case (telemetry was added)", func(t *testing.T) {
		t.Parallel()
		registry := "some-registry"
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: registry,
			MultipleImages: []map[string]interface{}{
				{
					"runImage": "foo",
				},
				{
					"runImage": "bar",
				},
			},
		}

		utils := newCnbBuildTestsUtils()
		addBuilderFiles(&utils)

		telemetryData := &telemetry.CustomData{}
		err := callCnbBuild(&config, telemetryData, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		require.NoError(t, err)

		assert.Equal(t, "paketobuildpacks/builder-jammy-base:latest", telemetryData.CnbBuilder)
		assert.Equal(t, "foo,bar", telemetryData.CnbRunImage)
	})

	t.Run("error case, multiple artifacts in path", func(t *testing.T) {
		t.Parallel()
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			DockerConfigJSON:     "/path/to/config.json",
			ProjectDescriptor:    "project.toml",
			AdditionalTags:       []string{"latest"},
			Buildpacks:           []string{"paketobuildpacks/java", "gcr.io/paketo-buildpacks/node"},
			Path:                 "target/*.jar",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddDir("target")
		utils.FilesMock.AddFile("target/app.jar", []byte(`FFFFFF`))
		utils.FilesMock.AddFile("target/app-src.jar", []byte(`FFFFFF`))

		addBuilderFiles(&utils)

		telemetryData := telemetry.CustomData{}
		err := callCnbBuild(&config, &telemetryData, &utils, &cnbBuildCommonPipelineEnvironment{}, &piperhttp.Client{})
		require.EqualError(t, err, "could not resolve path: Failed to resolve glob for 'target/*.jar', matching 2 file(s)")
	})

	t.Run("success case, artifacts found by glob", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageName:   "my-image",
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: fmt.Sprintf("https://%s", imageRegistry),
			DockerConfigJSON:     "/path/to/config.json",
			ProjectDescriptor:    "project.toml",
			AdditionalTags:       []string{"latest"},
			Buildpacks:           []string{"paketobuildpacks/java", "gcr.io/paketo-buildpacks/node"},
			Path:                 "**/target",
		}

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		utils.FilesMock.AddDir("target")
		utils.FilesMock.AddFile("target/app.jar", []byte(`FFFFFF`))

		addBuilderFiles(&utils)

		telemetryData := telemetry.CustomData{}
		err := callCnbBuild(&config, &telemetryData, &utils, &commonPipelineEnvironment, &piperhttp.Client{})

		require.NoError(t, err)
		runner := utils.ExecMockRunner
		assert.Contains(t, runner.Env, "CNB_REGISTRY_AUTH={\"my-registry\":\"Basic dXNlcjpwYXNz\"}")
		assert.Contains(t, runner.Calls[1].Params, fmt.Sprintf("%s/%s:%s", imageRegistry, config.ContainerImageName, config.ContainerImageTag))
		assert.Equal(t, config.ContainerRegistryURL, commonPipelineEnvironment.container.registryURL)
		assert.Equal(t, "my-image:3.1.5", commonPipelineEnvironment.container.imageNameTag)
	})

	t.Run("success case (multiple images configured)", func(t *testing.T) {
		t.Parallel()
		commonPipelineEnvironment := cnbBuildCommonPipelineEnvironment{}
		config := cnbBuildOptions{
			ContainerImageTag:    "3.1.5",
			ContainerRegistryURL: imageRegistry,
			DockerConfigJSON:     "/path/to/my-config.json",
			AdditionalTags:       []string{"3", "3.1", "3.1", "3.1.5"},
			MultipleImages:       []map[string]interface{}{{"ContainerImageName": "my-image-0", "ContainerImageAlias": "simple"}, {"ContainerImageName": "my-image-1"}},
		}

		expectedImageCount := len(config.MultipleImages)

		utils := newCnbBuildTestsUtils()
		utils.FilesMock.AddFile(config.DockerConfigJSON, []byte(`{"auths":{"my-registry":{"auth":"dXNlcjpwYXNz"}}}`))
		addBuilderFiles(&utils)

		telemetryData := &telemetry.CustomData{}
		err := callCnbBuild(&config, telemetryData, &utils, &commonPipelineEnvironment, &piperhttp.Client{})
		require.NoError(t, err)

		runner := utils.ExecMockRunner
		require.Equal(t, expectedImageCount, len(runner.Calls)-1)
		for i, call := range runner.Calls {
			if i == 0 { // first call is -version
				continue
			}
			lifecycleCall := i - 1
			assertLifecycleCalls(t, runner, i+1)
			containerImageName := fmt.Sprintf("my-image-%d", lifecycleCall)
			assert.Contains(t, call.Params, fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, containerImageName, config.ContainerImageTag))
			assert.Contains(t, call.Params, fmt.Sprintf("%s/%s:3", config.ContainerRegistryURL, containerImageName))
			assert.Contains(t, call.Params, fmt.Sprintf("%s/%s:3.1", config.ContainerRegistryURL, containerImageName))
			assert.Contains(t, call.Params, fmt.Sprintf("%s/%s:3.1.5", config.ContainerRegistryURL, containerImageName))
		}

		assert.Equal(t, "my-image-0:3.1.5", commonPipelineEnvironment.container.imageNameTag)
		assert.Equal(t, []string{"simple", "my-image-1"}, commonPipelineEnvironment.container.imageNames)
	})
}
