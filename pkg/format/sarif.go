package format

// SARIF format related JSON structs
type SARIF struct {
	Schema  string `json:"$schema" default:"https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json"`
	Version string `json:"version" default:"2.1.0"`
	Runs    []Runs `json:"runs"`
}

// Runs of a Tool and related Results
type Runs struct {
	Results             []Results           `json:"results"`
	Tool                Tool                `json:"tool"`
	Invocations         []Invocations       `json:"invocations,omitempty"`
	OriginalUriBaseIds  *OriginalUriBaseIds `json:"originalUriBaseIds,omitempty"`
	Artifacts           []Artifact          `json:"artifacts,omitempty"`
	AutomationDetails   AutomationDetails   `json:"automationDetails,omitempty"`
	ColumnKind          string              `json:"columnKind,omitempty" default:"utf16CodeUnits"`
	ThreadFlowLocations []Locations         `json:"threadFlowLocations,omitempty"`
	Taxonomies          []Taxonomies        `json:"taxonomies,omitempty"`
}

// Results these structs are relevant to the Results object
type Results struct {
	RuleID           string            `json:"ruleId"`
	RuleIndex        int               `json:"ruleIndex"`
	Level            string            `json:"level,omitempty"`
	Message          *Message          `json:"message,omitempty"`
	AnalysisTarget   *ArtifactLocation `json:"analysisTarget,omitempty"`
	Locations        []Location        `json:"locations,omitempty"`
	CodeFlows        []CodeFlow        `json:"codeFlows,omitempty"`
	RelatedLocations []RelatedLocation `json:"relatedLocations,omitempty"`
	Properties       SarifProperties   `json:"properties"`
}

// Message to detail the finding
type Message struct {
	Text string `json:"text,omitempty"`
}

// Location of the finding
type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
	Message          *Message         `json:"message,omitempty"`
}

// PhysicalLocation
type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation  `json:"artifactLocation"`
	Region           Region            `json:"region,omitempty"`
	ContextRegion    ContextRegion     `json:"contextRegion,omitempty"`
	LogicalLocations []LogicalLocation `json:"logicalLocations,omitempty"`
}

// ArtifactLocation describing the path of the artifact
type ArtifactLocation struct {
	URI   string `json:"uri"`
	Index int    `json:"index"`
}

// Region where the finding was detected
type Region struct {
	StartLine   int           `json:"startLine,omitempty"`
	StartColumn int           `json:"startColumn,omitempty"`
	EndLine     int           `json:"endLine,omitempty"`
	EndColumn   int           `json:"endColumn,omitempty"`
	ByteOffset  int           `json:"byteOffset,omitempty"`
	ByteLength  int           `json:"byteLength,omitempty"`
	Snippet     *SnippetSarif `json:"snippet,omitempty"`
}

// LogicalLocation of the finding
type LogicalLocation struct {
	FullyQualifiedName string `json:"fullyQualifiedName"`
}

// SarifProperties adding additional information/context to the finding
type SarifProperties struct {
	InstanceID        string `json:"instanceID,omitempty"`
	InstanceSeverity  string `json:"instanceSeverity,omitempty"`
	Confidence        string `json:"confidence,omitempty"`
	FortifyCategory   string `json:"fortifyCategory,omitempty"`
	Audited           bool   `json:"audited"`
	ToolSeverity      string `json:"toolSeverity"`
	ToolSeverityIndex int    `json:"toolSeverityIndex"`
	ToolState         string `json:"toolState"`
	ToolStateIndex    int    `json:"toolStateIndex"`
	ToolAuditMessage  string `json:"toolAuditMessage"`
	UnifiedAuditState string `json:"unifiedAuditState"`
}

// Tool these structs are relevant to the Tool object
type Tool struct {
	Driver Driver `json:"driver"`
}

// Driver meta information for the scan and tool context
type Driver struct {
	Name                string                `json:"name"`
	Version             string                `json:"version"`
	InformationUri      string                `json:"informationUri,omitempty"`
	Rules               []SarifRule           `json:"rules"`
	SupportedTaxonomies []SupportedTaxonomies `json:"supportedTaxonomies,omitempty"`
}

// SarifRule related rule use to identify the finding
type SarifRule struct {
	ID                   string                `json:"id"`
	GUID                 string                `json:"guid,omitempty"`
	Name                 string                `json:"name,omitempty"`
	ShortDescription     *Message              `json:"shortDescription,omitempty"`
	FullDescription      *Message              `json:"fullDescription,omitempty"`
	DefaultConfiguration *DefaultConfiguration `json:"defaultConfiguration,omitempty"`
	HelpURI              string                `json:"helpUri,omitempty"`
	Help                 *Help                 `json:"help,omitempty"`
	Relationships        []Relationships       `json:"relationships,omitempty"`
	Properties           *SarifRuleProperties  `json:"properties,omitempty"`
}

// Help provides additional guidance to resolve the finding
type Help struct {
	Text     string `json:"text,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

// SnippetSarif holds the code snippet where the finding appears
type SnippetSarif struct {
	Text string `json:"text"`
}

// ContextRegion provides the context for the finding
type ContextRegion struct {
	StartLine int           `json:"startLine,omitempty"`
	EndLine   int           `json:"endLine,omitempty"`
	Snippet   *SnippetSarif `json:"snippet,omitempty"`
}

// CodeFlow
type CodeFlow struct {
	ThreadFlows []ThreadFlow `json:"threadFlows"`
}

// ThreadFlow
type ThreadFlow struct {
	Locations []Locations `json:"locations"`
}

// Locations
type Locations struct {
	Location *Location `json:"location,omitempty"`
	Kinds    []string  `json:"kinds,omitempty"`
	Index    int       `json:"index,omitempty"`
}

// RelatedLocation
type RelatedLocation struct {
	ID               int                     `json:"id"`
	PhysicalLocation RelatedPhysicalLocation `json:"physicalLocation"`
}

// RelatedPhysicalLocation
type RelatedPhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           RelatedRegion    `json:"region"`
}

// RelatedRegion
type RelatedRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
}

// SupportedTaxonomies
type SupportedTaxonomies struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Guid  string `json:"guid"`
}

// DefaultConfiguration
type DefaultConfiguration struct {
	Properties DefaultProperties `json:"properties,omitempty"`
	Level      string            `json:"level,omitempty"` //This exists in the template, but not sure how it is populated. TODO.
}

// DefaultProperties
type DefaultProperties struct {
	DefaultSeverity string `json:"defaultSeverity,omitempty"`
}

// Relationships
type Relationships struct {
	Target Target   `json:"target"`
	Kinds  []string `json:"kinds"`
}

// Target
type Target struct {
	Id            string        `json:"id"`
	ToolComponent ToolComponent `json:"toolComponent"`
}

// ToolComponent
type ToolComponent struct {
	Name string `json:"name"`
	Guid string `json:"guid"`
}

//SarifRuleProperties
type SarifRuleProperties struct {
	Accuracy    string   `json:"accuracy,omitempty"`
	Impact      string   `json:"impact,omitempty"`
	Probability string   `json:"probability,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Precision   string   `json:"precision,omitempty"`
}

// Invocations These structs are relevant to the Invocations object
type Invocations struct {
	CommandLine                string                       `json:"commandLine"`
	StartTimeUtc               string                       `json:"startTimeUtc"`
	ToolExecutionNotifications []ToolExecutionNotifications `json:"toolExecutionNotifications"`
	ExecutionSuccessful        bool                         `json:"executionSuccessful"`
	Machine                    string                       `json:"machine"`
	Account                    string                       `json:"account"`
	Properties                 InvocationProperties         `json:"properties"`
}

// ToolExecutionNotifications
type ToolExecutionNotifications struct {
	Message    Message    `json:"message"`
	Descriptor Descriptor `json:"descriptor"`
}

// Descriptor
type Descriptor struct {
	Id string `json:"id"`
}

// InvocationProperties
type InvocationProperties struct {
	Platform string `json:"platform"`
}

// OriginalUriBaseIds These structs are relevant to the originalUriBaseIds object
type OriginalUriBaseIds struct {
	SrcRoot SrcRoot `json:"%SRCROOT%"`
}

// SrcRoot
type SrcRoot struct {
	Uri string `json:"uri"`
}

// Artifact These structs are relevant to the artifacts object
type Artifact struct {
	Location SarifLocation `json:"location"`
	Length   int           `json:"length"`
	MimeType string        `json:"mimeType"`
	Encoding string        `json:"encoding"`
}

// SarifLocation
type SarifLocation struct {
	Uri       string `json:"uri"`
	UriBaseId string `json:"uriBaseId"`
}

// AutomationDetails These structs are relevant to the automationDetails object
type AutomationDetails struct {
	Id string `json:"id"`
}

// These structs are relevant to the threadFlowLocations object

// Taxonomies These structs are relevant to the taxonomies object
type Taxonomies struct {
	Guid             string  `json:"guid"`
	Name             string  `json:"name"`
	Organization     string  `json:"organization"`
	ShortDescription Message `json:"shortDescription"`
	Taxa             []Taxa  `json:"taxa"`
}

// Taxa
type Taxa struct {
	Id string `json:"id"`
}
