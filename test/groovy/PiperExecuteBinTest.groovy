import com.sap.piper.JenkinsUtils
import groovy.json.JsonSlurper
import hudson.AbortException
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class PiperExecuteBinTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()

    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)

    private List withEnvArgs = []
    private List credentials = []
    private List artifacts = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(exception)
        .around(new JenkinsReadYamlRule(this))
        .around(credentialsRule)
        .around(shellCallRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(fileExistsRule)
        .around(dockerExecuteRule)

    @Before
    void init() {
        credentials = []

        helper.registerAllowedMethod("withEnv", [List.class, Closure.class], {arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })

        helper.registerAllowedMethod('fileExists', [Map.class], {
            return true
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'testStep_reports.json')
                return [[target: "1234.pdf", mandatory: true]]
            if(m.file == 'testStep_links.json')
                return []
            if(m.text != null)
                return new JsonSlurper().parseText(m.text)
        })

        helper.registerAllowedMethod('libraryResource', [String.class], {s ->
            if (s == 'metadata/test.yaml') {
                return '''metadata:
  name: testStep
'''
            } else {
                return '''general:
  failOnError: true
'''
            }
        })

        helper.registerAllowedMethod('file', [Map], { m -> return m })
        helper.registerAllowedMethod('string', [Map], { m -> return m })
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            l.each {m ->
                credentials.add(m)
                if (m.credentialsId == 'credFile') {
                    binding.setProperty('PIPER_credFile', 'credFileContent')
                } else if (m.credentialsId == 'credToken') {
                    binding.setProperty('PIPER_credToken','credTokenContent')
                } else if (m.credentialsId == 'credUsernamePassword') {
                    binding.setProperty('PIPER_user', 'userId')
                    binding.setProperty('PIPER_password', '********')
                }
            }
            try {
                c()
            } finally {
                binding.setProperty('PIPER_credFile', null)
                binding.setProperty('PIPER_credToken', null)
                binding.setProperty('PIPER_user', null)
                binding.setProperty('PIPER_password', null)
            }
        })

        helper.registerAllowedMethod('archiveArtifacts', [Map.class], {m ->
            artifacts.add(m)
            return null
        })
    }

    @Test
    void testPiperExecuteBinDefault() {
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/test.yaml\'', '{"fileCredentialsId":"credFile", "tokenCredentialsId":"credToken", "credentialsId":"credUsernamePassword", "dockerImage":"my.Registry/my/image:latest"}')

        List stepCredentials = [
            [type: 'file', id: 'fileCredentialsId', env: ['PIPER_credFile']],
            [type: 'token', id: 'tokenCredentialsId', env: ['PIPER_credToken']],
            [type: 'usernamePassword', id: 'credentialsId', env: ['PIPER_user', 'PIPER_password']],
        ]
        stepRule.step.piperExecuteBin(
            [
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                testParam: "This is test content",
                script: nullScript
            ],
            'testStep',
            'metadata/test.yaml',
            stepCredentials
        )
        // asserts
        assertThat(writeFileRule.files['.pipeline/tmp/metadata/test.yaml'], containsString('name: testStep'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[1], is('./piper testStep'))
        assertThat(credentials.size(), is(3))
        assertThat(credentials[0], allOf(hasEntry('credentialsId', 'credFile'), hasEntry('variable', 'PIPER_credFile')))
        assertThat(credentials[1], allOf(hasEntry('credentialsId', 'credToken'), hasEntry('variable', 'PIPER_credToken')))
        assertThat(credentials[2], allOf(hasEntry('credentialsId', 'credUsernamePassword'), hasEntry('usernameVariable', 'PIPER_user') , hasEntry('passwordVariable', 'PIPER_password')))

        assertThat(dockerExecuteRule.dockerParams.dockerImage, is('my.Registry/my/image:latest'))

        assertThat(artifacts[0], allOf(hasEntry('artifacts', '1234.pdf'), hasEntry('allowEmptyArchive', false)))
    }

    @Test
    void testPiperExecuteBinSomeCredentials() {
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/test.yaml\'', '{"fileCredentialsId":"credFile", "tokenCredentialsId":"credToken", "dockerImage":"my.Registry/my/image:latest"}')

        List stepCredentials = [
            [type: 'file', id: 'fileCredentialsId', env: ['PIPER_credFile']],
            [type: 'token', id: 'tokenCredentialsId', env: ['PIPER_credToken']],
            [type: 'usernamePassword', id: 'credentialsId', env: ['PIPER_user', 'PIPER_password']],
        ]
        stepRule.step.piperExecuteBin(
            [
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                testParam: "This is test content",
                script: nullScript
            ],
            'testStep',
            'metadata/test.yaml',
            stepCredentials
        )
        // asserts
        assertThat(credentials.size(), is(2))
        assertThat(credentials[0], allOf(hasEntry('credentialsId', 'credFile'), hasEntry('variable', 'PIPER_credFile')))
        assertThat(credentials[1], allOf(hasEntry('credentialsId', 'credToken'), hasEntry('variable', 'PIPER_credToken')))
    }

    @Test
    void testPiperExecuteBinNoDockerNoCredentials() {
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/test.yaml\'', '{}')

        stepRule.step.piperExecuteBin(
            [
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                testParam: "This is test content",
                script: nullScript
            ],
            'testStep',
            'metadata/test.yaml',
            []
        )

        assertThat(writeFileRule.files['.pipeline/tmp/metadata/test.yaml'], containsString('name: testStep'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[1], is('./piper testStep'))
        assertThat(credentials.size(), is(0))

        assertThat(dockerExecuteRule.dockerParams.size(), is(0))

        assertThat(artifacts[0], allOf(hasEntry('artifacts', '1234.pdf'), hasEntry('allowEmptyArchive', false)))

    }

    @Test
    void testPiperExecuteBinNoReportFound() {
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/test.yaml\'', '{}')
        helper.registerAllowedMethod('fileExists', [Map], {
            return false
        })

        exception.expect(AbortException)
        exception.expectMessage("Expected to find testStep_reports.json in workspace but it is not there")

        stepRule.step.piperExecuteBin(
            [
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                testParam: "This is test content",
                script: nullScript
            ],
            'testStep',
            'metadata/test.yaml',
            [],
            true,
            false
        )
    }
}
