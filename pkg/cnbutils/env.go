package cnbutils

import (
	"fmt"
	"path/filepath"
)

func CreateEnvFiles(utils BuildUtils, platformPath string, env map[string]interface{}) error {
	envDir := filepath.Join(platformPath, "env")
	err := utils.MkdirAll(envDir, 0755)
	if err != nil {
		return err
	}

	for k, v := range env {
		err = utils.FileWrite(filepath.Join(envDir, k), []byte(fmt.Sprintf("%v", v)), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
