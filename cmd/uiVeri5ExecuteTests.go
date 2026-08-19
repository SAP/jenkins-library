package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func uiVeri5ExecuteTests(config uiVeri5ExecuteTestsOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runUIVeri5(&config, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runUIVeri5(config *uiVeri5ExecuteTestsOptions, command command.ExecRunner) error {
	envs := []string{"NPM_CONFIG_PREFIX=~/.npm-global"}
	path := "PATH=" + os.Getenv("PATH") + ":~/.npm-global/bin"
	envs = append(envs, path)
	if config.TestServerURL != "" {
		envs = append(envs, "TARGET_SERVER_URL="+config.TestServerURL)
	}
	command.SetEnv(envs)

	installCommandTokens := strings.Split(config.InstallCommand, " ")
	if err := command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...); err != nil {
		log.SetErrorCategory(log.ErrorCustom)
		return fmt.Errorf("failed to execute install command: %v: %w", config.InstallCommand, err)
	}

	if config.TestOptions != "" {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("parameter testOptions no longer supported, please use runOptions parameter instead.")
	}
	if err := command.RunExecutable(config.RunCommand, config.RunOptions...); err != nil {
		log.SetErrorCategory(log.ErrorTest)
		return fmt.Errorf("failed to execute run command: %v %v: %w", config.RunCommand, strings.Join(config.RunOptions, " "), err)
	}
	return nil
}
