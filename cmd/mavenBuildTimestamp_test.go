//go:build unit

package cmd

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const missingTimestampBOM = `<bom xmlns="http://cyclonedx.org/schema/bom/1.4"><metadata><component bom-ref="root"/></metadata></bom>`

type mavenTimestampMockUtils struct {
	mavenMockUtils
	globError   error
	beforeWrite func()
}

func (u *mavenTimestampMockUtils) Glob(pattern string) ([]string, error) {
	if u.globError != nil {
		return nil, u.globError
	}
	return u.FilesMock.Glob(pattern)
}

func (u *mavenTimestampMockUtils) FileWrite(path string, content []byte, mode os.FileMode) error {
	if u.beforeWrite != nil {
		u.beforeWrite()
	}
	return u.FilesMock.FileWrite(path, content, mode)
}

func TestEnsureMavenBOMTimestamps(t *testing.T) {
	for _, pomPath := range []string{"", "./pom.xml", "project/custom-pom.xml"} {
		t.Run("repairs both outputs recursively "+pomPath, func(t *testing.T) {
			utils := newMavenMockUtils()
			root := ""
			if strings.HasPrefix(pomPath, "project/") {
				root = "project/"
			}
			paths := []string{root + "target/bom-maven.xml", root + "target/simple-bom-maven.xml", root + "srv/custom-output/bom-maven.xml", root + "srv/custom-output/simple-bom-maven.xml"}
			for _, path := range paths {
				utils.AddFile(path, []byte(missingTimestampBOM))
			}
			utils.AddFile("unrelated/bom.xml", []byte("do not parse"))
			if root != "" {
				utils.AddFile("other/target/bom-maven.xml", []byte("outside POM subtree"))
			}
			require.NoError(t, ensureMavenBOMTimestamps(pomPath, &utils))
			var repaired []byte
			for _, path := range paths {
				content, err := utils.FileRead(path)
				require.NoError(t, err)
				assert.True(t, utils.HasWrittenFile(path))
				assert.Contains(t, string(content), "<timestamp>")
				if repaired != nil {
					assert.Equal(t, repaired, content, "all BOMs use one timestamp")
				}
				repaired = content
			}
			assert.False(t, utils.HasWrittenFile("unrelated/bom.xml"))
			assert.False(t, utils.HasWrittenFile("other/target/bom-maven.xml"))
		})
	}

	t.Run("does not write an existing timestamp", func(t *testing.T) {
		utils := newMavenMockUtils()
		input := []byte(`<bom xmlns="http://cyclonedx.org/schema/bom/1.4"><metadata><timestamp>2026-01-01T00:00:00Z</timestamp></metadata></bom>`)
		utils.AddFile("target/bom-maven.xml", input)
		require.NoError(t, ensureMavenBOMTimestamps("pom.xml", &utils))
		assert.False(t, utils.HasWrittenFile("target/bom-maven.xml"))
		actual, err := utils.FileRead("target/bom-maven.xml")
		require.NoError(t, err)
		assert.Equal(t, input, actual)
	})

	for _, failure := range []string{"glob", "read", "write", "invalid XML", "no files"} {
		t.Run(failure, func(t *testing.T) {
			utils := mavenTimestampMockUtils{mavenMockUtils: newMavenMockUtils()}
			path := "target/bom-maven.xml"
			utils.AddFile(path, []byte(missingTimestampBOM))
			cause := errors.New("injected failure")
			var want string
			switch failure {
			case "glob":
				utils.globError, want = cause, "find **/bom-maven.xml"
			case "read":
				utils.FileReadErrors = map[string]error{path: cause}
				want = "read " + path
			case "write":
				utils.FileWriteError, want = cause, "write "+path
			case "invalid XML":
				utils.AddFile(path, []byte("invalid"))
				want = "post-process " + path
			case "no files":
				utils.mavenMockUtils = newMavenMockUtils()
				want = "no Maven BOM files found"
			}
			err := ensureMavenBOMTimestamps("pom.xml", &utils)
			require.ErrorContains(t, err, want)
			assert.False(t, utils.HasWrittenFile(path))
			if failure == "glob" || failure == "read" || failure == "write" {
				assert.ErrorIs(t, err, cause)
			}
		})
	}
}

func TestMavenBuildEnsureBOMTimestamp(t *testing.T) {
	SetConfigOptions(ConfigCommandOptions{OpenFile: config.OpenPiperFile})
	for _, tc := range []struct {
		name              string
		createBOM, ensure bool
	}{
		{"default off", true, false},
		{"enabled", true, true},
		{"no BOM generation", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			utils := mavenTimestampMockUtils{mavenMockUtils: newMavenMockUtils()}
			paths := []string{"target/bom-maven.xml", "srv/target/simple-bom-maven.xml"}
			for _, path := range paths {
				utils.AddFile(path, []byte(missingTimestampBOM))
			}
			utils.beforeWrite = func() {
				require.Len(t, utils.Calls, 2, "repair runs after both Maven invocations")
				assert.Contains(t, utils.Calls[0].Params, mvnCycloneDXPackage+":makeAggregateBom")
				assert.Contains(t, utils.Calls[1].Params, mvnCycloneDXPackage+":makeBom")
			}
			options := mavenBuildOptions{CreateBOM: tc.createBOM, EnsureBOMTimestamp: tc.ensure}
			env := mavenBuildCommonPipelineEnvironment{}
			require.NoError(t, runMavenBuild(&options, nil, &utils, &env))
			for _, path := range paths {
				assert.Equal(t, tc.createBOM && tc.ensure, utils.HasWrittenFile(path))
			}
		})
	}

	t.Run("repair failure prevents publish", func(t *testing.T) {
		utils := newMavenMockUtils()
		utils.AddFile("target/bom-maven.xml", []byte("invalid XML"))
		options := mavenBuildOptions{CreateBOM: true, EnsureBOMTimestamp: true, Publish: true}
		env := mavenBuildCommonPipelineEnvironment{}
		err := runMavenBuild(&options, nil, &utils, &env)
		require.ErrorContains(t, err, "failed to ensure Maven BOM timestamps")
		require.Len(t, utils.Calls, 2)
		for _, call := range utils.Calls {
			assert.NotContains(t, call.Params, "deploy")
		}
		assert.False(t, utils.HasWrittenFile("target/bom-maven.xml"))
	})
}
