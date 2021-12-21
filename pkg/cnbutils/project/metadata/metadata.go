// Package metadata handles generation of the project-metadata.toml
package metadata

import (
	"bytes"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/buildpacks/lifecycle/platform"
	"github.com/pelletier/go-toml"
)

var metadataFilePath = "/layers/project-metadata.toml"

func writeProjectMetadata(metadata platform.ProjectMetadata, path string, utils cnbutils.BuildUtils) error {
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

func extractMetadataFromCPE(piperEnvRoot string, utils cnbutils.BuildUtils) platform.ProjectMetadata {
	cpePath := filepath.Join(piperEnvRoot, "commonPipelineEnvironment")
	return platform.ProjectMetadata{
		Source: &platform.ProjectSource{
			Type: "git",
			Version: map[string]interface{}{
				"commit":   piperenv.GetResourceParameter(cpePath, "git", "headCommitId"),
				"describe": piperenv.GetResourceParameter(cpePath, "git", "commitMessage"),
			},
			Metadata: map[string]interface{}{
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
