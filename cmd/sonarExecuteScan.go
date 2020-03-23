package cmd

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	SliceUtils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type sonarSettings struct {
	binary      string
	environment []string
	options     []string
}

func (s *sonarSettings) addEnvironment(element string) { s.environment = append(s.environment, element) }

func (s *sonarSettings) addOption(element string) { s.options = append(s.options, element) }

var sonar sonarSettings

var execLookPath = exec.LookPath
var fileUtilsExists = FileUtils.FileExists
var fileUtilsUnzip = FileUtils.Unzip
var osRename = os.Rename

func sonarExecuteScan(options sonarExecuteScanOptions, _ *telemetry.CustomData) {
	runner := command.Command{}
	// reroute command output to logging framework
	runner.Stdout(log.Entry().Writer())
	runner.Stderr(log.Entry().Writer())

	client := piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{TransportTimeout: 20 * time.Second})

	sonar = sonarSettings{
		binary:      "sonar-scanner",
		environment: []string{},
		options:     []string{},
	}

	if err := runSonar(options, &client, &runner); err != nil {
		log.Entry().WithError(err).Fatal("Execution failed")
	}
}

func runSonar(options sonarExecuteScanOptions, client piperhttp.Downloader, runner execRunner) error {
	if len(options.Host) > 0 {
		sonar.addEnvironment("SONAR_HOST_URL=" + options.Host)
	}
	if len(options.Token) > 0 {
		sonar.addEnvironment("SONAR_AUTH_TOKEN=" + options.Token)
	}
	if len(options.Organization) > 0 {
		sonar.addOption("sonar.organization=" + options.Organization)
	}
	if len(options.ProjectVersion) > 0 {
		sonar.addOption("sonar.projectVersion=" + options.ProjectVersion)
	}
	if err := handlePullRequest(options); err != nil {
		return err
	}
	if err := loadSonarScanner(options.SonarScannerDownloadURL, client); err != nil {
		return err
	}
	if err := loadCertificates(options.CustomTLSCertificateLinks, client, runner); err != nil {
		return err
	}

	log.Entry().
		WithField("command", sonar.binary).
		WithField("options", sonar.options).
		WithField("environment", sonar.environment).
		Debug("Executing sonar scan command")

	sonar.options = SliceUtils.Prefix(SliceUtils.Trim(sonar.options), "-D")

	if len(options.Options) > 0 {
		sonar.addOption(options.Options)
	}
	runner.SetEnv(sonar.environment)
	return runner.RunExecutable(sonar.binary, sonar.options...)
}

func handlePullRequest(options sonarExecuteScanOptions) error {
	if len(options.ChangeID) > 0 {
		if options.LegacyPRHandling {
			// see https://docs.sonarqube.org/display/PLUG/GitHub+Plugin
			sonar.addOption("sonar.analysis.mode=preview")
			sonar.addOption("sonar.github.pullRequest=" + options.ChangeID)
			if len(options.GithubAPIURL) > 0 {
				sonar.addOption("sonar.github.endpoint=" + options.GithubAPIURL)
			}
			if len(options.GithubToken) > 0 {
				sonar.addOption("sonar.github.oauth=" + options.GithubToken)
			}
			if len(options.Owner) > 0 && len(options.Repository) > 0 {
				sonar.addOption("sonar.github.repository=" + options.Owner + "/" + options.Repository)
			}
			if options.DisableInlineComments {
				sonar.addOption("sonar.github.disableInlineComments=" + strconv.FormatBool(options.DisableInlineComments))
			}
		} else {
			// see https://sonarcloud.io/documentation/analysis/pull-request/
			provider := strings.ToLower(options.PullRequestProvider)
			if provider == "github" {
				sonar.addOption("sonar.pullrequest.github.repository=" + options.Owner + "/" + options.Repository)
			} else {
				return errors.New("Pull-Request provider '" + provider + "' is not supported!")
			}
			sonar.addOption("sonar.pullrequest.key=" + options.ChangeID)
			sonar.addOption("sonar.pullrequest.base=" + options.ChangeTarget)
			sonar.addOption("sonar.pullrequest.branch=" + options.ChangeBranch)
			sonar.addOption("sonar.pullrequest.provider=" + provider)
		}
	}
	return nil
}

func loadSonarScanner(url string, client piperhttp.Downloader) error {
	if scannerPath, err := execLookPath(sonar.binary); err == nil {
		// using existing sonar-scanner
		log.Entry().WithField("path", scannerPath).Debug("Using local sonar-scanner")
	} else if len(url) != 0 {
		// download sonar-scanner-cli into TEMP folder
		log.Entry().WithField("url", url).Debug("Downloading sonar-scanner")
		tmpFolder := getTempDir()
		defer os.RemoveAll(tmpFolder) // clean up
		archive := filepath.Join(tmpFolder, path.Base(url))
		if err := client.DownloadFile(url, archive, nil, nil); err != nil {
			return errors.Wrap(err, "Download of sonar-scanner failed")
		}
		// unzip sonar-scanner-cli
		log.Entry().WithField("source", archive).WithField("target", tmpFolder).Debug("Extracting sonar-scanner")
		if _, err := fileUtilsUnzip(archive, tmpFolder); err != nil {
			return errors.Wrap(err, "Extraction of sonar-scanner failed")
		}
		// move sonar-scanner-cli to .sonar-scanner/
		toolPath := ".sonar-scanner"
		foldername := strings.ReplaceAll(strings.ReplaceAll(archive, ".zip", ""), "cli-", "")
		log.Entry().WithField("source", foldername).WithField("target", toolPath).Debug("Moving sonar-scanner")
		if err := osRename(foldername, toolPath); err != nil {
			return errors.Wrap(err, "Moving of sonar-scanner failed")
		}
		// update binary path
		sonar.binary = filepath.Join(getWorkingDir(), toolPath, "bin", sonar.binary)
		log.Entry().Debug("Download completed")
	}
	return nil
}

func loadCertificates(certificateString string, client piperhttp.Downloader, runner execRunner) error {
	trustStoreFile := filepath.Join(getWorkingDir(), ".certificates", "cacerts")

	if exists, _ := fileUtilsExists(trustStoreFile); exists {
		// use local existing trust store
		sonar.addEnvironment("SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore=" + trustStoreFile)
		log.Entry().WithField("trust store", trustStoreFile).Info("Using local trust store")
	} else
	//TODO: certificate loading is deactivated due to the missing JAVA keytool
	// see https://github.com/SAP/jenkins-library/issues/1072
	if os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES") == "true" && len(certificateString) > 0 {
		// use local created trust store with downloaded certificates
		keytoolOptions := []string{
			"-import",
			"-noprompt",
			"-storepass", "changeit",
			"-keystore", trustStoreFile,
		}
		tmpFolder := getTempDir()
		defer os.RemoveAll(tmpFolder) // clean up
		certificateList := strings.Split(certificateString, ",")

		for _, certificate := range certificateList {
			filename := path.Base(certificate) // decode?
			target := filepath.Join(tmpFolder, filename)

			log.Entry().WithField("source", certificate).WithField("target", target).Info("Downloading TLS certificate")
			// download certificate
			if err := client.DownloadFile(certificate, target, nil, nil); err != nil {
				return errors.Wrapf(err, "Download of TLS certificate failed")
			}
			options := append(keytoolOptions, "-file", target)
			options = append(options, "-alias", filename)
			// add certificate to keystore
			if err := runner.RunExecutable("keytool", options...); err != nil {
				return errors.Wrap(err, "Adding certificate to keystore failed")
			}
		}
		sonar.addEnvironment("SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore=" + trustStoreFile)
		log.Entry().WithField("trust store", trustStoreFile).Info("Using local trust store")
	} else {
		log.Entry().Debug("Download of TLS certificates skipped")
	}
	return nil
}

func getWorkingDir() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Entry().WithError(err).WithField("path", workingDir).Debug("Retrieving of work directory failed")
	}
	return workingDir
}

func getTempDir() string {
	tmpFolder, err := ioutil.TempDir(".", "temp-")
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
	}
	return tmpFolder
}
