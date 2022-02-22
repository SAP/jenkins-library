package format

// SARIF format related JSON structs
type SARIF struct {
	Schema  string `json:"$schema" default:"https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos01/schemas/sarif-schema-2.1.0.json"`
	Version string `json:"version" default:"2.1.0"`
	Runs    []Runs `json:"runs"`
}

// Runs of a Tool and related Results
type Runs struct {
	Results []Results `json:"results"`
	Tool    Tool      `json:"tool"`
	/*Invocations         []Invocations      `json:"invocations"`
	OriginalUriBaseIds  OriginalUriBaseIds `json:"originalUriBaseIds"`
	Artifacts           []Artifact         `json:"artifacts"`
	AutomationDetails   AutomationDetails  `json:"automationDetails"`
	ColumnKind          string             `json:"columnKind" default:"utf16CodeUnits"`
	ThreadFlowLocations []Locations        `json:"threadFlowLocations"`
	Taxonomies          []Taxonomies       `json:"taxonomies"`*/
}

// Results these structs are relevant to the Results object
type Results struct {
	RuleID         string           `json:"ruleId"`
	RuleIndex      int              `json:"ruleIndex"`
	Level          string           `json:"level,omitempty"`
	Message        Message          `json:"message"`
	AnalysisTarget ArtifactLocation `json:"analysisTarget,omitempty"`
	Locations      []Location       `json:"locations"`
	/*CodeFlows        []CodeFlow        `json:"codeFlows"`
	RelatedLocations []RelatedLocation `json:"relatedLocations"`*/
	Properties SarifProperties `json:"properties"`
}

// Message to detail the finding
type Message struct {
	Text string `json:"text,omitempty"`
}

// Location of the finding
type Location struct {
	PhysicalLocation ArtifactLocation  `json:"physicalLocation,omitempty"`
	Region           Region            `json:"region,omitempty"`
	LogicalLocations []LogicalLocation `json:"logicalLocations,omitempty"`
}

// ArtifactLocation describing the path of the artifact
type ArtifactLocation struct {
	URI   string `json:"uri"`
	Index int    `json:"index,omitempty"`
}

// Region where the finding was detected
type Region struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"EndLine,omitempty"`
	EndColumn   int `json:"EndColumn,omitempty"`
	ByteOffset  int `json:"ByteOffset,omitempty"`
	ByteLength  int `json:"ByteLength,omitempty"`
}

// LogicalLocation of the finding
type LogicalLocation struct {
	FullyQualifiedName string `json:"fullyQualifiedName"`
}

// SarifProperties adding additional information/context to the finding
type SarifProperties struct {
	InstanceID        string `json:"InstanceID"`
	InstanceSeverity  string `json:"InstanceSeverity"`
	Confidence        string `json:"Confidence"`
	Audited           bool   `json:"Audited"`
	ToolSeverity      string `json:"ToolSeverity"`
	ToolSeverityIndex int    `json:"ToolSeverityIndex"`
	ToolState         string `json:"ToolState"`
	ToolStateIndex    int    `json:"ToolStateIndex"`
	ToolAuditMessage  string `json:"ToolAuditMessage"`
	UnifiedAuditState string `json:"UnifiedAuditState"`
}

// Tool these structs are relevant to the Tool object
type Tool struct {
	Driver Driver `json:"driver"`
}

// Driver meta information for the scan and tool context
type Driver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationUri string      `json:"informationUri,omitempty"`
	Rules          []SarifRule `json:"rules"`
	//SupportedTaxonomies []SupportedTaxonomies `json:"supportedTaxonomies"`
}

// SarifRule related rule use to identify the finding
type SarifRule struct {
	ID                   string               `json:"id"`
	GUID                 string               `json:"guid"`
	Name                 string               `json:"name,omitempty"`
	ShortDescription     Message              `json:"shortDescription"`
	FullDescription      Message              `json:"fullDescription"`
	DefaultConfiguration DefaultConfiguration `json:"defaultConfiguration"`
	HelpURI              string               `json:"helpUri,omitempty"`
	Help                 Help                 `json:"help,omitempty"`
	Relationships        []Relationships      `json:"relationships,omitempty"`
	Properties           *SarifRuleProperties `json:"properties,omitempty"`
}

// Help provides additional guidance to resolve the finding
type Help struct {
	Text     string `json:"text,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

// SupportedTaxonomies
type SupportedTaxonomies struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	GUID  string `json:"guid"`
}

// DefaultConfiguration
type DefaultConfiguration struct {
	Properties DefaultProperties `json:"properties"`
	Level      string            `json:"level,omitempty"` //This exists in the template, but not sure how it is populated. TODO.
}

//DefaultProperties
type DefaultProperties struct {
	DefaultSeverity string `json:"DefaultSeverity"`
}

// Relationships
type Relationships struct {
	Target Target   `json:"target"`
	Kinds  []string `json:"kinds"`
}

// Target
type Target struct {
	ID            string        `json:"id"`
	ToolComponent ToolComponent `json:"toolComponent"`
}

//ToolComponent
type ToolComponent struct {
	Name string `json:"name"`
	GUID string `json:"guid"`
}

// SarifRuleProperties
type SarifRuleProperties struct {
	Accuracy    string   `json:"Accuracy,omitempty"`
	Impact      string   `json:"Impact,omitempty"`
	Probability string   `json:"Probability,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Precision   string   `json:"precision,omitempty"`
}
