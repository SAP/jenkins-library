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
	ScanID                   string       `xml:"ScanId,attr"`
	ProjectID                string       `xml:"ProjectId,attr"`
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

// CxxmlQuery CxxmlQuery
type CxxmlQuery struct {
	XMLName            xml.Name      `xml:"Query"`
	ID                 string        `xml:"id,attr"`
	Categories         string        `xml:"categories,attr"`
	CweID              string        `xml:"cweId,attr"`
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

// CxxmlResult CxxmlResult
type CxxmlResult struct {
	XMLName       xml.Name `xml:"Result"`
	NodeID        string   `xml:"NodeId,attr"`
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

// Path Path
type Path struct {
	XMLName           xml.Name   `xml:"Path"`
	ResultID          string     `xml:"ResultId,attr"`
	PathID            int        `xml:"PathId,attr"`
	SimilarityID      string     `xml:"SimilarityId,attr"`
	SourceMethod      string     `xml:"SourceMethod,attr"`
	DestinationMethod string     `xml:"DestinationMethod,attr"`
	PathNode          []PathNode `xml:"PathNode"`
}

// PathNode PathNode
type PathNode struct {
	XMLName  xml.Name `xml:"PathNode"`
	FileName string   `xml:"FileName"`
	Line     int      `xml:"Line"`
	Column   int      `xml:"Column"`
	NodeID   int      `xml:"NodeId"`
	Name     string   `xml:"Name"`
	Type     string   `xml:"Type"`
	Length   int      `xml:"Length"`
	Snippet  Snippet  `xml:"Snippet"`
}

// Snippet Snippet
type Snippet struct {
	XMLName xml.Name `xml:"Snippet"`
	Line    Line     `xml:"Line"`
}

// Line Line
type Line struct {
	XMLName xml.Name `xml:"Line"`
	Number  int      `xml:"Number"`
	Code    string   `xml:"Code"`
}

// ConvertCxxmlToSarif is the entrypoint for the Parse function
func ConvertCxxmlToSarif(xmlReportName string) (format.SARIF, error) {
	var sarif format.SARIF
	log.Entry().Debug("Reading audit file.")
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
	baseURL := "https://" + strings.Split(cxxml.DeepLink, "/")[2] + "/CxWebClient/ScanQueryDescription.aspx?"
	cweIdsForTaxonomies := make(map[string]int) //use a map to avoid duplicates
	cweCounter := 0

	//CxXML files contain a CxXMLResults > Query object, which represents a broken rule or type of vuln
	//This Query object contains a list of Result objects, each representing an occurence
	//Each Result object contains a ResultPath, which represents the exact location of the occurence (the "Snippet")
	log.Entry().Debug("[SARIF] Now handling results.")
	for i := 0; i < len(cxxml.Query); i++ {
		//add cweid to array
		cweIdsForTaxonomies[cxxml.Query[i].CweID] = cweCounter
		cweCounter = cweCounter + 1
		for j := 0; j < len(cxxml.Query[i].Result); j++ {
			result := *new(format.Results)

			//General
			result.RuleID = "checkmarx-" + cxxml.Query[i].ID
			result.RuleIndex = cweIdsForTaxonomies[cxxml.Query[i].CweID]
			result.Level = "none"
			msg := new(format.Message)
			//msg.Text = cxxml.Query[i].Name + ": " + cxxml.Query[i].Categories
			msg.Text = cxxml.Query[i].Name
			result.Message = msg
			//analysisTarget := new(format.ArtifactLocation)
			//analysisTarget.URI = cxxml.Query[i].Result[j].FileName
			//analysisTarget.Index = index

			//result.AnalysisTarget = analysisTarget
			if cxxml.Query[i].Name != "" {
				msg := new(format.Message)
				msg.Text = cxxml.Query[i].Name
			}
			//Locations
			for k := 0; k < len(cxxml.Query[i].Result[j].Path.PathNode); k++ {
				loc := *new(format.Location)
				/*index := 0
				//Check if that artifact has been added
				added := false
				for file := 0; file < len(sarif.Runs[0].Artifacts); file++ {
					if sarif.Runs[0].Artifacts[file].Location.Uri == cxxml.Query[i].Result[j].FileName {
						added = true
						index = file
						break
					}
				}
				if !added {
					artifact := format.Artifact{Location: format.SarifLocation{Uri: cxxml.Query[i].Result[j].FileName}}
					sarif.Runs[0].Artifacts = append(sarif.Runs[0].Artifacts, artifact)
					index = len(sarif.Runs[0].Artifacts) - 1
				}
				loc.PhysicalLocation.ArtifactLocation.Index = index */
				loc.PhysicalLocation.ArtifactLocation.URI = cxxml.Query[i].Result[j].FileName
				loc.PhysicalLocation.Region.StartLine = cxxml.Query[i].Result[j].Path.PathNode[k].Line
				loc.PhysicalLocation.Region.EndLine = cxxml.Query[i].Result[j].Path.PathNode[k].Line
				loc.PhysicalLocation.Region.StartColumn = cxxml.Query[i].Result[j].Path.PathNode[k].Column
				snip := new(format.SnippetSarif)
				snip.Text = cxxml.Query[i].Result[j].Path.PathNode[k].Snippet.Line.Code
				loc.PhysicalLocation.Region.Snippet = snip
				//loc.PhysicalLocation.ContextRegion.StartLine = cxxml.Query[i].Result[j].Path.PathNode[k].Line
				//loc.PhysicalLocation.ContextRegion.EndLine = cxxml.Query[i].Result[j].Path.PathNode[k].Line
				//loc.PhysicalLocation.ContextRegion.Snippet = snip
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

			result.PartialFingerprints.CheckmarxSimilarityID = cxxml.Query[i].Result[j].Path.SimilarityID
			result.PartialFingerprints.PrimaryLocationLineHash = cxxml.Query[i].Result[j].Path.SimilarityID

			//Properties
			props := new(format.SarifProperties)
			props.Audited = false
			if cxxml.Query[i].Result[j].Remark != "" {
				props.Audited = true
			}
			props.CheckmarxSimilarityID = cxxml.Query[i].Result[j].Path.SimilarityID
			props.InstanceID = cxxml.Query[i].Result[j].Path.ResultID + "-" + strconv.Itoa(cxxml.Query[i].Result[j].Path.PathID)
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
					candidate := strings.Split(remarks[cnd], "]: ")
					if len(candidate) == 1 {
						if len(candidate[0]) != 0 {
							messageCandidates = append([]string{strings.Trim(candidate[0], "\r\n")}, messageCandidates...)
						}
						continue
					} else if len(candidate) == 0 {
						continue
					}
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
		rule.ID = "checkmarx-" + cxxml.Query[i].ID
		words := strings.Split(cxxml.Query[i].Name, "_")
		for w := 0; w < len(words); w++ {
			words[w] = strings.Title(strings.ToLower(words[w]))
		}
		rule.Name = strings.Join(words, "")
		rule.HelpURI = baseURL + "queryID=" + cxxml.Query[i].ID + "&queryVersionCode=" + cxxml.Query[i].QueryVersionCode + "&queryTitle=" + cxxml.Query[i].Name
		rule.Help = new(format.Help)
		rule.Help.Text = rule.HelpURI
		rule.ShortDescription = new(format.Message)
		rule.ShortDescription.Text = cxxml.Query[i].Name
		if cxxml.Query[i].Categories != "" {
			rule.FullDescription = new(format.Message)
			rule.FullDescription.Text = cxxml.Query[i].Categories
		}
		rule.Properties = new(format.SarifRuleProperties)
		if cxxml.Query[i].CweID != "" {
			rule.Properties.Tags = append(rule.Properties.Tags, "external/cwe/cwe-"+cxxml.Query[i].CweID)
		}
		rulesArray = append(rulesArray, rule)
	}

	// Handle driver object
	log.Entry().Debug("[SARIF] Now handling driver object.")
	tool := *new(format.Tool)
	tool.Driver = *new(format.Driver)
	tool.Driver.Name = "Checkmarx SCA"
	versionData := strings.Split(cxxml.CheckmarxVersion, "V ")
	if len(versionData) > 1 { // Safety check
		tool.Driver.Version = strings.Split(cxxml.CheckmarxVersion, "V ")[1]
	} else {
		tool.Driver.Version = cxxml.CheckmarxVersion // Safe case
	}
	tool.Driver.InformationUri = "https://checkmarx.atlassian.net/wiki/spaces/KC/pages/1170245301/Navigating+Scan+Results+v9.0.0+to+v9.2.0"
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
