package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/docker"
	piperDocker "github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/toolrecord"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

const (
	webReportPath    = "%s/#/product/%v/"
	scanResultFile   = "protecodescan_vulns.json"
	stepResultFile   = "protecodeExecuteScan.json"
	dockerConfigFile = ".pipeline/docker/config.json"
)

type protecodeUtils interface {
	piperutils.FileUtils
	piperDocker.Download
}

type protecodeUtilsBundle struct {
	*piperutils.Files
	*piperDocker.Client
}

func protecodeExecuteScan(config protecodeExecuteScanOptions, telemetryData *telemetry.CustomData, influx *protecodeExecuteScanInflux) {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	//create client for sending api request
	log.Entry().Debug("Create protecode client")
	client := createProtecodeClient(&config)

	dClientOptions := piperDocker.ClientOptions{ImageName: config.ScanImage, RegistryURL: config.DockerRegistryURL, LocalPath: config.FilePath, ImageFormat: "legacy"}
	dClient := &piperDocker.Client{}
	dClient.SetOptions(dClientOptions)

	utils := protecodeUtilsBundle{
		Client: dClient,
		Files:  &piperutils.Files{},
	}

	influx.step_data.fields.protecode = false
	if err := runProtecodeScan(&config, influx, client, utils, "./cache"); err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute protecode scan.")
	}
	influx.step_data.fields.protecode = true
}

func runProtecodeScan(config *protecodeExecuteScanOptions, influx *protecodeExecuteScanInflux, client protecode.Protecode, utils protecodeUtils, cachePath string) error {
	// make sure cache exists
	if err := utils.MkdirAll(cachePath, 0755); err != nil {
		return err
	}

	if err := correctDockerConfigEnvVar(config, utils); err != nil {
		return err
	}

	var fileName, filePath string
	var err error

	if len(config.FetchURL) == 0 && len(config.FilePath) == 0 {
		log.Entry().Debugf("Get docker image: %v, %v, %v", config.ScanImage, config.DockerRegistryURL, config.FilePath)
		fileName, filePath, err = getDockerImage(utils, config, cachePath)
		if err != nil {
			return errors.Wrap(err, "failed to get Docker image")
		}
		if len(config.FilePath) <= 0 {
			(*config).FilePath = filePath
			log.Entry().Debugf("Filepath for upload image: %v", config.FilePath)
		}
	} else if len(config.FilePath) > 0 {
		parts := strings.Split(config.FilePath, "/")
		pathFragment := strings.Join(parts[:len(parts)-1], "/")
		if len(pathFragment) > 0 {
			(*config).FilePath = pathFragment
		} else {
			(*config).FilePath = "./"
		}
		fileName = parts[len(parts)-1]

	} else if len(config.FetchURL) > 0 {
		// Get filename from a fetch URL
		fileName = filepath.Base(config.FetchURL)
		log.Entry().Debugf("[DEBUG] ===> Filename from fetch URL: %v", fileName)
	}

	log.Entry().Debug("Execute protecode scan")
	if err := executeProtecodeScan(influx, client, config, fileName, utils); err != nil {
		return err
	}

	defer func() { _ = utils.FileRemove(config.FilePath) }()

	if err := utils.RemoveAll(cachePath); err != nil {
		log.Entry().Warnf("Error during cleanup folder %v", err)
	}

	return nil
}

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

func getDockerImage(utils protecodeUtils, config *protecodeExecuteScanOptions, cachePath string) (string, string, error) {
	m := regexp.MustCompile(`[\s@:/]`)

	tarFileName := fmt.Sprintf("%s.tar", m.ReplaceAllString(config.ScanImage, "-"))
	tarFilePath, err := filepath.Abs(filepath.Join(cachePath, tarFileName))

	if err != nil {
		return "", "", err
	}

	if _, err = utils.DownloadImage(config.ScanImage, tarFilePath); err != nil {
		return "", "", errors.Wrap(err, "failed to download docker image")
	}

	return filepath.Base(tarFilePath), filepath.Dir(tarFilePath), nil
}

func executeProtecodeScan(influx *protecodeExecuteScanInflux, client protecode.Protecode, config *protecodeExecuteScanOptions, fileName string, utils protecodeUtils) error {
	reportPath := "./"

	log.Entry().Debugf("[DEBUG] ===> Load existing product Group:%v, VerifyOnly:%v, Filename:%v, replaceProductId:%v", config.Group, config.VerifyOnly, fileName, config.ReplaceProductID)

	var productID int

	// If replaceProductId is not provided then switch to automatic existing product detection
	if config.ReplaceProductID > 0 {

		log.Entry().Infof("replaceProductID has been provided (%v) and checking ...", config.ReplaceProductID)

		// Validate provided product id, if not valid id then throw an error
		if client.VerifyProductID(config.ReplaceProductID) {
			log.Entry().Infof("replaceProductID has been checked and it's valid")
			productID = config.ReplaceProductID
		} else {
			log.Entry().Debugf("[DEBUG] ===> ReplaceProductID doesn't exist")
			return fmt.Errorf("ERROR -> the product id is not valid '%d'", config.ReplaceProductID)
		}

	} else {
		// Get existing product id by filename
		log.Entry().Infof("replaceProductID is not provided and automatic search starts from group: %v ... ", config.Group)
		productID = client.LoadExistingProduct(config.Group, fileName)

		if productID > 0 {
			log.Entry().Infof("Automatic search completed and found following product id: %v", productID)
		} else {
			log.Entry().Infof("Automatic search completed but not found any similar product scan, now starts new scan creation")
		}
	}

	// check if no existing is found
	productID = uploadScanOrDeclareFetch(utils, *config, productID, client, fileName)

	if productID <= 0 {
		return fmt.Errorf("the product id is not valid '%d'", productID)
	}

	//pollForResult
	log.Entry().Debugf("Poll for scan result %v", productID)
	result := client.PollForResult(productID, config.TimeoutMinutes)
	// write results to file
	jsonData, _ := json.Marshal(result)
	if err := utils.FileWrite(filepath.Join(reportPath, scanResultFile), jsonData, 0644); err != nil {
		log.Entry().Warningf("failed to write result file: %v", err)
	}

	//check if result is ok else notify
	if protecode.HasFailed(result) {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("protecode scan failed: %v/products/%v", config.ServerURL, productID)
	}

	//loadReport
	log.Entry().Debugf("Load report %v for %v", config.ReportFileName, productID)
	resp := client.LoadReport(config.ReportFileName, productID)

	buf, err := io.ReadAll(*resp)

	if err != nil {
		return fmt.Errorf("unable to process protecode report %v", err)
	}

	if err = utils.FileWrite(config.ReportFileName, buf, 0644); err != nil {
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
		}, reportPath, stepResultFile, parsedResult, utils); err != nil {
		log.Entry().Warningf("failed to write report: %v", err)
	}

	log.Entry().Debug("Write influx data")
	setInfluxData(influx, parsedResult)

	// write reports JSON
	reports := []piperutils.Path{
		{Target: config.ReportFileName, Mandatory: true},
		{Target: stepResultFile, Mandatory: true},
		{Target: scanResultFile, Mandatory: true},
	}
	// write links JSON
	webuiURL := fmt.Sprintf(webReportPath, config.ServerURL, productID)
	links := []piperutils.Path{
		{Name: "Protecode WebUI", Target: webuiURL},
		{Name: "Protecode Report", Target: path.Join("artifact", config.ReportFileName), Scope: "job"},
	}

	// write custom report
	scanReport := protecode.CreateCustomReport(fileName, productID, parsedResult, vulns)
	paths, err := protecode.WriteCustomReports(scanReport, fileName, fmt.Sprint(productID), utils)
	if err != nil {
		// do not fail - consider failing later on
		log.Entry().Warning("failed to create custom HTML/MarkDown file ...", err)
	} else {
		reports = append(reports, paths...)
	}

	// create toolrecord file
	toolRecordFileName, err := createToolRecordProtecode(utils, "./", config, productID, webuiURL)
	if err != nil {
		// do not fail until the framework is well established
		log.Entry().Warning("TR_PROTECODE: Failed to create toolrecord file ...", err)
	} else {
		reports = append(reports, piperutils.Path{Target: toolRecordFileName})
	}

	piperutils.PersistReportsAndLinks("protecodeExecuteScan", "", utils, reports, links)

	if config.FailOnSevereVulnerabilities && protecode.HasSevereVulnerabilities(result.Result, config.ExcludeCVEs) {
		log.SetErrorCategory(log.ErrorCompliance)
		return fmt.Errorf("the product is not compliant")
	} else if protecode.HasSevereVulnerabilities(result.Result, config.ExcludeCVEs) {
		log.Entry().Infof("policy violation(s) found - step will only create data but not fail due to setting failOnSevereVulnerabilities: false")
	}
	return nil
}

func setInfluxData(influx *protecodeExecuteScanInflux, result map[string]int) {
	influx.protecode_data.fields.historical_vulnerabilities = result["historical_vulnerabilities"]
	influx.protecode_data.fields.triaged_vulnerabilities = result["triaged_vulnerabilities"]
	influx.protecode_data.fields.excluded_vulnerabilities = result["excluded_vulnerabilities"]
	influx.protecode_data.fields.minor_vulnerabilities = result["minor_vulnerabilities"]
	influx.protecode_data.fields.major_vulnerabilities = result["major_vulnerabilities"]
	influx.protecode_data.fields.vulnerabilities = result["vulnerabilities"]
}

func createProtecodeClient(config *protecodeExecuteScanOptions) protecode.Protecode {
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
		ServerURL:  config.ServerURL,
		Logger:     log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode"),
		Duration:   duration,
		Username:   config.Username,
		Password:   config.Password,
		UserAPIKey: config.UserAPIKey,
	}

	pc.SetOptions(protecodeOptions)

	return pc
}

func uploadScanOrDeclareFetch(utils protecodeUtils, config protecodeExecuteScanOptions, productID int, client protecode.Protecode, fileName string) int {

	// check if product doesn't exist then create a new one.
	if productID <= 0 {
		log.Entry().Infof("New product creation started ... ")
		productID = uploadFile(utils, config, productID, client, fileName, false)

		log.Entry().Infof("New product has been successfully created: %v", productID)
		return productID

		// In case product already exists and "VerifyOnly (reuseExisting)" is false then we replace binary without creating a new product.
	} else if (productID > 0) && !config.VerifyOnly {
		log.Entry().Infof("Product already exists and 'VerifyOnly (reuseExisting)' is false then product (%v) binary and scan result will be replaced without creating a new product.", productID)
		productID = uploadFile(utils, config, productID, client, fileName, true)

		return productID

		// If product already exists and "reuseExisting" option is enabled then return the latest similar scan result.
	} else {
		log.Entry().Infof("VerifyOnly (reuseExisting) option is enabled and returned productID: %v", productID)
		return productID
	}
}

func uploadFile(utils protecodeUtils, config protecodeExecuteScanOptions, productID int, client protecode.Protecode, fileName string, replaceBinary bool) int {

	// get calculated version for Version field
	version := getProcessedVersion(&config)

	if len(config.FetchURL) > 0 {
		log.Entry().Debugf("Declare fetch url %v", config.FetchURL)
		resultData := client.DeclareFetchURL(config.CleanupMode, config.Group, config.CustomDataJSONMap, config.FetchURL, version, productID, replaceBinary)
		productID = resultData.Result.ProductID
	} else {
		log.Entry().Debugf("Upload file path: %v", config.FilePath)
		if len(config.FilePath) <= 0 {
			log.Entry().Fatalf("There is no file path configured for upload : %v", config.FilePath)
		}
		pathToFile := filepath.Join(config.FilePath, fileName)
		if exists, err := utils.FileExists(pathToFile); err != nil && !exists {
			log.Entry().Fatalf("There is no file for upload: %v", pathToFile)
		}

		combinedFileName := fileName
		if len(config.PullRequestName) > 0 {
			combinedFileName = fmt.Sprintf("%v_%v", config.PullRequestName, fileName)
		}

		resultData := client.UploadScanFile(config.CleanupMode, config.Group, config.CustomDataJSONMap, pathToFile, combinedFileName, version, productID, replaceBinary)
		productID = resultData.Result.ProductID
	}
	return productID
}

func correctDockerConfigEnvVar(config *protecodeExecuteScanOptions, utils protecodeUtils) error {
	var err error
	path := config.DockerConfigJSON

	if len(config.DockerConfigJSON) > 0 && len(config.DockerRegistryURL) > 0 && len(config.ContainerRegistryPassword) > 0 && len(config.ContainerRegistryUser) > 0 {
		path, err = docker.CreateDockerConfigJSON(config.DockerRegistryURL, config.ContainerRegistryUser, config.ContainerRegistryPassword, dockerConfigFile, config.DockerConfigJSON, utils)
	}

	if err != nil {
		return errors.Wrap(err, "failed to create / update docker config json file")
	}

	if len(path) > 0 {
		log.Entry().Infof("Docker credentials configuration: %v", path)
		path, _ := filepath.Abs(path)
		// use parent directory
		path = filepath.Dir(path)
		os.Setenv("DOCKER_CONFIG", path)
	} else {
		log.Entry().Info("Docker credentials configuration: NONE")
	}
	return nil
}

// Calculate version based on versioning model and artifact version or return custom scan version provided by user
func getProcessedVersion(config *protecodeExecuteScanOptions) string {
	processedVersion := config.CustomScanVersion
	if len(processedVersion) > 0 {
		log.Entry().Infof("Using custom version: %v", processedVersion)
	} else {
		if len(config.VersioningModel) > 0 {
			processedVersion = versioning.ApplyVersioningModel(config.VersioningModel, config.Version)
		} else {
			// By default 'major' if <config.VersioningModel> not provided
			processedVersion = versioning.ApplyVersioningModel("major", config.Version)
		}
	}
	return processedVersion
}

// create toolrecord file for protecode
// todo: check if group and product names can be retrieved
func createToolRecordProtecode(utils protecodeUtils, workspace string, config *protecodeExecuteScanOptions, productID int, webuiURL string) (string, error) {
	record := toolrecord.New(utils, workspace, "protecode", config.ServerURL)
	groupURL := config.ServerURL + "/#/groups/" + config.Group
	err := record.AddKeyData("group",
		config.Group,
		config.Group, // todo figure out display name
		groupURL)
	if err != nil {
		return "", err
	}
	err = record.AddKeyData("product",
		strconv.Itoa(productID),
		strconv.Itoa(productID), // todo figure out display name
		webuiURL)
	if err != nil {
		return "", err
	}
	err = record.Persist()
	if err != nil {
		return "", err
	}
	return record.GetFileName(), nil
}
