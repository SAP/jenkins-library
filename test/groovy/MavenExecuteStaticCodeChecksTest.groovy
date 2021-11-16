import com.sap.piper.ReportAggregator
import org.hamcrest.Matcher
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsFileExistsRule
import util.JenkinsReadJsonRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.startsWith
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class MavenExecuteStaticCodeChecksTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])

    private List withEnvArgs = []
    private boolean spotBugsStepCalled = false
    private boolean pmdParserStepCalled = false

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(exception)
        .around(new JenkinsReadYamlRule(this))
        .around(credentialsRule)
        .around(readJsonRule)
        .around(shellCallRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(fileExistsRule)

    @Before
    void init() {
        helper.registerAllowedMethod("withEnv", [List.class, Closure.class], { arguments, closure ->
            arguments.each { arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
        helper.registerAllowedMethod("dockerExecute", [Map.class, Closure.class], {
            Map params, Closure c ->
                c.call()
        })
        helper.registerAllowedMethod("recordIssues", [Map.class], { Map config
            ->
        })
        helper.registerAllowedMethod("spotBugs", [Map.class], { Map config
            -> spotBugsStepCalled = true
        })
        helper.registerAllowedMethod("pmdParser", [Map.class], { Map config
            -> pmdParserStepCalled = true
        })

        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/mavenExecuteStaticCodeChecks.yaml\'', '{"dockerImage": "maven:3.6-jdk-8"}')

        helper.registerAllowedMethod('findFiles', [Map.class], {return null})
        helper.registerAllowedMethod("writePipelineEnv", [Map.class], {m -> return })
        helper.registerAllowedMethod("readPipelineEnv", [Map.class], {m -> return })
    }

    @Test
    void 'MavenExecuteStaticCodeChecks should be executed, results recorded and reported in Reportaggregator'() {
        spotBugsStepCalled = false
        pmdParserStepCalled = false

        nullScript.commonPipelineEnvironment.configuration = [steps: [:]]
        stepRule.step.mavenExecuteStaticCodeChecks(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript
        )

        assertThat(writeFileRule.files['.pipeline/tmp/metadata/mavenExecuteStaticCodeChecks.yaml'], containsString('name: mavenExecuteStaticCodeChecks'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[shellCallRule.shell.size() -1], is('./piper mavenExecuteStaticCodeChecks'))
        assertTrue(spotBugsStepCalled)
        assertTrue(pmdParserStepCalled)
        assertThat(ReportAggregator.instance.staticCodeScans, hasItems("Findbugs Static Code Checks", "PMD Static Code Checks"))
    }

    @Test
    void 'MavenExecuteStaticCodeChecks should not record results and not report in Reportaggregator when turned off'() {
        spotBugsStepCalled = false
        pmdParserStepCalled = false

        nullScript.commonPipelineEnvironment.configuration = [steps: [mavenExecuteStaticCodeChecks: [spotBugs: false, pmd: false]]]
        stepRule.step.mavenExecuteStaticCodeChecks(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript
        )

        assertThat(writeFileRule.files['.pipeline/tmp/metadata/mavenExecuteStaticCodeChecks.yaml'], containsString('name: mavenExecuteStaticCodeChecks'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[shellCallRule.shell.size() -1], is('./piper mavenExecuteStaticCodeChecks'))
        assertFalse(spotBugsStepCalled)
        assertFalse(pmdParserStepCalled)
        assertTrue(ReportAggregator.instance.staticCodeScans.isEmpty())
    }
}
