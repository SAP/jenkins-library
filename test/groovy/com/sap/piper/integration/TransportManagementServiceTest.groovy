package com.sap.piper.integration

import hudson.AbortException
import org.junit.Rule
import org.junit.Test
import org.junit.Ignore
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class TransportManagementServiceTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, 'test/resources/TransportManagementService')
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, ['responseFileUpload.txt'])

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsErrorRule(this))
        .around(new JenkinsReadJsonRule(this))
        .around(shellRule)
        .around(loggingRule)
        .around(readFileRule)
        .around(fileExistsRule)
        .around(thrown)

    @Test
    void retrieveOAuthToken__successfully() {
        Map requestParams
        helper.registerAllowedMethod('httpRequest', [Map.class], { m ->
            requestParams = m
            return [content: '{ "access_token": "myOAuthToken" }', status: 200]
        })

        def uaaUrl = 'http://dummy.sap.com/oauth'
        def clientId = 'myId'
        def clientSecret = 'mySecret'

        def tms = new TransportManagementService(nullScript, [:])
        def token = tms.authentication(uaaUrl, clientId, clientSecret)

        assertThat(loggingRule.log, containsString("[TransportManagementService] OAuth Token retrieval started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] OAuth Token retrieved successfully."))
        assertThat(token, is('myOAuthToken'))
        assertThat(requestParams, hasEntry('url', "${uaaUrl}/oauth/token/?grant_type=client_credentials&response_type=token"))
        assertThat(requestParams, hasEntry('requestBody', "grant_type=password&username=${clientId}&password=${clientSecret}".toString()))
        assertThat(requestParams.customHeaders[1].value, is("Basic ${"${clientId}:${clientSecret}".bytes.encodeBase64()}"))
    }

    @Test
    void retrieveOAuthToken__inVerboseMode__yieldsMoreEchos() {
        helper.registerAllowedMethod('httpRequest', [Map.class], {
            return [content: '{ "access_token": "myOAuthToken" }', status: 200]
        })

        def uaaUrl = 'http://dummy.sap.com/oauth'
        def clientId = 'myId'
        def clientSecret = 'mySecret'

        def tms = new TransportManagementService(nullScript, [verbose: true])
        def token = tms.authentication(uaaUrl, clientId, clientSecret)
        assertThat(loggingRule.log, containsString("[TransportManagementService] OAuth Token retrieval started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] UAA-URL: '${uaaUrl}', ClientId: '${clientId}'"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] OAuth Token retrieved successfully."))
        assertThat(token, is('myOAuthToken'))
    }

    @Test
    void retrieveOAuthToken__failure() {
        def uaaUrl = 'http://dummy.sap.com/oauth'
        def clientId = 'myId'
        def clientSecret = 'mySecret'
        def responseStatusCode = 400
        def responseContent = 'Here an error message is expected (THIS PART IS HERE TO CHECK THAT ERROR MESSAGE IS EXPOSED IN NON-VERBOSE MODE)'

        thrown.expect(AbortException)
        thrown.expectMessage("[TransportManagementService] OAuth Token retrieval failed (HTTP status code '${responseStatusCode}'). Response content '${responseContent}'.")
        loggingRule.expect("[TransportManagementService] OAuth Token retrieval started.")

        helper.registerAllowedMethod('httpRequest', [Map.class], {
            return [content: responseContent, status: responseStatusCode]
        })

        def tms = new TransportManagementService(nullScript, [verbose: false])
        tms.authentication(uaaUrl, clientId, clientSecret)
    }

    @Test
    void retrieveOAuthToken__failure__status__400__inVerboseMode() {
        def uaaUrl = 'http://dummy.sap.com/oauth'
        def clientId = 'myId'
        def clientSecret = 'mySecret'
        def responseStatusCode = 400
        def responseContent = 'Here an error message is expected (THIS PART IS HERE TO CHECK THAT ERROR MESSAGE IS EXPOSED IN VERBOSE MODE)'

        thrown.expect(AbortException)
        thrown.expectMessage("[TransportManagementService] OAuth Token retrieval failed (HTTP status code '${responseStatusCode}'). Response content '${responseContent}'.")
        loggingRule.expect("[TransportManagementService] OAuth Token retrieval started.")
        loggingRule.expect("[TransportManagementService] UAA-URL: '${uaaUrl}', ClientId: '${clientId}'")

        helper.registerAllowedMethod('httpRequest', [Map.class], {
            return [content: responseContent, status: responseStatusCode]
        })

        def tms = new TransportManagementService(nullScript, [verbose: true])
        tms.authentication(uaaUrl, clientId, clientSecret)
    }

    @Test
    void uploadFile__successfully() {

        def url = 'http://dummy.sap.com'
        def token = 'myToken'
        def file = 'myFile.mtar'
        def namedUser = 'myUser'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX,'.*curl.*', '201')

        def tms = new TransportManagementService(nullScript, [:])
        def responseDetails = tms.uploadFile(url, token, file, namedUser)

        // replace needed since the curl command is spread over several lines.
        def oAuthShellCall = shellRule.shell[0].replaceAll('\\\\ ', '')

        assertThat(loggingRule.log, containsString("[TransportManagementService] File upload started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] File upload successful."))
        assertThat(oAuthShellCall, startsWith("#!/bin/sh -e "))
        assertThat(oAuthShellCall, endsWith("curl --write-out '%{response_code}' -H 'Authorization: Bearer ${token}' -F 'file=@${file}' -F 'namedUser=${namedUser}' --output responseFileUpload.txt '${url}/v2/files/upload'"))
        assertThat(responseDetails, hasEntry("fileId", 1234))
    }

    @Test
    void uploadFile__verboseMode__withHttpErrorResponse__throwsError() {

        def url = 'http://dummy.sap.com'
        def token = 'myWrongToken'
        def file = 'myFile.mtar'
        def namedUser = 'myUser'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '400')

        readFileRule.files << [ 'responseFileUpload.txt': 'Something went wrong during file upload (WE ARE IN VERBOSE MODE)']

        thrown.expect(AbortException.class)
        thrown.expectMessage('Unexpected response code received from file upload (400). 201 expected')

        loggingRule.expect('[TransportManagementService] URL: \'http://dummy.sap.com\', File: \'myFile.mtar\'')
        loggingRule.expect('[TransportManagementService] Response body: Something went wrong during file upload (WE ARE IN VERBOSE MODE)')

        // The log entries which are present in non verbose mode must be present in verbose mode also, of course
        loggingRule.expect('[TransportManagementService] File upload started.')
        loggingRule.expect('[TransportManagementService] Unexpected response code received from file upload (400). 201 expected. Response body: Something went wrong during file upload')

        new TransportManagementService(nullScript, [verbose:true])
            .uploadFile(url, token, file, namedUser)
    }

    @Test
    void uploadFile__NonVerboseMode__withHttpErrorResponse__throwsError() {

        def url = 'http://dummy.sap.com'
        def token = 'myWrongToken'
        def file = 'myFile.mtar'
        def namedUser = 'myUser'

        // 418 (tea-pot)? Other than 400 which is used in verbose mode in order to be sure that we don't mix up
        // with any details from the other test for the verbose mode. The log message below (Unexpected response code ...)
        // reflects that 418 instead of 400.
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '418')

        readFileRule.files << [ 'responseFileUpload.txt': 'Something went wrong during file upload. WE ARE IN NON VERBOSE MODE.']

        thrown.expect(AbortException.class)
        thrown.expectMessage('Unexpected response code received from file upload (418). 201 expected')

        loggingRule.expect('[TransportManagementService] File upload started.')
        loggingRule.expect('[TransportManagementService] Unexpected response code received from file upload (418). 201 expected. Response body: Something went wrong during file upload. WE ARE IN NON VERBOSE MODE.')

        new TransportManagementService(nullScript, [verbose:false])
            .uploadFile(url, token, file, namedUser)
    }

    @Test
    void uploadFile__inVerboseMode__yieldsMoreEchos() {

        def url = 'http://dummy.sap.com'
        def token = 'myToken'
        def file = 'myFile.mtar'
        def namedUser = 'myUser'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".* curl .*", '201')
        fileExistsRule.existingFiles.add('responseFileUpload.txt')
        readFileRule.files.put('responseFileUpload.txt', '{"fileId": 1234}')

        def tms = new TransportManagementService(nullScript, [verbose: true])
        tms.uploadFile(url, token, file, namedUser)

        assertThat(loggingRule.log, containsString("[TransportManagementService] File upload started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] URL: '${url}', File: '${file}'"))
        assertThat(loggingRule.log, containsString("\"fileId\": 1234"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] File upload successful."))
    }

    @Test
    void uploadFileToNode__successfully() {
        Map requestParams
        helper.registerAllowedMethod('httpRequest', [Map.class], { m ->
            requestParams = m
            return [content: '{ "upload": "success" }', status: 200]
        })

        def url = 'http://dummy.sap.com'
        def token = 'myToken'
        def nodeName = 'myNode'
        def fileId = 1234
        def description = "My description."
        def namedUser = 'myUser'

        def tms = new TransportManagementService(nullScript, [:])
        def responseDetails = tms.uploadFileToNode(url, token, nodeName, fileId, description, namedUser)

        def bodyRegEx = /^\{\s+"nodeName":\s+"myNode",\s+"contentType":\s+"MTA",\s+"description":\s+"My\s+description.",\s+"storageType":\s+"FILE",\s+"namedUser":\s+"myUser",\s+"entries":\s+\[\s+\{\s+"uri":\s+1234\s+}\s+]\s+}$/

        assertThat(loggingRule.log, containsString("[TransportManagementService] Node upload started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Node upload successful."))
        assertThat(requestParams, hasEntry('url', "${url}/v2/nodes/upload"))
        assert requestParams.requestBody ==~ bodyRegEx
        assertThat(requestParams.customHeaders[0].value, is("Bearer ${token}"))
        assertThat(responseDetails, hasEntry("upload", "success"))
    }

    @Test
    void uploadFileToNode__inVerboseMode__yieldsMoreEchos() {
        def url = 'http://dummy.sap.com'
        def token = 'myToken'
        def nodeName = 'myNode'
        def fileId = 1234
        def description = "My description."
        def namedUser = 'myUser'
        def responseContent = '{ "upload": "success" }'

        helper.registerAllowedMethod('httpRequest', [Map.class], {
            return [content: responseContent, status: 200]
        })

        def tms = new TransportManagementService(nullScript, [verbose: true])
        tms.uploadFileToNode(url, token, nodeName, fileId, description, namedUser)

        assertThat(loggingRule.log, containsString("[TransportManagementService] Node upload started."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] URL: '${url}', NodeName: '${nodeName}', FileId: '${fileId}'"))
        assertThat(loggingRule.log, containsString("\"upload\": \"success\""))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Node upload successful. Response content '${responseContent}'."))
    }

    @Test
    void uploadFileToNode__failure() {
        def url = 'http://dummy.sap.com'
        def token = 'myToken'
        def nodeName = 'myNode'
        def fileId = 1234
        def description = "My description."
        def namedUser = 'myUser'
        def responseStatusCode = 400
        def responseContent = '{ "errorType": "TsInternalServerErrorException", "message": "The application has encountered an unexpected error (THIS PART IS HERE TO CHECK THAT ERROR MESSAGE IS EXPOSED IN NON-VERBOSE MODE)." }'

        thrown.expect(AbortException)
        thrown.expectMessage("[TransportManagementService] Node upload failed (HTTP status code '${responseStatusCode}'). Response content '${responseContent}'.")
        loggingRule.expect("[TransportManagementService] Node upload started.")

        helper.registerAllowedMethod('httpRequest', [Map.class], {
            return [content: responseContent, status: responseStatusCode]
        })

        def tms = new TransportManagementService(nullScript, [verbose: false])
        tms.uploadFileToNode(url, token, nodeName, fileId, description, namedUser)
    }

    @Test
    void uploadFileToNode__failure__status__400__inVerboseMode() {
        def url = 'http://dummy.sap.com'
        def token = 'myToken'
        def nodeName = 'myNode'
        def fileId = 1234
        def description = "My description."
        def namedUser = 'myUser'
        def responseStatusCode = 400
        def responseContent = '{ "errorType": "TsInternalServerErrorException", "message": "The application has encountered an unexpected error (THIS PART IS HERE TO CHECK THAT ERROR MESSAGE IS EXPOSED IN VERBOSE MODE)." }'

        thrown.expect(AbortException)
        thrown.expectMessage("[TransportManagementService] Node upload failed (HTTP status code '${responseStatusCode}'). Response content '${responseContent}'.")
        loggingRule.expect("[TransportManagementService] Node upload started.")
        loggingRule.expect("[TransportManagementService] URL: '${url}', NodeName: '${nodeName}', FileId: '${fileId}'")

        helper.registerAllowedMethod('httpRequest', [Map.class], {
            return [content: responseContent, status: responseStatusCode]
        })

        def tms = new TransportManagementService(nullScript, [verbose: true])
        tms.uploadFileToNode(url, token, nodeName, fileId, description, namedUser)
    }
}
