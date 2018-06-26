import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.Rules

import hudson.AbortException


public class TransportRequestReleaseTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(thrown)
        .around(jsr)
        .around(jlr)

    @Before
    public void setup() {

        helper.registerAllowedMethod('usernamePassword', [Map.class], {m -> return m})

        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->

            credentialsId = l[0].credentialsId
            binding.setProperty('username', 'anonymous')
            binding.setProperty('password', '********')
            try {
                c()
            } finally {
                binding.setProperty('username', null)
                binding.setProperty('password', null)
            }
         })

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 0 })

        nullScript.commonPipelineEnvironment.configuration = [steps:
                                     [transportRequestRelease:
                                         [
                                          cmCredentialsId: 'CM',
                                          cmEndpoint: 'https://example.org/cm'
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void changeIdNotProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Change id not provided (parameter: 'changeId').")

        jsr.step.call(script: nullScript, transportRequestId: '001')
    }

    @Test
    public void transportRequestIdNotProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Transport Request id not provided (parameter: 'transportRequestId').")

        jsr.step.call(script: nullScript, changeId: '001')
    }

    @Test
    public void releaseTransportRequestFailureTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 1 })

        thrown.expect(AbortException)
        thrown.expectMessage("Cannot release Transport Request '001'. Return code from cmclient: 1.")

        jsr.step.call(script: nullScript, changeId: '001', transportRequestId: '001')
    }

    @Test
    public void releaseTransportRequestSuccessTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 0 })

        jsr.step.call(script: nullScript, changeId: '001', transportRequestId: '001')

        assert jlr.log.contains("[INFO] Closing transport request '001' for change document '001'.")
        assert jlr.log.contains("[INFO] Transport Request '001' has been successfully closed.")
    }
}
