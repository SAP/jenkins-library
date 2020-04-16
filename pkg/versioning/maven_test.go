package versioning

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/stretchr/testify/assert"
)

type mavenMockRunner struct {
	evaluateErrorString string
	executeErrorString  string
	stdout              string
	opts                *maven.ExecuteOptions
	expression          string
	pomFile             string
}

func (m *mavenMockRunner) Evaluate(pomFile, expression string, runner mavenExecRunner) (string, error) {
	m.pomFile = pomFile
	m.expression = expression
	if len(m.evaluateErrorString) > 0 {
		return "", fmt.Errorf(m.evaluateErrorString)
	}
	return m.stdout, nil
}

func (m *mavenMockRunner) Execute(opts *maven.ExecuteOptions, runner mavenExecRunner) (string, error) {
	m.opts = opts
	if len(m.executeErrorString) > 0 {
		return "", fmt.Errorf(m.executeErrorString)
	}
	if opts.ReturnStdout {
		return m.stdout, nil
	}
	return "", nil
}

func TestMavenGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		runner := mavenMockRunner{
			stdout: "1.2.3",
		}
		mvn := &Maven{
			runner:  &runner,
			pomPath: "path/to/pom.xml",
		}
		version, err := mvn.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
		assert.Equal(t, "project.version", runner.expression)
		assert.Equal(t, "path/to/pom.xml", runner.pomFile)
	})

	t.Run("error case", func(t *testing.T) {
		runner := mavenMockRunner{
			stdout:              "1.2.3",
			evaluateErrorString: "maven eval failed",
		}
		mvn := &Maven{
			runner: &runner,
		}
		version, err := mvn.GetVersion()
		assert.EqualError(t, err, "Maven - getting version failed: maven eval failed")
		assert.Equal(t, "", version)
	})

}

func TestMavenSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		runner := mavenMockRunner{
			stdout: "testGroup",
		}
		mvn := &Maven{
			runner:              &runner,
			pomPath:             "path/to/pom.xml",
			projectSettingsFile: "project-settings.xml",
			globalSettingsFile:  "global-settings.xml",
			m2Path:              "m2/path",
		}
		expectedOptions := maven.ExecuteOptions{
			PomPath:             "path/to/pom.xml",
			Defines:             []string{"-DnewVersion=1.2.4", "-DgroupId=testGroup", "-DartifactId=*", "-DoldVersion=*", "-DgenerateBackupPoms=false"},
			Goals:               []string{"org.codehaus.mojo:versions-maven-plugin:2.7:set"},
			ProjectSettingsFile: "project-settings.xml",
			GlobalSettingsFile:  "global-settings.xml",
			M2Path:              "m2/path",
		}
		err := mvn.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Equal(t, &expectedOptions, runner.opts)
	})

	t.Run("evaluate error", func(t *testing.T) {
		runner := mavenMockRunner{
			stdout:              "testGroup",
			evaluateErrorString: "maven eval failed",
		}
		mvn := &Maven{
			runner:  &runner,
			pomPath: "path/to/pom.xml",
		}
		err := mvn.SetVersion("1.2.4")
		assert.EqualError(t, err, "Maven - getting groupId failed: maven eval failed")
	})

	t.Run("execute error", func(t *testing.T) {
		runner := mavenMockRunner{
			stdout:             "testGroup",
			executeErrorString: "maven exec failed",
		}
		mvn := &Maven{
			runner:  &runner,
			pomPath: "path/to/pom.xml",
		}
		err := mvn.SetVersion("1.2.4")
		assert.EqualError(t, err, "Maven - setting version 1.2.4 failed: maven exec failed")
	})
}
