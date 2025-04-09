//go:build unit

package fortify

import (
	"github.com/SAP/jenkins-library/pkg/format"
	"github.com/piper-validation/fortify-client-go/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {

	//use a test FVDL file here. The file should not be committed unless stripped of information.
	testFvdl := `
	<?xml version="1.0" encoding="UTF-8"?>
<FVDL xmlns="xmlns://www.fortifysoftware.com/schema/fvdl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" version="1.12" xsi:type="FVDL">
<CreatedTS date="2022-01-14" time="16:18:46"/>
<UUID>UUID</UUID>
<Build>
  <Project>PROJECTNAME</Project>
  <Label>https://api.github.com/repos///commits/</Label>
  <BuildID>BUILDID</BuildID>
  <NumberFiles>42</NumberFiles>
  <LOC type="Fortify">301</LOC>
  <LOC type="Line Count">747</LOC>
  <LOC type="Function Declarations">465</LOC>
  <LOC type="Function Definitions">198</LOC>
  <JavaClasspath>C:/java</JavaClasspath>
  <SourceBasePath>C:/fortify-reference-pipeline</SourceBasePath>
  <SourceFiles>
    <File size="17171" timestamp="1641914061433" type="xml" encoding="windows-1252">
      <Name>result/rules/Custom_Rules_for_Annotation_Management.xml</Name>
    </File>
  </SourceFiles>
  <ScanTime value="77"/>
</Build>
<Vulnerabilities>
<Vulnerability>
  <ClassInfo>
    <ClassID>B5C0FEFD-DUMMY</ClassID>
    <Type>SAST Configuration</Type>
    <Subtype>Custom Rules</Subtype>
    <AnalyzerName>configuration</AnalyzerName>
    <DefaultSeverity>5.0</DefaultSeverity>
  </ClassInfo>
  <InstanceInfo>
    <InstanceID>DUMMYDUMMYDUMMY</InstanceID>
    <InstanceSeverity>5.0</InstanceSeverity>
    <Confidence>5.0</Confidence>
  </InstanceInfo>
  <AnalysisInfo>
    <Unified>
      <Context/>
      <Trace>
        <Primary>
          <Entry>
            <Node isDefault="true">
              <SourceLocation path="result/rules/Custom_Rules_for_Annotation_Management.xml" line="2" colStart="0" colEnd="0" snippet="DUMMYDUMMY#result/rules/Custom_Rules_for_Annotation_Management.xml:2:2"/>
            </Node>
          </Entry>
          <Entry>
            <NodeRef id="4491"/>
          </Entry>
        </Primary>
      </Trace>
    </Unified>
  </AnalysisInfo>
</Vulnerability>
<Vulnerability>
  <ClassInfo>
    <ClassID>B5C0FEFD-DUMMY</ClassID>
    <Type>SAST Configuration</Type>
    <Subtype>Custom Rules</Subtype>
    <AnalyzerName>configuration</AnalyzerName>
    <DefaultSeverity>5.0</DefaultSeverity>
  </ClassInfo>
  <InstanceInfo>
    <InstanceID>DUMMYDUMMYDUMMY</InstanceID>
    <InstanceSeverity>5.0</InstanceSeverity>
    <Confidence>5.0</Confidence>
  </InstanceInfo>
  <AnalysisInfo>
    <Unified>
      <Context/>
      <Trace>
        <Primary>
          <Entry>
            <Node isDefault="true">
              <SourceLocation path="result/rules/Custom_Rules_for_Annotation_Management.xml" line="2" colStart="0" colEnd="0" snippet="DUMMYDUMMY#result/rules/Custom_Rules_for_Annotation_Management.xml:2:2"/>
              <Action>Dummy action</Action>
            </Node>
          </Entry>
        </Primary>
      </Trace>
    </Unified>
  </AnalysisInfo>
</Vulnerability>
</Vulnerabilities>
<ContextPool>
  <Context id="1">
    <Function name="toResponse" namespace="exceptionmappers" enclosingClass="ThrowableMapper"/>
    <FunctionDeclarationSourceLocation path="src/file.java" line="25" lineEnd="30" colStart="59" colEnd="0"/>
  </Context>
</ContextPool>
<UnifiedNodePool>
  <Node id="0">
    <SourceLocation path="src/file.java" line="28" lineEnd="28" colStart="76" colEnd="0" contextId="1" snippet="DUMMYDUMMY#src/file.java:28:28"/>
    <Action type="OutCall">getMessage(return)</Action>
    <Reason>
      <Rule ruleID="A6172DC7-DUMMY"/>
    </Reason>
    <Knowledge>
      <Fact primary="false" type="Call">Direct : java.lang.Throwable.getMessage</Fact>
    </Knowledge>
  </Node>
</UnifiedNodePool>
<Description contentType="preformatted" classID="B5C0FEFD-DUMMY">
  <Abstract>This scan contains project-specific custom rules. Please see the recommendation section on how to proceed.</Abstract>
  <Explanation>Custom rules can help improve scan quality. They can reduce both false positives and false negatives by tailoring the scan settings to match the threat model and other specifics of an application. At the same time, custom rules need to be part of the review when a scan is reviewed by an auditor. This issue is a reminder of this fact.</Explanation>
  <Recommendations>If you are an auditor reviewing this project, please review the custom rules and the associated documentation. If unsure, please consult the Security Testing team.
                
If you are a developer or other project member, please mark this finding as "Not an issue".</Recommendations>
</Description>
<Description contentType="preformatted" classID="C02261BC-DUMMY">
  <Abstract>&lt;Content&gt;&lt;Paragraph&gt;The function &lt;Replace key="EnclosingFunction.name"/&gt; in &lt;Replace key="PrimaryLocation.file"/&gt; reveals system data or debug information by calling &lt;Replace key="PrimaryCall.name"/&gt; on line &lt;Replace key="PrimaryLocation.line"/&gt;. The information revealed by &lt;Replace key="PrimaryCall.name"/&gt; could help an adversary form a plan of attack.&lt;AltParagraph&gt;Revealing system data or debugging information helps an adversary learn about the system and form a plan of attack.&lt;/AltParagraph&gt;&lt;/Paragraph&gt;&lt;/Content&gt;</Abstract>
  <Explanation>&lt;Content&gt;An external information leak occurs when system data or debug information leaves the program to a remote machine via a socket or network connection. External leaks can help an attacker by revealing specific data about operating systems, full pathnames, the existence of usernames, or locations of configuration files, and are more serious than internal information leaks, which are more difficult for an attacker to access.

&lt;Paragraph&gt;
In this case, &lt;Replace key="PrimaryCall.name" link="PrimaryLocation"/&gt; is called in &lt;Replace key="PrimaryLocation.file"/&gt; at line &lt;Replace key="PrimaryLocation.line"/&gt;.
&lt;/Paragraph&gt;

&lt;b&gt;Example 1:&lt;/b&gt; The following code leaks Exception information in the HTTP response:

&lt;pre&gt;
protected void doPost (HttpServletRequest req, HttpServletResponse res) throws IOException {
    ...
    PrintWriter out = res.getWriter();
    try {
        ...
    } catch (Exception e) {
      out.println(e.getMessage());
    }
}
&lt;/pre&gt;

This information can be exposed to a remote user. In some cases, the error message provides the attacker with the precise type of attack to which the system is vulnerable. For example, a database error message can reveal that the application is vulnerable to a SQL injection attack. Other error messages can reveal more oblique clues about the system. In &lt;code&gt;Example 1&lt;/code&gt;, the leaked information could imply information about the type of operating system, the applications installed on the system, and the amount of care that the administrators have put into configuring the program.

Information leaks are also a concern in a mobile computing environment. With mobile platforms, applications are downloaded from various sources and are run alongside each other on the same device. The likelihood of running a piece of malware next to a banking application is high, which is why application authors need to be careful about what information they include in messages addressed to other applications running on the device.

&lt;b&gt;Example 2:&lt;/b&gt; The following code broadcasts the stack trace of a caught exception to all the registered Android receivers.
&lt;pre&gt;
...
try {
  ...
} catch (Exception e) {
    String exception = Log.getStackTraceString(e);
    Intent i = new Intent();
    i.setAction("SEND_EXCEPTION");
    i.putExtra("exception", exception);
    view.getContext().sendBroadcast(i);
}
...
&lt;/pre&gt;

This is another scenario specific to the mobile environment. Most mobile devices now implement a Near-Field Communication (NFC) protocol for quickly sharing information between devices using radio communication. It works by bringing devices in close proximity or having the devices touch each other. Even though the communication range of NFC is limited to just a few centimeters, eavesdropping, data modification and various other types of attacks are possible, because NFC alone does not ensure secure communication.

&lt;b&gt;Example 3:&lt;/b&gt; The Android platform provides support for NFC. The following code creates a message that gets pushed to the other device within range.
&lt;pre&gt;
...
public static final String TAG = "NfcActivity";
private static final String DATA_SPLITTER = "__:DATA:__";
private static final String MIME_TYPE = "application/my.applications.mimetype";
...
TelephonyManager tm = (TelephonyManager)Context.getSystemService(Context.TELEPHONY_SERVICE);
String VERSION = tm.getDeviceSoftwareVersion();
...
NfcAdapter nfcAdapter = NfcAdapter.getDefaultAdapter(this);
if (nfcAdapter == null)
  return;

String text = TAG + DATA_SPLITTER + VERSION;
NdefRecord record = new NdefRecord(NdefRecord.TNF_MIME_MEDIA,
            MIME_TYPE.getBytes(), new byte[0], text.getBytes());
NdefRecord[] records = { record };
NdefMessage msg = new NdefMessage(records);
nfcAdapter.setNdefPushMessage(msg, this);
...
&lt;/pre&gt;

An NFC Data Exchange Format (NDEF) message contains typed data, a URI, or a custom application payload. If the message contains information about the application, such as its name, MIME type, or device software version, this information could be leaked to an eavesdropper.&lt;/Content&gt;</Explanation>
  <Recommendations>&lt;Content&gt;Write error messages with security in mind. In production environments, turn off detailed error information in favor of brief messages. Restrict the generation and storage of detailed output that can help administrators and programmers diagnose problems. Debug traces can sometimes appear in non-obvious places (embedded in comments in the HTML for an error page, for example).

Even brief error messages that do not reveal stack traces or database dumps can potentially aid an attacker. For example, an "Access Denied" message can reveal that a file or user exists on the system. Because of this, never send information to a resource directly outside the program.

&lt;b&gt;Example 4:&lt;/b&gt; The following code broadcasts the stack trace of a caught exception within your application only, so that it cannot be leaked to other apps on the system. Additionally, this technique is more efficient than globally broadcasting through the system.

&lt;pre&gt;
...
try {
  ...
} catch (Exception e) {
    String exception = Log.getStackTraceString(e);
    Intent i = new Intent();
    i.setAction("SEND_EXCEPTION");
    i.putExtra("exception", exception);
    LocalBroadcastManager.getInstance(view.getContext()).sendBroadcast(i);
}
...
&lt;/pre&gt;

If you are concerned about leaking system data via NFC on an Android device, you could do one of the following three things. Do not include system data in the messages pushed to other devices in range, encrypt the payload of the message, or establish a secure communication channel at a higher layer.&lt;/Content&gt;</Recommendations>
  <Tips>
    <Tip>Do not rely on wrapper scripts, corporate IT policy, or quick-thinking system administrators to prevent system information leaks. Write software that is secure on its own.</Tip>
    <Tip>This category of vulnerability does not apply to all types of programs. For example, if your application executes on a client machine where system information is already available to an attacker, or if you print system information only to a trusted log file, you can use Audit Guide to filter out this category from your scan results.</Tip>
  </Tips>
  <References>
    <Reference>
      <Title>Security in Near Field Communication (NFC): Strengths and Weaknesses</Title>
      <Author>Ernst Haselsteiner and Klemens Breitfuss</Author>
      <Source>http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.475.3812&amp;rep=rep1&amp;type=pdf</Source>
    </Reference>
  </References>
</Description>
<Snippets>
  <Snippet id="DUMMYDUMMY#result/rules/Custom_Rules_for_Annotation_Management.xml:2:2">
    <File>result/rules/Custom_Rules_for_Annotation_Management.xml</File>
    <StartLine>1</StartLine>
    <EndLine>5</EndLine>
    <Text><![CDATA[<?xml version="1.0" encoding="UTF-8"?>
<RulePack xmlns="xmlns://www.fortifysoftware.com/schema/rules">
    <RulePackID>57658246-DUMMY</RulePackID>
    <SKU>SKU-094b9c82-DUMMY</SKU>
    <Name><![CDATA[Custom Rules for Annotation Management]]]]><![CDATA[></Name>
]]></Text>
  </Snippet>
</Snippets>
<ProgramData>
  <Sources>
    <SourceInstance ruleID="07CF967B-DUMMY">
      <FunctionCall>
        <SourceLocation path="src/file.java" line="28" lineEnd="28" colStart="47" colEnd="0"/>
        <Function name="getName" namespace="java.lang" enclosingClass="Class"/>
      </FunctionCall>
      <TaintFlags>
        <TaintFlag name="CLASS_NAME"/>
      </TaintFlags>
    </SourceInstance>
  </Sources>
  <Sinks>
    <SinkInstance ruleID="9667C493-DUMMY">
      <FunctionCall>
        <SourceLocation path="src/file.java" line="99" lineEnd="99" colStart="16" colEnd="0"/>
        <Function name="trace" namespace="org.slf4j" enclosingClass="Logger"/>
      </FunctionCall>
    </SinkInstance>
  </Sinks>
  <CalledWithNoDef>
    <Function name="setLoadTimeWeaver" namespace="org.springframework.orm.jpa" enclosingClass="LocalContainerEntityManagerFactoryBean"/>
  </CalledWithNoDef>
</ProgramData>
<EngineData>
  <EngineVersion>20.2.0.0139</EngineVersion>
  <RulePacks>
    <RulePack>
      <RulePackID>14EE50EB-DUMMY</RulePackID>
      <SKU>RUL13078</SKU>
      <Name>Fortify Secure Coding Rules, Core, Annotations</Name>
      <Version>2020.4.0.0007</Version>
      <MAC>DUMMY==</MAC>
    </RulePack>
  </RulePacks>
  <Properties type="System">
    <Property>
      <name>os.name</name>
      <value>Windows 10</value>
    </Property>
  </Properties>
  <CommandLine>
    <Argument>-verbose</Argument>
  </CommandLine>
  <Errors>
    <Error code="12003"><![CDATA[Assuming Java source level to be 1.8 as it was not specified. Note that the default value may change in future versions.]]></Error>
  </Errors>
  <MachineInfo>
    <Hostname>W-E</Hostname>
    <Username>XXXXXXX</Username>
    <Platform>Windows 10</Platform>
  </MachineInfo>
  <FilterResult/>
  <RuleInfo>
    <Rule id="B5C0FEFD-DUMMY">
      <MetaInfo>
        <Group name="Accuracy">5</Group>
        <Group name="Impact">5</Group>
        <Group name="RemediationEffort">1</Group>
        <Group name="Probability">5</Group>
        <Group name="altcategoryCWE">CWE ID 111</Group>
      </MetaInfo>
    </Rule>
  </RuleInfo>
  <LicenseInfo>
    <Metadata>
      <name>owner</name>
      <value>S - FAN23043</value>
    </Metadata>
    <Capability>
      <Name>VSPlugins</Name>
      <Expiration>2032-12-31</Expiration>
    </Capability>
  </LicenseInfo>
</EngineData>
</FVDL>	`

	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Split(req.URL.Path, "/")[1] == "projectVersions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{
          "data": [
            {
              "projectVersionId": 11037,
              "issueInstanceId": "DUMMYDUMMYDUMMY",
              "issueName": "Dummy issue",
              "primaryTag": "Exploitable",
              "audited": true,
              "issueStatus": "Reviewed",
              "folderGuid": "aaaaaaaa-1111-aaaa-1111-1111aaaaaaaa",
              "hasComments": true,
              "friority": "High",
              "_href": "https://fortify-stage.tools.sap/ssc/api/v1/projectVersions/11037"
            }
          ],
          "count": 1,
          "responseCode": 200}`))
			return
		}
		if strings.Split(req.URL.Path, "/")[1] == "issues" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{
          "data": [
            {
              "issueId": 47009919,
              "comment": "Dummy comment."
            }
          ],
          "count": 1,
          "responseCode": 200}`))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	filterSet := new(models.FilterSet)
	filterSet.Folders = append(filterSet.Folders, &models.FolderDto{GUID: "aaaaaaaa-1111-aaaa-1111-1111aaaaaaaa", Name: "Audit All"})

	t.Run("Valid config", func(t *testing.T) {
		projectVersion := models.ProjectVersion{ID: 11037}
		sarif, sarifSimplified, err := Parse(sys, &projectVersion, []byte(testFvdl), filterSet)
		assert.NoError(t, err, "error")
		assert.Equal(t, len(sarif.Runs[0].Results), 2)
		assert.Equal(t, len(sarif.Runs[0].Results[0].Locations), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].CodeFlows), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].RelatedLocations), 1)
		assert.Equal(t, len(sarif.Runs[0].Tool.Driver.Rules), 1)
		assert.Equal(t, sarif.Runs[0].Results[0].Properties.ToolState, "Exploitable")
		assert.Equal(t, sarif.Runs[0].Results[0].Properties.ToolAuditMessage, "Dummy comment.")
		assert.Equal(t, sarif.Runs[0].OriginalUriBaseIds, &format.OriginalUriBaseIds{SrcRoot: format.SrcRoot{Uri: "file:///C:/fortify-reference-pipeline/"}})

		//test simplified structure
		assert.Equal(t, len(sarifSimplified.Runs[0].Results), 2)                     // same results
		assert.Equal(t, len(sarifSimplified.Runs[0].Tool.Driver.Rules), 1)           // same rules
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].Locations), 0)        // without location
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].CodeFlows), 0)        // without code flows
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].RelatedLocations), 0) // without related location
		assert.Equal(t, sarifSimplified.Runs[0].Results[0].Properties.ToolState, "Exploitable")
		assert.Equal(t, sarifSimplified.Runs[0].Results[0].Properties.ToolAuditMessage, "Dummy comment.")
		assert.Equal(t, sarifSimplified.Runs[0].OriginalUriBaseIds, (*format.OriginalUriBaseIds)(nil)) // without OriginalUriBaseIds
	})

	t.Run("Missing data", func(t *testing.T) {
		projectVersion := models.ProjectVersion{ID: 11037}
		_, _, err := Parse(sys, &projectVersion, []byte{}, filterSet)
		assert.Error(t, err, "EOF")
	})

	t.Run("No system instance", func(t *testing.T) {
		projectVersion := models.ProjectVersion{ID: 11037}
		sarif, sarifSimplified, err := Parse(nil, &projectVersion, []byte(testFvdl), filterSet)
		assert.NoError(t, err, "error")
		assert.Equal(t, len(sarif.Runs[0].Results), 2)
		assert.Equal(t, len(sarif.Runs[0].Tool.Driver.Rules), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].Locations), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].CodeFlows), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].RelatedLocations), 1)
		assert.Equal(t, sarif.Runs[0].Results[0].Properties.ToolState, "Unknown")
		assert.Equal(t, sarif.Runs[0].Results[0].Properties.ToolAuditMessage, "Cannot fetch audit state: no sys instance")
		assert.Equal(t, sarif.Runs[0].OriginalUriBaseIds, &format.OriginalUriBaseIds{SrcRoot: format.SrcRoot{Uri: "file:///C:/fortify-reference-pipeline/"}})

		//test simplified structure
		assert.Equal(t, len(sarifSimplified.Runs[0].Results), 2)                     // same results
		assert.Equal(t, len(sarifSimplified.Runs[0].Tool.Driver.Rules), 1)           // same rules
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].Locations), 0)        // without location
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].CodeFlows), 0)        // without code flows
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].RelatedLocations), 0) // without related location
		assert.Equal(t, sarifSimplified.Runs[0].Results[0].Properties.ToolState, "Unknown")
		assert.Equal(t, sarifSimplified.Runs[0].Results[0].Properties.ToolAuditMessage, "Cannot fetch audit state: no sys instance")
		assert.Equal(t, sarifSimplified.Runs[0].OriginalUriBaseIds, (*format.OriginalUriBaseIds)(nil)) // without OriginalUriBaseIds
	})
}

func TestParse_EmptySourceBasePath(t *testing.T) {

	//use a test FVDL file here. The file should not be committed unless stripped of information.
	testFvdl := `
	<?xml version="1.0" encoding="UTF-8"?>
<FVDL xmlns="xmlns://www.fortifysoftware.com/schema/fvdl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" version="1.12" xsi:type="FVDL">
<CreatedTS date="2022-01-14" time="16:18:46"/>
<UUID>UUID</UUID>
<Build>
  <Project>PROJECTNAME</Project>
  <Label>https://api.github.com/repos///commits/</Label>
  <BuildID>BUILDID</BuildID>
  <NumberFiles>42</NumberFiles>
  <LOC type="Fortify">301</LOC>
  <LOC type="Line Count">747</LOC>
  <LOC type="Function Declarations">465</LOC>
  <LOC type="Function Definitions">198</LOC>
  <JavaClasspath>C:/java</JavaClasspath>
  <SourceBasePath/>
  <SourceFiles>
    <File size="17171" timestamp="1641914061433" type="xml" encoding="windows-1252">
      <Name>result/rules/Custom_Rules_for_Annotation_Management.xml</Name>
    </File>
  </SourceFiles>
  <ScanTime value="77"/>
</Build>
<Vulnerabilities>
<Vulnerability>
  <ClassInfo>
    <ClassID>B5C0FEFD-DUMMY</ClassID>
    <Type>SAST Configuration</Type>
    <Subtype>Custom Rules</Subtype>
    <AnalyzerName>configuration</AnalyzerName>
    <DefaultSeverity>5.0</DefaultSeverity>
  </ClassInfo>
  <InstanceInfo>
    <InstanceID>DUMMYDUMMYDUMMY</InstanceID>
    <InstanceSeverity>5.0</InstanceSeverity>
    <Confidence>5.0</Confidence>
  </InstanceInfo>
  <AnalysisInfo>
    <Unified>
      <Context/>
      <Trace>
        <Primary>
          <Entry>
            <Node isDefault="true">
              <SourceLocation path="result/rules/Custom_Rules_for_Annotation_Management.xml" line="2" colStart="0" colEnd="0" snippet="DUMMYDUMMY#result/rules/Custom_Rules_for_Annotation_Management.xml:2:2"/>
            </Node>
          </Entry>
          <Entry>
            <NodeRef id="4491"/>
          </Entry>
        </Primary>
      </Trace>
    </Unified>
  </AnalysisInfo>
</Vulnerability>
<Vulnerability>
  <ClassInfo>
    <ClassID>B5C0FEFD-DUMMY</ClassID>
    <Type>SAST Configuration</Type>
    <Subtype>Custom Rules</Subtype>
    <AnalyzerName>configuration</AnalyzerName>
    <DefaultSeverity>5.0</DefaultSeverity>
  </ClassInfo>
  <InstanceInfo>
    <InstanceID>DUMMYDUMMYDUMMY</InstanceID>
    <InstanceSeverity>5.0</InstanceSeverity>
    <Confidence>5.0</Confidence>
  </InstanceInfo>
  <AnalysisInfo>
    <Unified>
      <Context/>
      <Trace>
        <Primary>
          <Entry>
            <Node isDefault="true">
              <SourceLocation path="result/rules/Custom_Rules_for_Annotation_Management.xml" line="2" colStart="0" colEnd="0" snippet="DUMMYDUMMY#result/rules/Custom_Rules_for_Annotation_Management.xml:2:2"/>
              <Action>Dummy action</Action>
            </Node>
          </Entry>
        </Primary>
      </Trace>
    </Unified>
  </AnalysisInfo>
</Vulnerability>
</Vulnerabilities>
<ContextPool>
  <Context id="1">
    <Function name="toResponse" namespace="exceptionmappers" enclosingClass="ThrowableMapper"/>
    <FunctionDeclarationSourceLocation path="src/file.java" line="25" lineEnd="30" colStart="59" colEnd="0"/>
  </Context>
</ContextPool>
<UnifiedNodePool>
  <Node id="0">
    <SourceLocation path="src/file.java" line="28" lineEnd="28" colStart="76" colEnd="0" contextId="1" snippet="DUMMYDUMMY#src/file.java:28:28"/>
    <Action type="OutCall">getMessage(return)</Action>
    <Reason>
      <Rule ruleID="A6172DC7-DUMMY"/>
    </Reason>
    <Knowledge>
      <Fact primary="false" type="Call">Direct : java.lang.Throwable.getMessage</Fact>
    </Knowledge>
  </Node>
</UnifiedNodePool>
<Description contentType="preformatted" classID="B5C0FEFD-DUMMY">
  <Abstract>This scan contains project-specific custom rules. Please see the recommendation section on how to proceed.</Abstract>
  <Explanation>Custom rules can help improve scan quality. They can reduce both false positives and false negatives by tailoring the scan settings to match the threat model and other specifics of an application. At the same time, custom rules need to be part of the review when a scan is reviewed by an auditor. This issue is a reminder of this fact.</Explanation>
  <Recommendations>If you are an auditor reviewing this project, please review the custom rules and the associated documentation. If unsure, please consult the Security Testing team.
                
If you are a developer or other project member, please mark this finding as "Not an issue".</Recommendations>
</Description>
<Description contentType="preformatted" classID="C02261BC-DUMMY">
  <Abstract>&lt;Content&gt;&lt;Paragraph&gt;The function &lt;Replace key="EnclosingFunction.name"/&gt; in &lt;Replace key="PrimaryLocation.file"/&gt; reveals system data or debug information by calling &lt;Replace key="PrimaryCall.name"/&gt; on line &lt;Replace key="PrimaryLocation.line"/&gt;. The information revealed by &lt;Replace key="PrimaryCall.name"/&gt; could help an adversary form a plan of attack.&lt;AltParagraph&gt;Revealing system data or debugging information helps an adversary learn about the system and form a plan of attack.&lt;/AltParagraph&gt;&lt;/Paragraph&gt;&lt;/Content&gt;</Abstract>
  <Explanation>&lt;Content&gt;An external information leak occurs when system data or debug information leaves the program to a remote machine via a socket or network connection. External leaks can help an attacker by revealing specific data about operating systems, full pathnames, the existence of usernames, or locations of configuration files, and are more serious than internal information leaks, which are more difficult for an attacker to access.

&lt;Paragraph&gt;
In this case, &lt;Replace key="PrimaryCall.name" link="PrimaryLocation"/&gt; is called in &lt;Replace key="PrimaryLocation.file"/&gt; at line &lt;Replace key="PrimaryLocation.line"/&gt;.
&lt;/Paragraph&gt;

&lt;b&gt;Example 1:&lt;/b&gt; The following code leaks Exception information in the HTTP response:

&lt;pre&gt;
protected void doPost (HttpServletRequest req, HttpServletResponse res) throws IOException {
    ...
    PrintWriter out = res.getWriter();
    try {
        ...
    } catch (Exception e) {
      out.println(e.getMessage());
    }
}
&lt;/pre&gt;

This information can be exposed to a remote user. In some cases, the error message provides the attacker with the precise type of attack to which the system is vulnerable. For example, a database error message can reveal that the application is vulnerable to a SQL injection attack. Other error messages can reveal more oblique clues about the system. In &lt;code&gt;Example 1&lt;/code&gt;, the leaked information could imply information about the type of operating system, the applications installed on the system, and the amount of care that the administrators have put into configuring the program.

Information leaks are also a concern in a mobile computing environment. With mobile platforms, applications are downloaded from various sources and are run alongside each other on the same device. The likelihood of running a piece of malware next to a banking application is high, which is why application authors need to be careful about what information they include in messages addressed to other applications running on the device.

&lt;b&gt;Example 2:&lt;/b&gt; The following code broadcasts the stack trace of a caught exception to all the registered Android receivers.
&lt;pre&gt;
...
try {
  ...
} catch (Exception e) {
    String exception = Log.getStackTraceString(e);
    Intent i = new Intent();
    i.setAction("SEND_EXCEPTION");
    i.putExtra("exception", exception);
    view.getContext().sendBroadcast(i);
}
...
&lt;/pre&gt;

This is another scenario specific to the mobile environment. Most mobile devices now implement a Near-Field Communication (NFC) protocol for quickly sharing information between devices using radio communication. It works by bringing devices in close proximity or having the devices touch each other. Even though the communication range of NFC is limited to just a few centimeters, eavesdropping, data modification and various other types of attacks are possible, because NFC alone does not ensure secure communication.

&lt;b&gt;Example 3:&lt;/b&gt; The Android platform provides support for NFC. The following code creates a message that gets pushed to the other device within range.
&lt;pre&gt;
...
public static final String TAG = "NfcActivity";
private static final String DATA_SPLITTER = "__:DATA:__";
private static final String MIME_TYPE = "application/my.applications.mimetype";
...
TelephonyManager tm = (TelephonyManager)Context.getSystemService(Context.TELEPHONY_SERVICE);
String VERSION = tm.getDeviceSoftwareVersion();
...
NfcAdapter nfcAdapter = NfcAdapter.getDefaultAdapter(this);
if (nfcAdapter == null)
  return;

String text = TAG + DATA_SPLITTER + VERSION;
NdefRecord record = new NdefRecord(NdefRecord.TNF_MIME_MEDIA,
            MIME_TYPE.getBytes(), new byte[0], text.getBytes());
NdefRecord[] records = { record };
NdefMessage msg = new NdefMessage(records);
nfcAdapter.setNdefPushMessage(msg, this);
...
&lt;/pre&gt;

An NFC Data Exchange Format (NDEF) message contains typed data, a URI, or a custom application payload. If the message contains information about the application, such as its name, MIME type, or device software version, this information could be leaked to an eavesdropper.&lt;/Content&gt;</Explanation>
  <Recommendations>&lt;Content&gt;Write error messages with security in mind. In production environments, turn off detailed error information in favor of brief messages. Restrict the generation and storage of detailed output that can help administrators and programmers diagnose problems. Debug traces can sometimes appear in non-obvious places (embedded in comments in the HTML for an error page, for example).

Even brief error messages that do not reveal stack traces or database dumps can potentially aid an attacker. For example, an "Access Denied" message can reveal that a file or user exists on the system. Because of this, never send information to a resource directly outside the program.

&lt;b&gt;Example 4:&lt;/b&gt; The following code broadcasts the stack trace of a caught exception within your application only, so that it cannot be leaked to other apps on the system. Additionally, this technique is more efficient than globally broadcasting through the system.

&lt;pre&gt;
...
try {
  ...
} catch (Exception e) {
    String exception = Log.getStackTraceString(e);
    Intent i = new Intent();
    i.setAction("SEND_EXCEPTION");
    i.putExtra("exception", exception);
    LocalBroadcastManager.getInstance(view.getContext()).sendBroadcast(i);
}
...
&lt;/pre&gt;

If you are concerned about leaking system data via NFC on an Android device, you could do one of the following three things. Do not include system data in the messages pushed to other devices in range, encrypt the payload of the message, or establish a secure communication channel at a higher layer.&lt;/Content&gt;</Recommendations>
  <Tips>
    <Tip>Do not rely on wrapper scripts, corporate IT policy, or quick-thinking system administrators to prevent system information leaks. Write software that is secure on its own.</Tip>
    <Tip>This category of vulnerability does not apply to all types of programs. For example, if your application executes on a client machine where system information is already available to an attacker, or if you print system information only to a trusted log file, you can use Audit Guide to filter out this category from your scan results.</Tip>
  </Tips>
  <References>
    <Reference>
      <Title>Security in Near Field Communication (NFC): Strengths and Weaknesses</Title>
      <Author>Ernst Haselsteiner and Klemens Breitfuss</Author>
      <Source>http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.475.3812&amp;rep=rep1&amp;type=pdf</Source>
    </Reference>
  </References>
</Description>
<Snippets>
  <Snippet id="DUMMYDUMMY#result/rules/Custom_Rules_for_Annotation_Management.xml:2:2">
    <File>result/rules/Custom_Rules_for_Annotation_Management.xml</File>
    <StartLine>1</StartLine>
    <EndLine>5</EndLine>
    <Text><![CDATA[<?xml version="1.0" encoding="UTF-8"?>
<RulePack xmlns="xmlns://www.fortifysoftware.com/schema/rules">
    <RulePackID>57658246-DUMMY</RulePackID>
    <SKU>SKU-094b9c82-DUMMY</SKU>
    <Name><![CDATA[Custom Rules for Annotation Management]]]]><![CDATA[></Name>
]]></Text>
  </Snippet>
</Snippets>
<ProgramData>
  <Sources>
    <SourceInstance ruleID="07CF967B-DUMMY">
      <FunctionCall>
        <SourceLocation path="src/file.java" line="28" lineEnd="28" colStart="47" colEnd="0"/>
        <Function name="getName" namespace="java.lang" enclosingClass="Class"/>
      </FunctionCall>
      <TaintFlags>
        <TaintFlag name="CLASS_NAME"/>
      </TaintFlags>
    </SourceInstance>
  </Sources>
  <Sinks>
    <SinkInstance ruleID="9667C493-DUMMY">
      <FunctionCall>
        <SourceLocation path="src/file.java" line="99" lineEnd="99" colStart="16" colEnd="0"/>
        <Function name="trace" namespace="org.slf4j" enclosingClass="Logger"/>
      </FunctionCall>
    </SinkInstance>
  </Sinks>
  <CalledWithNoDef>
    <Function name="setLoadTimeWeaver" namespace="org.springframework.orm.jpa" enclosingClass="LocalContainerEntityManagerFactoryBean"/>
  </CalledWithNoDef>
</ProgramData>
<EngineData>
  <EngineVersion>20.2.0.0139</EngineVersion>
  <RulePacks>
    <RulePack>
      <RulePackID>14EE50EB-DUMMY</RulePackID>
      <SKU>RUL13078</SKU>
      <Name>Fortify Secure Coding Rules, Core, Annotations</Name>
      <Version>2020.4.0.0007</Version>
      <MAC>DUMMY==</MAC>
    </RulePack>
  </RulePacks>
  <Properties type="System">
    <Property>
      <name>os.name</name>
      <value>Windows 10</value>
    </Property>
  </Properties>
  <CommandLine>
    <Argument>-verbose</Argument>
  </CommandLine>
  <Errors>
    <Error code="12003"><![CDATA[Assuming Java source level to be 1.8 as it was not specified. Note that the default value may change in future versions.]]></Error>
  </Errors>
  <MachineInfo>
    <Hostname>W-E</Hostname>
    <Username>XXXXXXX</Username>
    <Platform>Windows 10</Platform>
  </MachineInfo>
  <FilterResult/>
  <RuleInfo>
    <Rule id="B5C0FEFD-DUMMY">
      <MetaInfo>
        <Group name="Accuracy">5</Group>
        <Group name="Impact">5</Group>
        <Group name="RemediationEffort">1</Group>
        <Group name="Probability">5</Group>
        <Group name="altcategoryCWE">CWE ID 111</Group>
      </MetaInfo>
    </Rule>
  </RuleInfo>
  <LicenseInfo>
    <Metadata>
      <name>owner</name>
      <value>S - FAN23043</value>
    </Metadata>
    <Capability>
      <Name>VSPlugins</Name>
      <Expiration>2032-12-31</Expiration>
    </Capability>
  </LicenseInfo>
</EngineData>
</FVDL>	`

	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Split(req.URL.Path, "/")[1] == "projectVersions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{
          "data": [
            {
              "projectVersionId": 11037,
              "issueInstanceId": "DUMMYDUMMYDUMMY",
              "issueName": "Dummy issue",
              "primaryTag": "Exploitable",
              "audited": true,
              "issueStatus": "Reviewed",
              "folderGuid": "aaaaaaaa-1111-aaaa-1111-1111aaaaaaaa",
              "hasComments": true,
              "friority": "High",
              "_href": "https://fortify-stage.tools.sap/ssc/api/v1/projectVersions/11037"
            }
          ],
          "count": 1,
          "responseCode": 200}`))
			return
		}
		if strings.Split(req.URL.Path, "/")[1] == "issues" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{
          "data": [
            {
              "issueId": 47009919,
              "comment": "Dummy comment."
            }
          ],
          "count": 1,
          "responseCode": 200}`))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	filterSet := new(models.FilterSet)
	filterSet.Folders = append(filterSet.Folders, &models.FolderDto{GUID: "aaaaaaaa-1111-aaaa-1111-1111aaaaaaaa", Name: "Audit All"})

	t.Run("Valid config", func(t *testing.T) {
		projectVersion := models.ProjectVersion{ID: 11037}
		sarif, sarifSimplified, err := Parse(sys, &projectVersion, []byte(testFvdl), filterSet)
		assert.NoError(t, err, "error")
		assert.Equal(t, len(sarif.Runs[0].Results), 2)
		assert.Equal(t, len(sarif.Runs[0].Tool.Driver.Rules), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].Locations), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].CodeFlows), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].RelatedLocations), 1)
		assert.Equal(t, sarif.Runs[0].Results[0].Properties.ToolState, "Exploitable")
		assert.Equal(t, sarif.Runs[0].Results[0].Properties.ToolAuditMessage, "Dummy comment.")
		assert.Equal(t, sarif.Runs[0].OriginalUriBaseIds, (*format.OriginalUriBaseIds)(nil))

		// test simplified sarif structure
		assert.Equal(t, len(sarifSimplified.Runs[0].Results), 2)
		assert.Equal(t, len(sarifSimplified.Runs[0].Tool.Driver.Rules), 1)
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].Locations), 0)
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].CodeFlows), 0)
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].RelatedLocations), 0)
		assert.Equal(t, sarifSimplified.Runs[0].Results[0].Properties.ToolState, "Exploitable")
		assert.Equal(t, sarifSimplified.Runs[0].Results[0].Properties.ToolAuditMessage, "Dummy comment.")
		assert.Equal(t, sarifSimplified.Runs[0].OriginalUriBaseIds, (*format.OriginalUriBaseIds)(nil))
	})

	t.Run("Missing data", func(t *testing.T) {
		projectVersion := models.ProjectVersion{ID: 11037}
		_, _, err := Parse(sys, &projectVersion, []byte{}, filterSet)
		assert.Error(t, err, "EOF")
	})

	t.Run("No system instance", func(t *testing.T) {
		projectVersion := models.ProjectVersion{ID: 11037}
		sarif, sarifSimplified, err := Parse(nil, &projectVersion, []byte(testFvdl), filterSet)
		assert.NoError(t, err, "error")
		assert.Equal(t, len(sarif.Runs[0].Results), 2)
		assert.Equal(t, len(sarif.Runs[0].Tool.Driver.Rules), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].Locations), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].CodeFlows), 1)
		assert.Equal(t, len(sarif.Runs[0].Results[0].RelatedLocations), 1)
		assert.Equal(t, sarif.Runs[0].Results[0].Properties.ToolState, "Unknown")
		assert.Equal(t, sarif.Runs[0].Results[0].Properties.ToolAuditMessage, "Cannot fetch audit state: no sys instance")
		assert.Equal(t, sarif.Runs[0].OriginalUriBaseIds, (*format.OriginalUriBaseIds)(nil))

		assert.Equal(t, len(sarifSimplified.Runs[0].Results), 2)
		assert.Equal(t, len(sarifSimplified.Runs[0].Tool.Driver.Rules), 1)
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].Locations), 0)
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].CodeFlows), 0)
		assert.Equal(t, len(sarifSimplified.Runs[0].Results[0].RelatedLocations), 0)
		assert.Equal(t, sarifSimplified.Runs[0].Results[0].Properties.ToolState, "Unknown")
		assert.Equal(t, sarifSimplified.Runs[0].Results[0].Properties.ToolAuditMessage, "Cannot fetch audit state: no sys instance")
		assert.Equal(t, sarifSimplified.Runs[0].OriginalUriBaseIds, (*format.OriginalUriBaseIds)(nil))
	})
}

func TestIntegrateAuditData(t *testing.T) {
	sys, server := spinUpServer(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Split(req.URL.Path, "/")[1] == "projectVersions" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{
          "data": [
            {
              "projectVersionId": 11037,
              "issueInstanceId": "DUMMYDUMMYDUMMY",
              "issueName": "Dummy issue",
              "primaryTag": "Exploitable",
              "audited": true,
              "issueStatus": "Reviewed",
              "folderGuid": "aaaaaaaa-1111-aaaa-1111-1111aaaaaaaa",
              "hasComments": true,
              "friority": "High",
              "_href": "https://fortify-stage.tools.sap/ssc/api/v1/projectVersions/11037"
            }
          ],
          "count": 1,
          "responseCode": 200}`))
			return
		}
		if strings.Split(req.URL.Path, "/")[1] == "issues" {
			header := rw.Header()
			header.Add("Content-type", "application/json")
			rw.Write([]byte(
				`{
          "data": [
            {
              "issueId": 47009919,
              "comment": "Dummy comment."
            }
          ],
          "count": 1,
          "responseCode": 200}`))
			return
		}
	})
	// Close the server when test finishes
	defer server.Close()

	filterSet := new(models.FilterSet)
	filterSet.Folders = append(filterSet.Folders, &models.FolderDto{GUID: "aaaaaaaa-1111-aaaa-1111-1111aaaaaaaa", Name: "Audit All"})

	t.Run("Successful lookup", func(t *testing.T) {
		ruleProp := *new(format.SarifProperties)
		projectVersion := models.ProjectVersion{ID: 11037}
		auditData, _ := sys.GetAllIssueDetails(projectVersion.ID)
		err := integrateAuditData(&ruleProp, "DUMMYDUMMYDUMMY", sys, &projectVersion, auditData, filterSet, false, 5)
		assert.NoError(t, err, "error")
		assert.Equal(t, ruleProp.Audited, true)
		assert.Equal(t, ruleProp.ToolState, "Exploitable")
		assert.Equal(t, ruleProp.ToolStateIndex, 5)
		assert.Equal(t, ruleProp.ToolSeverity, "High")
		assert.Equal(t, ruleProp.ToolSeverityIndex, 3)
		assert.Equal(t, ruleProp.ToolAuditMessage, "Dummy comment.")
		assert.Equal(t, ruleProp.FortifyCategory, "Audit All")
		assert.Equal(t, ruleProp.AuditRequirementIndex, format.AUDIT_REQUIREMENT_GROUP_1_INDEX)
		assert.Equal(t, ruleProp.AuditRequirement, format.AUDIT_REQUIREMENT_GROUP_1_DESC)
		assert.Equal(t, ruleProp.CheckmarxSimilarityID, "") // ensure the existence of not applicable field (specific Checkmarx)
	})

	t.Run("Missing project version", func(t *testing.T) {
		ruleProp := *new(format.SarifProperties)
		auditData, _ := sys.GetAllIssueDetails(11037)
		err := integrateAuditData(&ruleProp, "DUMMYDUMMYDUMMY", sys, nil, auditData, filterSet, false, 5)
		assert.Error(t, err, "project or projectVersion is undefined: lookup aborted for 11037")
	})

	t.Run("Missing sys", func(t *testing.T) {
		ruleProp := *new(format.SarifProperties)
		projectVersion := models.ProjectVersion{ID: 11037}
		auditData, _ := sys.GetAllIssueDetails(projectVersion.ID)
		err := integrateAuditData(&ruleProp, "DUMMYDUMMYDUMMY", nil, &projectVersion, auditData, filterSet, false, 5)
		assert.Error(t, err, "no system instance, lookup impossible for DUMMYDUMMYDUMMY")
	})

	t.Run("Missing filterSet", func(t *testing.T) {
		ruleProp := *new(format.SarifProperties)
		projectVersion := models.ProjectVersion{ID: 11037}
		auditData, _ := sys.GetAllIssueDetails(projectVersion.ID)
		err := integrateAuditData(&ruleProp, "DUMMYDUMMYDUMMY", sys, &projectVersion, auditData, nil, false, 5)
		assert.Error(t, err, "no filter set defined, category will be missing from 11037")
	})

	t.Run("Missing Audit Data", func(t *testing.T) {
		ruleProp := *new(format.SarifProperties)
		projectVersion := models.ProjectVersion{ID: 11037}
		err := integrateAuditData(&ruleProp, "DUMMYDUMMYDUMMY", sys, &projectVersion, nil, filterSet, false, 5)
		assert.Error(t, err, "not exactly 1 issue found for instance ID 11037, found 0")
	})

	t.Run("Successful lookup in oneRequestPerInstance mode", func(t *testing.T) {
		ruleProp := *new(format.SarifProperties)
		projectVersion := models.ProjectVersion{ID: 11037}
		err := integrateAuditData(&ruleProp, "DUMMYDUMMYDUMMY", sys, &projectVersion, nil, filterSet, true, 5)
		assert.NoError(t, err, "error")
		assert.Equal(t, ruleProp.Audited, true)
		assert.Equal(t, ruleProp.ToolState, "Exploitable")
		assert.Equal(t, ruleProp.ToolStateIndex, 5)
		assert.Equal(t, ruleProp.ToolSeverity, "High")
		assert.Equal(t, ruleProp.ToolSeverityIndex, 3)
		assert.Equal(t, ruleProp.ToolAuditMessage, "Dummy comment.")
		assert.Equal(t, ruleProp.FortifyCategory, "Audit All")
		assert.Equal(t, ruleProp.AuditRequirementIndex, format.AUDIT_REQUIREMENT_GROUP_1_INDEX)
		assert.Equal(t, ruleProp.AuditRequirement, format.AUDIT_REQUIREMENT_GROUP_1_DESC)
		assert.Equal(t, ruleProp.CheckmarxSimilarityID, "") // ensure the existence of not applicable field (specific Checkmarx)
	})

	t.Run("Max retries set to 0: error raised", func(t *testing.T) {
		ruleProp := *new(format.SarifProperties)
		projectVersion := models.ProjectVersion{ID: 11037}
		auditData, _ := sys.GetAllIssueDetails(11037)
		err := integrateAuditData(&ruleProp, "DUMMYDUMMYDUMMY", sys, &projectVersion, auditData, filterSet, false, 0)
		assert.Error(t, err, "request failed: maximum number of retries reached, placeholder values will be set from now on for audit data")
	})

	t.Run("Max retries set to -1: fail silently", func(t *testing.T) {
		ruleProp := *new(format.SarifProperties)
		projectVersion := models.ProjectVersion{ID: 11037}
		auditData, _ := sys.GetAllIssueDetails(11037)
		err := integrateAuditData(&ruleProp, "DUMMYDUMMYDUMMY", sys, &projectVersion, auditData, filterSet, false, -1)
		assert.NoError(t, err)
	})
}
