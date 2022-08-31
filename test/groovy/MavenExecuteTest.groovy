import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsFileExistsRule
import util.JenkinsMavenExecuteRule
import util.JenkinsReadJsonRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class MavenExecuteTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()

    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
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
        .around(new JenkinsReadJsonRule(this))
        .around(shellCallRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(fileExistsRule)

    @Before
    void init() {
        helper.registerAllowedMethod("withEnv", [List, Closure], { arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
        credentialsRule.withCredentials('idOfCxCredential', "admin", "admin123")
        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue(
            './piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/mavenExecute.yaml\'',
            '{"credentialsId": "idOfCxCredential", "verbose": false}'
        )

        helper.registerAllowedMethod('findFiles', [Map.class], {return null})
        helper.registerAllowedMethod("writePipelineEnv", [Map.class], {m -> return })
        helper.registerAllowedMethod("readPipelineEnv", [Map.class], {m -> return })
    }

    @Test
    void testExecute() {
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
        assertThat(shellCallRule.shell[2], is('./piper mavenExecute'))
    }

    @Test
    void testOutputIsReturned() {
        // init
        String outputFile = '.pipeline/maven_output.txt'
        String expectedOutput = 'the output'
        fileExistsRule.registerExistingFile(outputFile)
        helper.registerAllowedMethod('readFile', [String], {file ->
            if (file == outputFile) {
                return expectedOutput
            }
            return ''
        })

        // test
       String receivedOutput = stepRule.step.mavenExecute(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript,
            returnStdout: true,
        )

        // asserts
        assertThat(receivedOutput, is(expectedOutput))
    }

    @Test
    void testOutputIsMissing() {
        // init
        fileExistsRule.setExistingFiles([])
        helper.registerAllowedMethod('readFile', [String], {file ->
            return ''
        })
        String errorMessage = ''
        helper.registerAllowedMethod('error', [String], {message ->
            errorMessage = message
        })

        // test
        stepRule.step.mavenExecute(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript,
            returnStdout: true,
        )

        // asserts
        assertThat(errorMessage, containsString('Internal error. A text file with the contents of the maven output was expected'))
    }
}
