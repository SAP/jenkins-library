package whitesource

// System declares an interface for talking to the Whitesource service.
type System interface {
	GetProductsMetaInfo() ([]Product, error)
	GetProductByName(productName string) (Product, error)
	GetProjectsMetaInfo(productToken string) ([]Project, error)
	GetProjectToken(productToken, projectName string) (string, error)
	GetProjectByToken(projectToken string) (Project, error)
	GetProjectByName(productToken, projectName string) (Project, error)
	GetProjectTokens(productToken string, projectNames []string) ([]string, error)
	GetProductName(productToken string) (string, error)
	GetProjectRiskReport(projectToken string) ([]byte, error)
	GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error)
	GetProjectAlerts(projectToken string) ([]Alert, error)
	GetProjectLibraryLocations(projectToken string) ([]Library, error)
}
