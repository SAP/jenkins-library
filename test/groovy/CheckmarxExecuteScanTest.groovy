import groovy.json.JsonSlurper
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class CheckmarxExecuteScanTest extends BasePiperTest {

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
        .around(new JenkinsReadYamlRule(this))
        .around(credentialsRule)
        .around(readJsonRule)
        .around(shellCallRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(fileExistsRule)

    @Before
    void init() {
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'reports.json')
                return [[target: "1234.pdf", mandatory: true]]
            if(m.file == 'links.json')
                return []
            if(m.text != null)
                return new JsonSlurper().parseText(m.text)
        })
        helper.registerAllowedMethod("withEnv", [List.class, Closure.class], {arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
        credentialsRule.withCredentials('idOfCxCredential', "PIPER_username", "PIPER_password")
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'metadata/checkmarx.yaml\'', '{"checkmarxCredentialsId": "idOfCxCredential", "verbose": false}')
    }

    @Test
    void testCheckmarxExecuteScanDefault() {
        stepRule.step.checkmarxExecuteScan(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript
        )
        // asserts
        assertThat(writeFileRule.files['metadata/checkmarx.yaml'], containsString('name: checkmarxExecuteScan'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[1], is('./piper checkmarxExecuteScan --verbose false'))
    }
}
