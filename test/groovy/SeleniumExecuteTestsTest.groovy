import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class SeleniumExecuteTestsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)

    private List credentials = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(dockerExecuteRule)
        .around(stepRule) // needs to be activated after dockerExecuteRule, otherwise executeDocker is not mocked

    boolean bodyExecuted = false

    def gitMap

    @Before
    void init() throws Exception {
        credentials = []
        bodyExecuted = false
        helper.registerAllowedMethod('stash', [String.class], null)
        helper.registerAllowedMethod('git', [Map.class], {m ->
            gitMap = m
        })
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            l.each {m ->
                credentials.add(m)
                if (m.credentialsId == 'MyCredentialId') {
                    binding.setProperty('PIPER_SELENIUM_HUB_USER', 'seleniumUser')
                    binding.setProperty('PIPER_SELENIUM_HUB_PASSWORD', '********')
                }
            }
            try {
                c()
            } finally {
                binding.setProperty('PIPER_SELENIUM_HUB_USER', null)
                binding.setProperty('PIPER_SELENIUM_HUB_PASSWORD', null)
            }
        })
    }

    @Test
    void testExecuteSeleniumDefault() {
        def expectedDefaultEnvVars = [
            'PIPER_SELENIUM_HOSTNAME': 'npm',
            'PIPER_SELENIUM_WEBDRIVER_HOSTNAME': 'selenium',
            'PIPER_SELENIUM_WEBDRIVER_PORT': '4444'
        ]

        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assertThat(dockerExecuteRule.dockerParams.containerPortMappings, is(['selenium/standalone-chrome': [[containerPort: 4444, hostPort: 4444]]]))
        assertThat(dockerExecuteRule.dockerParams.dockerImage, is('node:lts-bookworm'))
        assertThat(dockerExecuteRule.dockerParams.dockerName, is('npm'))
        assertThat(dockerExecuteRule.dockerParams.dockerWorkspace, is('/home/node'))
        assertThat(dockerExecuteRule.dockerParams.sidecarEnvVars, is(null))
        assertThat(dockerExecuteRule.dockerParams.sidecarImage, is('selenium/standalone-chrome'))
        assertThat(dockerExecuteRule.dockerParams.sidecarName, is('selenium'))
        assertThat(dockerExecuteRule.dockerParams.sidecarVolumeBind, is(['/dev/shm': '/dev/shm']))
        expectedDefaultEnvVars.each { key, value ->
            assert dockerExecuteRule.dockerParams.dockerEnvVars[key] == value
        }
    }

    @Test
    void testNoNullPointerExceptionWithEmptyContainerPortMapping() {
        nullScript.commonPipelineEnvironment.configuration = [steps:[seleniumExecuteTests:[
            containerPortMappings: []
        ]]]

        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assert dockerExecuteRule.dockerParams.dockerEnvVars.PIPER_SELENIUM_WEBDRIVER_PORT == null
    }

    @Test
    void testNoNullPointerExceptionWithUnrelatedSidecarImageAndContainerPortMapping() {
        nullScript.commonPipelineEnvironment.configuration = [steps:[seleniumExecuteTests:[
            sidecarImage: 'myCustomImage',
            containerPortMappings: [
                'someImageOtherThanMyCustomImage': [
                    [containerPort: 5555, hostPort: 5555]
                ]
            ]
        ]]]
        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assert dockerExecuteRule.dockerParams.dockerEnvVars.PIPER_SELENIUM_WEBDRIVER_PORT == null
    }

    @Test
    void testDockerFromCustomStepConfiguration() {

        def expectedImage = 'image:test'
        def expectedEnvVars = ['env1': 'value1', 'env2': 'value2']
        def expectedOptions = '--opt1=val1 --opt2=val2 --opt3'
        def expectedWorkspace = '/path/to/workspace'

        nullScript.commonPipelineEnvironment.configuration = [steps:[seleniumExecuteTests:[
            dockerImage: expectedImage,
            dockerOptions: expectedOptions,
            dockerEnvVars: expectedEnvVars,
            dockerWorkspace: expectedWorkspace
            ]]]

        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
        }

        assert expectedImage == dockerExecuteRule.dockerParams.dockerImage
        assert expectedOptions == dockerExecuteRule.dockerParams.dockerOptions
        assert expectedWorkspace == dockerExecuteRule.dockerParams.dockerWorkspace
        expectedEnvVars.each { key, value ->
            assert dockerExecuteRule.dockerParams.dockerEnvVars[key] == value
        }
    }

    @Test
    void testExecuteSeleniumCustomBuildTool() {
        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            buildTool: 'maven',
            juStabUtils: utils
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assertThat(dockerExecuteRule.dockerParams.dockerImage, is('maven:3.5-jdk-8'))
        assertThat(dockerExecuteRule.dockerParams.dockerName, is('maven'))
        assertThat(dockerExecuteRule.dockerParams.dockerWorkspace, is(''))
    }
    @Test
    void testExecuteSeleniumError() {
        thrown.expectMessage('Error occured')
        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
            throw new AbortException('Error occured')
        }
    }

    @Test
    void testExecuteSeleniumIgnoreError() {
        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            failOnError: false,
            juStabUtils: utils
        ) {
            bodyExecuted = true
            throw new AbortException('Error occured')
        }
        assertThat(bodyExecuted, is(true))
    }

    @Test
    void testExecuteSeleniumCustomRepo() {
        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            gitBranch: 'test',
            gitSshKeyCredentialsId: 'testCredentials',
            juStabUtils: utils,
            testRepository: 'git@test/test.git'
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assertThat(gitMap, hasEntry('branch', 'test'))
        assertThat(gitMap, hasEntry('credentialsId', 'testCredentials'))
        assertThat(gitMap, hasEntry('url', 'git@test/test.git'))
    }

    @Test
    void testSeleniumHubCredentials() {
        nullScript.commonPipelineEnvironment.configuration = [steps:[seleniumExecuteTests:[
            seleniumHubCredentialsId: 'MyCredentialId'
        ]]]

        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
            bodyExecuted = true
        }

        assertThat(bodyExecuted, is(true))
        assertThat(credentials.size(), is(1))
    }
}
