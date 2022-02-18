package cnbutils

import (
	"os"
	"path/filepath"
	"strings"
)

func CopyGlob(src, dest string, globs []string, utils BuildUtils) error {
	for _, glob := range globs {
		matches, err := utils.Glob(filepath.Join(src, glob))
		if err != nil {
			return err
		}

		for _, match := range matches {
			destPath := filepath.Join(dest, strings.Replace(match, src, "", 1))
			err = utils.MkdirAll(filepath.Base(destPath), os.ModePerm)
			if err != nil {
				return err
			}
			_, err = utils.Copy(match, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
