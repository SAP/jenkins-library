package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
)

func ensureMavenBOMTimestamps(pomPath string, utils maven.Utils) error {
	if pomPath == "" {
		pomPath = "pom.xml"
	}
	root := filepath.Dir(pomPath)
	now := time.Now().UTC()
	matched := 0
	for _, name := range []string{mvnBomFilename, mvnSimpleBomFilename} {
		pattern := filepath.ToSlash(filepath.Join(root, "**", name+".xml"))
		files, err := utils.Glob(pattern)
		if err != nil {
			return fmt.Errorf("find %s: %w", pattern, err)
		}
		for _, file := range files {
			matched++
			input, err := utils.FileRead(file)
			if err != nil {
				return fmt.Errorf("read %s: %w", file, err)
			}
			output, changed, err := maven.EnsureBOMTimestamp(input, now)
			if err != nil {
				return fmt.Errorf("post-process %s: %w", file, err)
			}
			if changed {
				if err := utils.FileWrite(file, output, 0o644); err != nil {
					return fmt.Errorf("write %s: %w", file, err)
				}
				log.Entry().Infof("Added Maven SBOM metadata timestamp: %s", file)
			}
		}
	}
	if matched == 0 {
		return fmt.Errorf("no Maven BOM files found below %s", root)
	}
	return nil
}
