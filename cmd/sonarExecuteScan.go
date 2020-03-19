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
	Binary      string
	Environment []string
	Options     []string
}

var sonar sonarSettings

var execLookPath = exec.LookPath
var fileUtilsExists = FileUtils.FileExists
var fileUtilsUnzip = FileUtils.Unzip
var osRename = os.Rename

func sonarExecuteScan(options sonarExecuteScanOptions, telemetryData *telemetry.CustomData) {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	client := piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{Timeout: time.Second * 180})

	sonar = sonarSettings{
		Binary:      "sonar-scanner",
		Environment: []string{},
		Options:     []string{},
	}

	if err := runSonar(options, &c, &client); err != nil {
		log.Entry().WithError(err).Fatal("Execution failed")
	}
}

func runSonar(options sonarExecuteScanOptions, runner execRunner, client piperhttp.Downloader) error {
	// Provided by withSonarQubeEnv:
	// - SONAR_CONFIG_NAME
	// - SONAR_EXTRA_PROPS
	// - SONAR_HOST_URL
	// - SONAR_AUTH_TOKEN
	// - SONARQUBE_SCANNER_PARAMS = { "sonar.host.url" : "https:\/\/sonar", "sonar.login" : "******"}
	// - SONAR_MAVEN_GOAL
	if len(options.Host) > 0 {
		sonar.Environment = append(sonar.Environment, "SONAR_HOST_URL="+options.Host)
	}
	//TODO: SONAR_AUTH_TOKEN or SONAR_TOKEN, both seem to work
	if len(options.Token) > 0 {
		sonar.Environment = append(sonar.Environment, "SONAR_TOKEN="+options.Token)
	}
	if len(options.Organization) > 0 {
		sonar.Options = append(sonar.Options, "sonar.organization="+options.Organization)
	}
	if len(options.ProjectVersion) > 0 {
		sonar.Options = append(sonar.Options, "sonar.projectVersion="+options.ProjectVersion)
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
		WithField("command", sonar.Binary).
		WithField("options", sonar.Options).
		WithField("environment", sonar.Environment).
		Debug("Executing sonar scan command")

	sonar.Options = SliceUtils.Prefix(sonar.Options, "-D")
	runner.SetEnv(sonar.Environment)
	return runner.RunExecutable(sonar.Binary, sonar.Options...)
}

func handlePullRequest(options sonarExecuteScanOptions) error {
	if len(options.ChangeID) > 0 {
		if options.LegacyPRHandling {
			// see https://docs.sonarqube.org/display/PLUG/GitHub+Plugin
			sonar.Options = append(sonar.Options, "sonar.analysis.mode=preview")
			sonar.Options = append(sonar.Options, "sonar.github.pullRequest="+options.ChangeID)
			if len(options.GithubAPIURL) > 0 {
				sonar.Options = append(sonar.Options, "sonar.github.endpoint="+options.GithubAPIURL)
			}
			if len(options.GithubToken) > 0 {
				sonar.Options = append(sonar.Options, "sonar.github.oauth="+options.GithubToken)
			}
			if len(options.Owner) > 0 && len(options.Repository) > 0 {
				sonar.Options = append(sonar.Options, "sonar.github.repository="+options.Owner+"/"+options.Repository)
			}
			if options.DisableInlineComments {
				sonar.Options = append(sonar.Options, "sonar.github.disableInlineComments="+strconv.FormatBool(options.DisableInlineComments))
			}
		} else {
			// see https://sonarcloud.io/documentation/analysis/pull-request/
			provider := strings.ToLower(options.PullRequestProvider)
			if provider == "github" {
				sonar.Options = append(sonar.Options, "sonar.pullrequest.github.repository="+options.Owner+"/"+options.Repository)
			} else {
				return errors.New("Pull-Request provider '" + provider + "' is not supported!")
			}
			sonar.Options = append(sonar.Options, "sonar.pullrequest.key="+options.ChangeID)
			sonar.Options = append(sonar.Options, "sonar.pullrequest.base="+options.ChangeTarget)
			sonar.Options = append(sonar.Options, "sonar.pullrequest.branch="+options.ChangeBranch)
			sonar.Options = append(sonar.Options, "sonar.pullrequest.provider="+provider)
		}
	}
	return nil
}

func loadSonarScanner(url string, client piperhttp.Downloader) error {
	if scannerPath, err := execLookPath(sonar.Binary); err == nil {
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
		sonar.Binary = filepath.Join(getWorkingDir(), toolPath, "bin", sonar.Binary)
		log.Entry().Debug("Download completed")
	}
	return nil
}

func loadCertificates(certificateString string, client piperhttp.Downloader, runner execRunner) error {
	trustStoreFile := filepath.Join(getWorkingDir(), ".certificates", "cacerts")

	if exists, _ := fileUtilsExists(trustStoreFile); exists {
		// use local existing trust store
		sonar.Environment = append(sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+trustStoreFile)
		log.Entry().WithField("trust store", trustStoreFile).Info("Using local trust store")
	} else
	//TODO: certificate loading is deactivated due to the missing JAVA keytool
	// see https://github.com/SAP/jenkins-library/issues/1072
	if os.Getenv("PIPER_SONAR_LOAD_CERTIFICATES") == "true" && len(certificateString) > 0 {
		// use local created trust store with downloaded certificates
		keytoolOptions := []string{
			"-import",
			"-noprompt",
			"-storepass changeit",
			"-keystore " + trustStoreFile,
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
			options := append(keytoolOptions, "-file \""+target+"\"")
			options = append(options, "-alias \""+filename+"\"")
			// add certificate to keystore
			if err := runner.RunExecutable("keytool", options...); err != nil {
				return errors.Wrap(err, "Adding certificate to keystore failed")
			}
		}
		sonar.Environment = append(sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+trustStoreFile)
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
