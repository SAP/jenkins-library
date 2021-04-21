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

	"github.com/pkg/errors"

	"github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/command"
	piperDocker "github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	StepResults "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const (
	webReportPath  = "%s/product/%v/"
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
		log.Entry().WithError(err).Fatal("Failed to execute protecode scan.")
	}
}

func runProtecodeScan(config *protecodeExecuteScanOptions, influx *protecodeExecuteScanInflux, dClient piperDocker.Download) error {
	correctDockerConfigEnvVar(config)
	var fileName, filePath string
	var err error
	//create client for sending api request
	log.Entry().Debug("Create protecode client")
	client := createClient(config)
	if len(config.FetchURL) <= 0 {
		log.Entry().Debugf("Get docker image: %v, %v, %v, %v", config.ScanImage, config.DockerRegistryURL, config.FilePath, config.IncludeLayers)
		fileName, filePath, err = getDockerImage(dClient, config)
		if err != nil {
			return errors.Wrap(err, "failed to get Docker image")
		}
		if len(config.FilePath) <= 0 {
			(*config).FilePath = filePath
			log.Entry().Debugf("Filepath for upload image: %v", config.FilePath)
		}
	} else {
		// Get filename from a fetch URL
		fileName = filepath.Base(config.FetchURL)
		log.Entry().Debugf("[DEBUG] ===> Filepath from fetch URL: %v", fileName)
	}

	log.Entry().Debug("Execute protecode scan")
	if err := executeProtecodeScan(influx, client, config, fileName, writeReportToFile); err != nil {
		return err
	}

	defer os.Remove(config.FilePath)

	if err := os.RemoveAll(filepath.Join(cachePath, cacheProtecodePath)); err != nil {
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

func getDockerImage(dClient piperDocker.Download, config *protecodeExecuteScanOptions) (string, string, error) {

	cacheImagePath := filepath.Join(cachePath, cacheProtecodeImagePath)
	deletePath := filepath.Join(cachePath, cacheProtecodePath)
	err := os.RemoveAll(deletePath)

	os.Mkdir(cacheImagePath, 600)

	imageSource, err := dClient.GetImageSource()
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", "", errors.Wrap(err, "failed to get docker image")
	}
	image, err := dClient.DownloadImageToPath(imageSource, cacheImagePath)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to download docker image")
	}

	var fileName string
	if util.IsTar(config.ScanImage) {
		fileName = config.ScanImage
	} else {
		fileName = getTarName(config)
		tarFilePath := filepath.Join(cachePath, fileName)
		tarFile, err := os.Create(tarFilePath)
		if err != nil {
			log.SetErrorCategory(log.ErrorCustom)
			return "", "", errors.Wrap(err, "failed to create tar for the docker image")
		}
		defer tarFile.Close()
		if err := os.Chmod(tarFilePath, 0644); err != nil {
			log.SetErrorCategory(log.ErrorCustom)
			return "", "", errors.Wrap(err, "failed to set permissions on tar for the docker image")
		}
		if err = dClient.TarImage(tarFile, image); err != nil {
			return "", "", errors.Wrap(err, "failed to tar the docker image")
		}
	}

	resultFilePath := config.FilePath

	if len(config.FilePath) <= 0 {
		resultFilePath = cachePath
	}

	return fileName, resultFilePath, nil
}

func executeProtecodeScan(influx *protecodeExecuteScanInflux, client protecode.Protecode, config *protecodeExecuteScanOptions, fileName string, writeReportToFile func(resp io.ReadCloser, reportFileName string) error) error {

	log.Entry().Debugf("[DEBUG] ===> Load existing product Group:%v Reuse:%v and Filenam:%v", config.Group, config.ReuseExisting, fileName)

	// Get existing product id by filename
	productID := client.LoadExistingProduct(config.Group, fileName)

	log.Entry().Debugf("[DEBUG] ===> Returned productID: %v", productID)

	// check if no existing is found or reuse existing is false
	productID = uploadScanOrDeclareFetch(*config, productID, client, fileName)

	log.Entry().Debugf("[DEBUG] ===> After 'uploadScanOrDeclareFetch' returned productID: %v", productID)

	if productID <= 0 {
		return fmt.Errorf("the product id is not valid '%d'", productID)
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
		return fmt.Errorf("protecode scan failed: %v/products/%v", config.ServerURL, productID)
	}

	//loadReport
	log.Entry().Debugf("Load report %v for %v", config.ReportFileName, productID)
	resp := client.LoadReport(config.ReportFileName, productID)
	//save report to filesystem
	if err := writeReportToFile(*resp, config.ReportFileName); err != nil {
		log.Entry().Warningf("failed to write report: %s", err)
	}
	//clean scan from server
	log.Entry().Debugf("Delete scan %v for %v", config.CleanupMode, productID)
	client.DeleteScan(config.CleanupMode, productID)

	//count vulnerabilities
	log.Entry().Debug("Parse scan result")
	parsedResult, vulns := client.ParseResultForInflux(result.Result, config.ExcludeCVEs)

	log.Entry().Debug("Write report to filesystem")
	if err := protecode.WriteReport(
		protecode.ReportData{
			ServerURL:                   config.ServerURL,
			FailOnSevereVulnerabilities: config.FailOnSevereVulnerabilities,
			ExcludeCVEs:                 config.ExcludeCVEs,
			Target:                      config.ReportFileName,
			Vulnerabilities:             vulns,
			ProductID:                   fmt.Sprintf("%v", productID),
		}, reportPath, stepResultFile, parsedResult, ioutil.WriteFile); err != nil {
		log.Entry().Warningf("failed to write report: %v", err)
	}

	log.Entry().Debug("Write influx data")
	setInfluxData(influx, parsedResult)

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
		return fmt.Errorf("the product is not compliant")
	}
	return nil
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
	//if !hasExisting(productID, config.ReuseExisting) {

	log.Entry().Debugf("[DEBUG] ===> In uploadScanOrDeclareFetch: %v", productID)

	// check if product doesn't exist then create a new one.
	if productID <= 0 {
		log.Entry().Debugf("[DEBUG] ===> New product creation started: %v", productID)
		productID = uploadFile(config, productID, client, fileName, false)

		log.Entry().Debugf("[DEBUG] ===> After uploading [productID < 0] file returned productID: %v", productID)
		return productID

		// In case product already exists and "reuseExisting" is false then we replace binary without creating a new product.
	} else if (productID > 0) && !config.ReuseExisting {
		log.Entry().Debugf("[DEBUG] ===> Replace binary entry point started %v", productID)
		productID = uploadFile(config, productID, client, fileName, true)

		log.Entry().Debugf("[DEBUG] ===> After uploading file [(productID > 0) && !config.ReuseExisting] returned productID: %v", productID)
		return productID

		// If product already exists and "reuseExisting" option is enabled then return the latest similar scan result.
	} else {
		log.Entry().Debugf("[DEBUG] ===> reuseExisting option is enabled and returned productID: %v", productID)
		return productID
	}

}

func uploadFile(config protecodeExecuteScanOptions, productID int, client protecode.Protecode, fileName string, replaceBinary bool) int {

	if len(config.FetchURL) > 0 {
		log.Entry().Debugf("Declare fetch url %v", config.FetchURL)
		resultData := client.DeclareFetchURL(config.CleanupMode, config.Group, config.FetchURL, productID, replaceBinary)
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

		resultData := client.UploadScanFile(config.CleanupMode, config.Group, pathToFile, combinedFileName, productID, replaceBinary)
		productID = resultData.Result.ProductID
		log.Entry().Debugf("[DEBUG] ===> uploadFile return FINAL product id: %v", productID)
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

func getTarName(config *protecodeExecuteScanOptions) string {
	// remove original version
	fileName := strings.TrimSuffix(config.ScanImage, ":"+config.ArtifactVersion)
	// remove sha digest if exists
	sha256 := "@sha256"
	if index := strings.Index(fileName, sha256); index > -1 {
		fileName = fileName[:index]
	}
	// append trimmed version
	if version := handleArtifactVersion(config.ArtifactVersion); len(version) > 0 {
		fileName = fileName + "_" + version
	}
	fileName = strings.ReplaceAll(fileName, "/", "_")
	return fileName + ".tar"
}
