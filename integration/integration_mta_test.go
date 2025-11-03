//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestMTAIntegration ./integration/...

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
)

const defaultDockerImage = "devxci/mbtci-java21-node22"

func TestMTAIntegrationMavenProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	req := testcontainers.ContainerRequest{
		Image: defaultDockerImage,
		User:  "root",
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestMtaIntegration", "maven"),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", pwd),
			}
		},
	}

	mtaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer mtaContainer.Terminate(ctx)

	code, reader, err := mtaContainer.Exec(ctx, []string{"/piperbin/piper", "mtaBuild", "--installArtifacts", "--m2Path=mym2"}, exec.WithWorkingDir("/maven"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "Installing /maven/.flattened-pom.xml to /maven/mym2/mygroup/mymvn/1.0-SNAPSHOT/mymvn-1.0-SNAPSHOT.pom")
	assert.Contains(t, output, "Installing /maven/app/target/mymvn-app-1.0-SNAPSHOT.war to /maven/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT.war")
	assert.Contains(t, output, "Installing /maven/app/target/mymvn-app-1.0-SNAPSHOT-classes.jar to /maven/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT-classes.jar")
	assert.Contains(t, output, "added 2 packages, and audited 3 packages in")
}

func TestMTAIntegrationMavenSpringProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	req := testcontainers.ContainerRequest{
		Image: defaultDockerImage,
		User:  "root",
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestMtaIntegration", "maven-spring"),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", pwd),
			}
		},
	}

	mtaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer mtaContainer.Terminate(ctx)

	code, reader, err := mtaContainer.Exec(ctx, []string{"/piperbin/piper", "mtaBuild", "--installArtifacts", "--m2Path=mym2"}, exec.WithWorkingDir("/maven-spring"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	_, err = io.ReadAll(reader)
	assert.NoError(t, err)

	code, reader, err = mtaContainer.Exec(ctx, []string{"/piperbin/piper", "mavenExecuteIntegration", "--m2Path=mym2"}, exec.WithWorkingDir("/maven-spring"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")
}

func TestMTAIntegrationNPMProject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	req := testcontainers.ContainerRequest{
		Image: defaultDockerImage,
		User:  "root",
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestMtaIntegration", "npm"),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", pwd),
			}
		},
	}

	mtaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer mtaContainer.Terminate(ctx)

	code, reader, err := mtaContainer.Exec(ctx, []string{"/piperbin/piper", "mtaBuild"}, exec.WithWorkingDir("/npm"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "INFO the MTA archive generated at: /npm/test-mta-js.mtar")
}

func TestMTAIntegrationNPMProjectInstallsDevDependencies(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	req := testcontainers.ContainerRequest{
		Image: defaultDockerImage,
		User:  "root",
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestMtaIntegration", "npm-install-dev-dependencies"),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", pwd),
			}
		},
	}

	mtaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer mtaContainer.Terminate(ctx)

	code, reader, err := mtaContainer.Exec(ctx, []string{"/piperbin/piper", "mtaBuild", "--installArtifacts"}, exec.WithWorkingDir("/npm-install-dev-dependencies"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "added 2 packages, and audited 3 packages in")
}

func TestMTAIntegrationNPMProjectWithSeparateBOMValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	pwd, err := os.Getwd()
	assert.NoError(t, err, "Getting current working directory failed.")
	pwd = filepath.Dir(pwd)

	req := testcontainers.ContainerRequest{
		Image: defaultDockerImage,
		User:  "root",
		Cmd:   []string{"tail", "-f"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(pwd, "integration", "testdata", "TestMtaIntegration", "npm"),
				ContainerFilePath: "/",
				FileMode:          0755,
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{
				fmt.Sprintf("%s:/piperbin", pwd),
			}
		},
	}

	mtaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer mtaContainer.Terminate(ctx)

	code, reader, err := mtaContainer.Exec(ctx, []string{"/piperbin/piper", "mtaBuild", "--createBOM"}, exec.WithWorkingDir("/npm"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	_, err = io.ReadAll(reader)
	assert.NoError(t, err)

	code, _, err = mtaContainer.Exec(ctx, []string{"test", "-f", "/npm/sbom-gen/bom-mta.xml"})
	assert.NoError(t, err)
	assert.Equal(t, 0, code, "BOM file /npm/sbom-gen/bom-mta.xml should exist")

	code, reader, err = mtaContainer.Exec(ctx, []string{"/piperbin/piper", "validateBOM", "--bomPattern", "**/sbom-gen/bom-*.xml"}, exec.WithWorkingDir("/npm"))
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	outputBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-mta.xml")
	assert.Contains(t, output, "info  validateBOM - BOM validation passed:")
	assert.Contains(t, output, "info  validateBOM - BOM PURL:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 1/1 files validated successfully")
}
