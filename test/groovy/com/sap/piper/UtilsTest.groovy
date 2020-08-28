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
}
