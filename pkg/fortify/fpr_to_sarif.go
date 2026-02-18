package fortify

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/piper-validation/fortify-client-go/models"

	"github.com/SAP/jenkins-library/pkg/format"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// FVDL This struct encapsulates everyting in the FVDL document
type FVDL struct {
	XMLName         xml.Name        `xml:"FVDL"`
	Xmlns           string          `xml:"xmlns,attr"`
	XmlnsXsi        string          `xml:"xsi,attr"`
	Version         string          `xml:"version,attr"`
	XsiType         string          `xml:"type,attr"`
	Created         CreatedTS       `xml:"CreatedTS"`
	Uuid            UUID            `xml:"UUID"`
	Build           Build           `xml:"Build"`
	Vulnerabilities Vulnerabilities `xml:"Vulnerabilities"`
	ContextPool     ContextPool     `xml:"ContextPool"`
	UnifiedNodePool UnifiedNodePool `xml:"UnifiedNodePool"`
	Description     []Description   `xml:"Description"`
	Snippets        []Snippet       `xml:"Snippets>Snippet"`
	ProgramData     ProgramData     `xml:"ProgramData"`
	EngineData      EngineData      `xml:"EngineData"`
}

// CreatedTS
type CreatedTS struct {
	XMLName xml.Name `xml:"CreatedTS"`
	Date    string   `xml:"date,attr"`
	Time    string   `xml:"time,attr"`
}

// UUIF
type UUID struct {
	XMLName xml.Name `xml:"UUID"`
	Uuid    string   `xml:",innerxml"`
}

// LOC
type LOC struct {
	XMLName  xml.Name `xml:"LOC"`
	LocType  string   `xml:"type,attr"`
	LocValue string   `xml:",innerxml"`
}

// These structures are relevant to the Build object
// The Build object transports all build and scan related information
type Build struct {
	XMLName        xml.Name `xml:"Build"`
	Project        string   `xml:"Project"`
	Version        string   `xml:"Version"`
	Label          string   `xml:"Label"`
	BuildID        string   `xml:"BuildID"`
	NumberFiles    int      `xml:"NumberFiles"`
	Locs           []LOC    `xml:"LOC"`
	JavaClassPath  string   `xml:"JavaClasspath"`
	SourceBasePath string   `xml:"SourceBasePath"`
	SourceFiles    []File   `xml:"SourceFiles>File"`
	Scantime       ScanTime `xml:"ScanTime"`
}

// File
type File struct {
	XMLName       xml.Name `xml:"File"`
	FileSize      int      `xml:"size,attr"`
	FileTimestamp string   `xml:"timestamp,attr"`
	FileLoc       int      `xml:"loc,attr,omitempty"`
	FileType      string   `xml:"type,attr"`
	Encoding      string   `xml:"encoding,attr"`
	Name          string   `xml:"Name"`
	Locs          []LOC    `xml:",any,omitempty"`
}

// ScanTime
type ScanTime struct {
	XMLName xml.Name `xml:"ScanTime"`
	Value   int      `xml:"value,attr"`
}

// Vulnerabilities These structures are relevant to the Vulnerabilities object
type Vulnerabilities struct {
	XMLName       xml.Name        `xml:"Vulnerabilities"`
	Vulnerability []Vulnerability `xml:"Vulnerability"`
}

// Vulnerability
type Vulnerability struct {
	XMLName      xml.Name     `xml:"Vulnerability"`
	ClassInfo    ClassInfo    `xml:"ClassInfo"`
	InstanceInfo InstanceInfo `xml:"InstanceInfo"`
	AnalysisInfo AnalysisInfo `xml:"AnalysisInfo>Unified"`
}

// ClassInfo
type ClassInfo struct {
	XMLName         xml.Name `xml:"ClassInfo"`
	ClassID         string   `xml:"ClassID"`
	Kingdom         string   `xml:"Kingdom,omitempty"`
	Type            string   `xml:"Type"`
	Subtype         string   `xml:"Subtype,omitempty"`
	AnalyzerName    string   `xml:"AnalyzerName"`
	DefaultSeverity float64  `xml:"DefaultSeverity"`
}

// InstanceInfo
type InstanceInfo struct {
	XMLName          xml.Name `xml:"InstanceInfo"`
	InstanceID       string   `xml:"InstanceID"`
	InstanceSeverity float64  `xml:"InstanceSeverity"`
	Confidence       string   `xml:"Confidence"`
}

// AnalysisInfo
type AnalysisInfo struct { //Note that this is directly the "Unified" object
	Context                Context
	ReplacementDefinitions ReplacementDefinitions `xml:"ReplacementDefinitions"`
	Trace                  []Trace                `xml:"Trace"`
}

// Context
type Context struct {
	XMLName   xml.Name                          `xml:"Context"`
	ContextId string                            `xml:"id,attr,omitempty"`
	Function  Function                          `xml:"Function"`
	FDSL      FunctionDeclarationSourceLocation `xml:"FunctionDeclarationSourceLocation"`
}

// Function
type Function struct {
	XMLName                xml.Name `xml:"Function"`
	FunctionName           string   `xml:"name,attr"`
	FunctionNamespace      string   `xml:"namespace,attr"`
	FunctionEnclosingClass string   `xml:"enclosingClass,attr"`
}

// FunctionDeclarationSourceLocation
type FunctionDeclarationSourceLocation struct {
	XMLName      xml.Name `xml:"FunctionDeclarationSourceLocation"`
	FDSLPath     string   `xml:"path,attr"`
	FDSLLine     string   `xml:"line,attr"`
	FDSLLineEnd  string   `xml:"lineEnd,attr"`
	FDSLColStart string   `xml:"colStart,attr"`
	FDSLColEnd   string   `xml:"colEnd,attr"`
}

// ReplacementDefinitions
type ReplacementDefinitions struct {
	XMLName     xml.Name      `xml:"ReplacementDefinitions"`
	Def         []Def         `xml:"Def"`
	LocationDef []LocationDef `xml:"LocationDef"`
}

// Def
type Def struct {
	XMLName  xml.Name `xml:"Def"`
	DefKey   string   `xml:"key,attr"`
	DefValue string   `xml:"value,attr"`
}

// LocationDef
type LocationDef struct {
	XMLName  xml.Name `xml:"LocationDef"`
	Path     string   `xml:"path,attr"`
	Line     int      `xml:"line,attr"`
	LineEnd  int      `xml:"lineEnd,attr"`
	ColStart int      `xml:"colStart,attr"`
	ColEnd   int      `xml:"colEnd,attr"`
	Key      string   `xml:"key,attr"`
}

// Trace
type Trace struct {
	XMLName xml.Name `xml:"Trace"`
	Primary Primary  `xml:"Primary"`
}

// Primary
type Primary struct {
	XMLName xml.Name `xml:"Primary"`
	Entry   []Entry  `xml:"Entry"`
}

// Entry
type Entry struct {
	XMLName xml.Name `xml:"Entry"`
	NodeRef NodeRef  `xml:"NodeRef,omitempty"`
	Node    Node     `xml:"Node,omitempty"`
}

// NodeRef
type NodeRef struct {
	XMLName xml.Name `xml:"NodeRef"`
	RefId   int      `xml:"id,attr"`
}

// Node
type Node struct {
	XMLName        xml.Name       `xml:"Node"`
	IsDefault      string         `xml:"isDefault,attr,omitempty"`
	NodeLabel      string         `xml:"label,attr,omitempty"`
	ID             int            `xml:"id,attr,omitempty"`
	SourceLocation SourceLocation `xml:"SourceLocation"`
	Action         Action         `xml:"Action,omitempty"`
	Reason         Reason         `xml:"Reason,omitempty"`
	Knowledge      Knowledge      `xml:"Knowledge,omitempty"`
}

// SourceLocation
type SourceLocation struct {
	XMLName   xml.Name `xml:"SourceLocation"`
	Path      string   `xml:"path,attr"`
	Line      int      `xml:"line,attr"`
	LineEnd   int      `xml:"lineEnd,attr"`
	ColStart  int      `xml:"colStart,attr"`
	ColEnd    int      `xml:"colEnd,attr"`
	ContextId string   `xml:"contextId,attr"`
	Snippet   string   `xml:"snippet,attr"`
}

// Action
type Action struct {
	XMLName    xml.Name `xml:"Action"`
	Type       string   `xml:"type,attr"`
	ActionData string   `xml:",innerxml"`
}

// Reason
type Reason struct {
	XMLName xml.Name `xml:"Reason"`
	Rule    Rule     `xml:"Rule,omitempty"`
	Trace   Trace    `xml:"Trace,omitempty"`
}

// Rule
type Rule struct {
	XMLName xml.Name `xml:"Rule"`
	RuleID  string   `xml:"ruleID,attr"`
}

// Group
type Group struct {
	XMLName xml.Name `xml:"Group"`
	Name    string   `xml:"name,attr"`
	Data    string   `xml:",innerxml"`
}

// Knowledge
type Knowledge struct {
	XMLName xml.Name `xml:"Knowledge"`
	Facts   []Fact   `xml:"Fact"`
}

// Fact
type Fact struct {
	XMLName  xml.Name `xml:"Fact"`
	Primary  string   `xml:"primary,attr"`
	Type     string   `xml:"type,attr,omitempty"`
	FactData string   `xml:",innerxml"`
}

// ContextPool These structures are relevant to the ContextPool object
type ContextPool struct {
	XMLName xml.Name  `xml:"ContextPool"`
	Context []Context `xml:"Context"`
}

// UnifiedNodePool These structures are relevant to the UnifiedNodePool object
type UnifiedNodePool struct {
	XMLName xml.Name `xml:"UnifiedNodePool"`
	Node    []Node   `xml:"Node"`
}

// Description These structures are relevant to the Description object
type Description struct {
	XMLName           xml.Name          `xml:"Description"`
	ContentType       string            `xml:"contentType,attr"`
	ClassID           string            `xml:"classID,attr"`
	Abstract          Abstract          `xml:"Abstract"`
	Explanation       Explanation       `xml:"Explanation"`
	Recommendations   Recommendations   `xml:"Recommendations"`
	Tips              []Tip             `xml:"Tips>Tip,omitempty"`
	References        []Reference       `xml:"References>Reference"`
	CustomDescription CustomDescription `xml:"CustomDescription,omitempty"`
}

// Abstract
type Abstract struct {
	XMLName xml.Name `xml:"Abstract"`
	Text    string   `xml:",innerxml"`
}

// Explanation
type Explanation struct {
	XMLName xml.Name `xml:"Explanation"`
	Text    string   `xml:",innerxml"`
}

// Recommendations
type Recommendations struct {
	XMLName xml.Name `xml:"Recommendations"`
	Text    string   `xml:",innerxml"`
}

// Reference
type Reference struct {
	XMLName xml.Name `xml:"Reference"`
	Title   string   `xml:"Title"`
	Author  string   `xml:"Author"`
}

// Tip
type Tip struct {
	XMLName xml.Name `xml:"Tip"`
	Tip     string   `xml:",innerxml"`
}

// CustomDescription
type CustomDescription struct {
	XMLName         xml.Name        `xml:"CustomDescription"`
	ContentType     string          `xml:"contentType,attr"`
	RuleID          string          `xml:"ruleID,attr"`
	Explanation     Explanation     `xml:"Explanation"`
	Recommendations Recommendations `xml:"Recommendations"`
	References      []Reference     `xml:"References>Reference"`
}

// Snippet These structures are relevant to the Snippets object
type Snippet struct {
	XMLName   xml.Name `xml:"Snippet"`
	SnippetId string   `xml:"id,attr"`
	File      string   `xml:"File"`
	StartLine int      `xml:"StartLine"`
	EndLine   int      `xml:"EndLine"`
	Text      string   `xml:"Text"`
}

// ProgramData These structures are relevant to the ProgramData object
type ProgramData struct {
	XMLName         xml.Name         `xml:"ProgramData"`
	Sources         []SourceInstance `xml:"Sources>SourceInstance"`
	Sinks           []SinkInstance   `xml:"Sinks>SinkInstance"`
	CalledWithNoDef []Function       `xml:"CalledWithNoDef>Function"`
}

// SourceInstance
type SourceInstance struct {
	XMLName        xml.Name       `xml:"SourceInstance"`
	RuleID         string         `xml:"ruleID,attr"`
	FunctionCall   FunctionCall   `xml:"FunctionCall,omitempty"`
	FunctionEntry  FunctionEntry  `xml:"FunctionEntry,omitempty"`
	SourceLocation SourceLocation `xml:"SourceLocation,omitempty"`
	TaintFlags     TaintFlags     `xml:"TaintFlags"`
}

// FunctionCall
type FunctionCall struct {
	XMLName        xml.Name       `xml:"FunctionCall"`
	SourceLocation SourceLocation `xml:"SourceLocation"`
	Function       Function       `xml:"Function"`
}

// FunctionEntry
type FunctionEntry struct {
	XMLName        xml.Name       `xml:"FunctionEntry"`
	SourceLocation SourceLocation `xml:"SourceLocation"`
	Function       Function       `xml:"Function"`
}

// TaintFlags
type TaintFlags struct {
	XMLName   xml.Name    `xml:"TaintFlags"`
	TaintFlag []TaintFlag `xml:"TaintFlag"`
}

// TaintFlag
type TaintFlag struct {
	XMLName       xml.Name `xml:"TaintFlag"`
	TaintFlagName string   `xml:"name,attr"`
}

// SinkInstance
type SinkInstance struct {
	XMLName        xml.Name       `xml:"SinkInstance"`
	RuleID         string         `xml:"ruleID,attr"`
	FunctionCall   FunctionCall   `xml:"FunctionCall,omitempty"`
	SourceLocation SourceLocation `xml:"SourceLocation,omitempty"`
}

// EngineData These structures are relevant to the EngineData object
type EngineData struct {
	XMLName       xml.Name     `xml:"EngineData"`
	EngineVersion string       `xml:"EngineVersion"`
	RulePacks     []RulePack   `xml:"RulePacks>RulePack"`
	Properties    []Properties `xml:"Properties"`
	CLArguments   []string     `xml:"CommandLine>Argument"`
	Errors        []Error      `xml:"Errors>Error"`
	MachineInfo   MachineInfo  `xml:"MachineInfo"`
	FilterResult  FilterResult `xml:"FilterResult"`
	RuleInfo      []RuleInfo   `xml:"RuleInfo>Rule"`
	LicenseInfo   LicenseInfo  `xml:"LicenseInfo"`
}

// RulePack
type RulePack struct {
	XMLName    xml.Name `xml:"RulePack"`
	RulePackID string   `xml:"RulePackID"`
	SKU        string   `xml:"SKU"`
	Name       string   `xml:"Name"`
	Version    string   `xml:"Version"`
	MAC        string   `xml:"MAC"`
}

// Properties
type Properties struct {
	XMLName        xml.Name   `xml:"Properties"`
	PropertiesType string     `xml:"type,attr"`
	Property       []Property `xml:"Property"`
}

// Property
type Property struct {
	XMLName xml.Name `xml:"Property"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

// Error
type Error struct {
	XMLName      xml.Name `xml:"Error"`
	ErrorCode    string   `xml:"code,attr"`
	ErrorMessage string   `xml:",innerxml"`
}

// MachineInfo
type MachineInfo struct {
	XMLName  xml.Name `xml:"MachineInfo"`
	Hostname string   `xml:"Hostname"`
	Username string   `xml:"Username"`
	Platform string   `xml:"Platform"`
}

// FilterResult
type FilterResult struct {
	XMLName xml.Name `xml:"FilterResult"`
	//Todo? No data in sample audit file
}

// RuleInfo
type RuleInfo struct {
	XMLName       xml.Name `xml:"Rule"`
	RuleID        string   `xml:"id,attr"`
	MetaInfoGroup []Group  `xml:"MetaInfo>Group,omitempty"`
}

// LicenseInfo
type LicenseInfo struct {
	XMLName    xml.Name     `xml:"LicenseInfo"`
	Metadata   []Metadata   `xml:"Metadata"`
	Capability []Capability `xml:"Capability"`
}

// Metadata
type Metadata struct {
	XMLName xml.Name `xml:"Metadata"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

// Capability
type Capability struct {
	XMLName    xml.Name  `xml:"Capability"`
	Name       string    `xml:"Name"`
	Expiration string    `xml:"Expiration"`
	Attribute  Attribute `xml:"Attribute"`
}

// Attribute
type Attribute struct {
	XMLName xml.Name `xml:"Attribute"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

// ConvertFprToSarif converts the FPR file contents into SARIF format
func ConvertFprToSarif(sys System, projectVersion *models.ProjectVersion, resultFilePath string, filterSet *models.FilterSet) (format.SARIF, format.SARIF, error) {
	log.Entry().Debug("Extracting FPR.")
	var sarif format.SARIF
	var sarifSimplified format.SARIF
	tmpFolder, err := os.MkdirTemp(".", "temp-")
	defer os.RemoveAll(tmpFolder)
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
		return sarif, sarifSimplified, err
	}

	_, err = piperutils.Unzip(resultFilePath, tmpFolder)
	if err != nil {
		return sarif, sarifSimplified, err
	}

	log.Entry().Debug("Reading audit file.")
	data, err := os.ReadFile(filepath.Join(tmpFolder, "audit.fvdl"))
	if err != nil {
		return sarif, sarifSimplified, err
	}
	if len(data) == 0 {
		log.Entry().Error("Error reading audit file at " + filepath.Join(tmpFolder, "audit.fvdl") + ". This might be that the file is missing, corrupted, or too large. Aborting procedure.")
		err := errors.New("cannot read audit file")
		return sarif, sarifSimplified, err
	}

	log.Entry().Debug("Calling Parse.")
	return Parse(sys, projectVersion, data, filterSet)
}

// Parse parses the FPR file
func Parse(sys System, projectVersion *models.ProjectVersion, data []byte, filterSet *models.FilterSet) (format.SARIF, format.SARIF, error) {
	//To read XML data, Unmarshal or Decode can be used, here we use Decode to work on the stream
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	start := time.Now() // For the conversion start time

	var fvdl FVDL
	err := decoder.Decode(&fvdl)
	if err != nil {
		return format.SARIF{}, format.SARIF{}, err
	}

	//Create an object containing all audit data
	log.Entry().Debug("Querying Fortify SSC for batch audit data")
	oneRequestPerIssueMode := false
	var auditData []*models.ProjectVersionIssue
	maxretries := 5 // Maximum number of requests allowed to fail before stopping them
	if sys != nil && projectVersion != nil {
		auditData, err = sys.GetAllIssueDetails(projectVersion.ID)
		if err != nil || len(auditData) == 0 { // It's reasonable to admit that with a length of 0, something went wrong
			log.Entry().WithError(err).Error("failed to get all audit data, defaulting to one-request-per-issue basis")
			oneRequestPerIssueMode = true
			// We do not lower maxretries here in case a "real" bug happened
		} else {
			log.Entry().Debug("Request successful, data frame size: ", len(auditData), " audits")
		}
	} else {
		log.Entry().Error("no system instance or project version found, lookup impossible")
		oneRequestPerIssueMode = true
		maxretries = 1 // Set to 1 if the sys instance isn't defined: chances are it couldn't be created, we'll live a chance if there was an unknown bug
		log.Entry().Debug("request failed: remaining retries ", maxretries)
	}

	//Now, we handle the sarif
	var sarif format.SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var fortifyRun format.Runs
	fortifyRun.ColumnKind = "utf16CodeUnits"
	cweIdsForTaxonomies := make(map[string]string) //Defining this here and filling it in the course of the program helps filling the Taxonomies object easily. Map because easy to check for keys
	sarif.Runs = append(sarif.Runs, fortifyRun)

	// Initialize the simplified version
	var sarifSimplified format.SARIF
	sarifSimplified.Schema = sarif.Schema
	sarifSimplified.Version = sarif.Version
	sarifSimplified.Runs = append(sarifSimplified.Runs, fortifyRun)

	// Handle results/vulnerabilities
	log.Entry().Debug("[SARIF] Now handling results.")
	for i := 0; i < len(fvdl.Vulnerabilities.Vulnerability); i++ {
		result := *new(format.Results)
		//result.RuleID = fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.ClassID
		// Handle ruleID the same way than in Rule
		idArray := []string{}
		/*if fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.Kingdom != "" {
			idArray = append(idArray, fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.Kingdom)
		}*/
		if fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.Type != "" {
			idArray = append(idArray, fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.Type)
		}
		if fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.Subtype != "" {
			idArray = append(idArray, fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.Subtype)
		}
		result.RuleID = "fortify-" + strings.Join(idArray, "/")
		result.Kind = "fail" // Default value, Level must not be set if kind is not fail
		// This is an "easy" treatment of result.Level. It does not follow the spec exactly, but the idea is there
		// An exact processing algorithm can be found here https://docs.oasis-open.org/sarif/sarif/v2.1.0/os/sarif-v2.1.0-os.html#_Toc34317648
		if fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceSeverity >= 3.0 {
			result.Level = "error"
		} else if fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceSeverity >= 1.5 {
			result.Level = "warning"
		} else if fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceSeverity < 1.5 {
			result.Level = "note"
		} else {
			result.Level = "none"
		}
		//get message
		for j := 0; j < len(fvdl.Description); j++ {
			if fvdl.Description[j].ClassID == fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.ClassID {
				result.RuleIndex = j
				rawMessage := unescapeXML(fvdl.Description[j].Abstract.Text)
				// Replacement defintions in message
				for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def); l++ {
					rawMessage = strings.ReplaceAll(rawMessage, "<Replace key=\""+fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def[l].DefKey+"\"/>", fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def[l].DefValue)
				}
				msg := new(format.Message)
				msg.Text = rawMessage
				result.Message = msg
				break
			}
		}

		// Handle all locations items
		location := *new(format.Location)
		//get location
		for k := 0; k < len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace); k++ { // k iterates on traces
			//In each trace/primary, there can be one or more entries
			//Each trace represents a codeflow, each entry represents a location in threadflow
			codeFlow := *new(format.CodeFlow)
			threadFlow := *new(format.ThreadFlow)
			//We now iterate on Entries in the trace/primary
			for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry); l++ { // l iterates on entries
				tfla := *new([]format.Locations)             //threadflowlocationarray. Useful for the node-in-node edge case
				threadFlowLocation := *new(format.Locations) //One is created regardless of the path taken afterwards
				//this will populate both threadFlowLocation AND the parent location object (result.Locations[0])
				// We check if a noderef is present: if no (index of ref is the default 0), this is a "real" node. As a measure of safety (in case a node refers to nodeid 0), we add another check: the node must have a label or a isdefault value
				if fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].NodeRef.RefId == 0 && (fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.NodeLabel != "" || fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.IsDefault != "") {
					//initalize the current location object, it will be added to threadFlowLocation.Location
					tfloc := new(format.Location)
					//get artifact location
					for j := 0; j < len(fvdl.Build.SourceFiles); j++ { // j iterates on source files
						if fvdl.Build.SourceFiles[j].Name == fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.Path {
							tfloc.PhysicalLocation.ArtifactLocation.Index = j + 1
							tfloc.PhysicalLocation.ArtifactLocation.URI = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.Path
							tfloc.PhysicalLocation.ArtifactLocation.URIBaseId = "%SRCROOT%"
							break
						}
					}
					//get region & context region
					tfloc.PhysicalLocation.Region.StartLine = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.Line
					tfloc.PhysicalLocation.Region.EndLine = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.LineEnd
					tfloc.PhysicalLocation.Region.StartColumn = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.ColStart
					tfloc.PhysicalLocation.Region.EndColumn = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.ColEnd
					//Snippet is handled last
					targetSnippetId := fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.Snippet
					for j := 0; j < len(fvdl.Snippets); j++ {
						if fvdl.Snippets[j].SnippetId == targetSnippetId {
							tfloc.PhysicalLocation.ContextRegion = new(format.ContextRegion)
							tfloc.PhysicalLocation.ContextRegion.StartLine = fvdl.Snippets[j].StartLine
							tfloc.PhysicalLocation.ContextRegion.EndLine = fvdl.Snippets[j].EndLine
							break
						}
					}
					// if a label is passed, put it as message
					if fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.NodeLabel != "" {
						tfloc.Message = new(format.Message)
						tfloc.Message.Text = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.NodeLabel
					} else {
						// otherwise check for existance of action object, and if yes, save message
						if !(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData == "") {
							tfloc.Message = new(format.Message)
							tfloc.Message.Text = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData
						}
					}
					location = *tfloc
					//set Kinds
					threadFlowLocation.Location = tfloc
					//threadFlowLocation.Kinds = append(threadFlowLocation.Kinds, "review") //TODO
					threadFlowLocation.Index = 0 // to be safe?
					tfla = append(tfla, threadFlowLocation)

					// "Node-in-node" edge case! in some cases the "Reason" object will contain a "Trace>Primary>Entry>Node" object
					// Check for it at depth 1 only, as an in-case
					if len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry) > 0 {
						ninThreadFlowLocation := *new(format.Locations)
						if fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].NodeRef.RefId != 0 {
							// As usual, only the index for a ref
							ninThreadFlowLocation.Index = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].NodeRef.RefId + 1
						} else {
							// Build a new "node-in-node" tfloc, it will be appended to tfla
							nintfloc := new(format.Location)

							// artifactlocation
							for j := 0; j < len(fvdl.Build.SourceFiles); j++ {
								if fvdl.Build.SourceFiles[j].Name == fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.SourceLocation.Path {
									nintfloc.PhysicalLocation.ArtifactLocation.Index = j + 1
									nintfloc.PhysicalLocation.ArtifactLocation.URI = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.SourceLocation.Path
									nintfloc.PhysicalLocation.ArtifactLocation.URIBaseId = "%SRCROOT%"
									break
								}
							}

							// region & context region
							nintfloc.PhysicalLocation.Region.StartLine = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.SourceLocation.Line
							nintfloc.PhysicalLocation.Region.EndLine = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.SourceLocation.LineEnd
							nintfloc.PhysicalLocation.Region.StartColumn = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.SourceLocation.ColStart
							nintfloc.PhysicalLocation.Region.EndColumn = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.SourceLocation.ColEnd
							// snippet
							targetSnippetId := fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.SourceLocation.Snippet
							for j := 0; j < len(fvdl.Snippets); j++ {
								if fvdl.Snippets[j].SnippetId == targetSnippetId {
									nintfloc.PhysicalLocation.ContextRegion = new(format.ContextRegion)
									nintfloc.PhysicalLocation.ContextRegion.StartLine = fvdl.Snippets[j].StartLine
									nintfloc.PhysicalLocation.ContextRegion.EndLine = fvdl.Snippets[j].EndLine
									break
								}
							}
							// label as message
							if fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.NodeLabel != "" {
								nintfloc.Message = new(format.Message)
								nintfloc.Message.Text = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Reason.Trace.Primary.Entry[0].Node.NodeLabel
							}

							ninThreadFlowLocation.Location = nintfloc
							ninThreadFlowLocation.Index = 0 // Safety
						}
						tfla = append(tfla, ninThreadFlowLocation)
					}
					// END edge case

				} else { //is not a main threadflow: just register NodeRef index in threadFlowLocation
					// Sarif does not provision 0 as a valid array index, so we increment the node ref id
					// Each index i serves to reference the i-th object in run.threadFlowLocations
					threadFlowLocation.Index = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].NodeRef.RefId + 1
					tfla = append(tfla, threadFlowLocation)
				}
				threadFlow.Locations = append(threadFlow.Locations, tfla...)
			}
			codeFlow.ThreadFlows = append(codeFlow.ThreadFlows, threadFlow)
			result.CodeFlows = append(result.CodeFlows, codeFlow)
		}

		//For some reason, the principal object only has 1 location: here we keep the last one
		//Void message
		location.Message = nil
		result.Locations = append(result.Locations, location)

		//handle relatedLocation
		relatedLocation := *new(format.RelatedLocation)
		relatedLocation.ID = 1
		relatedLocation.PhysicalLocation = *new(format.RelatedPhysicalLocation)
		relatedLocation.PhysicalLocation.ArtifactLocation = location.PhysicalLocation.ArtifactLocation
		relatedLocation.PhysicalLocation.Region = *new(format.RelatedRegion)
		relatedLocation.PhysicalLocation.Region.StartLine = location.PhysicalLocation.Region.StartLine
		relatedLocation.PhysicalLocation.Region.StartColumn = location.PhysicalLocation.Region.StartColumn
		result.RelatedLocations = append(result.RelatedLocations, relatedLocation)

		//handle partialFingerprints
		result.PartialFingerprints.FortifyInstanceID = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceID
		result.PartialFingerprints.PrimaryLocationLineHash = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceID //Fixit

		//handle properties
		prop := new(format.SarifProperties)
		prop.InstanceSeverity = strconv.FormatFloat(fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceSeverity, 'f', 1, 64)
		prop.Confidence = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.Confidence
		prop.InstanceID = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceID
		prop.RuleGUID = fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.ClassID
		//Get the audit data
		if err := integrateAuditData(prop, fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceID, sys, projectVersion, auditData, filterSet, oneRequestPerIssueMode, maxretries); err != nil {
			log.Entry().Debug(err)
			maxretries = maxretries - 1
			if maxretries >= 0 {
				log.Entry().Debug("request failed: remaining retries ", maxretries)
			}
		}
		result.Properties = prop

		sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)

		// Handle simplified version of a result
		resultSimplified := *new(format.Results)
		resultSimplified.RuleID = result.RuleID
		resultSimplified.Kind = result.Kind
		resultSimplified.Level = result.Level
		resultSimplified.Message = result.Message
		resultSimplified.Properties = result.Properties
		sarifSimplified.Runs[0].Results = append(sarifSimplified.Runs[0].Results, resultSimplified)
	}

	//handle the tool object
	log.Entry().Debug("[SARIF] Now handling driver object.")
	tool := *new(format.Tool)
	tool.Driver = *new(format.Driver)
	tool.Driver.Name = "MicroFocus Fortify SCA"
	tool.Driver.Version = fvdl.EngineData.EngineVersion
	tool.Driver.InformationUri = "https://www.microfocus.com/documentation/fortify-static-code-analyzer-and-tools/2020/SCA_Guide_20.2.0.pdf"

	//handle the simplified tool object
	toolSimplified := *new(format.Tool)
	toolSimplified.Driver = tool.Driver

	//handles rules
	for i := 0; i < len(fvdl.EngineData.RuleInfo); i++ { //i iterates on rules
		sarifRule := *new(format.SarifRule)
		sarifRule.ID = fvdl.EngineData.RuleInfo[i].RuleID
		sarifRule.GUID = fvdl.EngineData.RuleInfo[i].RuleID
		for j := 0; j < len(fvdl.Vulnerabilities.Vulnerability); j++ { //j iterates on vulns to find the name
			if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.ClassID == fvdl.EngineData.RuleInfo[i].RuleID {
				var nameArray []string
				var idArray []string
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Kingdom != "" {
					//idArray = append(idArray, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Kingdom)
					words := strings.Split(fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Kingdom, " ")
					for index, element := range words { // These are required to ensure that titlecase is respected in titles, part of sarif "friendly name" rules
						words[index] = piperutils.Title(strings.ToLower(element))
					}
					nameArray = append(nameArray, words...)
				}
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Type != "" {
					idArray = append(idArray, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Type)
					words := strings.Split(fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Type, " ")
					for index, element := range words {
						words[index] = piperutils.Title(strings.ToLower(element))
					}
					nameArray = append(nameArray, words...)
				}
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Subtype != "" {
					idArray = append(idArray, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Subtype)
					words := strings.Split(fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Subtype, " ")
					for index, element := range words {
						words[index] = piperutils.Title(strings.ToLower(element))
					}
					nameArray = append(nameArray, words...)
				}
				sarifRule.ID = "fortify-" + strings.Join(idArray, "/")
				sarifRule.Name = strings.Join(nameArray, "")
				defaultConfig := new(format.DefaultConfiguration)
				defaultConfig.Level = "warning" // Default value
				defaultConfig.Enabled = true    // Default value
				defaultConfig.Rank = -1.0       // Default value
				defaultConfig.Properties.DefaultSeverity = strconv.FormatFloat(fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.DefaultSeverity, 'f', 1, 64)
				sarifRule.DefaultConfiguration = defaultConfig

				//Descriptions
				for j := 0; j < len(fvdl.Description); j++ {
					if fvdl.Description[j].ClassID == sarifRule.GUID {
						//rawAbstract := strings.Join(idArray, "/")
						rawAbstract := unescapeXML(fvdl.Description[j].Abstract.Text)
						rawExplanation := unescapeXML(fvdl.Description[j].Explanation.Text)

						// Replacement defintions in abstract/explanation
						for k := 0; k < len(fvdl.Vulnerabilities.Vulnerability); k++ { // Iterate on vulns to find the correct one (where ReplacementDefinitions are)
							if fvdl.Vulnerabilities.Vulnerability[k].ClassInfo.ClassID == fvdl.Description[j].ClassID {
								for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def); l++ {
									rawAbstract = strings.ReplaceAll(rawAbstract, "<Replace key=\""+fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def[l].DefKey+"\"/>", fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def[l].DefValue)
									rawExplanation = strings.ReplaceAll(rawExplanation, "<Replace key=\""+fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def[l].DefKey+"\"/>", fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def[l].DefValue)
								}
								// Replacement locationdef in explanation
								for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.LocationDef); l++ {
									rawExplanation = strings.ReplaceAll(rawExplanation, fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.LocationDef[l].Key, fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.LocationDef[l].Path)
								}
								// If Description has a CustomDescription, add it for good measure
								if fvdl.Description[j].CustomDescription.RuleID != "" {
									rawExplanation = rawExplanation + " \n; " + unescapeXML(fvdl.Description[j].CustomDescription.Explanation.Text)
								}
								sd := new(format.Message)
								sd.Text = rawAbstract
								sarifRule.ShortDescription = sd
								fd := new(format.Message)
								fd.Text = rawExplanation
								sarifRule.FullDescription = fd
								break
							}
						}
						break
					}
				}

				//properties
				//Prepare a CWE id object as an in-case
				cweIds := []string{}
				//scan for the properties we want:
				var propArray [][]string
				for j := 0; j < len(fvdl.EngineData.RuleInfo[i].MetaInfoGroup); j++ {
					if (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Accuracy") || (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Impact") || (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Probability") {
						propArray = append(propArray, []string{fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name, fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Data})
					} else if fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "altcategoryCWE" {
						//Get all CWE IDs. First, split on ", "
						rawCweIds := strings.Split(fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Data, ", ")
						//If not "None", split each string on " " and add its 2nd index
						if rawCweIds[0] != "None" {
							for k := range rawCweIds {
								cweId := strings.Split(rawCweIds[k], " ")[2]
								//Fill the cweIdsForTaxonomies map if not already in
								if _, isIn := cweIdsForTaxonomies[cweId]; !isIn {
									cweIdsForTaxonomies[cweId] = cweId
								}
								cweIds = append(cweIds, cweId)
							}
						} else {
							cweIds = append(cweIds, rawCweIds[0])
						}
					}
				}
				var ruleProp *format.SarifRuleProperties
				ruleProp = new(format.SarifRuleProperties)
				if len(propArray) != 0 {
					for j := 0; j < len(propArray); j++ {
						if propArray[j][0] == "Accuracy" {
							ruleProp.Accuracy = propArray[j][1]
						} else if propArray[j][0] == "Impact" {
							ruleProp.Impact = propArray[j][1]
						} else if propArray[j][0] == "Probability" {
							ruleProp.Probability = propArray[j][1]
						}
					}
				}

				// Add each part of the "name" in the tags
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Kingdom != "" {
					ruleProp.Tags = append(ruleProp.Tags, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Kingdom)
				}
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Type != "" {
					ruleProp.Tags = append(ruleProp.Tags, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Type)
				}
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Subtype != "" {
					ruleProp.Tags = append(ruleProp.Tags, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Subtype)
				}

				//Add the SecuritySeverity parameter for GHAS tagging
				ruleProp.SecuritySeverity = strconv.FormatFloat(2*fvdl.Vulnerabilities.Vulnerability[j].InstanceInfo.InstanceSeverity, 'f', 1, 64)

				sarifRule.Properties = ruleProp

				//relationships: will most likely require some expansion
				//One relationship per CWE id
				for j := 0; j < len(cweIds); j++ {
					if cweIds[j] == "None" {
						continue
					}
					sarifRule.Properties.Tags = append(sarifRule.Properties.Tags, "external/cwe/cwe-"+cweIds[j])

					rls := *new(format.Relationships)
					rls.Target.Id = cweIds[j]
					rls.Target.ToolComponent.Name = "CWE"
					rls.Target.ToolComponent.Guid = "25F72D7E-8A92-459D-AD67-64853F788765" //This might not be exact, it is taken from the Microsoft tool converter
					rls.Kinds = append(rls.Kinds, "relevant")
					sarifRule.Relationships = append(sarifRule.Relationships, rls)
				}

				// Add a helpURI as some processors require it
				sarifRule.HelpURI = "https://vulncat.fortify.com/en/weakness"

				//Finalize: append the rule
				tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)

				// Handle simplified version of tool
				sarifRuleSimplified := *new(format.SarifRule)
				sarifRuleSimplified.ID = sarifRule.ID
				sarifRuleSimplified.GUID = sarifRule.GUID
				sarifRuleSimplified.Name = sarifRule.Name
				sarifRuleSimplified.DefaultConfiguration = sarifRule.DefaultConfiguration
				sarifRuleSimplified.Properties = sarifRule.Properties
				toolSimplified.Driver.Rules = append(toolSimplified.Driver.Rules, sarifRuleSimplified)

				// A rule vuln has been found for this rule, no need to keep iterating
				break
			}
		}
	}
	//supportedTaxonomies
	sTax := *new(format.SupportedTaxonomies) //This object seems fixed, but it will have to be checked
	sTax.Name = "CWE"
	sTax.Index = 1
	sTax.Guid = "25F72D7E-8A92-459D-AD67-64853F788765"
	tool.Driver.SupportedTaxonomies = append(tool.Driver.SupportedTaxonomies, sTax)

	//Add additional rulepacks
	for pack := 0; pack < len(fvdl.EngineData.RulePacks); pack++ {
		extension := *new(format.Driver)
		extension.Name = fvdl.EngineData.RulePacks[pack].Name
		extension.Version = fvdl.EngineData.RulePacks[pack].Version
		extension.GUID = fvdl.EngineData.RulePacks[pack].RulePackID
		tool.Extensions = append(tool.Extensions, extension)
	}

	//Finalize: tool
	sarif.Runs[0].Tool = tool
	sarifSimplified.Runs[0].Tool = toolSimplified

	//handle invocations object
	log.Entry().Debug("[SARIF] Now handling invocation.")
	invocation := *new(format.Invocation)
	for i := 0; i < len(fvdl.EngineData.Properties); i++ { //i selects the properties type
		if fvdl.EngineData.Properties[i].PropertiesType == "Fortify" { // This is the correct type, now iterate on props
			for j := 0; j < len(fvdl.EngineData.Properties[i].Property); j++ {
				if fvdl.EngineData.Properties[i].Property[j].Name == "com.fortify.SCAExecutablePath" {
					splitPath := strings.Split(fvdl.EngineData.Properties[i].Property[j].Value, "/")
					invocation.CommandLine = splitPath[len(splitPath)-1]
					break
				}
			}
			break
		}
	}
	invocation.CommandLine = strings.Join(append([]string{invocation.CommandLine}, fvdl.EngineData.CLArguments...), " ")
	invocation.StartTimeUtc = strings.Join([]string{fvdl.Created.Date, fvdl.Created.Time}, "T") + ".000Z"
	for i := 0; i < len(fvdl.EngineData.Errors); i++ {
		ten := *new(format.ToolExecutionNotifications)
		ten.Message.Text = fvdl.EngineData.Errors[i].ErrorMessage
		ten.Descriptor.Id = fvdl.EngineData.Errors[i].ErrorCode
		invocation.ToolExecutionNotifications = append(invocation.ToolExecutionNotifications, ten)
	}
	invocation.ExecutionSuccessful = true //fvdl doesn't seem to plan for this setting
	invocation.Machine = fvdl.EngineData.MachineInfo.Hostname
	invocation.Account = fvdl.EngineData.MachineInfo.Username
	invocProp := new(format.InvocationProperties)
	invocProp.Platform = fvdl.EngineData.MachineInfo.Platform
	invocation.Properties = invocProp
	sarif.Runs[0].Invocations = append(sarif.Runs[0].Invocations, invocation)

	//handle originalUriBaseIds
	if fvdl.Build.SourceBasePath != "" {
		oubi := new(format.OriginalUriBaseIds)
		prefix := "file://"
		if fvdl.Build.SourceBasePath[0] == '/' {
			oubi.SrcRoot.Uri = prefix + fvdl.Build.SourceBasePath + "/"
		} else {
			oubi.SrcRoot.Uri = prefix + "/" + fvdl.Build.SourceBasePath + "/"
		}
		sarif.Runs[0].OriginalUriBaseIds = oubi
	} else {
		log.Entry().Warn("SourceBaesPath is empty")
	}

	//handle artifacts
	log.Entry().Debug("[SARIF] Now handling artifacts.")
	for i := 0; i < len(fvdl.Build.SourceFiles); i++ { //i iterates on source files
		artifact := *new(format.Artifact)
		artifact.Location.Uri = fvdl.Build.SourceFiles[i].Name
		artifact.Location.UriBaseId = "%SRCROOT%"
		artifact.Length = fvdl.Build.SourceFiles[i].FileSize
		switch fvdl.Build.SourceFiles[i].FileType {
		case "java":
			artifact.MimeType = "text/x-java-source"
		case "xml":
			artifact.MimeType = "text/xml"
		default:
			artifact.MimeType = "text"
		}
		artifact.Encoding = fvdl.Build.SourceFiles[i].Encoding
		sarif.Runs[0].Artifacts = append(sarif.Runs[0].Artifacts, artifact)
	}

	//handle automationDetails
	sarif.Runs[0].AutomationDetails = &format.AutomationDetails{Id: fvdl.Build.BuildID}

	//handle threadFlowLocations
	log.Entry().Debug("[SARIF] Now handling threadFlowLocations.")
	threadFlowLocationsObject := []format.Locations{}
	//to ensure an exact replacement in case a threadFlowLocation object refers to another, we prepare a map
	threadFlowIndexMap := make(map[int]([]int)) // This  will store indexes, we will work with it only to reduce item copies to a minimum
	for i := 0; i < len(fvdl.UnifiedNodePool.Node); i++ {
		threadFlowIndexMap[i+1] = append(threadFlowIndexMap[i+1], i+1)
		loc := new(format.Location)
		//get artifact location
		for j := 0; j < len(fvdl.Build.SourceFiles); j++ { // j iterates on source files
			if fvdl.Build.SourceFiles[j].Name == fvdl.UnifiedNodePool.Node[i].SourceLocation.Path {
				loc.PhysicalLocation.ArtifactLocation.Index = j + 1
				loc.PhysicalLocation.ArtifactLocation.URI = fvdl.UnifiedNodePool.Node[i].SourceLocation.Path
				loc.PhysicalLocation.ArtifactLocation.URIBaseId = "%SRCROOT%"
				break
			}
		}
		//get region & context region
		loc.PhysicalLocation.Region.StartLine = fvdl.UnifiedNodePool.Node[i].SourceLocation.Line
		loc.PhysicalLocation.Region.EndLine = fvdl.UnifiedNodePool.Node[i].SourceLocation.LineEnd
		loc.PhysicalLocation.Region.StartColumn = fvdl.UnifiedNodePool.Node[i].SourceLocation.ColStart
		loc.PhysicalLocation.Region.EndColumn = fvdl.UnifiedNodePool.Node[i].SourceLocation.ColEnd
		targetSnippetId := fvdl.UnifiedNodePool.Node[i].SourceLocation.Snippet
		for j := 0; j < len(fvdl.Snippets); j++ {
			if fvdl.Snippets[j].SnippetId == targetSnippetId {
				loc.PhysicalLocation.ContextRegion = new(format.ContextRegion)
				loc.PhysicalLocation.ContextRegion.StartLine = fvdl.Snippets[j].StartLine
				loc.PhysicalLocation.ContextRegion.EndLine = fvdl.Snippets[j].EndLine
				break
			}
		}
		loc.Message = new(format.Message)
		loc.Message.Text = fvdl.UnifiedNodePool.Node[i].Action.ActionData

		log.Entry().Debug("Compute eventual sub-nodes")
		threadFlowIndexMap[i+1] = computeLocationPath(fvdl, i+1) // Recursively traverse array
		locs := format.Locations{Location: loc}
		threadFlowLocationsObject = append(threadFlowLocationsObject, locs)
	}

	sarif.Runs[0].ThreadFlowLocations = threadFlowLocationsObject

	// Now, iterate on threadflows in each result, and replace eventual indexes...
	for i := 0; i < len(sarif.Runs[0].Results); i++ {
		for cf := 0; cf < len(sarif.Runs[0].Results[i].CodeFlows); cf++ {
			for tf := 0; tf < len(sarif.Runs[0].Results[i].CodeFlows[cf].ThreadFlows); tf++ {
				log.Entry().Debug("Handling tf: ", tf, "from instance ", sarif.Runs[0].Results[i].PartialFingerprints.FortifyInstanceID)
				newLocations := *new([]format.Locations)
				for j := 0; j < len(sarif.Runs[0].Results[i].CodeFlows[cf].ThreadFlows[tf].Locations); j++ {
					if sarif.Runs[0].Results[i].CodeFlows[cf].ThreadFlows[tf].Locations[j].Index != 0 {
						indexes := threadFlowIndexMap[sarif.Runs[0].Results[i].CodeFlows[cf].ThreadFlows[tf].Locations[j].Index]
						log.Entry().Debug("Indexes found: ", indexes)
						for rep := range indexes {
							newLocations = append(newLocations, sarif.Runs[0].ThreadFlowLocations[indexes[rep]-1])
							newLocations[rep].Index = 0 // void index
						}
					} else {
						newLocations = append(newLocations, sarif.Runs[0].Results[i].CodeFlows[cf].ThreadFlows[tf].Locations[j])
					}
				}
				sarif.Runs[0].Results[i].CodeFlows[cf].ThreadFlows[tf].Locations = newLocations
			}
		}
	}

	// Threadflowlocations is no loger useful: voiding it will make for smaller reports
	sarif.Runs[0].ThreadFlowLocations = []format.Locations{}

	// Add a conversion object to highlight this isn't native SARIF
	conversion := new(format.Conversion)
	conversion.Tool.Driver.Name = "Piper FPR to SARIF converter"
	conversion.Tool.Driver.InformationUri = "https://github.com/SAP/jenkins-library"
	conversion.Invocation.ExecutionSuccessful = true
	conversion.Invocation.StartTimeUtc = fmt.Sprintf("%s", start.Format("2006-01-02T15:04:05.000Z")) // "YYYY-MM-DDThh:mm:ss.sZ" on 2006-01-02 15:04:05
	conversion.Invocation.Machine = fvdl.EngineData.MachineInfo.Hostname
	conversion.Invocation.Account = fvdl.EngineData.MachineInfo.Username
	convInvocProp := new(format.InvocationProperties)
	convInvocProp.Platform = fvdl.EngineData.MachineInfo.Platform
	conversion.Invocation.Properties = convInvocProp
	sarif.Runs[0].Conversion = conversion

	//handle taxonomies
	//Only one exists apparently: CWE. It is fixed
	taxonomy := *new(format.Taxonomies)
	taxonomy.GUID = "25F72D7E-8A92-459D-AD67-64853F788765"
	taxonomy.Name = "CWE"
	taxonomy.Organization = "MITRE"
	taxonomy.ShortDescription.Text = "The MITRE Common Weakness Enumeration"
	for key := range cweIdsForTaxonomies {
		taxa := *new(format.Taxa)
		taxa.Id = key
		taxonomy.Taxa = append(taxonomy.Taxa, taxa)
	}
	sarif.Runs[0].Taxonomies = append(sarif.Runs[0].Taxonomies, taxonomy)

	return sarif, sarifSimplified, nil
}

func integrateAuditData(ruleProp *format.SarifProperties, issueInstanceID string, sys System, projectVersion *models.ProjectVersion, auditData []*models.ProjectVersionIssue, filterSet *models.FilterSet, oneRequestPerIssue bool, maxretries int) error {

	// Set default values
	ruleProp.Audited = false
	ruleProp.FortifyCategory = "Unknown"
	ruleProp.ToolSeverity = "Unknown"
	ruleProp.ToolState = "Unknown"
	ruleProp.ToolAuditMessage = "Error fetching audit state" // We set this as default for the error phase, then reset it to nothing
	ruleProp.ToolSeverityIndex = 0
	ruleProp.ToolStateIndex = 0
	ruleProp.AuditRequirementIndex = 0
	ruleProp.AuditRequirement = "Unknown"

	// These default values allow for the property bag to be filled even if an error happens later. They all should be overwritten by a normal course of the progrma.
	if maxretries == 0 {
		// Max retries reached, we stop there to avoid a longer execution time
		err := errors.New("request failed: maximum number of retries reached, placeholder values will be set from now on for audit data")
		return err
	} else if maxretries < 0 {
		return nil // Avoid spamming logfile
	}
	if sys == nil {
		ruleProp.ToolAuditMessage = "Cannot fetch audit state: no sys instance"
		err := errors.New("no system instance, lookup impossible for " + issueInstanceID)
		return err
	}
	if projectVersion == nil {
		err := errors.New("project or projectVersion is undefined: lookup aborted for " + issueInstanceID)
		return err
	}
	// Reset the audit message
	ruleProp.ToolAuditMessage = ""
	var data []*models.ProjectVersionIssue
	var err error
	if oneRequestPerIssue {
		log.Entry().Debug("operating in one-request-per-issue mode: looking up audit state of " + issueInstanceID)
		data, err = sys.GetIssueDetails(projectVersion.ID, issueInstanceID)
		if err != nil {
			return err
		}
	} else {
		for i := range auditData {
			if issueInstanceID == *auditData[i].IssueInstanceID {
				data = append(data, auditData[i])
				break
			}
		}
	}
	if len(data) != 1 { //issueInstanceID is supposedly unique so len(data) = 1
		return errors.New("not exactly 1 issue found, found " + fmt.Sprint(len(data)))
	}
	if filterSet != nil {
		for i := 0; i < len(filterSet.Folders); i++ {
			if filterSet.Folders[i].GUID == *data[0].FolderGUID {
				ruleProp.FortifyCategory = filterSet.Folders[i].Name
				//  classify into audit groups
				switch ruleProp.FortifyCategory {
				case "Corporate Security Requirements", "Audit All":
					ruleProp.AuditRequirementIndex = format.AUDIT_REQUIREMENT_GROUP_1_INDEX
					ruleProp.AuditRequirement = format.AUDIT_REQUIREMENT_GROUP_1_DESC
				case "Spot Checks of Each Category":
					ruleProp.AuditRequirementIndex = format.AUDIT_REQUIREMENT_GROUP_2_INDEX
					ruleProp.AuditRequirement = format.AUDIT_REQUIREMENT_GROUP_2_DESC
				case "Optional":
					ruleProp.AuditRequirementIndex = format.AUDIT_REQUIREMENT_GROUP_3_INDEX
					ruleProp.AuditRequirement = format.AUDIT_REQUIREMENT_GROUP_3_DESC
				}
				break
			}
		}
	} else {
		err := errors.New("no filter set defined, category will be missing from " + issueInstanceID)
		return err
	}
	ruleProp.Audited = data[0].Audited
	ruleProp.ToolSeverity = *data[0].Friority
	switch ruleProp.ToolSeverity {
	case "Critical":
		ruleProp.ToolSeverityIndex = 5
	case "Urgent":
		ruleProp.ToolSeverityIndex = 4
	case "High":
		ruleProp.ToolSeverityIndex = 3
	case "Medium":
		ruleProp.ToolSeverityIndex = 2
	case "Low":
		ruleProp.ToolSeverityIndex = 1
	}
	if ruleProp.Audited {
		ruleProp.ToolState = *data[0].PrimaryTag
		switch ruleProp.ToolState { //This is as easy as it can get, seeing that the index is not in the response.
		case "Exploitable":
			ruleProp.ToolStateIndex = 5
		case "Suspicious":
			ruleProp.ToolStateIndex = 4
		case "Bad Practice":
			ruleProp.ToolStateIndex = 3
		case "Reliability Issue":
			ruleProp.ToolStateIndex = 2
		case "Not an Issue":
			ruleProp.ToolStateIndex = 1
		}
	} else {
		ruleProp.ToolState = "Unreviewed"
	}
	if *data[0].HasComments { //fetch latest message if comments exist
		//Fetch the ID
		parentID := data[0].ID
		commentData, err := sys.GetIssueComments(parentID)
		if err != nil {
			return err
		}
		ruleProp.ToolAuditMessage = unescapeXML(*commentData[0].Comment)
	}
	return nil
}

// Factorizes some code used to obtain the relevant value for a snippet based on the type given by Fortify
// Note: snippet text is no longer part of .sarif due to size issue.
// This function however is helpful to explain how to get snippet out of FPR
func handleSnippet(snippetType string, snippet string) string {
	snippetTarget := ""
	switch snippetType {
	case "Assign":
		snippetWords := strings.Split(snippet, " ")
		if snippetWords[0] == "Assignment" {
			snippetTarget = snippetWords[2]
		} else {
			snippetTarget = snippet
		}
	case "InCall":
		snippetTarget = strings.Split(snippet, "(")[0]
	case "OutCall":
		snippetTarget = strings.Split(snippet, "(")[0]
	case "InOutCall":
		snippetTarget = strings.Split(snippet, "(")[0]
	case "Return":
		snippetTarget = snippet
	case "Read":
		snippetWords := strings.Split(snippet, " ")
		if len(snippetWords) > 1 {
			snippetTarget = " " + snippetWords[1]
		} else {
			snippetTarget = snippetWords[0]
		}
	default:
		snippetTarget = snippet
	}
	return snippetTarget
}

func unescapeXML(input string) string {
	raw := input
	// Post-treat string to change the XML escaping generated by Unmarshal
	raw = strings.ReplaceAll(raw, "&amp;", "&")
	raw = strings.ReplaceAll(raw, "&lt;", "<")
	raw = strings.ReplaceAll(raw, "&gt;", ">")
	raw = strings.ReplaceAll(raw, "&apos;", "'")
	raw = strings.ReplaceAll(raw, "&quot;", "\"")
	return raw
}

// Used to build a reference array of index for the successors of each node in the UnifiedNodePool
func computeLocationPath(fvdl FVDL, input int) []int {
	log.Entry().Debug("Computing for ID ", input)
	// Find the successors of input
	var subnodes []int
	var result []int
	for j := 0; j < len(fvdl.UnifiedNodePool.Node[input-1].Reason.Trace.Primary.Entry); j++ {
		if fvdl.UnifiedNodePool.Node[input-1].Reason.Trace.Primary.Entry[j].NodeRef.RefId != 0 && fvdl.UnifiedNodePool.Node[input-1].Reason.Trace.Primary.Entry[j].NodeRef.RefId != (input-1) {
			subnodes = append(subnodes, fvdl.UnifiedNodePool.Node[input-1].Reason.Trace.Primary.Entry[j].NodeRef.RefId+1)
		}
	}
	result = append(result, input)
	log.Entry().Debug("Successors: ", subnodes)
	for j := 0; j < len(subnodes); j++ {
		result = append(result, computeLocationPath(fvdl, subnodes[j])...)
	}
	log.Entry().Debug("Finishing computing for ID ", input)
	return result
}
