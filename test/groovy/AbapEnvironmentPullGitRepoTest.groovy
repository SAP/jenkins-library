import java.util.Map
import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import org.junit.Before
import org.junit.After
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsStepRule
import util.JenkinsReadJsonRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsDockerExecuteRule
import util.JenkinsWriteFileRule
import util.JenkinsShellCallRule
import util.Rules

import hudson.AbortException

public class AbapEnvironmentPullGitRepoTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this).withCredentials('test_credentialsId', 'user', 'password')
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(dockerExecuteRule)
        .around(stepRule)
        .around(loggingRule)
        .around(readJsonRule)
        .around(credentialsRule)
        .around(shellRule)
        .around(writeFileRule)

    private List withEnvArgs = []

    @Test
    public void test() {
        helper.registerAllowedMethod("fileExists", [String.class], { file -> return false })
        helper.registerAllowedMethod("fileExists", [Map.class], { file -> return false})
        helper.registerAllowedMethod('findFiles', [Map.class], {m -> return null})
        helper.registerAllowedMethod("withEnv", [List.class, Closure.class], {arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })
        credentialsRule.withCredentials('credentialsId', 'testUser', 'testPassword')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.\/piper getConfig --contextConfig --stepMetadata '.pipeline\/tmp\/metadata\/abapEnvironmentPullGitRepo.yaml'/, /{"credentialsId":"credentialsId"}/ )

        stepRule.step.abapEnvironmentPullGitRepo(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            host: 'example.com', 
            repositoryName: 'Z_DEMO_DM', 
            credentialsId: 'test_credentialsId'
        )

        assertThat(shellRule.shell[0], containsString(/.\/piper getConfig --contextConfig --stepMetadata '.pipeline\/tmp\/metadata\/abapEnvironmentPullGitRepo.yaml'/))
        assertThat(shellRule.shell[1], containsString(/.\/piper abapEnvironmentPullGitRepo/))
    }
}
