package com.sap.piper.cm
import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertEquals

import org.hamcrest.Matchers
import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.GitUtils

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsScriptLoaderRule
import util.JenkinsShellCallRule
import util.JenkinsCredentialsRule
import util.JenkinsDockerExecuteRule
import util.Rules

import hudson.AbortException

public class ChangeManagementTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()

    private JenkinsShellCallRule script = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule logging = new JenkinsLoggingRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
        .around(thrown)
        .around(script)
        .around(logging)
        .around(new JenkinsCredentialsRule(this).withCredentials('me','user','password'))
        .around(dockerExecuteRule)

    @Test
    public void testRetrieveChangeDocumentIdOutsideGitWorkTreeTest() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve ChangeDocumentId. ' +
                             'Not in a git work tree. ' +
                             'ChangeDocumentId is extracted from git commit messages.')

        new ChangeManagement(nullScript, gitUtilsMock(false, new String[0])).getChangeDocumentId()
    }

    @Test
    public void testRetrieveChangeDocumentIdNothingFound() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve ChangeDocumentId from git commits.')

        new ChangeManagement(nullScript, gitUtilsMock(true, new String[0])).getChangeDocumentId()
    }

    @Test
    public void testRetrieveChangeDocumentIdReturnsArrayWithNullValue() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve ChangeDocumentId from git commits.')

        new ChangeManagement(nullScript, gitUtilsMock(true, (String[])[ null ])).getChangeDocumentId()
    }

    @Test
    public void testRetrieveChangeDocumentNotUnique() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Multiple ChangeDocumentIds found')

        String[] changeIds = [ 'a', 'b' ]
        new ChangeManagement(nullScript, gitUtilsMock(true, changeIds)).getChangeDocumentId()
    }

    @Test
    public void testRetrieveChangeDocumentSameChangeIdFoundTwice() {

        String[] changeIds = [ 'a', 'a' ]
        def changeID = new ChangeManagement(nullScript, gitUtilsMock(true, changeIds)).getChangeDocumentId()

        assert changeID == 'a'
    }

    @Test
    public void testIsChangeInDevelopmentReturnsTrueWhenChangeIsInDevelopent() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, "cmclient.*is-change-in-development -cID '001'", 0)
        boolean inDevelopment = new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'me')

        assertThat(inDevelopment, is(equalTo(true)))
        assertThat(script.shell[0], allOf(containsString("cmclient"),
                                            containsString("-u 'user'"),
                                            containsString("-p 'password'"),
                                            containsString("-e 'endpoint'"),
                                            containsString('is-change-in-development'),
                                            containsString("-cID '001'"),
                                            containsString("-t SOLMAN")))
    }

    @Test
    public void testIsChangeInDevelopmentReturnsFalseWhenChangeIsNotInDevelopent() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, "cmclient.*is-change-in-development -cID '001'", 3)

        boolean inDevelopment = new ChangeManagement(nullScript, null)
                                    .isChangeInDevelopment('001',
                                                           'endpoint',
                                                           'me')

        assertThat(inDevelopment, is(equalTo(false)))
    }

    @Test
    public void testIsChangeInDevelopmentThrowsExceptionWhenCMClientReturnsUnexpectedExitCode() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve status for change document \'001\'. Does this change exist? Return code from cmclient: 1.')

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, "cmclient.*is-change-in-development -cID '001'", 1)
        new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'me')
    }

    @Test
    public void testGetCommandLineWithoutCMClientOpts() {
        String commandLine = new ChangeManagement(nullScript, null)
            .getCMCommandLine(BackendType.SOLMAN,
                              'https://example.org/cm',
                              "me",
                              "topSecret",
                              "the-command",
                              ["-key1", "val1", "-key2", "val2"])
        commandLine = commandLine.replaceAll(' +', " ")
        assertThat(commandLine, not(containsString("CMCLIENT_OPTS")))
        assertThat(commandLine, containsString("cmclient -e 'https://example.org/cm' -u 'me' -p 'topSecret' -t SOLMAN the-command -key1 val1 -key2 val2"))
    }

@Test
public void testGetCommandLineWithCMClientOpts() {
    String commandLine = new ChangeManagement(nullScript, null)
        .getCMCommandLine(BackendType.SOLMAN,
                          'https://example.org/cm',
                          "me",
                          "topSecret",
                          "the-command",
                          ["-key1", "val1", "-key2", "val2"],
                          '-Djavax.net.debug=all')
    commandLine = commandLine.replaceAll(' +', " ")
    assertThat(commandLine, containsString('export CMCLIENT_OPTS="-Djavax.net.debug=all"'))
}

    @Test
    public void testCreateTransportRequestSOLMANSucceeds() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".*cmclient.*create-transport -cID 001 -dID 002.*", '004')
        def transportRequestId = new ChangeManagement(nullScript).createTransportRequestSOLMAN( '001', '002', '003', 'me')

        // the check for the transportRequestID is sufficient. This checks implicit the command line since that value is
        // returned only in case the shell call matches.
        assert transportRequestId == '004'

    }

    @Test
    public void testCreateTransportRequestRFCSucceeds() {

        script.setReturnValue('cts createTransportRequest', '{"REQUESTID":"XYZK9000004"}')

        def transportRequestId = new ChangeManagement(nullScript).createTransportRequestRFC(
            'rfc', // docker image
            [], // docker options
            'https://example.org/rfc', // endpoint
            '01', // client
            '001', // instance
            'me', // credentialsId
            'Lorem ipsum' // description
        )

        assert dockerExecuteRule.dockerParams.dockerImage == 'rfc'

        assert dockerExecuteRule.dockerParams.dockerOptions == [
            '--env TRANSPORT_DESCRIPTION=Lorem ipsum',
            '--env ABAP_DEVELOPMENT_INSTANCE=001',
            '--env ABAP_DEVELOPMENT_CLIENT=01',
            '--env ABAP_DEVELOPMENT_SERVER=https://example.org/rfc',
            '--env ABAP_DEVELOPMENT_USER=user',
            '--env ABAP_DEVELOPMENT_PASSWORD=password'
        ]

        assert transportRequestId == 'XYZK9000004'

    }

    @Test
    public void testCreateTransportRequestRFCFails() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot create transport request: script returned exit code 3')

        script.setReturnValue('cts createTransportRequest',
            { throw new AbortException('script returned exit code 3')})

        def transportRequestId = new ChangeManagement(nullScript).createTransportRequestRFC(
            'rfc', // docker image
            [], // docker options
            'https://example.org/rfc', // endpoint
            '01', // client
            '001', // instance
            'me', // credentialsId
            'Lorem ipsum' // description
        )
    }

    @Test
    public void testCreateTransportRequestCTSSucceeds() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'cmclient.* -t CTS .*create-transport -tt W -ts XYZ -d "desc 123"$', '004')
        def transportRequestId = new ChangeManagement(nullScript)
            .createTransportRequestCTS(
                'W', // transport type
                'XYZ', // target system
                'desc 123', // description
                'https://example.org/cm',
                'me')

        // the check for the transportRequestID is sufficient. This checks implicit the command line since that value is
        // returned only in case the shell call matches.
        assert transportRequestId == '004'

    }

    @Test
    public void testUploadFileToTransportSucceedsSOLMAN() {

        // the regex provided below is an implicit check that the command line is fine.
        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'upload-file-to-transport.*-cID 001 -tID 002 XXX "/path"', 0)

        new ChangeManagement(nullScript).uploadFileToTransportRequestSOLMAN(
            '001',
            '002',
            'XXX',
            '/path',
            'https://example.org/cm',
            'me')

        // no assert required here, since the regex registered above to the script rule is an implicit check for
        // the command line.
    }

    @Test
    public void testUploadFileToTransportSucceedsCTS() {

        // the regex provided below is an implicit check that the command line is fine.
        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, '-t CTS upload-file-to-transport -tID 002 "/path"', 0)

        new ChangeManagement(nullScript).uploadFileToTransportRequestCTS(
            '002',
            '/path',
            'https://example.org/cm',
            'me')

        // no assert required here, since the regex registered above to the script rule is an implicit check for
        // the command line.
    }

    @Test
    public void testUploadFileToTransportSucceedsRFC() {

        new ChangeManagement(nullScript).uploadFileToTransportRequestRFC(
            'rfc',
            [],
            '002', //transportRequestId
            '001', // applicationId
            'https://example.org/mypath/deployArtifact.zip',
            'https://example.org/rfc',
            'me',
            '01', //developmentInstance
            '00', // developmentClient
            'Lorem ipsum', // applicationDescription
            'XYZ' // abapPackage
            )

            assert dockerExecuteRule.dockerParams.dockerImage == 'rfc'

            assert dockerExecuteRule.dockerParams.dockerOptions ==
            [
                '--env ABAP_DEVELOPMENT_INSTANCE=01',
                '--env ABAP_DEVELOPMENT_CLIENT=00',
                '--env ABAP_APPLICATION_NAME=001',
                '--env ABAP_APPLICATION_DESC=Lorem ipsum',
                '--env ABAP_PACKAGE=XYZ',
                '--env ZIP_FILE_URL=https://example.org/mypath/deployArtifact.zip',
                '--env ABAP_DEVELOPMENT_SERVER=https://example.org/rfc',
                '--env ABAP_DEVELOPMENT_USER=user',
                '--env ABAP_DEVELOPMENT_PASSWORD=password',
            ]

            assertThat(script.shell, contains('cts uploadToABAP:002'))
    }

    @Test
    public void testUploadFileToTransportFailsRFC() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot upload file into transport request. Return code from rfc client: 1.')

        script.setReturnValue('cts uploadToABAP:002', 1)

        new ChangeManagement(nullScript).uploadFileToTransportRequestRFC(
            'rfc',
            [],
            '002', //transportRequestId
            '001', // applicationId
            'https://example.org/mypath/deployArtifact.zip',
            'https://example.org/rfc',
            'me',
            '01', //developmentInstance
            '00', // developmentClient
            'Lorem ipsum', // applicationDescription
            'XYZ' // abapPackage
            )
    }

    @Test
    public void testUploadFileToTransportFailsSOLMAN() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage("Cannot upload file into transport request. " +
            "Return code from cm client: 1.")

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX,, 'upload-file-to-transport', 1)

        new ChangeManagement(nullScript).uploadFileToTransportRequestSOLMAN(
            '001',
            '002',
            'XXX',
            '/path',
            'https://example.org/cm',
            'me')
    }

    @Test
    public void testReleaseTransportRequestSucceedsSOLMAN() {

        // the regex provided below is an implicit check that the command line is fine.
        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, '-t SOLMAN release-transport.*-cID 001.*-tID 002', 0)

        new ChangeManagement(nullScript).releaseTransportRequestSOLMAN(
            '001',
            '002',
            'https://example.org',
            'me',
            'openSesame')

        // no assert required here, since the regex registered above to the script rule is an implicit check for
        // the command line.
    }

    @Test
    public void testReleaseTransportRequestSucceedsCTS() {

        // the regex provided below is an implicit check that the command line is fine.
        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, '-t CTS export-transport.*-tID 002', 0)

        new ChangeManagement(nullScript).releaseTransportRequestCTS(
            '002',
            'https://example.org',
            'me',
            'openSesame')

        // no assert required here, since the regex registered above to the script rule is an implicit check for
        // the command line.
    }

    @Test
    public void testReleaseTransportRequestSucceedsRFC() {

        new ChangeManagement(nullScript).releaseTransportRequestRFC(
            '',
            [],
            '002',
            'https://example.org',
            '001',
            'me')

        List stringDockerOptions = []
        for(item in dockerExecuteRule.dockerParams.dockerOptions)
        {
            stringDockerOptions = stringDockerOptions.plus([item.toString()])
        }
        assertThat(stringDockerOptions, hasItem('--env ABAP_DEVELOPMENT_SERVER=https://example.org'))
        assertThat(stringDockerOptions, hasItem('--env ABAP_DEVELOPMENT_USER=user'))
        assertThat(stringDockerOptions, hasItem('--env ABAP_DEVELOPMENT_PASSWORD=password'))
        assertThat(stringDockerOptions, hasItem('--env ABAP_DEVELOPMENT_CLIENT=001'))
        assertThat(script.shell, hasItem('cts releaseTransport:002'))
    }

    @Test
    public void testReleaseTransportRequestFailsSOLMAN() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage("Cannot release Transport Request '002'. Return code from cmclient: 1.")

        // the regex provided below is an implicit check that the command line is fine.
        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'release-transport.*-cID 001.*-tID 002', 1)

        new ChangeManagement(nullScript).releaseTransportRequestSOLMAN(
            '001',
            '002',
            'https://example.org',
            'me')
    }

    private GitUtils gitUtilsMock(boolean insideWorkTree, String[] changeIds) {
        return new GitUtils() {
            public boolean insideWorkTree() {
                return insideWorkTree
            }

            public String[] extractLogLines(
                         String filter = '',
                         String from = 'origin/master',
                         String to = 'HEAD',
                         String format = '%b') {
                return changeIds
            }
        }
    }
}
