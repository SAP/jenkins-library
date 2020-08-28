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

    @Test
    void testGetStageNameFromEnv() {
        File stageScript = new File('vars', 'piperPipelineStageMavenStaticCodeChecks.groovy')
        Script step = loadScript(stageScript.getPath())

        nullScript.env = ['STAGE_NAME': 'stageNameFromEnv']
        nullScript.commonPipelineEnvironment.useTechnicalStageNames = false

        String result = Utils.getStageName(nullScript, [:], step)
        assertEquals('stageNameFromEnv', result)
    }

    @Test
    void testGetStageNameFromField() {
        File stageScript = new File('vars', 'piperPipelineStageMavenStaticCodeChecks.groovy')
        Script step = loadScript(stageScript.getPath())

        nullScript.env = ['STAGE_NAME': 'stageNameFromEnv']
        nullScript.commonPipelineEnvironment.useTechnicalStageNames = true

        String result = Utils.getStageName(nullScript, [:], step)
        assertEquals('mavenExecuteStaticCodeChecks', result)
    }

    @Test
    void testGetStageNameFromParametersTechnical() {
        File stageScript = new File('vars', 'piperPipelineStageMavenStaticCodeChecks.groovy')
        Script step = loadScript(stageScript.getPath())

        nullScript.env = ['STAGE_NAME': 'stageNameFromEnv']
        nullScript.commonPipelineEnvironment.useTechnicalStageNames = true

        Map parameters = ['stageName': 'stageNameFromParams']

        String result = Utils.getStageName(nullScript, parameters, step)
        assertEquals('stageNameFromParams', result)
    }

    @Test
    void testGetStageNameFromParametersEnv() {
        File stageScript = new File('vars', 'piperPipelineStageMavenStaticCodeChecks.groovy')
        Script step = loadScript(stageScript.getPath())

        nullScript.env = ['STAGE_NAME': 'stageNameFromEnv']
        nullScript.commonPipelineEnvironment.useTechnicalStageNames = false

        Map parameters = ['stageName': 'stageNameFromParams']

        String result = Utils.getStageName(nullScript, parameters, step)
        assertEquals('stageNameFromParams', result)
    }
}
