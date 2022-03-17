package reporting

import (
	"fmt"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

type Uploader interface {
	CreateIssue(ghCreateIssueOptions *piperGithub.CreateIssueOptions) error
}

// UploadSingleReportToGithub uploads a single report to GitHub
func UploadSingleReportToGithub(scanReport IssueDetail, token, APIURL, owner, repository string, assignees []string, uploader Uploader) error {
	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	markdownReport, _ := scanReport.ToMarkdown()
	options := piperGithub.CreateIssueOptions{
		Token:          token,
		APIURL:         APIURL,
		Owner:          owner,
		Repository:     repository,
		Title:          scanReport.Title(),
		Body:           markdownReport,
		Assignees:      assignees,
		UpdateExisting: true,
	}
	err := uploader.CreateIssue(&options)
	if err != nil {
		return fmt.Errorf("failed to upload results for '%v' into GitHub issue: %w", scanReport.Title(), err)
	}
	return nil
}

// UploadMultipleReportsToGithub uploads a number of reports to GitHub, one per IssueDetail to create transparency
func UploadMultipleReportsToGithub(scanReports *[]IssueDetail, token, APIURL, owner, repository string, assignees, trustedCerts []string) error {
	for i := 0; i < len(*scanReports); i++ {
		vuln := (*scanReports)[i]
		title := vuln.Title()
		markdownReport, _ := vuln.ToMarkdown()
		options := piperGithub.CreateIssueOptions{
			Token:          token,
			APIURL:         APIURL,
			Owner:          owner,
			Repository:     repository,
			Title:          title,
			Body:           markdownReport,
			Assignees:      assignees,
			UpdateExisting: true,
			TrustedCerts:   trustedCerts,
		}

		log.Entry().Debugf("Creating/updating GitHub issue(s) with title %v in org %v and repo %v", title, owner, repository)
		err := piperGithub.CreateIssue(&options)
		if err != nil {
			return errors.Wrapf(err, "Failed to upload WhiteSource result for %v into GitHub issue", vuln.Title())
		}
	}

	return nil
}