package reporting

import (
	"fmt"
	"testing"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/stretchr/testify/assert"
)

type mockUploader struct {
	issueOptions *piperGithub.CreateIssueOptions
	uploadError  error
}

func (m *mockUploader) CreateIssue(ghCreateIssueOptions *piperGithub.CreateIssueOptions) error {
	m.issueOptions = ghCreateIssueOptions
	return m.uploadError
}

type issueDetailMock struct {
	vulnerabilityType       string
	vulnerabilityName       string
	libraryName             string
	vulnerabilitySeverity   string
	vulnerabilityScore      float64
	vulnerabilityCVSS3Score float64
}

func (idm issueDetailMock) Title() string {
	return fmt.Sprintf("%v/%v/%v", idm.vulnerabilityType, idm.vulnerabilityName, idm.libraryName)
}

func (idm issueDetailMock) ToMarkdown() ([]byte, error) {
	return []byte(fmt.Sprintf(`**Vulnerability %v**
| Severity | Package | Installed Version | Description | Fix Resolution | Link |
| --- | --- | --- | --- | --- | --- |
|%v|%v|%v|%v|%v|[%v](%v)|
`, idm.vulnerabilityName, idm.vulnerabilitySeverity, idm.libraryName, "", "", "", "", "")), nil
}

func (idm issueDetailMock) ToTxt() string {
	return fmt.Sprintf(`Vulnerability %v
Severity: %v
Package: %v
Installed Version: %v
Description: %v
Fix Resolution: %v
Link: %v
`, idm.vulnerabilityName, idm.vulnerabilitySeverity, idm.libraryName, "", "", "", "")
}

func TestUploadSingleReportToGithub(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		testUploader := mockUploader{}
		testData := struct {
			scanReport ScanReport
			token      string
			apiurl     string
			owner      string
			repository string
			assignees  []string
			uploader   Uploader
		}{
			scanReport: ScanReport{ReportTitle: "testReportTitle"},
			token:      "testToken",
			apiurl:     "testApiUrl",
			owner:      "testOwner",
			repository: "testRepository",
			assignees:  []string{"testAssignee1", "testAssignee2"},
			uploader:   &testUploader,
		}

		err := UploadSingleReportToGithub(testData.scanReport, testData.token, testData.apiurl, testData.owner, testData.repository, testData.assignees, testData.uploader)

		assert.NoError(t, err)

		assert.Equal(t, testData.token, testUploader.issueOptions.Token)
		assert.Equal(t, testData.apiurl, testUploader.issueOptions.APIURL)
		assert.Equal(t, testData.owner, testUploader.issueOptions.Owner)
		assert.Equal(t, testData.repository, testUploader.issueOptions.Repository)
		assert.Equal(t, testData.scanReport.ReportTitle, testUploader.issueOptions.Title)
		assert.Contains(t, string(testUploader.issueOptions.Body), "testReportTitle")
		assert.Equal(t, testData.assignees, testUploader.issueOptions.Assignees)
		assert.True(t, testUploader.issueOptions.UpdateExisting)
	})

	t.Run("error case", func(t *testing.T) {
		t.Parallel()
		testUploader := mockUploader{uploadError: fmt.Errorf("upload failed")}
		var report IssueDetail
		report = ScanReport{}
		err := UploadSingleReportToGithub(report, "", "", "", "", []string{}, &testUploader)

		assert.Contains(t, fmt.Sprint(err), "upload failed")
	})
}

func TestUploadMultipleReportsToGithub(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		testUploader := mockUploader{}
		testData := struct {
			reports    []IssueDetail
			token      string
			apiurl     string
			owner      string
			repository string
			assignees  []string
			uploader   Uploader
		}{
			reports:    []IssueDetail{issueDetailMock{vulnerabilityType: "SECURITY_VULNERABILITY", libraryName: "test-component", vulnerabilityName: "CVE-2022001", vulnerabilitySeverity: "MEDIUM", vulnerabilityScore: 5.3}},
			token:      "testToken",
			apiurl:     "testApiUrl",
			owner:      "testOwner",
			repository: "testRepository",
			assignees:  []string{"testAssignee1", "testAssignee2"},
			uploader:   &testUploader,
		}

		err := UploadMultipleReportsToGithub(&testData.reports, testData.token, testData.apiurl, testData.owner, testData.repository, testData.assignees, []string{}, testData.uploader)

		assert.NoError(t, err)

		assert.Equal(t, testData.token, testUploader.issueOptions.Token)
		assert.Equal(t, testData.apiurl, testUploader.issueOptions.APIURL)
		assert.Equal(t, testData.owner, testUploader.issueOptions.Owner)
		assert.Equal(t, testData.repository, testUploader.issueOptions.Repository)
		assert.Equal(t, testData.reports[0].Title(), testUploader.issueOptions.Title)
		assert.Contains(t, string(testUploader.issueOptions.Body), "CVE-2022001")
		assert.Equal(t, testData.assignees, testUploader.issueOptions.Assignees)
		assert.True(t, testUploader.issueOptions.UpdateExisting)
	})

	t.Run("error case", func(t *testing.T) {
		t.Parallel()
		testUploader := mockUploader{uploadError: fmt.Errorf("upload failed")}
		reports := []IssueDetail{issueDetailMock{vulnerabilityType: "SECURITY_VULNERABILITY", libraryName: "test-component", vulnerabilityName: "CVE-2022001", vulnerabilitySeverity: "MEDIUM", vulnerabilityScore: 5.3}}
		err := UploadMultipleReportsToGithub(&reports, "", "", "", "", []string{}, []string{}, &testUploader)

		assert.Contains(t, fmt.Sprint(err), "upload failed")
	})
}
