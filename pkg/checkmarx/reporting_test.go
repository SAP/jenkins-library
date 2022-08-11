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
