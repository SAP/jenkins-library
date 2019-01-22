package com.sap.piper

import hudson.AbortException
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.notNullValue
import static org.hamcrest.Matchers.startsWith

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertNull
import static org.junit.Assert.assertThat

import org.springframework.beans.factory.annotation.Autowired

class GitUtilsTest extends BasePiperTest {

    @Autowired
    GitUtils gitUtils

    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(loggingRule)
        .around(shellRule)
        .around(thrown)

    @Before
    void init() throws Exception {
        shellRule.setReturnValue('git rev-parse HEAD', 'testCommitId')
    }

    @Test
    void TestIsInsideWorkTree() {
        shellRule.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 0)
        assertTrue(gitUtils.insideWorkTree())
    }

    @Test
    void TestIsNotInsideWorkTree() {
        shellRule.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 1)
        assertFalse(gitUtils.insideWorkTree())
    }

    @Test
    void testWorkTreeDirty() {
        shellRule.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 0)
        shellRule.setReturnValue('git diff --quiet HEAD', 0)
        assertFalse(gitUtils.isWorkTreeDirty())
    }

    @Test
    void testWorkTreeDirtyOutsideWorktree() {
        thrown.expect(AbortException)
        thrown.expectMessage('Method \'isWorkTreeClean\' called outside a git work tree.')
        shellRule.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 1)
        gitUtils.isWorkTreeDirty()
    }

    @Test
    void testWorkTreeNotDirty() {
        shellRule.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 0)
        shellRule.setReturnValue('git diff --quiet HEAD', 1)
        assertTrue(gitUtils.isWorkTreeDirty())
    }

    @Test
    void testWorkTreeDirtyGeneralGitTrouble() {
        thrown.expect(AbortException)
        thrown.expectMessage('git command \'git diff --quiet HEAD\' return with code \'129\'. This indicates general trouble with git.')
        shellRule.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 0)
        shellRule.setReturnValue('git diff --quiet HEAD', 129) // e.g. when called outside work tree
        gitUtils.isWorkTreeDirty()
    }

    @Test
    void testGetGitCommitId() {
        shellRule.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 0)
        assertEquals('testCommitId', gitUtils.getGitCommitIdOrNull())
    }

    @Test
    void testGetGitCommitIdNotAGitRepo() {
        shellRule.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 128)
        assertNull(gitUtils.getGitCommitIdOrNull())
    }

    @Test
    void testExtractLogLinesWithDefaults() {
        gitUtils.extractLogLines()
        assertTrue(shellRule.shell
                         .stream()
                           .anyMatch( { it ->
                             it.contains('git log --pretty=format:%b origin/master..HEAD')}))
    }

    @Test
    void testExtractLogLinesWithCustomValues() {
        gitUtils.extractLogLines('myFilter', 'HEAD~5', 'HEAD~1', '%B')
        assertTrue( shellRule.shell
                          .stream()
                            .anyMatch( { it ->
                               it.contains('git log --pretty=format:%B HEAD~5..HEAD~1')}))
    }

    @Test
    void testExtractLogLinesFilter() {
        shellRule.setReturnValue('#!/bin/bash git log --pretty=format:%b origin/master..HEAD', 'abc\n123')
        String[] log = gitUtils.extractLogLines('12.*')
        assertThat(log, is(notNullValue()))
        assertThat(log.size(),is(equalTo(1)))
        assertThat(log[0], is(equalTo('123')))
    }

    @Test
    void testExtractLogLinesFilterNoMatch() {
        shellRule.setReturnValue('#!/bin/bash git log --pretty=format:%b origin/master..HEAD', 'abc\n123')
        String[] log = gitUtils.extractLogLines('xyz')
        assertNotNull(log)
        assertThat(log.size(),is(equalTo(0)))
    }

    @Test
    void testHandleTestRepository() {
        def result, gitMap, stashName, config = [
            testRepository: 'repoUrl',
            gitSshKeyCredentialsId: 'abc',
            gitBranch: 'master'
        ]

        helper.registerAllowedMethod('git', [Map.class], {m -> gitMap = m })
        helper.registerAllowedMethod("stash", [String.class], { s -> stashName = s})

        result = GitUtils.handleTestRepository(nullScript, config)
        // asserts
        assertThat(gitMap, hasEntry('url', config.testRepository))
        assertThat(gitMap, hasEntry('credentialsId', config.gitSshKeyCredentialsId))
        assertThat(gitMap, hasEntry('branch', config.gitBranch))
        assertThat(stashName, startsWith('testContent-'))
        assertThat(result, startsWith('testContent-'))
    }
}
