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

// Maven defines a maven artifact used for versioning
type Maven struct {
	pomPath             string
	runner              mavenRunner
	execRunner          mavenExecRunner
	projectSettingsFile string
	globalSettingsFile  string
	m2Path              string
}

func (m *Maven) init() {
	if len(m.pomPath) == 0 {
		m.pomPath = "pom.xml"
	}

	if m.execRunner == nil {
		m.execRunner = &command.Command{}
	}
}

// BuildDescriptorPattern returns the pattern for the relevant build descriptor files
func (m *Maven) BuildDescriptorPattern() string {
	return "*pom.xml"
}

// VersioningScheme returns the relevant versioning scheme
func (m *Maven) VersioningScheme() string {
	return "maven"
}

// GetVersion returns the current version of the artifact
func (m *Maven) GetVersion() (string, error) {
	m.init()

	version, err := m.runner.Evaluate(m.pomPath, "project.version", m.execRunner)
	if err != nil {
		return "", errors.Wrap(err, "Maven - getting version failed")
	}
	//ToDo: how to deal with SNAPSHOT replacement?
	return version, nil
}

// SetVersion updates the version of the artifact
func (m *Maven) SetVersion(version string) error {
	m.init()

	groupID, err := m.runner.Evaluate(m.pomPath, "project.groupId", m.execRunner)
	if err != nil {
		return errors.Wrap(err, "Maven - getting groupId failed")
	}
	opts := maven.ExecuteOptions{
		PomPath:             m.pomPath,
		ProjectSettingsFile: m.projectSettingsFile,
		GlobalSettingsFile:  m.globalSettingsFile,
		M2Path:              m.m2Path,
		Goals:               []string{"org.codehaus.mojo:versions-maven-plugin:2.7:set"},
		Defines: []string{
			fmt.Sprintf("-DnewVersion=%v", version),
			fmt.Sprintf("-DgroupId=%v", groupID),
			"-DartifactId=*",
			"-DoldVersion=*",
			"-DgenerateBackupPoms=false",
		},
	}
	_, err = m.runner.Execute(&opts, m.execRunner)
	if err != nil {
		return errors.Wrapf(err, "Maven - setting version %v failed", version)
	}
	return nil
}
