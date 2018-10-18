#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class HealthExecuteCheckTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)
        .around(thrown)


    @Before
    void init() throws Exception {
        // register Jenkins commands with mock values
        def command1 = "curl -so /dev/null -w '%{response_code}' http://testserver"
        def command2 = "curl -so /dev/null -w '%{response_code}' http://testserver/endpoint"
        helper.registerAllowedMethod('sh', [Map.class], {map ->
            return map.script == command1 || map.script == command2 ? "200" : "404"
        })
    }

    @Test
    void testHealthCheckOk() throws Exception {
        def testUrl = 'http://testserver/endpoint'

        jsr.step.healthExecuteCheck(
            script: nullScript,
            testServerUrl: testUrl
        )

        assertThat(jlr.log, containsString("Health check for ${testUrl} successful"))
    }

    @Test
    void testHealthCheck404() throws Exception {
        def testUrl = 'http://testserver/404'

        thrown.expect(Exception)
        thrown.expectMessage('Health check failed: 404')

        jsr.step.healthExecuteCheck(
            script: nullScript,
            testServerUrl: testUrl
        )
    }


    @Test
    void testHealthCheckWithEndPoint() throws Exception {
        jsr.step.healthExecuteCheck(
            script: nullScript,
            testServerUrl: 'http://testserver',
            healthEndpoint: 'endpoint'
        )

        assertThat(jlr.log, containsString("Health check for http://testserver/endpoint successful"))
    }

    @Test
    void testHealthCheckWithEndPointTrailingSlash() throws Exception {
        jsr.step.healthExecuteCheck(
            script: nullScript,
            testServerUrl: 'http://testserver/',
            healthEndpoint: 'endpoint'
        )

        assertThat(jlr.log, containsString("Health check for http://testserver/endpoint successful"))
    }

}
