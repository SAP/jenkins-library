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
import static org.junit.Assert.assertThat

class MavenExecuteTest extends BasePiperTest {
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
        helper.registerAllowedMethod('fileExists', [Map], {
            if (m.file == 'mavenExecute_reports.json')
                return false
            if (m.file == 'mavenExecute_links.json')
                return false
            return true
        })
        helper.registerAllowedMethod("withEnv", [List, Closure], { arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
        credentialsRule.withCredentials('idOfCxCredential', "admin", "admin123")
        shellCallRule.setReturnValue(
            './piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/mavenExecute.yaml\'',
            '{"credentialsId": "idOfCxCredential", "verbose": false}'
        )
    }

    @Test
    void testDeployPom() {
        stepRule.step.mavenExecute(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript,
        )
        // asserts
        assertThat(writeFileRule.files['.pipeline/tmp/metadata/mavenExecute.yaml'], containsString('name: mavenExecute'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'),
            containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[1], is('./piper mavenExecute'))
    }
}
