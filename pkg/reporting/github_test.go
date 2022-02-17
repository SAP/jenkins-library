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
			title      string
			assignees  []string
			uploader   Uploader
		}{
			scanReport: ScanReport{Title: "testReportTitle"},
			token:      "testToken",
			apiurl:     "testApiUrl",
			owner:      "testOwner",
			repository: "testRepository",
			title:      "testTitle",
			assignees:  []string{"testAssignee1", "testAssignee2"},
			uploader:   &testUploader,
		}

		err := UploadSingleReportToGithub(testData.scanReport, testData.token, testData.apiurl, testData.owner, testData.repository, testData.title, testData.assignees, testData.uploader)

		assert.NoError(t, err)

		assert.Equal(t, testData.token, testUploader.issueOptions.Token)
		assert.Equal(t, testData.apiurl, testUploader.issueOptions.APIURL)
		assert.Equal(t, testData.owner, testUploader.issueOptions.Owner)
		assert.Equal(t, testData.repository, testUploader.issueOptions.Repository)
		assert.Equal(t, testData.title, testUploader.issueOptions.Title)
		assert.Contains(t, string(testUploader.issueOptions.Body), "testReportTitle")
		assert.Equal(t, testData.assignees, testUploader.issueOptions.Assignees)
		assert.True(t, testUploader.issueOptions.UpdateExisting)
	})

	t.Run("error case", func(t *testing.T) {
		t.Parallel()
		testUploader := mockUploader{uploadError: fmt.Errorf("upload failed")}
		err := UploadSingleReportToGithub(ScanReport{}, "", "", "", "", "", []string{}, &testUploader)

		assert.Contains(t, fmt.Sprint(err), "upload failed")
	})
}
