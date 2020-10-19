// +build !release

package whitesource

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"net/http"
	"os"
)

func newTestScan(config *ScanOptions) *Scan {
	return &Scan{
		AggregateProjectName: config.ProjectName,
		ProductVersion:       "product-version",
	}
}

type NpmInstall struct {
	currentDir  string
	packageJSON []string
}

type DownloadedFile struct {
	sourceURL string
	filePath  string
}

type ScanUtilsMock struct {
	*mock.FilesMock
	*mock.ExecMockRunner
	NpmInstalledModules []NpmInstall
	DownloadedFiles     []DownloadedFile
}

func (m *ScanUtilsMock) RemoveAll(path string) error {
	// TODO: Implement in FS Mock
	return nil
}

func (m *ScanUtilsMock) FindPackageJSONFiles(_ *ScanOptions) ([]string, error) {
	matches, _ := m.Glob("**/package.json")
	return matches, nil
}

func (m *ScanUtilsMock) InstallAllNPMDependencies(_ *ScanOptions, packageJSONs []string) error {
	m.NpmInstalledModules = append(m.NpmInstalledModules, NpmInstall{
		currentDir:  m.CurrentDir,
		packageJSON: packageJSONs,
	})
	return nil
}

func (m *ScanUtilsMock) DownloadFile(url, filename string, _ http.Header, _ []*http.Cookie) error {
	m.DownloadedFiles = append(m.DownloadedFiles, DownloadedFile{sourceURL: url, filePath: filename})
	return nil
}

func (m *ScanUtilsMock) FileOpen(name string, flag int, perm os.FileMode) (File, error) {
	return m.Open(name, flag, perm)
}

func newScanUtilsMock() *ScanUtilsMock {
	return &ScanUtilsMock{
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
}
