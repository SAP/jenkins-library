package cmd

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/contrast"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type contrastExecuteScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newContrastExecuteScanTestsUtils() contrastExecuteScanMockUtils {
	utils := contrastExecuteScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestGetAuth(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := &contrastExecuteScanOptions{
			UserAPIKey: "user-api-key",
			Username:   "username",
			ServiceKey: "service-key",
		}
		authString := getAuth(config)
		assert.NotEmpty(t, authString)
		data, err := base64.StdEncoding.DecodeString(authString)
		assert.NoError(t, err)
		assert.Equal(t, "username:service-key", string(data))
	})
}

func TestGetApplicationUrls(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := &contrastExecuteScanOptions{
			Server:         "https://server.com",
			OrganizationID: "orgId",
			ApplicationID:  "appId",
		}
		appUrl, guiUrl := getApplicationUrls(config)
		assert.Equal(t, "https://server.com/api/v4/organizations/orgId/applications/appId", appUrl)
		assert.Equal(t, "https://server.com/Contrast/static/ng/index.html#/orgId/applications/appId", guiUrl)
	})
}

func TestValidateConfigs(t *testing.T) {
	t.Parallel()
	validConfig := contrastExecuteScanOptions{
		UserAPIKey:     "user-api-key",
		ServiceKey:     "service-key",
		Username:       "username",
		Server:         "https://server.com",
		OrganizationID: "orgId",
		ApplicationID:  "appId",
	}

	t.Run("Valid config", func(t *testing.T) {
		config := validConfig
		err := validateConfigs(&config)
		assert.NoError(t, err)
	})

	t.Run("Valid config, server url without https://", func(t *testing.T) {
		config := validConfig
		config.Server = "server.com"
		err := validateConfigs(&config)
		assert.NoError(t, err)
		assert.Equal(t, config.Server, "https://server.com")
	})

	t.Run("Empty config", func(t *testing.T) {
		config := contrastExecuteScanOptions{}

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty userAPIKey", func(t *testing.T) {
		config := validConfig
		config.UserAPIKey = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty username", func(t *testing.T) {
		config := validConfig
		config.Username = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty serviceKey", func(t *testing.T) {
		config := validConfig
		config.ServiceKey = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty server", func(t *testing.T) {
		config := validConfig
		config.Server = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty organizationId", func(t *testing.T) {
		config := validConfig
		config.OrganizationID = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})

	t.Run("Empty applicationID", func(t *testing.T) {
		config := validConfig
		config.ApplicationID = ""

		err := validateConfigs(&config)
		assert.Error(t, err)
	})
}

// Test constants for mock and end-to-end tests
const (
	// Mock test constants
	mockContrastAPIKey     = "mock-api-key"
	mockContrastServiceKey = "mock-service-key"
	mockContrastUsername   = "mock@example.com"
	mockContrastOrgID      = "org-mock-123"
	mockContrastServerURL  = "https://mock.contrastsecurity.com"
	mockContrastAppID      = "app-mock-456"

	// End-to-end test constants - Fill these with your real values for end-to-end testing
	e2eContrastAPIKey     = "YOUR_API_KEY"
	e2eContrastServiceKey = "YOUR_SERVICE_KEY"
	e2eContrastUsername   = "YOUR_USERNAME"
	e2eContrastOrgID      = "YOUR_ORG_ID"
	e2eContrastServerURL  = "https://YOUR_SERVER.contrastsecurity.com"
	e2eContrastAppID      = "YOUR_APP_ID"
)

// Mock-based unit tests (no real credentials needed)

func TestGenerateSarifReportMockSuccess(t *testing.T) {
	// Setup mock HTTP server
	serverURL := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.Contains(path, "/sarif/async") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "uuid": "test-sarif-uuid"}`))
		} else if strings.Contains(path, "/status") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "status": "ACTIVE", "downloadUrl": "` + serverURL + `/download"}`))
		} else if strings.Contains(path, "/download") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": "2.1.0"}`))
		}
	}))
	defer server.Close()
	serverURL = server.URL

	mockConfig := &contrastExecuteScanOptions{
		UserAPIKey:     mockContrastAPIKey,
		ServiceKey:     mockContrastServiceKey,
		Username:       mockContrastUsername,
		OrganizationID: mockContrastOrgID,
		Server:         server.URL,
		ApplicationID:  mockContrastAppID,
	}

	mockUtils := newContrastExecuteScanTestsUtils()
	mockClient := contrast.NewClient(
		mockContrastAPIKey,
		mockContrastServiceKey,
		mockContrastUsername,
		mockContrastOrgID,
		server.URL,
		server.URL+"/api/v4/organizations/"+mockContrastOrgID+"/applications/"+mockContrastAppID,
	)

	reports, err := generateSarifReport(mockConfig, mockUtils, mockClient)

	assert.NoError(t, err, "generateSarifReport should not return error")
	assert.NotEmpty(t, reports, "Expected reports to be generated")
	assert.Equal(t, 1, len(reports))
	assert.Equal(t, "Contrast SARIF Report", reports[0].Name)
}

func TestGeneratePdfReportMockSuccess(t *testing.T) {
	// Setup mock HTTP server
	serverURL := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.Contains(path, "/attestation") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "uuid": "test-pdf-uuid"}`))
		} else if strings.Contains(path, "/status") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "status": "ACTIVE", "downloadUrl": "` + serverURL + `/download"}`))
		} else if strings.Contains(path, "/download") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("PDF content"))
		}
	}))
	defer server.Close()
	serverURL = server.URL

	mockConfig := &contrastExecuteScanOptions{
		UserAPIKey:     mockContrastAPIKey,
		ServiceKey:     mockContrastServiceKey,
		Username:       mockContrastUsername,
		OrganizationID: mockContrastOrgID,
		Server:         server.URL,
		ApplicationID:  mockContrastAppID,
	}

	mockUtils := newContrastExecuteScanTestsUtils()
	mockClient := contrast.NewClient(
		mockContrastAPIKey,
		mockContrastServiceKey,
		mockContrastUsername,
		mockContrastOrgID,
		server.URL,
		server.URL+"/api/v4/organizations/"+mockContrastOrgID+"/applications/"+mockContrastAppID,
	)

	reports, err := generatePdfReport(mockConfig, mockUtils, mockClient)

	assert.NoError(t, err, "generatePdfReport should not return error")
	assert.NotEmpty(t, reports, "Expected reports to be generated")
	assert.Equal(t, 1, len(reports))
	assert.Equal(t, "Contrast PDF Attestation Report", reports[0].Name)
}

// TestContrastExecuteScanEndToEnd performs an end-to-end test of the runContrastExecuteScan function.
// It requires valid Contrast credentials to be set in the constants above.
// This test is skipped if the credentials are not filled in.
func TestContrastExecuteScanEndToEnd(t *testing.T) {
	if e2eContrastAPIKey == "YOUR_API_KEY" {
		t.Skip("Skipping end-to-end test: Contrast credentials not provided.")
	}

	outputDir := "./contrast-e2e-output"
	_ = os.RemoveAll(outputDir) // Best-effort cleanup
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	oldCWD, _ := os.Getwd()
	err := os.Chdir(outputDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(oldCWD)

	config := contrastExecuteScanOptions{
		Server:         e2eContrastServerURL,
		OrganizationID: e2eContrastOrgID,
		ApplicationID:  e2eContrastAppID,
		UserAPIKey:     e2eContrastAPIKey,
		Username:       e2eContrastUsername,
		ServiceKey:     e2eContrastServiceKey,
		GenerateSarif:  true,
		GeneratePdf:    true,
	}

	utils := newContrastExecuteScanUtils()

	reports, err := runContrastExecuteScan(&config, nil, utils)

	assert.NoError(t, err, "runContrastExecuteScan should not return an error")
	assert.NotEmpty(t, reports, "Expected reports to be generated")

	// Verify SARIF report
	sarifPath := filepath.Join(".", "contrast", "piper_contrast.sarif")
	assert.FileExists(t, sarifPath, "SARIF report file should exist")
	foundSarif := false
	for _, report := range reports {
		if filepath.Clean(report.Target) == filepath.Clean(sarifPath) {
			foundSarif = true
			break
		}
	}
	assert.True(t, foundSarif, "SARIF report should be in the returned reports list")

	// Verify PDF report
	pdfPath := filepath.Join(".", "contrast", "piper_contrast_attestation.pdf")
	assert.FileExists(t, pdfPath, "PDF report file should exist")
	foundPdf := false
	for _, report := range reports {
		if filepath.Clean(report.Target) == filepath.Clean(pdfPath) {
			foundPdf = true
			break
		}
	}
	assert.True(t, foundPdf, "PDF report should be in the returned reports list")
}
