package cmd

import (
	"errors"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

type mavenMockUtils struct {
	shouldFail bool
	*mock.FilesMock
	*mock.ExecMockRunner
}

func (m *mavenMockUtils) DownloadFile(_, _ string, _ http.Header, _ []*http.Cookie) error {
	return errors.New("Test should not download files.")
}

func newMavenMockUtils() mavenMockUtils {
	utils := mavenMockUtils{
		shouldFail:     false,
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

var cpe mavenBuildCommonPipelineEnvironment

func TestMavenBuild(t *testing.T) {
	t.Run("mavenBuild should install the artifact", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)
		expectedParamsFirstCall := []string{"org.cyclonedx:cyclonedx-maven-plugin:2.7.8:makeBom"}
		expectedParamsSecondCall := []string{"install"}

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two maven invocations (makeBOM and main build)") {
			assert.Equal(t, "mvn", mockedUtils.Calls[0].Exec)
			assert.Contains(t, mockedUtils.Calls[0].Params, expectedParamsFirstCall[0], "First call should contain makeBom goal")

			assert.Equal(t, "mvn", mockedUtils.Calls[1].Exec)
			assert.Contains(t, mockedUtils.Calls[1].Params, expectedParamsSecondCall[0], "Second call should contain install goal")
		}
	})

	t.Run("mavenBuild should accept profiles", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Profiles: []string{"profile1", "profile2"}}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)

		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two maven invocations (makeBOM and main build)") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "--activate-profiles")
			assert.True(t, strings.Contains(mockedUtils.Calls[0].Params[1], "profile1,profile2"), "Profiles should be activated")
			assert.Contains(t, mockedUtils.Calls[1].Params, "--activate-profiles")
			assert.True(t, strings.Contains(mockedUtils.Calls[1].Params[1], "profile1,profile2"), "Profiles should be activated in the second call as well")
		}
	})

	t.Run("mavenBuild should createBOM", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{CreateBOM: true}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (makeBOM and makeAggregateBOM or main goals)") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "org.cyclonedx:cyclonedx-maven-plugin:2.7.8:makeBom")
			assert.Contains(t, mockedUtils.Calls[1].Params, "org.cyclonedx:cyclonedx-maven-plugin:2.7.8:makeAggregateBom")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DschemaVersion=1.4")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DincludeBomSerialNumber=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DincludeCompileScope=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DincludeProvidedScope=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DincludeRuntimeScope=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DincludeSystemScope=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DincludeTestScope=false")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DincludeLicenseText=false")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DoutputFormat=xml")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DoutputName=bom-maven")
		}
	})

	t.Run("mavenBuild include install and deploy when publish is true", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[0].Params, "install")
			assert.NotContains(t, mockedUtils.Calls[0].Params, "verify")
			assert.Contains(t, mockedUtils.Calls[1].Params, "deploy")
		}
	})

	t.Run("mavenBuild with deploy must skip build, install and test", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false, DeployFlags: []string{"-Dmaven.main.skip=true", "-Dmaven.test.skip=true", "-Dmaven.install.skip=true"}}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.main.skip=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.test.skip=true")
			assert.Contains(t, mockedUtils.Calls[1].Params, "-Dmaven.install.skip=true")
		}
	})

	t.Run("mavenBuild with deploy must include alt repo id and url when passed as parameter", func(t *testing.T) {
		mockedUtils := newMavenMockUtils()

		config := mavenBuildOptions{Publish: true, Verify: false, AltDeploymentRepositoryID: "ID", AltDeploymentRepositoryURL: "http://sampleRepo.com"}

		err := runMavenBuild(&config, nil, &mockedUtils, &cpe)

		assert.Nil(t, err)
		if assert.Equal(t, 2, len(mockedUtils.Calls), "Expected two Maven invocations (main and deploy)") {
			assert.Contains(t, mockedUtils.Calls[1].Params, "-DaltDeploymentRepository=ID::default::http://sampleRepo.com")
		}
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
