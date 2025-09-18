package cmd

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type artifactVersioningMock struct {
	originalVersion  string
	newVersion       string
	getVersionError  string
	setVersionError  string
	versioningScheme string
	coordinates      versioning.Coordinates
	coordinatesError error
}

func (a *artifactVersioningMock) VersioningScheme() string {
	return a.versioningScheme
}

func (a *artifactVersioningMock) GetVersion() (string, error) {
	if len(a.getVersionError) > 0 {
		return "", fmt.Errorf("%s", a.getVersionError)
	}
	return a.originalVersion, nil
}

func (a *artifactVersioningMock) SetVersion(version string) error {
	if len(a.setVersionError) > 0 {
		return fmt.Errorf("%s", a.setVersionError)
	}
	a.newVersion = version
	return nil
}

func (a *artifactVersioningMock) GetCoordinates() (versioning.Coordinates, error) {
	if a.coordinatesError != nil {
		return versioning.Coordinates{}, a.coordinatesError
	}
	return a.coordinates, nil
}

type gitRepositoryMock struct {
	createRemoteConfigs []*gitConfig.RemoteConfig
	createRemoteCalls   int
	createRemoteError   []string
	deleteRemoteNames   []string
	deleteRemoteCalls   int
	deleteRemoteError   []string
	pushCalled          bool
	pushOptions         *git.PushOptions
	pushError           string
	remote              *git.Remote
	remoteError         string
	revision            string
	revisionHash        plumbing.Hash
	revisionError       string
	tag                 string
	tagHash             plumbing.Hash
	tagError            string
	worktree            *git.Worktree
	worktreeError       string
	commitObjectHash    string
}

func (r *gitRepositoryMock) CommitObject(hash plumbing.Hash) (*object.Commit, error) {
	r.commitObjectHash = hash.String()
	return &object.Commit{Hash: hash, Message: "Test commit message"}, nil
}

func (r *gitRepositoryMock) CreateTag(name string, hash plumbing.Hash, opts *git.CreateTagOptions) (*plumbing.Reference, error) {
	if len(r.tagError) > 0 {
		return nil, fmt.Errorf("%s", r.tagError)
	}
	r.tag = name
	r.tagHash = hash
	return nil, nil
}

func (r *gitRepositoryMock) CreateRemote(config *gitConfig.RemoteConfig) (*git.Remote, error) {
	r.createRemoteCalls++
	if len(r.createRemoteError) >= r.createRemoteCalls && len(r.createRemoteError[r.createRemoteCalls-1]) > 0 {
		return nil, fmt.Errorf("%s", r.createRemoteError[r.createRemoteCalls-1])
	}
	r.createRemoteConfigs = append(r.createRemoteConfigs, config)
	return nil, nil
}

func (r *gitRepositoryMock) DeleteRemote(name string) error {
	r.deleteRemoteCalls++
	if len(r.deleteRemoteError) >= r.deleteRemoteCalls && len(r.deleteRemoteError[r.deleteRemoteCalls-1]) > 0 {
		return fmt.Errorf("%s", r.deleteRemoteError[r.deleteRemoteCalls-1])
	}
	r.deleteRemoteNames = append(r.deleteRemoteNames, name)
	return nil
}

func (r *gitRepositoryMock) Push(o *git.PushOptions) error {
	if len(r.pushError) > 0 {
		return fmt.Errorf("%s", r.pushError)
	}
	r.pushCalled = true
	r.pushOptions = o
	return nil
}

func (r *gitRepositoryMock) Remote(name string) (*git.Remote, error) {
	if len(r.remoteError) > 0 {
		return &git.Remote{}, fmt.Errorf("%s", r.remoteError)
	}
	return r.remote, nil
}

func (r *gitRepositoryMock) ResolveRevision(rev plumbing.Revision) (*plumbing.Hash, error) {
	if len(r.revisionError) > 0 {
		return nil, fmt.Errorf("%s", r.revisionError)
	}
	r.revision = rev.String()
	return &r.revisionHash, nil
}

func (r *gitRepositoryMock) Worktree() (*git.Worktree, error) {
	if len(r.worktreeError) > 0 {
		return nil, fmt.Errorf("%s", r.worktreeError)
	}
	return r.worktree, nil
}

type gitWorktreeMock struct {
	checkoutError string
	checkoutOpts  *git.CheckoutOptions
	commitHash    plumbing.Hash
	commitMsg     string
	commitOpts    *git.CommitOptions
	commitError   string
}

func (w *gitWorktreeMock) Checkout(opts *git.CheckoutOptions) error {
	if len(w.checkoutError) > 0 {
		return fmt.Errorf("%s", w.checkoutError)
	}
	w.checkoutOpts = opts
	return nil
}
func (w *gitWorktreeMock) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	if len(w.commitError) > 0 {
		return plumbing.Hash{}, fmt.Errorf("%s", w.commitError)
	}
	w.commitMsg = msg
	w.commitOpts = opts
	return w.commitHash, nil
}

type artifactPrepareVersionMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.HttpClientMock
}

func newArtifactPrepareVersionMockUtils() *artifactPrepareVersionMockUtils {
	utils := artifactPrepareVersionMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return &utils
}

func (a *artifactPrepareVersionMockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	// so far no dedicated logic required for testing
	return nil
}

func (a *artifactPrepareVersionMockUtils) GetConfigProvider() (orchestrator.ConfigProvider, error) {
	return &orchestrator.UnknownOrchestratorConfigProvider{}, nil
}

func TestRunArtifactPrepareVersion(t *testing.T) {

	t.Run("success case - cloud", func(t *testing.T) {

		config := artifactPrepareVersionOptions{
			BuildTool:       "maven",
			IncludeCommitID: true,
			Password:        "****",
			TagPrefix:       "v",
			Username:        "testUser",
			VersioningType:  "cloud",
		}
		telemetryData := telemetry.CustomData{}

		cpe := artifactPrepareVersionCommonPipelineEnvironment{}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
		}

		utils := newArtifactPrepareVersionMockUtils()

		worktree := gitWorktreeMock{
			commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{2, 3, 4}),
		}

		conf := gitConfig.RemoteConfig{Name: "origin", URLs: []string{"https://my.test.server"}}

		repo := gitRepositoryMock{
			revisionHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3}),
			remote:       git.NewRemote(nil, &conf),
		}

		err := runArtifactPrepareVersion(&config, &telemetryData, &cpe, &versioningMock, utils, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })

		assert.NoError(t, err)

		assert.Contains(t, versioningMock.newVersion, "1.2.3")
		assert.Contains(t, versioningMock.newVersion, fmt.Sprintf("_%v", repo.revisionHash.String()))

		assert.Equal(t, "HEAD", repo.revision)
		assert.Contains(t, repo.tag, "v1.2.3")
		assert.Equal(t, &git.CheckoutOptions{Hash: repo.revisionHash, Keep: true}, worktree.checkoutOpts)
		assert.True(t, repo.pushCalled)

		assert.Contains(t, cpe.artifactVersion, "1.2.3")
		assert.Contains(t, cpe.originalArtifactVersion, "1.2.3")
		assert.Equal(t, worktree.commitHash.String(), cpe.git.commitID)
		assert.Equal(t, "Test commit message", cpe.git.commitMessage)

		assert.Equal(t, telemetry.CustomData{BuildTool: "maven", FilePath: ""}, telemetryData)
	})

	t.Run("success case - cloud_noTag", func(t *testing.T) {

		config := artifactPrepareVersionOptions{
			BuildTool:       "maven",
			IncludeCommitID: true,
			Password:        "****",
			TagPrefix:       "v",
			Username:        "testUser",
			VersioningType:  "cloud_noTag",
		}
		telemetryData := telemetry.CustomData{}

		cpe := artifactPrepareVersionCommonPipelineEnvironment{}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
		}

		utils := newArtifactPrepareVersionMockUtils()

		worktree := gitWorktreeMock{
			commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{2, 3, 4}),
		}

		conf := gitConfig.RemoteConfig{Name: "origin", URLs: []string{"https://my.test.server"}}

		repo := gitRepositoryMock{
			revisionHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3}),
			remote:       git.NewRemote(nil, &conf),
		}

		err := runArtifactPrepareVersion(&config, &telemetryData, &cpe, &versioningMock, utils, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })

		assert.NoError(t, err)

		assert.False(t, repo.pushCalled)
		assert.Contains(t, cpe.artifactVersion, "1.2.3")
		assert.Contains(t, cpe.originalArtifactVersion, "1.2.3")
		assert.Equal(t, repo.revisionHash.String(), cpe.git.commitID)
	})

	t.Run("success case - compatibility", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			BuildTool:          "maven",
			VersioningType:     "cloud",
			VersioningTemplate: "${version}",
		}

		cpe := artifactPrepareVersionCommonPipelineEnvironment{}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
		}

		worktree := gitWorktreeMock{}
		repo := gitRepositoryMock{}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &cpe, &versioningMock, nil, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", cpe.artifactVersion)
	})

	t.Run("success case - library", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			BuildTool:      "maven",
			VersioningType: "library",
		}

		cpe := artifactPrepareVersionCommonPipelineEnvironment{}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
		}

		worktree := gitWorktreeMock{
			commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{2, 3, 4}),
		}
		repo := gitRepositoryMock{
			revisionHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3}),
		}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &cpe, &versioningMock, nil, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", cpe.artifactVersion)
		assert.Equal(t, repo.revisionHash.String(), cpe.git.commitID)
	})

	t.Run("success case - coordinates", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			BuildTool:        "maven",
			VersioningType:   "library",
			FetchCoordinates: true,
		}

		cpe := artifactPrepareVersionCommonPipelineEnvironment{}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
			coordinates:      versioning.Coordinates{GroupID: "my.testGroup", ArtifactID: "testArtifact", Packaging: "testPackaging"},
		}

		worktree := gitWorktreeMock{
			commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{2, 3, 4}),
		}
		repo := gitRepositoryMock{
			revisionHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3}),
		}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &cpe, &versioningMock, nil, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })

		assert.NoError(t, err)
		assert.Equal(t, "testArtifact", cpe.artifactID)
		assert.Equal(t, "my.testGroup", cpe.groupID)
		assert.Equal(t, "testPackaging", cpe.packaging)
	})

	t.Run("error - failed to retrieve version", func(t *testing.T) {
		config := artifactPrepareVersionOptions{}

		versioningMock := artifactVersioningMock{
			getVersionError: "getVersion error",
		}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, nil, &versioningMock, nil, nil, nil)
		assert.EqualError(t, err, "failed to retrieve version: getVersion error")

	})

	t.Run("error - failed to retrieve git commit ID", func(t *testing.T) {
		config := artifactPrepareVersionOptions{}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
		}

		repo := gitRepositoryMock{revisionError: "revision error"}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, nil, &versioningMock, nil, &repo, nil)
		assert.EqualError(t, err, "failed to retrieve git commit ID: revision error")
	})

	t.Run("error - versioning template", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			VersioningType: "cloud",
		}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "notSupported",
		}

		utils := newArtifactPrepareVersionMockUtils()

		repo := gitRepositoryMock{}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &artifactPrepareVersionCommonPipelineEnvironment{}, &versioningMock, utils, &repo, nil)
		assert.Contains(t, fmt.Sprint(err), "failed to get versioning template for scheme 'notSupported'")
	})

	t.Run("error - failed to retrieve git worktree", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			VersioningType: "cloud",
		}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
		}

		utils := newArtifactPrepareVersionMockUtils()

		repo := gitRepositoryMock{}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &artifactPrepareVersionCommonPipelineEnvironment{}, &versioningMock, utils, &repo, func(r gitRepository) (gitWorktree, error) { return nil, fmt.Errorf("worktree error") })
		assert.EqualError(t, err, "failed to retrieve git worktree: worktree error")
	})

	t.Run("error - failed to initialize git worktree: ", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			VersioningType: "cloud",
		}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
		}

		utils := newArtifactPrepareVersionMockUtils()

		worktree := gitWorktreeMock{checkoutError: "checkout error"}
		repo := gitRepositoryMock{}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &artifactPrepareVersionCommonPipelineEnvironment{}, &versioningMock, utils, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })
		assert.EqualError(t, err, "failed to initialize worktree: checkout error")
	})

	t.Run("error - failed to set version", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			VersioningType: "cloud",
		}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			setVersionError:  "setVersion error",
			versioningScheme: "maven",
		}

		utils := newArtifactPrepareVersionMockUtils()

		worktree := gitWorktreeMock{}
		repo := gitRepositoryMock{}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &artifactPrepareVersionCommonPipelineEnvironment{}, &versioningMock, utils, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })
		assert.EqualError(t, err, "failed to write version: setVersion error")
	})

	t.Run("error - failed to push changes", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			VersioningType: "cloud",
		}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
		}

		utils := newArtifactPrepareVersionMockUtils()

		worktree := gitWorktreeMock{}
		repo := gitRepositoryMock{}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &artifactPrepareVersionCommonPipelineEnvironment{}, &versioningMock, utils, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })
		assert.Contains(t, fmt.Sprint(err), "failed to push changes for version '1.2.3")
	})

	t.Run("error - failed to get coordinates", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			BuildTool:        "maven",
			VersioningType:   "library",
			FetchCoordinates: true,
		}

		cpe := artifactPrepareVersionCommonPipelineEnvironment{}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
			coordinatesError: fmt.Errorf("coordinatesError"),
		}

		utils := newArtifactPrepareVersionMockUtils()

		worktree := gitWorktreeMock{
			commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{2, 3, 4}),
		}
		repo := gitRepositoryMock{
			revisionHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3}),
		}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &cpe, &versioningMock, utils, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })

		assert.EqualError(t, err, "failed to get coordinates: coordinatesError")
	})

	t.Run("warning - failed to get coordinates", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			BuildTool:        "maven",
			VersioningType:   "library",
			FetchCoordinates: false,
		}

		cpe := artifactPrepareVersionCommonPipelineEnvironment{}

		versioningMock := artifactVersioningMock{
			originalVersion:  "1.2.3",
			versioningScheme: "maven",
			coordinatesError: fmt.Errorf("coordinatesError"),
		}

		utils := newArtifactPrepareVersionMockUtils()

		worktree := gitWorktreeMock{
			commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{2, 3, 4}),
		}
		repo := gitRepositoryMock{
			revisionHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3}),
		}

		err := runArtifactPrepareVersion(&config, &telemetry.CustomData{}, &cpe, &versioningMock, utils, &repo, func(r gitRepository) (gitWorktree, error) { return &worktree, nil })

		assert.NoError(t, err)
	})
}

func TestVersioningTemplate(t *testing.T) {
	tt := []struct {
		scheme      string
		expected    string
		expectedErr string
	}{
		{scheme: "maven", expected: "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}_{{.CommitID}}{{end}}{{end}}"},
		{scheme: "semver2", expected: "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}+{{.CommitID}}{{end}}{{end}}"},
		{scheme: "pep440", expected: "{{.Version}}{{if .Timestamp}}.{{.Timestamp}}{{if .CommitID}}+{{.CommitID}}{{end}}{{end}}"},
		{scheme: "notSupported", expected: "", expectedErr: "versioning scheme 'notSupported' not supported"},
	}

	for _, test := range tt {
		scheme, err := versioningTemplate(test.scheme)
		assert.Equal(t, test.expected, scheme)
		if len(test.expectedErr) == 0 {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, test.expectedErr)
		}
	}
}

func TestCalculateNewVersion(t *testing.T) {

	currentVersion := "1.2.3"
	testTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	commitID := plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3}).String()

	tt := []struct {
		versioningTemplate string
		includeCommitID    bool
		shortCommitID      bool
		unixTimestamp      bool
		expected           string
		expectedErr        string
	}{
		{versioningTemplate: "", expectedErr: "failed calculate version, new version is ''"},
		{versioningTemplate: "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}+{{.CommitID}}{{end}}{{end}}", expected: "1.2.3-20200101000000"},
		{versioningTemplate: "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}+{{.CommitID}}{{end}}{{end}}", includeCommitID: true, expected: "1.2.3-20200101000000+428ecf70bc22df0ba3dcf194b5ce53e769abab07"},
		{versioningTemplate: "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}+{{.CommitID}}{{end}}{{end}}", includeCommitID: true, shortCommitID: true, expected: "1.2.3-20200101000000+428ecf7"},
		{versioningTemplate: "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}+{{.CommitID}}{{end}}{{end}}", includeCommitID: true, unixTimestamp: true, expected: "1.2.3-1577836800+428ecf70bc22df0ba3dcf194b5ce53e769abab07"},
	}

	for _, test := range tt {
		version, err := calculateNewVersion(test.versioningTemplate, currentVersion, commitID, test.includeCommitID, test.shortCommitID, test.unixTimestamp, testTime)
		assert.Equal(t, test.expected, version)
		if len(test.expectedErr) == 0 {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, test.expectedErr)
		}
	}
}

func TestPushChanges(t *testing.T) {

	newVersion := "1.2.3"
	testTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	conf := gitConfig.RemoteConfig{Name: "origin", URLs: []string{"https://my.test.server"}}
	remote := git.NewRemote(nil, &conf)

	t.Run("success - username/password", func(t *testing.T) {
		config := artifactPrepareVersionOptions{Username: "testUser", Password: "****", CommitUserName: "Project Piper"}
		repo := gitRepositoryMock{remote: remote}
		worktree := gitWorktreeMock{commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})}

		commitID, err := pushChanges(&config, newVersion, &repo, &worktree, testTime, nil)
		assert.NoError(t, err)
		assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", commitID)
		assert.Equal(t, "update version 1.2.3", worktree.commitMsg)
		assert.Equal(t, &git.CommitOptions{All: true, AllowEmptyCommits: true, Author: &object.Signature{Name: "Project Piper", When: testTime}}, worktree.commitOpts)
		assert.Equal(t, "1.2.3", repo.tag)
		assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", repo.tagHash.String())
		assert.Equal(t, &git.PushOptions{RefSpecs: []gitConfig.RefSpec{"refs/tags/1.2.3:refs/tags/1.2.3"}, Auth: &gitHttp.BasicAuth{Username: config.Username, Password: config.Password}}, repo.pushOptions)
	})

	t.Run("success - ssh fallback", func(t *testing.T) {
		config := artifactPrepareVersionOptions{CommitUserName: "Project Piper"}
		repo := gitRepositoryMock{remote: remote}
		worktree := gitWorktreeMock{commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})}
		customCerts := []byte("custom certs")

		originalSSHAgentAuth := sshAgentAuth
		sshAgentAuth = func(u string) (*ssh.PublicKeysCallback, error) { return &ssh.PublicKeysCallback{}, nil }
		commitID, err := pushChanges(&config, newVersion, &repo, &worktree, testTime, customCerts)
		sshAgentAuth = originalSSHAgentAuth

		assert.NoError(t, err)
		assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", commitID)
		assert.Equal(t, "update version 1.2.3", worktree.commitMsg)
		assert.Equal(t, &git.CommitOptions{All: true, AllowEmptyCommits: true, Author: &object.Signature{Name: "Project Piper", When: testTime}}, worktree.commitOpts)
		assert.Equal(t, "1.2.3", repo.tag)
		assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", repo.tagHash.String())
		assert.Equal(t, &git.PushOptions{RefSpecs: []gitConfig.RefSpec{"refs/tags/1.2.3:refs/tags/1.2.3"}, Auth: &ssh.PublicKeysCallback{}, CABundle: customCerts}, repo.pushOptions)
	})

	t.Run("success - ssh", func(t *testing.T) {
		confSSH := gitConfig.RemoteConfig{Name: "origin", URLs: []string{"git@my.test.server"}}
		remoteSSH := git.NewRemote(nil, &confSSH)

		config := artifactPrepareVersionOptions{}
		repo := gitRepositoryMock{remote: remoteSSH}
		worktree := gitWorktreeMock{commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})}

		originalSSHAgentAuth := sshAgentAuth
		sshAgentAuth = func(u string) (*ssh.PublicKeysCallback, error) { return &ssh.PublicKeysCallback{}, nil }
		commitID, err := pushChanges(&config, newVersion, &repo, &worktree, testTime, nil)
		sshAgentAuth = originalSSHAgentAuth

		assert.NoError(t, err)
		assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", commitID)
		assert.Equal(t, &git.PushOptions{RefSpecs: []gitConfig.RefSpec{"refs/tags/1.2.3:refs/tags/1.2.3"}, Auth: &ssh.PublicKeysCallback{}}, repo.pushOptions)
	})

	t.Run("error - commit", func(t *testing.T) {
		config := artifactPrepareVersionOptions{}
		repo := gitRepositoryMock{}
		worktree := gitWorktreeMock{commitError: "commit error", commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})}

		commitID, err := pushChanges(&config, newVersion, &repo, &worktree, testTime, nil)
		assert.Equal(t, "0000000000000000000000000000000000000000", commitID)
		assert.EqualError(t, err, "failed to commit new version: commit error")
	})

	t.Run("error - create tag", func(t *testing.T) {
		config := artifactPrepareVersionOptions{}
		repo := gitRepositoryMock{tagError: "tag error"}
		worktree := gitWorktreeMock{commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})}

		commitID, err := pushChanges(&config, newVersion, &repo, &worktree, testTime, nil)
		assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", commitID)
		assert.EqualError(t, err, "tag error")
	})

	t.Run("error - no remote url", func(t *testing.T) {
		config := artifactPrepareVersionOptions{}
		repo := gitRepositoryMock{}
		worktree := gitWorktreeMock{commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})}

		commitID, err := pushChanges(&config, newVersion, &repo, &worktree, testTime, nil)
		assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", commitID)
		assert.EqualError(t, err, "no remote url maintained")
	})

	t.Run("error - ssh fallback", func(t *testing.T) {

		config := artifactPrepareVersionOptions{}
		worktree := gitWorktreeMock{commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})}

		sshSuccess := func(u string) (*ssh.PublicKeysCallback, error) { return nil, nil }
		sshFailure := func(u string) (*ssh.PublicKeysCallback, error) { return nil, fmt.Errorf("ssh error") }

		tt := []struct {
			repo          gitRepositoryMock
			sshAgentAuth  func(string) (*ssh.PublicKeysCallback, error)
			expectedError string
		}{
			{repo: gitRepositoryMock{remote: remote, deleteRemoteError: []string{"delete error"}}, sshAgentAuth: sshSuccess, expectedError: "failed to update remote origin - remove: delete error"},
			{repo: gitRepositoryMock{remote: remote, createRemoteError: []string{"update error"}}, sshAgentAuth: sshSuccess, expectedError: "failed to update remote origin - create: update error"},
			{repo: gitRepositoryMock{remote: remote}, sshAgentAuth: sshFailure, expectedError: "failed to retrieve ssh authentication: ssh error"},
			{repo: gitRepositoryMock{remote: remote, deleteRemoteError: []string{"", "delete error"}}, sshAgentAuth: sshSuccess, expectedError: "failed to restore remote origin - remove: delete error"},
			{repo: gitRepositoryMock{remote: remote, createRemoteError: []string{"", "update error"}}, sshAgentAuth: sshSuccess, expectedError: "failed to restore remote origin - create: update error"},
		}

		originalSSHAgentAuth := sshAgentAuth

		for _, test := range tt {
			sshAgentAuth = test.sshAgentAuth
			commitID, err := pushChanges(&config, newVersion, &test.repo, &worktree, testTime, nil)
			sshAgentAuth = originalSSHAgentAuth

			assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", commitID)
			assert.EqualError(t, err, test.expectedError)
		}
	})

	t.Run("error - push", func(t *testing.T) {
		config := artifactPrepareVersionOptions{Username: "testUser", Password: "****"}
		repo := gitRepositoryMock{remote: remote, pushError: "push error"}
		worktree := gitWorktreeMock{commitHash: plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})}

		commitID, err := pushChanges(&config, newVersion, &repo, &worktree, testTime, nil)
		assert.Equal(t, "428ecf70bc22df0ba3dcf194b5ce53e769abab07", commitID)
		assert.EqualError(t, err, "push error")
	})
}

func TestTemplateCompatibility(t *testing.T) {
	tt := []struct {
		groovy         string
		versioningType string
		timestamp      bool
		commitID       bool
	}{
		{groovy: `${version}`, versioningType: "library", timestamp: false, commitID: false},
		{groovy: `${version}-${timestamp}`, versioningType: "cloud", timestamp: true, commitID: false},
		{groovy: `${version}-${timestamp}${commitId?"_"+commitId:""`, versioningType: "cloud", timestamp: true, commitID: true},
	}

	for _, test := range tt {
		versioningType, timestamp, commitID := templateCompatibility(test.groovy)
		assert.Equal(t, test.versioningType, versioningType)
		assert.Equal(t, test.timestamp, timestamp)
		assert.Equal(t, test.commitID, commitID)
	}
}

func TestConvertHTTPToSSHURL(t *testing.T) {
	tt := []struct {
		httpURL  string
		expected string
	}{
		{httpURL: "https://my.test.server/owner/repo.git", expected: "git@my.test.server:owner/repo.git"},
	}

	for _, test := range tt {
		assert.Equal(t, test.expected, convertHTTPToSSHURL(test.httpURL))
	}
}

func TestPropagateVersion(t *testing.T) {
	t.Parallel()

	gitCommitID := "theGitCommitId"
	testTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) //20200101000000

	t.Run("success case", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			VersioningType:        "cloud",
			AdditionalTargetTools: []string{"helm"},
		}

		chartMetadata := chart.Metadata{Version: "1.2.3"}
		content, err := yaml.Marshal(chartMetadata)
		assert.NoError(t, err)

		utils := newArtifactPrepareVersionMockUtils()
		utils.AddFile("myChart/Chart.yaml", content)
		artifactOpts := versioning.Options{}

		err = propagateVersion(&config, utils, &artifactOpts, "1.2.4", gitCommitID, testTime)
		assert.NoError(t, err)
	})

	t.Run("success case - dedicated build descriptors", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			VersioningType:              "cloud",
			AdditionalTargetTools:       []string{"helm"},
			AdditionalTargetDescriptors: []string{"myChart/Chart.yaml"},
			IncludeCommitID:             true,
		}

		chartMetadata := chart.Metadata{Version: "1.2.3"}
		content, err := yaml.Marshal(chartMetadata)
		assert.NoError(t, err)

		utils := newArtifactPrepareVersionMockUtils()
		utils.AddFile("myChart/Chart.yaml", content)
		artifactOpts := versioning.Options{}

		err = propagateVersion(&config, utils, &artifactOpts, "1.2.4", gitCommitID, testTime)
		assert.NoError(t, err)

		chartContent, err := utils.FileRead("myChart/Chart.yaml")
		assert.NoError(t, err)
		chartMeta := chart.Metadata{}
		err = yaml.Unmarshal(chartContent, &chartMeta)
		assert.NoError(t, err)

		assert.Equal(t, "1.2.4-20200101000000_theGitCommitId", chartMeta.AppVersion)
		assert.Equal(t, "1.2.4-20200101000000+theGitCommitId", chartMeta.Version)
	})

	t.Run("success case - dedicated build descriptors / no cloud", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			VersioningType:              "library",
			AdditionalTargetTools:       []string{"helm"},
			AdditionalTargetDescriptors: []string{"myChart/Chart.yaml"},
		}

		chartMetadata := chart.Metadata{Version: "1.2.3"}
		content, err := yaml.Marshal(chartMetadata)
		assert.NoError(t, err)

		utils := newArtifactPrepareVersionMockUtils()
		utils.AddFile("myChart/Chart.yaml", content)
		artifactOpts := versioning.Options{}

		err = propagateVersion(&config, utils, &artifactOpts, "1.2.4", gitCommitID, testTime)
		assert.NoError(t, err)

		chartContent, err := utils.FileRead("myChart/Chart.yaml")
		assert.NoError(t, err)
		chartMeta := chart.Metadata{}
		err = yaml.Unmarshal(chartContent, &chartMeta)
		assert.NoError(t, err)

		assert.Equal(t, "1.2.4", chartMeta.AppVersion)
		assert.Equal(t, "1.2.4", chartMeta.Version)
	})

	t.Run("success case - noop", func(t *testing.T) {
		config := artifactPrepareVersionOptions{}
		utils := newArtifactPrepareVersionMockUtils()
		artifactOpts := versioning.Options{}

		err := propagateVersion(&config, utils, &artifactOpts, "1.2.4", gitCommitID, testTime)
		assert.NoError(t, err)
	})

	t.Run("error case - wrong config", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			AdditionalTargetDescriptors: []string{"pom.xml"},
			AdditionalTargetTools:       []string{"maven", "helm"},
		}
		utils := newArtifactPrepareVersionMockUtils()
		artifactOpts := versioning.Options{}

		err := propagateVersion(&config, utils, &artifactOpts, "1.2.4", gitCommitID, testTime)
		assert.EqualError(t, err, "additionalTargetDescriptors cannot have a different number of entries than additionalTargetTools")
	})

	t.Run("error case - wrong target tool", func(t *testing.T) {
		config := artifactPrepareVersionOptions{
			AdditionalTargetTools: []string{"notKnown"},
		}
		utils := newArtifactPrepareVersionMockUtils()
		artifactOpts := versioning.Options{}

		err := propagateVersion(&config, utils, &artifactOpts, "1.2.4", gitCommitID, testTime)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve artifact")
	})
}

func TestTruncateString(t *testing.T) {
	t.Run("input string longer than maxLength - truncate", func(t *testing.T) {
		inputStr := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor"
		expected := "Lorem ipsum dolor sit amet, consectetur adipiscing..."

		outputStr := truncateString(inputStr, 50)
		assert.Equal(t, outputStr, expected)
	})

	t.Run("input string shorter than maxLength - return as is", func(t *testing.T) {
		inputStr := "Lorem ipsum dolor sit amet"
		outputStr := truncateString(inputStr, 50)

		assert.Equal(t, outputStr, inputStr)
	})

	t.Run("input string contains unicode chars", func(t *testing.T) {
		inputStr := "パイパーは素晴らしい図書館です"
		expected := "パイパーは..."

		outputStr := truncateString(inputStr, 5)
		assert.Equal(t, outputStr, expected)
	})

	t.Run("input string is empty", func(t *testing.T) {
		outputStr := truncateString("", 5)
		assert.Equal(t, outputStr, "")
	})
}
