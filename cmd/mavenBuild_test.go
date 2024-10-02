//go:build unit
// +build unit

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
)

func TestMavenBuild(t *testing.T) {

	cpe := mavenBuildCommonPipelineEnvironment{}

	t.Run("mavenBuild should install the artifact", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
	})

	t.Run("mavenBuild should skip integration tests", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("integration-tests/pom.xml", []byte{})

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-pl", "!integration-tests")
	})

	t.Run("mavenBuild should flatten", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Flatten: true}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "flatten:flatten")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-Dflatten.mode=resolveCiFriendliesOnly")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DupdatePomFile=true")
	})

	t.Run("mavenBuild should run only verify", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Verify: true}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "verify")
		assert.NotContains(t, mockedUtils.Calls[0].Params, "install")
	})

	t.Run("mavenBuild should createBOM", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{CreateBOM: true}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "org.cyclonedx:cyclonedx-maven-plugin:2.7.8:makeAggregateBom")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DschemaVersion=1.4")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeBomSerialNumber=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeCompileScope=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeProvidedScope=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeRuntimeScope=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeSystemScope=true")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeTestScope=false")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DincludeLicenseText=false")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DoutputFormat=xml")
		assert.Contains(t, mockedUtils.Calls[0].Params, "-DoutputName=bom-maven")
	})

	t.Run("mavenBuild include install and deploy when publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
		assert.NotContains(t, mockedUtils.Calls[0].Params, "verify")
		assert.Contains(t, mockedUtils.Calls[1].Params, "deploy")

	})

	t.Run("mavenBuild with deploy must skip build, install and test", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false, DeployFlags: []string{"-Dmaven.main.skip=true", "-Dmaven.test.skip=true", "-Dmaven.install.skip=true"}}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.main.skip=true")
		assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.test.skip=true")
		assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.install.skip=true")

	})

	t.Run("mavenBuild with deploy must include alt repo id and url when passed as parameter", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com"}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=ID::default::http://sampleRepo.com")
	})

	t.Run("mavenBuild accepts profiles", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Profiles: []string{"profile1", "profile2"}}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
		assert.Contains(t, mockedUtils.Calls[0].Params, "profile1,profile2")
	})

	t.Run("mavenBuild should not create build artifacts metadata when CreateBuildArtifactsMetadata is false and Publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		config := mavenBuildOptions{CreateBuildArtifactsMetadata: false, Publish: true}
		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)
		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Contains(t, mockedUtils.Calls[0].Params, "install")
		assert.Empty(t, cpe.custom.mavenBuildArtifacts)
	})

	t.Run("mavenBuild should not create build artifacts metadata when CreateBuildArtifactsMetadata is true and Publish is false", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()
		mockedUtils.AddFile("pom.xml", []byte{})
		config := mavenBuildOptions{CreateBuildArtifactsMetadata: true, Publish: false}
		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)
		assert.Nil(t, err)
		assert.Equal(t, mockedUtils.Calls[0].Exec, "mvn")
		assert.Empty(t, cpe.custom.mavenBuildArtifacts)
	})

}

func TestIsAggregatedBOM(t *testing.T) {
	t.Run("is aggregated BOM", func(t *testing.T) {
		bom := piperutils.Bom{
			Metadata: piperutils.Metadata{
				Properties: []piperutils.BomProperty{
					{Name: "maven.goal", Value: "makeAggregateBom"},
				},
			},
		}
		assert.True(t, isAggregatedBOM(bom))
	})

	t.Run("is not aggregated BOM", func(t *testing.T) {
		bom := piperutils.Bom{
			Metadata: piperutils.Metadata{
				Properties: []piperutils.BomProperty{
					{Name: "some.property", Value: "someValue"},
				},
			},
		}
		assert.False(t, isAggregatedBOM(bom))
	})
}

func createTempFile(t *testing.T, dir string, filename string, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0666)
	if err != nil {
		t.Fatalf("Failed to create temp file: %s", err)
	}
	return filePath
}

func TestGetPurlForThePomAndDeleteIndividualBom(t *testing.T) {
	t.Run("valid BOM file, non-aggregated", func(t *testing.T) {
		tempDir, err := piperutils.Files{}.TempDir("", "test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %s", err)
		}

		bomContent := `<bom>
			<metadata>
				<component>
					<purl>pkg:maven/com.example/mycomponent@1.0.0</purl>
				</component>
				<properties>
					<property name="name1" value="value1" />
				</properties>
			</metadata>
		</bom>`
		pomFilePath := createTempFile(t, tempDir, "pom.xml", "")
		bomDir := filepath.Join(tempDir, "target")
		if err := os.MkdirAll(bomDir, 0777); err != nil {
			t.Fatalf("Failed to create temp directory: %s", err)
		}
		bomFilePath := createTempFile(t, bomDir, mvnBomFilename+".xml", bomContent)
		defer os.Remove(bomFilePath)

		purl := getPurlForThePomAndDeleteIndividualBom(pomFilePath)
		assert.Equal(t, "pkg:maven/com.example/mycomponent@1.0.0", purl)
		_, err = os.Stat(bomFilePath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("valid BOM file, aggregated BOM", func(t *testing.T) {
		tempDir, err := piperutils.Files{}.TempDir("", "test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %s", err)
		}

		bomContent := `<bom>
			<metadata>
				<component>
					<purl>pkg:maven/com.example/aggregatecomponent@1.0.0</purl>
				</component>
				<properties>
					<property name="maven.goal" value="makeAggregateBom" />
				</properties>
			</metadata>
		</bom>`
		pomFilePath := createTempFile(t, tempDir, "pom.xml", "")
		bomDir := filepath.Join(tempDir, "target")
		if err := os.MkdirAll(bomDir, 0777); err != nil {
			t.Fatalf("Failed to create temp directory: %s", err)
		}
		bomFilePath := createTempFile(t, bomDir, mvnBomFilename+".xml", bomContent)

		purl := getPurlForThePomAndDeleteIndividualBom(pomFilePath)
		assert.Equal(t, "pkg:maven/com.example/aggregatecomponent@1.0.0", purl)
		_, err = os.Stat(bomFilePath)
		assert.False(t, os.IsNotExist(err)) // File should not be deleted
	})

	t.Run("BOM file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		pomFilePath := createTempFile(t, tempDir, "pom.xml", "") // Create a temp pom file

		purl := getPurlForThePomAndDeleteIndividualBom(pomFilePath)
		assert.Equal(t, "", purl)
	})
}
