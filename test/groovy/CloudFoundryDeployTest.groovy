import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsStepRule
import util.Rules
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

class CloudFoundryDeployTest extends BasePiperTest {

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(credentialsRule)
        .around(stepRule) // needs to be activated after dockerExecuteRule, otherwise executeDocker is not mocked

    @Before
    void init() {
        // removing additional credentials tests might have added; adding default credentials
        credentialsRule.reset()
            .withCredentials('test_cfCredentialsId', 'test_cf', '********')
    }

    @Test
    void testGoStepWithMtaExtensionCredentialsFromParams() {
        String calledStep = ''
        String usedMetadataFile = ''
        List credInfo = []
        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            Map parameters, String stepName,
            String metadataFile, List credentialInfo ->
                calledStep = stepName
                usedMetadataFile = metadataFile
                credInfo = credentialInfo
        })

        stepRule.step.cloudFoundryDeploy([
            script                 : nullScript,
            juStabUtils            : utils,
            useGoStep              : true,
            mtaExtensionCredentials: [myCred: 'Mta.ExtensionCredential~Credential_Id1'],
        ])

        assertEquals('cloudFoundryDeploy', calledStep)
        assertEquals('metadata/cloudFoundryDeploy.yaml', usedMetadataFile)

        // contains assertion does not work apparently when comparing a list of lists agains an expected list.
        boolean found = false
        credInfo.each { entry ->
            if (entry == [type: 'token', id: 'Mta.ExtensionCredential~Credential_Id1', env: ['MTA_EXTENSION_CREDENTIAL_CREDENTIAL_ID1'], resolveCredentialsId: false]) {
                found = true
            }
        }
        assertTrue(found)
    }

    @Test
    void testGoStepWithMtaExtensionCredentialsFromConfig() {
        String calledStep = ''
        String usedMetadataFile = ''
        List credInfo = []
        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            Map parameters, String stepName,
            String metadataFile, List credentialInfo ->
                calledStep = stepName
                usedMetadataFile = metadataFile
                credInfo = credentialInfo
        })

        nullScript.commonPipelineEnvironment.configuration = [steps:[cloudFoundryDeploy:[
            mtaExtensionCredentials: [myCred: 'Mta.ExtensionCredential~Credential_Id1']
        ]]]

        stepRule.step.cloudFoundryDeploy([
            script                 : nullScript,
            juStabUtils            : utils,
            useGoStep              : true,
        ])

        assertEquals('cloudFoundryDeploy', calledStep)
        assertEquals('metadata/cloudFoundryDeploy.yaml', usedMetadataFile)

        // contains assertion does not work apparently when comparing a list of lists agains an expected list.
        boolean found = false
        credInfo.each { entry ->
            if (entry == [type: 'token', id: 'Mta.ExtensionCredential~Credential_Id1', env: ['MTA_EXTENSION_CREDENTIAL_CREDENTIAL_ID1'], resolveCredentialsId: false]) {
                found = true
            }
        }
        assertTrue(found)
    }
}
