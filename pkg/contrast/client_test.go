package contrast

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testClientOrgID      = "org-123"
	testClientApiKey     = "test-api-key"
	testClientUsername   = "test-username"
	testClientServiceKey = "test-service-key"
	testClientBaseURL    = "https://test.contrastsecurity.com"
	testClientAppURL     = "https://test.contrastsecurity.com/api/v4/organizations/org-123/applications/app-123"
	testReportUUID       = "test-report-uuid"

	// HTTP constants
	apiKeyHeader      = "API-Key"
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
	expectedNoError   = "Expected no error, got %v"
	downloadURL       = "http://example.com/download"
	defaultBaseURL    = "https://cs003.contrastsecurity.com"
)

func TestClientCreation(t *testing.T) {
	t.Run("with custom base URL", func(t *testing.T) {
		customURL := "https://custom.contrastsecurity.com"
		client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, customURL, testClientAppURL)
		verifyClientFields(t, client, customURL)
	})

	t.Run("with empty base URL uses default", func(t *testing.T) {
		client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, "", testClientAppURL)
		verifyClientFields(t, client, defaultBaseURL)
	})
}

func verifyClientFields(t *testing.T, client *Client, expectedURL string) {
	t.Helper()
	if client.ApiKey != testClientApiKey {
		t.Errorf("Expected ApiKey %s, got %s", testClientApiKey, client.ApiKey)
	}
	if client.ServiceKey != testClientServiceKey {
		t.Errorf("Expected ServiceKey %s, got %s", testClientServiceKey, client.ServiceKey)
	}
	if client.Username != testClientUsername {
		t.Errorf("Expected Username %s, got %s", testClientUsername, client.Username)
	}
	if client.OrgID != testClientOrgID {
		t.Errorf("Expected OrgID %s, got %s", testClientOrgID, client.OrgID)
	}
	if client.BaseURL != expectedURL {
		t.Errorf("Expected BaseURL %s, got %s", expectedURL, client.BaseURL)
	}
	if client.AppURL != testClientAppURL {
		t.Errorf("Expected AppURL %s, got %s", testClientAppURL, client.AppURL)
	}
	if client.Auth == "" {
		t.Error("Expected Auth to be set")
	}
	if client.HttpClient == nil {
		t.Error("Expected HttpClient to be initialized")
	}
	if client.HttpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.HttpClient.Timeout)
	}
}

func TestAddAuth(t *testing.T) {
	client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, testClientBaseURL, testClientAppURL)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client.addAuth(req)

	verifyBasicAuth(t, req)
	verifyAPIKeyHeader(t, req)
}

func verifyBasicAuth(t *testing.T, req *http.Request) {
	t.Helper()
	username, password, ok := req.BasicAuth()
	if !ok {
		t.Error("Expected Basic Auth to be set")
	}
	if username != testClientUsername {
		t.Errorf("Expected username %s, got %s", testClientUsername, username)
	}
	if password != testClientServiceKey {
		t.Errorf("Expected password %s, got %s", testClientServiceKey, password)
	}
}

func verifyAPIKeyHeader(t *testing.T, req *http.Request) {
	t.Helper()
	apiKey := req.Header.Get(apiKeyHeader)
	if apiKey != testClientApiKey {
		t.Errorf("Expected API-Key %s, got %s", testClientApiKey, apiKey)
	}
}

func TestCheckReportStatusSuccess(t *testing.T) {
	expectedResponse := ReportStatusResponse{
		Success:     true,
		Status:      "ACTIVE",
		DownloadUrl: downloadURL,
		Messages:    []string{"Report ready"},
	}

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		verifyGetRequest(t, r)
		verifyAuthHeaders(t, r)
		verifyJSONContentType(t, r)

		w.Header().Set(contentTypeHeader, applicationJSON)
		json.NewEncoder(w).Encode(expectedResponse)
	})
	defer server.Close()

	client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, testClientBaseURL, testClientAppURL)

	response, err := client.checkReportStatus(server.URL)
	if err != nil {
		t.Fatalf(expectedNoError, err)
	}

	verifyActiveResponse(t, response)
}

func createTestServer(t *testing.T, handler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(handler))
}

func verifyGetRequest(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Method != "GET" {
		t.Errorf("Expected GET method, got %s", r.Method)
	}
}

func verifyAuthHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	username, password, ok := r.BasicAuth()
	if !ok || username != testClientUsername || password != testClientServiceKey {
		t.Error("Expected proper Basic Auth")
	}
	if r.Header.Get(apiKeyHeader) != testClientApiKey {
		t.Error("Expected proper API-Key header")
	}
}

func verifyJSONContentType(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Header.Get(contentTypeHeader) != applicationJSON {
		t.Error("Expected Content-Type application/json")
	}
}

func verifyActiveResponse(t *testing.T, response *ReportStatusResponse) {
	t.Helper()
	if !response.Success {
		t.Error("Expected Success to be true")
	}
	if response.Status != "ACTIVE" {
		t.Errorf("Expected Status ACTIVE, got %s", response.Status)
	}
	if response.DownloadUrl != downloadURL {
		t.Errorf("Expected DownloadUrl %s, got %s", downloadURL, response.DownloadUrl)
	}
}

func TestCheckReportStatusServerError(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, testClientBaseURL, testClientAppURL)

	_, err := client.checkReportStatus(server.URL)
	if err == nil {
		t.Fatal("Expected error for server error")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("Expected status code error, got %v", err)
	}
}

func TestCheckReportStatusInvalidJSON(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, applicationJSON)
		w.Write([]byte("invalid json"))
	})
	defer server.Close()

	client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, testClientBaseURL, testClientAppURL)

	_, err := client.checkReportStatus(server.URL)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("Expected parse error, got %v", err)
	}
}

func TestPollReportStatusSuccess(t *testing.T) {
	callCount := 0
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++

		response := createPollResponse(callCount)
		w.Header().Set(contentTypeHeader, applicationJSON)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, server.URL, testClientAppURL)

	// Note: This test will take actual time due to real sleeps in the polling function
	response, err := client.PollReportStatus(testReportUUID, "TEST")
	if err != nil {
		t.Fatalf(expectedNoError, err)
	}

	verifyPollSuccess(t, response, callCount)
}

func verifyPollSuccess(t *testing.T, response *ReportStatusResponse, callCount int) {
	t.Helper()
	if !response.Success {
		t.Error("Expected Success to be true")
	}
	if response.Status != "ACTIVE" {
		t.Errorf("Expected Status ACTIVE, got %s", response.Status)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestDownloadReportSuccess(t *testing.T) {
	expectedData := []byte("test report data")

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		verifyPostRequest(t, r)
		verifyAuthHeaders(t, r)
		verifyJSONContentType(t, r)

		w.WriteHeader(http.StatusOK)
		w.Write(expectedData)
	})
	defer server.Close()

	client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, testClientBaseURL, testClientAppURL)

	data, err := client.DownloadReport(server.URL, "TEST")
	if err != nil {
		t.Fatalf(expectedNoError, err)
	}

	if string(data) != string(expectedData) {
		t.Errorf("Expected data %s, got %s", string(expectedData), string(data))
	}
}

func verifyPostRequest(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Method != "POST" {
		t.Errorf("Expected POST method, got %s", r.Method)
	}
}

func TestDownloadReportServerError(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	})
	defer server.Close()

	client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, testClientBaseURL, testClientAppURL)

	_, err := client.DownloadReport(server.URL, "TEST")
	if err == nil {
		t.Fatal("Expected error for server error")
	}

	verifyServerError(t, err)
}

func verifyServerError(t *testing.T, err error) {
	t.Helper()
	if !strings.Contains(err.Error(), "unexpected status code: 404") {
		t.Errorf("Expected status code error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Not found") {
		t.Errorf("Expected error body in message, got %v", err)
	}
}

func TestNewPollConfig(t *testing.T) {
	config := newPollConfig()
	verifyPollConfigValues(t, config)
}

func verifyPollConfigValues(t *testing.T, config pollConfig) {
	t.Helper()

	expectedValues := map[string]interface{}{
		"maxTotalWait":    5 * time.Minute,
		"maxPollInterval": 60 * time.Second,
		"initialDelay":    15 * time.Second,
		"pollInterval":    5 * time.Second,
		"backoffFactor":   1.5,
	}

	actualValues := map[string]interface{}{
		"maxTotalWait":    config.maxTotalWait,
		"maxPollInterval": config.maxPollInterval,
		"initialDelay":    config.initialDelay,
		"pollInterval":    config.pollInterval,
		"backoffFactor":   config.backoffFactor,
	}

	for key, expected := range expectedValues {
		actual := actualValues[key]
		if actual != expected {
			t.Errorf("Expected %s %v, got %v", key, expected, actual)
		}
	}
}

func TestWaitAndBackoff(t *testing.T) {
	client := NewClient(testClientApiKey, testClientServiceKey, testClientUsername, testClientOrgID, testClientBaseURL, testClientAppURL)
	config := newPollConfig()

	testNormalBackoff(t, client, config)
	testMaxIntervalCapping(t, client, config)
}

func createPollResponse(callCount int) ReportStatusResponse {
	switch callCount {
	case 1:
		return ReportStatusResponse{Success: true, Status: "CREATING"}
	case 2:
		return ReportStatusResponse{Success: true, Status: "ACTIVE", DownloadUrl: downloadURL}
	default:
		return ReportStatusResponse{Success: false, Messages: []string{"Too many calls"}}
	}
}

func testNormalBackoff(t *testing.T, client *Client, config pollConfig) {
	t.Helper()

	totalWaited, nextInterval := client.waitAndBackoff(0, config, "TEST")
	expectedTotal := config.pollInterval
	expectedNext := time.Duration(float64(config.pollInterval) * config.backoffFactor)

	if totalWaited != expectedTotal {
		t.Errorf("Expected totalWaited %v, got %v", expectedTotal, totalWaited)
	}
	if nextInterval != expectedNext {
		t.Errorf("Expected nextInterval %v, got %v", expectedNext, nextInterval)
	}
}

func testMaxIntervalCapping(t *testing.T, client *Client, config pollConfig) {
	t.Helper()

	config.pollInterval = 50 * time.Second
	_, nextInterval := client.waitAndBackoff(0, config, "TEST")

	if nextInterval != config.maxPollInterval {
		t.Errorf("Expected nextInterval to be capped at %v, got %v", config.maxPollInterval, nextInterval)
	}
}
