package rc

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	CredentialUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

const (
	yarnConfigFilename = ".yarnrc"
)

// Yarn implements the RCManager interface for managing .yarnrc files
type Yarn struct {
	filepath string
	content  string
	utils    Utils
}

// NewYarn creates a new Yarn RC manager instance
func NewYarn(path string, utils Utils) *Yarn {
	if !strings.HasSuffix(path, yarnConfigFilename) {
		path = filepath.Join(path, yarnConfigFilename)
	}
	return &Yarn{filepath: path, utils: utils}
}

// Write saves the current content to the .yarnrc file
func (rc *Yarn) Write() error {
	if err := propertiesWriteFile(rc.filepath, []byte(rc.content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", rc.filepath, err)
	}
	return nil
}

// Load reads the content from the .yarnrc file
func (rc *Yarn) Load() error {
	bytes, err := propertiesLoadFile(rc.filepath)
	if err != nil {
		return err
	}
	rc.content = string(bytes)
	return nil
}

// Set updates or adds a key-value pair in the .yarnrc content
func (rc *Yarn) Set(key, value string) {
	r := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*"%s"\s+.*$`, key))

	// Yarn uses a different format with quotes around keys and no equals sign
	keyValue := fmt.Sprintf(`"%s" "%s"`, key, value)

	if r.MatchString(rc.content) {
		rc.content = r.ReplaceAllString(rc.content, keyValue)
	} else {
		rc.content += keyValue + "\n"
	}
}

// SetRegistry configures the yarn registry and authentication
func (rc *Yarn) SetRegistry(registry, username, password, scope string) error {
	if len(registry) == 0 {
		log.Entry().Debug("no registry provided")
		return nil
	}

	exists, err := rc.utils.FileExists(rc.filepath)
	if err != nil {
		return fmt.Errorf("failed to check if %s exists: %w", rc.filepath, err)
	}
	if exists {
		log.Entry().Debugf("loading existing %s file", rc.filepath)
		if err := rc.Load(); err != nil {
			return fmt.Errorf("failed to read existing %s file: %w", rc.filepath, err)
		}
	} else {
		log.Entry().Debugf("creating new yarnrc file at %s", rc.filepath)
	}

	log.Entry().Debugf("adding registry %s", registry)

	// Set main registry entry
	rc.Set("registry", registry)

	// Set scoped registry if provided
	if scope != "" {
		rc.Set(fmt.Sprintf("%s:registry", scope), registry)
	}

	// Set auth credentials if provided
	if len(username) > 0 && len(password) > 0 {
		authString := CredentialUtils.EncodeUsernamePassword(username, password)
		rc.Set(fmt.Sprintf("%s:_auth", strings.TrimPrefix(registry, "https:")), authString)
		rc.Set("always-auth", "true")
	}

	return rc.Write()
}

func (rc *Yarn) GetFilePath() string {
	return rc.filepath
}
