package buildpacks

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/cnbutils/privacy"
	"github.com/SAP/jenkins-library/pkg/cnbutils/project"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

const version = 3

type Telemetry struct {
	customData *telemetry.CustomData
	data       *BuildpacksTelemetry
}

func NewTelemetry(customData *telemetry.CustomData) *Telemetry {
	return &Telemetry{
		customData: customData,
		data: &BuildpacksTelemetry{
			Version: version,
		},
	}
}

func (d *Telemetry) Export() error {
	d.customData.Custom1Label = "cnbBuildStepData"
	customData, err := json.Marshal(d.data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal custom telemetry data")
	}
	d.customData.Custom1 = string(customData)
	return nil
}

func (d *Telemetry) WithImage(image string) {
	d.data.builder = image
}

func (d *Telemetry) AddSegment(segment *Segment) {
	segment.data.Builder = d.data.builder
	d.data.Data = append(d.data.Data, segment.data)
}

type BuildpacksTelemetry struct {
	builder string
	Version int                      `json:"version"`
	Data    []*cnbBuildTelemetryData `json:"data"`
}

type cnbBuildTelemetryData struct {
	ImageTag          string                                 `json:"imageTag"`
	AdditionalTags    []string                               `json:"additionalTags"`
	BindingKeys       []string                               `json:"bindingKeys"`
	Path              PathEnum                               `json:"path"`
	BuildEnv          cnbBuildTelemetryDataBuildEnv          `json:"buildEnv"`
	Buildpacks        cnbBuildTelemetryDataBuildpacks        `json:"buildpacks"`
	ProjectDescriptor cnbBuildTelemetryDataProjectDescriptor `json:"projectDescriptor"`
	BuildTool         string                                 `json:"buildTool"`
	Builder           string                                 `json:"builder"`
}

type cnbBuildTelemetryDataBuildEnv struct {
	KeysFromConfig            []string               `json:"keysFromConfig"`
	KeysFromProjectDescriptor []string               `json:"keysFromProjectDescriptor"`
	KeysOverall               []string               `json:"keysOverall"`
	JVMVersion                string                 `json:"jvmVersion"`
	KeyValues                 map[string]interface{} `json:"keyValues"`
}

type cnbBuildTelemetryDataBuildpacks struct {
	FromConfig            []string `json:"FromConfig"`
	FromProjectDescriptor []string `json:"FromProjectDescriptor"`
	Overall               []string `json:"overall"`
}

type cnbBuildTelemetryDataProjectDescriptor struct {
	Used        bool `json:"used"`
	IncludeUsed bool `json:"includeUsed"`
	ExcludeUsed bool `json:"excludeUsed"`
}

type Segment struct {
	data *cnbBuildTelemetryData
}

func NewSegment() *Segment {
	return &Segment{
		data: &cnbBuildTelemetryData{},
	}
}

func (s *Segment) WithBindings(bindings map[string]interface{}) *Segment {
	var bindingKeys []string
	for k := range bindings {
		bindingKeys = append(bindingKeys, k)
	}
	s.data.BindingKeys = bindingKeys
	return s
}

func (s *Segment) WithEnv(env map[string]interface{}) *Segment {
	s.data.BuildEnv.KeysFromConfig = []string{}
	s.data.BuildEnv.KeysOverall = []string{}
	for key := range env {
		s.data.BuildEnv.KeysFromConfig = append(s.data.BuildEnv.KeysFromConfig, key)
		s.data.BuildEnv.KeysOverall = append(s.data.BuildEnv.KeysOverall, key)
	}
	return s
}

func (s *Segment) WithTags(tag string, additionalTags []string) *Segment {
	s.data.ImageTag = tag
	s.data.AdditionalTags = additionalTags
	return s
}

func (s *Segment) WithPath(path PathEnum) *Segment {
	s.data.Path = path
	return s
}

func (s *Segment) WithBuildTool(buildTool string) *Segment {
	s.data.BuildTool = buildTool
	return s
}

func (s *Segment) WithBuilder(builder string) *Segment {
	s.data.Builder = privacy.FilterBuilder(builder)
	return s
}

func (s *Segment) WithBuildpacksFromConfig(buildpacks []string) *Segment {
	s.data.Buildpacks.FromConfig = privacy.FilterBuildpacks(buildpacks)
	return s
}

func (s *Segment) WithBuildpacksOverall(buildpacks []string) *Segment {
	s.data.Buildpacks.Overall = privacy.FilterBuildpacks(buildpacks)
	return s
}

func (s *Segment) WithKeyValues(env map[string]interface{}) *Segment {
	s.data.BuildEnv.KeyValues = privacy.FilterEnv(env)
	return s
}

func (s *Segment) WithProjectDescriptor(descriptor *project.Descriptor) *Segment {
	descriptorKeys := s.data.BuildEnv.KeysFromProjectDescriptor
	overallKeys := s.data.BuildEnv.KeysOverall
	for key := range descriptor.EnvVars {
		descriptorKeys = append(descriptorKeys, key)
		overallKeys = append(overallKeys, key)
	}
	s.data.BuildEnv.KeysFromProjectDescriptor = descriptorKeys
	s.data.BuildEnv.KeysOverall = overallKeys
	s.data.Buildpacks.FromProjectDescriptor = privacy.FilterBuildpacks(descriptor.Buildpacks)
	s.data.ProjectDescriptor.Used = true
	s.data.ProjectDescriptor.IncludeUsed = descriptor.Include != nil
	s.data.ProjectDescriptor.ExcludeUsed = descriptor.Exclude != nil
	return s
}
