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
	opts                *maven.EvaluateOptions
	execOpts            *maven.ExecuteOptions
	expression          string
}

func (m *mavenMockRunner) Evaluate(opts *maven.EvaluateOptions, expression string, utils maven.Utils) (string, error) {
	m.opts = opts
	m.expression = expression
	if len(m.evaluateErrorString) > 0 {
		return "", fmt.Errorf("%s", m.evaluateErrorString)
	}
	return m.stdout, nil
}

func (m *mavenMockRunner) Execute(opts *maven.ExecuteOptions, utils maven.Utils) (string, error) {
	m.execOpts = opts
	if len(m.executeErrorString) > 0 {
		return "", fmt.Errorf("%s", m.executeErrorString)
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
			options: maven.EvaluateOptions{PomPath: "path/to/pom.xml", M2Path: "path/to/m2"},
		}
		version, err := mvn.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
		assert.Equal(t, "project.version", runner.expression)
		assert.Equal(t, "path/to/pom.xml", runner.opts.PomPath)
		assert.Equal(t, "path/to/m2", runner.opts.M2Path)
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
			runner: &runner,
			options: maven.EvaluateOptions{
				PomPath:             "path/to/pom.xml",
				ProjectSettingsFile: "project-settings.xml",
				GlobalSettingsFile:  "global-settings.xml",
				M2Path:              "m2/path"},
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
		assert.Equal(t, &expectedOptions, runner.execOpts)
	})

	t.Run("evaluate error", func(t *testing.T) {
		runner := mavenMockRunner{
			stdout:              "testGroup",
			evaluateErrorString: "maven eval failed",
		}
		mvn := &Maven{
			runner:  &runner,
			options: maven.EvaluateOptions{PomPath: "path/to/pom.xml"},
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
			options: maven.EvaluateOptions{PomPath: "path/to/pom.xml"},
		}
		err := mvn.SetVersion("1.2.4")
		assert.EqualError(t, err, "Maven - setting version 1.2.4 failed: maven exec failed")
	})
}
