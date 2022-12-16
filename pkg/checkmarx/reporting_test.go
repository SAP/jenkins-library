package checkmarx

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateJSONReport(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<CxXMLResults InitiatorName="admin" Owner="admin" ScanId="1000005" ProjectId="2" ProjectName="Project 1" TeamFullPathOnReportDate="CxServer" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2" ScanStart="Sunday, December 3, 2017 4:50:34 PM" Preset="Checkmarx Default" ScanTime="00h:03m:18s" LinesOfCodeScanned="6838" FilesScanned="34" ReportCreationTime="Sunday, December 3, 2017 6:13:45 PM" Team="CxServer" CheckmarxVersion="8.6.0" ScanComments="" ScanType="Incremental" SourceOrigin="LocalPath" Visibility="Public">
	<Query id="430" categories="PCI DSS v3.2;PCI DSS (3.2) - 6.5.1 - Injection flaws - particularly SQL injection,OWASP Top 10 2013;A1-Injection,FISMA 2014;System And Information Integrity,NIST SP 800-53;SI-10 Information Input Validation (P1),OWASP Top 10 2017;A1-Injection" cweId="89" name="SQL_Injection" group="CSharp_High_Risk" Severity="High" Language="CSharp" LanguageHash="1363215419077432" LanguageChangeDate="2017-12-03T00:00:00.0000000" SeverityIndex="3" QueryPath="CSharp\Cx\CSharp High Risk\SQL Injection Version:0" QueryVersionCode="430">
	</Query>
	</CxXMLResults>`

	var xmlResult DetailedResult
	xml.Unmarshal([]byte(data), &xmlResult)
	resultMap := map[string]interface{}{}
	resultMap["InitiatorName"] = xmlResult.InitiatorName
	resultMap["Owner"] = xmlResult.Owner
	resultMap["ScanId"] = xmlResult.ScanID
	resultMap["ProjectId"] = xmlResult.ProjectID
	resultMap["ProjectName"] = xmlResult.ProjectName
	resultMap["Team"] = xmlResult.Team
	resultMap["TeamFullPathOnReportDate"] = xmlResult.TeamFullPathOnReportDate
	resultMap["ScanStart"] = xmlResult.ScanStart
	resultMap["ScanTime"] = xmlResult.ScanTime
	resultMap["LinesOfCodeScanned"] = xmlResult.LinesOfCodeScanned
	resultMap["FilesScanned"] = xmlResult.FilesScanned
	resultMap["CheckmarxVersion"] = xmlResult.CheckmarxVersion
	resultMap["ScanType"] = xmlResult.ScanType
	resultMap["Preset"] = xmlResult.Preset
	resultMap["DeepLink"] = xmlResult.DeepLink
	resultMap["ReportCreationTime"] = xmlResult.ReportCreationTime
	resultMap["High"] = map[string]int{}
	resultMap["Medium"] = map[string]int{}
	resultMap["Low"] = map[string]int{}
	resultMap["Information"] = map[string]int{}
	submap := map[string]int{}
	submap["Issues"] = 10
	submap["NotFalsePositive"] = 10
	resultMap["High"] = submap

	submap = map[string]int{}
	submap["Issues"] = 4
	submap["NotFalsePositive"] = 0
	resultMap["Medium"] = submap

	submap = map[string]int{}
	submap["Issues"] = 2
	submap["NotFalsePositive"] = 2
	submap["Confirmed"] = 1
	submap["NotExploitable"] = 1
	resultMap["Low"] = submap

	submap = map[string]int{}
	submap["Issues"] = 5
	submap["NotFalsePositive"] = 5
	resultMap["Information"] = submap

	lowPerQuery := map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 4
	submap["Confirmed"] = 0
	submap["NotExploitable"] = 0
	lowPerQuery["Low_Query_Name_1"] = submap

	submap = map[string]int{}
	submap["Issues"] = 5
	submap["Confirmed"] = 2
	submap["NotExploitable"] = 3
	lowPerQuery["Low_Query_Name_2"] = submap

	resultMap["LowPerQuery"] = lowPerQuery

	reportingData := CreateJSONReport(resultMap)
	assert.Equal(t, int64(1000005), reportingData.ScanID)
	assert.Equal(t, "Project 1", reportingData.ProjectName)
	assert.Equal(t, int64(2), reportingData.ProjectID)
	assert.Equal(t, "CxServer", reportingData.TeamName)
	assert.Equal(t, "checkmarx", reportingData.ToolName)
	assert.Equal(t, "CxServer", reportingData.TeamPath)
	assert.Equal(t, "http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&projectid=2", reportingData.DeepLink)
	assert.Equal(t, "Checkmarx Default", reportingData.Preset)
	assert.Equal(t, "8.6.0", reportingData.CheckmarxVersion)
	assert.Equal(t, "Incremental", reportingData.ScanType)
	assert.Equal(t, 10, reportingData.HighTotal)
	assert.Equal(t, 0, reportingData.HighAudited)
	assert.Equal(t, 4, reportingData.MediumTotal)
	assert.Equal(t, 4, reportingData.MediumAudited)
	assert.Equal(t, 2, reportingData.LowTotal)
	assert.Equal(t, 2, reportingData.LowAudited)
	assert.Equal(t, 5, reportingData.InformationTotal)
	assert.Equal(t, 0, reportingData.InformationAudited)
	assert.Equal(t, false, reportingData.IsLowPerQueryAudited)
	assert.Equal(t, 2, len(*reportingData.LowPerQuery))
	if (*reportingData.LowPerQuery)[0].QueryName == "Low_Query_Name_1" {
		assert.Equal(t, "Low_Query_Name_1", (*reportingData.LowPerQuery)[0].QueryName)
		assert.Equal(t, 0, (*reportingData.LowPerQuery)[0].Audited)
		assert.Equal(t, 4, (*reportingData.LowPerQuery)[0].Total)
		assert.Equal(t, "Low_Query_Name_2", (*reportingData.LowPerQuery)[1].QueryName)
		assert.Equal(t, 5, (*reportingData.LowPerQuery)[1].Audited)
		assert.Equal(t, 5, (*reportingData.LowPerQuery)[1].Total)
	} else {
		assert.Equal(t, "Low_Query_Name_1", (*reportingData.LowPerQuery)[1].QueryName)
		assert.Equal(t, 0, (*reportingData.LowPerQuery)[1].Audited)
		assert.Equal(t, 4, (*reportingData.LowPerQuery)[1].Total)
		assert.Equal(t, "Low_Query_Name_2", (*reportingData.LowPerQuery)[0].QueryName)
		assert.Equal(t, 5, (*reportingData.LowPerQuery)[0].Audited)
		assert.Equal(t, 5, (*reportingData.LowPerQuery)[0].Total)
	}

	lowPerQuery = map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 100
	submap["Confirmed"] = 10
	submap["NotExploitable"] = 0
	lowPerQuery["Low_Query_Name_1"] = submap

	submap = map[string]int{}
	submap["Issues"] = 5
	submap["Confirmed"] = 2
	submap["NotExploitable"] = 3
	lowPerQuery["Low_Query_Name_2"] = submap

	resultMap["LowPerQuery"] = lowPerQuery
	reportingData = CreateJSONReport(resultMap)
	assert.Equal(t, true, reportingData.IsLowPerQueryAudited)

	lowPerQuery = map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 200
	submap["Confirmed"] = 3
	submap["NotExploitable"] = 2
	lowPerQuery["Low_Query_Name_1"] = submap

	resultMap["LowPerQuery"] = lowPerQuery
	reportingData = CreateJSONReport(resultMap)
	assert.Equal(t, false, reportingData.IsLowPerQueryAudited)

	lowPerQuery = map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 200
	submap["Confirmed"] = 5
	submap["NotExploitable"] = 5
	lowPerQuery["Low_Query_Name_1"] = submap

	resultMap["LowPerQuery"] = lowPerQuery
	reportingData = CreateJSONReport(resultMap)
	assert.Equal(t, true, reportingData.IsLowPerQueryAudited)
}

func TestJsonReportWithNoLowVulnData(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<CxXMLResults InitiatorName="admin" Owner="admin" ScanId="1000005" ProjectId="2" ProjectName="Project 1" TeamFullPathOnReportDate="CxServer" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2" ScanStart="Sunday, December 3, 2017 4:50:34 PM" Preset="Checkmarx Default" ScanTime="00h:03m:18s" LinesOfCodeScanned="6838" FilesScanned="34" ReportCreationTime="Sunday, December 3, 2017 6:13:45 PM" Team="CxServer" CheckmarxVersion="8.6.0" ScanComments="" ScanType="Incremental" SourceOrigin="LocalPath" Visibility="Public">
	<Query id="430" categories="PCI DSS v3.2;PCI DSS (3.2) - 6.5.1 - Injection flaws - particularly SQL injection,OWASP Top 10 2013;A1-Injection,FISMA 2014;System And Information Integrity,NIST SP 800-53;SI-10 Information Input Validation (P1),OWASP Top 10 2017;A1-Injection" cweId="89" name="SQL_Injection" group="CSharp_High_Risk" Severity="High" Language="CSharp" LanguageHash="1363215419077432" LanguageChangeDate="2017-12-03T00:00:00.0000000" SeverityIndex="3" QueryPath="CSharp\Cx\CSharp High Risk\SQL Injection Version:0" QueryVersionCode="430">
	</Query>
	</CxXMLResults>`

	var xmlResult DetailedResult
	xml.Unmarshal([]byte(data), &xmlResult)
	resultMap := map[string]interface{}{}
	resultMap["InitiatorName"] = xmlResult.InitiatorName
	resultMap["Owner"] = xmlResult.Owner
	resultMap["ScanId"] = xmlResult.ScanID
	resultMap["ProjectId"] = xmlResult.ProjectID
	resultMap["ProjectName"] = xmlResult.ProjectName
	resultMap["Team"] = xmlResult.Team
	resultMap["TeamFullPathOnReportDate"] = xmlResult.TeamFullPathOnReportDate
	resultMap["ScanStart"] = xmlResult.ScanStart
	resultMap["ScanTime"] = xmlResult.ScanTime
	resultMap["LinesOfCodeScanned"] = xmlResult.LinesOfCodeScanned
	resultMap["FilesScanned"] = xmlResult.FilesScanned
	resultMap["CheckmarxVersion"] = xmlResult.CheckmarxVersion
	resultMap["ScanType"] = xmlResult.ScanType
	resultMap["Preset"] = xmlResult.Preset
	resultMap["DeepLink"] = xmlResult.DeepLink
	resultMap["ReportCreationTime"] = xmlResult.ReportCreationTime
	resultMap["High"] = map[string]int{}
	resultMap["Medium"] = map[string]int{}
	resultMap["Low"] = map[string]int{}
	resultMap["Information"] = map[string]int{}
	submap := map[string]int{}
	submap["Issues"] = 10
	submap["NotFalsePositive"] = 10
	resultMap["High"] = submap

	submap = map[string]int{}
	submap["Issues"] = 4
	submap["NotFalsePositive"] = 4
	resultMap["Medium"] = submap

	submap = map[string]int{}
	submap["Issues"] = 2
	submap["NotFalsePositive"] = 1
	resultMap["Information"] = submap

	reportingData := CreateJSONReport(resultMap)
	assert.Equal(t, int64(1000005), reportingData.ScanID)
	assert.Equal(t, "Project 1", reportingData.ProjectName)
	assert.Equal(t, int64(2), reportingData.ProjectID)
	assert.Equal(t, "CxServer", reportingData.TeamName)
	assert.Equal(t, "checkmarx", reportingData.ToolName)
	assert.Equal(t, "CxServer", reportingData.TeamPath)
	assert.Equal(t, "http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&projectid=2", reportingData.DeepLink)
	assert.Equal(t, "Checkmarx Default", reportingData.Preset)
	assert.Equal(t, "8.6.0", reportingData.CheckmarxVersion)
	assert.Equal(t, "Incremental", reportingData.ScanType)
	assert.Equal(t, 10, reportingData.HighTotal)
	assert.Equal(t, 0, reportingData.HighAudited)
	assert.Equal(t, 4, reportingData.MediumTotal)
	assert.Equal(t, 0, reportingData.MediumAudited)
	assert.Equal(t, 0, reportingData.LowTotal)
	assert.Equal(t, 0, reportingData.LowAudited)
	assert.Equal(t, 2, reportingData.InformationTotal)
	assert.Equal(t, 0, reportingData.InformationAudited)
}

func TestCreateCustomReport(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
	<CxXMLResults InitiatorName="admin" Owner="admin" ScanId="1000005" ProjectId="2" ProjectName="Project 1" TeamFullPathOnReportDate="CxServer" DeepLink="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&amp;projectid=2" ScanStart="Sunday, December 3, 2017 4:50:34 PM" Preset="Checkmarx Default" ScanTime="00h:03m:18s" LinesOfCodeScanned="6838" FilesScanned="34" ReportCreationTime="Sunday, December 3, 2017 6:13:45 PM" Team="CxServer" CheckmarxVersion="8.6.0" ScanComments="" ScanType="Incremental" SourceOrigin="LocalPath" Visibility="Public">
	<Query id="430" categories="PCI DSS v3.2;PCI DSS (3.2) - 6.5.1 - Injection flaws - particularly SQL injection,OWASP Top 10 2013;A1-Injection,FISMA 2014;System And Information Integrity,NIST SP 800-53;SI-10 Information Input Validation (P1),OWASP Top 10 2017;A1-Injection" cweId="89" name="SQL_Injection" group="CSharp_High_Risk" Severity="High" Language="CSharp" LanguageHash="1363215419077432" LanguageChangeDate="2017-12-03T00:00:00.0000000" SeverityIndex="3" QueryPath="CSharp\Cx\CSharp High Risk\SQL Injection Version:0" QueryVersionCode="430">
	</Query>
	</CxXMLResults>`

	var xmlResult DetailedResult
	xml.Unmarshal([]byte(data), &xmlResult)
	resultMap := map[string]interface{}{}
	resultMap["InitiatorName"] = xmlResult.InitiatorName
	resultMap["Owner"] = xmlResult.Owner
	resultMap["ScanId"] = xmlResult.ScanID
	resultMap["ProjectId"] = xmlResult.ProjectID
	resultMap["ProjectName"] = xmlResult.ProjectName
	resultMap["Team"] = xmlResult.Team
	resultMap["TeamFullPathOnReportDate"] = xmlResult.TeamFullPathOnReportDate
	resultMap["ScanStart"] = xmlResult.ScanStart
	resultMap["ScanTime"] = xmlResult.ScanTime
	resultMap["LinesOfCodeScanned"] = xmlResult.LinesOfCodeScanned
	resultMap["FilesScanned"] = xmlResult.FilesScanned
	resultMap["CheckmarxVersion"] = xmlResult.CheckmarxVersion
	resultMap["ScanType"] = xmlResult.ScanType
	resultMap["Preset"] = xmlResult.Preset
	resultMap["DeepLink"] = xmlResult.DeepLink
	resultMap["ReportCreationTime"] = xmlResult.ReportCreationTime
	resultMap["High"] = map[string]int{}
	resultMap["Medium"] = map[string]int{}
	resultMap["Low"] = map[string]int{}
	resultMap["Information"] = map[string]int{}
	submap := map[string]int{}
	submap["Issues"] = 10
	submap["NotFalsePositive"] = 10
	resultMap["High"] = submap

	submap = map[string]int{}
	submap["Issues"] = 4
	submap["NotFalsePositive"] = 0
	resultMap["Medium"] = submap

	submap = map[string]int{}
	submap["Issues"] = 2
	submap["NotFalsePositive"] = 2
	submap["Confirmed"] = 1
	submap["NotExploitable"] = 1
	resultMap["Low"] = submap

	submap = map[string]int{}
	submap["Issues"] = 5
	submap["NotFalsePositive"] = 5
	resultMap["Information"] = submap

	lowPerQuery := map[string]map[string]int{}
	submap = map[string]int{}
	submap["Issues"] = 4
	submap["Confirmed"] = 0
	submap["NotExploitable"] = 0
	lowPerQuery["Low_Query_Name_1"] = submap

	submap = map[string]int{}
	submap["Issues"] = 5
	submap["Confirmed"] = 2
	submap["NotExploitable"] = 3
	lowPerQuery["Low_Query_Name_2"] = submap

	resultMap["LowPerQuery"] = lowPerQuery

	insecure := []string{"insecure"}
	neutral := []string{"neutral"}

	reportingData := CreateCustomReport(resultMap, insecure, neutral)
	assert.Equal(t, "Checkmarx SAST Report", reportingData.ReportTitle)
	assert.Equal(t, 15, len(reportingData.Subheaders))
	assert.Equal(t, 2, len(reportingData.Overview))

	subheaders := make(map[string]string)
	for _, subheader := range reportingData.Subheaders {
		subheaders[subheader.Description] = subheader.Details
	}
	assert.Equal(t, "Project 1", subheaders["Project name"])
	assert.Equal(t, "2", subheaders["Project ID"])
	assert.Equal(t, "admin", subheaders["Owner"])
	assert.Equal(t, "1000005", subheaders["Scan ID"])
	assert.Equal(t, "CxServer", subheaders["Team"])
	assert.Equal(t, "CxServer", subheaders["Team full path"])
	assert.Equal(t, "Sunday, December 3, 2017 4:50:34 PM", subheaders["Scan start"])
	assert.Equal(t, "00h:03m:18s", subheaders["Scan duration"])
	assert.Equal(t, "Incremental", subheaders["Scan type"])
	assert.Equal(t, "Checkmarx Default", subheaders["Preset"])
	assert.Equal(t, "Sunday, December 3, 2017 6:13:45 PM", subheaders["Report creation time"])
	assert.Equal(t, "6838", subheaders["Lines of code scanned"])
	assert.Equal(t, "34", subheaders["Files scanned"])
	assert.Equal(t, "8.6.0", subheaders["Checkmarx version"])
	assert.Equal(t, `<a href="http://WIN2K12-TEMP/CxWebClient/ViewerMain.aspx?scanid=1000005&projectid=2" target="_blank">Link to scan in CX UI</a>`, subheaders["Deep link"])

	detailRows := make(map[string]string)
	for _, detailRow := range reportingData.DetailTable.Rows {
		detailRows[detailRow.Columns[0].Content] = detailRow.Columns[1].Content
	}
	assert.Equal(t, "10", detailRows["High issues"])
	assert.Equal(t, "10", detailRows["High not false positive issues"])
	assert.Equal(t, "0", detailRows["High not exploitable issues"])
	assert.Equal(t, "0", detailRows["High confirmed issues"])
	assert.Equal(t, "0", detailRows["High urgent issues"])
	assert.Equal(t, "0", detailRows["High proposed not exploitable issues"])
	assert.Equal(t, "0", detailRows["High to verify issues"])
	assert.Equal(t, "4", detailRows["Medium issues"])
	assert.Equal(t, "0", detailRows["Medium not false positive issues"])
	assert.Equal(t, "0", detailRows["Medium not exploitable issues"])
	assert.Equal(t, "0", detailRows["Medium confirmed issues"])
	assert.Equal(t, "0", detailRows["Medium urgent issues"])
	assert.Equal(t, "0", detailRows["Medium proposed not exploitable issues"])
	assert.Equal(t, "0", detailRows["Medium to verify issues"])
	assert.Equal(t, "2", detailRows["Low issues"])
	assert.Equal(t, "2", detailRows["Low not false positive issues"])
	assert.Equal(t, "1", detailRows["Low not exploitable issues"])
	assert.Equal(t, "1", detailRows["Low confirmed issues"])
	assert.Equal(t, "0", detailRows["Low urgent issues"])
	assert.Equal(t, "0", detailRows["Low proposed not exploitable issues"])
	assert.Equal(t, "0", detailRows["Low to verify issues"])
	assert.Equal(t, "5", detailRows["Informational issues"])
	assert.Equal(t, "5", detailRows["Informational not false positive issues"])
	assert.Equal(t, "0", detailRows["Informational not exploitable issues"])
	assert.Equal(t, "0", detailRows["Informational confirmed issues"])
	assert.Equal(t, "0", detailRows["Informational urgent issues"])
	assert.Equal(t, "0", detailRows["Informational proposed not exploitable issues"])
	assert.Equal(t, "0", detailRows["Informational to verify issues"])
}
