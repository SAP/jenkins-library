// Package metadata handles generation of the project-metadata.toml
package metadata

import (
	"bytes"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/buildpacks/lifecycle/platform/files"
)

var metadataFilePath = "/layers/project-metadata.toml"

func writeProjectMetadata(metadata files.ProjectMetadata, path string, utils cnbutils.BuildUtils) error {
	var buf bytes.Buffer

	err := toml.NewEncoder(&buf).Encode(metadata)
	if err != nil {
		return err
	}

	err = utils.FileWrite(path, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func extractMetadataFromCPE(piperEnvRoot string, utils cnbutils.BuildUtils) files.ProjectMetadata {
	cpePath := filepath.Join(piperEnvRoot, "commonPipelineEnvironment")
	return files.ProjectMetadata{
		Source: &files.ProjectSource{
			Type: "git",
			Version: map[string]any{
				"commit": piperenv.GetResourceParameter(cpePath, "git", "headCommitId"),
			},
			Metadata: map[string]any{
				"refs": []string{
					piperenv.GetResourceParameter(cpePath, "git", "branch"),
				},
			},
		},
	}
}

func WriteProjectMetadata(piperEnvRoot string, utils cnbutils.BuildUtils) {
	projectMetadata := extractMetadataFromCPE(piperEnvRoot, utils)

	err := writeProjectMetadata(projectMetadata, metadataFilePath, utils)
	if err != nil {
		log.Entry().Warnf("failed write 'project-metadata.toml', error: %s", err.Error())
		return
	}
}
