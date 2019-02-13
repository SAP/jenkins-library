#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class GaugeExecuteTestsTest extends BasePiperTest {
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(shellRule)
        .around(loggingRule)
        .around(environmentRule)
        .around(stepRule)
        .around(thrown)

    def gitParams = [:]
    def seleniumParams = [:]

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("git", [Map.class], { map -> gitParams = map })
        helper.registerAllowedMethod("unstash", [String.class], { s -> return [s]})

        helper.registerAllowedMethod('seleniumExecuteTests', [Map.class, Closure.class], {map, body ->
            seleniumParams = map
            return body()
        })
    }

    @Test
    void testExecuteGaugeDefaultSuccess() throws Exception {
        stepRule.step.gaugeExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            testServerUrl: 'http://test.url'
        )
        assertThat(shellRule.shell, hasItem(stringContainsInOrder([
            'export HOME=${HOME:-$(pwd)}',
            'if [ "$HOME" = "/" ]; then export HOME=$(pwd); fi',
            'export PATH=$HOME/bin/gauge:$PATH',
            'mkdir -p $HOME/bin/gauge',
            'curl -SsL https://downloads.gauge.org/stable | sh -s -- --location=$HOME/bin/gauge',
            'gauge telemetry off',
            'gauge install java',
            'gauge install html-report',
            'gauge install xml-report',
            'mvn test-compile gauge:execute -DspecsDir=specs'
        ])))
        assertThat(seleniumParams.dockerImage, is('maven:3.5-jdk-8'))
        assertThat(seleniumParams.dockerEnvVars, hasEntry('TARGET_SERVER_URL', 'http://test.url'))
        assertThat(seleniumParams.dockerName, is('maven'))
        assertThat(seleniumParams.dockerWorkspace, is(''))
        assertThat(seleniumParams.stashContent, hasSize(2))
        assertThat(seleniumParams.stashContent, allOf(hasItem('buildDescriptor'), hasItem('tests')))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteGaugeNode() throws Exception {
        stepRule.step.gaugeExecuteTests(
            script: nullScript,
            buildTool: 'npm',
            dockerEnvVars: ['TARGET_SERVER_URL':'http://custom.url'],
            juStabUtils: utils,
            testOptions: 'testSpec'
        )
        assertThat(shellRule.shell, hasItem(stringContainsInOrder([
            'gauge install js',
            'gauge run testSpec'
        ])))
        assertThat(seleniumParams.dockerImage, is('node:8-stretch'))
        assertThat(seleniumParams.dockerEnvVars, hasEntry('TARGET_SERVER_URL', 'http://custom.url'))
        assertThat(seleniumParams.dockerName, is('npm'))
        assertThat(seleniumParams.dockerWorkspace, is('/home/node'))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteCustomWithError() throws Exception {
        helper.registerAllowedMethod("sh", [String.class], { s ->
            throw new RuntimeException('Test Error')
        })
        thrown.expect(RuntimeException)
        thrown.expectMessage('Test Error')
        try {
            stepRule.step.gaugeExecuteTests(
                script: nullScript,
                juStabUtils: utils,
                dockerImage: 'testImage',
                dockerName: 'testImageName',
                dockerWorkspace: '/home/test',
                failOnError: true,
                stashContent: ['testStash'],
            )

        } finally{
            assertThat(seleniumParams.dockerImage, is('testImage'))
            assertThat(seleniumParams.dockerName, is('testImageName'))
            assertThat(seleniumParams.dockerWorkspace, is('/home/test'))
            assertThat(seleniumParams.stashContent, hasSize(1))
            assertThat(loggingRule.log, containsString('[gaugeExecuteTests] One or more tests failed'))
            assertThat(nullScript.currentBuild.result, is('UNSTABLE'))

        }

    }

    @Test
    void testExecuteGaugeCustomRepo() throws Exception {
        helper.registerAllowedMethod('git', [String.class], null)
        helper.registerAllowedMethod('stash', [String.class], null)

        stepRule.step.gaugeExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            testRepository: 'myTestRepo',
            failOnError: true
        )

        // nested matchers do not work correctly
        assertThat(seleniumParams.stashContent, hasItem(startsWith('testContent-')))
        assertJobStatusSuccess()
    }
}
