package com.sap.piper

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not

import static org.junit.Assert.assertThat
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.Rules

class NotifyTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(thrown)

    @Before
    void init() throws Exception {
        // prepare
        nullScript.STEP_NAME = 'anyStep'
        utils.env.JOB_NAME = 'testJob'
        Notify.instance = utils
    }

    @Test
    void testWarning() {
        // execute test
        Notify.warning(nullScript, "test message")
        // asserts
        assertThat(jlr.log, containsString('[WARNING] test message (piper-lib-os/anyStep)'))
    }

    @Test
    void testError() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[ERROR] test message (piper-lib-os/anyOtherStep)')
        // execute test
        Notify.error(nullScript, "test message", "anyOtherStep")
    }
}
