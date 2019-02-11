#!groovy

import static org.hamcrest.Matchers.*

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.junit.rules.ExpectedException

import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

class UiVeri5ExecuteTestsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsDockerExecuteRule dockerRule = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(new JenkinsReadYamlRule(this))
        .around(dockerRule)
        .around(loggingRule)
        .around(stepRule)

    def gitParams = [:]
    def shellCommands = []
    def seleniumMap = [:]

    class MockBuild {
        MockBuild(){}
        def getRootDir(){
            return new MockPath()
        }
        class MockPath {
            MockPath(){}
            // return default
            String getAbsolutePath(){
                return 'myPath'
            }
        }
    }

    @Before
    void init() {

        binding.setVariable('currentBuild', [
            result: 'SUCCESS',
            rawBuild: new MockBuild()
        ])

        helper.registerAllowedMethod("git", [Map.class], { map -> gitParams = map})
        helper.registerAllowedMethod("stash", [String.class], null)
        helper.registerAllowedMethod("unstash", [String.class], { s -> return [s]})
        helper.registerAllowedMethod("sh", [String.class], { s ->
            if (s.contains('failure')) throw new RuntimeException('Test Error')
            shellCommands.add(s.toString())
        })
        helper.registerAllowedMethod("sh", [Map.class], { map ->
            return 'available'
        })
        helper.registerAllowedMethod('seleniumExecuteTests', [Map.class, Closure.class], { m, body ->
            seleniumMap = m
            return body()
        })
    }

    @Test
    void testDefault() throws Exception {
        // execute test
        stepRule.step.uiVeri5ExecuteTests([
            script: nullScript,
            juStabUtils: utils,
        ])
        // asserts
        assertThat(shellCommands, hasItem(containsString('npm install @ui5/uiveri5 --global --quiet')))
        assertThat(shellCommands, hasItem(containsString('uiveri5 --seleniumAddress=\'http://selenium:4444/wd/hub\'')))
        assertThat(seleniumMap.dockerImage, isEmptyOrNullString())
        assertThat(seleniumMap.dockerWorkspace, isEmptyOrNullString())
    }

    @Test
    void testDefaultOnK8s() throws Exception {
        // prepare
        binding.variables.env.ON_K8S = 'true'
        // execute test
        stepRule.step.uiVeri5ExecuteTests([
            script: nullScript,
            juStabUtils: utils,
        ])
        // asserts
        assertThat(shellCommands, hasItem(containsString('npm install @ui5/uiveri5 --global --quiet')))
        assertThat(shellCommands, hasItem(containsString('uiveri5 --seleniumAddress=\'http://localhost:4444/wd/hub\'')))
    }

    @Test
    void testWithCustomSidecar() throws Exception {
        // execute test
        stepRule.step.uiVeri5ExecuteTests([
            script: nullScript,
            juStabUtils: utils,
            sidecarEnvVars: [myEnv: 'testValue'],
            sidecarImage: 'myImage'
        ])
        // asserts
        assertThat(seleniumMap.sidecarImage, is('myImage'))
        assertThat(seleniumMap.sidecarEnvVars.myEnv, is('testValue'))
    }

    @Test
    void testWithTestRepository() throws Exception {
        // execute test
        stepRule.step.uiVeri5ExecuteTests([
            script: nullScript,
            juStabUtils: utils,
            testRepository: 'git@myGitUrl'
        ])
        // asserts
        assertThat(seleniumMap, hasKey('stashContent'))
        assertThat(seleniumMap.stashContent, hasItem(startsWith('testContent-')))
        assertThat(gitParams, hasEntry('url', 'git@myGitUrl'))
        assertThat(gitParams, not(hasKey('credentialsId')))
        assertThat(gitParams, not(hasKey('branch')))
        assertJobStatusSuccess()
    }

    @Test
    void testWithTestRepositoryWithGitBranchAndCredentials() throws Exception {
        // execute test
        stepRule.step.uiVeri5ExecuteTests([
            script: nullScript,
            juStabUtils: utils,
            testRepository: 'git@myGitUrl',
            gitSshKeyCredentialsId: 'myCredentials',
            gitBranch: 'myBranch'
        ])
        // asserts
        assertThat(gitParams, hasEntry('url', 'git@myGitUrl'))
        assertThat(gitParams, hasEntry('credentialsId', 'myCredentials'))
        assertThat(gitParams, hasEntry('branch', 'myBranch'))
        assertJobStatusSuccess()
    }

    @Test
    void testWithFailOnError() throws Exception {
        thrown.expect(RuntimeException)
        thrown.expectMessage('Test Error')
        // execute test
        stepRule.step.uiVeri5ExecuteTests([
            juStabUtils: utils,
            failOnError: true,
            testOptions: 'failure',
            script: nullScript
        ])
    }

    @Test
    void testWithoutFailOnError() throws Exception {
        // execute test
        stepRule.step.uiVeri5ExecuteTests([
            juStabUtils: utils,
            testOptions: 'failure',
            script: nullScript
        ])
        // asserts
        assertJobStatusSuccess()
    }
}
