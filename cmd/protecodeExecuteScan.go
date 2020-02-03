package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/google/go-containerregistry/pkg/legacy/tarball"
	"github.com/google/go-containerregistry/pkg/name"
)

type protecodeData struct {
	Target                               string `json:"target,omitempty"`
	Mandatory                            bool   `json:"mandatory,omitempty"`
	ProductID                            string `json:"productID,omitempty"`
	ProtecodeServerURL                   string `json:"protecodeServerUrl,omitempty"`
	ProtecodeFailOnSevereVulnerabilities bool   `json:"protecodeFailOnSevereVulnerabilities,omitempty"`
	ProtecodeExcludeCVEs                 string `json:"protecodeExcludeCVEs,omitempty"`
	Count                                string `json:"count,omitempty"`
	Cvss2GreaterOrEqualSeven             string `json:"cvss2GreaterOrEqualSeven,omitempty"`
	Cvss3GreaterOrEqualSeven             string `json:"cvss3GreaterOrEqualSeven,omitempty"`
	ExcludedVulnerabilities              string `json:"excludedVulnerabilities,omitempty"`
	TriagedVulnerabilities               string `json:"triagedVulnerabilities,omitempty"`
	HistoricalVulnerabilities            string `json:"historicalVulnerabilities,omitempty"`
}

var cachePath = "./cache"
var cacheProtecodeImagePath = "/protecode/Image"
var cacheProtecodePath = "/protecode"

func protecodeExecuteScan(config protecodeExecuteScanOptions, influx *protecodeExecuteScanInflux) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	return runProtecodeScan(&config, influx)
}

func runProtecodeScan(config *protecodeExecuteScanOptions, influx *protecodeExecuteScanInflux) error {

	//create client for sending api request
	log.Entry().Debug("Protecode scan debug, create protecode client")
	client := createClient(config)

	log.Entry().Debugf("Protecode scan debug, get docker image: %v, %v, %v, %v", config.ScanImage, config.DockerRegistryURL, config.FilePath, config.IncludeLayers)
	image := getDockerImage(config.ScanImage, config.DockerRegistryURL, config.FilePath, config.IncludeLayers)

	artifactVersion := handleArtifactVersion(config.ArtifactVersion)
	fileName, filePath := createImageTar(image, config.ScanImage, config.FilePath, artifactVersion)

	if len(config.FilePath) <= 0 {
		(*config).FilePath = filePath
		log.Entry().Debugf("Protecode scan debug, filepath: %v", config.FilePath)
	}

	log.Entry().Debug("Protecode scan debug, execute protecode scan")
	parsedResult, productID := executeProtecodeScan(client, config, fileName, writeReportToFile)

	log.Entry().Debug("Protecode scan debug, write influx data")
	setInfluxData(influx, parsedResult)

	log.Entry().Debug("Protecode scan debug, write report to filesystem")
	writeReportDataToJSONFile(config, parsedResult, productID, ioutil.WriteFile)

	defer os.Remove(config.FilePath)

	deletePath := filepath.Join(cachePath, cacheProtecodePath)
	err := os.RemoveAll(deletePath)
	if err != nil {
		log.Entry().Warnf("Protecode scan debug, exception during cleanup folder %v", err)
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
		log.Entry().Fatalf("Protecode scan failed, exception during get docker image: %v", err)
	}

	return image
}

func getDockerImage(scanImage string, registryURL string, path string, includeLayers bool) pkgutil.Image {
	return getImage(scanImage, registryURL, path, includeLayers)
}

func createImageTar(image pkgutil.Image, fileName string, path string, artifactVersion string) (string, string) {

	if !pkgutil.IsTar(fileName) {
		fileName = fmt.Sprintf("%v%v.tar", strings.ReplaceAll(fileName, "/", "_"), artifactVersion)
		tarFileName := filepath.Join(cachePath, fileName)

		tarImageData(tarFileName, image)
	}

	var resultFilePath string

	if len(path) <= 0 {
		resultFilePath = filepath.Join(cachePath, fileName)
		if len(resultFilePath) <= 0 {
			log.Entry().Fatalf("Protecode scan failed, there is no file path configured  : %v (filename:%v, PSPath: %v)", path, fileName, image.FSPath)
		}
	}

	return fileName, resultFilePath
}

func tarImageData(tarFileName string, image pkgutil.Image) {
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		log.Entry().WithError(err).Fatal("Protecode scan failed, error during create tar for the docker image")
	}
	if err := os.Chmod(tarFileName, 0644); err != nil {
		log.Entry().WithError(err).Fatal("Protecode scan failed, error during create tar for the docker image")
	}
	defer tarFile.Close()

	reference, err := name.ParseReference(image.Digest.String(), name.WeakValidation)
	if err != nil {
		log.Entry().WithError(err).Fatal("Protecode scan failed, not possible to parse reference of docker image")
	}
	err = tarball.Write(reference, image.Image, tarFile)
	if err != nil {
		log.Entry().WithError(err).Fatal("Protecode scan failed, error during create tar archive of docker image via tarball")
	}
}

func getURLAndFileNameFromDockerImage(scanImage string, registryURL string, filePath string) string {

	completeURL := scanImage

	if len(registryURL) > 0 && len(filePath) <= 0 {
		if strings.HasSuffix(registryURL, "/") {
			completeURL = fmt.Sprintf("remote://%v%v", registryURL, scanImage)
		} else {
			completeURL = fmt.Sprintf("remote://%v/%v", registryURL, scanImage)
		}
	} else if len(filePath) > 0 {
		completeURL = filePath
		if !pkgutil.IsTar(filePath) {
			completeURL = fmt.Sprintf("daemon://%v", filePath)
		}
	}

	if len(completeURL) <= 0 {
		log.Entry().Fatal("Protecode scan failed, there is no scan image configured")
	}

	return completeURL
}

func executeProtecodeScan(client protecode.Protecode, config *protecodeExecuteScanOptions, fileName string, writeReportToFile func(resp io.ReadCloser, reportFileName string) error) (map[string]int, int) {

	var parsedResult map[string]int = make(map[string]int)
	//load existing product by filename
	log.Entry().Debug("Protecode scan debug, load existing product")
	productID := client.LoadExistingProduct(config.ProtecodeGroup, config.ReuseExisting)

	// check if no existing is found or reuse existing is false
	productID = uploadScanOrDeclareFetch(*config, productID, client, fileName)
	if productID <= 0 {
		log.Entry().Fatalf("Protecode scan failed, the product id is not valid (product id %v <= zero)", productID)
	}
	//pollForResult
	log.Entry().Debug("Protecode scan debug, poll for scan result")
	result := client.PollForResult(productID, config.ProtecodeTimeoutMinutes)

	jsonData, _ := json.Marshal(result)
	ioutil.WriteFile("protecodescan_vulns.json", jsonData, 0644)

	//check if result is ok else notify
	if len(result.Result.Status) > 0 && result.Result.Status == "F" {
		log.Entry().Fatal("Protecode scan failed, please check the log and protecode backend for more details.")
	}
	//loadReport
	log.Entry().Debug("Protecode scan debug, load report")
	resp := client.LoadReport(config.ReportFileName, productID)

	//save report to filesystem
	err := writeReportToFile(*resp, config.ReportFileName)
	if err != nil {
		return parsedResult, productID
	}
	//clean scan from server
	log.Entry().Debug("Protecode scan debug, delete scan")
	client.DeleteScan(config.CleanupMode, productID)

	//count vulnerabilities
	log.Entry().Debug("Protecode scan debug, parse result")
	parsedResult, _ = client.ParseResultForInflux(result.Result, config.ProtecodeExcludeCVEs)

	return parsedResult, productID
}

func setInfluxData(influx *protecodeExecuteScanInflux, result map[string]int) {

	influx.protecodeData.fields.historicalVulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
	influx.protecodeData.fields.triagedVulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	influx.protecodeData.fields.excludedVulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	influx.protecodeData.fields.minorVulnerabilities = fmt.Sprintf("%v", result["minor_vulnerabilities"])
	influx.protecodeData.fields.majorVulnerabilities = fmt.Sprintf("%v", result["major_vulnerabilities"])
	influx.protecodeData.fields.vulnerabilities = fmt.Sprintf("%v", result["vulnerabilities"])
}

func writeReportDataToJSONFile(config *protecodeExecuteScanOptions, result map[string]int, productID int, writeToFile func(f string, d []byte, p os.FileMode) error) {

	protecodeData := protecodeData{}
	protecodeData.ProtecodeServerURL = config.ProtecodeServerURL
	protecodeData.ProtecodeFailOnSevereVulnerabilities = config.ProtecodeFailOnSevereVulnerabilities
	protecodeData.ProtecodeExcludeCVEs = config.ProtecodeExcludeCVEs
	protecodeData.Target = config.ReportFileName
	protecodeData.Mandatory = true
	protecodeData.ProductID = fmt.Sprintf("%v", productID)
	protecodeData.Count = fmt.Sprintf("%v", result["count"])
	protecodeData.Cvss2GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss2GreaterOrEqualSeven"])
	protecodeData.Cvss3GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss3GreaterOrEqualSeven"])
	protecodeData.ExcludedVulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	protecodeData.TriagedVulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	protecodeData.HistoricalVulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])

	jsonData, _ := json.Marshal(protecodeData)

	log.Entry().Infof("Protecode scan info, %v %v of which %v had a CVSS v2 score >= 7.0 and %v had a CVSS v3 score >= 7.0.\n %v vulnerabilities were excluded via configuration (%v) and %v vulnerabilities were triaged via the webUI.\nIn addition %v historical vulnerabilities were spotted.",
		protecodeData.Count, "json.results.summary.verdict.detailed", protecodeData.Cvss2GreaterOrEqualSeven, protecodeData.Cvss3GreaterOrEqualSeven, protecodeData.ExcludedVulnerabilities, protecodeData.ProtecodeExcludeCVEs, protecodeData.TriagedVulnerabilities, protecodeData.HistoricalVulnerabilities)

	writeToFile("protecodescan_report.json", jsonData, 0644)
}

func createClient(config *protecodeExecuteScanOptions) protecode.Protecode {

	var duration time.Duration = time.Duration(time.Minute * 1)

	if len(config.ProtecodeTimeoutMinutes) > 0 {
		dur, err := time.ParseDuration(fmt.Sprintf("%vm", config.ProtecodeTimeoutMinutes))
		if err != nil {
			log.Entry().Warnf("Protecode scan failed, failed to parse timeout %v, switched back to default timeout %v minutes", config.ProtecodeTimeoutMinutes, duration)
		} else {
			duration = dur
		}
	}

	pc := protecode.Protecode{}

	protecodeOptions := protecode.Options{
		ServerURL: config.ProtecodeServerURL,
		Logger:    log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode"),
		Duration:  duration,
		Username:  config.User,
		Password:  config.Password,
	}

	pc.SetOptions(protecodeOptions)

	return pc
}

func uploadScanOrDeclareFetch(config protecodeExecuteScanOptions, productID int, client protecode.Protecode, filaName string) int {

	//check if the LoadExistingProduct) before returns an valid product id, than scip this
	if productID <= 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			log.Entry().Debug("Protecode scan debug, declare fetch url")
			resultData := client.DeclareFetchURL(config.CleanupMode, config.ProtecodeGroup, config.FetchURL)
			productID = resultData.ProductID

		} else {
			log.Entry().Debugf("Protecode scan debug, upload file path: %v", config.FilePath)
			if len(config.FilePath) <= 0 {
				log.Entry().Fatalf("Protecode scan failed, there is no file path configured for upload : %v", config.FilePath)
			}
			resultData := client.UploadScanFile(config.CleanupMode, config.ProtecodeGroup, config.FilePath, filaName)
			productID = resultData.Result.ProductID
		}
	}

	return productID
}

var writeReportToFile = func(resp io.ReadCloser, reportFileName string) error {
	f, err := os.Create(reportFileName)
	if err == nil {
		defer f.Close()
		_, err = io.Copy(f, resp)
	}

	return err
}
