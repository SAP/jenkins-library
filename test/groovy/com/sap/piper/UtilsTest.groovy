package com.sap.piper

import org.junit.Rule
import org.junit.Before
import org.junit.Test
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import static org.hamcrest.Matchers.is

import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.BasePiperTest
import util.Rules

class UtilsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(shellRule)
        .around(loggingRule)

    private parameters

    @Before
    void setup() {
        parameters = [:]
    }

    @Test
    void testGenerateSHA1() {
        def result = utils.generateSha1('ContinuousDelivery')
        // asserts
        // generated with "echo -n 'ContinuousDelivery' | sha1sum | sed 's/  -//'"
        assertThat(result, is('0dad6c33b6246702132454f604dee80740f399ad'))
    }

    @Test
    void testUnstashAllSkipNull() {
        def stashResult = utils.unstashAll(['a', null, 'b'])
        assert stashResult == ['a', 'b']
    }

    @Test
    void testAppendNonExistingParameterToStringList() {
        Map parameters = [:]
        List result = Utils.appendParameterToStringList([], parameters, 'non-existing')
        assertTrue(result.isEmpty())
    }

    @Test
    void testAppendStringParameterToStringList() {
        Map parameters = ['param': 'string']
        List result = Utils.appendParameterToStringList([], parameters, 'param')
        assertEquals(1, result.size())
    }

    @Test
    void testAppendListParameterToStringList() {
        Map parameters = ['param': ['string2', 'string3']]
        List result = Utils.appendParameterToStringList(['string1'], parameters, 'param')
        assertEquals(['string1', 'string2', 'string3'], result)
    }

    @Test
    void testAppendEmptyListParameterToStringList() {
        Map parameters = ['param': []]
        List result = Utils.appendParameterToStringList(['string'], parameters, 'param')
        assertEquals(['string'], result)
    }

    def newExaminee(Map results) {
        results.stashProperties = null
        def examinee = new Utils()
        examinee.steps = [
            stash: { Map stashProperties ->
                results.stashProperties = stashProperties
            },
        ]
        examinee.echo = {}
        return examinee
    }

    @Test
    void testStash_noParentheses() {
        Map results = [:]
        newExaminee(results).stash 'test'
        assertEquals([name: 'test', includes: '**/*.*', excludes: ''], results.stashProperties)
    }

    @Test
    void testStashAndLog_noParentheses() {
        Map results = [:]
        newExaminee(results).stash name: 'test'
        assertEquals([name: 'test', includes: '**/*.*', excludes: ''], results.stashProperties)
    }

    @Test
    void testStash_simpleSignature1Param() {
        Map results = [:]
        Map expected = [name: 'test', includes: '**/*.*', excludes: '']
        
        newExaminee(results).stash('test')
        assertEquals(expected, results.stashProperties)
        
        newExaminee(results).stash(name: 'test')
        assertEquals(expected, results.stashProperties)
    }

    @Test
    void testStash_simpleSignature2Params() {
        Map results = [:]
        Map expected = [name: 'test', includes: 'includesX', excludes: '']
        
        newExaminee(results).stash('test', 'includesX')
        assertEquals(expected, results.stashProperties)
        
        newExaminee(results).stash(name: 'test', includes: 'includesX')
        assertEquals(expected, results.stashProperties)
    }

    @Test
    void testStash_simpleSignature3Params() {
        Map results = [:]
        Map expected = [name: 'test', includes: 'includesX', excludes: 'excludesX']
        
        newExaminee(results).stash('test', 'includesX', 'excludesX')
        assertEquals(expected, results.stashProperties)
        
        newExaminee(results).stash(name: 'test', includes: 'includesX', excludes: 'excludesX')
        assertEquals(expected, results.stashProperties)
    }

    @Test
    void testStash_simpleSignature4Params() {
        Map results = [:]
        Map expected = [name: 'test', includes: 'includesX', excludes: 'excludesX', useDefaultExcludes: false]
        
        newExaminee(results).stash('test', 'includesX', 'excludesX', false)
        assertEquals(expected, results.stashProperties)
        
        newExaminee(results).stash(name: 'test', includes: 'includesX', excludes: 'excludesX', useDefaultExcludes: false)
        assertEquals(expected, results.stashProperties)
    }

    @Test
    void testStash_simpleSignature5Params() {
        Map results = [:]
        Map expected = [name: 'test', includes: 'includesX', excludes: 'excludesX', useDefaultExcludes: false, allowEmpty: true]
        
        newExaminee(results).stash('test', 'includesX', 'excludesX', false, true)
        assertEquals(expected, results.stashProperties)
        
        newExaminee(results).stash(name: 'test', includes: 'includesX', excludes: 'excludesX', useDefaultExcludes: false, allowEmpty: true)
        assertEquals(expected, results.stashProperties)
    }

    @Test
    void testStash_explicitDefaults() {
        Map results = [:]
        Map expected = [name: 'test', includes: 'includesX', excludes: 'excludesX']
       
        newExaminee(results).stash('test', 'includesX', 'excludesX', true, false)
        assertEquals(expected, results.stashProperties)
        
        newExaminee(results).stash(name: 'test', includes: 'includesX', excludes: 'excludesX', useDefaultExcludes: true, allowEmpty: false)
        assertEquals(expected, results.stashProperties)
    }

    @Test(expected = IllegalArgumentException.class)
    void testStashAndLog_noName_fails() {
        Map results = [:]
        newExaminee(results).stash([:])
        assertEquals([includes: 'includesX'], results.stashProperties)
    }

    @Test
    void testStashAndLog_includes() {
        Map results = [:]
        newExaminee(results).stash(name: 'test', includes: 'includesX')
        assertEquals([name: 'test', includes: 'includesX', excludes: ''], results.stashProperties)
    }

    @Test
    void testStashAndLog_excludes() {
        Map results = [:]
        newExaminee(results).stash(name: 'test', excludes: 'excludesX')
        assertEquals([name: 'test', includes: '**/*.*', excludes: 'excludesX'], results.stashProperties)
    }

    @Test
    void testStashAndLog_useDefaultExcludes() {
        Map results = [:]
        newExaminee(results).stash(name: 'test', useDefaultExcludes: true)
        assertEquals([name: 'test', includes: '**/*.*', excludes: ''], results.stashProperties)
        newExaminee(results).stash(name: 'test', useDefaultExcludes: false)
        assertEquals([name: 'test', includes: '**/*.*', excludes: '', useDefaultExcludes: false], results.stashProperties)
    }

    @Test
    void testStashAndLog_allowEmpty() {
        Map results = [:]
        newExaminee(results).stash(name: 'test', allowEmpty: true)
        assertEquals([name: 'test', includes: '**/*.*', excludes: '', allowEmpty: true], results.stashProperties)
        newExaminee(results).stash(name: 'test', allowEmpty: false)
        assertEquals([name: 'test', includes: '**/*.*', excludes: ''], results.stashProperties)
    }

}
