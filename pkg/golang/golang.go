package golang

import (
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
)

type utilsBundle struct {
	command.Command
}

// prepare golang private packages for whitesource and blackduck(detectExecuteScan)
func PrepareGolangPrivatePackages(stepName, privateModules, privateModulesGitToken string) error {
	utils := &utilsBundle{
		Command: command.Command{
			StepName: stepName,
		},
	}
	os.Setenv("GOPRIVATE", privateModules)
	err := gitConfigurationForPrivateModules(privateModules, privateModulesGitToken, utils)
	if err != nil {
		return err
	}
	return nil
}

func gitConfigurationForPrivateModules(privateMod string, token string, utils *utilsBundle) error {
	privateMod = strings.ReplaceAll(privateMod, "/*", "")
	privateMod = strings.ReplaceAll(privateMod, "*.", "")
	modules := strings.Split(privateMod, ",")
	for _, v := range modules {
		authenticatedRepoURL := fmt.Sprintf("https://%s@%s", token, v)
		repoBaseURL := fmt.Sprintf("https://%s", v)
		err := utils.RunExecutable("git", "config", "--global", fmt.Sprintf("url.%s.insteadOf", authenticatedRepoURL), repoBaseURL)
		if err != nil {
			return err
		}
	}
	return nil
}
