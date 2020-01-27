package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
	Target                    string `json:"target,omitempty"`
	Mandatory                 bool   `json:"mandatory,omitempty"`
	ProductID                 string `json:"productID,omitempty"`
	ProtecodeServerURL        string `json:"protecodeServerUrl,omitempty"`
	Count                     string `json:"count,omitempty"`
	Cvss2GreaterOrEqualSeven  string `json:"cvss2GreaterOrEqualSeven,omitempty"`
	Cvss3GreaterOrEqualSeven  string `json:"cvss3GreaterOrEqualSeven,omitempty"`
	ExcludedVulnerabilities   string `json:"excludedVulnerabilities,omitempty"`
	TriagedVulnerabilities    string `json:"triagedVulnerabilities,omitempty"`
	HistoricalVulnerabilities string `json:"historicalVulnerabilities,omitempty"`
}

func protecodeExecuteScan(config protecodeExecuteScanOptions, influx *protecodeExecuteScanInflux) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	return runProtecodeScan(&config, influx)
}

func runProtecodeScan(config *protecodeExecuteScanOptions, influx *protecodeExecuteScanInflux) error {

	//create client for sending api request
	if config.Verbose {
		log.Entry().Info("Protecode scan debug, create protecode client")
	}
	client := createClient(config)

	if config.Verbose {
		log.Entry().Info("Protecode scan debug, get docker image")
	}
	fileName := getDockerImage(config, client)

	if config.Verbose {
		log.Entry().Info("Protecode scan debug, execute protecode scan")
	}
	parsedResult, productID := executeProtecodeScan(client, config, fileName, writeReportToFile)

	if config.Verbose {
		log.Entry().Info("Protecode scan debug, write influx data")
	}
	setInfluxData(influx, parsedResult)

	if config.Verbose {
		log.Entry().Info("Protecode scan debug, write report to filesystem")
	}
	writeReportDataToJSONFile(config, parsedResult, productID, ioutil.WriteFile)

	err := os.Remove(config.FilePath)
	if err != nil {
		log.Entry().WithError(err).Warnf("Protecode scan warning, failed to delete tar source code")
	}

	return nil
}

func getDockerImage(config *protecodeExecuteScanOptions, client protecode.Protecode) string {
	fileName := config.ScanImage

	cachePath := "./cache"
	cacheProtecodeImagePath := "/protecodeImage"
	cacheImagePath := filepath.Join(cachePath, cacheProtecodeImagePath)
	os.Mkdir(cacheImagePath, 600)

	tarFileName := filepath.Join(cacheImagePath, fileName)

	completeURL, err := getURLAndFileNameFromDockerImage(config)
	if err != nil {
		log.Entry().Fatalf("Protecode scan failed, exception during get url creation for get the docker image %v", err)
	}

	image, err := pkgutil.GetImage(completeURL, config.IncludeLayers, cacheImagePath)
	if err != nil {
		log.Entry().Fatalf("Protecode scan failed, exception during get docker image %v", err)
	}

	if !strings.Contains(config.ScanImage, ".tar") {
		fileName = fmt.Sprintf("%v.tar", strings.ReplaceAll(config.ScanImage, "/", "_"))
		tarFileName = filepath.Join(cachePath, fileName)

		if config.Verbose {
			log.Entry().Info("Protecode scan debug, tar the image")
		}
		tarImageData(tarFileName, image, cacheImagePath)
	}

	if len(config.FilePath) <= 0 {
		(*config).FilePath = fmt.Sprintf("./%v", filepath.Join("./", tarFileName))
		if len(config.FilePath) <= 0 {
			log.Entry().Fatalf("Protecode scan failed, there is no file path configured  : %v (filename:%v, PSPath: %v)", config.FilePath, fileName, image.FSPath)
		}
	}

	return fileName
}

func tarImageData(tarFileName string, image pkgutil.Image, cacheImagePath string) {
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

	os.RemoveAll(cacheImagePath)
}

func getURLAndFileNameFromDockerImage(config *protecodeExecuteScanOptions) (string, error) {

	completeURL := config.ScanImage

	if len(config.DockerRegistryURL) > 0 {
		if strings.HasSuffix(config.DockerRegistryURL, "/") {
			completeURL = fmt.Sprintf("remote://%v%v", config.DockerRegistryURL, config.ScanImage)
		} else {
			completeURL = fmt.Sprintf("remote://%v/%v", config.DockerRegistryURL, config.ScanImage)
		}
	}

	if len(completeURL) <= 0 {
		return completeURL, errors.New("Protecode scan failed, there is no scan image configured")
	}

	return completeURL, nil
}

func executeProtecodeScan(client protecode.Protecode, config *protecodeExecuteScanOptions, fileName string, writeReportToFile func(resp io.ReadCloser, reportFileName string) error) (map[string]int, int) {

	var parsedResult map[string]int = make(map[string]int)
	//load existing product by filename
	if config.Verbose {
		log.Entry().Info("Protecode scan debug, load existing product")
	}
	productID := client.LoadExistingProduct(config.ProtecodeGroup, config.FilePath, config.ReuseExisting)

	// check if no existing is found or reuse existing is false
	productID = uploadScanOrDeclareFetch(*config, productID, client, fileName)
	if productID <= 0 {
		log.Entry().Fatalf("Protecode scan failed, the product id is not valid (product id %v <= zero)", productID)
	}
	//pollForResult
	if config.Verbose {
		log.Entry().Info("Protecode scan debug, poll for scan result")
	}
	result := client.PollForResult(productID, config.Verbose)

	jsonData, _ := json.Marshal(result)
	ioutil.WriteFile("Vulns.json", jsonData, 0644)

	//check if result is ok else notify
	if len(result.Result.Status) > 0 && result.Result.Status == "F" {
		log.Entry().Fatal("Protecode scan failed, please check the log and protecode backend for more details.")
	}
	//loadReport
	if config.Verbose {
		log.Entry().Info("Protecode scan debug, load report")
	}
	resp := client.LoadReport(config.ReportFileName, productID)

	//save report to filesystem
	err := writeReportToFile(*resp, config.ReportFileName)
	if err != nil {
		return parsedResult, productID
	}
	//clean scan from server
	if config.Verbose {
		log.Entry().Info("Protecode scan debug, delete scan")
	}
	client.DeleteScan(config.CleanupMode, productID)

	//count vulnerabilities
	if config.Verbose {
		log.Entry().Info("Protecode scan debug, parse result")
	}
	parsedResult = client.ParseResultForInflux(result.Result, config.ProtecodeExcludeCVEs)

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

	writeToFile("report.json", jsonData, 0644)
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

	protecodeOptions := protecode.ProtecodeOptions{
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

	// check if no existing is found or reuse existing is false
	if productID <= 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			if config.Verbose {
				log.Entry().Info("Protecode scan debug, declare fetch url")
			}
			resultData := client.DeclareFetchUrl(config.CleanupMode, config.ProtecodeGroup, config.FetchURL)
			productID = resultData.ProductId

		} else {
			if config.Verbose {
				log.Entry().Infof("Protecode scan debug, upload file path: %v", config.FilePath)
			}
			if len(config.FilePath) <= 0 {
				log.Entry().Fatalf("Protecode scan failed, there is no file path configured for upload : %v", config.FilePath)
			}
			resultData := client.UploadScanFile(config.CleanupMode, config.ProtecodeGroup, config.FilePath, filaName)
			productID = resultData.Result.ProductId
		}
	}

	return productID
}

func writeReportToFile(resp io.ReadCloser, reportFileName string) error {
	f, err := os.Create(reportFileName)
	if err == nil {
		defer f.Close()
		_, err = io.Copy(f, resp)
	}

	return err
}
