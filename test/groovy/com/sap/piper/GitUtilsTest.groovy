package com.sap.piper

import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasItem
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

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(thrown)

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

    @Test
    void testExtractLogLinesWithDefaults() {
        gitUtils.extractLogLines()
        assertTrue(jscr.shell
                         .stream()
                           .anyMatch( { it ->
                             it.contains('git log --pretty=format:%b origin/master..HEAD')}))
    }

    @Test
    void testExtractLogLinesWithCustomValues() {
        gitUtils.extractLogLines('myFilter', 'HEAD~5', 'HEAD~1', '%B')
        assertTrue( jscr.shell
                          .stream()
                            .anyMatch( { it ->
                               it.contains('git log --pretty=format:%B HEAD~5..HEAD~1')}))
    }

    @Test
    void testExtractLogLinesFilter() {
        jscr.setReturnValue('#!/bin/bash git log --pretty=format:%b origin/master..HEAD', 'abc\n123')
        String[] log = gitUtils.extractLogLines('12.*')
        assertThat(log, is(notNullValue()))
        assertThat(log.size(),is(equalTo(1)))
        assertThat(log[0], is(equalTo('123')))
    }

    @Test
    void testExtractLogLinesFilterNoMatch() {
        jscr.setReturnValue('#!/bin/bash git log --pretty=format:%b origin/master..HEAD', 'abc\n123')
        String[] log = gitUtils.extractLogLines('xyz')
        assertNotNull(log)
        assertThat(log.size(),is(equalTo(0)))
	}

    @Test
    void testHandleTestRepository() {
        helper.registerAllowedMethod("git", [Map.class], { m ->
            assertThat(m, hasEntry('url', 'repoUrl'))
            assertThat(m, hasEntry('credentialsId', 'abc'))
            assertThat(m, hasEntry('branch', 'master'))
        })
        helper.registerAllowedMethod("stash", [String.class], { s ->
            assertThat(s, startsWith('testContent-'))
        })

        def config = [
            testRepository: 'repoUrl',
            gitSshKeyCredentialsId: 'abc',
            gitBranch: 'master'
        ]

        GitUtils.handleTestRepository(nullScript, config)
        println("LOG: ${jlr.log}")
        // asserts
        assertThat(config.stashContent, hasItem(startsWith('testContent-')))
	}
}
