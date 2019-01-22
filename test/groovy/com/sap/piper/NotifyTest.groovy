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
import util.JenkinsShellCallRule
import util.Rules

class NotifyTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    private Map config

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(thrown)

    @Before
    void init() throws Exception {
        // prepare
        config = [collectTelemetryData: true]
        nullScript.STEP_NAME = 'anyStep'
        utils.env.JOB_NAME = 'testJob'
        Notify.instance = utils
    }

    @Test
    void testWarning() {
        // execute test
        Notify.warning(config, nullScript, "test message")
        // asserts
        assertThat(jlr.log, containsString('[WARNING] test message (piper-lib-os/anyStep)'))
        assertThat(jscr.shell, hasItem(containsString('curl -G -v "https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log"')))
        assertThat(jscr.shell, hasItem(containsString('--data-urlencode "custom14=WARNING"')))
    }

    @Test
    void testError() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[ERROR] test message (piper-lib-os/anyOtherStep)')
        // execute test
        try{
            Notify.error(config, nullScript, "test message", "anyOtherStep")
        }finally{
            // asserts
            assertThat(jscr.shell, hasItem(containsString('curl -G -v "https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log"')))
            assertThat(jscr.shell, hasItem(containsString('--data-urlencode "custom14=ERROR"')))
        }
    }
}
