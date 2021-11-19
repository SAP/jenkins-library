package gcs

import "path"

func GetTargetFolder(folderPath string, stepResultType string, subFolder string) string {
	return path.Join(folderPath, stepResultType, subFolder)
}
