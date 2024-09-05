package cnbutils

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
)

func CopyProject(source, target string, include, exclude *ignore.GitIgnore, utils BuildUtils, follow bool) error {
	sourceFiles, err := utils.Glob(path.Join(source, "**"))
	if err != nil {
		return err
	}

	knownSymlinks := []string{}

	for _, sourceFile := range sourceFiles {
		skipFile, err := isIgnored(source, sourceFile, include, exclude, knownSymlinks)
		if err != nil {
			return err
		}

		if !skipFile {
			target := path.Join(target, strings.ReplaceAll(sourceFile, source, ""))

			isDir, err := utils.DirExists(sourceFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorBuild)
				return errors.Wrapf(err, "Checking file info '%s' failed", target)
			}

			isSymlink, err := symlinkExists(sourceFile, utils)
			if err != nil {
				return err
			}

			if isSymlink {
				linkTarget, err := utils.Readlink(sourceFile)
				if err != nil {
					return err
				}
				if !isDir || !follow {
					log.Entry().Debugf("Creating symlink from %s to %s", linkTarget, target)
					err = utils.Symlink(linkTarget, target)
					if err != nil {
						return err
					}
				}
				if !follow {
					log.Entry().Debugf("Adding %s to list of known symlinks", sourceFile)
					knownSymlinks = append(knownSymlinks, sourceFile)
				}
			} else if isDir {
				log.Entry().Debugf("Creating directory %s", target)
				err = utils.MkdirAll(target, os.ModePerm)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return errors.Wrapf(err, "Creating directory '%s' failed", target)
				}
			} else {
				log.Entry().Debugf("Copying '%s' to '%s'", sourceFile, target)
				err = copyFile(sourceFile, target, utils)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return errors.Wrapf(err, "Copying '%s' to '%s' failed", sourceFile, target)
				}
			}
		}
	}
	return nil
}

func symlinkExists(path string, utils BuildUtils) (bool, error) {
	lstat, err := utils.Lstat(path)
	return lstat.Mode().Type() == fs.ModeSymlink, err
}

func copyFile(source, target string, utils BuildUtils) error {
	targetDir := filepath.Dir(target)

	exists, err := utils.DirExists(targetDir)
	if err != nil {
		return err
	}

	if !exists {
		log.Entry().Debugf("Creating directory %s", targetDir)
		err = utils.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	log.Entry().Debugf("Copying %s to %s", source, target)
	_, err = utils.Copy(source, target)
	return err
}

func isIgnored(source, sourceFile string, include, exclude *ignore.GitIgnore, knownSymlinks []string) (bool, error) {
	find, err := filepath.Rel(source, sourceFile)
	if err != nil {
		log.SetErrorCategory(log.ErrorBuild)
		return false, errors.Wrapf(err, "Calculating relative path for '%s' failed", sourceFile)
	}

	for _, link := range knownSymlinks {
		if strings.HasPrefix(sourceFile, link) {
			return true, nil
		}
	}

	if exclude != nil {
		filtered := exclude.MatchesPath(find)

		if filtered {
			log.Entry().Debugf("%s matches exclude pattern, ignoring", find)
			return true, nil
		}
	}

	if include != nil {
		filtered := !include.MatchesPath(find)

		if filtered {
			log.Entry().Debugf("%s doesn't match include pattern, ignoring", find)
			return true, nil
		} else {
			log.Entry().Debugf("%s matches include pattern", find)
			return false, nil
		}
	}

	return false, nil
}
