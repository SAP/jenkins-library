package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	file "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const toolFolder = ".sonar-scanner"

func sonarExecuteScan(options sonarExecuteScanOptions, telemetryData *telemetry.CustomData) error {
	c := command.Command{}
	// reroute command output to loging framework
	// also log stdout as Karma reports into it
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	client := piperhttp.Client{}

	runSonar(options, &c, &client)
	return nil
}

func runSonar(options sonarExecuteScanOptions, command execRunner, client piperhttp.Downloader) {
	arguments := []string{}

	// Provided by withSonarQubeEnv: SONAR_HOST_URL, SONAR_AUTH_TOKEN, SONARQUBE_SCANNER_PARAMS
	// SONARQUBE_SCANNER_PARAMS={ "sonar.host.url" : "https:\/\/sonar", "sonar.login" : "******"}
	//sonarHost := os.Getenv("SONAR_HOST_URL")
	if len(options.Host) > 0 {
		arguments = append(arguments, "sonar.host.url="+options.Host)
	}
	//sonarToken := os.Getenv("SONAR_AUTH_TOKEN")
	if len(options.Token) > 0 {
		arguments = append(arguments, "sonar.login="+options.Token)
	}
	if len(options.Organization) > 0 {
		arguments = append(arguments, "sonar.organization="+options.Organization)
	}
	if len(options.ProjectVersion) > 0 {
		arguments = append(arguments, "sonar.projectVersion="+options.ProjectVersion)
	}

	//if(configuration.options instanceof String)
	//configuration.options = [].plus(configuration.options)

	if len(options.ChangeID) > 0 {
		if options.LegacyPRHandling {
			// see https://docs.sonarqube.org/display/PLUG/GitHub+Plugin
			arguments = append(arguments, "sonar.analysis.mode=preview")
			arguments = append(arguments, "sonar.github.pullRequest="+options.ChangeID)

			//githubToken := os.Getenv("GITHUB_TOKEN")
			if len(options.GithubToken) > 0 {
				arguments = append(arguments, "sonar.github.oauth="+options.GithubToken)
			}
			arguments = append(arguments, "sonar.github.repository=${config.githubOrg}/${config.githubRepo}")
			if len(options.GithubAPIURL) > 0 {
				arguments = append(arguments, "sonar.github.endpoint="+options.GithubAPIURL)
			}
			if options.DisableInlineComments {
				arguments = append(arguments, "sonar.github.disableInlineComments="+strconv.FormatBool(options.DisableInlineComments))
			}
		} else {
			// see https://sonarcloud.io/documentation/analysis/pull-request/
			arguments = append(arguments, "sonar.pullrequest.key="+options.ChangeID)
			arguments = append(arguments, "sonar.pullrequest.base={{ env.CHANGE_toolFolder }}")
			arguments = append(arguments, "sonar.pullrequest.branch={{ env.CHANGE_BRANCH }}")
			arguments = append(arguments, "sonar.pullrequest.provider={{ options.pullRequestProvider }}")
			/*if options.PullRequestProvider == "GitHub" {
				arguments = append(arguments, "sonar.pullrequest.github.repository={{ options.githubOrg }}/{{ options.githubRepo }}")
			} else {
				log.Entry().Fatal("Pull-Request provider '{{ options.pullRequestProvider }}' is not supported!")
			}*/
		}
	}

	//loadSonarScanner(options.SonarScannerDownloadURL, client)

	loadCertificates(command, "", client)

	scan(command, arguments)
}

func loadSonarScanner(url string, client piperhttp.Downloader) {
	if len(url) > 0 {
		log.Entry().WithField("url", url).Debug("download Sonar scanner cli")
		// create temp folder to extract archive with CLI
		tmpFolder, err := ioutil.TempDir(".", "temp-")
		if err != nil {
			log.Entry().WithError(err).WithField("tempFolder", tmpFolder).Debug("creation of temp directory failed")
		}
		archive := filepath.Join(tmpFolder, path.Base(url))

		if err := client.DownloadFile(url, archive, nil, nil); err != nil {
			log.Entry().WithError(err).WithField("source", url).WithField("target", archive).
				Fatal("download of Sonar scanner cli failed")
		}
		if _, err := file.Unzip(archive, tmpFolder); err != nil {
			log.Entry().WithError(err).WithField("source", archive).WithField("target", tmpFolder).
				Fatal("extraction of Sonar scanner cli failed")
		}
		// derive foldername from archive
		foldername := strings.ReplaceAll(strings.ReplaceAll(archive, ".zip", ""), "cli-", "")
		if err := os.Rename(foldername, toolFolder); err != nil {
			log.Entry().WithError(err).WithField("source", foldername).WithField("target", toolFolder).
				Fatal("renaming of tool folder failed")
		}
		if err := os.Remove(tmpFolder); err != nil {
			log.Entry().WithError(err).WithField("target", tmpFolder).
				Warn("deletion of archive failed")
		}
		log.Entry().Debug("download completed")
	} else {
		log.Entry().WithField("url", url).Debug("download of Sonar scanner cli skipped")
	}
}

//TODO: extract to Helper?
func loadCertificates(command execRunner, certificateString string, client piperhttp.Downloader) {
	if len(certificateString) > 0 {
		//certificateFolder := ".certificates"

		// create temp folder to extract archive with CLI
		tmpFolder, err := ioutil.TempDir(".", "temp-")
		if err != nil {
			log.Entry().WithError(err).WithField("tempFolder", tmpFolder).Debug("creation of temp directory failed")
		}

		keystore := filepath.Join(toolFolder, "jre", "lib", "security", "cacerts")
		keytoolOptions := []string{"-import", "-noprompt", "-storepass changeit", "-keystore " + keystore}
		certificateList := strings.Split(certificateString, ",")

		for _, certificate := range certificateList {
			filename := path.Base(certificate) // decode?
			target := filepath.Join(tmpFolder, filename)

			log.Entry().
				WithField("source", certificate).
				WithField("target", target).
				Info("download of TLS certificate")
			// download certificate
			if err := client.DownloadFile(certificate, target, nil, nil); err != nil {
				log.Entry().
					WithField("url", certificate).
					WithError(err).
					Fatal("download of TLS certificate failed")
			}
			options := append(keytoolOptions, "-file \""+target+"\"")
			options = append(options, "-alias \""+filename+"\"")
			// add certificate to keystore
			if err := command.RunExecutable("keytool", keytoolOptions...); err != nil {
				log.Entry().WithError(err).WithField("source", target).Fatal("adding certificate to keystore failed")
			}
			// sh "keytool ${keytoolOptions.join(" ")} -alias "${filename}" -file "${certificateFolder}${filename}""
		}
	} else {
		log.Entry().
			WithField("certificates", certificateString).
			Debug("download of TLS certificates skipped")
	}
}

func scan(command execRunner, options []string) {
	executable := filepath.Join(toolFolder, "bin", "sonar-scanner")
	for idx, element := range options {
		element = strings.TrimSpace(element)
		if !strings.HasPrefix(element, "-D") {
			element = "-D" + element
		}
		options[idx] = element
	}
	log.Entry().
		WithField("command", executable).
		WithField("options", strings.Join(options, " ")).
		Debug("executing sonar scan command")

	if err := command.RunExecutable(executable, options...); err != nil {
		log.Entry().WithError(err).Fatal("failed to execute scan command")
	}
}

func setOption(options *[]string, id, value string) {
	if len(value) > 0 {
		o := append(*options, "sonar."+id+"="+value)
		options = &o
	}
}
