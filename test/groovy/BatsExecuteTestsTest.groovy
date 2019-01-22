import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.startsWith
import static org.junit.Assert.assertThat

class BatsExecuteTestsTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule jder = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jder)
        .around(shellRule)
        .around(loggingRule)
        .around(jsr)

    List withEnvArgs = []

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("withEnv", [List.class, Closure.class], {arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
    }

    @Test
    void testDefault() {
        nullScript.commonPipelineEnvironment.configuration = [general: [container: 'test-container']]
        jsr.step.batsExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            dockerContainerName: 'test-container',
            dockerImageNameAndTag: 'test/image',
            envVars: [
                IMAGE_NAME: 'test/image',
                CONTAINER_NAME: '${commonPipelineEnvironment.configuration.general.container}'
            ],
            testPackage: 'testPackage'
        )
        // asserts
        assertThat(withEnvArgs, hasItem('IMAGE_NAME=test/image'))
        assertThat(withEnvArgs, hasItem('CONTAINER_NAME=test-container'))
        assertThat(shellRule.shell, hasItem('git clone https://github.com/bats-core/bats-core.git'))
        assertThat(shellRule.shell, hasItem('bats-core/bin/bats --recursive --tap src/test > \'TEST-testPackage.tap\''))
        assertThat(shellRule.shell, hasItem('cat \'TEST-testPackage.tap\''))

        assertThat(jder.dockerParams.dockerImage, is('node:8-stretch'))
        assertThat(jder.dockerParams.dockerWorkspace, is('/home/node'))

        assertThat(shellRule.shell, hasItem('npm install tap-xunit -g'))
        assertThat(shellRule.shell, hasItem('cat \'TEST-testPackage.tap\' | tap-xunit --package=\'testPackage\' > TEST-testPackage.xml'))

        assertJobStatusSuccess()
    }

    @Test
    void testTap() {
        jsr.step.batsExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            outputFormat: 'tap'
        )
        assertThat(jder.dockerParams.size(), is(0))
    }

    @Test
    void testFailOnError() {
        helper.registerAllowedMethod('sh', [String.class], {s ->
            if (s.startsWith('bats-core/bin/bats')) {
                throw new Exception('Shell call failed')
            } else {
                return null
            }
        })
        thrown.expectMessage('Shell call failed')
        jsr.step.batsExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            failOnError: true,
        )

    }

    @Test
    void testGit() {
        def gitRepository
        helper.registerAllowedMethod('git', [Map.class], {m ->
            gitRepository = m
        })
        helper.registerAllowedMethod('stash', [String.class], {s ->
            assertThat(s, startsWith('testContent-'))
        })

        jsr.step.batsExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            testRepository: 'testRepo',
        )

        assertThat(gitRepository.size(), is(1))
        assertThat(gitRepository.url, is('testRepo'))
        assertThat(jder.dockerParams.stashContent, hasItem(startsWith('testContent-')))
    }

    @Test
    void testGitBranchAndCredentials() {
        def gitRepository
        helper.registerAllowedMethod('git', [Map.class], {m ->
            gitRepository = m
        })
        helper.registerAllowedMethod('stash', [String.class], {s ->
            assertThat(s, startsWith('testContent-'))
        })

        jsr.step.batsExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            gitBranch: 'test',
            gitSshKeyCredentialsId: 'testCredentials',
            testRepository: 'testRepo',
        )
        assertThat(gitRepository.size(), is(3))
        assertThat(gitRepository.credentialsId, is('testCredentials'))
        assertThat(gitRepository.branch, is('test'))
    }


}
