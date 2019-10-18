import java.util.Map

import org.hamcrest.Matchers
import org.hamcrest.core.StringContains
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.Rules

import hudson.AbortException

public class PullSoftwareComponentToAbapCloudPlatformTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(stepRule)
        .around(loggingRule)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('CM', 'anonymous', '********'))

    @Before
    public void setup() {
    }

    @Test
    public void test() {
    }

    @Test
    public void checkRepositoryProvided() {
       thrown.expect(IllegalArgumentException)
       thrown.expectMessage("Repository / Software Component not provided")
       stepRule.step.pullSoftwareComponentToAbapCloudPlatform(script: nullScript, host: 'https://www.example.com', username: 'user', password: 'password')
    }


    @Test
    public void checkHostProvided() {
       thrown.expect(IllegalArgumentException)
       thrown.expectMessage("Host not provided")
       stepRule.step.pullSoftwareComponentToAbapCloudPlatform(script: nullScript, repositoryName: 'REPO', username: 'user', password: 'password')
    }
}