import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

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
    def archiveStepPatterns
    @Before
    void init() throws Exception {
        sonarInstance = null
        archiveStepPatterns = []
        helper.registerAllowedMethod("withSonarQubeEnv", [String.class, Closure.class], { string, closure ->
            sonarInstance = string
            return closure()
        })
        helper.registerAllowedMethod("unstash", [String.class], { stashInput -> return [] })
        helper.registerAllowedMethod("fileExists", [String.class], { file -> return file })
        helper.registerAllowedMethod('string', [Map], { m -> m })
        helper.registerAllowedMethod("archiveArtifacts", [Map.class], {
            parameters -> archiveStepPatterns.push(parameters.artifacts)
        })
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

    @Test
    void testExecuteScript() {
        def piperGoPath = "/usr/local/bin/piper"
        def stepName = "my-step"
        def customDefaultConfig = "--config=custom"
        def customConfigArg = "-Dfoo=bar"
        def shouldArchiveArtifacts = true

        jsr.step.executeScript(
            piperGoPath: piperGoPath,
            stepName: stepName,
            customDefaultConfig: customDefaultConfig,
            customConfigArg: customConfigArg,
            shouldArchiveArtifacts: shouldArchiveArtifacts
        )
        def shellCommand = jscr.shell.join(" ")
        assertThat(shellCommand, containsString("-Dfoo=bar"))
        if (shouldArchiveArtifacts) {
            assertThat(archiveStepPatterns, hasSize(1))
            assertThat(archiveStepPatterns, hasItem('sonarscan.json'))
        }
    }
}
