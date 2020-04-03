package versioning

import (
	"fmt"
	"io"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/maven"

	"github.com/pkg/errors"
)

type mavenExecRunner interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

type mavenRunner interface {
	Execute(*maven.ExecuteOptions, mavenExecRunner) (string, error)
	Evaluate(string, string, mavenExecRunner) (string, error)
}

// Maven ...
type Maven struct {
	PomPath             string
	Runner              mavenRunner
	ExecRunner          mavenExecRunner
	ProjectSettingsFile string
	GlobalSettingsFile  string
	M2Path              string
}

// InitBuildDescriptor ...
func (m *Maven) init() {
	if len(m.PomPath) == 0 {
		m.PomPath = "pom.xml"
	}

	if m.ExecRunner == nil {
		m.ExecRunner = &command.Command{}
	}
}

// VersioningScheme ...
func (m *Maven) VersioningScheme() string {
	return "maven"
}

// GetVersion ...
func (m *Maven) GetVersion() (string, error) {
	m.init()

	version, err := m.Runner.Evaluate(m.PomPath, "project.version", m.ExecRunner)
	if err != nil {
		return "", errors.Wrap(err, "Maven - getting version failed")
	}
	//ToDo: how to deal with SNAPSHOT replacement?
	return version, nil
}

// SetVersion ...
func (m *Maven) SetVersion(version string) error {
	m.init()

	groupID, err := m.Runner.Evaluate(m.PomPath, "project.groupId", m.ExecRunner)
	if err != nil {
		return errors.Wrap(err, "Maven - getting groupId failed")
	}
	opts := maven.ExecuteOptions{
		PomPath:             m.PomPath,
		ProjectSettingsFile: m.ProjectSettingsFile,
		GlobalSettingsFile:  m.GlobalSettingsFile,
		M2Path:              m.M2Path,
		Goals:               []string{"org.codehaus.mojo:versions-maven-plugin:2.3:set"},
		Defines: []string{
			fmt.Sprintf("-DnewVersion=%v}", version),
			fmt.Sprintf("-DgroupId=%v}", groupID),
			"-DartifactId='*'",
			"-DoldVersion='*'",
			"-DgenerateBackupPoms=false",
		},
	}
	_, err = m.Runner.Execute(&opts, m.ExecRunner)
	if err != nil {
		return errors.Wrapf(err, "Maven - setting version %v failed", version)
	}
	return nil
}
