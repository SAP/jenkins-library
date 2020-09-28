package whitesource

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
	"time"
)

func TestAppendScannedProjectVersion(t *testing.T) {
	t.Parallel()
	t.Run("single module", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		// test
		err := scan.AppendScannedProjectVersion("module-a - 1")
		// assert
		assert.NoError(t, err)
		expected := make(map[string]Project)
		expected["module-a - 1"] = Project{Name: "module-a - 1"}
		assert.Equal(t, expected, scan.scannedProjects)
		_, exists := scan.scanTimes["module-a - 1"]
		assert.True(t, exists)
	})
	t.Run("two modules", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		// test
		err1 := scan.AppendScannedProjectVersion("module-a - 1")
		err2 := scan.AppendScannedProjectVersion("module-b - 1")
		// assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		expected := make(map[string]Project)
		expected["module-a - 1"] = Project{Name: "module-a - 1"}
		expected["module-b - 1"] = Project{Name: "module-b - 1"}
		assert.Equal(t, expected, scan.scannedProjects)
		_, exists := scan.scanTimes["module-b - 1"]
		assert.True(t, exists)
	})
	t.Run("module without version", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		// test
		err := scan.AppendScannedProjectVersion("module-a")
		// assert
		assert.EqualError(t, err, "projectName is expected to include the product version")
		assert.Len(t, scan.scannedProjects, 0)
	})
	t.Run("duplicate module", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		// test
		err1 := scan.AppendScannedProjectVersion("module-a - 1")
		err2 := scan.AppendScannedProjectVersion("module-a - 1")
		// assert
		assert.NoError(t, err1)
		assert.EqualError(t, err2, "project with name 'module-a - 1' was already scanned")
		expected := make(map[string]Project)
		expected["module-a - 1"] = Project{Name: "module-a - 1"}
		assert.Equal(t, expected, scan.scannedProjects)
		assert.Len(t, scan.scanTimes, 1)
	})
}

func TestAppendScannedProject(t *testing.T) {
	t.Parallel()
	t.Run("product version is appended", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		// test
		err := scan.AppendScannedProject("module-a")
		// assert
		assert.NoError(t, err)
		expected := make(map[string]Project)
		expected["module-a - 1"] = Project{Name: "module-a - 1"}
		assert.Equal(t, expected, scan.scannedProjects)
	})
}

func TestProjectByName(t *testing.T) {
	t.Parallel()
	t.Run("no init", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		// test
		project, exists := scan.ProjectByName("not there")
		// assert
		assert.False(t, exists)
		assert.Equal(t, Project{}, project)
	})
	t.Run("happy path", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		err := scan.AppendScannedProject("module-a")
		require.NoError(t, err)
		// test
		project, exists := scan.ProjectByName("module-a - 1")
		// assert
		assert.True(t, exists)
		assert.Equal(t, Project{Name: "module-a - 1"}, project)
	})
	t.Run("no such project", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		err := scan.AppendScannedProject("module-a")
		require.NoError(t, err)
		// test
		project, exists := scan.ProjectByName("not there")
		// assert
		assert.False(t, exists)
		assert.Equal(t, Project{}, project)
	})
}

func TestScannedProjects(t *testing.T) {
	t.Parallel()
	t.Run("no init", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		// test
		projects := scan.ScannedProjects()
		// assert
		assert.Len(t, projects, 0)
	})
	t.Run("single module", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("module-a")
		// test
		projects := scan.ScannedProjects()
		// assert
		assert.Len(t, projects, 1)
		assert.Contains(t, projects, Project{Name: "module-a - 1"})
	})
	t.Run("two modules", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("module-a")
		_ = scan.AppendScannedProject("module-b")
		// test
		projects := scan.ScannedProjects()
		// assert
		assert.Len(t, projects, 2)
		assert.Contains(t, projects, Project{Name: "module-a - 1"})
		assert.Contains(t, projects, Project{Name: "module-b - 1"})
	})
}

func TestScanTime(t *testing.T) {
	t.Parallel()
	t.Run("no init", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		// test
		timeStamp := scan.ScanTime("module-b - 1")
		// assert
		assert.Equal(t, time.Time{}, timeStamp)
	})
	t.Run("happy path", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("module-a")
		// test
		timeStamp := scan.ScanTime("module-a - 1")
		// assert
		assert.NotEqual(t, time.Time{}, timeStamp)
	})
	t.Run("project not scanned", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("module-a")
		// test
		timeStamp := scan.ScanTime("module-b - 1")
		// assert
		assert.Equal(t, time.Time{}, timeStamp)
	})
}

func TestScanUpdateProjects(t *testing.T) {
	t.Parallel()
	t.Run("update single project which exist", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("mock-project")
		mockSystem := NewSystemMock("just-now")
		// test
		err := scan.UpdateProjects("mock-product-token", mockSystem)
		// assert
		assert.NoError(t, err)
		expected := make(map[string]Project)
		expected["mock-project - 1"] = Project{
			Name:           "mock-project - 1",
			ID:             42,
			PluginName:     "mock-plugin-name",
			Token:          "mock-project-token",
			UploadedBy:     "MrBean",
			CreationDate:   "last-thursday",
			LastUpdateDate: "just-now",
		}
		assert.Equal(t, expected, scan.scannedProjects)
	})
	t.Run("update two projects, one of which exist", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("mock-project")
		_ = scan.AppendScannedProject("unknown-project")
		mockSystem := NewSystemMock("just-now")
		// test
		err := scan.UpdateProjects("mock-product-token", mockSystem)
		// assert
		assert.NoError(t, err, "no error expected if not all projects exist (yet)")
		expected := make(map[string]Project)
		expected["mock-project - 1"] = Project{
			Name:           "mock-project - 1",
			ID:             42,
			PluginName:     "mock-plugin-name",
			Token:          "mock-project-token",
			UploadedBy:     "MrBean",
			CreationDate:   "last-thursday",
			LastUpdateDate: "just-now",
		}
		expected["unknown-project - 1"] = Project{
			Name: "unknown-project - 1",
		}
		assert.Equal(t, expected, scan.scannedProjects)
	})
	t.Run("update single project which exist", func(t *testing.T) {
		// init
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("mock-project")
		mockSystem := &SystemMock{}
		// test
		err := scan.UpdateProjects("mock-product-token", mockSystem)
		// assert
		assert.EqualError(t, err, "failed to retrieve WhiteSource projects meta info: no product with that token")
	})
}

func TestDownloadReports(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		// init
		options := ReportOptions{
			ReportDirectory:           "report-dir",
			VulnerabilityReportFormat: "txt",
		}
		utils := &mock.FilesMock{}
		system := NewSystemMock("2010-05-30 00:15:00 +0100")
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("mock-project")
		_ = scan.UpdateProjects("mock-product-token", system)
		// test
		paths, err := scan.DownloadReports(options, utils, system)
		// assert
		if assert.NoError(t, err) && assert.Len(t, paths, 2) {
			vPath := filepath.Join("report-dir", "mock-project - 1-vulnerability-report.txt")
			assert.True(t, utils.HasWrittenFile(vPath))
			vContent, _ := utils.FileRead(vPath)
			assert.Equal(t, []byte("mock-vulnerability-report"), vContent)

			rPath := filepath.Join("report-dir", "mock-project - 1-risk-report.pdf")
			assert.True(t, utils.HasWrittenFile(rPath))
			rContent, _ := utils.FileRead(rPath)
			assert.Equal(t, []byte("mock-risk-report"), rContent)
		}
	})
	t.Run("invalid project token", func(t *testing.T) {
		// init
		options := ReportOptions{
			ReportDirectory:           "report-dir",
			VulnerabilityReportFormat: "txt",
		}
		utils := &mock.FilesMock{}
		system := NewSystemMock("2010-05-30 00:15:00 +0100")
		scan := &Scan{ProductVersion: "1"}
		_ = scan.AppendScannedProject("no-such-project")
		_ = scan.UpdateProjects("mock-product-token", system)
		// test
		paths, err := scan.DownloadReports(options, utils, system)
		// assert
		assert.EqualError(t, err, "no project with token '' found in Whitesource")
		assert.Nil(t, paths)
	})
}
