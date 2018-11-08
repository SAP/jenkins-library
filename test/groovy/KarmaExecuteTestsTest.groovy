#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class KarmaExecuteTestsTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jscr)
        .around(jlr)
        .around(jer)
        .around(jsr)
        .around(thrown)

    def seleniumParams = [:]

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("unstash", [String.class], { s -> return [s]})

        helper.registerAllowedMethod('seleniumExecuteTests', [Map.class, Closure.class], {map, body ->
            seleniumParams = map
            return body()
        })
    }

    @Test
    void testDefaults() throws Exception {
        jsr.step.karmaExecuteTests(
            script: nullScript,
            juStabUtils: utils
        )
        assertThat(jscr.shell, hasItems(
            containsString("cd '.' && npm install --quiet"),
            containsString("cd '.' && npm run karma")
        ))
        assertThat(seleniumParams.dockerImage, is('node:8-stretch'))
        assertThat(seleniumParams.dockerName, is('karma'))
        assertThat(seleniumParams.dockerWorkspace, is('/home/node'))
        assertJobStatusSuccess()
    }
}
