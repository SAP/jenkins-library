package checkmarxOne

import (
	"fmt"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// ConvertCxJSONToSarif is the entrypoint for the Parse function
func ConvertCxJSONToSarif(sys System, serverURL string, scanResults *[]ScanResult, scanMeta *ScanMetadata, scan *Scan) (format.SARIF, error) {
	// Process sarif
	start := time.Now()

	var sarif format.SARIF
	sarif.Schema = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/cos02/schemas/sarif-schema-2.1.0.json"
	sarif.Version = "2.1.0"
	var checkmarxRun format.Runs
	checkmarxRun.ColumnKind = "utf16CodeUnits"
	checkmarxRun.Results = make([]format.Results, 0)
	sarif.Runs = append(sarif.Runs, checkmarxRun)
	rulesArray := []format.SarifRule{}

	baseURL := serverURL + "/results/" + scanMeta.ScanID + "/" + scanMeta.ProjectID

	cweIdsForTaxonomies := make(map[int]int) //use a map to avoid duplicates
	cweCounter := 0
	//maxretries := 5

	//JSON contains a ScanResultData > Query object, which represents a broken rule or type of vuln
	//This Query object contains a list of Result objects, each representing an occurence
	//Each Result object contains a ResultPath, which represents the exact location of the occurence (the "Snippet")
	log.Entry().Debug("[SARIF] Now handling results.")

	for _, r := range *scanResults {
		_, haskey := cweIdsForTaxonomies[r.VulnerabilityDetails.CweId]

		if !haskey {
			cweIdsForTaxonomies[r.VulnerabilityDetails.CweId] = cweCounter
			cweCounter++
		}

		simidString := fmt.Sprintf("%d", r.SimilarityID)

		var apiDescription string
		result := *new(format.Results)

		//General
		result.RuleID = fmt.Sprintf("checkmarxOne-%v/%d", r.Data.LanguageName, r.Data.QueryID)
		result.RuleIndex = cweIdsForTaxonomies[r.VulnerabilityDetails.CweId]
		result.Level = "none"
		msg := new(format.Message)
		if apiDescription != "" {
			msg.Text = apiDescription
		} else {
			msg.Text = r.Data.QueryName
		}
		result.Message = msg

		//Locations
		codeflow := *new(format.CodeFlow)
		threadflow := *new(format.ThreadFlow)
		locationSaved := false
		for k := 0; k < len(r.Data.Nodes); k++ {
			loc := *new(format.Location)
			loc.PhysicalLocation.ArtifactLocation.URI = r.Data.Nodes[0].FileName
			loc.PhysicalLocation.Region.StartLine = r.Data.Nodes[k].Line
			loc.PhysicalLocation.Region.EndLine = r.Data.Nodes[k].Line
			loc.PhysicalLocation.Region.StartColumn = r.Data.Nodes[k].Column
			snip := new(format.SnippetSarif)
			snip.Text = r.Data.Nodes[k].Name
			loc.PhysicalLocation.Region.Snippet = snip
			if !locationSaved { // To avoid overloading log file, we only save the 1st location, or source, as in the webview
				result.Locations = append(result.Locations, loc)
				locationSaved = true
			}

			//Related Locations
			relatedLocation := *new(format.RelatedLocation)
			relatedLocation.ID = k + 1
			relatedLocation.PhysicalLocation = *new(format.RelatedPhysicalLocation)
			relatedLocation.PhysicalLocation.ArtifactLocation = loc.PhysicalLocation.ArtifactLocation
			relatedLocation.PhysicalLocation.Region = *new(format.RelatedRegion)
			relatedLocation.PhysicalLocation.Region.StartLine = loc.PhysicalLocation.Region.StartLine
			relatedLocation.PhysicalLocation.Region.StartColumn = r.Data.Nodes[k].Column
			result.RelatedLocations = append(result.RelatedLocations, relatedLocation)

			threadFlowLocation := *new(format.Locations)
			tfloc := new(format.Location)
			tfloc.PhysicalLocation.ArtifactLocation.URI = r.Data.Nodes[0].FileName
			tfloc.PhysicalLocation.Region.StartLine = r.Data.Nodes[k].Line
			tfloc.PhysicalLocation.Region.EndLine = r.Data.Nodes[k].Line
			tfloc.PhysicalLocation.Region.StartColumn = r.Data.Nodes[k].Column
			tfloc.PhysicalLocation.Region.Snippet = snip
			threadFlowLocation.Location = tfloc
			threadflow.Locations = append(threadflow.Locations, threadFlowLocation)

		}
		codeflow.ThreadFlows = append(codeflow.ThreadFlows, threadflow)
		result.CodeFlows = append(result.CodeFlows, codeflow)

		result.PartialFingerprints.CheckmarxSimilarityID = simidString
		result.PartialFingerprints.PrimaryLocationLineHash = simidString

		//Properties
		props := new(format.SarifProperties)
		props.Audited = false
		props.CheckmarxSimilarityID = simidString
		props.InstanceID = r.ResultID // no more PathID in cx1
		props.ToolSeverity = r.Severity

		// classify into audit groups
		switch r.Severity {
		case "HIGH":
			props.AuditRequirement = format.AUDIT_REQUIREMENT_GROUP_1_DESC
			props.AuditRequirementIndex = format.AUDIT_REQUIREMENT_GROUP_1_INDEX
			props.ToolSeverityIndex = 3
			break
		case "MEDIUM":
			props.AuditRequirement = format.AUDIT_REQUIREMENT_GROUP_1_DESC
			props.AuditRequirementIndex = format.AUDIT_REQUIREMENT_GROUP_1_INDEX
			props.ToolSeverityIndex = 2
			break
		case "LOW":
			props.AuditRequirement = format.AUDIT_REQUIREMENT_GROUP_2_DESC
			props.AuditRequirementIndex = format.AUDIT_REQUIREMENT_GROUP_2_INDEX
			props.ToolSeverityIndex = 1
			break
		case "INFORMATION":
			props.AuditRequirement = format.AUDIT_REQUIREMENT_GROUP_3_DESC
			props.AuditRequirementIndex = format.AUDIT_REQUIREMENT_GROUP_3_INDEX
			props.ToolSeverityIndex = 0
			break
		}

		switch r.State {
		case "NOT_EXPLOITABLE":
			props.ToolState = "NOT_EXPLOITABLE"
			props.ToolStateIndex = 1
			props.Audited = true
			break
		case "CONFIRMED":
			props.ToolState = "CONFIRMED"
			props.ToolStateIndex = 2
			props.Audited = true
			break
		case "URGENT", "URGENT ":
			props.ToolState = "URGENT"
			props.ToolStateIndex = 3
			props.Audited = true
			break
		case "PROPOSED_NOT_EXPLOITABLE":
			props.ToolState = "PROPOSED_NOT_EXPLOITABLE"
			props.ToolStateIndex = 4
			props.Audited = true
			break
		default:
			props.ToolState = "TO_VERIFY" // Includes case 0
			props.ToolStateIndex = 0

			break
		}

		props.ToolAuditMessage = ""
		// currently disabled due to the extra load (one api call per finding)
		/*predicates, err := sys.GetResultsPredicates(r.SimilarityID, scanMeta.ProjectID)
		if err == nil {
			log.Entry().Infof("Retrieved %d results predicates", len(predicates))
			messageCandidates := []string{}
			for _, p := range predicates {
				messageCandidates = append([]string{strings.Trim(p.Comment, "\r\n")}, messageCandidates...) //Append in reverse order, trim to remove extra \r
			}
			log.Entry().Info(strings.Join(messageCandidates, "; "))
			props.ToolAuditMessage = strings.Join(messageCandidates, " \n ")
		} else {
			log.Entry().Warningf("Error while retrieving result predicates: %s", err)
		}*/

		props.RuleGUID = fmt.Sprintf("%d", r.Data.QueryID)
		props.UnifiedAuditState = ""
		result.Properties = props

		//Finalize
		sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)

		//handle the rules array
		rule := *new(format.SarifRule)

		rule.ID = fmt.Sprintf("checkmarxOne-%v/%d", r.Data.LanguageName, r.Data.QueryID)
		words := strings.Split(r.Data.QueryName, "_")
		for w := 0; w < len(words); w++ {
			words[w] = piperutils.Title(strings.ToLower(words[w]))
		}
		rule.Name = strings.Join(words, "")

		rule.HelpURI = fmt.Sprintf("%v/sast/description/%v/%v", baseURL, r.VulnerabilityDetails.CweId, r.Data.QueryID)
		rule.Help = new(format.Help)
		rule.Help.Text = rule.HelpURI
		rule.ShortDescription = new(format.Message)
		rule.ShortDescription.Text = r.Data.QueryName
		rule.Properties = new(format.SarifRuleProperties)

		if len(r.VulnerabilityDetails.Compliances) > 0 {
			rule.FullDescription = new(format.Message)
			rule.FullDescription.Text = strings.Join(r.VulnerabilityDetails.Compliances[:], ";")

			for cat := 0; cat < len(r.VulnerabilityDetails.Compliances); cat++ {
				rule.Properties.Tags = append(rule.Properties.Tags, r.VulnerabilityDetails.Compliances[cat])
			}
		}
		switch r.Severity {
		case "INFORMATION":
			rule.Properties.SecuritySeverity = "0.0"
		case "LOW":
			rule.Properties.SecuritySeverity = "2.0"
		case "MEDIUM":
			rule.Properties.SecuritySeverity = "5.0"
		case "HIGH":
			rule.Properties.SecuritySeverity = "7.0"
		default:
			rule.Properties.SecuritySeverity = "10.0"
		}

		if r.VulnerabilityDetails.CweId != 0 {
			rule.Properties.Tags = append(rule.Properties.Tags, fmt.Sprintf("external/cwe/cwe-%d", r.VulnerabilityDetails.CweId))
		}

		match := false
		for _, r := range rulesArray {
			if r.ID == rule.ID {
				match = true
				break
			}
		}
		if !match {
			rulesArray = append(rulesArray, rule)
		}
	}

	// Handle driver object
	log.Entry().Debug("[SARIF] Now handling driver object.")
	tool := *new(format.Tool)
	tool.Driver = *new(format.Driver)
	tool.Driver.Name = "CheckmarxOne SCA"

	// TODO: a way to fetch/store the version
	tool.Driver.Version = "1" //strings.Split(cxxml.CheckmarxVersion, "V ")
	tool.Driver.InformationUri = "https://checkmarx.com/resource/documents/en/34965-165898-results-details-per-scanner.html"
	tool.Driver.Rules = rulesArray
	sarif.Runs[0].Tool = tool

	//handle automationDetails
	sarif.Runs[0].AutomationDetails = &format.AutomationDetails{Id: fmt.Sprintf("%v/sast", baseURL)} // Use deeplink to pass a maximum of information

	//handle taxonomies
	//Only one exists apparently: CWE. It is fixed
	taxonomy := *new(format.Taxonomies)
	taxonomy.Name = "CWE"
	taxonomy.Organization = "MITRE"
	taxonomy.ShortDescription.Text = "The MITRE Common Weakness Enumeration"
	for key := range cweIdsForTaxonomies {
		taxa := *new(format.Taxa)
		taxa.Id = fmt.Sprintf("%d", key)
		taxonomy.Taxa = append(taxonomy.Taxa, taxa)
	}
	sarif.Runs[0].Taxonomies = append(sarif.Runs[0].Taxonomies, taxonomy)

	// Add a conversion object to highlight this isn't native SARIF
	conversion := new(format.Conversion)
	conversion.Tool.Driver.Name = "Piper CheckmarxOne JSON to SARIF converter"
	conversion.Tool.Driver.InformationUri = "https://github.com/SAP/jenkins-library"
	conversion.Invocation.ExecutionSuccessful = true
	conversion.Invocation.StartTimeUtc = fmt.Sprintf("%s", start.Format("2006-01-02T15:04:05.000Z")) // "YYYY-MM-DDThh:mm:ss.sZ" on 2006-01-02 15:04:05
	conversion.Invocation.Account = scan.Initiator
	sarif.Runs[0].Conversion = conversion

	return sarif, nil
}

func getQuery(queries []Query, queryID uint64) *Query {
	for id := range queries {
		if queries[id].QueryID == queryID {
			return &queries[id]
		}
	}
	return nil
}
