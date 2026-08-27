//go:build unit
// +build unit

package versioning

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/stretchr/testify/assert"
)

type mavenMockRunner struct {
	evaluateErrorString         string
	evaluateMultipleErrorString string
	executeErrorString          string
	stdout                      string
	// values allows to return a dedicated value per expression, stdout is used if it is not set
	values                map[string]string
	opts                  *maven.EvaluateOptions
	execOpts              *maven.ExecuteOptions
	expression            string
	expressions           []string
	evaluateCalls         int
	evaluateMultipleCalls int
	executeCalls          int
}

func (m *mavenMockRunner) value(expression string) string {
	if m.values != nil {
		return m.values[expression]
	}
	return m.stdout
}

func (m *mavenMockRunner) Evaluate(opts *maven.EvaluateOptions, expression string, utils maven.Utils) (string, error) {
	m.opts = opts
	m.expression = expression
	m.evaluateCalls++
	if len(m.evaluateErrorString) > 0 {
		return "", fmt.Errorf("%s", m.evaluateErrorString)
	}
	return m.value(expression), nil
}

func (m *mavenMockRunner) EvaluateMultiple(opts *maven.EvaluateOptions, expressions []string, utils maven.Utils) ([]string, error) {
	m.opts = opts
	m.expressions = expressions
	m.evaluateMultipleCalls++
	if len(m.evaluateMultipleErrorString) > 0 {
		return nil, fmt.Errorf("%s", m.evaluateMultipleErrorString)
	}
	// a combined evaluation is a single mvn call, thus it fails as well if evaluation fails
	if len(m.evaluateErrorString) > 0 {
		return nil, fmt.Errorf("%s", m.evaluateErrorString)
	}
	values := []string{}
	for _, expression := range expressions {
		values = append(values, m.value(expression))
	}
	return values, nil
}

func (m *mavenMockRunner) Execute(opts *maven.ExecuteOptions, utils maven.Utils) (string, error) {
	m.execOpts = opts
	m.executeCalls++
	if len(m.executeErrorString) > 0 {
		return "", fmt.Errorf("%s", m.executeErrorString)
	}
	if opts.ReturnStdout {
		return m.stdout, nil
	}
	return "", nil
}

func mavenCoordinateValues() map[string]string {
	return map[string]string{
		"project.groupId":    "com.sap.cp",
		"project.artifactId": "my-app",
		"project.version":    "1.0.0-SNAPSHOT",
		"project.packaging":  "jar",
	}
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
		assert.Equal(t, mavenCoordinateExpressions, runner.expressions)
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

	t.Run("falls back to single evaluation", func(t *testing.T) {
		runner := mavenMockRunner{
			stdout:                      "1.2.3",
			evaluateMultipleErrorString: "combined evaluation failed",
		}
		mvn := &Maven{
			runner:  &runner,
			options: maven.EvaluateOptions{PomPath: "path/to/pom.xml"},
		}
		version, err := mvn.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
		assert.Equal(t, "project.version", runner.expression)
		assert.Equal(t, 1, runner.evaluateCalls)
	})
}

func TestMavenGetCoordinates(t *testing.T) {
	t.Run("reads all coordinates with one maven call", func(t *testing.T) {
		runner := mavenMockRunner{values: mavenCoordinateValues()}
		mvn := &Maven{
			runner:  &runner,
			options: maven.EvaluateOptions{PomPath: "path/to/pom.xml"},
		}

		coordinates, err := mvn.GetCoordinates()

		assert.NoError(t, err)
		assert.Equal(t, Coordinates{GroupID: "com.sap.cp", ArtifactID: "my-app", Version: "1.0.0-SNAPSHOT", Packaging: "jar"}, coordinates)
		assert.Equal(t, 1, runner.evaluateMultipleCalls)
		assert.Equal(t, 0, runner.evaluateCalls)
	})

	t.Run("caches coordinates", func(t *testing.T) {
		runner := mavenMockRunner{values: mavenCoordinateValues()}
		mvn := &Maven{runner: &runner}

		first, err := mvn.GetCoordinates()
		assert.NoError(t, err)
		second, err := mvn.GetCoordinates()
		assert.NoError(t, err)

		assert.Equal(t, first, second)
		assert.Equal(t, 1, runner.evaluateMultipleCalls)
		assert.Equal(t, 0, runner.evaluateCalls)
	})

	t.Run("falls back to single evaluations", func(t *testing.T) {
		runner := mavenMockRunner{
			values:                      mavenCoordinateValues(),
			evaluateMultipleErrorString: "combined evaluation failed",
		}
		mvn := &Maven{runner: &runner}

		coordinates, err := mvn.GetCoordinates()

		assert.NoError(t, err)
		assert.Equal(t, Coordinates{GroupID: "com.sap.cp", ArtifactID: "my-app", Version: "1.0.0-SNAPSHOT", Packaging: "jar"}, coordinates)
		assert.Equal(t, 4, runner.evaluateCalls)
		// a failing combined evaluation must not be repeated for every single expression
		assert.Equal(t, 1, runner.evaluateMultipleCalls)
	})

	t.Run("falls back if a coordinate is empty", func(t *testing.T) {
		values := mavenCoordinateValues()
		values["project.packaging"] = ""
		runner := mavenMockRunner{values: values}
		mvn := &Maven{runner: &runner}

		coordinates, err := mvn.GetCoordinates()

		assert.NoError(t, err)
		assert.Equal(t, "com.sap.cp", coordinates.GroupID)
		assert.Equal(t, "", coordinates.Packaging)
		assert.Equal(t, 4, runner.evaluateCalls)
	})

	t.Run("error case", func(t *testing.T) {
		runner := mavenMockRunner{evaluateErrorString: "maven eval failed"}
		mvn := &Maven{runner: &runner}

		_, err := mvn.GetCoordinates()

		assert.EqualError(t, err, "Maven - getting groupId failed: maven eval failed")
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

	t.Run("version is updated without another maven call", func(t *testing.T) {
		runner := mavenMockRunner{values: mavenCoordinateValues()}
		mvn := &Maven{
			runner:  &runner,
			options: maven.EvaluateOptions{PomPath: "path/to/pom.xml"},
		}

		// this is the sequence artifactPrepareVersion runs for a maven project
		version, err := mvn.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.0.0-SNAPSHOT", version)

		assert.NoError(t, mvn.SetVersion("1.0.0-20260827_deadbeef"))

		coordinates, err := mvn.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, Coordinates{
			GroupID:    "com.sap.cp",
			ArtifactID: "my-app",
			Version:    "1.0.0-20260827_deadbeef",
			Packaging:  "jar",
		}, coordinates)

		// one evaluation and the versions:set call, instead of 7 maven calls
		assert.Equal(t, 1, runner.evaluateMultipleCalls)
		assert.Equal(t, 0, runner.evaluateCalls)
		assert.Equal(t, 1, runner.executeCalls)
	})
}
