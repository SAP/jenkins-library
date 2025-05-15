package codeql

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
)

type RepoInfo struct {
	ServerUrl   string
	Owner       string
	Repo        string
	CommitId    string
	AnalyzedRef string
	FullRef     string
	FullUrl     string
	ScanUrl     string
}

func GetRepoInfo(repository, analyzedRef, commitID, targetGithubRepoURL, targetGithubBranchName string) (*RepoInfo, error) {
	repoInfo := &RepoInfo{}
	err := setRepoInfoFromRepoUri(repository, repoInfo)
	if err != nil {
		log.Entry().Error(err)
	}
	repoInfo.AnalyzedRef = analyzedRef
	repoInfo.CommitId = commitID

	getRepoInfoFromOrchestrator(repoInfo)

	if len(targetGithubRepoURL) > 0 {
		log.Entry().Infof("Checking target GitHub repo URL: %s", targetGithubRepoURL)
		if err := setTargetGithubRepoInfo(targetGithubRepoURL, targetGithubBranchName, repoInfo); err != nil {
			return repoInfo, err
		}
	}

	repoUrl := fmt.Sprintf("%s/%s/%s", repoInfo.ServerUrl, repoInfo.Owner, repoInfo.Repo)
	repoInfo.FullUrl = repoUrl
	repoInfo.ScanUrl = fmt.Sprintf("%s/security/code-scanning?query=is:open+ref:%s", repoUrl, repoInfo.AnalyzedRef)

	repoRef, err := buildRepoReference(repoUrl, repoInfo.AnalyzedRef)
	if err != nil {
		return nil, err
	}
	repoInfo.FullRef = repoRef

	return repoInfo, nil
}

func buildRepoReference(repository, analyzedRef string) (string, error) {
	ref := strings.Split(analyzedRef, "/")
	if len(ref) < 3 {
		return "", fmt.Errorf("wrong analyzedRef format: %s", analyzedRef)
	}
	if strings.Contains(analyzedRef, "pull") {
		if len(ref) < 4 {
			return "", fmt.Errorf("wrong analyzedRef format: %s", analyzedRef)
		}
		return fmt.Sprintf("%s/pull/%s", repository, ref[2]), nil
	}
	return fmt.Sprintf("%s/tree/%s", repository, ref[2]), nil
}

func setRepoInfoFromRepoUri(repoUri string, repoInfo *RepoInfo) error {
	if repoUri == "" {
		return errors.New("repository param is not set or it cannot be auto populated")
	}
	serverUrl, owner, repo, err := parseRepoUri(repoUri)
	if err != nil {
		return err
	}
	repoInfo.ServerUrl = serverUrl
	repoInfo.Owner = owner
	repoInfo.Repo = repo
	return nil
}

func parseRepoUri(repoUri string) (string, string, string, error) {
	pat := regexp.MustCompile(`^(https:\/\/|git@)([\S]+:[\S]+@)?([^\/:]+)[\/:]([^\/:]+\/[\S]+)$`)
	matches := pat.FindAllStringSubmatch(repoUri, -1)
	if len(matches) > 0 {
		match := matches[0]
		serverUrl := "https://" + match[3]
		repoData := strings.Split(strings.TrimSuffix(match[4], ".git"), "/")
		if len(repoData) != 2 {
			return "", "", "", fmt.Errorf("invalid repository %s", repoUri)
		}
		owner, repo := repoData[0], repoData[1]
		return serverUrl, owner, repo, nil
	}
	return "", "", "", fmt.Errorf("invalid repository %s", repoUri)
}

func getRepoInfoFromOrchestrator(repoInfo *RepoInfo) {
	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().Warn("No orchestrator found. We assume piper is running locally.")
	} else {
		if repoInfo.AnalyzedRef == "" {
			repoInfo.AnalyzedRef = provider.GitReference()
		}
		if repoInfo.CommitId == "" || repoInfo.CommitId == "NA" {
			repoInfo.CommitId = provider.CommitSHA()
		}
		if repoInfo.ServerUrl == "" {
			err := setRepoInfoFromRepoUri(provider.RepoURL(), repoInfo)
			if err != nil {
				log.Entry().WithError(err).Error("failed to get repo info from orchestrator")
			}
		}
	}
}

func setTargetGithubRepoInfo(targetGHRepoURL, targetGHBranchName string, repoInfo *RepoInfo) error {
	if strings.Contains(repoInfo.ServerUrl, "github") {
		return errors.New("TargetGithubRepoURL should not be set as the source repo is on github")
	}
	err := setRepoInfoFromRepoUri(targetGHRepoURL, repoInfo)
	if err != nil {
		log.Entry().WithError(err).Error("Failed to get target github repo info")
		return err
	}
	if len(targetGHBranchName) > 0 {
		log.Entry().Infof("Target GitHub branch name: %s", targetGHBranchName)
		repoInfo.AnalyzedRef = getFullBranchName(targetGHBranchName)
	}
	return nil
}

func getFullBranchName(branchName string) string {
	if len(strings.Split(branchName, "/")) < 3 {
		return "refs/heads/" + branchName
	}
	return branchName
}
