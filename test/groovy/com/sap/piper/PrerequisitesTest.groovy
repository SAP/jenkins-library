package com.sap.piper

import util.JenkinsLoggingRule

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import hudson.AbortException
import util.BasePiperTest
import util.Rules

class PrerequisitesTest extends BasePiperTest {

    def result = 'SUCCESS'

    @Rule
    public ExpectedException thrown = ExpectedException.none()

    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
                                      .around(loggingRule)

    @Before
    public void init() {
        nullScript.currentBuild = [
            'setResult' : { r -> result = r },
            STEP_NAME: 'dummy',
        ]
        nullScript.STEP_NAME = 'dummy'
    }

    @Test
    public void checkScriptProvidedTest() {

        def script = Prerequisites.checkScript(nullScript, [script:{}])

        assert script != null
        assert result == 'SUCCESS'

    }

    @Test
    public void checkScriptMissingTest() {

        loggingRule.expect('No reference to surrounding script provided with key \'script\'')

        def script = Prerequisites.checkScript(nullScript, [:])

        assert script == null
        assert result == 'UNSTABLE'
    }

    @Test
    public void checkScriptMissingTestFeatureFlagSet() {

        thrown.expect(AbortException)
        thrown.expectMessage('No reference to surrounding script provided')

        try {
            System.setProperty('com.sap.piper.featureFlag.failOnMissingScript', 'true')
            Prerequisites.checkScript(nullScript, [:])
        } finally {
            System.clearProperty('com.sap.piper.featureFlag.failOnMissingScript')
        }
    }
}
