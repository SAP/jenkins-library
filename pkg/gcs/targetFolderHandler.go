package gcs

import "path"

// GetTargetFolder calculates the target folder in GCS bucket
func GetTargetFolder(folderPath string, stepResultType string, subFolder string) string {
	return path.Join(folderPath, stepResultType, subFolder)
}
