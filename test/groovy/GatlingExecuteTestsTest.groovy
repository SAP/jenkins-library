import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class GatlingExecuteTestsTest extends BasePiperTest {
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(stepRule)
        .around(thrown)

    def mavenParams = [:]

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("mavenExecute", [Map], { map -> mavenParams = map })
    }

    @Test
    void testModuleDoesNotExist() throws Exception {
        helper.registerAllowedMethod("fileExists", [String], { path -> return false })
        thrown.expectMessage("The Maven module 'does-not-exist' does not exist.")

        stepRule.step.gatlingExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            testModule: 'does-not-exist'
        )

        assertJobStatusFailure()
    }

    @Test
    void executionWithoutAppUrls() throws Exception {
        helper.registerAllowedMethod("fileExists", [String], { path ->
            switch (path) {
            case "performance-tests/pom.xml":
                return true
            default:
                return false
            }
        })

        stepRule.step.gatlingExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            testModule: 'performance-tests/pom.xml'
        )

        assertThat(mavenParams, is([
            script: nullScript,
            flags: ['--update-snapshots'],
            pomPath: 'performance-tests/pom.xml',
            goals: ['test']
        ]))

        assertJobStatusSuccess()
    }
}
