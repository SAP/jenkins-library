package com.sap.piper

import org.jenkinsci.plugins.workflow.steps.MissingContextVariableException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class JenkinsUtilsTest extends BasePiperTest {
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(shellRule)
        .around(jlr)

    @Test
    void testNodeAvailable() {
        def result = jenkinsUtils.nodeAvailable()
        assertThat(shellRule.shell, contains("echo 'Node is available!'"))
        assertThat(result, is(true))
    }

    @Test
    void testNoNodeAvailable() {
        helper.registerAllowedMethod('sh', [String.class], {s ->
            throw new MissingContextVariableException(String.class)
        })

        def result = jenkinsUtils.nodeAvailable()
        assertThat(jlr.log, containsString('No node context available.'))
        assertThat(result, is(false))
    }

}
