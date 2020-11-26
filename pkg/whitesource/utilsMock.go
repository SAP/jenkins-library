// +build !release

package whitesource

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"net/http"
	"os"
)

func newTestScan(config *ScanOptions) *Scan {
	return &Scan{
		AggregateProjectName: config.ProjectName,
		ProductVersion:       "product-version",
	}
}

// NpmInstall records in which directory "npm install" has been invoked and for which package.json files.
type NpmInstall struct {
	CurrentDir  string
	PackageJSON []string
}

// DownloadedFile records what URL has been downloaded to which file.
type DownloadedFile struct {
	SourceURL string
	FilePath  string
}

// ScanUtilsMock is an implementation of the Utils interface that can be used during tests.
type ScanUtilsMock struct {
	*mock.FilesMock
	*mock.ExecMockRunner
	NpmInstalledModules []NpmInstall
	DownloadedFiles     []DownloadedFile
}

// RemoveAll mimics os.RemoveAll().
func (m *ScanUtilsMock) RemoveAll(_ string) error {
	// Can be removed once implemented in mock.FilesMock.
	return nil
}

// FindPackageJSONFiles mimics npm.FindPackageJSONFiles() based on the FilesMock setup.
func (m *ScanUtilsMock) FindPackageJSONFiles(options *ScanOptions) ([]string, error) {
	unfilteredMatches, _ := m.Glob("**/package.json")
	return piperutils.ExcludeFiles(unfilteredMatches, options.BuildDescriptorExcludeList)
}

// InstallAllNPMDependencies mimics npm.InstallAllNPMDependencies() and records the "npm install".
func (m *ScanUtilsMock) InstallAllNPMDependencies(_ *ScanOptions, packageJSONs []string) error {
	m.NpmInstalledModules = append(m.NpmInstalledModules, NpmInstall{
		CurrentDir:  m.CurrentDir,
		PackageJSON: packageJSONs,
	})
	return nil
}

// DownloadFile mimics http.Downloader and records the downloaded file.
func (m *ScanUtilsMock) DownloadFile(url, filename string, _ http.Header, _ []*http.Cookie) error {
	m.DownloadedFiles = append(m.DownloadedFiles, DownloadedFile{SourceURL: url, FilePath: filename})
	return nil
}

// FileOpen mimics os.FileOpen() based on FilesMock Open().
func (m *ScanUtilsMock) FileOpen(name string, flag int, perm os.FileMode) (File, error) {
	return m.Open(name, flag, perm)
}

// NewScanUtilsMock returns an initialized ScanUtilsMock instance.
func NewScanUtilsMock() *ScanUtilsMock {
	return &ScanUtilsMock{
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
}
