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

func sonarExecuteScan(options sonarExecuteScanOptions, telemetryData *telemetry.CustomData) error {
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

	runSonar(options, &c, &client)
	return nil
}

func runSonar(options sonarExecuteScanOptions, runner execRunner, client piperhttp.Downloader) {
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

	handlePullRequest(options)

	loadSonarScanner(options.SonarScannerDownloadURL, client)

	loadCertificates(runner, "", client)

	log.Entry().
		WithField("command", sonar.Binary).
		WithField("options", sonar.Options).
		WithField("environment", sonar.Environment).
		Debug("Executing sonar scan command")

	runner.SetEnv(sonar.Environment)

	sonar.Options = SliceUtils.Prefix(sonar.Options, "-D")
	if err := runner.RunExecutable(sonar.Binary, sonar.Options...); err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute scan command")
	}
}

func handlePullRequest(options sonarExecuteScanOptions) {
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
			sonar.Options = append(sonar.Options, "sonar.pullrequest.key="+options.ChangeID)
			sonar.Options = append(sonar.Options, "sonar.pullrequest.base="+options.ChangeTarget)
			sonar.Options = append(sonar.Options, "sonar.pullrequest.branch="+options.ChangeBranch)
			sonar.Options = append(sonar.Options, "sonar.pullrequest.provider="+options.PullRequestProvider)
			if options.PullRequestProvider == "GitHub" {
				sonar.Options = append(sonar.Options, "sonar.pullrequest.github.repository="+options.Owner+"/"+options.Repository)
			} else {
				log.Entry().Fatal("Pull-Request provider '" + options.PullRequestProvider + "' is not supported!")
			}
		}
	}
}

func loadSonarScanner(url string, client piperhttp.Downloader) {
	if scannerPath, err := execLookPath(sonar.Binary); err == nil {
		// using existing sonar-scanner
		log.Entry().WithField("path", scannerPath).Debug("Using local Sonar scanner cli")
	} else {
		// download sonar-scanner-cli from url to .sonar-scanner/
		log.Entry().WithField("url", url).Debug("Downloading Sonar scanner cli")
		if len(url) == 0 {
			log.Entry().Error("Download url for Sonar scanner cli missing")
		}
		// download sonar-scanner-cli into TEMP folder
		tmpFolder := getTempDir()
		defer os.RemoveAll(tmpFolder) // clean up
		archive := filepath.Join(tmpFolder, path.Base(url))
		if err := client.DownloadFile(url, archive, nil, nil); err != nil {
			log.Entry().WithError(err).
				WithField("source", url).
				WithField("target", archive).
				Fatal("Download of Sonar scanner cli failed")
		}
		// unzip sonar-scanner-cli
		if _, err := fileUtilsUnzip(archive, tmpFolder); err != nil {
			log.Entry().WithError(err).
				WithField("source", archive).
				WithField("target", tmpFolder).
				Fatal("Extraction of Sonar scanner cli failed")
		}
		// move sonar-scanner-cli to .sonar-scanner/
		toolPath := ".sonar-scanner"
		foldername := strings.ReplaceAll(strings.ReplaceAll(archive, ".zip", ""), "cli-", "")
		if err := os.Rename(foldername, toolPath); err != nil {
			log.Entry().WithError(err).
				WithField("source", foldername).
				WithField("target", toolPath).
				Fatal("Renaming of tool folder failed")
		}
		// remove TEMP folder
		if err := os.Remove(tmpFolder); err != nil {
			log.Entry().WithError(err).WithField("target", tmpFolder).
				Warn("Deletion of archive failed")
		}
		// update binary path
		sonar.Binary = filepath.Join(getWorkingDir(), toolPath, "bin", sonar.Binary)
		log.Entry().Debug("Download completed")
	}
}

func loadCertificates(runner execRunner, certificateString string, client piperhttp.Downloader) {
	certPath := ".certificates"
	workingDir := getWorkingDir()
	if len(certificateString) > 0 {
		// create temp folder to extract archive with CLI
		tmpFolder := getTempDir()
		defer os.RemoveAll(tmpFolder) // clean up
		keystore := filepath.Join(workingDir, certPath, "cacerts")
		keytoolOptions := []string{"-import", "-noprompt", "-storepass changeit", "-keystore " + keystore}
		certificateList := strings.Split(certificateString, ",")

		for _, certificate := range certificateList {
			filename := path.Base(certificate) // decode?
			target := filepath.Join(tmpFolder, filename)

			log.Entry().
				WithField("source", certificate).
				WithField("target", target).
				Info("Download of TLS certificate")
			// download certificate
			if err := client.DownloadFile(certificate, target, nil, nil); err != nil {
				log.Entry().
					WithField("url", certificate).
					WithError(err).
					Fatal("Download of TLS certificate failed")
			}
			options := append(keytoolOptions, "-file \""+target+"\"")
			options = append(options, "-alias \""+filename+"\"")
			// add certificate to keystore
			if err := runner.RunExecutable("keytool", keytoolOptions...); err != nil {
				log.Entry().WithError(err).WithField("source", target).Fatal("Adding certificate to keystore failed")
			}
			// sh "keytool ${keytoolOptions.join(" ")} -alias "${filename}" -file "${certPath}${filename}""
		}
	} else {
		log.Entry().Debug("Download of TLS certificates skipped")
	}
	// use custom trust store
	trustStoreFile := filepath.Join(workingDir, certPath, "cacerts")
	if exists, _ := fileUtilsExists(filepath.Join(workingDir, certPath, "cacerts")); exists {
		log.Entry().
			WithField("trust store", trustStoreFile).
			Debug("Using local trust store")
		sonar.Environment = append(sonar.Environment, "SONAR_SCANNER_OPTS=-Djavax.net.ssl.trustStore="+trustStoreFile)
	}
}

func getWorkingDir() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Entry().WithError(err).
			WithField("path", workingDir).
			Debug("Retrieving of work directory failed")
	}
	return workingDir
}

func getTempDir() string {
	// create temp folder
	tmpFolder, err := ioutil.TempDir("", "temp-")
	if err != nil {
		log.Entry().WithError(err).
			WithField("path", tmpFolder).
			Debug("Creation of temp directory failed")
	}
	return tmpFolder
}
