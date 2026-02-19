package cnbutils

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"

	ignore "github.com/sabhiram/go-gitignore"
)

func shouldBeFiltered(path string, knownSymlinks []string) bool {
	for _, symlink := range knownSymlinks {
		if strings.HasPrefix(path, symlink) {
			return true
		}
	}
	return false
}

func filterSymlinks(sourceFiles []string, utils BuildUtils) ([]string, error) {
	filteredFiles := []string{}
	knownSymlinks := []string{}

	for _, sourceFile := range sourceFiles {
		if shouldBeFiltered(sourceFile, knownSymlinks) {
			continue
		}

		isSymlink, err := symlinkExists(sourceFile, utils)
		if err != nil {
			return nil, err
		}

		if isSymlink {
			log.Entry().Debugf("Ignoring any path below %q", sourceFile)
			knownSymlinks = append(knownSymlinks, sourceFile)
		}
		filteredFiles = append(filteredFiles, sourceFile)
	}
	return filteredFiles, nil
}

func CopyProject(source, target string, include, exclude *ignore.GitIgnore, utils BuildUtils, follow bool) error {
	sourceFiles, err := utils.Glob(path.Join(source, "**"))
	if err != nil {
		return err
	}

	if !follow {
		sourceFiles, err = filterSymlinks(sourceFiles, utils)
		if err != nil {
			return err
		}
	}

	for _, sourceFile := range sourceFiles {
		relPath, err := filepath.Rel(source, sourceFile)
		if err != nil {
			log.SetErrorCategory(log.ErrorBuild)
			return fmt.Errorf("Calculating relative path for '%s' failed: %w", sourceFile, err)
		}

		if !isIgnored(relPath, include, exclude) {
			target := path.Join(target, strings.ReplaceAll(sourceFile, source, ""))

			isSymlink, err := symlinkExists(sourceFile, utils)
			if err != nil {
				return err
			}

			isDir, err := utils.DirExists(sourceFile)
			if err != nil {
				return err
			}

			if isSymlink {
				linkTarget, err := utils.Readlink(sourceFile)
				if err != nil {
					return err
				}
				log.Entry().Debugf("Creating symlink from %q to %q", target, linkTarget)
				err = utils.Symlink(linkTarget, target)
				if err != nil {
					return err
				}
			} else if isDir {
				err = utils.MkdirAll(target, os.ModePerm)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return fmt.Errorf("Creating directory '%s' failed: %w", target, err)
				}
			} else {
				log.Entry().Debugf("Copying '%s' to '%s'", sourceFile, target)
				err = copyFile(sourceFile, target, utils)
				if err != nil {
					log.SetErrorCategory(log.ErrorBuild)
					return fmt.Errorf("Copying '%s' to '%s' failed: %w", sourceFile, target, err)
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
		log.Entry().Debugf("Creating directory '%s'", targetDir)
		err = utils.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	log.Entry().Debugf("Copying '%s' to '%s'", source, target)
	_, err = utils.Copy(source, target)
	return err
}

func isIgnored(find string, include, exclude *ignore.GitIgnore) bool {
	if exclude != nil {
		filtered := exclude.MatchesPath(find)

		if filtered {
			log.Entry().Debugf("%s matches exclude pattern, ignoring", find)
			return true
		}
	}

	if include != nil {
		filtered := !include.MatchesPath(find)

		if filtered {
			log.Entry().Debugf("%s doesn't match include pattern, ignoring", find)
			return true
		} else {
			log.Entry().Debugf("%s matches include pattern", find)
			return false
		}
	}

	return false
}
