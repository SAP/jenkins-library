import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.allOf

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.junit.rules.ExpectedException
import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsShellCallRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.Rules

class SonarExecuteScanTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule jedr = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(thrown)
        .around(jedr)
        .around(jscr)
        .around(jlr)
        .around(jsr)

    def sonarInstance

    @Before
    void init() throws Exception {
        sonarInstance = null
        helper.registerAllowedMethod("withSonarQubeEnv", [String.class, Closure.class], { string, closure ->
            sonarInstance = string
            return closure()
        })
        helper.registerAllowedMethod("unstash", [String.class], { stashInput -> return [] })
        helper.registerAllowedMethod("fileExists", [String.class], { file -> return file })
        helper.registerAllowedMethod('string', [Map], { m -> m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            try {
                binding.setProperty(l[0].variable, 'TOKEN_' + l[0].credentialsId)
                c()
            } finally {
                binding.setProperty(l[0].variable, null)
            }
        })
        helper.registerAllowedMethod('withEnv', [List.class, Closure.class], { List envVars, Closure body ->
            body()
        })
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3-20180101')
    }

    @Test
    void testWithCustomTlsCertificates() throws Exception {
        jsr.step.sonarExecuteScan.loadCertificates(
            customTlsCertificateLinks: [
                'http://url.to/my.cert'
            ]
        )
        // asserts
        assertThat(jscr.shell, allOf(
            hasItem(containsString('wget --directory-prefix .certificates/ --no-verbose http://url.to/my.cert')),
            hasItem(containsString('keytool -import -noprompt -storepass changeit -keystore .certificates/cacerts -alias \'my.cert\' -file \'.certificates/my.cert\''))
        ))
        assertJobStatusSuccess()
    }
}
