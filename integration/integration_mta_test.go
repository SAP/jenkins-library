//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestMTAIntegration ./integration/...

package main

import (
	"testing"

	"github.com/SAP/jenkins-library/integration/testhelper"
	"github.com/stretchr/testify/assert"
)

const defaultDockerImage = "devxci/mbtci-java21-node22"

func TestMTAIntegrationMavenProject(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    defaultDockerImage,
		TestData: "TestMtaIntegration/maven",
		WorkDir:  "/maven",
		User:     "root",
	})

	output := testhelper.RunPiper(t, container, "/maven", "mtaBuild", "--installArtifacts", "--m2Path=mym2")

	assert.Contains(t, output, "Installing /maven/.flattened-pom.xml to /maven/mym2/mygroup/mymvn/1.0-SNAPSHOT/mymvn-1.0-SNAPSHOT.pom")
	assert.Contains(t, output, "Installing /maven/app/target/mymvn-app-1.0-SNAPSHOT.war to /maven/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT.war")
	assert.Contains(t, output, "Installing /maven/app/target/mymvn-app-1.0-SNAPSHOT-classes.jar to /maven/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT-classes.jar")
	assert.Contains(t, output, "added 2 packages, and audited 3 packages in")
}

func TestMTAIntegrationMavenSpringProject(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    defaultDockerImage,
		TestData: "TestMtaIntegration/maven-spring",
		WorkDir:  "/maven-spring",
		User:     "root",
	})

	testhelper.RunPiper(t, container, "/maven-spring", "mtaBuild", "--installArtifacts", "--m2Path=mym2")

	output := testhelper.RunPiper(t, container, "/maven-spring", "mavenExecuteIntegration", "--m2Path=mym2")
	assert.Contains(t, output, "Tests run: 1, Failures: 0, Errors: 0, Skipped: 0")
}

func TestMTAIntegrationNPMProject(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    defaultDockerImage,
		TestData: "TestMtaIntegration/npm",
		WorkDir:  "/npm",
		User:     "root",
	})

	output := testhelper.RunPiper(t, container, "/npm", "mtaBuild")
	assert.Contains(t, output, "INFO the MTA archive generated at: /npm/test-mta-js.mtar")
}

func TestMTAIntegrationNPMProjectInstallsDevDependencies(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    defaultDockerImage,
		TestData: "TestMtaIntegration/npm-install-dev-dependencies",
		WorkDir:  "/npm-install-dev-dependencies",
		User:     "root",
	})

	output := testhelper.RunPiper(t, container, "/npm-install-dev-dependencies", "mtaBuild", "--installArtifacts")
	assert.Contains(t, output, "added 2 packages, and audited 3 packages in")
}

func TestMTAIntegrationNPMProjectWithSeparateBOMValidation(t *testing.T) {
	t.Parallel()

	container := testhelper.StartPiperContainer(t, testhelper.ContainerConfig{
		Image:    defaultDockerImage,
		TestData: "TestMtaIntegration/npm",
		WorkDir:  "/npm",
		User:     "root",
	})

	testhelper.RunPiper(t, container, "/npm", "mtaBuild", "--createBOM")

	testhelper.AssertFileExists(t, container, "/npm/sbom-gen/bom-mta.xml")

	output := testhelper.RunPiper(t, container, "/npm", "validateBOM", "--bomPattern", "**/sbom-gen/bom-*.xml")
	assert.Contains(t, output, "info  validateBOM - Found 1 BOM file(s) to validate")
	assert.Contains(t, output, "info  validateBOM - Validating BOM file:")
	assert.Contains(t, output, "bom-mta.xml")
	assert.Contains(t, output, "info  validateBOM - BOM validation passed:")
	assert.Contains(t, output, "info  validateBOM - BOM PURL:")
	assert.Contains(t, output, "info  validateBOM - BOM validation complete: 1/1 files validated successfully")
}
