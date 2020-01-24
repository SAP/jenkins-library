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
	ReportFileName            string `json:"reportFileName,omitempty"`
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
	client := createClient(config)

	fileName, err := getDockerImage(config, client)
	if err != nil {
		log.Entry().Fatalf("Exception during getting the image %v", err)
	}
	parsedResult, productID, err := executeProtecodeScan(client, config, fileName, writeReportToFile)
	if err != nil {
		log.Entry().Fatalf("Exception during the execute of the scan %v", err)
	}

	setInfluxData(influx, parsedResult)

	writeDataToJSONFile(config, parsedResult, productID)

	log.Entry().Debugf("Cleanup tar archive")
	err = os.Remove(config.FilePath)
	if err != nil {
		log.Entry().WithError(err).Warnf("Failed to delete tar source code")
	}

	return nil
}

func getDockerImage(config *protecodeExecuteScanOptions, client protecode.Protecode) (string, error) {

	cacheImagePath := "./cache/protecodeImage"
	os.Mkdir(cacheImagePath, 600)
	completeURL, err := getURLAndFileNameFromDockerImage(config)
	if err != nil {
		log.Entry().Fatalf("Exception during get url creation for get the docker image %v", err)
	}

	image, err := pkgutil.GetImage(completeURL, config.IncludeLayers, cacheImagePath)
	if err != nil {
		log.Entry().Fatalf("Exception during get docker image %v", err)
	}
	fileName := fmt.Sprintf("%v.tar", strings.ReplaceAll(config.ScanImage, "/", "_"))

	//tar folder
	cacheTarPath := "./cache"
	tarFileName := filepath.Join(cacheTarPath, fileName)
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to create tar of docker image")
	}
	if err := os.Chmod(tarFileName, 0644); err != nil {
		log.Entry().WithError(err)
	}
	defer tarFile.Close()

	reference, err := name.ParseReference(image.Digest.String(), name.WeakValidation)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to parse reference of docker image")
	}
	err = tarball.Write(reference, image.Image, tarFile)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to create tar archive of docker image via tarball")
	}

	os.RemoveAll(cacheImagePath)

	if len(config.FilePath) <= 0 {
		(*config).FilePath = fmt.Sprintf("./%v", filepath.Join("./", tarFileName))
		if len(config.FilePath) <= 0 {
			log.Entry().Fatalf("Protecode scan failed, there is no file path configured  : %v (filename:%v, PSPath: %v)", config.FilePath, fileName, image.FSPath)
		}
	}

	return fileName, nil
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

func executeProtecodeScan(client protecode.Protecode, config *protecodeExecuteScanOptions, fileName string, writeReportToFile func(resp io.ReadCloser, reportFileName string) error) (map[string]int, int, error) {

	var parsedResult map[string]int = make(map[string]int)
	//load existing product by filename
	productID := client.LoadExistingProduct(config.ProtecodeGroup, config.FilePath, config.ReuseExisting)

	// check if no existing is found or reuse existing is false
	productID, err := uploadScanOrDeclareFetch(*config, productID, client, fileName)
	if err != nil {
		return parsedResult, productID, err
	}
	if productID <= 0 {
		return parsedResult, productID, errors.New(fmt.Sprintf("Protecode scan failed, the product id is not valid (product id %v <= zero)", productID))
	}
	//pollForResult
	result := client.PollForResult(productID, config.Verbose)

	jsonData, _ := json.Marshal(result)
	ioutil.WriteFile("Vulns.json", jsonData, 0644)

	//check if result is ok else notify
	if len(result.Result.Status) > 0 && result.Result.Status == "F" {
		log.Entry().Fatal("Protecode scan failed, please check the log and protecode backend for more details.")
		return parsedResult, productID, errors.New("Protecode scan failed, please check the log and protecode backend for more details")
	}
	//loadReport
	resp := client.LoadReport(config.ReportFileName, productID)

	//save report to filesystem
	err = writeReportToFile(*resp, config.ReportFileName)
	if err != nil {
		return parsedResult, productID, err
	}
	//clean scan from server
	client.DeleteScan(config.CleanupMode, productID)

	//count vulnerabilities
	parsedResult = client.ParseResultForInflux(result.Result, config.ProtecodeExcludeCVEs)

	log.Entry().Infof("Protecode scan result: %v", parsedResult)

	return parsedResult, productID, nil
}

func setInfluxData(influx *protecodeExecuteScanInflux, result map[string]int) {

	influx.protecodeData.fields.historicalVulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
	influx.protecodeData.fields.triagedVulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	influx.protecodeData.fields.excludedVulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	influx.protecodeData.fields.minorVulnerabilities = fmt.Sprintf("%v", result["minor_vulnerabilities"])
	influx.protecodeData.fields.majorVulnerabilities = fmt.Sprintf("%v", result["major_vulnerabilities"])
	influx.protecodeData.fields.vulnerabilities = fmt.Sprintf("%v", result["vulnerabilities"])
}

func writeDataToJSONFile(config *protecodeExecuteScanOptions, result map[string]int, productID int) {

	protecodeData := protecodeData{}
	protecodeData.ProtecodeServerURL = config.ProtecodeServerURL
	protecodeData.ReportFileName = config.ReportFileName
	protecodeData.ProductID = fmt.Sprintf("%v", productID)
	protecodeData.Count = fmt.Sprintf("%v", result["count"])
	protecodeData.Cvss2GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss2GreaterOrEqualSeven"])
	protecodeData.Cvss3GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss3GreaterOrEqualSeven"])
	protecodeData.ExcludedVulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	protecodeData.TriagedVulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	protecodeData.HistoricalVulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])

	jsonData, _ := json.Marshal(protecodeData)

	ioutil.WriteFile("ProtecodeData.json", jsonData, 0644)
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

func uploadScanOrDeclareFetch(config protecodeExecuteScanOptions, productID int, client protecode.Protecode, filaName string) (int, error) {

	// check if no existing is found or reuse existing is false
	if productID <= 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			fmt.Printf("triggering Protecode scan - url: %v, group: %v", config.FetchURL, config.ProtecodeGroup)
			resultData := client.DeclareFetchUrl(config.CleanupMode, config.ProtecodeGroup, config.FetchURL)
			log.Entry().Infof("Protecode scan declare fetch url result: %v", resultData)
			productID = resultData.ProductId

		} else {
			if len(config.FilePath) <= 0 {
				return -1, errors.New(fmt.Sprintf("Protecode scan failed, there is no file path configured for upload : %v", config.FilePath))
			}
			fmt.Printf("triggering Protecode scan - file: %v, group: %v", config.FilePath, config.ProtecodeGroup)
			resultData := client.UploadScanFile(config.CleanupMode, config.ProtecodeGroup, config.FilePath, filaName)
			log.Entry().Infof("Protecode scan upload result: %v", resultData)
			productID = resultData.Result.ProductId
		}
	}

	return productID, nil
}

func writeReportToFile(resp io.ReadCloser, reportFileName string) error {
	f, err := os.Create(reportFileName)
	if err == nil {
		defer f.Close()
		_, err = io.Copy(f, resp)
	}

	return err
}
