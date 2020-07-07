package versioning

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"io"
	"regexp"
	"strings"
)

type gradleExecRunner interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

// GradleDescriptor holds the unique identifier combination for Gradle built Java artifacts
type GradleDescriptor struct {
	GroupID    string
	ArtifactID string
	Version    string
	Packaging  string
}

// Gradle defines a maven artifact used for versioning
type Gradle struct {
	execRunner     gradleExecRunner
	gradlePropsOut []byte
}

func (g *Gradle) init() error {
	if g.execRunner == nil {
		g.execRunner = &command.Command{}
	}

	if g.gradlePropsOut == nil {
		gradlePropsBuffer := &bytes.Buffer{}
		g.execRunner.Stdout(gradlePropsBuffer)
		err := g.execRunner.RunExecutable("gradle", "properties", "--no-daemon", "--console=plain", "-q")
		if err != nil {
			return err
		}
		g.gradlePropsOut = gradlePropsBuffer.Bytes()
		g.execRunner.Stdout(log.Writer())
	}
	return nil
}

// VersioningScheme returns the relevant versioning scheme
func (g *Gradle) VersioningScheme() string {
	return "semver2"
}

// GetCoordinates reads the coordinates from the maven pom.xml descriptor file
func (g *Gradle) GetCoordinates() (Coordinates, error) {
	result := &GradleDescriptor{}
	var err error
	// result.GroupID, err = g.GetGroupID()
	// if err != nil {
	// 	return nil, err
	// }
	result.ArtifactID, err = g.GetArtifactID()
	if err != nil {
		return nil, err
	}
	result.Version, err = g.GetVersion()
	if err != nil {
		return nil, err
	}
	// result.Packaging, err = g.GetPackaging()
	// if err != nil {
	// 	return nil, err
	// }
	return result, nil
}

// GetPackaging returns the current ID of the Group
// func (g *Gradle) GetPackaging() (string, error) {
// 	g.init()

// 	packaging, err := g.runner.Evaluate(&g.options, "project.packaging", g.execRunner)
// 	if err != nil {
// 		return "", errors.Wrap(err, "Gradle - getting packaging failed")
// 	}
// 	return packaging, nil
// }

// GetGroupID returns the current ID of the Group
// func (g *Gradle) GetGroupID() (string, error) {
// 	g.init()

// 	groupID, err := g.runner.Evaluate(&g.options, "project.groupId", g.execRunner)
// 	if err != nil {
// 		return "", errors.Wrap(err, "Gradle - getting groupId failed")
// 	}
// 	return groupID, nil
// }

// GetArtifactID returns the current ID of the artifact
func (g *Gradle) GetArtifactID() (string, error) {
	err := g.init()
	if err != nil {
		return "", err
	}

	regex := regexp.MustCompile(`(?m:^rootProject: root project '(.*)')`)
	match := string(regex.Find(g.gradlePropsOut))
	artifactID := strings.Split(match, `'`)[1]

	return artifactID, nil
}

// GetVersion returns the current version of the artifact
func (g *Gradle) GetVersion() (string, error) {
	versionID := "unspecified"
	err := g.init()
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile("(?m:^version: (.*))")
	match := r.FindString(string(g.gradlePropsOut))
	versionIDSlice := strings.Split(match, ` `)
	if len(versionIDSlice) > 1 {
		versionID = versionIDSlice[1]
	}

	return versionID, nil
}

// SetVersion updates the version of the artifact
func (g *Gradle) SetVersion(version string) error {
	return nil
}
