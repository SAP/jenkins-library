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
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.startsWith
import static org.junit.Assert.assertThat

class MavenExecuteStaticCodeChecksTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])

    private List withEnvArgs = []

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
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/metadata/mavenStaticCodeChecks.yaml\'', '{"dockerImage": "maven:3.6-jdk-8"}')
    }

    @Test
    void testMavenExecuteStaticCodeChecksDefault() {
        stepRule.step.mavenExecuteStaticCodeChecks(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript
        )
        // asserts
        assertThat(writeFileRule.files['.pipeline/metadata/mavenStaticCodeChecks.yaml'], containsString('name: mavenExecuteStaticCodeChecks'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[1], is('./piper mavenExecuteStaticCodeChecks'))
    }
}
