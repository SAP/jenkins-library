package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertNull

class GitUtilsTest extends BasePiperTest {

    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this).around(jscr).around(thrown)

    @Before
    void init() throws Exception {
        jscr.setReturnValue('git rev-parse HEAD', 'testCommitId')
    }

    @Test
    void TestIsInsideWorkTree() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 0)
        assertTrue(gitUtils.insideWorkTree())
    }

    @Test
    void TestIsNotInsideWorkTree() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 1)
        assertFalse(gitUtils.insideWorkTree())
    }


    @Test
    void testGetGitCommitId() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 0)
        assertEquals('testCommitId', gitUtils.getGitCommitIdOrNull())
    }

    @Test
    void testGetGitCommitIdNotAGitRepo() {
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 128)
        assertNull(gitUtils.getGitCommitIdOrNull())
    }

}
