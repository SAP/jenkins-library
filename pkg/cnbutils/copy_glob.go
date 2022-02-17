package cnbutils

import (
	"path/filepath"
	"strings"
)

func CopyGlob(src, dest, globs string, utils BuildUtils) error {
	for _, glob := range strings.Split(globs, ",") {
		matches, err := utils.Glob(filepath.Join(src, glob))
		if err != nil {
			return err
		}

		for _, match := range matches {
			destPath := filepath.Join(dest, strings.Replace(match, src, "", 1))
			_, err = utils.Copy(match, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
