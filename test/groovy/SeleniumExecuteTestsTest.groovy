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
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule jedr = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jedr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

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
        jsr.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assertThat(jedr.dockerParams.containerPortMappings, is(['selenium/standalone-chrome': [[containerPort: 4444, hostPort: 4444]]]))
        assertThat(jedr.dockerParams.dockerEnvVars, is(null))
        assertThat(jedr.dockerParams.dockerImage, is('node:8-stretch'))
        assertThat(jedr.dockerParams.dockerName, is('npm'))
        assertThat(jedr.dockerParams.dockerWorkspace, is('/home/node'))
        assertThat(jedr.dockerParams.sidecarEnvVars, is(null))
        assertThat(jedr.dockerParams.sidecarImage, is('selenium/standalone-chrome'))
        assertThat(jedr.dockerParams.sidecarName, is('selenium'))
        assertThat(jedr.dockerParams.sidecarVolumeBind, is(['/dev/shm': '/dev/shm']))
    }

    @Test
    void testExecuteSeleniumCustomBuildTool() {
        jsr.step.seleniumExecuteTests(
            script: nullScript,
            buildTool: 'maven',
            juStabUtils: utils
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assertThat(jedr.dockerParams.dockerImage, is('maven:3.5-jdk-8'))
        assertThat(jedr.dockerParams.dockerName, is('maven'))
        assertThat(jedr.dockerParams.dockerWorkspace, is(''))
    }
    @Test
    void testExecuteSeleniumError() {
        thrown.expectMessage('Error occured')
        jsr.step.seleniumExecuteTests(
            script: nullScript,
            juStabUtils: utils
        ) {
            throw new AbortException('Error occured')
        }
    }

    @Test
    void testExecuteSeleniumIgnoreError() {
        jsr.step.seleniumExecuteTests(
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
        jsr.step.seleniumExecuteTests(
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
