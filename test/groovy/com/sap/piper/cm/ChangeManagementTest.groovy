package com.sap.piper.cm

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

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
import util.Rules

import hudson.AbortException

public class ChangeManagementTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()

    private JenkinsShellCallRule script = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule logging = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
        .around(thrown)
        .around(script)
        .around(logging)
        .around(new JenkinsCredentialsRule(this).withCredentials('me','user','password'))

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

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'cmclient.*is-change-in-development -cID "001"', 0)
        boolean inDevelopment = new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'me')

        assertThat(inDevelopment, is(equalTo(true)))
        assertThat(script.shell[0], allOf(containsString("cmclient"),
                                            containsString("-u 'user'"),
                                            containsString("-p 'password'"),
                                            containsString("-e 'endpoint'"),
                                            containsString('is-change-in-development'),
                                            containsString('-cID "001"'),
                                            containsString("-t SOLMAN")))
    }

    @Test
    public void testIsChangeInDevelopmentReturnsFalseWhenChangeIsNotInDevelopent() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'cmclient.*is-change-in-development -cID "001"', 3)

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

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'cmclient.*is-change-in-development -cID "001"', 1)
        new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'me')
    }

    @Test
    public void testGetCommandLineWithoutCMClientOpts() {
        String commandLine = new ChangeManagement(nullScript, null)
            .getCMCommandLine('https://example.org/cm',
                              "me",
                              "topSecret",
                              "the-command",
                              [new ChangeManagement.KeyValue("key1", "val1"), new ChangeManagement.KeyValue("key2", "val2")])
        commandLine = commandLine.replaceAll(' +', " ")
        assertThat(commandLine, not(containsString("CMCLIENT_OPTS")))
        assertThat(commandLine, containsString("cmclient -e 'https://example.org/cm' -u 'me' -p 'topSecret' -t SOLMAN the-command -key1 \"val1\" -key2 \"val2\""))
    }

@Test
public void testGetCommandLineWithCMClientOpts() {
    String commandLine = new ChangeManagement(nullScript, null)
        .getCMCommandLine('https://example.org/cm',
                          "me",
                          "topSecret",
                          "the-command",
                          [new ChangeManagement.KeyValue("key1", "val1"), new ChangeManagement.KeyValue("key2", "val2")],
                          '-Djavax.net.debug=all')
    commandLine = commandLine.replaceAll(' +', " ")
    assertThat(commandLine, containsString('export CMCLIENT_OPTS="-Djavax.net.debug=all"'))
}

    @Test
    public void testGetCommandLineWithBlanksInFilePath() {
        String commandLine = new ChangeManagement(nullScript, null)
            .getCMCommandLine('https://example.org/cm',
                              "me",
                              "topSecret",
                              "the-command",
                              [new ChangeManagement.KeyValue("key1", "/file path")])
        commandLine = commandLine.replaceAll(' +', " ")
        assertThat(commandLine, containsString("cmclient -e 'https://example.org/cm' -u 'me' -p 'topSecret' -t SOLMAN the-command -key1 \"/file path\""))
    }

    @Test
    public void testCreateTransportRequestSucceeds() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*cmclient.*create-transport -cID "001" -dID "002".*', '004')
        def transportRequestId = new ChangeManagement(nullScript).createTransportRequest('001', '002', '003', 'me')

        // the check for the transportRequestID is sufficient. This checks implicit the command line since that value is
        // returned only in case the shell call matches.
        assert transportRequestId == '004'

    }

    @Test
    public void testCreateTransportRequestFails() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*upload-file-to-transport.*', 1)

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot upload file \'/path\' for change document \'001\''+
                             ' with transport request \'002\'. Return code from cmclient: 1.')

        new ChangeManagement(nullScript).uploadFileToTransportRequest('001',
                                                                      '002',
                                                                      'XXX',
                                                                      '/path',
                                                                      'https://example.org/cm',
                                                                      'me')
    }

    @Test
    public void testUploadFileToTransportSucceeds() {

        // the regex provided below is an implicit check that the command line is fine.
        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'upload-file-to-transport.*-cID "001" -tID "002" XXX "/path"', 0)

        new ChangeManagement(nullScript).uploadFileToTransportRequest('001',
            '002',
            'XXX',
            '/path',
            'https://example.org/cm',
            'me')

        // no assert required here, since the regex registered above to the script rule is an implicit check for
        // the command line.
    }

    @Test
    public void testUploadFileToTransportFails() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage("Cannot upload file '/path' for change document '001' with transport request '002'. " +
            "Return code from cmclient: 1.")

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'upload-file-to-transport', 1)

        new ChangeManagement(nullScript).uploadFileToTransportRequest('001',
            '002',
            'XXX',
            '/path',
            'https://example.org/cm',
            'me')
    }

    @Test
    public void testReleaseTransportRequestSucceeds() {

        // the regex provided below is an implicit check that the command line is fine.
        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'release-transport.*-cID "001".*-tID "002"', 0)

        new ChangeManagement(nullScript).releaseTransportRequest('001',
            '002',
            'https://example.org',
            'me',
            'openSesame')

        // no assert required here, since the regex registered above to the script rule is an implicit check for
        // the command line.
    }

    @Test
    public void testReleaseTransportRequestFails() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage("Cannot release Transport Request '002'. Return code from cmclient: 1.")

        // the regex provided below is an implicit check that the command line is fine.
        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'release-transport.*-cID "001".*-tID "002"', 1)

        new ChangeManagement(nullScript).releaseTransportRequest('001',
            '002',
            'https://example.org',
            'me',
            'openSesame')
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
    
    @Test
    public void opt_keyvalue_key_missing() {

        thrown.expect(NullPointerException)
        new ChangeManagement.KeyValue(null,"value")
    }

    @Test
    public void opt_keyvalue_value_missing() {

        thrown.expect(NullPointerException)
        new ChangeManagement.KeyValue("key",null)
    }

    @Test
    public void opt_keyvalue_tostring() {

        assert new ChangeManagement.KeyValue("key","value").toString() == '-key "value"'
        assert new ChangeManagement.KeyValue("key","value").setQuotes(false).toString() == '-key value'
    }

    @Test
    public void opt_switch_key_missing() {

        thrown.expect(NullPointerException)
        new ChangeManagement.Switch(null)
    }

    @Test
    public void opt_switch_tostring() {

        assert new ChangeManagement.Switch("key").toString() == "-key"
    }
    
    @Test
    public void opt_value_data_missing() {

        thrown.expect(NullPointerException)
        new ChangeManagement.Value(null)
    }

    @Test
    public void opt_value_tostring() {

        assert new ChangeManagement.Value("value").toString() == '"value"'
        assert new ChangeManagement.Value("value").setQuotes(false).toString() == "value"
    }
}
