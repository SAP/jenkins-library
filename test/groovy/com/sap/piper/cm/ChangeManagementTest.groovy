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
