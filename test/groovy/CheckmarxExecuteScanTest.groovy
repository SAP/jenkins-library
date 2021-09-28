import groovy.json.JsonSlurper
import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class CheckmarxExecuteScanTest extends BasePiperTest {
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
            return true
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'checkmarxExecuteScan_reports.json')
                return [[target: "1234.pdf", mandatory: true]]
            if(m.file == 'checkmarxExecuteScan_links.json')
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
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/checkmarx.yaml\'', '{"checkmarxCredentialsId": "idOfCxCredential", "verbose": false}')

        helper.registerAllowedMethod('findFiles', [Map.class], {return null})
        helper.registerAllowedMethod("writePipelineEnv", [Map.class], {m -> return })
        helper.registerAllowedMethod("readPipelineEnv", [Map.class], {m -> return })
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
        assertThat(writeFileRule.files['.pipeline/tmp/metadata/checkmarx.yaml'], containsString('name: checkmarxExecuteScan'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[1], is('./piper checkmarxExecuteScan'))
    }

    @Test
    void testCheckmarxExecuteScanNoReports() {
        helper.registerAllowedMethod('fileExists', [Map], {
            return false
        })

        exception.expect(AbortException)
        exception.expectMessage("Expected to find checkmarxExecuteScan_reports.json in workspace but it is not there")

        stepRule.step.checkmarxExecuteScan(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript
        )
    }
}
