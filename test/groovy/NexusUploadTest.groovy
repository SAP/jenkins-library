import groovy.json.JsonSlurper
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class NexusUploadTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()

    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])
    //private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this, [])

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
    //    .around(dockerExecuteRule)

    @Before
    void init() {
        helper.registerAllowedMethod('fileExists', [Map], {
            return true
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if (m.file == 'nexusUpload_reports.json')
                return [[target: "1234.pdf", mandatory: true]]
            if (m.file == 'nexusUpload_links.json')
                return []
            if (m.text != null)
                return new JsonSlurper().parseText(m.text)
        })
        helper.registerAllowedMethod("withEnv", [List, Closure], { arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
//        helper.registerAllowedMethod("dockerExecute", [Map, Closure], { map, closure ->
//            // ignore
//        })
        credentialsRule.withCredentials('idOfCxCredential', "admin", "admin123")
        shellCallRule.setReturnValue(
            './piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/nexusUpload.yaml\'',
            '{"credentialsId": "idOfCxCredential", "verbose": false}'
        )
    }

    @Test
    void testDeployPom() {
        stepRule.step.nexusUpload(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript,
        )
        // asserts
        assertThat(writeFileRule.files['.pipeline/tmp/metadata/nexusUpload.yaml'], containsString('name: nexusUpload'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'),
            containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[1], is('./piper nexusUpload'))
    }
}
