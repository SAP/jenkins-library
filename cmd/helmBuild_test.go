//go:build unit
// +build unit

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/SAP/jenkins-library/pkg/build"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/kubernetes/mocks"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/versioning"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type helmMockUtilsBundle struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.HttpClientMock
}

func newHelmMockUtilsBundle() helmMockUtilsBundle {
	utils := helmMockUtilsBundle{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		HttpClientMock: &mock.HttpClientMock{
			FileUploads: map[string]string{},
		},
	}
	return utils
}

func setupConfigOpenFileMock(t *testing.T) {
	t.Helper()
	openFileBak := configOptions.OpenFile
	t.Cleanup(func() { configOptions.OpenFile = openFileBak })
	configOptions.OpenFile = configOpenFileMock
}

func TestRunHelmUpgrade(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	cpe := helmBuildCommonPipelineEnvironment{}
	testTable := []struct {
		config         helmBuildOptions
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmBuildOptions{
				HelmCommand: "upgrade",
			},
			methodError: nil,
		},
		{
			config: helmBuildOptions{
				HelmCommand: "upgrade",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute upgrade: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmUpgrade").Return(testCase.methodError)

			err := runHelmBuild(testCase.config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmLint(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	cpe := helmBuildCommonPipelineEnvironment{}
	testTable := []struct {
		config         helmBuildOptions
		expectedConfig []string
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmBuildOptions{
				HelmCommand: "lint",
			},
			methodError: nil,
		},
		{
			config: helmBuildOptions{
				HelmCommand: "lint",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm lint: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmLint").Return(testCase.methodError)

			err := runHelmBuild(testCase.config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmInstall(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	cpe := helmBuildCommonPipelineEnvironment{}
	testTable := []struct {
		config         helmBuildOptions
		expectedConfig []string
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmBuildOptions{
				HelmCommand: "install",
			},
			methodError: nil,
		},
		{
			config: helmBuildOptions{
				HelmCommand: "install",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm install: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmInstall").Return(testCase.methodError)

			err := runHelmBuild(testCase.config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmTest(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	cpe := helmBuildCommonPipelineEnvironment{}
	testTable := []struct {
		config         helmBuildOptions
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmBuildOptions{
				HelmCommand: "test",
			},
			methodError: nil,
		},
		{
			config: helmBuildOptions{
				HelmCommand: "test",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm test: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmTest").Return(testCase.methodError)

			err := runHelmBuild(testCase.config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmUninstall(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	cpe := helmBuildCommonPipelineEnvironment{}
	testTable := []struct {
		config         helmBuildOptions
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmBuildOptions{
				HelmCommand: "uninstall",
			},
			methodError: nil,
		},
		{
			config: helmBuildOptions{
				HelmCommand: "uninstall",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm uninstall: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmUninstall").Return(testCase.methodError)

			err := runHelmBuild(testCase.config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmDependency(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	cpe := helmBuildCommonPipelineEnvironment{}
	testTable := []struct {
		config         helmBuildOptions
		methodError    error
		expectedErrStr string
	}{
		{
			config: helmBuildOptions{
				HelmCommand: "dependency",
			},
			methodError: nil,
		},
		{
			config: helmBuildOptions{
				HelmCommand: "dependency",
			},
			methodError:    errors.New("some error"),
			expectedErrStr: "failed to execute helm dependency: some error",
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmDependency").Return(testCase.methodError)

			err := runHelmBuild(testCase.config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
		})

	}
}

func TestRunHelmPush(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	testTable := []struct {
		config             helmBuildOptions
		artifactInfo       versioning.Coordinates
		methodString       string
		methodError        error
		expectedErrStr     string
		expectArtifactsSet bool
	}{
		{
			config: helmBuildOptions{
				HelmCommand: "publish",
			},
			artifactInfo:       versioning.Coordinates{ArtifactID: "my-chart", Version: "1.2.3"},
			methodString:       "https://my.target.repository/my-chart-1.2.3.tgz",
			methodError:        nil,
			expectArtifactsSet: true,
		},
		{
			config: helmBuildOptions{
				HelmCommand: "publish",
			},
			methodError:        errors.New("some error"),
			expectedErrStr:     "failed to execute helm publish: some error",
			expectArtifactsSet: false,
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			cpe := helmBuildCommonPipelineEnvironment{}
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmPublish").Return(testCase.methodString, testCase.methodError)

			err := runHelmBuild(testCase.config, helmExecutor, &fileHandlerMock{}, &cpe, testCase.artifactInfo, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
			if testCase.expectedErrStr != "" {
				assert.EqualError(t, err, testCase.expectedErrStr)
			} else {
				assert.NoError(t, err)
			}

			if testCase.expectArtifactsSet {
				assert.NotEmpty(t, cpe.custom.helmBuildArtifacts, "helmBuildArtifacts must be set after successful publish")
				var artifacts build.BuildArtifacts
				require.NoError(t, json.Unmarshal([]byte(cpe.custom.helmBuildArtifacts), &artifacts))
				require.Len(t, artifacts.Coordinates, 1)
				assert.Equal(t, testCase.artifactInfo.ArtifactID, artifacts.Coordinates[0].ArtifactID)
				assert.Equal(t, testCase.artifactInfo.Version, artifacts.Coordinates[0].Version)
				assert.Equal(t, testCase.methodString, artifacts.Coordinates[0].URL)
			} else {
				assert.Empty(t, cpe.custom.helmBuildArtifacts, "helmBuildArtifacts must not be set on publish failure")
			}
		})
	}
}

func TestRunHelmPushSBOM(t *testing.T) {
	setupConfigOpenFileMock(t)
	// SBOM behaviour of the explicit "publish" command path in runHelmBuild.
	testTable := []struct {
		name       string
		createBOM  bool
		expectSyft bool
		expectPurl bool
	}{
		{
			name:       "createBOM disabled - no syft",
			createBOM:  false,
			expectSyft: false,
			expectPurl: false,
		},
		{
			name:       "createBOM enabled - syft + root purl injected",
			createBOM:  true,
			expectSyft: true,
			expectPurl: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			utils := newHelmMockUtilsBundle()
			execRunner := &mock.ExecMockRunner{}
			client := setupSyftDownloadMock(t)

			cpe := helmBuildCommonPipelineEnvironment{}
			config := helmBuildOptions{
				HelmCommand:            "publish",
				CreateBOM:              testCase.createBOM,
				SyftDownloadURL:        "http://test-syft-url.io",
				ContainerImageNameTags: []string{"registry.example.com/foo:1.0.0"},
			}
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmPublish").Return("https://my.target.repository/foo-1.0.0.tgz", nil)
			// helm template yields no images → SBOM falls back to the CPE list.
			helmExecutor.On("RunHelmTemplate").Return([]byte(nil), nil)

			// The fake syft binary does not write a BOM, so pre-create the
			// bom-docker-0.xml (root component without a PURL) syft would produce.
			require.NoError(t, os.WriteFile("bom-docker-0.xml", []byte(bomWithoutRootPurl), 0o644))
			defer os.Remove("bom-docker-0.xml")

			// Act
			err := runHelmBuild(config, helmExecutor, utils, &cpe, versioning.Coordinates{}, execRunner, &mock.FilesMock{}, client)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectSyft, syftScanInvoked(execRunner),
				"syft scan invocation must match createBOM=%v", testCase.createBOM)

			updated, err := os.ReadFile("bom-docker-0.xml")
			require.NoError(t, err)
			if testCase.expectPurl {
				assert.Contains(t, string(updated), "<purl>pkg:docker/nginx@1.25</purl>",
					"root PURL must be injected when createBOM=true")
			} else {
				assert.NotContains(t, string(updated), "<purl>",
					"no PURL must be injected when createBOM=false")
			}
		})
	}
}

func TestRunHelmDefaultCommand(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	cpe := helmBuildCommonPipelineEnvironment{}
	testTable := []struct {
		config             helmBuildOptions
		methodLintError    error
		methodPackageError error
		methodPublishError error
		expectedErrStr     string
		fileUtils          fileHandlerMock
		assertFunc         func(fileHandlerMock) error
	}{
		{
			config: helmBuildOptions{
				HelmCommand: "",
			},
			methodLintError:    nil,
			methodPackageError: nil,
			methodPublishError: nil,
			fileUtils:          fileHandlerMock{},
		},
		{
			// this test checks if parseAndRenderCPETemplate is called properly
			// when config.RenderValuesTemplate is true
			config: helmBuildOptions{
				HelmCommand:          "",
				RenderValuesTemplate: true,
			},
			methodLintError:    nil,
			methodPackageError: nil,
			methodPublishError: nil,
			fileUtils:          fileHandlerMock{},
			// we expect the values file is traversed since parsing and rendering according to cpe template is active
			assertFunc: func(f fileHandlerMock) error {
				if len(f.fileExistsCalled) == 1 && f.fileExistsCalled[0] == "/values.yaml" {
					return nil
				}
				return fmt.Errorf("expected FileExists called for ['/values.yaml'] but was: %+v", f.fileExistsCalled)
			},
		},
		{
			// this test checks if parseAndRenderCPETemplate is NOT called
			// when config.RenderValuesTemplate is false
			config: helmBuildOptions{
				HelmCommand:          "",
				RenderValuesTemplate: false,
			},
			methodLintError:    nil,
			methodPackageError: nil,
			methodPublishError: nil,
			fileUtils:          fileHandlerMock{},
			// we expect the values file is not traversed since parsing and rendering according to cpe template is not active
			assertFunc: func(f fileHandlerMock) error {
				if len(f.fileExistsCalled) > 0 {
					return fmt.Errorf("expected FileExists not called, but was for: %+v", f.fileExistsCalled)
				}
				return nil
			},
		},
		{
			config: helmBuildOptions{
				HelmCommand: "",
			},
			methodLintError: errors.New("some error"),
			expectedErrStr:  "failed to execute helm lint: some error",
			fileUtils:       fileHandlerMock{},
		},
		{
			config: helmBuildOptions{
				HelmCommand: "",
			},
			methodPackageError: errors.New("some error"),
			expectedErrStr:     "failed to execute helm dependency: some error",
			fileUtils:          fileHandlerMock{},
		},
		{
			config: helmBuildOptions{
				HelmCommand: "",
			},
			methodPublishError: errors.New("some error"),
			expectedErrStr:     "failed to execute helm publish: some error",
			fileUtils:          fileHandlerMock{},
		},
	}

	for i, testCase := range testTable {
		t.Run(fmt.Sprint("case ", i), func(t *testing.T) {
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmDependency").Return(testCase.methodPackageError)
			helmExecutor.On("RunHelmLint").Return(testCase.methodLintError)
			helmExecutor.On("RunHelmPublish").Return(testCase.methodPublishError)

			err := runHelmBuild(testCase.config, helmExecutor, &testCase.fileUtils, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
			if err != nil {
				assert.Equal(t, testCase.expectedErrStr, err.Error())
			}
			if testCase.assertFunc != nil {
				assert.NoError(t, testCase.assertFunc(testCase.fileUtils))
			}

		})
	}

}

func TestWriteHelmBuildArtifacts(t *testing.T) {
	setupConfigOpenFileMock(t)

	t.Run("default publish=true path sets helmBuildArtifacts in CPE", func(t *testing.T) {
		cpe := helmBuildCommonPipelineEnvironment{}
		artifactInfo := versioning.Coordinates{ArtifactID: "my-chart", Version: "0.5.0"}
		publishURL := "https://my.target.repository/my-chart-0.5.0.tgz"

		helmExecutor := &mocks.HelmExecutor{}
		helmExecutor.On("RunHelmLint").Return(nil)
		helmExecutor.On("RunHelmPublish").Return(publishURL, nil)

		config := helmBuildOptions{Publish: true}
		err := runHelmBuild(config, helmExecutor, &fileHandlerMock{}, &cpe, artifactInfo, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())

		assert.NoError(t, err)
		assert.NotEmpty(t, cpe.custom.helmBuildArtifacts)
		var artifacts build.BuildArtifacts
		require.NoError(t, json.Unmarshal([]byte(cpe.custom.helmBuildArtifacts), &artifacts))
		require.Len(t, artifacts.Coordinates, 1)
		assert.Equal(t, "my-chart", artifacts.Coordinates[0].ArtifactID)
		assert.Equal(t, "0.5.0", artifacts.Coordinates[0].Version)
		assert.Equal(t, publishURL, artifacts.Coordinates[0].URL)
	})

	t.Run("publish=false does not set helmBuildArtifacts", func(t *testing.T) {
		cpe := helmBuildCommonPipelineEnvironment{}
		helmExecutor := &mocks.HelmExecutor{}
		helmExecutor.On("RunHelmLint").Return(nil)

		config := helmBuildOptions{Publish: false}
		err := runHelmBuild(config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())

		assert.NoError(t, err)
		assert.Empty(t, cpe.custom.helmBuildArtifacts)
	})
}

func TestRunHelmDefaultCommandSBOM(t *testing.T) {
	setupConfigOpenFileMock(t)
	// SBOM behaviour of the default flow (helmCommand='' + publish=true) in
	// runHelmBuildDefault.
	testTable := []struct {
		name       string
		createBOM  bool
		expectSyft bool
		expectPurl bool
	}{
		{
			name:       "createBOM disabled - no syft",
			createBOM:  false,
			expectSyft: false,
			expectPurl: false,
		},
		{
			name:       "createBOM enabled - syft + root purl injected",
			createBOM:  true,
			expectSyft: true,
			expectPurl: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			utils := newHelmMockUtilsBundle()
			execRunner := &mock.ExecMockRunner{}
			client := setupSyftDownloadMock(t)

			cpe := helmBuildCommonPipelineEnvironment{}
			config := helmBuildOptions{
				HelmCommand:            "",
				Publish:                true,
				CreateBOM:              testCase.createBOM,
				SyftDownloadURL:        "http://test-syft-url.io",
				ContainerImageNameTags: []string{"registry.example.com/foo:1.0.0"},
			}
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmLint").Return(nil)
			helmExecutor.On("RunHelmPublish").Return("https://my.target.repository/foo-1.0.0.tgz", nil)
			// helm template yields no images → SBOM falls back to the CPE list.
			helmExecutor.On("RunHelmTemplate").Return([]byte(nil), nil)

			require.NoError(t, os.WriteFile("bom-docker-0.xml", []byte(bomWithoutRootPurl), 0o644))
			defer os.Remove("bom-docker-0.xml")

			// Act
			err := runHelmBuild(config, helmExecutor, utils, &cpe, versioning.Coordinates{}, execRunner, &mock.FilesMock{}, client)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectSyft, syftScanInvoked(execRunner),
				"syft scan invocation must match createBOM=%v", testCase.createBOM)

			updated, err := os.ReadFile("bom-docker-0.xml")
			require.NoError(t, err)
			if testCase.expectPurl {
				assert.Contains(t, string(updated), "<purl>pkg:docker/nginx@1.25</purl>",
					"root PURL must be injected when createBOM=true")
			} else {
				assert.NotContains(t, string(updated), "<purl>",
					"no PURL must be injected when createBOM=false")
			}
		})
	}
}

// bomWithoutRootPurl is a minimal image BOM as syft emits it: the root/parent
// component carries a registry-qualified name (with a host:port, the case that
// reproduces the real validator failure) and NO <purl> (anchore/syft#1408).
// Used by the SBOM tests to verify the root PURL gets injected, registry-free.
const bomWithoutRootPurl = `<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component bom-ref="abc123" type="container">
			<name>host.docker.internal:5000/library/nginx</name>
			<version>1.25</version>
		</component>
	</metadata>
	<components></components>
</bom>`

// bomEmptyRootName has a root component with no <name>: GetComponent returns an
// empty name, so injection skips this file (best-effort, no error, no PURL).
const bomEmptyRootName = `<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component bom-ref="abc123" type="container">
			<name></name>
			<version>1.25</version>
		</component>
	</metadata>
	<components></components>
</bom>`

// bomInvalidRootName has a root component whose name is not a valid image
// reference (spaces/uppercase), so purl.RefToPURL fails and injection skips
// this file (best-effort, no error, no PURL).
const bomInvalidRootName = `<bom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component bom-ref="abc123" type="container">
			<name>Invalid Name</name>
			<version>1.25</version>
		</component>
	</metadata>
	<components></components>
</bom>`

// bomWrongRootElement is parseable by the lenient GetComponent (yields a
// non-empty name) but is rejected by the strict CycloneDX decoder inside
// UpdatePurl (root element is not <bom>), driving the UpdatePurl error branch.
const bomWrongRootElement = `<notbom xmlns="http://cyclonedx.org/schema/bom/1.4" version="1">
	<metadata>
		<component bom-ref="abc123" type="container">
			<name>registry.example.com/foo</name>
			<version>1.0.0</version>
		</component>
	</metadata>
</notbom>`

func TestRunHelmInjectContainerBOMPurls(t *testing.T) {
	t.Run("injects a clean registry-free docker purl into the root component", func(t *testing.T) {
		require.NoError(t, os.WriteFile("bom-docker-0.xml", []byte(bomWithoutRootPurl), 0o644))
		defer os.Remove("bom-docker-0.xml")

		err := injectContainerBOMPurls()

		require.NoError(t, err)
		updated, err := os.ReadFile("bom-docker-0.xml")
		require.NoError(t, err)
		assert.Contains(t, string(updated), "<purl>pkg:docker/nginx@1.25</purl>",
			"root component must carry a clean docker PURL after injection")
		// The injected PURL must not carry the registry host:port. (The component
		// <name> still holds syft's original registry-qualified name; UpdatePurl
		// only adds the <purl>, it does not rewrite the name.)
		assert.NotContains(t, string(updated), "pkg:docker/host.docker.internal",
			"the injected PURL must not carry the registry host")
	})

	t.Run("no bom-docker files - no-op, no error", func(t *testing.T) {
		// Ensure there is no matching file in the working directory.
		_ = os.Remove("bom-docker-0.xml")

		err := injectContainerBOMPurls()
		assert.NoError(t, err, "injection must be a no-op when no bom-docker files exist")
	})

	t.Run("failing BOM is skipped but later BOMs are still processed", func(t *testing.T) {
		// For each failure mode: bom-docker-0 fails, bom-docker-1 is valid. The
		// glob is lexicographically sorted, so 0 is processed first; its failure
		// must not fail the step nor prevent 1 from getting its PURL injected.
		failingFixtures := []struct {
			name    string
			content string
		}{
			{name: "empty root name", content: bomEmptyRootName},
			{name: "invalid image name (RefToPURL fails)", content: bomInvalidRootName},
			{name: "undecodable BOM (UpdatePurl fails)", content: bomWrongRootElement},
		}
		for _, fixture := range failingFixtures {
			t.Run(fixture.name, func(t *testing.T) {
				require.NoError(t, os.WriteFile("bom-docker-0.xml", []byte(fixture.content), 0o644))
				defer os.Remove("bom-docker-0.xml")
				require.NoError(t, os.WriteFile("bom-docker-1.xml", []byte(bomWithoutRootPurl), 0o644))
				defer os.Remove("bom-docker-1.xml")

				err := injectContainerBOMPurls()

				assert.NoError(t, err, "a per-file failure must not fail the step")
				failed, err := os.ReadFile("bom-docker-0.xml")
				require.NoError(t, err)
				assert.NotContains(t, string(failed), "<purl>", "the failing BOM must be left untouched")
				succeeded, err := os.ReadFile("bom-docker-1.xml")
				require.NoError(t, err)
				assert.Contains(t, string(succeeded), "<purl>",
					"a later BOM must still get its PURL injected after an earlier failure")
			})
		}
	})

	t.Run("multiple valid BOMs all get a PURL", func(t *testing.T) {
		require.NoError(t, os.WriteFile("bom-docker-0.xml", []byte(bomWithoutRootPurl), 0o644))
		defer os.Remove("bom-docker-0.xml")
		require.NoError(t, os.WriteFile("bom-docker-1.xml", []byte(bomWithoutRootPurl), 0o644))
		defer os.Remove("bom-docker-1.xml")

		err := injectContainerBOMPurls()

		assert.NoError(t, err)
		for _, file := range []string{"bom-docker-0.xml", "bom-docker-1.xml"} {
			content, err := os.ReadFile(file)
			require.NoError(t, err)
			assert.Contains(t, string(content), "<purl>", "%s must get its PURL injected", file)
		}
	})
}

// setupSyftDownloadMock activates httpmock, registers a fake syft tar.gz archive
// at the test syft download URL, and returns a real http client wired to it.
// The caller relies on t.Cleanup for deactivation.
func setupSyftDownloadMock(t *testing.T) *piperhttp.Client {
	t.Helper()
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	fakeArchive, err := (&mock.FilesMock{}).CreateArchive(map[string][]byte{"syft": []byte("test")})
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "http://test-syft-url.io", httpmock.NewBytesResponder(http.StatusOK, fakeArchive))
	client := &piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
	return client
}

// syftScanInvoked reports whether a syft "scan" command was executed.
func syftScanInvoked(execRunner *mock.ExecMockRunner) bool {
	for _, call := range execRunner.Calls {
		if slices.Contains(call.Params, "scan") {
			return true
		}
	}
	return false
}

// syftScanCalls returns all executed syft "scan" invocations, in order.
func syftScanCalls(execRunner *mock.ExecMockRunner) []mock.ExecCall {
	var calls []mock.ExecCall
	for _, call := range execRunner.Calls {
		if len(call.Params) > 0 && call.Params[0] == "scan" {
			calls = append(calls, call)
		}
	}
	return calls
}

func TestRunHelmGenerateContainerSBOMs(t *testing.T) {
	// cpeFallbackExecutor returns a HelmExecutor whose `helm template` yields no
	// images, so generateContainerSBOMs falls back to the CPE image list. These
	t.Run("empty image list - no syft scan, no error", func(t *testing.T) {
		execRunner := &mock.ExecMockRunner{}
		config := helmBuildOptions{CreateBOM: true, SyftDownloadURL: "http://test-syft-url.io"}

		err := generateContainerSBOMs(config, nil, execRunner, &mock.FilesMock{}, &mock.HttpClientMock{})

		assert.NoError(t, err)
		assert.False(t, syftScanInvoked(execRunner), "syft must not run when there are no images")
	})

	t.Run("single image - one syft scan with derived registry and bom-docker-0.xml", func(t *testing.T) {
		execRunner := &mock.ExecMockRunner{}
		client := setupSyftDownloadMock(t)
		config := helmBuildOptions{CreateBOM: true, SyftDownloadURL: "http://test-syft-url.io"}
		images := []string{"registry.example.com/foo:1.0.0"}
		defer os.Remove("bom-docker-0.xml")

		err := generateContainerSBOMs(config, images, execRunner, &mock.FilesMock{}, client)

		assert.NoError(t, err)
		scanCalls := syftScanCalls(execRunner)
		require.Len(t, scanCalls, 1, "one syft scan per image")
		assert.Contains(t, scanCalls[0].Params, "registry:registry.example.com/foo:1.0.0",
			"registry must be derived from the full image reference")
		assert.Contains(t, scanCalls[0].Params, "cyclonedx-xml@1.4=bom-docker-0.xml")
	})

	t.Run("multiple images - one scan each with indexed bom filenames", func(t *testing.T) {
		execRunner := &mock.ExecMockRunner{}
		client := setupSyftDownloadMock(t)
		config := helmBuildOptions{CreateBOM: true, SyftDownloadURL: "http://test-syft-url.io"}
		images := []string{"registry.example.com/foo:1.0.0", "registry.example.com/bar:2.0.0"}
		defer os.Remove("bom-docker-0.xml")
		defer os.Remove("bom-docker-1.xml")

		err := generateContainerSBOMs(config, images, execRunner, &mock.FilesMock{}, client)

		assert.NoError(t, err)
		scanCalls := syftScanCalls(execRunner)
		require.Len(t, scanCalls, 2, "one syft scan per image")
		assert.Contains(t, scanCalls[0].Params, "registry:registry.example.com/foo:1.0.0")
		assert.Contains(t, scanCalls[0].Params, "cyclonedx-xml@1.4=bom-docker-0.xml")
		assert.Contains(t, scanCalls[1].Params, "registry:registry.example.com/bar:2.0.0")
		assert.Contains(t, scanCalls[1].Params, "cyclonedx-xml@1.4=bom-docker-1.xml")
	})

	t.Run("syft scan fails - error is surfaced", func(t *testing.T) {
		execRunner := &mock.ExecMockRunner{
			ShouldFailOnCommand: map[string]error{".*scan.*": fmt.Errorf("syft boom")},
		}
		client := setupSyftDownloadMock(t)
		config := helmBuildOptions{CreateBOM: true, SyftDownloadURL: "http://test-syft-url.io"}
		images := []string{"registry.example.com/foo:1.0.0"}

		err := generateContainerSBOMs(config, images, execRunner, &mock.FilesMock{}, client)

		require.Error(t, err, "a failing syft scan must surface as an error")
		assert.Contains(t, err.Error(), "syft boom", "the underlying syft error must be propagated")
	})

	t.Run("unparseable image reference - fails to derive registry, no scan", func(t *testing.T) {
		execRunner := &mock.ExecMockRunner{}
		config := helmBuildOptions{CreateBOM: true, SyftDownloadURL: "http://test-syft-url.io"}
		images := []string{"bad name"}

		err := generateContainerSBOMs(config, images, execRunner, &mock.FilesMock{}, &mock.HttpClientMock{})

		require.Error(t, err, "an unparseable image reference must surface as an error")
		assert.Contains(t, err.Error(), "failed to derive registry from image",
			"the error must come from the registry-derivation step")
		assert.False(t, syftScanInvoked(execRunner), "syft must not run when the registry cannot be derived")
	})

	t.Run("unparseable later image - fails to parse image name/tag, no scan", func(t *testing.T) {
		// The first image derives the registry fine; a later unparseable image
		// must surface from the name/tag parsing loop, not from registry derivation.
		execRunner := &mock.ExecMockRunner{}
		config := helmBuildOptions{CreateBOM: true, SyftDownloadURL: "http://test-syft-url.io"}
		images := []string{"registry.example.com/foo:1.0.0", "bad name"}

		err := generateContainerSBOMs(config, images, execRunner, &mock.FilesMock{}, &mock.HttpClientMock{})

		require.Error(t, err, "an unparseable later image must surface as an error")
		assert.Contains(t, err.Error(), "failed to parse image",
			"the error must come from the image name/tag parsing step")
		assert.False(t, syftScanInvoked(execRunner), "syft must not run when an image cannot be parsed")
	})
}

func TestRunHelmSBOMDiscoverImages(t *testing.T) {
	const validManifest = "apiVersion: apps/v1\nkind: Deployment\nspec:\n  template:\n    spec:\n      containers:\n        - image: registry.example.com/chart-app:9.9.9\n"

	tests := []struct {
		name           string
		templateReturn []byte
		templateErr    error
		cpeImages      []string
		expected       []string
	}{
		{
			name:        "helm template fails - falls back to CPE list",
			templateErr: fmt.Errorf("template boom"),
			cpeImages:   []string{"registry.example.com/cpe:1.0.0"},
			expected:    []string{"registry.example.com/cpe:1.0.0"},
		},
		{
			name:           "manifest parsing fails - falls back to CPE list",
			templateReturn: []byte("foo: [unclosed"),
			cpeImages:      []string{"registry.example.com/cpe:1.0.0"},
			expected:       []string{"registry.example.com/cpe:1.0.0"},
		},
		{
			name:           "no images rendered - falls back to CPE list",
			templateReturn: []byte("apiVersion: v1\nkind: ConfigMap\ndata:\n  key: value\n"),
			cpeImages:      []string{"registry.example.com/cpe:1.0.0"},
			expected:       []string{"registry.example.com/cpe:1.0.0"},
		},
		{
			name:           "templated images are returned",
			templateReturn: []byte(validManifest),
			cpeImages:      nil,
			expected:       []string{"registry.example.com/chart-app:9.9.9"},
		},
		{
			name:           "templated images take precedence over a non-empty CPE list",
			templateReturn: []byte(validManifest),
			cpeImages:      []string{"registry.example.com/stale:0.0.0"},
			expected:       []string{"registry.example.com/chart-app:9.9.9"},
		},
		{
			name:        "fallback with an empty CPE list yields no images",
			templateErr: fmt.Errorf("template boom"),
			cpeImages:   nil,
			expected:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmTemplate").Return(test.templateReturn, test.templateErr)
			config := helmBuildOptions{ContainerImageNameTags: test.cpeImages}

			images := discoverImages(config, helmExecutor)

			assert.Equal(t, test.expected, images)
		})
	}
}

func TestRunHelmSBOMFailureIsBestEffort(t *testing.T) {
	setupConfigOpenFileMock(t)
	// SBOM generation is best-effort: a failure inside generateContainerSBOMs
	// (here, an unparseable image that fails registry derivation) must be logged
	// and swallowed — the publish step itself must still succeed. Covers the
	// `if err := generateContainerSBOMs(...); err != nil` guards in both the
	// explicit "publish" command path and the default (publish=true) flow.
	tests := []struct {
		name   string
		config helmBuildOptions
	}{
		{
			name: "explicit publish command path",
			config: helmBuildOptions{
				HelmCommand:            "publish",
				CreateBOM:              true,
				SyftDownloadURL:        "http://test-syft-url.io",
				ContainerImageNameTags: []string{"bad name"},
			},
		},
		{
			name: "default flow with publish=true",
			config: helmBuildOptions{
				HelmCommand:            "",
				Publish:                true,
				CreateBOM:              true,
				SyftDownloadURL:        "http://test-syft-url.io",
				ContainerImageNameTags: []string{"bad name"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			utils := newHelmMockUtilsBundle()
			execRunner := &mock.ExecMockRunner{}
			cpe := helmBuildCommonPipelineEnvironment{}
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmLint").Return(nil)
			helmExecutor.On("RunHelmPublish").Return("https://my.target.repository/foo-1.0.0.tgz", nil)
			// helm template yields no images → SBOM falls back to the (bad) CPE list.
			helmExecutor.On("RunHelmTemplate").Return([]byte(nil), nil)

			err := runHelmBuild(test.config, helmExecutor, utils, &cpe, versioning.Coordinates{}, execRunner, &mock.FilesMock{}, &mock.HttpClientMock{})

			assert.NoError(t, err, "SBOM generation failure must not fail the publish step")
			assert.False(t, syftScanInvoked(execRunner), "syft must not run when the registry cannot be derived")
			assert.Equal(t, "https://my.target.repository/foo-1.0.0.tgz", cpe.custom.helmChartURL,
				"the chart must still be published despite the SBOM failure")
		})
	}
}

func TestRunHelmChartSBOM(t *testing.T) {
	setupConfigOpenFileMock(t)
	// The publish flow must produce the chart-level bom-helm.xml (in addition to
	// the container-image bom-docker-*.xml). Covers both publish entry points.
	tests := []struct {
		name   string
		config helmBuildOptions
	}{
		{
			name: "explicit publish command path",
			config: helmBuildOptions{
				HelmCommand:     "publish",
				ChartPath:       "chart",
				CreateBOM:       true,
				SyftDownloadURL: "http://test-syft-url.io",
			},
		},
		{
			name: "default flow with publish=true",
			config: helmBuildOptions{
				HelmCommand:     "",
				Publish:         true,
				ChartPath:       "chart",
				CreateBOM:       true,
				SyftDownloadURL: "http://test-syft-url.io",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			utils := newHelmMockUtilsBundle()
			utils.AddFile("chart/Chart.yaml", []byte("apiVersion: v2\nname: foo\nversion: 1.2.3\n"))
			execRunner := &mock.ExecMockRunner{}
			cpe := helmBuildCommonPipelineEnvironment{}
			helmExecutor := &mocks.HelmExecutor{}
			helmExecutor.On("RunHelmLint").Return(nil)
			helmExecutor.On("RunHelmPublish").Return("https://my.target.repository/foo-1.0.0.tgz", nil)
			helmExecutor.On("RunHelmTemplate").Return([]byte(nil), nil)

			err := runHelmBuild(test.config, helmExecutor, utils, &cpe, versioning.Coordinates{}, execRunner, utils.FilesMock, utils.HttpClientMock)

			require.NoError(t, err)
			content, err := utils.FileRead("bom-helm.xml")
			require.NoError(t, err, "bom-helm.xml must be produced on publish")
			assert.Contains(t, string(content), "pkg:helm/foo@1.2.3")
		})
	}
}

func TestRunHelmPopulatesBuildSettingsInfo(t *testing.T) {
	setupConfigOpenFileMock(t)
	// SLC-29: helmExecute must record build-settings traceability into
	// commonPipelineEnvironment.custom.buildSettingsInfo (non-empty JSON).
	t.Run("populates valid JSON on success", func(t *testing.T) {
		utils := newHelmMockUtilsBundle()
		cpe := helmBuildCommonPipelineEnvironment{}
		cfg := helmBuildOptions{
			HelmCommand: "lint",
			CreateBOM:   true,
			Publish:     false,
		}
		helmExecutor := &mocks.HelmExecutor{}
		helmExecutor.On("RunHelmLint").Return(nil)

		err := runHelmBuild(cfg, helmExecutor, utils, &cpe, versioning.Coordinates{}, utils.ExecMockRunner, utils.FilesMock, utils.HttpClientMock)

		require.NoError(t, err)
		require.NotEmpty(t, cpe.custom.buildSettingsInfo, "buildSettingsInfo must be populated")
		var payload map[string]interface{}
		assert.NoError(t, json.Unmarshal([]byte(cpe.custom.buildSettingsInfo), &payload),
			"buildSettingsInfo must be valid JSON")
	})

	t.Run("build settings failure is best-effort - step still succeeds", func(t *testing.T) {
		// A malformed inbound BuildSettingsInfo makes CreateBuildSettingsInfo
		// fail (json.Unmarshal of the existing value). The step must swallow it:
		// no error returned, and buildSettingsInfo left empty.
		utils := newHelmMockUtilsBundle()
		cpe := helmBuildCommonPipelineEnvironment{}
		cfg := helmBuildOptions{
			HelmCommand:       "lint",
			BuildSettingsInfo: "{not-valid-json",
		}
		helmExecutor := &mocks.HelmExecutor{}
		helmExecutor.On("RunHelmLint").Return(nil)

		err := runHelmBuild(cfg, helmExecutor, utils, &cpe, versioning.Coordinates{}, utils.ExecMockRunner, utils.FilesMock, utils.HttpClientMock)

		assert.NoError(t, err, "a build-settings failure must not fail the step")
		assert.Empty(t, cpe.custom.buildSettingsInfo, "buildSettingsInfo must stay empty on failure")
	})
}

func TestParseAndRenderCPETemplate(t *testing.T) {
	commonPipelineEnvironment := "commonPipelineEnvironment"
	valuesYaml := []byte(`
image: "image_1"
tag: {{ cpe "artifactVersion" }}
`)
	values1Yaml := []byte(`
image: "image_2"
tag: {{ cpe "artVersion" }}
`)
	values3Yaml := []byte(`
image: "image_3"
tag: {{ .CPE.artVersion
`)
	values4Yaml := []byte(`
image: "test-image"
tag: {{ imageTag "test-image" }}
`)

	tmpDir := t.TempDir()
	require.DirExists(t, tmpDir)
	err := os.Mkdir(path.Join(tmpDir, commonPipelineEnvironment), 0700)
	require.NoError(t, err)
	cpe := piperenv.CPEMap{
		"artifactVersion":         "1.0.0-123456789",
		"container/imageNameTags": []string{"test-image:1.0.0-123456789"},
	}
	err = cpe.WriteToDisk(tmpDir)
	require.NoError(t, err)

	defaultValueFile := "values.yaml"
	config := helmBuildOptions{
		ChartPath: ".",
	}

	tt := []struct {
		name             string
		defaultValueFile string
		config           helmBuildOptions
		expectedErr      error
		valueFile        []byte
	}{
		{
			name:             "'artifactVersion' file exists in CPE",
			defaultValueFile: defaultValueFile,
			config:           config,
			expectedErr:      nil,
			valueFile:        valuesYaml,
		},
		{
			name:             "'artVersion' file does not exist in CPE",
			defaultValueFile: defaultValueFile,
			config:           config,
			expectedErr:      nil,
			valueFile:        values1Yaml,
		},
		{
			name:             "Good template ({{ imageTag 'test-image' }})",
			defaultValueFile: defaultValueFile,
			config:           config,
			expectedErr:      nil,
			valueFile:        values4Yaml,
		},
		{
			name:             "Wrong template ({{ .CPE.artVersion)",
			defaultValueFile: defaultValueFile,
			config:           config,
			expectedErr:      fmt.Errorf("failed to parse template: failed to parse cpe template '\nimage: \"image_3\"\ntag: {{ .CPE.artVersion\n': template: cpetemplate:4: unclosed action started at cpetemplate:3"),
			valueFile:        values3Yaml,
		},
		{
			name:             "Multiple value files",
			defaultValueFile: defaultValueFile,
			config: helmBuildOptions{
				ChartPath:  ".",
				HelmValues: []string{"./values_1.yaml", "./values_2.yaml"},
			},
			expectedErr: nil,
			valueFile:   valuesYaml,
		},
		{
			name:             "No value file is provided",
			defaultValueFile: "",
			config: helmBuildOptions{
				ChartPath:  ".",
				HelmValues: []string{},
			},
			expectedErr: fmt.Errorf("no value file to proccess, please provide value file(s)"),
			valueFile:   valuesYaml,
		},
		{
			name:             "Wrong path to value file",
			defaultValueFile: defaultValueFile,
			config: helmBuildOptions{
				ChartPath:  ".",
				HelmValues: []string{"wrong/path/to/values_1.yaml"},
			},
			expectedErr: fmt.Errorf("failed to read file: could not read 'wrong/path/to/values_1.yaml'"),
			valueFile:   valuesYaml,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			utils := newHelmMockUtilsBundle()
			utils.AddFile(fmt.Sprintf("%s/%s", config.ChartPath, test.defaultValueFile), test.valueFile)

			if len(test.config.HelmValues) == 2 {
				for _, value := range test.config.HelmValues {
					utils.AddFile(value, test.valueFile)
				}
			}

			err := parseAndRenderCPETemplate(test.config, tmpDir, utils)
			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type fileHandlerMock struct {
	fileExistsCalled []string
	fileReadCalled   []string
	fileWriteCalled  []string
}

func (f *fileHandlerMock) FileWrite(name string, content []byte, mode os.FileMode) error {
	f.fileWriteCalled = append(f.fileWriteCalled, name)
	return nil
}

func (f *fileHandlerMock) FileRead(name string) ([]byte, error) {
	f.fileReadCalled = append(f.fileReadCalled, name)
	return []byte{}, nil
}

func (f *fileHandlerMock) FileExists(name string) (bool, error) {
	f.fileExistsCalled = append(f.fileExistsCalled, name)
	return true, nil
}

func TestRunHelmBuildSettingsInfo(t *testing.T) {
	t.Parallel()
	setupConfigOpenFileMock(t)

	t.Run("buildSettingsInfo written to CPE after successful run", func(t *testing.T) {
		cpe := helmBuildCommonPipelineEnvironment{}
		helmExecutor := &mocks.HelmExecutor{}
		helmExecutor.On("RunHelmLint").Return(nil)
		helmExecutor.On("RunHelmPublish").Return("", nil)

		config := helmBuildOptions{
			HelmCommand: "",
			Publish:     true,
		}

		err := runHelmBuild(config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
		require.NoError(t, err)
		assert.NotEmpty(t, cpe.custom.buildSettingsInfo)
		assert.Contains(t, cpe.custom.buildSettingsInfo, "helmBuild")
	})

	t.Run("existing buildSettingsInfo is appended to", func(t *testing.T) {
		cpe := helmBuildCommonPipelineEnvironment{}
		helmExecutor := &mocks.HelmExecutor{}
		helmExecutor.On("RunHelmLint").Return(nil)
		helmExecutor.On("RunHelmPublish").Return("", nil)

		config := helmBuildOptions{
			HelmCommand:       "",
			BuildSettingsInfo: `{"mavenBuild":[{"dockerImage":"maven:3.8"}]}`,
		}

		err := runHelmBuild(config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
		require.NoError(t, err)
		assert.Contains(t, cpe.custom.buildSettingsInfo, "helmBuild")
		assert.Contains(t, cpe.custom.buildSettingsInfo, "mavenBuild")
	})

	t.Run("buildSettingsInfo written even for lint-only run", func(t *testing.T) {
		cpe := helmBuildCommonPipelineEnvironment{}
		helmExecutor := &mocks.HelmExecutor{}
		helmExecutor.On("RunHelmLint").Return(nil)

		config := helmBuildOptions{
			HelmCommand: "lint",
		}

		err := runHelmBuild(config, helmExecutor, &fileHandlerMock{}, &cpe, versioning.Coordinates{}, newHelmMockUtilsBundle(), newHelmMockUtilsBundle(), newHelmMockUtilsBundle())
		require.NoError(t, err)
		assert.NotEmpty(t, cpe.custom.buildSettingsInfo)
		assert.Contains(t, cpe.custom.buildSettingsInfo, "helmBuild")
	})
}
