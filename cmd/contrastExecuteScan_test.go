package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func newMockContrastClient(server *httptest.Server) *contrast.Client {
	return contrast.NewClient(
		mockContrastAPIKey,
		mockContrastServiceKey,
		mockContrastUsername,
		mockContrastOrgID,
		server.URL,
		server.URL+"/api/v4/organizations/"+mockContrastOrgID+"/applications/"+mockContrastAppID,
	)
}

func newMockConfig(serverURL string) *contrastExecuteScanOptions {
	return &contrastExecuteScanOptions{
		UserAPIKey:     mockContrastAPIKey,
		ServiceKey:     mockContrastServiceKey,
		Username:       mockContrastUsername,
		OrganizationID: mockContrastOrgID,
		Server:         serverURL,
		ApplicationID:  mockContrastAppID,
	}
}

func TestCheckAgentSetup(t *testing.T) {
	t.Run("No servers connected returns hard error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true,"count":0,"servers":[]}`))
		}))
		defer server.Close()

		result, err := checkAgentSetup(newMockContrastClient(server), newMockConfig(server.URL))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no agents connected")
		assert.Nil(t, result)
	})

	t.Run("Active server within threshold sets no inactivity violation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/servers") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true,"count":1,"servers":[{"server_id":1,"last_activity":` +
					strings.TrimSpace(fmt.Sprintf("%d", (func() int64 { return time.Now().UnixMilli() })())) +
					`}]}`))
			} else if strings.Contains(r.URL.Path, "/route") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true,"discovered_count":0,"exercised_count":0}`))
			}
		}))
		defer server.Close()

		config := newMockConfig(server.URL)
		config.AgentInactivityThresholdDays = 7
		result, err := checkAgentSetup(newMockContrastClient(server), config)
		assert.NoError(t, err)
		assert.Nil(t, result.InactivityViolation)
	})

	t.Run("Inactive server beyond threshold sets inactivity violation", func(t *testing.T) {
		oldActivity := time.Now().Add(-10 * 24 * time.Hour).UnixMilli()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/servers") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{"success":true,"count":1,"servers":[{"server_id":1,"last_activity":%d}]}`, oldActivity)))
			} else if strings.Contains(r.URL.Path, "/route") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true,"discovered_count":0,"exercised_count":0}`))
			}
		}))
		defer server.Close()

		config := newMockConfig(server.URL)
		config.AgentInactivityThresholdDays = 7
		result, err := checkAgentSetup(newMockContrastClient(server), config)
		assert.NoError(t, err)
		assert.NotNil(t, result.InactivityViolation)
		assert.Contains(t, result.InactivityViolation.Error(), "No agent has been active")
	})
}

func TestCheckRouteCoverage(t *testing.T) {
	t.Run("Coverage below threshold sets violation and warns", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true,"discovered_count":100,"exercised_count":20}`))
		}))
		defer server.Close()

		config := newMockConfig(server.URL)
		config.RouteCoverageThreshold = 30
		result := &agentSetupResult{}
		checkRouteCoverage(newMockContrastClient(server), config, result)

		assert.NotNil(t, result.RouteCoverageViolation)
		assert.NotNil(t, result.RouteCoveragePct)
		assert.InDelta(t, 20.0, *result.RouteCoveragePct, 0.1)
		assert.Equal(t, 100, *result.RouteDiscoveredCount)
		assert.Equal(t, 20, *result.RouteExercisedCount)
	})

	t.Run("Coverage above threshold sets no violation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true,"discovered_count":100,"exercised_count":80}`))
		}))
		defer server.Close()

		config := newMockConfig(server.URL)
		config.RouteCoverageThreshold = 30
		result := &agentSetupResult{}
		checkRouteCoverage(newMockContrastClient(server), config, result)

		assert.Nil(t, result.RouteCoverageViolation)
		assert.NotNil(t, result.RouteCoveragePct)
		assert.InDelta(t, 80.0, *result.RouteCoveragePct, 0.1)
	})

	t.Run("No routes discovered sets no violation and nil coverage", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true,"discovered_count":0,"exercised_count":0}`))
		}))
		defer server.Close()

		result := &agentSetupResult{}
		checkRouteCoverage(newMockContrastClient(server), newMockConfig(server.URL), result)

		assert.Nil(t, result.RouteCoverageViolation)
		assert.Nil(t, result.RouteCoveragePct)
	})
}

func TestCheckForComplianceWithNewThresholds(t *testing.T) {
	inactivityErr := fmt.Errorf("No agent has been active in the last 7 day(s)")
	routeErr := fmt.Errorf("Route coverage check: only 10.0%% of discovered routes have been exercised")

	t.Run("CheckForCompliance false: violations do not fail build", func(t *testing.T) {
		setup := &agentSetupResult{
			InactivityViolation:    inactivityErr,
			RouteCoverageViolation: routeErr,
		}
		config := &contrastExecuteScanOptions{CheckForCompliance: false}
		err := enforceComplianceThresholds(config, setup)
		assert.NoError(t, err)
	})

	t.Run("CheckForCompliance true: inactivity violation fails build", func(t *testing.T) {
		setup := &agentSetupResult{InactivityViolation: inactivityErr}
		config := &contrastExecuteScanOptions{CheckForCompliance: true}
		err := enforceComplianceThresholds(config, setup)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "No agent has been active")
	})

	t.Run("CheckForCompliance true: route coverage violation fails build", func(t *testing.T) {
		setup := &agentSetupResult{RouteCoverageViolation: routeErr}
		config := &contrastExecuteScanOptions{CheckForCompliance: true}
		err := enforceComplianceThresholds(config, setup)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Route coverage check")
	})

	t.Run("CheckForCompliance true: no violations passes", func(t *testing.T) {
		setup := &agentSetupResult{}
		config := &contrastExecuteScanOptions{CheckForCompliance: true}
		err := enforceComplianceThresholds(config, setup)
		assert.NoError(t, err)
	})
}

func TestContrastAuditJSONReport(t *testing.T) {
	t.Run("Route counts included in JSON when present", func(t *testing.T) {
		discovered := 100
		exercised := 80
		audit := contrast.ContrastAudit{
			ToolName:             "contrast",
			ApplicationUrl:       "https://example.com",
			ScanResults:          []contrast.ContrastFindings{},
			RouteDiscoveredCount: &discovered,
			RouteExercisedCount:  &exercised,
		}
		data, err := json.Marshal(audit)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"routeDiscoveredCount":100`)
		assert.Contains(t, string(data), `"routeExercisedCount":80`)
	})

	t.Run("Route counts omitted from JSON when nil", func(t *testing.T) {
		audit := contrast.ContrastAudit{
			ToolName:       "contrast",
			ApplicationUrl: "https://example.com",
			ScanResults:    []contrast.ContrastFindings{},
		}
		data, err := json.Marshal(audit)
		assert.NoError(t, err)
		assert.NotContains(t, string(data), "routeDiscoveredCount")
		assert.NotContains(t, string(data), "routeExercisedCount")
	})
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
