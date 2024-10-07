package cmd

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

const hadolintCommand = "hadolint"

// HadolintPiperFileUtils abstracts piperutils.Files
type HadolintPiperFileUtils interface {
	FileExists(filename string) (bool, error)
	FileWrite(filename string, data []byte, perm os.FileMode) error
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// HadolintClient abstracts http.Client
type HadolintClient interface {
	SetOptions(options piperhttp.ClientOptions)
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
}

// hadolintRunner abstracts command.Command
type hadolintRunner interface {
	RunExecutable(executable string, params ...string) error
	Stdout(err io.Writer)
	Stderr(err io.Writer)
}

type hadolintUtils struct {
	HadolintPiperFileUtils
	HadolintClient
	hadolintRunner
}

func hadolintExecute(config hadolintExecuteOptions, _ *telemetry.CustomData) {
	runner := command.Command{
		ErrorCategoryMapping: map[string][]string{},
	}
	// reroute runner output to logging framework
	runner.Stdout(log.Writer())
	runner.Stderr(log.Writer())

	utils := hadolintUtils{
		HadolintPiperFileUtils: &piperutils.Files{},
		HadolintClient:         &piperhttp.Client{},
		hadolintRunner:         &runner,
	}

	if err := runHadolint(config, utils); err != nil {
		log.Entry().WithError(err).Fatal("Execution failed")
	}
}

func runHadolint(config hadolintExecuteOptions, utils hadolintUtils) error {
	var outputBuffer bytes.Buffer
	var errorBuffer bytes.Buffer
	utils.Stdout(&outputBuffer)
	utils.Stderr(&errorBuffer)

	options := []string{"--format", "checkstyle"}
	// load config file from URL
	if !hasConfigurationFile(config.ConfigurationFile, utils) && len(config.ConfigurationURL) > 0 {
		clientOptions := piperhttp.ClientOptions{
			TransportTimeout:          20 * time.Second,
			TransportSkipVerification: true,
		}
		if len(config.CustomTLSCertificateLinks) > 0 {
			clientOptions.TransportSkipVerification = false
			clientOptions.TrustedCerts = config.CustomTLSCertificateLinks
		}

		if len(config.ConfigurationUsername) > 0 {
			clientOptions.Username = config.ConfigurationUsername
			clientOptions.Password = config.ConfigurationPassword
		}
		utils.SetOptions(clientOptions)
		if err := loadConfigurationFile(config.ConfigurationURL, config.ConfigurationFile, utils); err != nil {
			return errors.Wrap(err, "failed to load configuration file from URL")
		}
	}
	// use config
	if hasConfigurationFile(config.ConfigurationFile, utils) {
		options = append(options, "--config", config.ConfigurationFile)
		log.Entry().WithField("file", config.ConfigurationFile).Debug("Using configuration file")
	} else {
		log.Entry().Debug("No configuration file found.")
	}
	// execute scan command
	err := utils.RunExecutable(hadolintCommand, append([]string{config.DockerFile}, options...)...)

	//TODO: related to https://github.com/hadolint/hadolint/issues/391
	// hadolint exists with 1 if there are processing issues but also if there are findings
	// thus check stdout first if a report was created
	if output := outputBuffer.String(); len(output) > 0 {
		log.Entry().WithField("report", output).Debug("Report created")
		if err := utils.FileWrite(config.ReportFile, []byte(output), 0666); err != nil {
			log.Entry().WithError(err).Warningf("failed to write report %v", config.ReportFile)
		}
	} else if err != nil {
		// if stdout is empty a processing issue occured
		return errors.Wrap(err, errorBuffer.String())
	}
	//TODO: mock away in tests
	// persist report information
	piperutils.PersistReportsAndLinks("hadolintExecute", "./", utils, []piperutils.Path{{Target: config.ReportFile}}, []piperutils.Path{})
	return nil
}

// loadConfigurationFile loads a file from the provided url
func loadConfigurationFile(url, file string, utils hadolintUtils) error {
	log.Entry().WithField("url", url).Debug("Loading configuration file from URL")
	return utils.DownloadFile(url, file, nil, nil)
}

// hasConfigurationFile checks if the given file exists
func hasConfigurationFile(file string, utils hadolintUtils) bool {
	exists, err := utils.FileExists(file)
	if err != nil {
		log.Entry().WithError(err).Error()
	}
	return exists
}
