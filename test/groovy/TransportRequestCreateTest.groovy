import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.Rules

import hudson.AbortException

public class TransportRequestCreateTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jsr)
        .around(jlr)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('CM', 'anonymous', '********'))

    @Before
    public void setup() {


        nullScript.commonPipelineEnvironment.configuration = [general:
                                     [changeManagement:

                                         [
                                          credentialsId: 'CM',
                                          type: 'SOLMAN',
                                          endpoint: 'https://example.org/cm',
                                          clientOpts: '-DmyProp=myVal',
                                          changeDocumentLabel: 'ChangeId\\s?:',
                                          git: [from: 'origin/master',
                                                to: 'HEAD',
                                                format: '%b']
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void changeIdNotProvidedTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' or via commit history).")
        ChangeManagement cm = new ChangeManagement(nullScript) {
            String getChangeDocumentId(
                                       String from,
                                       String to,
                                       String label,
                                       String format
                                      ) {
                                          throw new ChangeManagementException('Cannot retrieve changeId from git commits.')
                                      }
        }

        jsr.step.transportRequestCreate(script: nullScript, developmentSystemId: '001', cmUtils: cm)
    }

    @Test
    public void developmentSystemIdNotProvidedTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR developmentSystemId")

        jsr.step.transportRequestCreate(script: nullScript, changeDocumentId: '001')
    }

    @Test
    public void createTransportRequestFailureTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {

            String createTransportRequestSOLMAN(
                                          String changeId,
                                          String developmentSystemId,
                                          String cmEndpoint,
                                          String credentialId,
                                          String clientOpts) {

                    throw new ChangeManagementException('Exception message.')
            }
        }


        thrown.expect(AbortException)
        thrown.expectMessage("Exception message.")

        jsr.step.transportRequestCreate(script: nullScript, changeDocumentId: '001', developmentSystemId: '001', cmUtils: cm)
    }

    @Test
    public void createTransportRequestSuccessSOLMANTest() {

        def result = [:]

        ChangeManagement cm = new ChangeManagement(nullScript) {

            String createTransportRequestSOLMAN(
                                          String changeId,
                                          String developmentSystemId,
                                          String cmEndpoint,
                                          String credentialId,
                                          String clientOpts) {

                result.changeId = changeId
                result.developmentSystemId = developmentSystemId
                result.cmEndpoint = cmEndpoint
                result.credentialId = credentialId
                result.clientOpts = clientOpts
                return '001'
            }
        }

        def transportId = jsr.step.transportRequestCreate(script: nullScript, changeDocumentId: '001', developmentSystemId: '001', cmUtils: cm)

        assert transportId == '001'
        assert result == [changeId: '001',
                         developmentSystemId: '001',
                         cmEndpoint: 'https://example.org/cm',
                         credentialId: 'CM',
                         clientOpts: '-DmyProp=myVal'
                         ]

        assert jlr.log.contains("[INFO] Creating transport request for change document '001' and development system '001'.")
        assert jlr.log.contains("[INFO] Transport Request '001' has been successfully created.")
    }

    @Test
    public void createTransportRequestSuccessCTSTest() {

        def result = [:]

        ChangeManagement cm = new ChangeManagement(nullScript) {

            String createTransportRequestCTS(
                String transportType,
                String targetSystemId,
                String description,
                String endpoint,
                String credentialsId,
                String clientOpts
) {
                result.transportType = transportType
                result.targetSystemId = targetSystemId
                result.description = description
                result.endpoint = endpoint
                result.credentialsId = credentialsId
                result.clientOpts = clientOpts
                return '001'
            }
        }

        def transportId = jsr.step.transportRequestCreate(script: nullScript,
                                        transportType: 'W',
                                        targetSystem: 'XYZ',
                                        description: 'desc',
                                        changeManagement: [type: 'CTS'],
                                        cmUtils: cm)

        assert transportId == '001'
        assert result == [transportType: 'W',
                         targetSystemId: 'XYZ',
                         description: 'desc',
                         endpoint: 'https://example.org/cm',
                         credentialsId: 'CM',
                         clientOpts: '-DmyProp=myVal'
                         ]

        assert jlr.log.contains("[INFO] Creating transport request.")
        assert jlr.log.contains("[INFO] Transport Request '001' has been successfully created.")
    }

    @Test
    public void cmIntegrationSwichtedOffTest() {

        jlr.expect('[INFO] Change management integration intentionally switched off.')

        jsr.step.transportRequestCreate(script: nullScript,
            changeManagement: [type: 'NONE'])
    }
}
