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
	Vulnerabilities []Vulnerability `xml:"Vulnerabilities>Vulnerability"`
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
	Trace                  Trace
}

type Context struct {
	Function Function
	FDSL     FunctionDeclarationSourceLocation
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
