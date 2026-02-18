package orchestrator

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

type jenkinsConfigProvider struct {
	client         piperHttp.Client
	apiInformation map[string]any
}

func newJenkinsConfigProvider() *jenkinsConfigProvider {
	return &jenkinsConfigProvider{}
}

// Configure initializes the Jenkins orchestrator with credentials
func (j *jenkinsConfigProvider) Configure(opts *Options) error {
	j.client.SetOptions(piperHttp.ClientOptions{
		Username:         opts.JenkinsUsername,
		Password:         opts.JenkinsToken,
		MaxRetries:       3,
		TransportTimeout: time.Second * 10,
	})

	log.Entry().Debug("Successfully initialized Jenkins config provider")
	return nil
}

// OrchestratorVersion returns the orchestrator version currently running on
func (j *jenkinsConfigProvider) OrchestratorVersion() string {
	return getEnv("JENKINS_VERSION", "n/a")
}

// OrchestratorType returns the orchestrator type Jenkins
func (j *jenkinsConfigProvider) OrchestratorType() string {
	return "Jenkins"
}

func (j *jenkinsConfigProvider) fetchAPIInformation() {
	if len(j.apiInformation) == 0 {
		log.Entry().Debugf("apiInformation is empty, getting infos from API")
		URL := j.BuildURL() + "api/json"
		log.Entry().Debugf("API URL: %s", URL)
		response, err := j.client.GetRequest(URL, nil, nil)
		if err != nil {
			log.Entry().WithError(err).Error("could not get API information from Jenkins")
			j.apiInformation = map[string]any{}
			return
		}

		if response.StatusCode != 200 { //http.StatusNoContent
			log.Entry().Errorf("Response-Code is %v, could not get timestamp from Jenkins. Setting timestamp to 1970.", response.StatusCode)
			j.apiInformation = map[string]any{}
			return
		}
		err = piperHttp.ParseHTTPResponseBodyJSON(response, &j.apiInformation)
		if err != nil {
			log.Entry().WithError(err).Errorf("could not parse HTTP response body")
			j.apiInformation = map[string]any{}
			return
		}
		log.Entry().Debugf("successfully retrieved apiInformation")
	} else {
		log.Entry().Debugf("apiInformation already set")
	}
}

// BuildStatus returns build status of the current job
func (j *jenkinsConfigProvider) BuildStatus() string {
	j.fetchAPIInformation()
	if val, ok := j.apiInformation["result"]; ok {
		// cases in ADO: succeeded, failed, canceled, none, partiallySucceeded
		switch result := val; result {
		case "SUCCESS":
			return BuildStatusSuccess
		case "ABORTED":
			return BuildStatusAborted
		default:
			// FAILURE, NOT_BUILT
			return BuildStatusFailure
		}
	}
	return BuildStatusFailure
}

// ChangeSet returns the commitIds and timestamp of the changeSet of the current run
func (j *jenkinsConfigProvider) ChangeSets() []ChangeSet {
	j.fetchAPIInformation()

	marshal, err := json.Marshal(j.apiInformation)
	if err != nil {
		log.Entry().WithError(err).Debugf("could not marshal apiInformation")
		return []ChangeSet{}
	}
	jsonParsed, err := gabs.ParseJSON(marshal)
	if err != nil {
		log.Entry().WithError(err).Debugf("could not parse apiInformation")
		return []ChangeSet{}
	}

	var changeSetList []ChangeSet
	for _, child := range jsonParsed.Path("changeSets").Children() {
		if child.Path("kind").Data().(string) == "git" {
			for _, item := range child.S("items").Children() {
				tmpChangeSet := ChangeSet{
					CommitId:  item.Path("commitId").Data().(string),
					Timestamp: item.Path("timestamp").String(),
				}
				changeSetList = append(changeSetList, tmpChangeSet)
			}
		}

	}
	return changeSetList
}

// FullLogs returns the logfile from the current job as byte object
func (j *jenkinsConfigProvider) FullLogs() ([]byte, error) {
	URL := j.BuildURL() + "consoleText"

	response, err := j.client.GetRequest(URL, nil, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("could not GET Jenkins log file %v: %w", err, err)
	} else if response.StatusCode != 200 {
		log.Entry().Error("response code !=200 could not get log information from Jenkins, returning with empty log.")
		return []byte{}, nil
	}
	logFile, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("could not read Jenkins log file from request %v: %w", err, err)
	}
	defer response.Body.Close()
	return logFile, nil
}

// PipelineStartTime returns the pipeline start time in UTC
func (j *jenkinsConfigProvider) PipelineStartTime() time.Time {
	URL := j.BuildURL() + "api/json"
	response, err := j.client.GetRequest(URL, nil, nil)
	if err != nil {
		log.Entry().WithError(err).Errorf("could not getRequest to URL %s", URL)
		return time.Time{}.UTC()
	}

	if response.StatusCode != 200 { //http.StatusNoContent -> also empty log!
		log.Entry().Errorf("response code is %v . \n Could not get timestamp from Jenkins. Setting timestamp to 1970.", response.StatusCode)
		return time.Time{}.UTC()
	}
	var responseInterface map[string]any
	err = piperHttp.ParseHTTPResponseBodyJSON(response, &responseInterface)
	if err != nil {
		log.Entry().WithError(err).Infof("could not parse http response, returning 1970")
		return time.Time{}.UTC()
	}

	rawTimeStamp := responseInterface["timestamp"].(float64)
	timeStamp := time.Unix(int64(rawTimeStamp)/1000, 0)

	log.Entry().Debugf("Pipeline start time: %v", timeStamp.String())
	defer response.Body.Close()
	return timeStamp.UTC()
}

// JobName returns the job name of the current job e.g. foo/bar/BRANCH
func (j *jenkinsConfigProvider) JobName() string {
	return getEnv("JOB_NAME", "n/a")
}

// JobURL returns the current job URL e.g. https://jaas.url/job/foo/job/bar/job/main
func (j *jenkinsConfigProvider) JobURL() string {
	return getEnv("JOB_URL", "n/a")
}

// getJenkinsHome returns the jenkins home e.g. /var/lib/jenkins
func (j *jenkinsConfigProvider) getJenkinsHome() string {
	return getEnv("JENKINS_HOME", "n/a")
}

// BuildID returns the build ID of the current job, e.g. 1234
func (j *jenkinsConfigProvider) BuildID() string {
	return getEnv("BUILD_ID", "n/a")
}

// StageName returns the stage name the job is currently in, e.g. Promote
func (j *jenkinsConfigProvider) StageName() string {
	return getEnv("STAGE_NAME", "n/a")
}

// BuildReason returns the build reason of the current build
func (j *jenkinsConfigProvider) BuildReason() string {
	// BuildReasons are unified with AzureDevOps build reasons,see
	// https://docs.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops&tabs=yaml#build-variables-devops-services
	// ResourceTrigger, PullRequest, Manual, IndividualCI, Schedule
	j.fetchAPIInformation()
	marshal, err := json.Marshal(j.apiInformation)
	if err != nil {
		log.Entry().WithError(err).Debugf("could not marshal apiInformation")
		return BuildReasonUnknown
	}
	jsonParsed, err := gabs.ParseJSON(marshal)
	if err != nil {
		log.Entry().WithError(err).Debugf("could not parse apiInformation")
		return BuildReasonUnknown
	}

	for _, child := range jsonParsed.Path("actions").Children() {
		class := child.S("_class")
		if class == nil {
			continue
		}
		if class.Data().(string) == "hudson.model.CauseAction" {
			for _, val := range child.Path("causes").Children() {
				subclass := val.S("_class")
				if subclass.Data().(string) == "hudson.model.Cause$UserIdCause" {
					return BuildReasonManual
				} else if subclass.Data().(string) == "hudson.triggers.TimerTrigger$TimerTriggerCause" {
					return BuildReasonSchedule
				} else if subclass.Data().(string) == "jenkins.branch.BranchEventCause" {
					return BuildReasonPullRequest
				} else if subclass.Data().(string) == "org.jenkinsci.plugins.workflow.support.steps.build.BuildUpstreamCause" {
					return BuildReasonResourceTrigger
				} else {
					return BuildReasonUnknown
				}
			}
		}

	}
	return BuildReasonUnknown
}

// Branch returns the branch name, only works with the git plugin enabled
func (j *jenkinsConfigProvider) Branch() string {
	return getEnv("BRANCH_NAME", "n/a")
}

// GitReference returns the git reference, only works with the git plugin enabled
func (j *jenkinsConfigProvider) GitReference() string {
	ref := getEnv("BRANCH_NAME", "n/a")
	if ref == "n/a" {
		return ref
	} else if strings.HasPrefix(ref, "PR-") {
		return "refs/pull/" + strings.Split(ref, "-")[1] + "/head"
	} else if strings.HasPrefix(ref, "refs/") {
		return ref
	} else {
		return "refs/heads/" + ref
	}
}

// BuildURL returns the build url, e.g. https://jaas.url/job/foo/job/bar/job/main/1234/
func (j *jenkinsConfigProvider) BuildURL() string {
	return getEnv("BUILD_URL", "n/a")
}

// CommitSHA returns the commit SHA from the current build, only works with the git plugin enabled
func (j *jenkinsConfigProvider) CommitSHA() string {
	return getEnv("GIT_COMMIT", "n/a")
}

// RepoURL returns the repo URL of the current build, only works with the git plugin enabled
func (j *jenkinsConfigProvider) RepoURL() string {
	return getEnv("GIT_URL", "n/a")
}

// PullRequestConfig returns the pull request config
func (j *jenkinsConfigProvider) PullRequestConfig() PullRequestConfig {
	return PullRequestConfig{
		Branch: getEnv("CHANGE_BRANCH", "n/a"),
		Base:   getEnv("CHANGE_TARGET", "n/a"),
		Key:    getEnv("CHANGE_ID", "n/a"),
	}
}

// IsPullRequest returns boolean indicating if current job is a PR
func (j *jenkinsConfigProvider) IsPullRequest() bool {
	return envVarIsTrue("CHANGE_ID")
}

func isJenkins() bool {
	envVars := []string{"JENKINS_HOME", "JENKINS_URL"}
	return envVarsAreSet(envVars)
}
