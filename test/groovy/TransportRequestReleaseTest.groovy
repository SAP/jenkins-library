import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import groovy.json.StringEscapeUtils
import util.BasePiperTest
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.Rules

import hudson.AbortException
import hudson.scm.NullSCM


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
                                          credentialsId: 'CM',
                                          endpoint: 'https://example.org/cm'
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void changeIdNotProvidedTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR changeDocumentId")

        jsr.step.call(script: nullScript, transportRequestId: '001')
    }

    @Test
    public void transportRequestIdNotProvidedTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {
            String getTransportRequestId(String from,
                                         String to,
                                         String label,
                                         String format) {
                throw new ChangeManagementException('Cannot retrieve transportRequestId')
            }
        }

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Transport request id not provided (parameter: 'transportRequestId' or via commit history).")

        jsr.step.call(script: nullScript, changeDocumentId: '001', cmUtils: cm)
    }

    @Test
    public void releaseTransportRequestFailureTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 1 })

        thrown.expect(AbortException)
        thrown.expectMessage("Cannot release Transport Request '001'. Return code from cmclient: 1.")

        jsr.step.call(script: nullScript, changeDocumentId: '001', transportRequestId: '001')
    }

    @Test
    public void releaseTransportRequestSuccessTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 0 })

        jsr.step.call(script: nullScript, changeDocumentId: '001', transportRequestId: '001')

        assert jlr.log.contains("[INFO] Closing transport request '001' for change document '001'.")
        assert jlr.log.contains("[INFO] Transport Request '001' has been successfully closed.")
    }
}
