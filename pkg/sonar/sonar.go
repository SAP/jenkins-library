package sonar

import (
	"path"

	"github.com/magiconair/properties"
)

// ReportTaskData ...
// https://pkg.go.dev/github.com/magiconair/properties@v1.8.0?tab=doc#Properties.Decode
type ReportTaskData struct {
	ProjectKey    string `properties:"projectKey"`
	TaskID        string `properties:"ceTaskId"`
	DashboardURL  string `properties:"dashboardUrl"`
	TaskURL       string `properties:"ceTaskUrl"`
	ServerURL     string `properties:"serverUrl"`
	ServerVersion string `properties:"serverVersion"`
}

//ReadTaskReport ...
func ReadTaskReport(workspace string) (result ReportTaskData, err error) {
	reportFile := path.Join(workspace, ".scannerwork", "report-task.txt")
	// read file content
	reportContent, err := properties.LoadFile(reportFile, properties.UTF8)
	if err != nil {
		return
	}
	// read content into struct
	err = reportContent.Decode(&result)
	return
}
