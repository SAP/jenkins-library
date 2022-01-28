package fortify

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/piper-validation/fortify-client-go/models"

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
	ReplacementDefinitions ReplacementDefinitions `xml:"ReplacementDefinitions"`
	Trace                  []Trace                `xml:"Trace"`
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

type ReplacementDefinitions struct {
	XMLName     xml.Name      `xml:"ReplacementDefinitions"`
	Def         []Def         `xml:"Def"`
	LocationDef []LocationDef `xml:"LocationDef"`
}

type Def struct {
	XMLName  xml.Name `xml:"Def"`
	DefKey   string   `xml:"key,attr"`
	DefValue string   `xml:"value,attr"`
}

type LocationDef struct {
	XMLName  xml.Name `xml:"LocationDef"`
	Path     string   `xml:"path,attr"`
	Line     int      `xml:"line,attr"`
	LineEnd  int      `xml:"lineEnd,attr"`
	ColStart int      `xml:"colStart,attr"`
	ColEnd   int      `xml:"colEnd,attr"`
	Key      string   `xml:"key,attr"`
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

// These structs are relevant to the Results object

type Results struct {
	RuleID    string  `json:"ruleId"`
	RuleIndex int     `json:"ruleIndex"`
	Level     string  `json:"level,omitempty"`
	Message   Message `json:"message"`
	/*Locations        []Location        `json:"locations"`
	CodeFlows        []CodeFlow        `json:"codeFlows"`
	RelatedLocations []RelatedLocation `json:"relatedLocations"`*/
	Properties SarifProperties `json:"properties"`
}

type Message struct {
	Text string `json:"text,omitempty"`
}

type SarifProperties struct {
	InstanceID        string `json:"InstanceID"`
	InstanceSeverity  string `json:"InstanceSeverity"`
	Confidence        string `json:"Confidence"`
	Audited           bool   `json:"Audited"`
	ToolAuditState    string `json:"ToolAuditState"`
	ToolAuditMessage  string `json:"ToolAuditMessage"`
	UnifiedAuditState string `json:"UnifiedAuditState"`
}

// These structs are relevant to the Tool object

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationUri string      `json:"informationUri,omitempty"`
	Rules          []SarifRule `json:"rules"`
	//SupportedTaxonomies []SupportedTaxonomies `json:"supportedTaxonomies"`
}

type SarifRule struct {
	Id                   string               `json:"id"`
	Guid                 string               `json:"guid"`
	Name                 string               `json:"name,omitempty"`
	ShortDescription     Message              `json:"shortDescription"`
	FullDescription      Message              `json:"fullDescription"`
	DefaultConfiguration DefaultConfiguration `json:"defaultConfiguration"`
	Relationships        []Relationships      `json:"relationships,omitempty"`
	Properties           *SarifRuleProperties `json:"properties,omitempty"`
}

type SupportedTaxonomies struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Guid  string `json:"guid"`
}

type DefaultConfiguration struct {
	Properties DefaultProperties `json:"properties"`
	Level      string            `json:"level,omitempty"` //This exists in the template, but not sure how it is populated. TODO.
}

type DefaultProperties struct {
	DefaultSeverity string `json:"DefaultSeverity"`
}

type Relationships struct {
	Target Target   `json:"target"`
	Kinds  []string `json:"kinds"`
}

type Target struct {
	Id            string        `json:"id"`
	ToolComponent ToolComponent `json:"toolComponent"`
}

type ToolComponent struct {
	Name string `json:"name"`
	Guid string `json:"guid"`
}

type SarifRuleProperties struct {
	Accuracy    string `json:"Accuracy,omitempty"`
	Impact      string `json:"Impact,omitempty"`
	Probability string `json:"Probability,omitempty"`
}

func ConvertFprToSarif(sys System, project *models.Project, projectVersion *models.ProjectVersion, resultFilePath string) error {
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

	err = Parse(sys, project, projectVersion, data)
	return err
}

func Parse(sys System, project *models.Project, projectVersion *models.ProjectVersion, data []byte) error {
	//To read XML data, Unmarshal or Decode can be used, here we use Decode to work on the stream
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	var fvdl FVDL
	decoder.Decode(&fvdl)

	//Now, we handle the sarif
	var sarif SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos01/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var fortifyRun Runs
	sarif.Runs = append(sarif.Runs, fortifyRun)

	// Handle results/vulnerabilities
	for i := 0; i < len(fvdl.Vulnerabilities.Vulnerability); i++ {
		result := *new(Results)
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
				result.Message = Message{rawMessage}
				break
			}
		}

		//handle properties
		prop := *new(SarifProperties)
		prop.InstanceSeverity = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceSeverity
		prop.Confidence = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.Confidence
		prop.InstanceID = fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceID
		//Use a query to get the audit data
		// B5C0FEFD-CCB2-4F21-A9D7-87AE600A5885 is "custom rules": handle differently?
		if result.RuleID == "B5C0FEFD-CCB2-4F21-A9D7-87AE600A5885" {
			// Custom Rules has no audit value: it's notificaiton in the FVDL only.
			prop.Audited = true
			prop.ToolAuditMessage = "Custom Rules: not a vuln"
			prop.ToolAuditState = "Not an Issue"
		} else if sys != nil {
			if err := prop.IntegrateAuditData(fvdl.Vulnerabilities.Vulnerability[i].InstanceInfo.InstanceID, sys, project, projectVersion); err != nil {
				log.Entry().Debug(err)
				prop.Audited = false
				prop.ToolAuditState = "Unknown"
				prop.ToolAuditMessage = "Error fetching audit state"
			}
		} else {
			prop.Audited = false
			prop.ToolAuditState = "Unknown"
			prop.ToolAuditMessage = "Cannot fetch audit state"
		}
		result.Properties = prop

		sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)
	}

	//handle the tool object
	tool := *new(Tool)
	tool.Driver = *new(Driver)
	tool.Driver.Name = "MicroFocus Fortify SCA"
	tool.Driver.Version = fvdl.EngineData.EngineVersion
	tool.Driver.InformationUri = "https://www.microfocus.com/documentation/fortify-static-code-analyzer-and-tools/2020/SCA_Guide_20.2.0.pdf"

	//handles rules
	for i := 0; i < len(fvdl.EngineData.RuleInfo); i++ { //i iterates on rules
		sarifRule := *new(SarifRule)
		sarifRule.Id = fvdl.EngineData.RuleInfo[i].RuleID
		sarifRule.Guid = fvdl.EngineData.RuleInfo[i].RuleID
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
				sarifRule.DefaultConfiguration.Properties.DefaultSeverity = fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.DefaultSeverity
				break
			}
		}
		//Descriptions
		for j := 0; j < len(fvdl.Description); j++ {
			if fvdl.Description[j].ClassID == sarifRule.Id {
				rawAbstract := fvdl.Description[j].Abstract.Text
				rawExplanation := fvdl.Description[j].Explanation.Text
				// Replacement defintions in abstract/explanation
				for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def); l++ {
					rawAbstract = strings.ReplaceAll(rawAbstract, "Replace key=\""+fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def[l].DefKey+"\"", fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def[l].DefValue)
					rawExplanation = strings.ReplaceAll(rawExplanation, "Replace key=\""+fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def[l].DefKey+"\"", fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.Def[l].DefValue)
				}
				// Replacement locationdef in explanation
				for l := 0; l < len(fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.LocationDef); l++ {
					rawExplanation = strings.ReplaceAll(rawExplanation, fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.LocationDef[l].Key, fvdl.Vulnerabilities.Vulnerability[i].AnalysisInfo.ReplacementDefinitions.LocationDef[l].Path)
				}
				// If Description has a CustomDescription, add it for good measure
				if fvdl.Description[j].CustomDescription.RuleID != "" {
					rawExplanation = rawExplanation + "\n;" + fvdl.Description[j].CustomDescription.Explanation.Text
				}
				sarifRule.ShortDescription.Text = rawAbstract
				sarifRule.FullDescription.Text = rawExplanation
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
		//scan for the properties we want:
		var propArray [][]string
		for j := 0; j < len(fvdl.EngineData.RuleInfo[i].MetaInfoGroup); j++ {
			if (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Accuracy") || (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Impact") || (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Probability") {
				propArray = append(propArray, []string{fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name, fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Data})
			}
		}
		var ruleProp *SarifRuleProperties
		if len(propArray) != 0 {
			ruleProp = new(SarifRuleProperties)
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

		//Finalize: append the rule
		tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)
	}
	//Finalize: tool
	sarif.Runs[0].Tool = tool

	//Edit the json
	sarifJSON, err := json.MarshalIndent(sarif, "", "  ") //Marshal as json to write file
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("target/audit.sarif", sarifJSON, 0700)
	return err
}

func (RuleProp *SarifProperties) IntegrateAuditData(issueInstanceID string, sys System, project *models.Project, projectVersion *models.ProjectVersion) error {
	data, err := sys.GetIssueDetails(projectVersion.ID, issueInstanceID)
	log.Entry().Debug("Looking up audit state of " + issueInstanceID)
	if err != nil {
		return err
	}
	if len(data) != 1 { //issueInstanceID is supposedly unique so len(data) = 1
		log.Entry().Error("not exactly 1 issue found, found " + fmt.Sprint(len(data)))
		return errors.New("not exactly 1 issue found, found " + fmt.Sprint(len(data)))
	}
	RuleProp.Audited = data[0].Audited
	if RuleProp.Audited {
		RuleProp.ToolAuditState = *data[0].PrimaryTag
	} else {
		RuleProp.ToolAuditState = "Unreviewed"
	}
	if *data[0].HasComments { //fetch latest message if comments exist
		//Fetch the ID
		parentID := data[0].ID
		commentData, err := sys.GetIssueComments(parentID)
		if err != nil {
			return err
		}
		RuleProp.ToolAuditMessage = *commentData[0].Comment
	}
	return nil
}
