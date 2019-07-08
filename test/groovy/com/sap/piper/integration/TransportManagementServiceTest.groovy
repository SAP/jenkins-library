package com.sap.piper.integration

import hudson.AbortException
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsErrorRule
import util.JenkinsLoggingRule
import util.JenkinsReadFileRule
import util.JenkinsReadJsonRule
import util.JenkinsShellCallRule
import util.Rules

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.endsWith
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.startsWith
import static org.junit.Assert.assertThat

class TransportManagementServiceTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsErrorRule(this))
        .around(new JenkinsReadJsonRule(this))
        .around(shellRule)
        .around(loggingRule)
        .around(new JenkinsReadFileRule(this, 'test/resources/TransportManagementService'))
        .around(thrown)

    @Test
    void retrieveOAuthToken__successfully() {

        def uaaUrl = 'http://dummy.com/oauth'
        def clientId = 'myId'
        def clientSecret = 'mySecret'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '200')

        def tms = new TransportManagementService(nullScript, [:])
        def token = tms.authentication(uaaUrl, clientId, clientSecret)

        def oAuthShellCall = shellRule.shell[0]

        assertThat(loggingRule.log, containsString("[TransportManagementService] OAuth Token retrieval started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] OAuth Token retrieved successfully."))
        assertThat(oAuthShellCall, startsWith("#!/bin/sh -e"))
        assertThat(oAuthShellCall, endsWith("curl -XPOST -u '${clientId}':'${clientSecret}' -o responseAuth.txt --write-out '%{http_code}' '${uaaUrl}/oauth/token/?grant_type=client_credentials&response_type=token'"))
        assertThat(token, is('myOAuthToken'))
    }

    @Test
    void retrieveOAuthToken__withHttpErrorResponse__throwsError() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '401')

        thrown.expect(AbortException.class)
        thrown.expectMessage("[TransportManagementService] Retrieval of OAuth-Token failed. HTTP-Status: 401 \n [ERROR] Response:")

        def tms = new TransportManagementService(nullScript, [:])
        tms.authentication("", "", "")

    }

    @Test
    void retrieveOAuthToken__inVerboseMode__yieldsMoreEchos() {

        def uaaUrl = 'http://dummy.com/oauth'
        def clientId = 'myId'
        def clientSecret = 'mySecret'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '200')

        def tms = new TransportManagementService(nullScript, [verbose: true])
        tms.authentication(uaaUrl, clientId, clientSecret)

        assertThat(loggingRule.log, containsString("[TransportManagementService] OAuth Token retrieval started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] UAA-URL: '${uaaUrl}', ClientId: '${clientId}'"))
        assertThat(loggingRule.log, containsString("\"access_token\": \"myOAuthToken\""))
        assertThat(loggingRule.log, containsString("[TransportManagementService] OAuth Token retrieved successfully."))
    }

    @Test
    void uploadFileToTMS__successfully() {

        def url = 'http://dummy.com/oauth'
        def token = 'myToken'
        def file = 'myFile.mtar'
        def namedUser = 'myUser'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '200')

        def tms = new TransportManagementService(nullScript, [:])
        def responseDetails = tms.uploadFileToTMS(url, token, file, namedUser)

        def oAuthShellCall = shellRule.shell[0]

        assertThat(loggingRule.log, containsString("[TransportManagementService] File upload started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] File upload successful."))
        assertThat(oAuthShellCall, startsWith("#!/bin/sh -e"))
        assertThat(oAuthShellCall, endsWith("curl -XPOST -H 'Authorization: Bearer ${token}' -F 'file=@${file}' -F 'namedUser=${namedUser}' -o responseFileUpload.txt --write-out '%{http_code}' '${url}/v2/files/upload'"))
        assertThat(responseDetails, hasEntry("fileId", 1234))
    }

    @Test
    void uploadFileToTMS__withHttpErrorResponse__throwsError() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '400')

        thrown.expect(AbortException.class)
        thrown.expectMessage("[TransportManagementService] File upload failed. HTTP-Status: 400 \n [ERROR] Response:")

        def tms = new TransportManagementService(nullScript, [:])
        tms.uploadFileToTMS("", "", "", "")

    }

    @Test
    void uploadFileToTMS__inVerboseMode__yieldsMoreEchos() {

        def url = 'http://dummy.com/oauth'
        def token = 'myToken'
        def file = 'myFile.mtar'
        def namedUser = 'myUser'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '200')

        def tms = new TransportManagementService(nullScript, [verbose: true])
        tms.uploadFileToTMS(url, token, file, namedUser)

        assertThat(loggingRule.log, containsString("[TransportManagementService] File upload started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] URL: '${url}', File: '${file}'"))
        assertThat(loggingRule.log, containsString("\"fileId\": 1234"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] File upload successful."))
    }

    @Test
    void uploadFileToNode__successfully() {

        def url = 'http://dummy.com/oauth'
        def token = 'myToken'
        def nodeName = 'myNode'
        def fileId = 1234
        def description = "My description."
        def namedUser = 'myUser'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '200')

        def tms = new TransportManagementService(nullScript, [:])
        def responseDetails = tms.uploadFileToNode(url, token, nodeName, fileId, description, namedUser)

        def body = """{ "nodeName": "${nodeName}", "contentType": "MTA", "description": "${description}", "storageType": "FILE", "namedUser": "${namedUser}", "entries": [ { "uri": ${fileId} } ] }"""

        def oAuthShellCall = shellRule.shell[0]

        assertThat(loggingRule.log, containsString("[TransportManagementService] Node upload started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Node upload successful."))
        assertThat(oAuthShellCall, startsWith("#!/bin/sh -e"))
        assertThat(oAuthShellCall, endsWith("curl -XPOST -H 'Authorization: Bearer ${token}' -H 'Content-Type: application/json' -d '${body}' -o responseNodeUpload.txt --write-out '%{http_code}' '${url}/v2/nodes/upload'"))
        assertThat(responseDetails, hasEntry("upload", "success"))
    }

    @Test
    void uploadFileToNode__withHttpErrorResponse__throwsError() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '403')

        thrown.expect(AbortException.class)
        thrown.expectMessage("[TransportManagementService] Node upload failed. HTTP-Status: 403 \n [ERROR] Response:")

        def tms = new TransportManagementService(nullScript, [:])
        tms.uploadFileToNode("", "", "", 0, "", "")

    }

    @Test
    void uploadFileToNode__inVerboseMode__yieldsMoreEchos() {

        def url = 'http://dummy.com/oauth'
        def token = 'myToken'
        def nodeName = 'myNode'
        def fileId = 1234
        def description = "My description."
        def namedUser = 'myUser'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '200')

        def tms = new TransportManagementService(nullScript, [verbose: true])
        tms.uploadFileToNode(url, token, nodeName, fileId, description, namedUser)

        assertThat(loggingRule.log, containsString("[TransportManagementService] Node upload started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] URL: '${url}', NodeName: '${nodeName}', FileId: '${fileId}'"))
        assertThat(loggingRule.log, containsString("\"upload\": \"success\""))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Node upload successful."))
    }

}
