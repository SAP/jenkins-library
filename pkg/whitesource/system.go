package whitesource

// System declares an interface for talking to the Whitesource service.
type System interface {
	GetProductsMetaInfo() ([]Product, error)
	GetMetaInfoForProduct(productName string) (Product, error)
	GetProjectsMetaInfo(productToken string) ([]Project, error)
	GetProjectToken(productToken, projectName string) (string, error)
	GetProjectVitals(projectToken string) (*Project, error)
	GetProjectByName(productToken, projectName string) (*Project, error)
	GetProjectsByIDs(productToken string, projectIDs []int64) ([]Project, error)
	GetProjectTokens(productToken string, projectNames []string) ([]string, error)
	GetProductName(productToken string) (string, error)
	GetProjectRiskReport(projectToken string) ([]byte, error)
	GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error)
	GetOrganizationProductVitals() ([]Product, error)
	GetProductByName(productName string) (*Product, error)
	GetProjectAlerts(projectToken string) ([]Alert, error)
	GetProjectLibraryLocations(projectToken string) ([]Library, error)
}
