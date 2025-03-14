package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func onapsisExecuteScan(config onapsisExecuteScanOptions, telemetryData *telemetry.CustomData) {

	debugMode = config.DebugMode
	if debugMode {
		log.SetVerbose(true)
	}

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runOnapsisExecuteScan(config, telemetryData)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runOnapsisExecuteScan(config onapsisExecuteScanOptions, telemetryData *telemetry.CustomData) error {
	// Create a new ScanServer
	log.Entry().Info("Creating scan server...")
	server, err := NewScanServer(config)
	if err != nil {
		return errors.Wrap(err, "failed to create scan server")
	}

	// Call the ScanProject method
	log.Entry().Info("Scanning project...")
	startScanResponse, err := server.ScanProject(config, telemetryData)
	if err != nil {
		return errors.Wrap(err, "Failed to scan project")
	}

	// Monitor Job Status
	jobID := startScanResponse.Result.JobID
	log.Entry().Infof("Monitoring job %s status...", jobID)
	jobStatusResponse, err := server.MonitorJobStatus(jobID)
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
	loc, numMandatory, numOptional, totalTime := extractMetrics(metrics)
	log.Entry().Infof("Job Metrics - Lines of Code Scanned: %s, Mandatory Findings: %s, Optional Findings: %s, Total Time: %s seconds", loc, numMandatory, numOptional, totalTime)

	log.Entry().Infof("The findings can be viewed here: %s/ui/#/admin/scans/%s/%s/findings", config.ScanServiceURL, jobID, jobStatusResponse.Result.Details.Children[0])

	if config.FailOnMandatoryFinding && numMandatory != "0" {
		return errors.Errorf("Scan failed with %s mandatory findings", numMandatory)
	} else if config.FailOnOptionalFinding && numOptional != "0" {
		return errors.Errorf("Scan failed with %s optional findings", numOptional)
	}

	return nil
}

type ScanServer struct {
	serverUrl                   string
	client                      piperHttp.Sender
	authToken                   *AuthResponse
	authTokenExpirationTime     time.Time
	scanServiceCertificatePath  string
}

func NewScanServer(config onapsisExecuteScanOptions) (*ScanServer, error) {

	scanServiceUrl := config.ScanServiceURL

	scanServiceCertificatePath := config.OnapsisCertificatePath
	options := getHttpOptionsWithBasicAuth(config.OnapsisUsername, config.OnapsisPassword, scanServiceCertificatePath)

	client := &piperHttp.Client{}
	client.SetOptions(options)

	token, tokenErr := getJWTFromService(client, scanServiceUrl)
	if tokenErr != nil {
		log.Entry().WithError(tokenErr).Fatal("Error obtaining JWT")
	}
	tokenExpirationTime := getJwtExpirationTime(token.ExpiresIn)

	options = getHttpOptionsWithJwt(token.AccessToken, token.TokenType, scanServiceCertificatePath)
	client.SetOptions(options)

	server := &ScanServer{serverUrl: scanServiceUrl, client: client, authToken: token, authTokenExpirationTime: tokenExpirationTime, scanServiceCertificatePath: scanServiceCertificatePath}

	return server, nil
}

func (srv *ScanServer) RefreshServerAuth() {
	newJwt, refreshTokenErr := refreshJwtToken(srv.client, srv.authToken.RefreshToken, srv.serverUrl)

	if refreshTokenErr != nil {
		log.Entry().WithError(refreshTokenErr).Fatal("Error obtaining refreshed JWT")
	}

	options := getHttpOptionsWithJwt(newJwt.AccessToken, newJwt.TokenType, srv.scanServiceCertificatePath)
	srv.client.SetOptions(options)

	srv.authTokenExpirationTime = getJwtExpirationTime(newJwt.ExpiresIn)
	srv.authToken = newJwt
}

func getHttpOptionsWithBasicAuth(username string, password string, caCert string) piperHttp.ClientOptions {
	// Set BasicAuth for http client
	return piperHttp.ClientOptions{
		Username:                   username,
		Password:                   password,
		MaxRequestDuration:         60 * time.Second,
		DoLogRequestBodyOnDebug:    debugMode,
		DoLogResponseBodyOnDebug:   debugMode,
		TrustedCerts:               []string{caCert},
	}
}

// Obtain http.ClientOptions with JWT and tokenType "Bearer". caCert is the self-signed scan server certificate.
func getHttpOptionsWithJwt(jwt string, tokenType string, caCert string) piperHttp.ClientOptions {
	// Set authorization token for client
	return piperHttp.ClientOptions{
		Token:                      fmt.Sprintf("%s %s",tokenType, jwt),
		MaxRequestDuration:         60 * time.Second,
		DoLogRequestBodyOnDebug:    debugMode,
		DoLogResponseBodyOnDebug:   debugMode,
		TrustedCerts:               []string{caCert},
	}
}

func (srv *ScanServer) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	if time.Now().After(srv.authTokenExpirationTime) {
		srv.RefreshServerAuth()
	}

	return srv.client.SendRequest(method, url, body, header, cookies)
}

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

// getJWTFromService fetches a JWT using Basic Auth
func getJWTFromService(client piperHttp.Sender, scanServiceUrl string) (*AuthResponse, error) {

	jwtUrl := scanServiceUrl + "/cca/v1.2/auth_token"

	resp, err := client.SendRequest("GET", jwtUrl, nil, nil, nil)
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

func refreshJwtToken(client piperHttp.Sender, refreshToken, scanServiceUrl string) (*AuthResponse, error) {
	refreshTokenURL := scanServiceUrl + "/cca/v1.2/auth_token"

	// Create the refresh Cookie
	refreshCookie := http.Cookie{Name: "refresh_token", Value: refreshToken}

	resp, err := client.SendRequest("GET", refreshTokenURL, nil, nil, []*http.Cookie {&refreshCookie})
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Handle response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refresh token, status: %d, response: %s", resp.StatusCode, string(body))
	}

	// Extract new JWT from the response (assuming JSON response)
	var newJwt AuthResponse
	if err := json.Unmarshal(body, &newJwt); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &newJwt, nil
}

func (srv *ScanServer) ScanProject(config onapsisExecuteScanOptions, telemetryData *telemetry.CustomData) (ScanProjectResponse, error) {

	jobName, jobNameIsPresent := os.LookupEnv("JOB_BASE_NAME")
	if !jobNameIsPresent {
		jobName = "piper-ci-cd-scan"
	}

	jobDescription := fmt.Sprintf("Job triggered by CI/CD pipeline on git repo: %s, branch: %s", config.ScanGitURL, config.ScanGitBranch)

	// Create request data
	log.Entry().Info("Creating request data...")
	scanRequest := fmt.Sprintf(`{
		"engine_type": "GIT",
		"scan_information": {
			"name": "%s",
			"description": "%s"
		},
		"asset": {
			"type": "GITURL",
			"url": "%s"
		},
		"configuration": {
			"origin": "PIPER"
		},
		"scan_scope": {
			"languages": [
				"%s"
			],
			"branch_name": "%s",
			"exclude_packages": []
		}
	}`, jobName, jobDescription, config.ScanGitURL, config.AppType, config.ScanGitBranch)

	scanRequestReader := strings.NewReader(scanRequest)
	scanRequestHeader := http.Header{
		"Content-Type": {"application/json"},
	}

	// Send request
	log.Entry().Info("Sending scan request...")
	response, err := srv.SendRequest("POST", srv.serverUrl+"/cca/v1.2/scan", scanRequestReader, scanRequestHeader, nil)
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

func (srv *ScanServer) GetScanJobStatus(jobID string) (*GetScanJobStatusResponse, error) {
	// Send request
	response, err := srv.SendRequest("GET", srv.serverUrl+"/cca/v1.2/jobs/"+jobID, nil, nil, nil)
	if err != nil {
		return &GetScanJobStatusResponse{}, errors.Wrap(err, "failed to send request")
	}

	var responseData GetScanJobStatusResponse
	err = handleResponse(response, &responseData)
	if err != nil {
		return &responseData, errors.Wrap(err, "Failed to parse response")
	}

	return &responseData, nil
}

func (srv *ScanServer) MonitorJobStatus(jobID string) (*GetScanJobStatusResponse, error) {
	// Polling interval
	interval := time.Second * 10 // Check every 10 seconds
	for {
		// Get the job status
		response, err := srv.GetScanJobStatus(jobID)
		if err != nil {
			return &GetScanJobStatusResponse{}, errors.Wrap(err, "Failed to get scan job status")
		}

		// Log job progress
		log.Entry().Infof("Job %s progress: %d%%", jobID, response.Result.Progress)

		// Check if the job is complete
		if response.Result.Status == "SUCCESS" {
			log.Entry().Infof("Job %s progress: %d%%. Status: %s", jobID, response.Result.Progress, response.Result.Status)
			return response, nil
		} else if response.Result.Status == "FAILURE" {
			return &GetScanJobStatusResponse{}, errors.Errorf("Job %s failed with status: %s", jobID, response.Result.Status)
		}

		// Wait before checking again
		time.Sleep(interval)
	}
}

func (srv *ScanServer) GetJobReports(jobID string, reportArchiveName string) error {
	response, err := srv.SendRequest("GET", srv.serverUrl+"/cca/v1.2/jobs/"+jobID+"/result?format=ZIP", nil, nil, nil)
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
	log.Entry().Info("Writing report file...")
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		return errors.Wrap(err, "Failed to write report archive")
	}

	log.Entry().Info("Report written.")

	return nil
}

func (srv *ScanServer) GetJobResultMetrics(jobID string) (GetJobResultMetricsResponse, error) {
	// Send request
	response, err := srv.SendRequest("GET", srv.serverUrl+"/cca/v1.2/jobs/"+jobID+"/result/metrics", nil, nil, nil)
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

func extractMetrics(response GetJobResultMetricsResponse) (loc, numMandatory, numOptional, totalTime string) {
	for _, metric := range response.Result.Metrics {
		switch metric.Name {
		case "LOC":
			loc = metric.Value
		case "num_mandatory":
			numMandatory = metric.Value
		case "num_optional":
			numOptional = metric.Value
		case "total_time_used":
			totalTime = metric.Value
		}

	}

	return loc, numMandatory, numOptional, totalTime
}

func getJwtExpirationTime(expiresIn int) time.Time {
	tokenValidity := time.Duration(expiresIn * int(time.Second))

	return time.Now().Add(tokenValidity)
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
		Details       struct {
			Children []string `json:"children"`
		} `json:"details"`
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

var debugMode bool = false

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
