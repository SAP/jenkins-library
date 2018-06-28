package com.sap.piper.cm

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers
import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.GitUtils

import util.BasePiperTest
import util.JenkinsShellCallRule
import util.Rules

public class ChangeManagementTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()

    private JenkinsShellCallRule script = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
        .around(thrown)
        .around(script)

    @Test
    public void testRetrieveChangeDocumentIdOutsideGitWorkTreeTest() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve change document id. ' +
                             'Not in a git work tree. ' +
                             'Change document id is extracted from git commit messages.')

        new ChangeManagement(nullScript, gitUtilsMock(false, new String[0])).getChangeDocumentId()
    }

    @Test
    public void testRetrieveChangeDocumentIdNothingFound() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve changeId from git commits.')

        new ChangeManagement(nullScript, gitUtilsMock(true, new String[0])).getChangeDocumentId()
    }

    @Test
    public void testRetrieveChangeDocumentNotUnique() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Multiple ChangeIds found')

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
        def changeID = new ChangeManagement(nullScript, gitUtilsMock(true, changeIds)).getChangeDocumentId()

        assert changeID == 'a'
    }

    @Test
    public void testIsChangeInDevelopmentReturnsTrueWhenChangeIsInDevelopent() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, "cmclient.*is-change-in-development -cID '001'", 0)

        boolean inDevelopment = new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'user', 'password')

        assertThat(inDevelopment, is(equalTo(true)))
        assertThat(script.shell[0], allOf(containsString("cmclient"),
                                            containsString("-u 'user'"),
                                            containsString("-p 'password'"),
                                            containsString("-e \"endpoint\""),
                                            containsString('is-change-in-development'),
                                            containsString("-cID '001'"),
                                            containsString("-t SOLMAN")))
    }

    @Test
    public void testIsChangeInDevelopmentReturnsFalseWhenChangeIsNotInDevelopent() {

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, "cmclient.*is-change-in-development -cID '001'", 3)

        boolean inDevelopment = new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'user', 'password')

        assertThat(inDevelopment, is(equalTo(false)))
    }

    @Test
    public void testIsChangeInDevelopmentThrowsExceptionWhenCMClientReturnsUnexpectedExitCode() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve change status. Return code from cmclient: 1')

        script.setReturnValue(JenkinsShellCallRule.Type.REGEX, "cmclient.*is-change-in-development -cID '001'", 1)

        new ChangeManagement(nullScript, null).isChangeInDevelopment('001', 'endpoint', 'user', 'password')
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
