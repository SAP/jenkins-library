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


public class TransportRequestCreateTest extends BasePiperTest {

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
                                     [transportRequestCreate:
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

        jsr.step.call(script: nullScript, developmentSystemId: '001')
    }

    @Test
    public void developmentSystemIdNotProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Development system id not provided (parameter: 'developmentSystemId').")

        jsr.step.call(script: nullScript, changeId: '001')
    }

    @Test
    public void createTransportRequestFailureTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> throw new AbortException('Exception message.') })

        thrown.expect(AbortException)
        thrown.expectMessage("Cannot create a transport request for change id '001'. Exception message.")

        jsr.step.call(script: nullScript, changeId: '001', developmentSystemId: '001')
    }

    @Test
    public void createTransportRequestSuccessTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '001' })

        jsr.step.call(script: nullScript, changeId: '001', developmentSystemId: '001')

        assert jlr.log.contains("[INFO] Creating transport request for change document '001' and development system '001'.")
        assert jlr.log.contains("[INFO] Transport Request '001' has been successfully created.")
    }
}
