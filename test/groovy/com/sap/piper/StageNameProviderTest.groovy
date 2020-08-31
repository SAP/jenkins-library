package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.Rules

import static org.junit.Assert.*

class StageNameProviderTest extends BasePiperTest {

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)

    private Script testStep

    @Before
    void init() {
        File stageScript = new File('vars', 'piperPipelineStageMavenStaticCodeChecks.groovy')
        testStep = loadScript(stageScript.getPath())
        nullScript.env = ['STAGE_NAME': 'stageNameFromEnv']
    }

    @Test
    void testGetStageNameFromEnv() {
        StageNameProvider.instance.useTechnicalStageNames = false

        String result = StageNameProvider.instance.getStageName(nullScript, [:], testStep)
        assertEquals('stageNameFromEnv', result)
    }

    @Test
    void testGetStageNameFromField() {
        StageNameProvider.instance.useTechnicalStageNames = true

        String result = StageNameProvider.instance.getStageName(nullScript, [:], testStep)
        assertEquals('mavenExecuteStaticCodeChecks', result)
    }

    @Test
    void testGetStageNameFromParametersTechnical() {
        StageNameProvider.instance.useTechnicalStageNames = true

        Map parameters = ['stageName': 'stageNameFromParams']

        String result = StageNameProvider.instance.getStageName(nullScript, parameters, testStep)
        assertEquals('stageNameFromParams', result)
    }

    @Test
    void testGetStageNameFromParametersEnv() {
        StageNameProvider.instance.useTechnicalStageNames = false

        Map parameters = ['stageName': 'stageNameFromParams']

        String result = StageNameProvider.instance.getStageName(nullScript, parameters, testStep)
        assertEquals('stageNameFromParams', result)
    }
}
