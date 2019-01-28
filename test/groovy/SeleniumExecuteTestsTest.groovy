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
        bodyExecuted = false
        helper.registerAllowedMethod('stash', [String.class], null)
        helper.registerAllowedMethod('git', [Map.class], {m ->
            gitMap = m
        })
    }

    @Test
    void testExecuteSeleniumDefault() {
        stepRule.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assertThat(dockerExecuteRule.dockerParams.containerPortMappings, is(['selenium/standalone-chrome': [[containerPort: 4444, hostPort: 4444]]]))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, is(null))
        assertThat(dockerExecuteRule.dockerParams.dockerImage, is('node:8-stretch'))
        assertThat(dockerExecuteRule.dockerParams.dockerName, is('npm'))
        assertThat(dockerExecuteRule.dockerParams.dockerWorkspace, is('/home/node'))
        assertThat(dockerExecuteRule.dockerParams.sidecarEnvVars, is(null))
        assertThat(dockerExecuteRule.dockerParams.sidecarImage, is('selenium/standalone-chrome'))
        assertThat(dockerExecuteRule.dockerParams.sidecarName, is('selenium'))
        assertThat(dockerExecuteRule.dockerParams.sidecarVolumeBind, is(['/dev/shm': '/dev/shm']))
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
}
