package cmd

import (
	"github.com/SAP/jenkins-library/pkg/whitesource"
	"os"
	"testing"
)

type whitesourceSystemMock struct {
	productName         string
	products            []whitesource.Product
	projects            []whitesource.Project
	alerts              []whitesource.Alert
	libraries           []whitesource.Library
	riskReport          []byte
	vulnerabilityReport []byte
}

func (m *whitesourceSystemMock) GetProductsMetaInfo() ([]whitesource.Product, error) {
	return m.products, nil
}

func (m *whitesourceSystemMock) GetMetaInfoForProduct(productName string) (whitesource.Product, error) {
	return m.products[0], nil
}

func (m *whitesourceSystemMock) GetProjectsMetaInfo(productToken string) ([]whitesource.Project, error) {
	return m.projects, nil
}

func (m *whitesourceSystemMock) GetProjectToken(productToken, projectName string) (string, error) {
	return "mock-project-token", nil
}

func (m *whitesourceSystemMock) GetProjectVitals(projectToken string) (*whitesource.Project, error) {
	return &m.projects[0], nil
}

func (m *whitesourceSystemMock) GetProjectByName(productToken, projectName string) (*whitesource.Project, error) {
	return &m.projects[0], nil
}

func (m *whitesourceSystemMock) GetProjectsByIDs(productToken string, projectIDs []int64) ([]whitesource.Project, error) {
	return m.projects, nil
}

func (m *whitesourceSystemMock) GetProjectTokens(productToken string, projectNames []string) ([]string, error) {
	return []string{"mock-project-token-1", "mock-project-token-2"}, nil
}

func (m *whitesourceSystemMock) GetProductName(productToken string) (string, error) {
	return m.productName, nil
}

func (m *whitesourceSystemMock) GetProjectRiskReport(projectToken string) ([]byte, error) {
	return m.riskReport, nil
}

func (m *whitesourceSystemMock) GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error) {
	return m.vulnerabilityReport, nil
}

func (m *whitesourceSystemMock) GetOrganizationProductVitals() ([]whitesource.Product, error) {
	return m.products, nil
}

func (m *whitesourceSystemMock) GetProductByName(productName string) (*whitesource.Product, error) {
	return &m.products[0], nil
}

func (m *whitesourceSystemMock) GetProjectAlerts(projectToken string) ([]whitesource.Alert, error) {
	return m.alerts, nil
}

func (m *whitesourceSystemMock) GetProjectLibraryLocations(projectToken string) ([]whitesource.Library, error) {
	return m.libraries, nil
}

var mockLibrary = whitesource.Library{
	Name:     "mock-library",
	Filename: "mock-library-file",
	Version:  "mock-library-version",
	Project:  "mock-project",
}

func newWhitesourceSystemMock() *whitesourceSystemMock {
	return &whitesourceSystemMock{
		productName: "mock-product",
		products: []whitesource.Product{
			{
				Name:           "mock-product",
				Token:          "mock-product-token",
				CreationDate:   "yesterday",
				LastUpdateDate: "last-thursday",
			},
		},
		projects: []whitesource.Project{
			{
				ID:             42,
				Name:           "mock-project",
				PluginName:     "mock-plugin-name",
				Token:          "mock-project-token",
				UploadedBy:     "MrBean",
				CreationDate:   "yesterday",
				LastUpdateDate: "last-thursday",
			},
		},
		alerts: []whitesource.Alert{
			{
				Vulnerability: whitesource.Vulnerability{},
				Library:       mockLibrary,
				Project:       "mock-project",
				CreationDate:  "yesterday",
			},
		},
		libraries:           []whitesource.Library{mockLibrary},
		riskReport:          []byte("mock-risk-report"),
		vulnerabilityReport: []byte("mock-vulnerability-report"),
	}
}

func TestProjectCoordinates(t *testing.T) {
	t.Run("MTA project", func(t *testing.T) {
		err := os.Chdir("../../GettingStartedBookshop")
		if err != nil {
			t.Fatal("sample project not found")
		}
		config := ScanOptions{
			ScanType:               "mta",
			DefaultVersioningModel: "major",
		}
		utils := newUtils()
		systemMock := newWhitesourceSystemMock()
		err = resolveProjectIdentifiers(&config, utils, systemMock)
	})

}
