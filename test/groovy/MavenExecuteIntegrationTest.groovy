import groovy.json.JsonSlurper
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class MavenExecuteIntegrationTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()

    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)

    private List withEnvArgs = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(exception)
        .around(new JenkinsReadYamlRule(this))
        .around(credentialsRule)
        .around(new JenkinsReadJsonRule(this))
        .around(shellCallRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(new JenkinsFileExistsRule(this, []))

    @Before
    void init() {
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if (m.text != null)
                return new JsonSlurper().parseText(m.text)
        })
        helper.registerAllowedMethod("withEnv", [List, Closure], { arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
        shellCallRule.setReturnValue(
            './piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/mavenExecuteIntegration.yaml\'',
            '{"verbose": false}'
        )

        helper.registerAllowedMethod('fileExists', [String], {return true})
        helper.registerAllowedMethod('findFiles', [Map], {return null})
        helper.registerAllowedMethod('testsPublishResults', [Map], {return null})
    }

    @Test
    void testWithSidecar() {
        stepRule.step.mavenExecuteIntegration(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: 'This is test content',
            sidecarImage: 'some/image',
            script: nullScript,
        )
        // asserts
        assertThat(writeFileRule.files['.pipeline/tmp/metadata/mavenExecuteIntegration.yaml'], containsString('name: mavenExecuteIntegration'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'),
            containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[1], is('./piper mavenExecuteIntegration'))
    }
}
