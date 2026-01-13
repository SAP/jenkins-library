package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/google/go-github/v68/github"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type githubActionsConfigProvider struct {
	client      *github.Client
	ctx         context.Context
	owner       string
	repo        string
	runData     run
	jobs        []job
	jobsFetched bool
}

type run struct {
	fetched   bool
	StartedAt time.Time `json:"run_started_at"`
}

// used to unmarshal list jobs of the current workflow run into []job
type job struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	HtmlURL    string `json:"html_url"`
	Conclusion string `json:"conclusion"`
}

type fullLog struct {
	sync.Mutex
	b [][]byte
}

func newGithubActionsConfigProvider() *githubActionsConfigProvider {
	owner, repo := getOwnerAndRepoNames()
	return &githubActionsConfigProvider{
		owner: owner,
		repo:  repo,
	}
}

// Configure initializes http client for GitHubActionsDevopsConfigProvider
func (g *githubActionsConfigProvider) Configure(opts *Options) error {
	var err error
	g.ctx, g.client, err = piperGithub.NewClientBuilder(opts.GitHubToken, getEnv("GITHUB_API_URL", "")).Build()
	if err != nil {
		return errors.Wrap(err, "failed to create github client")
	}

	log.Entry().Debug("Successfully initialized GitHubActions config provider")
	return nil
}

func (g *githubActionsConfigProvider) OrchestratorVersion() string {
	log.Entry().Debugf("OrchestratorVersion() for GitHub Actions is not applicable.")
	return "n/a"
}

func (g *githubActionsConfigProvider) OrchestratorType() string {
	return "GitHubActions"
}

// BuildStatus returns current run status by looking at all jobs of the current workflow run
// if any job has conclusion "failure" the whole run is considered failed
// if any job has conclusion "cancelled" the whole run is considered aborted
// otherwise the run is considered successful
func (g *githubActionsConfigProvider) BuildStatus() string {
	if err := g.fetchJobs(); err != nil {
		log.Entry().Debugf("fetching jobs: %s", err)
		return BuildStatusFailure
	}

	for _, j := range g.jobs {
		switch j.Conclusion {
		case "failure":
			return BuildStatusFailure
		case "cancelled":
			return BuildStatusAborted
		}
	}

	return BuildStatusSuccess
}

// FullLogs returns the whole logfile for the current pipeline run
func (g *githubActionsConfigProvider) FullLogs() ([]byte, error) {
	if g.client == nil {
		log.Entry().Debug("ConfigProvider for GitHub Actions is not configured. Unable to fetch logs")
		return []byte{}, nil
	}

	if err := g.fetchJobs(); err != nil {
		return nil, err
	}
	// Ignore the last stage (job) as it is not possible in GitHub to fetch logs for a running job.
	jobs := g.jobs[:len(g.jobs)-1]

	fullLogs := fullLog{b: make([][]byte, len(jobs))}
	wg := errgroup.Group{}
	wg.SetLimit(10)
	for i := range jobs {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		wg.Go(func() error {
			_, resp, err := g.client.Actions.GetWorkflowJobLogs(g.ctx, g.owner, g.repo, jobs[i].ID, 1)
			if err != nil {
				// GetWorkflowJobLogs returns "200 OK" as error when log download is successful.
				// Therefore, ignore this error.
				// GitHub API returns redirect URL instead of plain text logs. See:
				// https://docs.github.com/en/enterprise-server@3.9/rest/actions/workflow-jobs?apiVersion=2022-11-28#download-job-logs-for-a-workflow-run
				if err.Error() != "unexpected status code: 200 OK" {
					return errors.Wrap(err, "fetching job logs failed")
				}
			}
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrap(err, "failed to read response body")
			}

			fullLogs.Lock()
			fullLogs.b[i] = b
			fullLogs.Unlock()

			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, errors.Wrap(err, "failed to fetch all logs")
	}

	return bytes.Join(fullLogs.b, []byte("")), nil
}

// BuildID returns current run ID
func (g *githubActionsConfigProvider) BuildID() string {
	return getEnv("GITHUB_RUN_ID", "n/a")
}

func (g *githubActionsConfigProvider) ChangeSets() []ChangeSet {
	log.Entry().Debug("ChangeSets for GitHubActions not implemented")
	return []ChangeSet{}
}

// PipelineStartTime returns the pipeline start time in UTC
func (g *githubActionsConfigProvider) PipelineStartTime() time.Time {
	g.fetchRunData()
	return g.runData.StartedAt.UTC()
}

// StageName returns the human-readable name given to a stage.
func (g *githubActionsConfigProvider) StageName() string {
	return getEnv("GITHUB_JOB", "unknown")
}

// BuildReason returns the reason of workflow trigger.
// BuildReasons are unified with AzureDevOps build reasons, see
// https://docs.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops&tabs=yaml#build-variables-devops-services
func (g *githubActionsConfigProvider) BuildReason() string {
	switch getEnv("GITHUB_EVENT_NAME", "") {
	case "workflow_dispatch":
		return BuildReasonManual
	case "schedule":
		return BuildReasonSchedule
	case "pull_request":
		return BuildReasonPullRequest
	case "workflow_call":
		return BuildReasonResourceTrigger
	case "push":
		return BuildReasonIndividualCI
	default:
		return BuildReasonUnknown
	}
}

// Branch returns the source branch name, e.g. main
func (g *githubActionsConfigProvider) Branch() string {
	return getEnv("GITHUB_REF_NAME", "n/a")
}

// GitReference return the git reference. For example, refs/heads/your_branch_name
func (g *githubActionsConfigProvider) GitReference() string {
	return getEnv("GITHUB_REF", "n/a")
}

// BuildURL returns the builds URL. The URL should point to the pipeline (not to the stage)
// that is currently being executed. For example, https://github.com/SAP/jenkins-library/actions/runs/5815297487
func (g *githubActionsConfigProvider) BuildURL() string {
	return g.RepoURL() + "/actions/runs/" + g.BuildID()
}

// JobURL returns the job URL. The URL should point to project’s pipelines.
// For example, https://github.com/SAP/jenkins-library/actions/workflows/workflow-file-name.yaml
func (g *githubActionsConfigProvider) JobURL() string {
	fileName := workflowFileName()
	if fileName == "" {
		return ""
	}

	return g.RepoURL() + "/actions/workflows/" + fileName
}

// JobName returns the current workflow name. For example, "Piper workflow"
func (g *githubActionsConfigProvider) JobName() string {
	return getEnv("GITHUB_WORKFLOW", "unknown")
}

// CommitSHA returns the commit SHA that triggered the workflow. For example, ffac537e6cbbf934b08745a378932722df287a53
func (g *githubActionsConfigProvider) CommitSHA() string {
	return getEnv("GITHUB_SHA", "n/a")
}

// RepoURL returns full url to repository. For example, https://github.com/SAP/jenkins-library
func (g *githubActionsConfigProvider) RepoURL() string {
	return getEnv("GITHUB_SERVER_URL", "n/a") + "/" + getEnv("GITHUB_REPOSITORY", "n/a")
}

// PullRequestConfig returns pull request configuration
func (g *githubActionsConfigProvider) PullRequestConfig() PullRequestConfig {
	// See https://docs.github.com/en/enterprise-server@3.6/actions/learn-github-actions/variables#default-environment-variables
	githubRef := getEnv("GITHUB_REF", "n/a")
	prNumber := strings.TrimSuffix(strings.TrimPrefix(githubRef, "refs/pull/"), "/merge")
	return PullRequestConfig{
		Branch: getEnv("GITHUB_HEAD_REF", "n/a"),
		Base:   getEnv("GITHUB_BASE_REF", "n/a"),
		Key:    prNumber,
	}
}

// IsPullRequest indicates whether the current build is triggered by a PR
func (g *githubActionsConfigProvider) IsPullRequest() bool {
	return envVarIsTrue("GITHUB_HEAD_REF")
}

func isGitHubActions() bool {
	envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}
	return envVarsAreSet(envVars)
}

// actionsURL returns URL to actions resource. For example,
// https://api.github.com/repos/SAP/jenkins-library/actions
func actionsURL() string {
	return getEnv("GITHUB_API_URL", "") + "/repos/" + getEnv("GITHUB_REPOSITORY", "") + "/actions"
}

func (g *githubActionsConfigProvider) fetchRunData() {
	if g.client == nil {
		log.Entry().Debug("ConfigProvider for GitHub Actions is not configured. Unable to fetch run data")
		return
	}

	if g.runData.fetched {
		return
	}

	runData, resp, err := g.client.Actions.GetWorkflowRunByID(g.ctx, g.owner, g.repo, g.runIdInt64())
	if err != nil || resp.StatusCode != 200 {
		log.Entry().Errorf("failed to get API data: %s", err)
		return
	}

	g.runData = convertRunData(runData)
	g.runData.fetched = true
}

func convertRunData(runData *github.WorkflowRun) run {
	startedAtTs := piperutils.SafeDereference(runData.RunStartedAt)
	return run{
		StartedAt: startedAtTs.Time,
	}
}

func (g *githubActionsConfigProvider) fetchJobs() error {
	if g.jobsFetched {
		return nil
	}

	jobs, resp, err := g.client.Actions.ListWorkflowJobs(g.ctx, g.owner, g.repo, g.runIdInt64(), nil)
	if err != nil || resp.StatusCode != 200 {
		return errors.Wrap(err, "failed to get API data")
	}
	if len(jobs.Jobs) == 0 {
		return fmt.Errorf("no jobs found in response")
	}

	filteredJobs := filterJobs(jobs.Jobs)
	g.jobs = convertJobs(filteredJobs)
	g.jobsFetched = true

	return nil
}

// filterJobs returns only the jobs associated with a runner.
// This is necessary because fetching jobs for a workflow run triggered by a pull request
// also includes extra PR check jobs.
// This also filters out skipped jobs.
func filterJobs(jobs []*github.WorkflowJob) []*github.WorkflowJob {
	filteredJobs := make([]*github.WorkflowJob, 0, len(jobs))
	for _, j := range jobs {
		if j.GetRunnerID() != 0 {
			filteredJobs = append(filteredJobs, j)
		}
	}

	return filteredJobs
}

func convertJobs(jobs []*github.WorkflowJob) []job {
	result := make([]job, 0, len(jobs))
	for _, j := range jobs {
		result = append(result, job{
			ID:         j.GetID(),
			Name:       j.GetName(),
			HtmlURL:    j.GetHTMLURL(),
			Conclusion: j.GetConclusion(),
		})
	}
	return result
}

func (g *githubActionsConfigProvider) runIdInt64() int64 {
	strRunId := g.BuildID()
	runId, err := strconv.ParseInt(strRunId, 10, 64)
	if err != nil {
		log.Entry().Debugf("invalid GITHUB_RUN_ID value %s: %s", strRunId, err)
		return 0
	}

	return runId
}

func getOwnerAndRepoNames() (string, string) {
	ownerAndRepo := getEnv("GITHUB_REPOSITORY", "")
	s := strings.Split(ownerAndRepo, "/")
	if len(s) != 2 {
		log.Entry().Errorf("unable to determine owner and repo: invalid value of GITHUB_REPOSITORY envvar: %s", ownerAndRepo)
		return "", ""
	}

	return s[0], s[1]
}

func workflowFileName() string {
	workflowRef := getEnv("GITHUB_WORKFLOW_REF", "")
	re := regexp.MustCompile(`\.github/workflows/([a-zA-Z0-9_-]+\.(yml|yaml))`)
	matches := re.FindStringSubmatch(workflowRef)
	if len(matches) > 1 {
		return matches[1]
	}

	log.Entry().Debugf("unable to determine workflow file name from GITHUB_WORKFLOW_REF: %s", workflowRef)
	return ""
}
