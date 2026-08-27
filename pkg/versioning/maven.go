package versioning

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
)

type mavenRunner interface {
	Execute(*maven.ExecuteOptions, maven.Utils) (string, error)
	Evaluate(*maven.EvaluateOptions, string, maven.Utils) (string, error)
	EvaluateMultiple(*maven.EvaluateOptions, []string, maven.Utils) ([]string, error)
}

const (
	mavenGroupIDExpression    = "project.groupId"
	mavenArtifactIDExpression = "project.artifactId"
	mavenVersionExpression    = "project.version"
	mavenPackagingExpression  = "project.packaging"
)

// mavenCoordinateExpressions are the expressions which are read with one single mvn call,
// see Maven.evaluateCoordinates()
var mavenCoordinateExpressions = []string{
	mavenGroupIDExpression,
	mavenArtifactIDExpression,
	mavenVersionExpression,
	mavenPackagingExpression,
}

// Maven defines a maven artifact used for versioning
type Maven struct {
	options maven.EvaluateOptions
	runner  mavenRunner
	utils   maven.Utils
	// evaluated holds the expression values matching the current state of the build descriptor.
	// Reading a single property requires a complete maven run, therefore values are read only
	// once and are kept in sync with the descriptor by SetVersion().
	evaluated map[string]string
	// coordinatesEvaluated is true as soon as it has been tried to read all coordinates with one
	// single mvn call. It makes sure that a failing combined evaluation is not repeated for every
	// single expression.
	coordinatesEvaluated bool
}

func (m *Maven) init() {
	if len(m.options.PomPath) == 0 {
		m.options.PomPath = "pom.xml"
	}

	if m.utils == nil {
		m.utils = maven.NewUtilsBundle()
	}

	if m.evaluated == nil {
		m.evaluated = map[string]string{}
	}
}

// evaluate returns the value of the given maven expression.
// Since every evaluation starts a JVM and builds the maven project model, which is expensive
// especially for multi module projects with a cold maven repository, the coordinates of the
// artifact are read with one single mvn call and are cached for subsequent calls.
func (m *Maven) evaluate(expression string) (string, error) {
	m.init()

	if value, exists := m.evaluated[expression]; exists {
		return value, nil
	}

	if !m.coordinatesEvaluated && isMavenCoordinateExpression(expression) {
		m.coordinatesEvaluated = true
		if err := m.evaluateCoordinates(); err != nil {
			// reading all coordinates at once is an optimization only, thus fall back to
			// evaluating the single expression in order to keep the previous behavior
			log.Entry().WithError(err).Debug("Maven - reading coordinates with one call failed, evaluating single expressions")
		} else if value, exists := m.evaluated[expression]; exists {
			return value, nil
		}
	}

	value, err := m.runner.Evaluate(&m.options, expression, m.utils)
	if err != nil {
		return "", err
	}
	m.evaluated[expression] = value
	return value, nil
}

// evaluateCoordinates reads all coordinates of the artifact with one single mvn call.
func (m *Maven) evaluateCoordinates() error {
	values, err := m.runner.EvaluateMultiple(&m.options, mavenCoordinateExpressions, m.utils)
	if err != nil {
		return err
	}

	if len(values) != len(mavenCoordinateExpressions) {
		return fmt.Errorf("expected %v values for file '%s', got %v", len(mavenCoordinateExpressions), m.options.PomPath, len(values))
	}

	for i, expression := range mavenCoordinateExpressions {
		if len(values[i]) == 0 {
			return fmt.Errorf("expression '%s' in file '%s' resolved to an empty value", expression, m.options.PomPath)
		}
	}

	for i, expression := range mavenCoordinateExpressions {
		m.evaluated[expression] = values[i]
	}
	return nil
}

func isMavenCoordinateExpression(expression string) bool {
	for _, coordinateExpression := range mavenCoordinateExpressions {
		if coordinateExpression == expression {
			return true
		}
	}
	return false
}

// VersioningScheme returns the relevant versioning scheme
func (m *Maven) VersioningScheme() string {
	return "maven"
}

// GetCoordinates reads the coordinates from the maven pom.xml descriptor file
func (m *Maven) GetCoordinates() (Coordinates, error) {
	result := Coordinates{}
	var err error
	result.GroupID, err = m.GetGroupID()
	if err != nil {
		return result, err
	}
	result.ArtifactID, err = m.GetArtifactID()
	if err != nil {
		return result, err
	}
	result.Version, err = m.GetVersion()
	if err != nil {
		return result, err
	}
	result.Packaging, err = m.GetPackaging()
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetPackaging returns the current ID of the Group
func (m *Maven) GetPackaging() (string, error) {
	packaging, err := m.evaluate(mavenPackagingExpression)
	if err != nil {
		return "", fmt.Errorf("Maven - getting packaging failed: %w", err)
	}
	return packaging, nil
}

// GetGroupID returns the current ID of the Group
func (m *Maven) GetGroupID() (string, error) {
	groupID, err := m.evaluate(mavenGroupIDExpression)
	if err != nil {
		return "", fmt.Errorf("Maven - getting groupId failed: %w", err)
	}
	return groupID, nil
}

// GetArtifactID returns the current ID of the artifact
func (m *Maven) GetArtifactID() (string, error) {
	artifactID, err := m.evaluate(mavenArtifactIDExpression)
	if err != nil {
		return "", fmt.Errorf("Maven - getting artifactId failed: %w", err)
	}
	return artifactID, nil
}

// GetVersion returns the current version of the artifact
func (m *Maven) GetVersion() (string, error) {
	version, err := m.evaluate(mavenVersionExpression)
	if err != nil {
		return "", fmt.Errorf("Maven - getting version failed: %w", err)
	}
	//ToDo: how to deal with SNAPSHOT replacement?
	return version, nil
}

// SetVersion updates the version of the artifact
func (m *Maven) SetVersion(version string) error {
	m.init()

	groupID, err := m.evaluate(mavenGroupIDExpression)
	if err != nil {
		return fmt.Errorf("Maven - getting groupId failed: %w", err)
	}
	opts := maven.ExecuteOptions{
		PomPath:             m.options.PomPath,
		ProjectSettingsFile: m.options.ProjectSettingsFile,
		GlobalSettingsFile:  m.options.GlobalSettingsFile,
		M2Path:              m.options.M2Path,
		Goals:               []string{"org.codehaus.mojo:versions-maven-plugin:2.7:set"},
		Defines: []string{
			fmt.Sprintf("-DnewVersion=%v", version),
			fmt.Sprintf("-DgroupId=%v", groupID),
			"-DartifactId=*",
			"-DoldVersion=*",
			"-DgenerateBackupPoms=false",
		},
	}
	_, err = m.runner.Execute(&opts, m.utils)
	if err != nil {
		return fmt.Errorf("Maven - setting version %v failed: %w", version, err)
	}
	// versions:set changes versions only, the remaining coordinates of the descriptor are untouched
	m.evaluated[mavenVersionExpression] = version
	return nil
}
