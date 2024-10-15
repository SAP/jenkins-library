package whitesource

// ScanOptions contains parameters needed during the scan.
type ScanOptions struct {
	BuildTool string
	// ScanType defines the type of scan. Can be "maven" or "mta" for scanning with Maven or "npm"/"yarn".
	ScanType       string
	OrgToken       string
	UserToken      string
	ProductName    string
	ProductToken   string
	ProductVersion string
	// ProjectName is an optional name for an "aggregator" project.
	// All scanned maven modules will be reflected in the aggregate project.
	ProjectName string

	BuildDescriptorFile        string
	BuildDescriptorExcludeList []string
	// PomPath is the path to root build descriptor file.
	PomPath string
	// M2Path is the path to the local maven repository.
	M2Path string
	// GlobalSettingsFile is an optional path to a global maven settings file.
	GlobalSettingsFile string
	// ProjectSettingsFile is an optional path to a local maven settings file.
	ProjectSettingsFile string
	// InstallArtifacts installs artifacts from all maven modules to the local repository
	InstallArtifacts bool

	// DefaultNpmRegistry is an optional default registry for NPM.
	DefaultNpmRegistry        string
	NpmIncludeDevDependencies bool

	AgentDownloadURL       string
	AgentFileName          string
	ConfigFilePath         string
	UseGlobalConfiguration bool

	JreDownloadURL string

	Includes []string
	Excludes []string

	AgentURL   string
	ServiceURL string

	ScanPath string

	InstallCommand string

	SkipParentProjectResolution     bool
	DisableNpmSubmodulesAggregation bool

	Verbose bool
}
