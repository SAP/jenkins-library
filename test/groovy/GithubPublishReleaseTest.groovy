import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsFileExistsRule
import util.JenkinsLoggingRule
import util.JenkinsReadJsonRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class GithubPublishReleaseTest extends BasePiperTest {

    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])

    private List withEnvArgs = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(credentialsRule)
        .around(readJsonRule)
        .around(shellCallRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(fileExistsRule)

    @Before
    void init() {
        helper.registerAllowedMethod('findFiles', [Map.class], {return null})
        helper.registerAllowedMethod("withEnv", [List.class, Closure.class], {arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
        credentialsRule.withCredentials('githubTokenId', 'thisIsATestToken')
        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/githubPublishRelease.yaml\'', '{"githubTokenCredentialsId":"githubTokenId"}')
        helper.registerAllowedMethod("writePipelineEnv", [Map.class], {m -> return })
        helper.registerAllowedMethod("readPipelineEnv", [Map.class], {m -> return })
    }

    @Test
    void testGithubPublishReleaseDefault() {
        // test
        stepRule.step.githubPublishRelease(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content"
        )
        // asserts
        assertThat(writeFileRule.files['.pipeline/tmp/metadata/githubPublishRelease.yaml'], containsString('name: githubPublishRelease'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[2], is('./piper githubPublishRelease'))
    }
}
