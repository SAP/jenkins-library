package sonar

import (
	"fmt"
	"path/filepath"

	"github.com/magiconair/properties"
)

// TaskReportData encapsulates information about an executed Sonar scan task.
// https://pkg.go.dev/github.com/magiconair/properties@v1.8.0?tab=doc#Properties.Decode
type TaskReportData struct {
	ProjectKey    string `properties:"projectKey"`
	TaskID        string `properties:"ceTaskId"`
	DashboardURL  string `properties:"dashboardUrl"`
	TaskURL       string `properties:"ceTaskUrl"`
	ServerURL     string `properties:"serverUrl"`
	ServerVersion string `properties:"serverVersion"`
}

// LoadProperties reads a properties file assembled from the given path segments.
func LoadProperties(path ...string) (*properties.Properties, error) {
	return properties.LoadFile(filepath.Join(path...), properties.UTF8)
}

// GetProperty reads sonar-project.properties from the workspace directory and returns the value of
// the given property, or an empty string if the property is absent or the file cannot be read.
func GetProperty(workspace string, property string) string {
	props, err := LoadProperties(workspace, "sonar-project.properties")
	if err != nil {
		return ""
	}
	return props.GetString(property, "")
}

// ReadTaskReport expects a file ".scannerwork/report-task.txt" to exist in the provided workspace directory,
// and parses its contents into the returned TaskReportData struct.
func ReadTaskReport(workspace string) (result TaskReportData, err error) {
	reportFile := filepath.Join(workspace, ".scannerwork", "report-task.txt")
	reportContent, err := LoadProperties(reportFile)
	if err != nil {
		return
	}
	err = reportContent.Decode(&result)
	if err != nil {
		err = fmt.Errorf("decode %s: %w", reportFile, err)
	}
	return
}
