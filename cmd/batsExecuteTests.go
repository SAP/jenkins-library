package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/SAP/jenkins-library/pkg/command"
	pipergit "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/go-git/go-git/v5"
)

type batsExecuteTestsUtils interface {
	CloneRepo(URL string) error
	Stdin(in io.Reader)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	SetEnv([]string)
	FileWrite(path string, content []byte, perm os.FileMode) error
	RunExecutable(e string, p ...string) error
}

type batsExecuteTestsUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newBatsExecuteTestsUtils() batsExecuteTestsUtils {
	utils := batsExecuteTestsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func batsExecuteTests(config batsExecuteTestsOptions, telemetryData *telemetry.CustomData, influx *batsExecuteTestsInflux) {
	utils := newBatsExecuteTestsUtils()

	influx.step_data.fields.bats = false
	err := runBatsExecuteTests(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	influx.step_data.fields.bats = true
}

func runBatsExecuteTests(config *batsExecuteTestsOptions, telemetryData *telemetry.CustomData, utils batsExecuteTestsUtils) error {
	if config.OutputFormat != "tap" && config.OutputFormat != "junit" {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("output format '%v' is incorrect. Possible drivers: tap, junit", config.OutputFormat)
	}

	err := utils.CloneRepo(config.Repository)

	if err != nil && !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return fmt.Errorf("couldn't pull %s repository: %w", config.Repository, err)
	}

	tapOutput := bytes.Buffer{}
	utils.Stdout(io.MultiWriter(&tapOutput, log.Writer()))

	utils.SetEnv(config.EnvVars)
	err = utils.RunExecutable("bats-core/bin/bats", "--recursive", "--tap", config.TestPath)
	if err != nil {
		return fmt.Errorf("failed to run bats test: %w", err)
	}

	err = utils.FileWrite("TEST-"+config.TestPackage+".tap", tapOutput.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write tap file: %w", err)
	}

	if config.OutputFormat == "junit" {
		output := bytes.Buffer{}
		utils.Stdout(io.MultiWriter(&output, log.Writer()))

		utils.SetEnv(append(config.EnvVars, "NPM_CONFIG_PREFIX=~/.npm-global"))
		err = utils.RunExecutable("npm", "install", "tap-xunit", "-g")
		if err != nil {
			return fmt.Errorf("failed to install tap-xunit: %w", err)
		}

		homedir, _ := os.UserHomeDir()
		path := "PATH=" + os.Getenv("PATH") + ":" + homedir + "/.npm-global/bin"

		output = bytes.Buffer{}
		utils.Stdout(&output)
		utils.Stdin(&tapOutput)
		utils.SetEnv(append(config.EnvVars, path))
		err = utils.RunExecutable("tap-xunit", "--package="+config.TestPackage)
		if err != nil {
			return fmt.Errorf("failed to run tap-xunit: %w", err)
		}
		err = utils.FileWrite("TEST-"+config.TestPackage+".xml", output.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("failed to write tap file: %w", err)
		}
	}

	return nil
}

func (b *batsExecuteTestsUtilsBundle) CloneRepo(URL string) error {
	// ToDo: BatsExecute test needs to check if the repo can come from a
	// enterprise github instance and needs ca-cert handelling seperately
	_, err := pipergit.PlainClone("", "", URL, "", "bats-core", []byte{})
	return err

}
