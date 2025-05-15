import com.sap.piper.JenkinsUtils

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.any
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasKey

import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.Rules
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.JenkinsShellCallRule

import com.sap.piper.Utils

import static com.lesfurets.jenkins.unit.MethodSignature.method

class PiperPublishWarningsTest extends BasePiperTest {
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    def warningsParserSettings
    def groovyScriptParserSettings
    def warningsPluginOptions

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(shellRule)
        .around(stepRule)

    @Before
    void init() throws Exception {
        warningsParserSettings = [:]
        groovyScriptParserSettings = [:]
        warningsPluginOptions = [:]

        // add handler for generic step call
        helper.registerAllowedMethod("writeFile", [Map.class], null)
        helper.registerAllowedMethod("recordIssues", [Map.class], {
            parameters -> warningsPluginOptions = parameters
        })
        helper.registerAllowedMethod("groovyScript", [Map.class], {
            parameters -> groovyScriptParserSettings = parameters
        })
        JenkinsUtils.metaClass.addWarningsNGParser = { String s1, String s2, String s3, String s4 ->
            warningsParserSettings = [id: s1, name: s2, regex: s3, script: s4]
            return true
        }
        JenkinsUtils.metaClass.static.getFullBuildLog = { def currentBuild -> return ""}
        JenkinsUtils.metaClass.static.isPluginActive = { id -> return true}

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        JenkinsUtils.metaClass.addWarningsNGParser = null
        JenkinsUtils.metaClass.static.getFullBuildLog = null
        JenkinsUtils.metaClass.static.isPluginActive = null
        Utils.metaClass = null
    }

    @Test
    void testPublishWarnings() throws Exception {
        stepRule.step.piperPublishWarnings(script: nullScript)
        // asserts
        assertThat(loggingRule.log, containsString('[piperPublishWarnings] Added warnings-ng plugin parser \'Piper\' configuration.'))
        assertThat(warningsParserSettings, hasEntry('id', 'piper'))
        assertThat(warningsParserSettings, hasEntry('name', 'Piper'))
        assertThat(warningsParserSettings, hasEntry('regex', '\\[(INFO|WARNING|ERROR)\\] (.*) \\(([^) ]*)\\/([^) ]*)\\)'))
        assertThat(warningsParserSettings, hasKey('script'))
        assertThat(warningsPluginOptions, allOf(
            hasEntry('enabledForFailure', true),
            hasEntry('skipBlames', true)
        ))
        assertThat(warningsPluginOptions, hasKey('tools'))

        assertThat(groovyScriptParserSettings, hasEntry('parserId', 'piper'))
        assertThat(groovyScriptParserSettings, hasEntry('pattern', 'build.log'))
    }
}
