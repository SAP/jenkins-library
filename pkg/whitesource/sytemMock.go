// +build !release

package whitesource

import "fmt"

// SystemMock stores a number of WhiteSource objects and, based on that, mocks the behavior of System.
type SystemMock struct {
	ProductName         string
	Products            []Product
	Projects            []Project
	Alerts              []Alert
	Libraries           []Library
	RiskReport          []byte
	VulnerabilityReport []byte
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

// GetProjectLibraryLocations returns the libraries stored in the SystemMock.
func (m *SystemMock) GetProjectLibraryLocations(projectToken string) ([]Library, error) {
	return m.Libraries, nil
}

// NewSystemMock returns a pointer to a new instance of SystemMock.
func NewSystemMock(lastUpdateDate string) *SystemMock {
	const projectName = "mock-project - 1"
	mockLibrary := Library{
		Name:     "mock-library",
		Filename: "mock-library-file",
		Version:  "mock-library-version",
		Project:  projectName,
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
