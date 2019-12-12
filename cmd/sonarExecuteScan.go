package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

const toolFolder = ".sonar-scanner"

func sonarExecuteScan(options sonarExecuteScanOptions) error {
	c := command.Command{}
	// reroute command output to loging framework
	// also log stdout as Karma reports into it
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())
	runSonar(options, &c)
	return nil
}

func runSonar(options sonarExecuteScanOptions, command execRunner) {
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

	loadSonarScanner(options.SonarScannerDownloadURL)

	//loadCertificates("", toolFolder)

	scan(arguments, command)
}

func loadSonarScanner(url string) {
	if len(url) > 0 {
		log.Entry().WithField("url", url).Debug("download Sonar scanner cli")
		// create temp folder to extract archive with CLI
		tmpFolder, err := ioutil.TempDir(".", "temp-")
		if err != nil {
			log.Entry().WithError(err).WithField("tempFolder", tmpFolder).Debug("creation of temp directory failed")
		}
		archive := filepath.Join(tmpFolder, path.Base(url))
		if err := DownloadFile(archive, url); err != nil {
			log.Entry().WithError(err).WithField("source", url).WithField("target", archive).
				Fatal("download of Sonar scanner cli failed")
		}
		if _, err := UnzipFile(archive, tmpFolder); err != nil {
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
func loadCertificates(certificateString string, toolFolder string) {
	if len(certificateString) > 0 {
		certificateFolder := ".certificates"

		//keystore := filepath.Join(toolFolder, "jre", "lib", "security", "cacerts")
		//keytoolOptions := []string{"-import", "-noprompt", "-storepass changeit", "-keystore " + keystore}
		certificateList := strings.Split(certificateString, ",")

		for _, certificate := range certificateList {
			filename := path.Base(certificate) // decode?

			log.Entry().
				WithField("filename", filename).
				Debug("download of TLS certificate")

			if err := DownloadFile(filepath.Join(certificateFolder, filename), certificate); err != nil {
				log.Entry().
					WithField("url", certificate).
					WithError(err).
					Fatal("download of TLS certificate failed")
			}
			// load
			// add to keytool
			// sh "keytool ${keytoolOptions.join(" ")} -alias "${filename}" -file "${certificateFolder}${filename}""
		}
	} else {
		log.Entry().
			WithField("certificates", certificateString).
			Debug("download of TLS certificates skipped")
	}
}

func scan(options []string, command execRunner) {
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

// extract to FileUtils
// https://golangcode.com/unzip-files-in-go/
// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func UnzipFile(src, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// extract to FileUtils
// https://golangcode.com/download-a-file-from-a-url/
// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
