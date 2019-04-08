#!groovy
package steps

import com.sap.piper.internal.JenkinsUtils

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasKey

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.Rules
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.JenkinsShellCallRule

import static com.lesfurets.jenkins.unit.MethodSignature.method

class PiperPublishNotificationsTest extends BasePiperTest {
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    def warningsParserSettings
    def warningsPluginOptions

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(jsr)

    @Before
    void init() throws Exception {
        warningsParserSettings = [:]
        warningsPluginOptions = [:]
        // add handler for generic step call
        helper.registerAllowedMethod("warnings", [Map.class], {
            parameters -> warningsPluginOptions = parameters
        })
        JenkinsUtils.metaClass.static.addWarningsParser = { Map m ->
            warningsParserSettings = m
            return true
        }
        helper.registerAllowedMethod( "deleteDir", [], null )
    }

    @Test
    void testPublishNotifications() throws Exception {
        jsr.step.sapPiperPublishNotifications(script: nullScript)
        // asserts
        assertThat(jlr.log, containsString('[piperPublishNotifications] New Warnings plugin parser \'Piper Notifications Parser\' configuration added.'))
        assertThat(warningsParserSettings, hasEntry('parserName', 'Piper Notifications Parser'))
        assertThat(warningsParserSettings, hasEntry('parserRegexp', '\\[(INFO|WARNING|ERROR)\\] (.*) \\(([^) ]*)\\/([^) ]*)\\)'))
        assertThat(warningsPluginOptions, hasEntry('canRunOnFailed', true))
        assertThat(warningsPluginOptions, hasKey('consoleParsers'))
        assertThat(warningsPluginOptions.consoleParsers, hasItem(hasEntry('parserName', 'Piper Notifications Parser')))
    }
}
