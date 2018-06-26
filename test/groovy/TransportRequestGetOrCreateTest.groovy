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


public class TransportRequestGetOrCreateTest extends BasePiperTest {

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
                                     [transportRequestGetOrCreate:
                                         [
                                          credentialsId: 'CM',
                                          endpoint: 'https://example.org/cm'
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void changeDocumentIdNotProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Change id not provided (parameter: 'changeDocumentId').")

        jsr.step.call(script: nullScript)
    }

    @Test
    public void developmentSystemIdNotProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Development system id not provided (parameter: 'developmentSystemId').")

        jsr.step.call(script: nullScript, changeDocumentId: '001')
    }

    @Test
    public void getTransportRequestsFailureTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> throw new AbortException('Exception message.') })

        thrown.expect(AbortException)
        thrown.expectMessage("Cannot get the transport requests for change document '001'. Exception message.")

        jsr.step.call(script: nullScript, changeDocumentId: '001', developmentSystemId: '001')
    }

    @Test
    public void tooManyTransportRequestsTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '001\n002\n003' })

        thrown.expect(AbortException)
        thrown.expectMessage("Too many open transport requests [001, 002, 003] for change document '001'.")

        jsr.step.call(script: nullScript, changeDocumentId: '001', developmentSystemId: '001')
    }

    @Test
    public void noOpenTransportRequestsTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '' })

        jsr.step.call(script: nullScript, changeDocumentId: '001', developmentSystemId: '001')

        assert jlr.log.contains("[INFO] Getting transport requests for change document '001'.")
        assert jlr.log.contains("[INFO] There is no open transport requests for change document '001'.")
    }

    @Test
    public void getTransportRequestsSuccessTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '005' })

        jsr.step.call(script: nullScript, changeDocumentId: '001', developmentSystemId: '001')

        assert jlr.log.contains("[INFO] Getting transport requests for change document '001'.")
        assert jlr.log.contains("[INFO] Transport request '005' available for change document '001'.")
    }
}
