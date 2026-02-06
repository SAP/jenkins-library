package contrast

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStartAsyncPdfGeneration_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/Contrast/api/ng/org-123/applications/app-456/attestation" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		// API now uses JSON body instead of query params
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "uuid": "pdf-uuid-123", "messages": []}`))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      "org-123",
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	uuid, err := client.StartAsyncPdfGeneration("app-456")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if uuid != "pdf-uuid-123" {
		t.Errorf("Expected uuid 'pdf-uuid-123', got '%s'", uuid)
	}
}

func TestStartAsyncPdfGeneration_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      "org-123",
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.StartAsyncPdfGeneration("app-456")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestStartAsyncPdfGeneration_FailedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": false, "messages": ["PDF generation failed"], "uuid": ""}`))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      "org-123",
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.StartAsyncPdfGeneration("app-456")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestStartAsyncPdfGeneration_EmptyUuid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "messages": [], "uuid": ""}`))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      "org-123",
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.StartAsyncPdfGeneration("app-456")
	if err == nil {
		t.Fatal("Expected error for empty UUID, got nil")
	}
}

func TestStartAsyncPdfGeneration_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      "org-123",
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.StartAsyncPdfGeneration("app-456")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestGeneratePdfReport_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch callCount {
		case 1:
			// StartAsyncPdfGeneration call
			if r.Method != "POST" || !contains(r.URL.Path, "/attestation") {
				t.Errorf("Expected PDF start call, got %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "uuid": "pdf-uuid-123"}`))

		case 2:
			// PollReportStatus call
			if r.Method != "GET" || !contains(r.URL.Path, "/status") {
				t.Errorf("Expected status poll call, got %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "status": "ACTIVE", "downloadUrl": ""}`))

		case 3:
			// DownloadReport call
			if r.Method != "POST" || !contains(r.URL.Path, "/download") {
				t.Errorf("Expected download call, got %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("PDF report content"))

		default:
			t.Errorf("Unexpected call count: %d", callCount)
		}
	}))
	defer server.Close()

	client := &Client{
		OrgID:      "org-123",
		BaseURL:    server.URL,
		HttpClient: &http.Client{Timeout: 1 * time.Second}, // Short timeout for tests
	}

	data, err := client.GeneratePdfReport("app-456")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(data) != "PDF report content" {
		t.Errorf("Expected 'PDF report content', got '%s'", string(data))
	}
	if callCount != 3 {
		t.Errorf("Expected 3 API calls, got %d", callCount)
	}
}

func TestGeneratePdfReport_StartFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	client := &Client{
		OrgID:      "org-123",
		BaseURL:    server.URL,
		HttpClient: server.Client(),
	}

	_, err := client.GeneratePdfReport("app-456")
	if err == nil {
		t.Fatal("Expected error when start fails, got nil")
	}
}

func TestGeneratePdfReport_PollFails(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// StartAsyncPdfGeneration succeeds
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "uuid": "pdf-uuid-123"}`))
		} else {
			// PollReportStatus fails
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := &Client{
		OrgID:      "org-123",
		BaseURL:    server.URL,
		HttpClient: &http.Client{Timeout: 1 * time.Second},
	}

	_, err := client.GeneratePdfReport("app-456")
	if err == nil {
		t.Fatal("Expected error when poll fails, got nil")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			(len(substr) < len(s) && s[1:len(substr)+1] == substr))))
}
