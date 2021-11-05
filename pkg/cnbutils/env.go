package cnbutils

import (
	"fmt"
	"path/filepath"
	"strings"
)

func CreateEnvFiles(utils BuildUtils, platformPath string, env []string) error {
	envDir := filepath.Join(platformPath, "env")
	err := utils.MkdirAll(envDir, 0755)
	if err != nil {
		return err
	}

	for _, e := range env {
		eSplit := strings.SplitN(e, "=", 2)

		if len(eSplit) != 2 {
			return fmt.Errorf("invalid environment variable: %s", e)
		}

		err = utils.FileWrite(filepath.Join(envDir, eSplit[0]), []byte(eSplit[1]), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
