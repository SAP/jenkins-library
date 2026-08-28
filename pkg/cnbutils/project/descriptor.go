// Package project handles project.toml parsing
package project

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project/types"
	v01 "github.com/SAP/jenkins-library/pkg/cnbutils/project/v01"
	v02 "github.com/SAP/jenkins-library/pkg/cnbutils/project/v02"
	"github.com/SAP/jenkins-library/pkg/cnbutils/registry"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	ignore "github.com/sabhiram/go-gitignore"
)

type project struct {
	Version string `toml:"schema-version"`
}

type versionDescriptor struct {
	Project project `toml:"_"`
}

var parsers = map[string]func(string) (types.Descriptor, error){
	"0.1": v01.NewDescriptor,
	"0.2": v02.NewDescriptor,
}

type Descriptor struct {
	Exclude        *ignore.GitIgnore
	Include        *ignore.GitIgnore
	EnvVars        map[string]any
	Buildpacks     []string
	PreBuildpacks  []string
	PostBuildpacks []string
	ProjectID      string
}

func ParseDescriptor(descriptorPath string, utils cnbutils.BuildUtils, httpClient piperhttp.Sender) (*Descriptor, error) {
	descriptor := &Descriptor{}

	descriptorContent, err := utils.FileRead(descriptorPath)
	if err != nil {
		return nil, err
	}

	var versionDescriptor versionDescriptor
	_, err = toml.Decode(string(descriptorContent), &versionDescriptor)
	if err != nil {
		return &Descriptor{}, fmt.Errorf("parsing schema version: %w", err)
	}

	version := versionDescriptor.Project.Version
	if version == "" {
		version = "0.1"
	}

	rawDescriptor, err := parsers[version](string(descriptorContent))
	if err != nil {
		return &Descriptor{}, err
	}

	if len(rawDescriptor.Build.Buildpacks) > 0 {
		descriptor.Buildpacks, err = searchBuildpacks(rawDescriptor.Build.Buildpacks, httpClient)
		if err != nil {
			return nil, err
		}
	}

	if len(rawDescriptor.Build.Pre.Buildpacks) > 0 {
		descriptor.PreBuildpacks, err = searchBuildpacks(rawDescriptor.Build.Pre.Buildpacks, httpClient)
		if err != nil {
			return nil, err
		}
	}

	if len(rawDescriptor.Build.Post.Buildpacks) > 0 {
		descriptor.PostBuildpacks, err = searchBuildpacks(rawDescriptor.Build.Post.Buildpacks, httpClient)
		if err != nil {
			return nil, err
		}
	}

	if len(rawDescriptor.Build.Env) > 0 {
		descriptor.EnvVars = envToMap(rawDescriptor.Build.Env)
	}

	if len(rawDescriptor.Build.Exclude) > 0 && len(rawDescriptor.Build.Include) > 0 {
		return nil, errors.New("project descriptor options 'exclude' and 'include' are mutually exclusive")
	}

	if len(rawDescriptor.Build.Exclude) > 0 {
		descriptor.Exclude = ignore.CompileIgnoreLines(rawDescriptor.Build.Exclude...)
	}

	if len(rawDescriptor.Build.Include) > 0 {
		descriptor.Include = ignore.CompileIgnoreLines(rawDescriptor.Build.Include...)
	}

	if len(rawDescriptor.Project.ID) > 0 {
		descriptor.ProjectID = rawDescriptor.Project.ID
	}

	return descriptor, nil
}

func envToMap(env []types.EnvVar) map[string]any {
	envMap := map[string]any{}

	for _, e := range env {
		if len(e.Name) == 0 {
			continue
		}

		envMap[e.Name] = e.Value
	}

	return envMap
}

func searchBuildpacks(buildpacks []types.Buildpack, httpClient piperhttp.Sender) ([]string, error) {
	var bpackImg []string

	for _, bpack := range buildpacks {
		if bpack.Script != (types.Script{}) {
			return nil, errors.New("inline buildpacks are not supported")
		}

		if bpack.URI != "" {
			log.Entry().Debugf("Adding buildpack using URI: %s", bpack.URI)
			bpackImg = append(bpackImg, bpack.URI)
		} else if bpack.ID != "" {
			imgURL, err := registry.SearchBuildpack(bpack.ID, bpack.Version, httpClient, "")
			if err != nil {
				return nil, err
			}

			bpackImg = append(bpackImg, imgURL)
		} else {
			return nil, errors.New("invalid buildpack entry in project.toml, either URI or ID should be specified")
		}

	}

	return bpackImg, nil
}
