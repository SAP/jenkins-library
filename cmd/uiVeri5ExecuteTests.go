package cmd

import (
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
	installCommandTokens := tokenize(config.InstallCommand) // strings.Split(command, " ")
	command.SetDir(config.ModulePath)
	command.SetEnv([]string{"NPM_CONFIG_PREFIX=/home/node/.npm-global"})
	err := command.RunExecutable(installCommandTokens[0], installCommandTokens[1:]...)
	if err != nil {
		log.Entry().WithError(err).WithField("command", config.InstallCommand).Fatal("failed to execute install command")
		return err
	}

	runCommandTokens := tokenize(config.RunCommand)
	runCommandTokens = append(runCommandTokens, config.ConfPath)
	command.SetDir(config.ModulePath)
	command.SetEnv([]string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/node/.npm-global/bin"})
	err = command.RunExecutable(runCommandTokens[0], runCommandTokens[1:]...)
	if err != nil {
		log.Entry().WithError(err).WithField("command", config.RunCommand).Fatal("failed to execute run command")
		return err
	}
	return nil
}
