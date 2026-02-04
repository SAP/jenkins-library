package contrast

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testOrgID = "org-123"
	testAppID = "app-456"
	testUUID  = "sarif-uuid-123"
)

func TestStartAsyncSarifGenerationSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		expectedPath := "/Contrast/api/ng/organizations/" + testOrgID + "/applications/" + testAppID + "/sarif/async"
		if r.URL.Path != expectedPath {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "uuid": "` + testUUID + `", "messages": []}`))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	uuid, err := client.StartAsyncSarifGeneration(testAppID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if uuid != testUUID {
		t.Errorf("Expected uuid '%s', got '%s'", testUUID, uuid)
	}
}

func TestStartAsyncSarifGenerationServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.StartAsyncSarifGeneration(testAppID)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestStartAsyncSarifGenerationFailedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": false, "messages": ["SARIF generation failed"], "uuid": ""}`))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.StartAsyncSarifGeneration(testAppID)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestStartAsyncSarifGenerationEmptyUuid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "messages": [], "uuid": ""}`))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.StartAsyncSarifGeneration(testAppID)
	if err == nil {
		t.Fatal("Expected error for empty UUID, got nil")
	}
}

func TestStartAsyncSarifGenerationInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.StartAsyncSarifGeneration(testAppID)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestGenerateSarifReportSuccess(t *testing.T) {
	callCount := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch callCount {
		case 1:
			// StartAsyncSarifGeneration call
			if r.Method != "POST" || !strings.Contains(r.URL.Path, "/sarif/async") {
				t.Errorf("Expected SARIF start call, got %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "uuid": "` + testUUID + `"}`))

		case 2:
			// PollReportStatus call
			if r.Method != "GET" || !strings.Contains(r.URL.Path, "/status") {
				t.Errorf("Expected status poll call, got %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "status": "ACTIVE", "downloadUrl": "` + serverURL + `/download"}`))

		case 3:
			// DownloadReport call
			if r.Method != "POST" || !strings.Contains(r.URL.Path, "/download") {
				t.Errorf("Expected download call, got %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"$schema": "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json", "version": "2.1.0"}`))
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: &http.Client{Timeout: 1 * time.Second},
	}

	data, err := client.GenerateSarifReport(testAppID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedContent := `{"$schema": "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json", "version": "2.1.0"}`
	if string(data) != expectedContent {
		t.Errorf("Expected SARIF content, got '%s'", string(data))
	}
	if callCount != 3 {
		t.Errorf("Expected 3 API calls, got %d", callCount)
	}
}

func TestGenerateSarifReportStartFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.GenerateSarifReport(testAppID)
	if err == nil {
		t.Fatal("Expected error when start fails, got nil")
	}
}

func TestGenerateSarifReportPollFails(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// StartAsyncSarifGeneration succeeds
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "uuid": "` + testUUID + `"}`))
		} else {
			// PollReportStatus fails
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: &http.Client{Timeout: 1 * time.Second},
	}

	_, err := client.GenerateSarifReport(testAppID)
	if err == nil {
		t.Fatal("Expected error when poll fails, got nil")
	}
}

func TestGenerateSarifReportNoDownloadURL(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch callCount {
		case 1:
			// StartAsyncSarifGeneration succeeds
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "uuid": "` + testUUID + `"}`))

		case 2:
			// PollReportStatus succeeds but no download URL
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "status": "ACTIVE", "downloadUrl": ""}`))
		}
	}))
	defer server.Close()

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: &http.Client{Timeout: 1 * time.Second},
	}

	_, err := client.GenerateSarifReport(testAppID)
	if err == nil {
		t.Fatal("Expected error when download URL is empty, got nil")
	}
}

func TestGenerateSarifReportDownloadFails(t *testing.T) {
	callCount := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch callCount {
		case 1:
			// StartAsyncSarifGeneration succeeds
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "uuid": "` + testUUID + `"}`))

		case 2:
			// PollReportStatus succeeds
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "status": "ACTIVE", "downloadUrl": "` + serverURL + `/download"}`))

		case 3:
			// DownloadReport fails
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client := &Client{
		OrgID:      testOrgID,
		BaseURL:    server.URL,
		HttpClient: &http.Client{Timeout: 1 * time.Second},
	}

	_, err := client.GenerateSarifReport(testAppID)
	if err == nil {
		t.Fatal("Expected error when download fails, got nil")
	}
}
