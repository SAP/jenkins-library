package checkmarx

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// CxXMLResults : This struct encapsulates everyting in the Cx XML document
type CxXMLResults struct {
	XMLName                  xml.Name     `xml:"CxXMLResults"`
	InitiatorName            string       `xml:"InitiatorName,attr"`
	Owner                    string       `xml:"Owner,attr"`
	ScanId                   string       `xml:"ScanId,attr"`
	ProjectId                string       `xml:"ProjectId,attr"`
	ProjectName              string       `xml:"ProjectName,attr"`
	TeamFullPathOnReportDate string       `xml:"TeamFullPathOnReportDate,attr"`
	DeepLink                 string       `xml:"DeepLink,attr"`
	ScanStart                string       `xml:"ScanStart,attr"`
	Preset                   string       `xml:"Preset,attr"`
	ScanTime                 string       `xml:"ScanTime,attr"`
	LinesOfCodeScanned       string       `xml:"LinesOfCodeScanned,attr"`
	FilesScanned             string       `xml:"FilesScanned,attr"`
	ReportCreationTime       string       `xml:"ReportCreationTime,attr"`
	Team                     string       `xml:"Team,attr"`
	CheckmarxVersion         string       `xml:"CheckmarxVersion,attr"`
	ScanComments             string       `xml:"ScanComments,attr"`
	ScanType                 string       `xml:"ScanType,attr"`
	SourceOrigin             string       `xml:"SourceOrigin,attr"`
	Visibility               string       `xml:"Visibility,attr"`
	Query                    []CxxmlQuery `xml:"Query"`
}

// Query
type CxxmlQuery struct {
	XMLName            xml.Name      `xml:"Query"`
	Id                 string        `xml:"id,attr"`
	Categories         string        `xml:"categories,attr"`
	CweId              string        `xml:"cweId,attr"`
	Name               string        `xml:"name,attr"`
	Group              string        `xml:"group,attr"`
	Severity           string        `xml:"Severity,attr"`
	Language           string        `xml:"Language,attr"`
	LanguageHash       string        `xml:"LanguageHash,attr"`
	LanguageChangeDate string        `xml:"LanguageChangeDate,attr"`
	SeverityIndex      int           `xml:"SeverityIndex,attr"`
	QueryPath          string        `xml:"QueryPath,attr"`
	QueryVersionCode   string        `xml:"QueryVersionCode,attr"`
	Result             []CxxmlResult `xml:"Result"`
}

// Result
type CxxmlResult struct {
	XMLName       xml.Name `xml:"Result"`
	NodeId        string   `xml:"NodeId,attr"`
	FileName      string   `xml:"FileName,attr"`
	Status        string   `xml:"Status,attr"`
	Line          int      `xml:"Line,attr"`
	Column        int      `xml:"Column,attr"`
	FalsePositive bool     `xml:"FalsePositive,attr"`
	Severity      string   `xml:"Severity,attr"`
	AssignToUser  string   `xml:"AssignToUser,attr"`
	State         int      `xml:"state,attr"`
	Remark        string   `xml:"Remark,attr"`
	DeepLink      string   `xml:"DeepLink,attr"`
	SeverityIndex int      `xml:"SeverityIndex,attr"`
	StatusIndex   int      `xml:"StatusIndex,attr"`
	DetectionDate string   `xml:"DetectionDate,attr"`
	Path          Path     `xml:"Path"`
}

// Path
type Path struct {
	XMLName           xml.Name   `xml:"Path"`
	ResultId          string     `xml:"ResultId,attr"`
	PathId            int        `xml:"PathId,attr"`
	SimilarityId      string     `xml:"SimilarityId,attr"`
	SourceMethod      string     `xml:"SourceMethod,attr"`
	DestinationMethod string     `xml:"DestinationMethod,attr"`
	PathNode          []PathNode `xml:"PathNode"`
}

// PathNode
type PathNode struct {
	XMLName  xml.Name `xml:"PathNode"`
	FileName string   `xml:"FileName"`
	Line     int      `xml:"Line"`
	Column   int      `xml:"Column"`
	NodeId   int      `xml:"NodeId"`
	Name     string   `xml:"Name"`
	Type     string   `xml:"Type"`
	Length   int      `xml:"Length"`
	Snippet  Snippet  `xml:"Snippet"`
}

// Snippet
type Snippet struct {
	XMLName xml.Name `xml:"Snippet"`
	Line    Line     `xml:"Line"`
}

// Line
type Line struct {
	XMLName xml.Name `xml:"Line"`
	Number  int      `xml:"Number"`
	Code    string   `xml:"Code"`
}

func ConvertCxxmlToSarif(xmlReportName string) (format.SARIF, error) {
	var sarif format.SARIF
	data, err := ioutil.ReadFile(xmlReportName)
	if err != nil {
		return sarif, err
	}
	if len(data) == 0 {
		log.Entry().Error("Error reading audit file at " + xmlReportName + ". This might be that the file is missing, corrupted, or too large. Aborting procedure.")
		err := errors.New("cannot read audit file")
		return sarif, err
	}

	log.Entry().Debug("Calling Parse.")
	return Parse(data)
}

// Parse function
func Parse(data []byte) (format.SARIF, error) {
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	var cxxml CxXMLResults
	err := decoder.Decode(&cxxml)
	if err != nil {
		return format.SARIF{}, err
	}

	// Process sarif
	var sarif format.SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var checkmarxRun format.Runs
	checkmarxRun.ColumnKind = "utf16CodeUnits"
	sarif.Runs = append(sarif.Runs, checkmarxRun)
	rulesArray := []format.SarifRule{}
	baseUrl := "https://" + strings.Split(cxxml.DeepLink, "/")[2] + "CxWebClient/ScanQueryDescription.aspx?"
	cweIdsForTaxonomies := make(map[string]int) //use a map to avoid duplicates
	cweCounter := 0

	//CxXML files contain a CxXMLResults > Query object, which represents a broken rule or type of vuln
	//This Query object contains a list of Result objects, each representing an occurence
	//Each Result object contains a ResultPath, which represents the exact location of the occurence (the "Snippet")
	for i := 0; i < len(cxxml.Query); i++ {
		//add cweid to array
		cweIdsForTaxonomies[cxxml.Query[i].CweId] = cweCounter
		cweCounter = cweCounter + 1
		for j := 0; j < len(cxxml.Query[i].Result); j++ {
			result := *new(format.Results)

			//General
			result.RuleID = cxxml.Query[i].Id
			result.RuleIndex = cweIdsForTaxonomies[cxxml.Query[i].CweId]
			result.Level = "none"
			msg := new(format.Message)
			msg.Text = cxxml.Query[i].Categories
			result.Message = msg
			analysisTarget := new(format.ArtifactLocation)
			analysisTarget.URI = cxxml.Query[i].Result[j].FileName
			result.AnalysisTarget = analysisTarget
			if cxxml.Query[i].Name != "" {
				msg := new(format.Message)
				msg.Text = cxxml.Query[i].Name
			}
			//Locations
			for k := 0; k < len(cxxml.Query[i].Result[j].Path.PathNode); k++ {
				loc := *new(format.Location)
				loc.PhysicalLocation.ArtifactLocation.URI = cxxml.Query[i].Result[j].FileName
				loc.PhysicalLocation.Region.StartLine = cxxml.Query[i].Result[j].Path.PathNode[k].Line
				snip := new(format.SnippetSarif)
				snip.Text = cxxml.Query[i].Result[j].Path.PathNode[k].Snippet.Line.Code
				loc.PhysicalLocation.Region.Snippet = snip
				loc.PhysicalLocation.ContextRegion.StartLine = cxxml.Query[i].Result[j].Path.PathNode[k].Line
				loc.PhysicalLocation.ContextRegion.EndLine = cxxml.Query[i].Result[j].Path.PathNode[k].Line
				loc.PhysicalLocation.ContextRegion.Snippet = snip
				result.Locations = append(result.Locations, loc)

				//Related Locations
				relatedLocation := *new(format.RelatedLocation)
				relatedLocation.ID = k + 1
				relatedLocation.PhysicalLocation = *new(format.RelatedPhysicalLocation)
				relatedLocation.PhysicalLocation.ArtifactLocation = loc.PhysicalLocation.ArtifactLocation
				relatedLocation.PhysicalLocation.Region = *new(format.RelatedRegion)
				relatedLocation.PhysicalLocation.Region.StartLine = loc.PhysicalLocation.Region.StartLine
				relatedLocation.PhysicalLocation.Region.StartColumn = cxxml.Query[i].Result[j].Path.PathNode[k].Column
				result.RelatedLocations = append(result.RelatedLocations, relatedLocation)

			}

			//Properties
			props := new(format.SarifProperties)
			props.Audited = false
			if cxxml.Query[i].Result[j].Remark != "" {
				props.Audited = true
			}
			props.CheckmarxSimilarityId = cxxml.Query[i].Result[j].Path.SimilarityId
			props.InstanceID = cxxml.Query[i].Result[j].Path.ResultId + "-" + strconv.Itoa(cxxml.Query[i].Result[j].Path.PathId)
			props.ToolSeverity = cxxml.Query[i].Result[j].Severity
			props.ToolSeverityIndex = cxxml.Query[i].Result[j].SeverityIndex
			props.ToolStateIndex = cxxml.Query[i].Result[j].State
			switch cxxml.Query[i].Result[j].State {
			case 1:
				props.ToolState = "NotExploitable"
				break
			case 2:
				props.ToolState = "Confirmed"
				break
			case 3:
				props.ToolState = "Urgent"
				break
			case 4:
				props.ToolState = "ProposedNotExploitable"
				break
			default:
				props.ToolState = "ToVerify" // Includes case 0
				break
			}
			props.ToolAuditMessage = ""
			if cxxml.Query[i].Result[j].Remark != "" {
				remarks := strings.Split(cxxml.Query[i].Result[j].Remark, "\n")
				messageCandidates := []string{}
				for cnd := 0; cnd < len(remarks); cnd++ {
					candidate := strings.Split(remarks[cnd], ": ")
					messageCandidates = append([]string{strings.Trim(candidate[1], "\r\n")}, messageCandidates...) //Append in reverse order, trim to remove extra \r
				}
				props.ToolAuditMessage = strings.Join(messageCandidates, " \n ")
			}
			props.UnifiedAuditState = ""
			result.Properties = props

			//Finalize
			sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)
		}

		//handle the rules array
		rule := *new(format.SarifRule)
		rule.ID = cxxml.Query[i].Id
		rule.Name = cxxml.Query[i].Name
		rule.HelpURI = baseUrl + "queryID=" + cxxml.Query[i].Id + "&queryVersionCode=" + cxxml.Query[i].QueryVersionCode + "&queryTitle=" + cxxml.Query[i].Name
		rulesArray = append(rulesArray, rule)
	}

	// Handle driver object
	tool := *new(format.Tool)
	tool.Driver = *new(format.Driver)
	tool.Driver.Name = "Checkmarx SCA"
	tool.Driver.Version = cxxml.CheckmarxVersion
	tool.Driver.Rules = rulesArray
	sarif.Runs[0].Tool = tool

	//handle automationDetails
	sarif.Runs[0].AutomationDetails.Id = cxxml.DeepLink // Use deeplink to pass a maximum of information

	//handle taxonomies
	//Only one exists apparently: CWE. It is fixed
	taxonomy := *new(format.Taxonomies)
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
