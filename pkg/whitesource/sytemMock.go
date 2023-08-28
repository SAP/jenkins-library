//go:build !release
// +build !release

package whitesource

import (
	"fmt"
	"strconv"
	"time"
)

// SystemMock stores a number of WhiteSource objects and, based on that, mocks the behavior of System.
type SystemMock struct {
	ProductName         string
	Products            []Product
	Projects            []Project
	Alerts              []Alert
	IgnoredAlerts       []Alert
	AlertType           string
	AlertError          error
	Libraries           []Library
	RiskReport          []byte
	VulnerabilityReport []byte
}

func (m *SystemMock) GetProjectIgnoredAlertsByType(projectToken string, alertType string) ([]Alert, error) {
	return m.IgnoredAlerts, nil
}

// GetProductByName mimics retrieving a Product by name. It returns an error of no Product is stored in the mock.
func (m *SystemMock) GetProductByName(productName string) (Product, error) {
	for _, product := range m.Products {
		if product.Name == productName {
			return product, nil
		}
	}
	return Product{}, fmt.Errorf("no product with name '%s' found in Whitesource", productName)
}

// CreateProduct appends a new Product to the system mock and returns its token ("mock-product-token-<index>").
func (m *SystemMock) CreateProduct(productName string) (string, error) {
	now := time.Now().Format(DateTimeLayout)
	productIndex := len(m.Products)
	product := Product{
		Name:           productName,
		Token:          "mock-product-token-" + strconv.Itoa(productIndex),
		CreationDate:   now,
		LastUpdateDate: now,
	}
	m.Products = append(m.Products, product)
	return product.Token, nil
}

// SetProductAssignments checks if the system mock contains a product with the given token and returns
// an error depending on that, but otherwise does nothing with the provided arguments.
func (m *SystemMock) SetProductAssignments(productToken string, _, _, _ *Assignment) error {
	for _, product := range m.Products {
		if product.Token == productToken {
			return nil
		}
	}
	return fmt.Errorf("no product with that token")
}

// GetProjectsMetaInfo returns the list of Projects stored in the mock or an error if token is unknown.
func (m *SystemMock) GetProjectsMetaInfo(productToken string) ([]Project, error) {
	for _, product := range m.Products {
		if product.Token == productToken {
			return m.Projects, nil
		}
	}
	return nil, fmt.Errorf("no product with that token")
}

// GetProjectToken checks the Projects stored in the mock and returns a valid token, or an empty token and no error.
func (m *SystemMock) GetProjectToken(productToken, projectName string) (string, error) {
	for _, project := range m.Projects {
		if project.Name == projectName {
			return project.Token, nil
		}
	}
	return "", nil
}

// GetProjectByToken checks the Projects stored in the mock and returns the one with the given token or an error.
func (m *SystemMock) GetProjectByToken(projectToken string) (Project, error) {
	for _, project := range m.Projects {
		if project.Token == projectToken {
			return project, nil
		}
	}
	return Project{}, fmt.Errorf("no project with token '%s' found in Whitesource", projectToken)
}

// GetProjectRiskReport mocks retrieving a risc report.
func (m *SystemMock) GetProjectRiskReport(projectToken string) ([]byte, error) {
	return m.RiskReport, nil
}

// GetProjectVulnerabilityReport mocks retrieving a vulnerability report.
// Behavior depends on what is stored in the mock.
func (m *SystemMock) GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error) {
	_, err := m.GetProjectByToken(projectToken)
	if err != nil {
		return nil, err
	}
	if m.VulnerabilityReport == nil {
		return nil, fmt.Errorf("no report available")
	}
	return m.VulnerabilityReport, nil
}

// GetProjectAlerts returns the alerts stored in the SystemMock.
func (m *SystemMock) GetProjectAlerts(projectToken string) ([]Alert, error) {
	return m.Alerts, nil
}

// GetProjectAlertsByType returns the alerts stored in the SystemMock and records the type.
func (m *SystemMock) GetProjectAlertsByType(projectToken, alertType string) ([]Alert, error) {
	if m.AlertError != nil {
		return m.Alerts, m.AlertError
	}
	m.AlertType = alertType
	return m.Alerts, nil
}

// GetProjectLibraryLocations returns the libraries stored in the SystemMock.
func (m *SystemMock) GetProjectLibraryLocations(projectToken string) ([]Library, error) {
	return m.Libraries, nil
}

// GetProjectHierarchy returns the libraries stored in the SystemMock.
func (m *SystemMock) GetProjectHierarchy(projectToken string, inHouse bool) ([]Library, error) {
	return m.Libraries, nil
}

// NewSystemMockWithProjectName returns a pointer to a new instance of SystemMock using a project with a defined name.
func NewSystemMockWithProjectName(lastUpdateDate, projectName string) *SystemMock {
	mockLibrary := Library{
		Name:     "mock-library",
		Filename: "mock-library-file",
		Version:  "mock-library-version",
	}
	return &SystemMock{
		ProductName: "mock-product",
		Products: []Product{
			{
				Name:           "mock-product",
				Token:          "mock-product-token",
				CreationDate:   "last-thursday",
				LastUpdateDate: lastUpdateDate,
			},
		},
		Projects: []Project{
			{
				ID:             42,
				Name:           projectName,
				PluginName:     "mock-plugin-name",
				Token:          "mock-project-token",
				UploadedBy:     "MrBean",
				CreationDate:   "last-thursday",
				LastUpdateDate: lastUpdateDate,
			},
		},
		Alerts: []Alert{
			{
				Vulnerability: Vulnerability{
					Name:  "something severe",
					Score: 5,
				},
				Library:      mockLibrary,
				Project:      projectName,
				CreationDate: "last-thursday",
			},
		},
		Libraries:           []Library{mockLibrary},
		RiskReport:          []byte("mock-risk-report"),
		VulnerabilityReport: []byte("mock-vulnerability-report"),
	}
}

// NewSystemMock returns a pointer to a new instance of SystemMock.
func NewSystemMock(lastUpdateDate string) *SystemMock {
	const projectName = "mock-project - 1"
	return NewSystemMockWithProjectName(lastUpdateDate, projectName)
}
