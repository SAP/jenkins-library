package whitesource

// ScanOptions contains parameters needed during the scan.
type ScanOptions struct {
	// ScanType defines the type of scan. Can be "maven" or "mta" for scanning with Maven or "npm"/"yarn".
	ScanType     string
	OrgToken     string
	UserToken    string
	ProductName  string
	ProductToken string
	// ProjectName is an optional name for an "aggregator" project.
	// All scanned maven modules will be reflected in the aggregate project.
	ProjectName                string
	BuildDescriptorExcludeList []string
	// PomPath is the path to root build descriptor file.
	PomPath string
	// M2Path is the path to the local maven repository.
	M2Path string
	// GlobalSettingsFile is an optional path to a global maven settings file.
	GlobalSettingsFile string
	// ProjectSettingsFile is an optional path to a local maven settings file.
	ProjectSettingsFile string

	// DefaultNpmRegistry is an optional default registry for NPM.
	DefaultNpmRegistry string

	AgentDownloadURL string
	AgentFileName    string
	ConfigFilePath   string

	Includes string
	Excludes string
}
