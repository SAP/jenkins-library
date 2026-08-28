package cnbutils

import (
	"bytes"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

type BuildSummary struct {
	Builder          string
	LifecycleVersion string

	Images []*ImageSummary
}

func (bs *BuildSummary) Print() {
	log.Entry().Infoln("")
	log.Entry().Info("*** Build Summary ***")
	log.Entry().Infof("  Builder: %q", bs.Builder)
	log.Entry().Infof("  Lifecycle: %q", bs.LifecycleVersion)
	log.Entry().Infof("  %d image(s) build:", len(bs.Images))
	log.Entry().Infoln("")
	for _, image := range bs.Images {
		image.Print()
		log.Entry().Infoln("")
	}
}

func NewBuildSummary(builder string, utils BuildUtils) *BuildSummary {
	return &BuildSummary{
		Builder:          builder,
		LifecycleVersion: lifecycleVersion(utils),
	}
}

type ImageSummary struct {
	ImageRef          string
	ProjectDescriptor string
	Buildpacks        []string
	EnvVars           []string
}

func (is *ImageSummary) Print() {
	log.Entry().Infof("  Image: %q", is.ImageRef)
	log.Entry().Infof("    Project descriptor: %q", is.ProjectDescriptor)
	log.Entry().Infof("    Env: %q", strings.Join(is.EnvVars, ", "))
}

func (is *ImageSummary) AddEnv(env map[string]any) {
	for key := range env {
		is.EnvVars = append(is.EnvVars, key)
	}
}

func lifecycleVersion(utils BuildUtils) string {
	currentStdout := utils.GetStdout()

	buf := bytes.NewBufferString("")
	utils.Stdout(buf)
	_ = utils.RunExecutable("/cnb/lifecycle/lifecycle", "-version")
	utils.Stdout(currentStdout)

	return strings.Trim(buf.String(), "\n")
}
