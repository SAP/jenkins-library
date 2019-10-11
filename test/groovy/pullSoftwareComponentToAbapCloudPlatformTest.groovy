import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString

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

public class TransportRequestCreateTest extends BasePiperTest {

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
    public void changeIdNotProvidedSOLANTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' or via commit history).")

        stepRule.step.pullSoftwareComponentToAbapCloudPlatform(script: nullScript, host: 'example.com', user: 'user', password: 'password')
    }
}