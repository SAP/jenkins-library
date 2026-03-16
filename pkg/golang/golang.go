package golang

import (
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
)

// PrepareGolangPrivatePackages prepares golang private packages for whitesource and blackduck(detectExecuteScan)
func PrepareGolangPrivatePackages(stepName, privateModules, privateModulesGitToken string) error {
	if err := os.Setenv("GOPRIVATE", privateModules); err != nil {
		return err
	}
	return gitConfigurationForPrivateModules(privateModules, privateModulesGitToken, &command.Command{StepName: stepName})
}

func gitConfigurationForPrivateModules(privateMod string, token string, cmd *command.Command) error {
	privateMod = strings.ReplaceAll(privateMod, "/*", "")
	privateMod = strings.ReplaceAll(privateMod, "*.", "")
	modules := strings.Split(privateMod, ",")
	for _, v := range modules {
		// Inline construction of authenticated URL key and base URL
		if err := cmd.RunExecutable(
			"git",
			"config",
			"--global",
			fmt.Sprintf("url.https://%s@%s.insteadOf", token, v),
			fmt.Sprintf("https://%s", v),
		); err != nil {
			return err
		}
	}
	return nil
}
