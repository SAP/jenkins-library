package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/google/go-containerregistry/pkg/legacy/tarball"
	"github.com/google/go-containerregistry/pkg/name"
)

type protecodeData struct {
	Target                      string           `json:"target,omitempty"`
	Mandatory                   bool             `json:"mandatory,omitempty"`
	ProductID                   string           `json:"productID,omitempty"`
	ServerURL                   string           `json:"serverUrl,omitempty"`
	FailOnSevereVulnerabilities bool             `json:"failOnSevereVulnerabilities,omitempty"`
	ExcludeCVEs                 string           `json:"excludeCVEs,omitempty"`
	Count                       string           `json:"count,omitempty"`
	Cvss2GreaterOrEqualSeven    string           `json:"cvss2GreaterOrEqualSeven,omitempty"`
	Cvss3GreaterOrEqualSeven    string           `json:"cvss3GreaterOrEqualSeven,omitempty"`
	ExcludedVulnerabilities     string           `json:"excludedVulnerabilities,omitempty"`
	TriagedVulnerabilities      string           `json:"triagedVulnerabilities,omitempty"`
	HistoricalVulnerabilities   string           `json:"historicalVulnerabilities,omitempty"`
	Vulnerabilities             []protecode.Vuln `json:"Vulnerabilities,omitempty"`
}

var reportPath = "./"
var cachePath = "./cache"
var cacheProtecodeImagePath = "/protecode/Image"
var cacheProtecodePath = "/protecode"

func protecodeExecuteScan(config protecodeExecuteScanOptions, telemetryData *telemetry.CustomData, influx *protecodeExecuteScanInflux) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	return runProtecodeScan(&config, influx)
}

func runProtecodeScan(config *protecodeExecuteScanOptions, influx *protecodeExecuteScanInflux) error {

	//create client for sending api request
	log.Entry().Debug("Create protecode client")
	client := createClient(config)

	log.Entry().Debugf("Get docker image: %v, %v, %v, %v", config.ScanImage, config.DockerRegistryURL, config.FilePath, config.IncludeLayers)
	image := getDockerImage(config.ScanImage, config.DockerRegistryURL, config.FilePath, config.IncludeLayers)

	artifactVersion := handleArtifactVersion(config.ArtifactVersion)
	fileName, filePath := createImageTar(image, config.ScanImage, config.FilePath, artifactVersion)

	if len(config.FilePath) <= 0 {
		(*config).FilePath = filePath
		log.Entry().Debugf("Filepath for upload image: %v", config.FilePath)
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

func handleArtifactVersion(artifactVersion string) string {
	matches, _ := regexp.MatchString("([\\d\\.]){1,}-[\\d]{14}([\\Wa-z\\d]{41})?", artifactVersion)
	if matches {
		split := strings.SplitN(artifactVersion, ".", 2)
		return split[0]
	}

	return artifactVersion
}

var getImage = func(scanImage string, registryURL string, filePath string, includeLayers bool) pkgutil.Image {

	cacheImagePath := filepath.Join(cachePath, cacheProtecodeImagePath)
	deletePath := filepath.Join(cachePath, cacheProtecodePath)
	err := os.RemoveAll(deletePath)

	os.Mkdir(cacheImagePath, 600)

	completeURL := getURLAndFileNameFromDockerImage(scanImage, registryURL, filePath)
	image, err := pkgutil.GetImage(completeURL, includeLayers, cacheImagePath)
	if err != nil {
		log.Entry().Fatalf("Error during get docker image: %v", err)
	}

	return image
}

func getDockerImage(scanImage string, registryURL string, path string, includeLayers bool) pkgutil.Image {
	return getImage(scanImage, registryURL, path, includeLayers)
}

func createImageTar(image pkgutil.Image, fileName string, path string, artifactVersion string) (string, string) {

	if !pkgutil.IsTar(fileName) {

		fileName = fmt.Sprintf("%v%v.tar", strings.ReplaceAll(fileName, "/", "_"), strings.ReplaceAll(artifactVersion, ":", "_"))
		tarFileName := filepath.Join(cachePath, fileName)

		tarImageData(tarFileName, image)
	}

	resultFilePath := path

	if len(path) <= 0 {
		resultFilePath = cachePath
	}

	return fileName, resultFilePath
}

func tarImageData(tarFileName string, image pkgutil.Image) {
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		log.Entry().WithError(err).Fatal("Error during create tar for the docker image")
	}
	if err := os.Chmod(tarFileName, 0644); err != nil {
		log.Entry().WithError(err).Fatal("Error during create tar for the docker image")
	}
	defer tarFile.Close()

	reference, err := name.ParseReference(image.Digest.String(), name.WeakValidation)
	if err != nil {
		log.Entry().WithError(err).Fatal("It is not possible to parse reference of docker image")
	}
	err = tarball.Write(reference, image.Image, tarFile)
	if err != nil {
		log.Entry().WithError(err).Fatal("Error during create tar archive of docker image via tarball")
	}
}

func getURLAndFileNameFromDockerImage(scanImage string, registryURL string, filePath string) string {

	pathToImage := scanImage

	if len(registryURL) > 0 && len(filePath) <= 0 {
		registry := registryURL

		url, _ := url.Parse(registryURL)
		if len(url.Scheme) > 0 {
			registry = strings.Replace(registryURL, fmt.Sprintf("%v://", url.Scheme), "", 1)
		}

		if strings.HasSuffix(registry, "/") {
			pathToImage = fmt.Sprintf("remote://%v%v", registry, scanImage)
		} else {
			pathToImage = fmt.Sprintf("remote://%v/%v", registry, scanImage)
		}
	} else if len(filePath) > 0 {
		pathToImage = filePath
		if !pkgutil.IsTar(filePath) {
			pathToImage = fmt.Sprintf("daemon://%v", filePath)
		}
	}

	if len(pathToImage) <= 0 {
		log.Entry().Fatal("There is no scan image configured")
	}

	return pathToImage
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

	jsonData, _ := json.Marshal(result)
	filePath := filepath.Join(reportPath, "protecodescan_vulns.json")
	ioutil.WriteFile(filePath, jsonData, 0644)

	//check if result is ok else notify
	if len(result.Result.Status) > 0 && result.Result.Status == "F" {
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
	writeReportDataToJSONFile(config, parsedResult, productID, vulns, ioutil.WriteFile)

	return parsedResult
}

func setInfluxData(influx *protecodeExecuteScanInflux, result map[string]int) {

	influx.protecodeData.fields.historicalVulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
	influx.protecodeData.fields.triagedVulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	influx.protecodeData.fields.excludedVulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	influx.protecodeData.fields.minorVulnerabilities = fmt.Sprintf("%v", result["minor_vulnerabilities"])
	influx.protecodeData.fields.majorVulnerabilities = fmt.Sprintf("%v", result["major_vulnerabilities"])
	influx.protecodeData.fields.vulnerabilities = fmt.Sprintf("%v", result["vulnerabilities"])
}

func writeReportDataToJSONFile(config *protecodeExecuteScanOptions, result map[string]int, productID int, vulns []protecode.Vuln, writeToFile func(f string, d []byte, p os.FileMode) error) {

	protecodeData := protecodeData{}
	protecodeData.ServerURL = config.ServerURL
	protecodeData.FailOnSevereVulnerabilities = config.FailOnSevereVulnerabilities
	protecodeData.ExcludeCVEs = config.ExcludeCVEs
	protecodeData.Target = config.ReportFileName
	protecodeData.Mandatory = true
	protecodeData.ProductID = fmt.Sprintf("%v", productID)
	protecodeData.Count = fmt.Sprintf("%v", result["count"])
	protecodeData.Cvss2GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss2GreaterOrEqualSeven"])
	protecodeData.Cvss3GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss3GreaterOrEqualSeven"])
	protecodeData.ExcludedVulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	protecodeData.TriagedVulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	protecodeData.HistoricalVulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
	protecodeData.Vulnerabilities = vulns

	jsonData, _ := json.Marshal(protecodeData)

	log.Entry().Infof("Protecode scan info, %v of which %v had a CVSS v2 score >= 7.0 and %v had a CVSS v3 score >= 7.0.\n %v vulnerabilities were excluded via configuration (%v) and %v vulnerabilities were triaged via the webUI.\nIn addition %v historical vulnerabilities were spotted. \n\n Vulnerabilities: %v",
		protecodeData.Count, protecodeData.Cvss2GreaterOrEqualSeven, protecodeData.Cvss3GreaterOrEqualSeven, protecodeData.ExcludedVulnerabilities, protecodeData.ExcludeCVEs, protecodeData.TriagedVulnerabilities, protecodeData.HistoricalVulnerabilities, protecodeData.Vulnerabilities)

	filePath := filepath.Join(reportPath, "protecodeExecuteScan.json")
	writeToFile(filePath, jsonData, 0644)
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
		Username:  config.User,
		Password:  config.Password,
	}

	pc.SetOptions(protecodeOptions)

	return pc
}

func uploadScanOrDeclareFetch(config protecodeExecuteScanOptions, productID int, client protecode.Protecode, fileName string) int {

	//check if the LoadExistingProduct) before returns an valid product id, than scip this
	if !hasExisting(productID, config.ReuseExisting) {
		if len(config.FetchURL) > 0 {
			log.Entry().Debugf("Declare fetch url %v", config.FetchURL)
			resultData := client.DeclareFetchURL(config.CleanupMode, config.Group, config.FetchURL)
			productID = resultData.ProductID

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
