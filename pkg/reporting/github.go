package reporting

import (
	"fmt"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
)

type Uploader interface {
	CreateIssue(ghCreateIssueOptions *piperGithub.CreateIssueOptions) error
}

func UploadSingleReportToGithub(scanReport ScanReport, token, APIURL, owner, repository, title string, assignees []string, uploader Uploader) error {
	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	markdownReport, _ := scanReport.ToMarkdown()
	options := piperGithub.CreateIssueOptions{
		Token:          token,
		APIURL:         APIURL,
		Owner:          owner,
		Repository:     repository,
		Title:          title,
		Body:           markdownReport,
		Assignees:      assignees,
		UpdateExisting: true,
	}
	err := uploader.CreateIssue(&options)
	if err != nil {
		return fmt.Errorf("failed to upload results for '%v' into GitHub issue: %w", title, err)
	}
	return nil
}
