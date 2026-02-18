package cnbutils

import (
	"fmt"
	"path/filepath"
)

func CreateEnvFiles(utils BuildUtils, platformPath string, env map[string]any) error {
	envDir := filepath.Join(platformPath, "env")
	err := utils.MkdirAll(envDir, 0755)
	if err != nil {
		return err
	}

	for k, v := range env {
		err = utils.FileWrite(filepath.Join(envDir, k), fmt.Appendf(nil, "%v", v), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
