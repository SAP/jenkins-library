package orchestrator

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJenkins(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_URL", "FOO BAR BAZ")
		os.Setenv("BUILD_URL", "jaas.com/foo/bar/main/42")
		os.Setenv("GIT_BRANCH", "main")
		os.Setenv("GIT_COMMIT", "abcdef42713")
		os.Setenv("GIT_URL", "github.com/foo/bar")

		p, _ := NewOrchestratorSpecificConfigProvider()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "jaas.com/foo/bar/main/42", p.GetBuildUrl())
		assert.Equal(t, "main", p.GetBranch())
		assert.Equal(t, "abcdef42713", p.GetCommit())
		assert.Equal(t, "github.com/foo/bar", p.GetRepoUrl())
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("BRANCH_NAME", "PR-42")
		os.Setenv("CHANGE_BRANCH", "feat/test-jenkins")
		os.Setenv("CHANGE_TARGET", "main")
		os.Setenv("CHANGE_ID", "42")

		p := JenkinsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-jenkins", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})
}

func TestJenkinsConfigProvider_GetLog(t *testing.T) {
	defer resetEnv(os.Environ())
	os.Clearenv()
	jenkinsHome := getTempDir()
	buildNumber := "42"
	jobName := "UnitTest"
	logPath := jenkinsHome + "/jobs/" + jobName + "/builds/" + buildNumber

	os.Setenv("JENKINS_HOME", jenkinsHome)
	os.Setenv("BUILD_NUMBER", buildNumber)
	os.Setenv("JOB_NAME", jobName)
	os.MkdirAll(logPath, 0700)

	dummyLog := []byte("First line of log\nSecond line of log\nOh Snap an error!\nExiting\n")

	tests := []struct {
		name    string
		want    []byte
		wantErr bool
	}{
		{
			name:    "Get log file -  success",
			want:    dummyLog,
			wantErr: false,
		},
		{
			name:    "Get log file -  not found failure",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &JenkinsConfigProvider{}
			if !tt.wantErr {
				// create dummy log file
				err := ioutil.WriteFile(logPath+"/log", dummyLog, 0644)
				if err != nil {
					log.Entry().Errorf("Could not create dummy log file. %v", err)
				}
			}

			got, err := a.GetLog()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLog() got = %v, want %v", got, tt.want)
			}
			os.RemoveAll(jenkinsHome)
		})
	}
}

func getTempDir() string {
	tmpFolder, err := ioutil.TempDir(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
	}
	return tmpFolder
}
