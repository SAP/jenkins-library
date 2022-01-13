package fortify

import (
	"bytes"
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
	Buildinfo       Build
	Vulnerabilities Vulnerabilities `xml:"Vulnerabilities"`
	ContextPool     ContextPool     `xml:"ContextPool"`
	UnifiedNodePool UnifiedNodePool `xml:"UnifiedNodePool"`
	Description     []Description   `xml:"Description"`
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
	FileSize      string   `xml:"size,attr"`
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
	XMLName      xml.Name `xml:"Vulnerability"`
	ClassInfo    ClassInfo
	InstanceInfo InstanceInfo
	AnalysisInfo AnalysisInfo `xml:"AnalysisInfo>Unified"`
}

type ClassInfo struct {
	XMLName         xml.Name `xml:"ClassInfo"`
	ClassID         string   `xml:"ClassID"`
	Kingdom         string   `xml:"Kingdom"`
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
	ReplacementDefinitions []Def `xml:"ReplacementDefinitions>Def"`
	Trace                  Trace `xml:"Trace"`
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
	RefId   string   `xml:"id,attr"`
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
	Line      string   `xml:"line,attr"`
	LineEnd   string   `xml:"lineEnd,attr"`
	ColStart  string   `xml:"colStart,attr"`
	ColEnd    string   `xml:"colEnd,attr"`
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
	//To read XML data, Unmarshal or Decode can be used. However, Unmarshal is not well-behaved when there are
	//multiple different XML tree roots. This is why a decoder is created from a reader, which allows us to
	//simply run Decode and get all well-formatted XML data for one type.
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	var result FVDL
	decoder.Decode(&result)
	return nil
}
