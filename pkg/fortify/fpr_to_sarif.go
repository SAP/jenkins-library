package fortify

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"

	"github.com/SAP/jenkins-library/pkg/log"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

// This struct encapsulates everyting in the FVDL document

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

type CreatedTS struct {
	XMLName xml.Name `xml:"CreatedTS"`
	Date    string   `xml:"date,attr"`
	Time    string   `xml:"time,attr"`
}

type UUID struct {
	XMLName xml.Name `xml:"UUID"`
	Uuid    string   `xml:",innerxml"`
}

// These structures are relevant to the Build object

type LOC struct {
	XMLName  xml.Name `xml:"LOC"`
	LocType  string   `xml:"type,attr"`
	LocValue string   `xml:",innerxml"`
}

type Build struct {
	XMLName        xml.Name `xml:"Build"`
	Project        string   `xml:"Project"`
	Label          string   `xml:"Label"`
	BuildID        string   `xml:"BuildID"`
	NumberFiles    int      `xml:"NumberFiles"`
	Locs           []LOC    `xml:",any"`
	JavaClassPath  string   `xml:"JavaClasspath"`
	SourceBasePath string   `xml:"SourceBasePath"`
	SourceFiles    []File   `xml:"SourceFiles>File"`
	Scantime       ScanTime `xml:"ScanTime"`
}

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

type ScanTime struct {
	XMLName xml.Name `xml:"ScanTime"`
	Value   int      `xml:"value,attr"`
}

// These structures are relevant to the Vulnerabilities object

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

type ClassInfo struct {
	XMLName         xml.Name `xml:"ClassInfo"`
	ClassID         string   `xml:"ClassID"`
	Kingdom         string   `xml:"Kingdom,omitempty"`
	Type            string   `xml:"Type"`
	Subtype         string   `xml:"Subtype,omitempty"`
	AnalyzerName    string   `xml:"AnalyzerName"`
	DefaultSeverity string   `xml:"DefaultSeverity"`
}

type InstanceInfo struct {
	XMLName          xml.Name `xml:"InstanceInfo"`
	InstanceID       string   `xml:"InstanceID"`
	InstanceSeverity string   `xml:"InstanceSeverity"`
	Confidence       string   `xml:"Confidence"`
}

type AnalysisInfo struct { //Note that this is directly the "Unified" object
	Context                Context
	ReplacementDefinitions []Def   `xml:"ReplacementDefinitions>Def"`
	Trace                  []Trace `xml:"Trace"`
}

type Context struct {
	XMLName   xml.Name `xml:"Context"`
	ContextId string   `xml:"id,attr,omitempty"`
	Function  Function
	FDSL      FunctionDeclarationSourceLocation
}

type Function struct {
	XMLName                xml.Name `xml:"Function"`
	FunctionName           string   `xml:"name,attr"`
	FunctionNamespace      string   `xml:"namespace,attr"`
	FunctionEnclosingClass string   `xml:"enclosingClass,attr"`
}

type FunctionDeclarationSourceLocation struct {
	XMLName      xml.Name `xml:"FunctionDeclarationSourceLocation"`
	FDSLPath     string   `xml:"path,attr"`
	FDSLLine     string   `xml:"line,attr"`
	FDSLLineEnd  string   `xml:"lineEnd,attr"`
	FDSLColStart string   `xml:"colStart,attr"`
	FDSLColEnd   string   `xml:"colEnd,attr"`
}

type Def struct {
	XMLName  xml.Name `xml:"Def"`
	DefKey   string   `xml:"key,attr"`
	DefValue string   `xml:"value,attr"`
}

type Trace struct {
	XMLName xml.Name `xml:"Trace"`
	Primary Primary  `xml:"Primary"`
}

type Primary struct {
	XMLName xml.Name `xml:"Primary"`
	Entry   []Entry  `xml:"Entry"`
}

type Entry struct {
	XMLName xml.Name `xml:"Entry"`
	NodeRef NodeRef  `xml:"NodeRef,omitempty"`
	Node    Node     `xml:"Node,omitempty"`
}

type NodeRef struct {
	XMLName xml.Name `xml:"NodeRef"`
	RefId   int      `xml:"id,attr"`
}

type Node struct {
	XMLName        xml.Name       `xml:"Node"`
	IsDefault      string         `xml:"isDefault,attr,omitempty"`
	NodeLabel      string         `xml:"label,attr,omitempty"`
	SourceLocation SourceLocation `xml:"SourceLocation"`
	Action         Action         `xml:"Action,omitempty"`
	Reason         Reason         `xml:"Reason,omitempty"`
	Knowledge      Knowledge      `xml:"Knowledge,omitempty"`
}

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

type Action struct {
	XMLName    xml.Name `xml:"Action"`
	Type       string   `xml:"type,attr"`
	ActionData string   `xml:",innerxml"`
}

type Reason struct {
	XMLName xml.Name `xml:"Reason"`
	Rule    Rule     `xml:"Rule,omitempty"`
	Trace   Trace    `xml:"Trace,omitempty"`
}

type Rule struct {
	XMLName xml.Name `xml:"Rule"`
	RuleID  string   `xml:"ruleID,attr"`
}

type Group struct {
	XMLName xml.Name `xml:"Group"`
	Name    string   `xml:"name,attr"`
	Data    string   `xml:",innerxml"`
}

type Knowledge struct {
	XMLName xml.Name `xml:"Knowledge"`
	Facts   []Fact   `xml:"Fact"`
}

type Fact struct {
	XMLName  xml.Name `xml:"Fact"`
	Primary  string   `xml:"primary,attr"`
	Type     string   `xml:"type,attr,omitempty"`
	FactData string   `xml:",innerxml"`
}

// These structures are relevant to the ContextPool object

type ContextPool struct {
	XMLName xml.Name  `xml:"ContextPool"`
	Context []Context `xml:"Context"`
}

// These structures are relevant to the UnifiedNodePool object

type UnifiedNodePool struct {
	XMLName xml.Name `xml:"UnifiedNodePool"`
	Node    []Node   `xml:"Node"`
}

// These structures are relevant to the Description object

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

type Abstract struct {
	XMLName xml.Name `xml:"Abstract"`
	Text    string   `xml:",innerxml"`
}

type Explanation struct {
	XMLName xml.Name `xml:"Explanation"`
	Text    string   `xml:",innerxml"`
}

type Recommendations struct {
	XMLName xml.Name `xml:"Recommendations"`
	Text    string   `xml:",innerxml"`
}

type Reference struct {
	XMLName xml.Name `xml:"Reference"`
	Title   string   `xml:"Title"`
	Author  string   `xml:"Author"`
}

type Tip struct {
	XMLName xml.Name `xml:"Tip"`
	Tip     string   `xml:",innerxml"`
}

type CustomDescription struct {
	XMLName         xml.Name        `xml:"CustomDescription"`
	ContentType     string          `xml:"contentType,attr"`
	RuleID          string          `xml:"ruleID,attr"`
	Explanation     Explanation     `xml:"Explanation"`
	Recommendations Recommendations `xml:"Recommendations"`
	References      []Reference     `xml:"References>Reference"`
}

// These structures are relevant to the Snippets object

type Snippet struct {
	XMLName   xml.Name `xml:"Snippet"`
	SnippetId string   `xml:"id,attr"`
	File      string   `xml:"File"`
	StartLine int      `xml:"StartLine"`
	EndLine   int      `xml:"EndLine"`
	Text      string   `xml:"Text"`
}

// These structures are relevant to the ProgramData object

type ProgramData struct {
	XMLName         xml.Name         `xml:"ProgramData"`
	Sources         []SourceInstance `xml:"Sources>SourceInstance"`
	Sinks           []SinkInstance   `xml:"Sinks>SinkInstance"`
	CalledWithNoDef []Function       `xml:"CalledWithNoDef>Function"`
}

type SourceInstance struct {
	XMLName        xml.Name       `xml:"SourceInstance"`
	RuleID         string         `xml:"ruleID,attr"`
	FunctionCall   FunctionCall   `xml:"FunctionCall,omitempty"`
	FunctionEntry  FunctionEntry  `xml:"FunctionEntry,omitempty"`
	SourceLocation SourceLocation `xml:"SourceLocation,omitempty"`
	TaintFlags     TaintFlags     `xml:"TaintFlags"`
}

type FunctionCall struct {
	XMLName        xml.Name       `xml:"FunctionCall"`
	SourceLocation SourceLocation `xml:"SourceLocation"`
	Function       Function       `xml:"Function"`
}

type FunctionEntry struct {
	XMLName        xml.Name       `xml:"FunctionEntry"`
	SourceLocation SourceLocation `xml:"SourceLocation"`
	Function       Function       `xml:"Function"`
}

type TaintFlags struct {
	XMLName   xml.Name    `xml:"TaintFlags"`
	TaintFlag []TaintFlag `xml:"TaintFlag"`
}

type TaintFlag struct {
	XMLName       xml.Name `xml:"TaintFlag"`
	TaintFlagName string   `xml:"name,attr"`
}

type SinkInstance struct {
	XMLName        xml.Name       `xml:"SinkInstance"`
	RuleID         string         `xml:"ruleID,attr"`
	FunctionCall   FunctionCall   `xml:"FunctionCall,omitempty"`
	SourceLocation SourceLocation `xml:"SourceLocation,omitempty"`
}

// These structures are relevant to the EngineData object

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

type RulePack struct {
	XMLName    xml.Name `xml:"RulePack"`
	RulePackID string   `xml:"RulePackID"`
	SKU        string   `xml:"SKU"`
	Name       string   `xml:"Name"`
	Version    string   `xml:"Version"`
	MAC        string   `xml:"MAC"`
}

type Properties struct {
	XMLName        xml.Name   `xml:"Properties"`
	PropertiesType string     `xml:"type,attr"`
	Property       []Property `xml:"Property"`
}

type Property struct {
	XMLName xml.Name `xml:"Property"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

type Error struct {
	XMLName      xml.Name `xml:"Error"`
	ErrorCode    string   `xml:"code,attr"`
	ErrorMessage string   `xml:",innerxml"`
}

type MachineInfo struct {
	XMLName  xml.Name `xml:"MachineInfo"`
	Hostname string   `xml:"Hostname"`
	Username string   `xml:"Username"`
	Platform string   `xml:"Platform"`
}

type FilterResult struct {
	XMLName xml.Name `xml:"FilterResult"`
	//Todo? No data in sample audit file
}

type RuleInfo struct {
	XMLName       xml.Name `xml:"Rule"`
	RuleID        string   `xml:"id,attr"`
	MetaInfoGroup []Group  `xml:"MetaInfo>Group,omitempty"`
}

type LicenseInfo struct {
	XMLName    xml.Name     `xml:"LicenseInfo"`
	Metadata   []Metadata   `xml:"Metadata"`
	Capability []Capability `xml:"Capability"`
}

type Metadata struct {
	XMLName xml.Name `xml:"Metadata"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

type Capability struct {
	XMLName    xml.Name  `xml:"Capability"`
	Name       string    `xml:"Name"`
	Expiration string    `xml:"Expiration"`
	Attribute  Attribute `xml:"Attribute"`
}

type Attribute struct {
	XMLName xml.Name `xml:"Attribute"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

// JSON receptacle structs

type SARIF struct {
	Schema  string `json:"$schema" default:"https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos01/schemas/sarif-schema-2.1.0.json"`
	Version string `json:"version" default:"2.1.0"`
	Runs    []Runs `json:"runs"`
}

func ConvertFprToSarif(resultFilePath string) error {
	log.Entry().Debug("Extracting FPR.")
	_, err := FileUtils.Unzip(resultFilePath, "result/")
	if err != nil {
		return err
	}
	//File is result/audit.fvdl
	data, err := ioutil.ReadFile("result/audit.fvdl")
	if err != nil {
		return err
	}

	err = Parse(data)
	return err
}

func Parse(data []byte) error {
	//To read XML data, Unmarshal or Decode can be used, here we use Decode to work on the stream
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	var result FVDL
	decoder.Decode(&result)
	resultJSON, err := json.MarshalIndent(result, "", "  ") //Temporary marshal as json to write file
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("target/audit.sarif", resultJSON, 0700)
	return err
}
