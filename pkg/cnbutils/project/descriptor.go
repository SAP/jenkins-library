// Package project handles project.toml parsing
package project

import (
	"errors"

	"github.com/SAP/jenkins-library/pkg/cnbutils"
	"github.com/SAP/jenkins-library/pkg/cnbutils/registry"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pelletier/go-toml"
	ignore "github.com/sabhiram/go-gitignore"
)

type script struct {
	API    string `toml:"api"`
	Inline string `toml:"inline"`
	Shell  string `toml:"shell"`
}
type buildpack struct {
	ID      string `toml:"id"`
	Version string `toml:"version"`
	URI     string `toml:"uri"`
	Script  script `toml:"script"`
}

type envVar struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

type build struct {
	Include    []string    `toml:"include"`
	Exclude    []string    `toml:"exclude"`
	Buildpacks []buildpack `toml:"buildpacks"`
	Env        []envVar    `toml:"env"`
}

type project struct {
	ID string `toml:"id"`
}

type projectDescriptor struct {
	Build    build                  `toml:"build"`
	Project  project                `toml:"project"`
	Metadata map[string]interface{} `toml:"metadata"`
}

type Descriptor struct {
	Exclude    *ignore.GitIgnore
	Include    *ignore.GitIgnore
	EnvVars    map[string]interface{}
	Buildpacks []string
	ProjectID  string
}

func ParseDescriptor(descriptorPath string, utils cnbutils.BuildUtils, httpClient piperhttp.Sender) (*Descriptor, error) {
	descriptor := &Descriptor{}

	descriptorContent, err := utils.FileRead(descriptorPath)
	if err != nil {
		return nil, err
	}

	rawDescriptor := projectDescriptor{}
	err = toml.Unmarshal(descriptorContent, &rawDescriptor)
	if err != nil {
		return nil, err
	}

	if rawDescriptor.Build.Buildpacks != nil && len(rawDescriptor.Build.Buildpacks) > 0 {
		buildpacksImg, err := rawDescriptor.Build.searchBuildpacks(httpClient)
		if err != nil {
			return nil, err
		}

		descriptor.Buildpacks = buildpacksImg
	}

	if rawDescriptor.Build.Env != nil && len(rawDescriptor.Build.Env) > 0 {
		descriptor.EnvVars = rawDescriptor.Build.envToMap()
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

func (b *build) envToMap() map[string]interface{} {
	envMap := map[string]interface{}{}

	for _, e := range b.Env {
		if len(e.Name) == 0 || len(e.Value) == 0 {
			continue
		}

		envMap[e.Name] = e.Value
	}

	return envMap
}

func (b *build) searchBuildpacks(httpClient piperhttp.Sender) ([]string, error) {
	var bpackImg []string

	for _, bpack := range b.Buildpacks {
		if bpack.Script != (script{}) {
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
