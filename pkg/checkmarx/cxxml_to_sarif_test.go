//go:build unit
// +build unit

package checkmarx

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/format"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {

	//Use a test CXXML doc
	testCxxml := `
<?xml version="1.0" encoding="utf-8"?>
<CxXMLResults InitiatorName="Test" Owner="Tester" ScanId="1111111" ProjectId="11037" ProjectName="test-project" TeamFullPathOnReportDate="CxServer" DeepLink="https://cxtext.test/CxWebClient/ViewerMain.aspx?scanid=1111111&amp;projectid=11037" ScanStart="Monday, March 7, 2022 1:58:49 PM" Preset="Checkmarx Default" ScanTime="00h:00m:22s" LinesOfCodeScanned="2682" FilesScanned="15" ReportCreationTime="Monday, March 7, 2022 1:59:25 PM" Team="SecurityTesting" CheckmarxVersion="V 9.4.3" ScanComments="Scan From Golang Script" ScanType="Incremental" SourceOrigin="LocalPath" Visibility="Public">
	<Query id="2415" categories="Dummy Categories" cweId="79" name="Dummy Vuln 1" group="JavaScript_High_Risk" Severity="High" Language="JavaScript" LanguageHash="9095271965336651" LanguageChangeDate="2022-01-16T00:00:00.0000000" SeverityIndex="3" QueryPath="JavaScript\Cx\JavaScript High Risk\Dummy Vuln 1:4" QueryVersionCode="14383421">
	<Result NodeId="143834211111" FileName="test/any.ts" Status="Recurrent" Line="7" Column="46" FalsePositive="False" Severity="High" AssignToUser="" state="0" Remark="" DeepLink="https://cxtext.test/CxWebClient/ViewerMain.aspx?" SeverityIndex="3" StatusIndex="1" DetectionDate="3/7/2022 12:21:30 PM">
		<Path ResultId="11037" PathId="4" SimilarityId="-1754124988" SourceMethod="function" DestinationMethod="function">
		<PathNode>
			<FileName>test/any.ts</FileName>
			<Line>7</Line>
			<Column>46</Column>
			<NodeId>1</NodeId>
			<Name>slice</Name>
			<Type></Type>
			<Length>5</Length>
			<Snippet>
			<Line>
				<Number>7</Number>
				<Code>dummy code</Code>
			</Line>
			</Snippet>
		</PathNode>
		<PathNode>
			<FileName>test/any.ts</FileName>
			<Line>7</Line>
			<Column>12</Column>
			<NodeId>2</NodeId>
			<Name>location</Name>
			<Type></Type>
			<Length>8</Length>
			<Snippet>
			<Line>
				<Number>7</Number>
				<Code>dummy code 2</Code>
			</Line>
			</Snippet>
		</PathNode>
		</Path>
	</Result>
	<Result NodeId="143834211112" FileName="html/ts.ts" Status="Recurrent" Line="7" Column="46" FalsePositive="False" Severity="High" AssignToUser="" state="0" Remark="" DeepLink="https://cxtext.test/CxWebClient/ViewerMain.aspx?" SeverityIndex="3" StatusIndex="1" DetectionDate="3/7/2022 12:21:30 PM">
		<Path ResultId="4845356468" PathId="5" SimilarityId="-1465173916" SourceMethod="function" DestinationMethod="function">
		<PathNode>
			<FileName>html/other.ts</FileName>
			<Line>7</Line>
			<Column>46</Column>
			<NodeId>1</NodeId>
			<Name>slice</Name>
			<Type></Type>
			<Length>5</Length>
			<Snippet>
			<Line>
				<Number>7</Number>
				<Code>dummycode</Code>
			</Line>
			</Snippet>
		</PathNode>
		<PathNode>
			<FileName>html/other.ts</FileName>
			<Line>7</Line>
			<Column>12</Column>
			<NodeId>2</NodeId>
			<Name>location</Name>
			<Type></Type>
			<Length>8</Length>
			<Snippet>
			<Line>
				<Number>7</Number>
				<Code>dummycode2</Code>
			</Line>
			</Snippet>
		</PathNode>
		</Path>
	</Result>
	</Query>
	<Query id="1111" categories="Dummy Categories" cweId="79" name="Dummy Vuln 2" group="JavaScript_High_Risk" Severity="High" Language="JavaScript" LanguageHash="9095271965336651" LanguageChangeDate="2022-01-16T00:00:00.0000000" SeverityIndex="3" QueryPath="JavaScript\Cx\JavaScript High Risk\Dummy Vuln 1:4" QueryVersionCode="14383421">
	<Result NodeId="143834211111" FileName="test/any.ts" Status="Recurrent" Line="7" Column="46" FalsePositive="False" Severity="High" AssignToUser="" state="2" Remark="Test-user Test-project, [Monday, March 7, 2022 1:57:26 PM]: Dummy comment&#xD;&#xA;Test-user Test-project, [Monday, March 7, 2022 1:57:26 PM]: Changed status to Confirmed" DeepLink="https://cxtext.test/CxWebClient/ViewerMain.aspx?" SeverityIndex="3" StatusIndex="1" DetectionDate="3/7/2022 12:21:30 PM">
		<Path ResultId="11037" PathId="4" SimilarityId="-1754124988" SourceMethod="function" DestinationMethod="function">
		<PathNode>
			<FileName>test/any.ts</FileName>
			<Line>7</Line>
			<Column>46</Column>
			<NodeId>1</NodeId>
			<Name>slice</Name>
			<Type></Type>
			<Length>5</Length>
			<Snippet>
			<Line>
				<Number>7</Number>
				<Code>dummy code</Code>
			</Line>
			</Snippet>
		</PathNode>
		</Path>
	</Result>
	</Query>
	</CxXMLResults>
`

	t.Run("Valid config", func(t *testing.T) {
		opts := piperHttp.ClientOptions{}
		logger := log.Entry().WithField("package", "SAP/jenkins-library/pkg/checkmarx_test")
		myTestClient := senderMock{responseBody: `{"shortDescription":"This is a dummy short description."}`, httpStatusCode: 200}
		sys := SystemInstance{serverURL: "https://cx.server.com", client: &myTestClient, logger: logger}
		myTestClient.SetOptions(opts)

		sarif, err := Parse(&sys, []byte(testCxxml), 11037)
		assert.NoError(t, err, "error")
		assert.Equal(t, len(sarif.Runs[0].Results), 3)
		assert.Equal(t, len(sarif.Runs[0].Tool.Driver.Rules), 2)
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.ToolState, "Confirmed")
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.ToolAuditMessage, "Changed status to Confirmed \n Dummy comment")
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.ToolSeverityIndex, 3)
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.ToolSeverity, "High")
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.AuditRequirementIndex, format.AUDIT_REQUIREMENT_GROUP_1_INDEX)
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.AuditRequirement, format.AUDIT_REQUIREMENT_GROUP_1_DESC)
		//assert.Equal(t, "This is a dummy short description.", sarif.Runs[0].Tool.Driver.Rules[0].FullDescription.Text)

		// ensure the existence of not applicable field (specific Fortify)
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.InstanceSeverity, "")
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.Confidence, "")
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.FortifyCategory, "")
	})

	t.Run("Missing sys", func(t *testing.T) {

		sarif, err := Parse(nil, []byte(testCxxml), 11037)
		assert.NoError(t, err, "error")
		assert.Equal(t, len(sarif.Runs[0].Results), 3)
		assert.Equal(t, len(sarif.Runs[0].Tool.Driver.Rules), 2)
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.ToolState, "Confirmed")
		assert.Equal(t, sarif.Runs[0].Results[2].Properties.ToolAuditMessage, "Changed status to Confirmed \n Dummy comment")
		assert.Equal(t, "Dummy Categories", sarif.Runs[0].Tool.Driver.Rules[0].FullDescription.Text)
	})

	t.Run("Missing data", func(t *testing.T) {
		_, err := Parse(nil, []byte{}, 11037)
		assert.Error(t, err, "EOF")
	})

}
