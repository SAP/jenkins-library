package com.sap.piper.cm

import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.GitUtils

import util.BasePiperTest
import util.Rules

public class ChangeManagementTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules.getCommonRules(this).around(thrown)

    @Test
    public void testRetrieveChangeDocumentIdOutsideGitWorkTreeTest() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve change document id. ' +
                             'Not in a git work tree. ' +
                             'Change document id is extracted from git commit messages.')

        GitUtils gitUtilsMock = new GitUtils() {
            public boolean insideWorkTree() {
                return false
            }
        }

        new ChangeManagement(nullScript, gitUtilsMock).getChangeDocumentId()
    }

    @Test
    public void testRetrieveChangeDocumentIdNothingFound() {

        thrown.expect(ChangeManagementException)
        thrown.expectMessage('Cannot retrieve changeId from git commits.')

        GitUtils gitUtilsMock = new GitUtils() {
            public boolean insideWorkTree() {
                return true
            }

            public String[] extractLogLines(
                         String filter = '',
                         String from = 'origin/master',
                         String to = 'HEAD',
                         String format = '%b') {
                return new String[0]
            }
        }

        new ChangeManagement(nullScript, gitUtilsMock).getChangeDocumentId()
    }
}
