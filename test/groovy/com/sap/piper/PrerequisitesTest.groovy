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

    @Rule
    public ExpectedException thrown = ExpectedException.none()

    public JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
                                      .around(jlr)

    @Before
    public void init() {
        nullScript.metaClass.STEP_NAME = 'dummy'
        nullScript.currentBuild.status = 'SUCCESS'
    }

    @Test
    public void checkScriptProvidedTest() {

        def script = Prerequisites.checkScript(nullScript, [script:{}])

        assert script != null
        assert nullScript.currentBuild.status == 'SUCCESS'

    }

    @Test
    public void checkScriptMissingTest() {

        jlr.expect('No reference to surrounding script provided with key \'script\'')
        assert nullScript.currentBuild.status == 'SUCCESS'

        def script = Prerequisites.checkScript(nullScript, [:])

        assert script == null
        assert nullScript.currentBuild.status == 'UNSTABLE'
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
