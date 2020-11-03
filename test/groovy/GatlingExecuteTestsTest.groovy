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

    List mavenParams = []

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("mavenExecute", [Map], { map -> mavenParams.add(map) })
    }

    @Test
    void pomPathDoesNotExist() throws Exception {
        helper.registerAllowedMethod("fileExists", [String], { path -> return false })
        thrown.expectMessage("The file 'does-not-exist' does not exist.")

        stepRule.step.gatlingExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            pomPath: 'does-not-exist'
        )

        assertJobStatusFailure()
    }

    @Test
    void executionWithoutAppUrls() throws Exception {
        registerPerformanceTestsModule()

        stepRule.step.gatlingExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            pomPath: 'performance-tests/pom.xml'
        )

        assertThat(mavenParams.size(), is(1))

        assertThat(mavenParams[0], is([
            script: nullScript,
            flags: ['--update-snapshots'],
            pomPath: 'performance-tests/pom.xml',
            goals: ['test']
        ]))

        assertJobStatusSuccess()
    }

    @Test
    void executionWithAppUrls() throws Exception {
        registerPerformanceTestsModule()

        final String url1 = 'url1'
        final String url2 = 'url2'
        final String username = 'test-username'
        final String password = 'test-password'

        helper.registerAllowedMethod("withCredentials", [List, Closure], { creds, body ->
            assertThat(creds.size(), is(1))
            binding.setVariable(creds[0].usernameVariable, username)
            binding.setVariable(creds[0].passwordVariable, password)
            body()
        })

        stepRule.step.gatlingExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            pomPath: 'performance-tests/pom.xml',
            appUrls: [[url: url1, credentialsId: 'credentials1'], [url: url2, credentialsId: 'credentials2']]
        )

        assertThat(mavenParams.size(), is(2))

        assertThat(mavenParams[0], is([
            script: nullScript,
            flags: ['--update-snapshots'],
            pomPath: 'performance-tests/pom.xml',
            goals: ['test'],
            defines: [
                "-DappUrl=$url1",
                "-Dusername=$username",
                "-Dpassword=$password"
            ]
        ]))

        assertThat(mavenParams[1], is([
            script: nullScript,
            flags: ['--update-snapshots'],
            pomPath: 'performance-tests/pom.xml',
            goals: ['test'],
            defines: [
                "-DappUrl=$url2",
                "-Dusername=$username",
                "-Dpassword=$password"
            ]
        ]))

        assertJobStatusSuccess()
    }

    private void registerPerformanceTestsModule() {
        helper.registerAllowedMethod("fileExists", [String], { path ->
            switch (path) {
                case "performance-tests/pom.xml":
                    return true
                default:
                    return false
            }
        })
    }
}
