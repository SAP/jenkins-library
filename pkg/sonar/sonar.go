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

// ReadTaskReport expects a file ".scannerwork/report-task.txt" to exist in the provided workspace directory,
// and parses its contents into the returned TaskReportData struct.
func ReadTaskReport(workspace string) (result TaskReportData, err error) {
	reportFile := filepath.Join(workspace, ".scannerwork", "report-task.txt")
	// read file content
	reportContent, err := properties.LoadFile(reportFile, properties.UTF8)
	if err != nil {
		return
	}
	// read content into struct
	err = reportContent.Decode(&result)
	if err != nil {
		err = fmt.Errorf("decode %s: %w", reportFile, err)
	}
	return
}
