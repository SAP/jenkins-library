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

	return Parse(sys, project, projectVersion, data)
}

func Parse(sys System, project *models.Project, projectVersion *models.ProjectVersion, data []byte) (format.SARIF, error) {
	//To read XML data, Unmarshal or Decode can be used, here we use Decode to work on the stream
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	var fvdl FVDL
	decoder.Decode(&fvdl)

	//Now, we handle the sarif
	var sarif format.SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos01/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var fortifyRun format.Runs
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
				result.Message = format.Message{rawMessage}
				break
			}
		}

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
				sarifRule.DefaultConfiguration.Properties.DefaultSeverity = fvdl.Vulnerabilities.Vulnerability[j].ClassInfo.DefaultSeverity
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
						sarifRule.ShortDescription.Text = rawAbstract
						sarifRule.FullDescription.Text = rawExplanation
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
		//scan for the properties we want:
		var propArray [][]string
		for j := 0; j < len(fvdl.EngineData.RuleInfo[i].MetaInfoGroup); j++ {
			if (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Accuracy") || (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Impact") || (fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name == "Probability") {
				propArray = append(propArray, []string{fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Name, fvdl.EngineData.RuleInfo[i].MetaInfoGroup[j].Data})
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

		//Finalize: append the rule
		tool.Driver.Rules = append(tool.Driver.Rules, sarifRule)
	}
	//Finalize: tool
	sarif.Runs[0].Tool = tool

	return sarif, nil
}

func integrateAuditData(ruleProp *format.SarifProperties, issueInstanceID string, sys System, project *models.Project, projectVersion *models.ProjectVersion) error {
	data, err := sys.GetIssueDetails(projectVersion.ID, issueInstanceID)
	log.Entry().Debug("Looking up audit state of " + issueInstanceID)
	if err != nil {
		return err
	}
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
