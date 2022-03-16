package fortify

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/piper-validation/fortify-client-go/models"

	"github.com/SAP/jenkins-library/pkg/format"

	"github.com/SAP/jenkins-library/pkg/log"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

// FVDL This struct encapsulates everyting in the FVDL document
type FVDL struct {
	XMLName         xml.Name `xml:"FVDL"`
	Xmlns           string   `xml:"xmlns,attr"`
	XmlnsXsi        string   `xml:"xsi,attr"`
	Version         string   `xml:"version,attr"`
	XsiType         string   `xml:"type,attr"`
	Created         CreatedTS
	Uuid            UUID
	Build           Build
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
	DefaultSeverity string   `xml:"DefaultSeverity"`
}

// InstanceInfo
type InstanceInfo struct {
	XMLName          xml.Name `xml:"InstanceInfo"`
	InstanceID       string   `xml:"InstanceID"`
	InstanceSeverity string   `xml:"InstanceSeverity"`
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
	XMLName   xml.Name `xml:"Context"`
	ContextId string   `xml:"id,attr,omitempty"`
	Function  Function
	FDSL      FunctionDeclarationSourceLocation
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

// Utils

func (n Node) isEmpty() bool {
	return n.IsDefault == ""
}

func (a Action) isEmpty() bool {
	return a.ActionData == ""
}

// ConvertFprToSarif converts the FPR file contents into SARIF format
func ConvertFprToSarif(sys System, project *models.Project, projectVersion *models.ProjectVersion, resultFilePath string) (format.SARIF, error) {
	log.Entry().Debug("Extracting FPR.")
	var sarif format.SARIF
	tmpFolder, err := ioutil.TempDir(".", "temp-")
	defer os.RemoveAll(tmpFolder)
	if err != nil {
		log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
		return sarif, err
	}

	_, err = FileUtils.Unzip(resultFilePath, tmpFolder)
	if err != nil {
		return sarif, err
	}

	data, err := ioutil.ReadFile(filepath.Join(tmpFolder, "audit.fvdl"))
	if err != nil {
		return sarif, err
	}
	if len(data) == 0 {
		log.Entry().Error("Error reading audit file at " + filepath.Join(tmpFolder, "audit.fvdl") + ". This might be that the file is missing, corrupted, or too large. Aborting procedure.")
		err := errors.New("cannot read audit file")
		return sarif, err
	}

	log.Entry().Debug("Calling Parse.")
	return Parse(sys, project, projectVersion, data)
}

// Parse parses the FPR file
func Parse(sys System, project *models.Project, projectVersion *models.ProjectVersion, data []byte) (format.SARIF, error) {
	//To read XML data, Unmarshal or Decode can be used, here we use Decode to work on the stream
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	var fvdl FVDL
	err := decoder.Decode(&fvdl)
	if err != nil {
		return format.SARIF{}, err
	}

	//Now, we handle the sarif
	var sarif format.SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos01/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var fortifyRun format.Runs
	fortifyRun.ColumnKind = "utf16CodeUnits"
	cweIdsForTaxonomies := make(map[string]string) //Defining this here and filling it in the course of the program helps filling the Taxonomies object easily. Map because easy to check for keys
	sarif.Runs = append(sarif.Runs, fortifyRun)

	// Handle results/vulnerabilities
	for i := 0; i < len(fvdl.Vulnerabilities.Vulnerability); i++ {
		result := *new(format.Results)
		result.RuleID = fvdl.Vulnerabilities.Vulnerability[i].ClassInfo.ClassID
		result.Level = "none" //TODO
		//get message
		for j := 0; j < len(fvdl.Description); j++ {
			if fvdl.Description[j].ClassID == result.RuleID {
				result.RuleIndex = j //Seems very abstract
				rawMessage := fvdl.Description[j].Abstract.Text
				// Replacement defintions in message
				for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def); l++ {
					rawMessage = strings.ReplaceAll(rawMessage, "Replace key=\""+fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def[l].DefKey+"\"", fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def[l].DefValue)
				}
				msg := new(format.Message)
				msg.Text = rawMessage
				result.Message = msg
				break
			}
		}

		// Handle all locations items
		location := *new(format.Location)
		var startingColumn int
		//get location
		for k := 0; k < len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace); k++ { // k iterates on traces
			//In each trace/primary, there can be one or more entries
			//Each trace represents a codeflow, each entry represents a location in threadflow
			codeFlow := *new(format.CodeFlow)
			threadFlow := *new(format.ThreadFlow)
			//We now iterate on Entries in the trace/primary
			for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry); l++ { // l iterates on entries
				threadFlowLocation := *new(format.Locations) //One is created regardless
				//the default node dictates the interesting threadflow (location, and so on)
				//this will populate both threadFlowLocation AND the parent location object (result.Locations[0])
				if !fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.isEmpty() && fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.IsDefault == "true" {
					//initalize threadFlowLocation.Location
					threadFlowLocation.Location = new(format.Location)
					//get artifact location
					for j := 0; j < len(fvdl.Build.SourceFiles); j++ { // j iterates on source files
						if fvdl.Build.SourceFiles[j].Name == fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.Path {
							threadFlowLocation.Location.PhysicalLocation.ArtifactLocation.Index = j
							break
						}
					}
					//get region & context region
					threadFlowLocation.Location.PhysicalLocation.Region.StartLine = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.Line
					//Snippet is handled last
					//threadFlowLocation.Location.PhysicalLocation.Region.Snippet.Text = "foobar"
					targetSnippetId := fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.Snippet
					for j := 0; j < len(fvdl.Snippets); j++ {
						if fvdl.Snippets[j].SnippetId == targetSnippetId {
							threadFlowLocation.Location.PhysicalLocation.ContextRegion.StartLine = fvdl.Snippets[j].StartLine
							threadFlowLocation.Location.PhysicalLocation.ContextRegion.EndLine = fvdl.Snippets[j].EndLine
							threadFlowLocation.Location.PhysicalLocation.ContextRegion.Snippet.Text = fvdl.Snippets[j].Text
							break
						}
					}
					//parse SourceLocation object for the startColumn value, store it appropriately
					startingColumn = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.SourceLocation.ColStart
					//check for existance of action object, and if yes, save message
					if !fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.isEmpty() {
						threadFlowLocation.Location.Message = new(format.Message)
						threadFlowLocation.Location.Message.Text = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData
						// Handle snippet
						snippetTarget := ""
						switch fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.Type {
						case "Assign":
							snippetWords := strings.Split(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData, " ")
							if snippetWords[0] == "Assignment" {
								snippetTarget = snippetWords[2]
							} else {
								snippetTarget = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData
							}
						case "InCall":
							snippetTarget = strings.Split(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData, "(")[0]
						case "OutCall":
							snippetTarget = strings.Split(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData, "(")[0]
						case "InOutCall":
							snippetTarget = strings.Split(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData, "(")[0]
						case "Return":
							snippetTarget = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData
						case "Read":
							snippetWords := strings.Split(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData, " ")
							if len(snippetWords) > 1 {
								snippetTarget = " " + snippetWords[1]
							} else {
								snippetTarget = snippetWords[0]
							}
						default:
							snippetTarget = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].Node.Action.ActionData
						}
						physLocationSnippetLines := strings.Split(threadFlowLocation.Location.PhysicalLocation.ContextRegion.Snippet.Text, "\n")
						snippetText := ""
						for j := 0; j < len(physLocationSnippetLines); j++ {
							if strings.Contains(physLocationSnippetLines[j], snippetTarget) {
								snippetText = physLocationSnippetLines[j]
								break
							}
						}
						if snippetText != "" {
							threadFlowLocation.Location.PhysicalLocation.Region.Snippet.Text = snippetText
						} else {
							threadFlowLocation.Location.PhysicalLocation.Region.Snippet.Text = threadFlowLocation.Location.PhysicalLocation.ContextRegion.Snippet.Text
						}
					} else {
						threadFlowLocation.Location.PhysicalLocation.Region.Snippet.Text = threadFlowLocation.Location.PhysicalLocation.ContextRegion.Snippet.Text
					}
					location = *threadFlowLocation.Location
					//set Kinds
					threadFlowLocation.Kinds = append(threadFlowLocation.Kinds, "unknown") //TODO
				} else { //is not a main threadflow: just register NodeRef index in threadFlowLocation
					threadFlowLocation.Index = fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.Trace[k].Primary.Entry[l].NodeRef.RefId
				}
				//add the threadflowlocation to the list of locations
				threadFlow.Locations = append(threadFlow.Locations, threadFlowLocation)
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
		relatedLocation.PhysicalLocation.Region.StartColumn = startingColumn
		result.RelatedLocations = append(result.RelatedLocations, relatedLocation)

		//handle properties
		prop := *new(format.SarifProperties)
		prop.InstanceSeverity = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceSeverity
		prop.Confidence = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.Confidence
		prop.InstanceID = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceID
		//Use a query to get the audit data
		// B5C0FEFD-CCB2-4F21-A9D7-87AE600A5885 is "custom rules": handle differently?
		if result.RuleID == "B5C0FEFD-CCB2-4F21-A9D7-87AE600A5885" {
			// Custom Rules has no audit value: it's notificaiton in the FVDL only.
			prop.Audited = true
			prop.ToolAuditMessage = "Custom Rules: not a vuln"
			prop.ToolState = "Not an Issue"
			prop.ToolStateIndex = 1
		} else if sys != nil {
			if err := integrateAuditData(&prop, fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceID, sys, project, projectVersion); err != nil {
				log.Entry().Debug(err)
				prop.Audited = false
				prop.ToolState = "Unknown"
				prop.ToolAuditMessage = "Error fetching audit state"
			}
		} else {
			prop.Audited = false
			prop.ToolState = "Unknown"
			prop.ToolAuditMessage = "Cannot fetch audit state"
		}
		result.Properties = prop

		sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)
	}

	//handle the tool object
	tool := *new(format.Tool)
	tool.Driver = *new(format.Driver)
	tool.Driver.Name = "MicroFocus Fortify SCA"
	tool.Driver.Version = fvdl.EngineData.EngineVersion
	tool.Driver.InformationUri = "https://www.microfocus.com/documentation/fortify-static-code-analyzer-and-tools/2020/SCA_Guide_20.2.0.pdf"

	//handles rules
	for i := 0; i < len(fvdl.EngineData.RuleInfo); i++ { //i iterates on rules
		sarifRule := *new(format.SarifRule)
		sarifRule.ID = fvdl.EngineData.RuleInfo[i].RuleID
		sarifRule.GUID = fvdl.EngineData.RuleInfo[i].RuleID
		for j := 0; j < len(fvdl.Vulnerabilities.Vulnerability); j++ { //j iterates on vulns to find the name
			if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.ClassID == fvdl.EngineData.RuleInfo[i].RuleID {
				var nameArray []string
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Kingdom != "" {
					nameArray = append(nameArray, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Kingdom)
				}
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Type != "" {
					nameArray = append(nameArray, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Type)
				}
				if fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Subtype != "" {
					nameArray = append(nameArray, fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.Subtype)
				}
				sarifRule.Name = strings.Join(nameArray, "/")
				defaultConfig := new(format.DefaultConfiguration)
				defaultConfig.Properties.DefaultSeverity = fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.DefaultSeverity
				sarifRule.DefaultConfiguration = defaultConfig
				break
			}
		}
		//Descriptions
		for j := 0; j < len(fvdl.Description); j++ {
			if fvdl.Description[j].ClassID == sarifRule.ID {
				rawAbstract := fvdl.Description[j].Abstract.Text
				rawExplanation := fvdl.Description[j].Explanation.Text
				// Replacement defintions in abstract/explanation
				for k := 0; k < len(fvdl.Vulnerabilities.Vulnerability); k++ { // Iterate on vulns to find the correct one (where ReplacementDefinitions are)
					if fvdl.Vulnerabilities.Vulnerability[k].ClassInfo.ClassID == fvdl.Description[j].ClassID {
						for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def); l++ {
							rawAbstract = strings.ReplaceAll(rawAbstract, "Replace key=\""+fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def[l].DefKey+"\"", fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def[l].DefValue)
							rawExplanation = strings.ReplaceAll(rawExplanation, "Replace key=\""+fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def[l].DefKey+"\"", fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.Def[l].DefValue)
						}
						// Replacement locationdef in explanation
						for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.LocationDef); l++ {
							rawExplanation = strings.ReplaceAll(rawExplanation, fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.LocationDef[l].Key, fvdl.Vulnerabilities.Vulnerability[k].AnalysisInfo.ReplacementDefinitions.LocationDef[l].Path)
						}
						// If Description has a CustomDescription, add it for good measure
						if fvdl.Description[j].CustomDescription.RuleID != "" {
							rawExplanation = rawExplanation + "\n;" + fvdl.Description[j].CustomDescription.Explanation.Text
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
		// Avoid empty descriptions to respect standard
		if sarifRule.ShortDescription.Text == "" {
			sarifRule.ShortDescription.Text = "None."
		}
		if sarifRule.FullDescription.Text == "" { // OR USE OMITEMPTY
			sarifRule.FullDescription.Text = "None."
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
					for k := 0; k < len(rawCweIds); k++ {
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
		if len(propArray) != 0 {
			ruleProp = new(format.SarifRuleProperties)
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
		sarifRule.Properties = ruleProp

		//relationships: will most likely require some expansion
		//One relationship per CWE id
		for j := 0; j < len(cweIds); j++ {
			rls := *new(format.Relationships)
			rls.Target.Id = cweIds[j]
			rls.Target.ToolComponent.Name = "CWE"
			rls.Target.ToolComponent.Guid = "25F72D7E-8A92-459D-AD67-64853F788765"
			rls.Kinds = append(rls.Kinds, "relevant")
			sarifRule.Relationships = append(sarifRule.Relationships, rls)
		}

		//Finalize: append the rule
		tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)
	}
	//supportedTaxonomies
	sTax := *new(format.SupportedTaxonomies) //This object seems fixed, but it will have to be checked
	sTax.Name = "CWE"
	sTax.Index = 0
	sTax.Guid = "25F72D7E-8A92-459D-AD67-64853F788765"
	tool.Driver.SupportedTaxonomies = append(tool.Driver.SupportedTaxonomies, sTax)

	//Finalize: tool
	sarif.Runs[0].Tool = tool

	//handle invocations object
	invocation := *new(format.Invocations)
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
	invocation.Properties.Platform = fvdl.EngineData.MachineInfo.Platform
	sarif.Runs[0].Invocations = append(sarif.Runs[0].Invocations, invocation)

	//handle originalUriBaseIds
	oubi := new(format.OriginalUriBaseIds)
	oubi.SrcRoot.Uri = "file:///" + fvdl.Build.SourceBasePath + "/"
	sarif.Runs[0].OriginalUriBaseIds = oubi

	//handle artifacts
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
	sarif.Runs[0].AutomationDetails.Id = fvdl.Build.BuildID

	//handle threadFlowLocations
	threadFlowLocationsObject := []format.Locations{}
	//prepare a check object
	for i := 0; i < len(fvdl.UnifiedNodePool.Node); i++ {
		unique := true
		//Uniqueness Check
		for check := 0; check < i; check++ {
			if fvdl.UnifiedNodePool.Node[i].SourceLocation.Snippet == fvdl.UnifiedNodePool.Node[check].SourceLocation.Snippet &&
				fvdl.UnifiedNodePool.Node[i].Action.ActionData == fvdl.UnifiedNodePool.Node[check].Action.ActionData {
				unique = false
			}
		}
		if !unique {
			continue
		}
		locations := *new(format.Locations)
		loc := new(format.Location)
		//get artifact location
		for j := 0; j < len(fvdl.Build.SourceFiles); j++ { // j iterates on source files
			if fvdl.Build.SourceFiles[j].Name == fvdl.UnifiedNodePool.Node[i].SourceLocation.Path {
				loc.PhysicalLocation.ArtifactLocation.Index = j
				break
			}
		}
		//get region & context region
		loc.PhysicalLocation.Region.StartLine = fvdl.UnifiedNodePool.Node[i].SourceLocation.Line
		//loc.PhysicalLocation.Region.Snippet.Text = "foobar" //TODO
		targetSnippetId := fvdl.UnifiedNodePool.Node[i].SourceLocation.Snippet
		for j := 0; j < len(fvdl.Snippets); j++ {
			if fvdl.Snippets[j].SnippetId == targetSnippetId {
				loc.PhysicalLocation.ContextRegion.StartLine = fvdl.Snippets[j].StartLine
				loc.PhysicalLocation.ContextRegion.EndLine = fvdl.Snippets[j].EndLine
				loc.PhysicalLocation.ContextRegion.Snippet.Text = fvdl.Snippets[j].Text
				break
			}
		}
		loc.Message = new(format.Message)
		loc.Message.Text = fvdl.UnifiedNodePool.Node[i].Action.ActionData
		// Handle snippet
		snippetTarget := ""
		switch fvdl.UnifiedNodePool.Node[i].Action.Type {
		case "Assign":
			snippetWords := strings.Split(fvdl.UnifiedNodePool.Node[i].Action.ActionData, " ")
			if snippetWords[0] == "Assignment" {
				snippetTarget = snippetWords[2]
			} else {
				snippetTarget = fvdl.UnifiedNodePool.Node[i].Action.ActionData
			}
		case "InCall":
			snippetTarget = strings.Split(fvdl.UnifiedNodePool.Node[i].Action.ActionData, "(")[0]
		case "OutCall":
			snippetTarget = strings.Split(fvdl.UnifiedNodePool.Node[i].Action.ActionData, "(")[0]
		case "InOutCall":
			snippetTarget = strings.Split(fvdl.UnifiedNodePool.Node[i].Action.ActionData, "(")[0]
		case "Return":
			snippetTarget = fvdl.UnifiedNodePool.Node[i].Action.ActionData
		case "Read":
			snippetWords := strings.Split(fvdl.UnifiedNodePool.Node[i].Action.ActionData, " ")
			if len(snippetWords) > 1 {
				snippetTarget = " " + snippetWords[1]
			} else {
				snippetTarget = snippetWords[0]
			}
		default:
			snippetTarget = fvdl.UnifiedNodePool.Node[i].Action.ActionData
		}
		physLocationSnippetLines := strings.Split(loc.PhysicalLocation.ContextRegion.Snippet.Text, "\n")
		snippetText := ""
		for j := 0; j < len(physLocationSnippetLines); j++ {
			if strings.Contains(physLocationSnippetLines[j], snippetTarget) {
				snippetText = physLocationSnippetLines[j]
				break
			}
		}
		if snippetText != "" {
			loc.PhysicalLocation.Region.Snippet.Text = snippetText
		} else {
			loc.PhysicalLocation.Region.Snippet.Text = loc.PhysicalLocation.ContextRegion.Snippet.Text
		}
		locations.Location = loc
		locations.Kinds = append(locations.Kinds, "unknown")
		threadFlowLocationsObject = append(threadFlowLocationsObject, locations)
	}

	sarif.Runs[0].ThreadFlowLocations = threadFlowLocationsObject

	//handle taxonomies
	//Only one exists apparently: CWE. It is fixed
	taxonomy := *new(format.Taxonomies)
	taxonomy.Guid = "25F72D7E-8A92-459D-AD67-64853F788765"
	taxonomy.Name = "CWE"
	taxonomy.Organization = "MITRE"
	taxonomy.ShortDescription.Text = "The MITRE Common Weakness Enumeration"
	for key := range cweIdsForTaxonomies {
		taxa := *new(format.Taxa)
		taxa.Id = key
		taxonomy.Taxa = append(taxonomy.Taxa, taxa)
	}
	sarif.Runs[0].Taxonomies = append(sarif.Runs[0].Taxonomies, taxonomy)

	return sarif, nil
}

func integrateAuditData(ruleProp *format.SarifProperties, issueInstanceID string, sys System, project *models.Project, projectVersion *models.ProjectVersion) error {
	if sys == nil {
		err := errors.New("no system instance, lookup impossible for " + issueInstanceID)
		return err
	}
	if project == nil || projectVersion == nil {
		err := errors.New("project or projectVersion is undefined: lookup aborted for " + issueInstanceID)
		return err
	}
	data, err := sys.GetIssueDetails(projectVersion.ID, issueInstanceID)
	if err != nil {
		return err
	}
	log.Entry().Debug("Looking up audit state of " + issueInstanceID)
	if len(data) != 1 { //issueInstanceID is supposedly unique so len(data) = 1
		log.Entry().Error("not exactly 1 issue found, found " + fmt.Sprint(len(data)))
		return errors.New("not exactly 1 issue found, found " + fmt.Sprint(len(data)))
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
		ruleProp.ToolAuditMessage = *commentData[0].Comment
	}
	return nil
}
