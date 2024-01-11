//go:build unit
// +build unit

package whitesource

import (
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	t.Run("success - projects with /", func(t *testing.T) {
		// init
		options := ReportOptions{
			ReportDirectory:           "report-dir",
			VulnerabilityReportFormat: "txt",
		}
		utils := &mock.FilesMock{}
		system := NewSystemMockWithProjectName("2010-05-30 00:15:00 +0100", "@test/mock-project - 1")
		scan := &Scan{ProductVersion: "1", scannedProjects: map[string]Project{"@test/mock-project - 1": system.Projects[0]}}
		//scan := &Scan{ProductVersion: "1", scannedProjects: map[string]Project{"mock-product-token": {Name:"@test/mock-project"}}}
		//_ = scan.AppendScannedProject("@test/mock-project")
		//_ = scan.UpdateProjects("mock-product-token", system)
		// test
		paths, err := scan.DownloadReports(options, utils, system)
		// assert
		if assert.NoError(t, err) && assert.Len(t, paths, 2) {
			vPath := filepath.Join("report-dir", "@test_mock-project - 1-vulnerability-report.txt")
			assert.True(t, utils.HasWrittenFile(vPath))
			vContent, _ := utils.FileRead(vPath)
			assert.Equal(t, []byte("mock-vulnerability-report"), vContent)

			rPath := filepath.Join("report-dir", "@test_mock-project - 1-risk-report.pdf")
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
		assert.EqualError(t, err, "unable to download vulnerability report from url: no project with token '' found in Whitesource")
		assert.Nil(t, paths)
	})
	t.Run("multiple scanned projects", func(t *testing.T) {
		// init
		options := ReportOptions{
			ReportDirectory:           "report-dir",
			VulnerabilityReportFormat: "txt",
		}
		utils := &mock.FilesMock{}
		system := NewSystemMock("2010-05-30 00:15:00 +0100")
		scan := &Scan{ProductVersion: "1"}
		err := scan.AppendScannedProjectVersion("mock-project - 1")
		require.NoError(t, err)
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
}
