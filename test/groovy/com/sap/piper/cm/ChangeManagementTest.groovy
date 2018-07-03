package com.sap.piper.cm

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.nullValue
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
import util.JenkinsShellCallRule
import util.Rules

public class ChangeManagementTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()

    private JenkinsShellCallRule script = new JenkinsShellCallRule(this)
	private JenkinsLoggingRule logging = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
        .around(thrown)
        .around(script)
        .around(logging)

	@Test
	public void testGetChangeIdFromConfigWhenProvidedInsideConfig() {
		String[] viaGitUtils = ['0815']
		def changeDocumentId = new ChangeManagement(nullScript, gitUtilsMock(false, viaGitUtils))
			.getChangeDocumentId([changeDocumentId: '0042'])

		assertThat(logging.log, containsString('[INFO] Use changeDocumentId \'0042\' from configuration.'))
		assertThat(changeDocumentId, is(equalTo('0042')))
	}
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
    public void testRetrieveChangeDocumentWithUniqueResult() {

        String[] changeIds = [ 'a' ];

        def params = [ git_from: 'origin/master',
                       git_to: 'HEAD',
                       git_label: 'ChangeDocument\\s?:',
                       git_format: '%b']

        def changeID = new ChangeManagement(nullScript, gitUtilsMock(true, changeIds)).getChangeDocumentId(params)

        assertThat(logging.log, containsString('[INFO] ChangeDocumentId \'a\' retrieved from git commit(s). '))
        assert changeID == 'a'
    }

    @Test
    public void testGetTransportRequestIdFromConfigWhenProvidedViaConfig() {

        def transportRequestCachedIdInsideCPE
        def cpe = [setTransportRequestId: { tID -> transportRequestCachedIdInsideCPE = tID},
                   getTransportRequestId: {return transportRequestCachedIdInsideCPE}]

        String[] viaGitUtils = ['0815']
        def transportRequestId = new ChangeManagement(nullScript, gitUtilsMock(true, viaGitUtils))
            .getTransportRequestId(cpe, '0042', 'TransportRequest\\s?:', 'origin/master', 'HEAD', '%b')

        // side effect check ...
        assertThat('TransportRequestId is not cached in cpe as expected.',
            transportRequestCachedIdInsideCPE, is (equalTo('0042')))

        assertThat(transportRequestId, is(equalTo('0042')))
    }

    @Test
    public void testGetTransportRequestIdFromConfigWhenNotViaConfig() {

        def transportRequestCachedIdInsideCPE
        def cpe = [setTransportRequestId: { tID -> transportRequestCachedIdInsideCPE = tID},
                   getTransportRequestId: {return transportRequestCachedIdInsideCPE}]

        def transportRequestId = new ChangeManagement(nullScript, gitUtilsMock(true, ['0815'] as String[]))
            .getTransportRequestId(cpe, null, 'TransportRequest\\s?:', 'origin/master', 'HEAD', '%b')

        // side effect check ...
        assertThat('TransportRequestId is not cached in cpe as expected.',
            transportRequestCachedIdInsideCPE, is (equalTo('0815')))

        assertThat(transportRequestId, is(equalTo('0815')))
    }

    @Test
    public void testGetTransportRequestReturnsNullInCaseNoValueIsProvided() {

        // in order to ensure no exception is thrown in this case.

        def transportRequestCachedIdInsideCPE
        def cpe = [setTransportRequestId: { tID -> transportRequestCachedIdInsideCPE = tID},
                   getTransportRequestId: {return transportRequestCachedIdInsideCPE}]

        def transportRequestId = new ChangeManagement(nullScript, gitUtilsMock(true, [] as String[]))
            .getTransportRequestId(cpe, null, 'TransportRequest\\s?:', 'origin/master', 'HEAD', '%b')

        // side effect check ...
        assertThat('TransportRequestId found in cache, but there should be nothing.',
            transportRequestCachedIdInsideCPE, is (nullValue()))

        // side effect check
        assertThat('Required log entry not found',
            logging.log, containsString('[WARN] Cannot retrieve transport request id from commit history.'))

        assertThat(transportRequestId, is(nullValue()))
    }

    @Test
    public void testIsChangeInDevelopmentReturnsTrueWhenChangeIsInDevelopent() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, "cmclient.*is-change-in-development -cID '001'", 0)

        boolean inDevelopment = new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'user', 'password')

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
                                                           'user',
                                                           'password')

        assertThat(inDevelopment, is(equalTo(false)))
    }

    @Test
    public void testIsChangeInDevelopmentThrowsExceptionWhenCMClientReturnsUnexpectedExitCode() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve status for change document \'001\'. Does this change exist? Return code from cmclient: 1.')

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, "cmclient.*is-change-in-development -cID '001'", 1)

        new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'user', 'password')
    }

    @Test
    public void testGetCommandLineWithoutCMClientOpts() {
        String commandLine = new ChangeManagement(nullScript, null)
            .getCMCommandLine('https://example.org/cm',
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
        .getCMCommandLine('https://example.org/cm',
                          "me",
                          "topSecret",
                          "the-command",
                          ["-key1", "val1", "-key2", "val2"],
                          '-Djavax.net.debug=all')
    commandLine = commandLine.replaceAll(' +', " ")
    assertThat(commandLine, containsString('export CMCLIENT_OPTS="-Djavax.net.debug=all"'))
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
