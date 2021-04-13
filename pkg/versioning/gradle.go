package versioning

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

type gradleExecRunner interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

// Gradle defines a maven artifact used for versioning
type Gradle struct {
	execRunner     gradleExecRunner
	gradlePropsOut []byte
	path           string
	propertiesFile *PropertiesFile
	versionField   string
	writeFile      func(string, []byte, os.FileMode) error
}

func (g *Gradle) init() error {
	if g.writeFile == nil {
		g.writeFile = ioutil.WriteFile
	}

	if g.propertiesFile == nil {
		g.propertiesFile = &PropertiesFile{
			path:             g.path,
			versioningScheme: g.VersioningScheme(),
			versionField:     g.versionField,
			writeFile:        g.writeFile,
		}
		err := g.propertiesFile.init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Gradle) initGetArtifact() error {
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
	result := Coordinates{}
	var err error
	// result.GroupID, err = g.GetGroupID()
	// if err != nil {
	// 	return nil, err
	// }
	result.ArtifactID, err = g.GetArtifactID()
	if err != nil {
		return result, err
	}
	result.Version, err = g.GetVersion()
	if err != nil {
		return result, err
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
	err := g.initGetArtifact()
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
	err := g.init()
	if err != nil {
		return "", err
	}

	return g.propertiesFile.GetVersion()
}

// SetVersion updates the version of the artifact
func (g *Gradle) SetVersion(version string) error {
	err := g.init()
	if err != nil {
		return err
	}
	return g.propertiesFile.SetVersion(version)
}
