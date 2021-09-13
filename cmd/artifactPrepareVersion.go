package cmd

import (
	"bytes"
	"fmt"
	"io"
	netHttp "net/http"
	"os"
	"strings"
	"text/template"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/command"
	gitUtils "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/pkg/errors"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type gitRepository interface {
	CommitObject(plumbing.Hash) (*object.Commit, error)
	CreateTag(string, plumbing.Hash, *git.CreateTagOptions) (*plumbing.Reference, error)
	CreateRemote(*gitConfig.RemoteConfig) (*git.Remote, error)
	DeleteRemote(string) error
	Push(*git.PushOptions) error
	Remote(string) (*git.Remote, error)
	ResolveRevision(plumbing.Revision) (*plumbing.Hash, error)
	Worktree() (*git.Worktree, error)
}

type gitWorktree interface {
	Checkout(*git.CheckoutOptions) error
	Commit(string, *git.CommitOptions) (plumbing.Hash, error)
}

func getGitWorktree(repository gitRepository) (gitWorktree, error) {
	return repository.Worktree()
}

type artifactPrepareVersionUtils interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error

	DownloadFile(url, filename string, header netHttp.Header, cookies []*netHttp.Cookie) error

	Glob(pattern string) (matches []string, err error)
	FileExists(filename string) (bool, error)
	Copy(src, dest string) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
	FileWrite(path string, content []byte, perm os.FileMode) error
	FileRead(path string) ([]byte, error)
}

type artifactPrepareVersionUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
}

func newArtifactPrepareVersionUtilsBundle() artifactPrepareVersionUtils {
	utils := artifactPrepareVersionUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &piperhttp.Client{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func artifactPrepareVersion(config artifactPrepareVersionOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *artifactPrepareVersionCommonPipelineEnvironment) {
	utils := newArtifactPrepareVersionUtilsBundle()

	// open local .git repository
	repository, err := openGit()
	if err != nil {
		log.Entry().WithError(err).Fatal("git repository required - none available")
	}

	err = runArtifactPrepareVersion(&config, telemetryData, commonPipelineEnvironment, nil, utils, repository, getGitWorktree)
	if err != nil {
		log.Entry().WithError(err).Fatal("artifactPrepareVersion failed")
	}
}

var sshAgentAuth = ssh.NewSSHAgentAuth

func runArtifactPrepareVersion(config *artifactPrepareVersionOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *artifactPrepareVersionCommonPipelineEnvironment, artifact versioning.Artifact, utils artifactPrepareVersionUtils, repository gitRepository, getWorktree func(gitRepository) (gitWorktree, error)) error {

	telemetryData.Custom1Label = "buildTool"
	telemetryData.Custom1 = config.BuildTool
	telemetryData.Custom2Label = "filePath"
	telemetryData.Custom2 = config.FilePath

	// Options for artifact
	artifactOpts := versioning.Options{
		GlobalSettingsFile:  config.GlobalSettingsFile,
		M2Path:              config.M2Path,
		ProjectSettingsFile: config.ProjectSettingsFile,
		VersionField:        config.CustomVersionField,
		VersionSection:      config.CustomVersionSection,
		VersioningScheme:    config.CustomVersioningScheme,
		VersionSource:       config.DockerVersionSource,
	}

	var err error
	if artifact == nil {
		artifact, err = versioning.GetArtifact(config.BuildTool, config.FilePath, &artifactOpts, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrap(err, "failed to retrieve artifact")
		}
	}

	versioningType := config.VersioningType

	// support former groovy versioning template and translate into new options
	if len(config.VersioningTemplate) > 0 {
		versioningType, _, config.IncludeCommitID = templateCompatibility(config.VersioningTemplate)
	}

	version, err := artifact.GetVersion()
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "failed to retrieve version")
	}
	log.Entry().Infof("Version before automatic versioning: %v", version)

	gitCommit, gitCommitMessage, err := getGitCommitID(repository)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return err
	}
	gitCommitID := gitCommit.String()

	commonPipelineEnvironment.git.headCommitID = gitCommitID
	newVersion := version

	if versioningType == "cloud" || versioningType == "cloud_noTag" {
		versioningTempl, err := versioningTemplate(artifact.VersioningScheme())
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "failed to get versioning template for scheme '%v'", artifact.VersioningScheme())
		}

		now := time.Now()

		newVersion, err = calculateNewVersion(versioningTempl, version, gitCommitID, config.IncludeCommitID, config.ShortCommitID, config.UnixTimestamp, now)
		if err != nil {
			return errors.Wrap(err, "failed to calculate new version")
		}

		worktree, err := getWorktree(repository)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrap(err, "failed to retrieve git worktree")
		}

		// opening repository does not seem to consider already existing files properly
		// behavior in case we do not run initializeWorktree:
		//   git.Add(".") will add the complete workspace instead of only changed files
		err = initializeWorktree(gitCommit, worktree)
		if err != nil {
			return err
		}

		// only update version in build descriptor if required in order to save prossing time (e.g. maven case)
		if newVersion != version {
			err = artifact.SetVersion(newVersion)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return errors.Wrap(err, "failed to write version")
			}
		}

		//ToDo: what about closure in current Groovy step. Discard the possibility or provide extension mechanism?

		if versioningType == "cloud" {
			// commit changes and push to repository (including new version tag)
			gitCommitID, err = pushChanges(config, newVersion, repository, worktree, now)
			if err != nil {
				if strings.Contains(fmt.Sprint(err), "reference already exists") {
					log.SetErrorCategory(log.ErrorCustom)
				}
				return errors.Wrapf(err, "failed to push changes for version '%v'", newVersion)
			}
		}
	}

	log.Entry().Infof("New version: '%v'", newVersion)

	commonPipelineEnvironment.git.commitID = gitCommitID // this commitID changes and is not necessarily the HEAD commitID
	commonPipelineEnvironment.artifactVersion = newVersion
	commonPipelineEnvironment.originalArtifactVersion = version
	commonPipelineEnvironment.git.commitMessage = gitCommitMessage

	// we may replace GetVersion() above with GetCoordinates() at some point ...
	coordinates, err := artifact.GetCoordinates()
	if err != nil && !config.FetchCoordinates {
		log.Entry().Warnf("fetchCoordinates is false and failed get artifact Coordinates")
	} else if err != nil && config.FetchCoordinates {
		return fmt.Errorf("failed to get coordinates: %w", err)
	} else {
		commonPipelineEnvironment.artifactID = coordinates.ArtifactID
		commonPipelineEnvironment.groupID = coordinates.GroupID
		commonPipelineEnvironment.packaging = coordinates.Packaging
	}

	// }

	return nil
}

func openGit() (gitRepository, error) {
	workdir, _ := os.Getwd()
	return gitUtils.PlainOpen(workdir)
}

func getGitCommitID(repository gitRepository) (plumbing.Hash, string, error) {
	commitID, err := repository.ResolveRevision(plumbing.Revision("HEAD"))
	if err != nil {
		return plumbing.Hash{}, "", errors.Wrap(err, "failed to retrieve git commit ID")
	}
	// ToDo not too elegant to retrieve the commit message here, must be refactored sooner than later
	// but to quickly address https://github.com/SAP/jenkins-library/pull/1515 let's revive this
	commitObject, err := repository.CommitObject(*commitID)
	if err != nil {
		return *commitID, "", errors.Wrap(err, "failed to retrieve git commit message")
	}
	return *commitID, commitObject.Message, nil
}

func versioningTemplate(scheme string) (string, error) {
	// generally: timestamp acts as build number providing a proper order
	switch scheme {
	case "docker":
		// from Docker documentation:
		// A tag name must be valid ASCII and may contain lowercase and uppercase letters, digits, underscores, periods and dashes.
		// A tag name may not start with a period or a dash and may contain a maximum of 128 characters.
		return "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}-{{.CommitID}}{{end}}{{end}}", nil
	case "maven":
		// according to https://www.mojohaus.org/versions-maven-plugin/version-rules.html
		return "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}_{{.CommitID}}{{end}}{{end}}", nil
	case "pep440":
		// according to https://www.python.org/dev/peps/pep-0440/
		return "{{.Version}}{{if .Timestamp}}.{{.Timestamp}}{{if .CommitID}}+{{.CommitID}}{{end}}{{end}}", nil
	case "semver2":
		// according to https://semver.org/spec/v2.0.0.html
		return "{{.Version}}{{if .Timestamp}}-{{.Timestamp}}{{if .CommitID}}+{{.CommitID}}{{end}}{{end}}", nil
	}
	return "", fmt.Errorf("versioning scheme '%v' not supported", scheme)
}

func calculateNewVersion(versioningTemplate, currentVersion, commitID string, includeCommitID, shortCommitID, unixTimestamp bool, t time.Time) (string, error) {
	tmpl, err := template.New("version").Parse(versioningTemplate)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create version template: %v", versioningTemplate)
	}

	timestamp := t.Format("20060102150405")
	if unixTimestamp {
		timestamp = fmt.Sprint(t.Unix())
	}

	buf := new(bytes.Buffer)
	versionParts := struct {
		Version   string
		Timestamp string
		CommitID  string
	}{
		Version:   currentVersion,
		Timestamp: timestamp,
	}

	if includeCommitID {
		versionParts.CommitID = commitID
		if shortCommitID {
			versionParts.CommitID = commitID[0:7]
		}
	}

	err = tmpl.Execute(buf, versionParts)
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute versioning template: %v", versioningTemplate)
	}

	newVersion := buf.String()
	if len(newVersion) == 0 {
		return "", fmt.Errorf("failed calculate version, new version is '%v'", newVersion)
	}
	return buf.String(), nil
}

func initializeWorktree(gitCommit plumbing.Hash, worktree gitWorktree) error {
	// checkout current revision in order to work on that
	err := worktree.Checkout(&git.CheckoutOptions{Hash: gitCommit, Keep: true})
	if err != nil {
		return errors.Wrap(err, "failed to initialize worktree")
	}

	return nil
}

func pushChanges(config *artifactPrepareVersionOptions, newVersion string, repository gitRepository, worktree gitWorktree, t time.Time) (string, error) {

	var commitID string

	commit, err := addAndCommit(config, worktree, newVersion, t)
	if err != nil {
		return commit.String(), err
	}

	commitID = commit.String()

	tag := fmt.Sprintf("%v%v", config.TagPrefix, newVersion)
	_, err = repository.CreateTag(tag, commit, nil)
	if err != nil {
		return commitID, err
	}

	ref := gitConfig.RefSpec(fmt.Sprintf("refs/tags/%v:refs/tags/%v", tag, tag))

	pushOptions := git.PushOptions{
		RefSpecs: []gitConfig.RefSpec{gitConfig.RefSpec(ref)},
	}

	currentRemoteOrigin, err := repository.Remote("origin")
	if err != nil {
		return commitID, errors.Wrap(err, "failed to retrieve current remote origin")
	}
	var updatedRemoteOrigin *git.Remote

	urls := originUrls(repository)
	if len(urls) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return commitID, fmt.Errorf("no remote url maintained")
	}
	if strings.HasPrefix(urls[0], "http") {
		if len(config.Username) == 0 || len(config.Password) == 0 {
			// handling compatibility: try to use ssh in case no credentials are available
			log.Entry().Info("git username/password missing - switching to ssh")

			remoteURL := convertHTTPToSSHURL(urls[0])

			// update remote origin url to point to ssh url instead of http(s) url
			err = repository.DeleteRemote("origin")
			if err != nil {
				return commitID, errors.Wrap(err, "failed to update remote origin - remove")
			}
			updatedRemoteOrigin, err = repository.CreateRemote(&gitConfig.RemoteConfig{Name: "origin", URLs: []string{remoteURL}})
			if err != nil {
				return commitID, errors.Wrap(err, "failed to update remote origin - create")
			}

			pushOptions.Auth, err = sshAgentAuth("git")
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return commitID, errors.Wrap(err, "failed to retrieve ssh authentication")
			}
			log.Entry().Infof("using remote '%v'", remoteURL)
		} else {
			pushOptions.Auth = &http.BasicAuth{Username: config.Username, Password: config.Password}
		}
	} else {
		pushOptions.Auth, err = sshAgentAuth("git")
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return commitID, errors.Wrap(err, "failed to retrieve ssh authentication")
		}
	}

	err = repository.Push(&pushOptions)
	if err != nil {
		errText := fmt.Sprint(err)
		switch {
		case strings.Contains(errText, "ssh: handshake failed"):
			log.SetErrorCategory(log.ErrorConfiguration)
		case strings.Contains(errText, "Permission"):
			log.SetErrorCategory(log.ErrorConfiguration)
		case strings.Contains(errText, "authorization failed"):
			log.SetErrorCategory(log.ErrorConfiguration)
		case strings.Contains(errText, "authentication required"):
			log.SetErrorCategory(log.ErrorConfiguration)
		case strings.Contains(errText, "knownhosts:"):
			err = errors.Wrap(err, "known_hosts file seems invalid")
			log.SetErrorCategory(log.ErrorConfiguration)
		case strings.Contains(errText, "unable to find any valid known_hosts file"):
			log.SetErrorCategory(log.ErrorConfiguration)
		case strings.Contains(errText, "connection timed out"):
			log.SetErrorCategory(log.ErrorInfrastructure)
		}
		return commitID, err
	}

	if updatedRemoteOrigin != currentRemoteOrigin {
		err = repository.DeleteRemote("origin")
		if err != nil {
			return commitID, errors.Wrap(err, "failed to restore remote origin - remove")
		}
		_, err := repository.CreateRemote(currentRemoteOrigin.Config())
		if err != nil {
			return commitID, errors.Wrap(err, "failed to restore remote origin - create")
		}
	}

	return commitID, nil
}

func addAndCommit(config *artifactPrepareVersionOptions, worktree gitWorktree, newVersion string, t time.Time) (plumbing.Hash, error) {
	//maybe more options are required: https://github.com/go-git/go-git/blob/master/_examples/commit/main.go
	commit, err := worktree.Commit(fmt.Sprintf("update version %v", newVersion), &git.CommitOptions{All: true, Author: &object.Signature{Name: config.CommitUserName, When: t}})
	if err != nil {
		return commit, errors.Wrap(err, "failed to commit new version")
	}
	return commit, nil
}

func originUrls(repository gitRepository) []string {
	remote, err := repository.Remote("origin")
	if err != nil || remote == nil {
		return []string{}
	}
	return remote.Config().URLs
}

func convertHTTPToSSHURL(url string) string {
	sshURL := strings.Replace(url, "https://", "git@", 1)
	return strings.Replace(sshURL, "/", ":", 1)
}

func templateCompatibility(groovyTemplate string) (versioningType string, useTimestamp bool, useCommitID bool) {
	useTimestamp = strings.Contains(groovyTemplate, "${timestamp}")
	useCommitID = strings.Contains(groovyTemplate, "${commitId")

	versioningType = "library"

	if useTimestamp {
		versioningType = "cloud"
	}

	return
}
