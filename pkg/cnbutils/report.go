package cnbutils

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/lifecycle/platform"
)

const reportFile = "/layers/report.toml"

func DigestFromReport(utils BuildUtils) (string, error) {
	report := platform.ExportReport{}

	data, err := utils.FileRead(reportFile)
	if err != nil {
		return "", err
	}

	err = toml.Unmarshal(data, &report)
	if err != nil {
		return "", err
	}

	if report.Image.Digest == "" {
		return "", fmt.Errorf("image digest is empty")
	}

	return report.Image.Digest, nil
}
