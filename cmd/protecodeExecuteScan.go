package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperDocker "github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	StepResults "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	webReportPath  = "%s/products/%v/"
	scanResultFile = "protecodescan_vulns.json"
	stepResultFile = "protecodeExecuteScan.json"
)

var reportPath = "./"
var cachePath = "./cache"
var cacheProtecodeImagePath = "/protecode/Image"
var cacheProtecodePath = "/protecode"

func protecodeExecuteScan(config protecodeExecuteScanOptions, telemetryData *telemetry.CustomData, influx *protecodeExecuteScanInflux) {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	dClient := createDockerClient(&config)
	if err := runProtecodeScan(&config, influx, dClient); err != nil {
		log.Entry().WithError(err).Fatal("Execution failed")
	}
}

func runProtecodeScan(config *protecodeExecuteScanOptions, influx *protecodeExecuteScanInflux, dClient piperDocker.Download) error {
	correctDockerConfigEnvVar(config)
	var fileName, filePath string
	//create client for sending api request
	log.Entry().Debug("Create protecode client")
	client := createClient(config)
	if len(config.FetchURL) <= 0 {
		log.Entry().Debugf("Get docker image: %v, %v, %v, %v", config.ScanImage, config.DockerRegistryURL, config.FilePath, config.IncludeLayers)
		fileName, filePath = getDockerImage(dClient, config)
		if len(config.FilePath) <= 0 {
			(*config).FilePath = filePath
			log.Entry().Debugf("Filepath for upload image: %v", config.FilePath)
		}
	}

	log.Entry().Debug("Execute protecode scan")
	parsedResult := executeProtecodeScan(client, config, fileName, writeReportToFile)

	log.Entry().Debug("Write influx data")
	setInfluxData(influx, parsedResult)

	defer os.Remove(config.FilePath)

	deletePath := filepath.Join(cachePath, cacheProtecodePath)
	err := os.RemoveAll(deletePath)
	if err != nil {
		log.Entry().Warnf("Error during cleanup folder %v", err)
	}

	return nil
}

// reused by cmd/sonarExecuteScan.go
// TODO: extract to version utils
func handleArtifactVersion(artifactVersion string) string {
	matches, _ := regexp.MatchString("([\\d\\.]){1,}-[\\d]{14}([\\Wa-z\\d]{41})?", artifactVersion)
	if matches {
		split := strings.SplitN(artifactVersion, ".", 2)
		log.Entry().WithField("old", artifactVersion).WithField("new", split[0]).Debug("Trimming version to major version digit.")
		return split[0]
	}
	return artifactVersion
}

func getDockerImage(dClient piperDocker.Download, config *protecodeExecuteScanOptions) (string, string) {

	cacheImagePath := filepath.Join(cachePath, cacheProtecodeImagePath)
	deletePath := filepath.Join(cachePath, cacheProtecodePath)
	err := os.RemoveAll(deletePath)

	os.Mkdir(cacheImagePath, 600)

	imageSource, err := dClient.GetImageSource()
	if err != nil {
		log.Entry().WithError(err).Fatal("Error during get docker image source")
	}
	image, err := dClient.DownloadImageToPath(imageSource, cacheImagePath)
	if err != nil {
		log.Entry().Fatalf("Error during get docker image: %v", err)
	}

	tarFile, fileName := tarImage(config)
	if tarFile != nil {
		defer tarFile.Close()
		err = dClient.TarImage(tarFile, image)
		if err != nil {
			log.Entry().WithError(err).Fatal("Error during tar the docker image")
		}
	}

	resultFilePath := config.FilePath

	if len(config.FilePath) <= 0 {
		resultFilePath = cachePath
	}

	return fileName, resultFilePath
}

func tarImage(config *protecodeExecuteScanOptions) (*os.File, string) {
	if !(filepath.Ext(config.ScanImage) == ".tar" ||
		filepath.Ext(config.ScanImage) == ".tar.gz" ||
		filepath.Ext(config.ScanImage) == ".tgz") {

		artifactVersion := handleArtifactVersion(config.ArtifactVersion)
		fileName := fmt.Sprintf("%v%v.tar", strings.ReplaceAll(config.ScanImage, "/", "_"), strings.ReplaceAll(artifactVersion, ":", "_"))
		tarFileName := filepath.Join(cachePath, fileName)

		tarFile, err := os.Create(tarFileName)
		if err != nil {
			log.Entry().WithError(err).Fatal("Error during create tar for the docker image")
		}
		if err := os.Chmod(tarFileName, 0644); err != nil {
			log.Entry().WithError(err).Fatal("Error during create tar for the docker image")
		}

		return tarFile, fileName
	}

	return nil, config.ScanImage
}

func executeProtecodeScan(client protecode.Protecode, config *protecodeExecuteScanOptions, fileName string, writeReportToFile func(resp io.ReadCloser, reportFileName string) error) map[string]int {

	var parsedResult map[string]int = make(map[string]int)
	//load existing product by filename
	log.Entry().Debugf("Load existing product Group:%v Reuse:%v", config.Group, config.ReuseExisting)
	productID := client.LoadExistingProduct(config.Group, config.ReuseExisting)

	// check if no existing is found or reuse existing is false
	productID = uploadScanOrDeclareFetch(*config, productID, client, fileName)
	if productID <= 0 {
		log.Entry().Fatalf("The product id is not valid (product id %v <= zero)", productID)
	}
	//pollForResult
	log.Entry().Debugf("Poll for scan result %v", productID)
	result := client.PollForResult(productID, config.TimeoutMinutes)
	// write results to file
	jsonData, _ := json.Marshal(result)
	ioutil.WriteFile(filepath.Join(reportPath, scanResultFile), jsonData, 0644)

	//check if result is ok else notify
	if protecode.HasFailed(result) {
		log.SetErrorCategory(log.ErrorService)
		log.Entry().Fatalf("Please check the log and protecode backend for more details. URL: %v/products/%v", config.ServerURL, productID)
	}

	//loadReport
	log.Entry().Debugf("Load report %v for %v", config.ReportFileName, productID)
	resp := client.LoadReport(config.ReportFileName, productID)
	//save report to filesystem
	err := writeReportToFile(*resp, config.ReportFileName)
	if err != nil {
		return parsedResult
	}
	//clean scan from server
	log.Entry().Debugf("Delete scan %v for %v", config.CleanupMode, productID)
	client.DeleteScan(config.CleanupMode, productID)

	//count vulnerabilities
	log.Entry().Debug("Parse scan reult")
	parsedResult, vulns := client.ParseResultForInflux(result.Result, config.ExcludeCVEs)

	log.Entry().Debug("Write report to filesystem")
	err = protecode.WriteReport(
		protecode.ReportData{
			ServerURL:                   config.ServerURL,
			FailOnSevereVulnerabilities: config.FailOnSevereVulnerabilities,
			ExcludeCVEs:                 config.ExcludeCVEs,
			Target:                      config.ReportFileName,
			Vulnerabilities:             vulns,
			ProductID:                   fmt.Sprintf("%v", productID),
		}, reportPath, stepResultFile, parsedResult, ioutil.WriteFile)
	if err != nil {
		log.Entry().Warningf("failed to write report: %v", err)
	}

	// write reports JSON
	reports := []StepResults.Path{
		{Target: config.ReportFileName, Mandatory: true},
		{Target: stepResultFile, Mandatory: true},
		{Target: scanResultFile, Mandatory: true},
	}
	// write links JSON
	links := []StepResults.Path{
		{Name: "Protecode WebUI", Target: fmt.Sprintf(webReportPath, config.ServerURL, productID)},
		{Name: "Protecode Report", Target: path.Join("artifact", config.ReportFileName), Scope: "job"},
	}
	StepResults.PersistReportsAndLinks("protecodeExecuteScan", "", reports, links)

	if config.FailOnSevereVulnerabilities && protecode.HasSevereVulnerabilities(result.Result, config.ExcludeCVEs) {
		log.SetErrorCategory(log.ErrorCompliance)
		log.Entry().Fatalf("The product is not compliant. URL: %v/products/%v", config.ServerURL, productID)
	}

	return parsedResult
}

func setInfluxData(influx *protecodeExecuteScanInflux, result map[string]int) {
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
	influx.protecode_data.fields.triaged_vulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	influx.protecode_data.fields.excluded_vulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	influx.protecode_data.fields.minor_vulnerabilities = fmt.Sprintf("%v", result["minor_vulnerabilities"])
	influx.protecode_data.fields.major_vulnerabilities = fmt.Sprintf("%v", result["major_vulnerabilities"])
	influx.protecode_data.fields.vulnerabilities = fmt.Sprintf("%v", result["vulnerabilities"])
}

func createClient(config *protecodeExecuteScanOptions) protecode.Protecode {

	var duration time.Duration = time.Duration(time.Minute * 1)

	if len(config.TimeoutMinutes) > 0 {
		dur, err := time.ParseDuration(fmt.Sprintf("%vm", config.TimeoutMinutes))
		if err != nil {
			log.Entry().Warnf("Failed to parse timeout %v, switched back to default timeout %v minutes", config.TimeoutMinutes, duration)
		} else {
			duration = dur
		}
	}

	pc := protecode.Protecode{}

	protecodeOptions := protecode.Options{
		ServerURL: config.ServerURL,
		Logger:    log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode"),
		Duration:  duration,
		Username:  config.Username,
		Password:  config.Password,
	}

	pc.SetOptions(protecodeOptions)

	return pc
}
func createDockerClient(config *protecodeExecuteScanOptions) piperDocker.Download {

	dClientOptions := piperDocker.ClientOptions{ImageName: config.ScanImage, RegistryURL: config.DockerRegistryURL, LocalPath: config.FilePath, IncludeLayers: config.IncludeLayers}
	dClient := &piperDocker.Client{}
	dClient.SetOptions(dClientOptions)

	return dClient
}

func uploadScanOrDeclareFetch(config protecodeExecuteScanOptions, productID int, client protecode.Protecode, fileName string) int {
	//check if the LoadExistingProduct) before returns an valid product id, than scip this
	if !hasExisting(productID, config.ReuseExisting) {
		if len(config.FetchURL) > 0 {
			log.Entry().Debugf("Declare fetch url %v", config.FetchURL)
			resultData := client.DeclareFetchURL(config.CleanupMode, config.Group, config.FetchURL)
			productID = resultData.Result.ProductID
		} else {
			log.Entry().Debugf("Upload file path: %v", config.FilePath)
			if len(config.FilePath) <= 0 {
				log.Entry().Fatalf("There is no file path configured for upload : %v", config.FilePath)
			}
			pathToFile := filepath.Join(config.FilePath, fileName)
			if !(fileExists(pathToFile)) {
				log.Entry().Fatalf("There is no file for upload: %v", pathToFile)
			}

			combinedFileName := fileName
			if len(config.PullRequestName) > 0 {
				combinedFileName = fmt.Sprintf("%v_%v", config.PullRequestName, fileName)
			}

			resultData := client.UploadScanFile(config.CleanupMode, config.Group, pathToFile, combinedFileName)
			productID = resultData.Result.ProductID
		}
	}
	return productID
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func hasExisting(productID int, reuseExisting bool) bool {
	if (productID > 0) || reuseExisting {
		return true
	}
	return false
}

var writeReportToFile = func(resp io.ReadCloser, reportFileName string) error {
	filePath := filepath.Join(reportPath, reportFileName)
	f, err := os.Create(filePath)
	if err == nil {
		defer f.Close()
		_, err = io.Copy(f, resp)
	}

	return err
}

func correctDockerConfigEnvVar(config *protecodeExecuteScanOptions) {
	path := config.DockerConfigJSON
	if len(path) > 0 {
		log.Entry().Infof("Docker credentials configuration: %v", path)
		path, _ := filepath.Abs(path)
		// use parent directory
		path = filepath.Dir(path)
		os.Setenv("DOCKER_CONFIG", path)
	} else {
		log.Entry().Info("Docker credentials configuration: NONE")
	}
}
