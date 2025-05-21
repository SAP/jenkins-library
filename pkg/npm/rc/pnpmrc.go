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
	pnpmConfigFilename = ".npmrc"
)

// PNPM implements the RCManager interface for managing .npmrc files for pnpm
type PNPM struct {
	filepath string
	content  string
	utils    Utils
}

// NewPNPM creates a new PNPM RC manager instance
func NewPNPM(path string, utils Utils) *PNPM {
	if !strings.HasSuffix(path, pnpmConfigFilename) {
		path = filepath.Join(path, pnpmConfigFilename)
	}
	return &PNPM{filepath: path, utils: utils}
}

// Write saves the current content to the .npmrc file
func (rc *PNPM) Write() error {
	if err := propertiesWriteFile(rc.filepath, []byte(rc.content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", rc.filepath, err)
	}
	return nil
}

// Load reads the content from the .npmrc file
func (rc *PNPM) Load() error {
	bytes, err := propertiesLoadFile(rc.filepath)
	if err != nil {
		return err
	}
	rc.content = string(bytes)
	return nil
}

// Set updates or adds a key-value pair in the .npmrc content
func (rc *PNPM) Set(key, value string) {
	r := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*%s\s*=.*$`, key))

	keyValue := fmt.Sprintf("%s=%s", key, value)

	if r.MatchString(rc.content) {
		rc.content = r.ReplaceAllString(rc.content, keyValue)
	} else {
		rc.content += keyValue + "\n"
	}
}

// SetRegistry configures the pnpm registry and authentication
func (rc *PNPM) SetRegistry(registry, username, password, scope string) error {
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
		log.Entry().Debugf("creating new npmrc file at %s", rc.filepath)
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
		// As per https://github.blog/changelog/2022-10-24-npm-v9-0-0-released/
		// auth settings must be scoped to a specific registry
		rc.Set(fmt.Sprintf("%s:%s", strings.TrimPrefix(registry, "https:"), "_auth"), CredentialUtils.EncodeUsernamePassword(username, password))
		rc.Set("always-auth", "true")
	}

	return rc.Write()
}

func (rc *PNPM) GetFilePath() string {
	return rc.filepath
}
