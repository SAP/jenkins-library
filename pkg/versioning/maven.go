package versioning

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/maven"

	"github.com/pkg/errors"
)

type mavenExecRunner interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

type mavenRunner interface {
	Execute(*maven.ExecuteOptions, maven.Utils) (string, error)
	Evaluate(*maven.EvaluateOptions, string, maven.Utils) (string, error)
}

// Maven defines a maven artifact used for versioning
type Maven struct {
	options maven.EvaluateOptions
	runner  mavenRunner
	utils   maven.Utils
}

func (m *Maven) init() {
	if len(m.options.PomPath) == 0 {
		m.options.PomPath = "pom.xml"
	}

	if m.utils == nil {
		m.utils = maven.NewUtilsBundle()
	}
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
	m.init()

	packaging, err := m.runner.Evaluate(&m.options, "project.packaging", m.utils)
	if err != nil {
		return "", errors.Wrap(err, "Maven - getting packaging failed")
	}
	return packaging, nil
}

// GetGroupID returns the current ID of the Group
func (m *Maven) GetGroupID() (string, error) {
	m.init()

	groupID, err := m.runner.Evaluate(&m.options, "project.groupId", m.utils)
	if err != nil {
		return "", errors.Wrap(err, "Maven - getting groupId failed")
	}
	return groupID, nil
}

// GetArtifactID returns the current ID of the artifact
func (m *Maven) GetArtifactID() (string, error) {
	m.init()

	artifactID, err := m.runner.Evaluate(&m.options, "project.artifactId", m.utils)
	if err != nil {
		return "", errors.Wrap(err, "Maven - getting artifactId failed")
	}
	return artifactID, nil
}

// GetVersion returns the current version of the artifact
func (m *Maven) GetVersion() (string, error) {
	m.init()

	version, err := m.runner.Evaluate(&m.options, "project.version", m.utils)
	if err != nil {
		return "", errors.Wrap(err, "Maven - getting version failed")
	}
	//ToDo: how to deal with SNAPSHOT replacement?
	return version, nil
}

// SetVersion updates the version of the artifact
func (m *Maven) SetVersion(version string) error {
	m.init()

	groupID, err := m.runner.Evaluate(&m.options, "project.groupId", m.utils)
	if err != nil {
		return errors.Wrap(err, "Maven - getting groupId failed")
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
		return errors.Wrapf(err, "Maven - setting version %v failed", version)
	}
	return nil
}
