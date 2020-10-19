// +build !release

package whitesource

import "github.com/SAP/jenkins-library/pkg/mock"

func newTestScan(config *ScanOptions) *Scan {
	return &Scan{
		AggregateProjectName: config.ProjectName,
		ProductVersion:       "product-version",
	}
}

type npmInstall struct {
	currentDir  string
	packageJSON []string
}

type scanUtilsMock struct {
	*mock.FilesMock
	*mock.ExecMockRunner
	npmInstalledModules []npmInstall
}

func (m *scanUtilsMock) RemoveAll(path string) error {
	// TODO: Implement in FS Mock
	return nil
}

func (m *scanUtilsMock) FindPackageJSONFiles(_ *ScanOptions) ([]string, error) {
	matches, _ := m.Glob("**/package.json")
	return matches, nil
}

func (m *scanUtilsMock) InstallAllNPMDependencies(_ *ScanOptions, packageJSONs []string) error {
	m.npmInstalledModules = append(m.npmInstalledModules, npmInstall{
		currentDir:  m.CurrentDir,
		packageJSON: packageJSONs,
	})
	return nil
}

func newScanUtilsMock() *scanUtilsMock {
	return &scanUtilsMock{
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
}
