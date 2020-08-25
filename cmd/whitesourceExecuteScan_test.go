package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/versioning"
	ws "github.com/SAP/jenkins-library/pkg/whitesource"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type whitesourceSystemMock struct {
	productName         string
	products            []ws.Product
	projects            []ws.Project
	alerts              []ws.Alert
	libraries           []ws.Library
	riskReport          []byte
	vulnerabilityReport []byte
}

func (m *whitesourceSystemMock) GetProductByName(productName string) (ws.Product, error) {
	for _, product := range m.products {
		if product.Name == productName {
			return product, nil
		}
	}
	return ws.Product{}, fmt.Errorf("no product with name '%s' found in Whitesource", productName)
}

func (m *whitesourceSystemMock) GetProjectsMetaInfo(productToken string) ([]ws.Project, error) {
	return m.projects, nil
}

func (m *whitesourceSystemMock) GetProjectToken(productToken, projectName string) (string, error) {
	return "mock-project-token", nil
}

func (m *whitesourceSystemMock) GetProjectByToken(projectToken string) (ws.Project, error) {
	for _, project := range m.projects {
		if project.Token == projectToken {
			return project, nil
		}
	}
	return ws.Project{}, fmt.Errorf("no project with token '%s' found in Whitesource", projectToken)
}

func (m *whitesourceSystemMock) GetProjectRiskReport(projectToken string) ([]byte, error) {
	return m.riskReport, nil
}

func (m *whitesourceSystemMock) GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error) {
	_, err := m.GetProjectByToken(projectToken)
	if err != nil {
		return nil, err
	}
	if m.vulnerabilityReport == nil {
		return nil, fmt.Errorf("no report available")
	}
	return m.vulnerabilityReport, nil
}

func (m *whitesourceSystemMock) GetProjectAlerts(projectToken string) ([]ws.Alert, error) {
	return m.alerts, nil
}

func (m *whitesourceSystemMock) GetProjectLibraryLocations(projectToken string) ([]ws.Library, error) {
	return m.libraries, nil
}

var mockLibrary = ws.Library{
	Name:     "mock-library",
	Filename: "mock-library-file",
	Version:  "mock-library-version",
	Project:  "mock-project",
}

func newWhitesourceSystemMock(lastUpdateDate string) *whitesourceSystemMock {
	return &whitesourceSystemMock{
		productName: "mock-product",
		products: []ws.Product{
			{
				Name:           "mock-product",
				Token:          "mock-product-token",
				CreationDate:   "last-thursday",
				LastUpdateDate: lastUpdateDate,
			},
		},
		projects: []ws.Project{
			{
				ID:             42,
				Name:           "mock-project",
				PluginName:     "mock-plugin-name",
				Token:          "mock-project-token",
				UploadedBy:     "MrBean",
				CreationDate:   "last-thursday",
				LastUpdateDate: lastUpdateDate,
			},
		},
		alerts: []ws.Alert{
			{
				Vulnerability: ws.Vulnerability{},
				Library:       mockLibrary,
				Project:       "mock-project",
				CreationDate:  "last-thursday",
			},
		},
		libraries:           []ws.Library{mockLibrary},
		riskReport:          []byte("mock-risk-report"),
		vulnerabilityReport: []byte("mock-vulnerability-report"),
	}
}

type coordinatesMock struct {
	GroupID    string
	ArtifactID string
	Version    string
}

type downloadedFile struct {
	sourceURL string
	filePath  string
}

type whitesourceUtilsMock struct {
	*mock.FilesMock
	*mock.ExecMockRunner
	coordinates     coordinatesMock
	downloadedFiles []downloadedFile
}

func (w *whitesourceUtilsMock) DownloadFile(url, filename string, _ http.Header, _ []*http.Cookie) error {
	w.downloadedFiles = append(w.downloadedFiles, downloadedFile{sourceURL: url, filePath: filename})
	return nil
}

func (w *whitesourceUtilsMock) FileOpen(name string, flag int, perm os.FileMode) (*os.File, error) {
	return nil, fmt.Errorf("FileOpen() is unsupported by the mock implementation")
}

func (w *whitesourceUtilsMock) RemoveAll(path string) error {
	return nil
}

func (w *whitesourceUtilsMock) GetArtifactCoordinates(_ *ScanOptions) (versioning.Coordinates, error) {
	return w.coordinates, nil
}

func (w *whitesourceUtilsMock) FindPackageJSONFiles(_ *ScanOptions) []string {
	matches, _ := w.Glob("**/package.json")
	return matches
}

func (w *whitesourceUtilsMock) InstallAllNPMDependencies(_ *ScanOptions, _ []string) error {
	return nil
}

func newWhitesourceUtilsMock() *whitesourceUtilsMock {
	return &whitesourceUtilsMock{
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
		coordinates: coordinatesMock{
			GroupID:    "mock-group-id",
			ArtifactID: "mock-artifact-id",
			Version:    "1.0.42",
		},
	}
}

func TestResolveProjectIdentifiers(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// init
		config := ScanOptions{
			ScanType:               "mta",
			DefaultVersioningModel: "major",
			ProductName:            "mock-product",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := newWhitesourceSystemMock("ignored")
		// test
		err := resolveProjectIdentifiers(&config, utilsMock, systemMock)
		// assert
		if assert.NoError(t, err) {
			assert.Equal(t, "mock-group-id-mock-artifact-id", config.ProjectName)
			assert.Equal(t, "1", config.ProductVersion)
			assert.Equal(t, "mock-project-token", config.ProjectToken)
			assert.Equal(t, "mock-product-token", config.ProductToken)
		}
	})
	t.Run("product not found", func(t *testing.T) {
		// init
		config := ScanOptions{
			ScanType:               "mta",
			DefaultVersioningModel: "major",
			ProductName:            "does-not-exist",
		}
		utilsMock := newWhitesourceUtilsMock()
		systemMock := newWhitesourceSystemMock("ignored")
		// test
		err := resolveProjectIdentifiers(&config, utilsMock, systemMock)
		// assert
		assert.EqualError(t, err, "no product with name 'does-not-exist' found in Whitesource")
	})
}

func TestBlockUntilProjectIsUpdated(t *testing.T) {
	t.Run("already new enough", func(t *testing.T) {
		// init
		nowString := "2010-05-30 00:15:00 +0100"
		now, err := time.Parse(whitesourceDateTimeLayout, nowString)
		if err != nil {
			t.Fatalf(err.Error())
		}
		lastUpdatedDate := "2010-05-30 00:15:01 +0100"
		systemMock := newWhitesourceSystemMock(lastUpdatedDate)
		config := &ScanOptions{
			ProjectToken: systemMock.projects[0].Token,
		}
		// test
		err = blockUntilProjectIsUpdated(config, systemMock, now, 2*time.Second, 1*time.Second, 2*time.Second)
		// assert
		assert.NoError(t, err)
	})
	t.Run("timeout while polling", func(t *testing.T) {
		// init
		nowString := "2010-05-30 00:15:00 +0100"
		now, err := time.Parse(whitesourceDateTimeLayout, nowString)
		if err != nil {
			t.Fatalf(err.Error())
		}
		lastUpdatedDate := "2010-05-30 00:07:00 +0100"
		systemMock := newWhitesourceSystemMock(lastUpdatedDate)
		config := &ScanOptions{
			ProjectToken: systemMock.projects[0].Token,
		}
		// test
		err = blockUntilProjectIsUpdated(config, systemMock, now, 2*time.Second, 1*time.Second, 1*time.Second)
		// assert
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "timeout while waiting")
		}
	})
}

func TestDownloadReports(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		// init
		config := &ScanOptions{
			ProjectToken:              "mock-project-token",
			ProjectName:               "mock-project",
			ReportDirectoryName:       "report-dir",
			VulnerabilityReportFormat: "txt",
		}
		utils := newWhitesourceUtilsMock()
		system := newWhitesourceSystemMock("2010-05-30 00:15:00 +0100")
		// test
		paths, err := downloadReports(config, utils, system)
		// assert
		if assert.NoError(t, err) && assert.Len(t, paths, 2) {
			vPath := filepath.Join("report-dir", "mock-project-vulnerability-report.txt")
			assert.True(t, utils.HasWrittenFile(vPath))
			vContent, _ := utils.FileRead(vPath)
			assert.Equal(t, []byte("mock-vulnerability-report"), vContent)

			rPath := filepath.Join("report-dir", "mock-project-risk-report.pdf")
			assert.True(t, utils.HasWrittenFile(rPath))
			rContent, _ := utils.FileRead(rPath)
			assert.Equal(t, []byte("mock-risk-report"), rContent)
		}
	})
	t.Run("invalid project token", func(t *testing.T) {
		// init
		config := &ScanOptions{
			ProjectToken: "<invalid>",
		}
		utils := newWhitesourceUtilsMock()
		system := newWhitesourceSystemMock("2010-05-30 00:15:00 +0100")
		// test
		path, err := downloadReports(config, utils, system)
		// assert
		assert.EqualError(t, err, "no project with token '<invalid>' found in Whitesource")
		assert.Nil(t, path)
	})
}

func TestWriteWhitesourceConfigJSON(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// init
		config := &ScanOptions{
			OrgToken:       "org-token",
			UserToken:      "user-token",
			ProductName:    "mock-product",
			ProjectName:    "mock-project",
			ProductVersion: "42",
		}
		utils := newWhitesourceUtilsMock()
		// test
		err := writeWhitesourceConfigJSON(config, utils, true)
		// assert
		if assert.NoError(t, err) && assert.True(t, utils.HasWrittenFile(whiteSourceConfig)) {
			contents, _ := utils.FileRead(whiteSourceConfig)
			expected := `{
		"apiKey": "org-token",
		"userKey": "user-token",
		"checkPolicies": true,
		"productName": "mock-product",
		"projectName": "mock-project",
		"productVer": "42",
		"devDep": true,
		"ignoreNpmLsErrors": true
	}`
			assert.Equal(t, expected, string(contents))
		}
	})
}
