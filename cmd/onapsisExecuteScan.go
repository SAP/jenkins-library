package cmd

import (
	"archive/zip"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"
)

type onapsisExecuteScanUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	Open(name string) (io.ReadWriteCloser, error)
	Getwd() (string, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The onapsisExecuteScanUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type onapsisExecuteScanUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to onapsisExecuteScanUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// onapsisExecuteScanUtilsBundle and forward to the implementation of the dependency.
}

func newOnapsisExecuteScanUtils() onapsisExecuteScanUtils {
	utils := onapsisExecuteScanUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

var includePatterns = []string{
	// TODO: Add more include patterns as needed (e.g., for ABAP scans)
	"**/*.js",
	"**/*.json",
}

var excludePatterns = []string{
	"**/.git/**",         // Exclude .git directory
	"**/.pipeline/**",    // Exclude .pipeline directory
	"**/node_modules/**", // Exclude node_modules directory
	"**/.gitignore",      // Exclude .gitignore file
	"**/*.log",           // Exclude all log files
	"workspace.zip",      // Exclude the zip file itself
}

func zipProject(folderPath string, outputPath string) error {
	log.Entry().Infof("Starting to zip folder: %s", folderPath)

	// Create the output file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	log.Entry().Infof("Created zip file: %s", outputPath)

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Track file count
	fileCount := 0

	// Walk through all the files in the folder
	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Entry().Errorf("Error accessing path %s: %v", path, err)
			return err
		}

		// Check if the file matches any of the exclude patterns
		for _, pattern := range excludePatterns {
			matched, _ := doublestar.Match(pattern, path)
			if matched {
				log.Entry().Infof("Excluding: %s (matches pattern: %s)", path, pattern)
				if info.IsDir() {
					return filepath.SkipDir // Skip the entire directory
				}
				return nil // Skip the file
			}
		}

		// Check if the file matches any of the include patterns
		included := false
		for _, pattern := range includePatterns {
			matched, _ := doublestar.Match(pattern, path)
			if matched {
				included = true
				break
			}
		}
		if !included {
			log.Entry().Infof("Skipping: %s (does not match include patterns)", path)
			return nil
		}

		// Log each file being processed
		log.Entry().Infof("Zipping file or directory: %s", path)

		// Create a header based on the file info
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Entry().Errorf("Failed to create zip header for file: %s", path)
			return err
		}

		// Ensure the correct relative file path in the zip
		header.Name, err = filepath.Rel(filepath.Dir(folderPath), path)
		if err != nil {
			log.Entry().Errorf("Failed to create relative path for file: %s", path)
			return err
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		// Create the writer for this file
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			log.Entry().Errorf("Failed to write header for file: %s", path)
			return err
		}

		// If it's a file, copy the content into the zip
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				log.Entry().Errorf("Failed to open file: %s", path)
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				log.Entry().Errorf("Failed to copy file content to zip for file: %s", path)
				return err
			}
		}

		fileCount++
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to zip folder: %w", err)
	}

	log.Entry().Infof("Successfully zipped %d files", fileCount)

	return nil
}

// AuthResponse for Onapsis response
// AuthResponse matches the Onapsis API response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// createSecureHTTPClient initializes and returns an HTTP client with a custom CA certificate
func createSecureHTTPClient(certPath string) (*http.Client, error) {
	// Read the CA certificate
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	// Create a certificate pool and append the internal CA
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// Create and return the secure HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
		Timeout: 10 * time.Second,
	}

	return client, nil
}

func refreshJwtToken(refreshToken, scanServiceUrl string) (string, error) {
	refreshTokenURL := scanServiceUrl + "/cca/v1.2/auth_token"

	// Create HTTP request
	req, err := http.NewRequest("GET", refreshTokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set and format the expected Cookie header
	req.Header.Set("Cookie", fmt.Sprintf("refresh_token=%s", refreshToken))

	certPath := "/home/ca.pem"

	client, err = createSecureHTTPClient(certPath)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Handle response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to refresh token, status: %d, response: %s", resp.StatusCode, string(body))
	}

	// Extract new JWT from the response (assuming JSON response)
	var refreshJwtToken AuthResponse
	if err := json.Unmarshal(body, &refreshJwtToken); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return refreshJwtToken.AccessToken, nil

}

// getJWTFromService fetches a JWT using Basic Auth
func getJWTFromService(username, password, scanServiceUrl string) (*AuthResponse, error) {

	url := scanServiceUrl + "/cca/v1.2/auth_token"
	certPath := "/home/ca.pem"

	fmt.Println("This is the scan servcie url: ", scanServiceUrl)

	// Create the request with GET method
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set Basic Auth (Postman Authorization tab)
	req.SetBasicAuth(username, password)

	client, err = createSecureHTTPClient(certPath)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to JWT service failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to obtain JWT, status code: %d, response: %s", resp.StatusCode, body)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &authResp, nil
}

func onapsisExecuteScan(config onapsisExecuteScanOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newOnapsisExecuteScanUtils()

	//display token and refresh token for tests purposes should be deleted after the merge
	token, tokenErr := getJWTFromService(config.OnapsisUsername, config.OnapsisPassword, config.ScanServiceURL)
	if tokenErr != nil {
		log.Entry().WithError(tokenErr).Fatal("Error obtaining JWT")
	}

	fmt.Println("Received JWT:", token.AccessToken)
	fmt.Println("Received Refresh Token:", token.RefreshToken)

	newJwt, refreshTokenErr := refreshJwtToken(token.RefreshToken, config.ScanServiceURL)

	if refreshTokenErr != nil {
		log.Entry().WithError(refreshTokenErr).Fatal("Error obtaining refreshed JWT")
	}

	fmt.Println("Received refreshed JWT:", newJwt)

	if config.DebugMode {
		log.SetVerbose(true)
	}

	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runOnapsisExecuteScan(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runOnapsisExecuteScan(config *onapsisExecuteScanOptions, telemetryData *telemetry.CustomData, utils onapsisExecuteScanUtils) error {
	// Create a new ScanServer
	log.Entry().Info("Creating scan server...")
	server, err := NewScanServer(&piperHttp.Client{}, config.ScanServiceURL, config.AccessToken)
	if err != nil {
		return errors.Wrap(err, "failed to create scan server")
	}

	// Call the ScanProject method
	log.Entry().Info("Scanning project...")
	response, err := server.ScanProject(config, telemetryData, utils, config.AppType)
	if err != nil {
		return errors.Wrap(err, "Failed to scan project")
	}

	// Monitor Job Status
	jobID := response.Result.JobID
	log.Entry().Infof("Monitoring job %s status...", jobID)
	err = server.MonitorJobStatus(jobID)
	if err != nil {
		return errors.Wrap(err, "Failed to scan project")
	}

	// Get Job Reports
	log.Entry().Info("Getting job reports...")
	err = server.GetJobReports(jobID, "onapsis_scan_report.zip")
	if err != nil {
		return errors.Wrap(err, "Failed to get job reports")
	}

	// Get Job Result Metrics
	log.Entry().Info("Getting job result metrics...")
	metrics, err := server.GetJobResultMetrics(jobID)
	if err != nil {
		return errors.Wrap(err, "Failed to get job result metrics")
	}

	// Analyze metrics
	loc, numMandatory, numOptional := extractMetrics(metrics)
	// TODO: Change logging to print lines of code scanned in what amount of time
	log.Entry().Infof("Job Metrics - Lines of Code Scanned: %s, Mandatory Findings: %s, Optional Findings: %s", loc, numMandatory, numOptional)

	if config.FailOnMandatoryFinding && numMandatory != "0" {
		return errors.Errorf("Scan failed with %s mandatory findings", numMandatory)
	} else if config.FailOnOptionalFinding && numOptional != "0" {
		return errors.Errorf("Scan failed with %s optional findings", numOptional)
	}

	return nil
}

type ScanServer struct {
	serverUrl string
	client    piperHttp.Uploader
}

func NewScanServer(client piperHttp.Uploader, serverUrl string, token string) (*ScanServer, error) {
	server := &ScanServer{serverUrl: serverUrl, client: client}

	log.Entry().Debugf("Token: %s", token)

	// Set authorization token for client
	options := piperHttp.ClientOptions{
		Token:                     "Bearer " + token,
		MaxRequestDuration:        60 * time.Second, // DEBUG
		TransportSkipVerification: true,             //DEBUG
		DoLogRequestBodyOnDebug:   true,
		DoLogResponseBodyOnDebug:  true,
	}
	server.client.SetOptions(options)

	return server, nil
}

func (srv *ScanServer) ScanProject(config *onapsisExecuteScanOptions, telemetryData *telemetry.CustomData, utils onapsisExecuteScanUtils, language string) (ScanProjectResponse, error) {
	// Get workspace path
	log.Entry().Info("Getting workspace path...") // DEBUG
	workspace, err := utils.Getwd()
	if err != nil {
		return ScanProjectResponse{}, errors.Wrap(err, "failed to get workspace path")
	}

	// Zip workspace files
	log.Entry().Info("Zipping workspace files...") // DEBUG
	zipFileName := "workspace.zip"
	zipFilePath := filepath.Join(workspace, zipFileName)
	err = zipProject(workspace, zipFilePath)
	if err != nil {
		return ScanProjectResponse{}, errors.Wrap(err, "failed to zip workspace files")
	}

	// Get zip file content
	log.Entry().Info("Getting zip file content...") // DEBUG
	fileHandle, err := utils.Open(zipFilePath)
	if err != nil {
		return ScanProjectResponse{}, errors.Wrapf(err, "unable to locate file %v", zipFilePath)
	}
	defer fileHandle.Close()

	// Construct ScanConfig form field
	log.Entry().Info("Constructing ScanConfig form field...") // DEBUG
	scanConfig := fmt.Sprintf(`{
		"engine_type": "FILE",
		"scan_information": {
			"name": "scenario",
			"description": "a scan with extracted source"
		},
		"asset": {
			"file_format": "ZIP",
			"recursive": "true",
			"language": "%s"
		},
		"configuration": {},
		"scan_scope": {}
	}`, language)

	formFields := map[string]string{
		"ScanConfig": scanConfig,
	}

	// Create request data
	log.Entry().Info("Creating request data...") // DEBUG
	requestData := piperHttp.UploadRequestData{
		Method:        "POST",
		URL:           srv.serverUrl + "/cca/v1.0/scan/file",
		File:          zipFileName,
		FileFieldName: "FileUploadContent",
		FileContent:   fileHandle,
		FormFields:    formFields,
		UploadType:    "form",
	}

	// Send request
	log.Entry().Info("Sending request...") // DEBUG
	response, err := srv.client.Upload(requestData)
	if err != nil {
		return ScanProjectResponse{}, errors.Wrap(err, "Failed to start scan")
	}

	// Handle response
	var responseData ScanProjectResponse
	err = handleResponse(response, &responseData)
	if err != nil {
		return responseData, errors.Wrap(err, "Failed to parse response")
	}

	return responseData, nil
}

func (srv *ScanServer) GetScanJobStatus(jobID string) (GetScanJobStatusResponse, error) {
	// Send request
	response, err := srv.client.SendRequest("GET", srv.serverUrl+"/cca/v1.2/job/"+jobID, nil, nil, nil)
	if err != nil {
		return GetScanJobStatusResponse{}, errors.Wrap(err, "failed to send request")
	}

	var responseData GetScanJobStatusResponse
	err = handleResponse(response, &responseData)
	if err != nil {
		return responseData, errors.Wrap(err, "Failed to parse response")
	}

	return responseData, nil
}

func (srv *ScanServer) MonitorJobStatus(jobID string) error {
	// Polling interval
	interval := time.Second * 10 // Check every 10 seconds
	for {
		// Get the job status
		response, err := srv.GetScanJobStatus(jobID)
		if err != nil {
			return errors.Wrap(err, "Failed to get scan job status")
		}

		// Log job progress
		log.Entry().Infof("Job %s progress: %d%%", jobID, response.Result.Progress)

		// Check if the job is complete
		if response.Result.Status == "SUCCESS" {
			return nil
		} else if response.Result.Status == "FAILURE" {
			return errors.Errorf("Job %s failed with status: %s", jobID, response.Result.Status)
		}

		// Wait before checking again
		time.Sleep(interval)
	}
}

func (srv *ScanServer) GetJobReports(jobID string, reportArchiveName string) error {
	response, err := srv.client.SendRequest("GET", srv.serverUrl+"/cca/v1.2/job/"+jobID+"/result?fileType=all", nil, nil, nil)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve job report")
	}

	// Create the destination zip file
	outFile, err := os.Create(reportArchiveName)
	if err != nil {
		return errors.Wrap(err, "Failed to create report archive")
	}
	defer outFile.Close()

	// Copy the response body to the file
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		return errors.Wrap(err, "Failed to write report archive")
	}

	return nil
}

func (srv *ScanServer) GetJobResultMetrics(jobID string) (GetJobResultMetricsResponse, error) {
	// Send request
	response, err := srv.client.SendRequest("GET", srv.serverUrl+"/cca/v1.2/job/"+jobID+"/result?type=metrics", nil, nil, nil)
	if err != nil {
		return GetJobResultMetricsResponse{}, errors.Wrap(err, "failed to send request")
	}

	var responseData GetJobResultMetricsResponse
	err = handleResponse(response, &responseData)
	if err != nil {
		return responseData, errors.Wrap(err, "Failed to parse response")
	}

	return responseData, nil
}

func extractMetrics(response GetJobResultMetricsResponse) (loc, numMandatory, numOptional string) {
	for _, metric := range response.Result.Metrics {
		switch metric.Name {
		case "LOC":
			loc = metric.Value
		case "num_mandatory":
			numMandatory = metric.Value
		case "num_optional":
			numOptional = metric.Value
		}
	}

	return loc, numMandatory, numOptional
}

type ScanProjectResponse struct {
	Success bool `json:"success"`
	Result  struct {
		JobID      string    `json:"job_id"`      // present only on success
		ResultCode int       `json:"result_code"` // present only on failure
		Timestamp  string    `json:"timestamp"`   // present only on success
		Messages   []Message `json:"messages"`
	} `json:"result"`
}

type GetScanJobStatusResponse struct {
	Success bool `json:"success"`
	Result  struct {
		JobID         string    `json:"job_id"`
		ReqRecvTime   string    `json:"req_recv_time"`
		ScanStartTime string    `json:"scan_start_time"`
		ScanEndTime   string    `json:"scan_end_time"`
		EngineType    string    `json:"engine_type"`
		Status        string    `json:"status"`
		Progress      int       `json:"progress"`
		Messages      []Message `json:"messages"`
	} `json:"result"`
}

type GetJobResultMetricsResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Metrics []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"metrics"`
	} `json:"result"`
}

type Message struct {
	Sequence  int    `json:"sequence"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	MessageID string `json:"message_id"`
	Param1    string `json:"param1"`
	Param2    string `json:"param2"`
	Param3    string `json:"param3"`
	Param4    string `json:"param4"`
}

func handleResponse(response *http.Response, responseData interface{}) error {
	err := piperHttp.ParseHTTPResponseBodyJSON(response, &responseData)
	if err != nil {
		return errors.Wrap(err, "Failed to parse file")
	}

	// Define a helper function to check success and handle error messages
	checkResponse := func(success bool, messages interface{}, resultCode int) error {
		if success {
			return nil
		}
		messageJSON, err := json.MarshalIndent(messages, "", "  ")
		if err != nil {
			return errors.Wrap(err, "Failed to marshal Messages")
		}
		return errors.Errorf("Request failed with result_code: %d, messages: %v", resultCode, string(messageJSON))
	}

	// Use type switch to handle different response types
	log.Entry().Debugf("responseData type: %T", responseData) // Log type using %T
	switch data := responseData.(type) {
	case *ScanProjectResponse:
		return checkResponse(data.Success, data.Result.Messages, data.Result.ResultCode)
	case *GetScanJobStatusResponse:
		return checkResponse(data.Success, data.Result.Messages, 0)
	case *GetJobResultMetricsResponse:
		return checkResponse(data.Success, data.Result.Metrics, 0)
	default:
		return errors.New("Unknown response type")
	}
}
