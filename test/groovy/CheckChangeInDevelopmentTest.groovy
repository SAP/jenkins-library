import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

import org.junit.Before
import org.junit.After
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException
import com.sap.piper.Utils

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

class CheckChangeInDevelopmentTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(stepRule)
        .around(loggingRule)
        .around(new JenkinsCredentialsRule(this)
        .withCredentials('CM', 'anonymous', '********'))

    @Before
    public void setup() {
        Utils.metaClass.echo = { def m -> }

        nullScript.commonPipelineEnvironment.configuration = [general:
                                     [changeManagement:
                                         [
                                          credentialsId: 'CM',
                                          type: 'SOLMAN',
                                          endpoint: 'https://example.org/cm'
                                         ]
                                     ]
                                 ]
        helper.registerAllowedMethod('addBadge', [Map], {return})
        helper.registerAllowedMethod('createSummary', [Map], {return})
    }

    @After
    public void tearDown() {
        cmUtilReceivedParams.clear()
        Utils.metaClass = null
    }

    private Map cmUtilReceivedParams = [:]

    @Test
    public void changeIsInStatusDevelopmentTest() {

        def calledWithParameters,
            calledWithStepName

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                if(stepName.equals("isChangeInDevelopment")) {
                    nullScript.commonPipelineEnvironment.setValue('isChangeInDevelopment', true)
                }
            })

        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            changeDocumentId: '001',
            failIfStatusIsNotInDevelopment: true)

        assertThat(nullScript.commonPipelineEnvironment.getValue('isChangeInDevelopment'), is(true))

        // no exception in thrown, so the change is in status 'in development'.
    }

    @Test
    public void changeIsNotInStatusDevelopmentTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Change '001' is not in status 'in development'")

        def calledWithParameters,
            calledWithStepName

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                if(stepName.equals("isChangeInDevelopment")) {
                    nullScript.commonPipelineEnvironment.setValue('isChangeInDevelopment', false)
                }
            })

        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            changeDocumentId: '001',
            failIfStatusIsNotInDevelopment: true)
    }

    @Test
    public void changeIsNotInStatusDevelopmentButWeWouldLikeToSkipFailureTest() {

        def calledWithParameters,
            calledWithStepName

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                if(stepName.equals("isChangeInDevelopment")) {
                    nullScript.commonPipelineEnvironment.setValue('isChangeInDevelopment', false)
                }
            })

        stepRule.step.checkChangeInDevelopment(
                                    script: nullScript,
                                    changeDocumentId: '001',
                                    failIfStatusIsNotInDevelopment: false)

        assertThat(nullScript.commonPipelineEnvironment.getValue('isChangeInDevelopment'), is(false))
    }

    @Test
    public void nullChangeDocumentIdTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("No changeDocumentId provided. Neither via parameter 'changeDocumentId' nor via " +
                             "label 'ChangeDocument\\s?:' in commit range [from: origin/master, to: HEAD].")

        def calledWithParameters,
            calledWithStepName

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                if(stepName.equals("transportRequestDocIDFromGit")) {
                    calledWithParameters = params
                }
            })

        stepRule.step.checkChangeInDevelopment(
            script: nullScript)

        assertThat(calledWithParameters,is(not(null)))
    }

    @Test
    public void emptyChangeDocumentIdTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("No changeDocumentId provided. Neither via parameter 'changeDocumentId' " +
                             "nor via label 'ChangeDocument\\s?:' in commit range " +
                             "[from: origin/master, to: HEAD].")

        def calledWithParameters,
            calledWithStepName

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                if(stepName.equals("transportRequestDocIDFromGit")) {
                    calledWithParameters = params
                    nullScript.commonPipelineEnvironment.setValue('changeDocumentId', '')
                }
            })

        stepRule.step.checkChangeInDevelopment(
            script: nullScript)

        assertThat(calledWithParameters,is(not(null)))
    }

    @Test
    public void cmIntegrationSwichtedOffTest() {

        loggingRule.expect('[INFO] Change management integration intentionally switched off.')

        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            changeManagement: [type: 'NONE'])

    }

    @Test
    public void stageConfigIsNotConsideredWithParamKeysTest() {

        nullScript.commonPipelineEnvironment.configuration.stages = [foo:[changeDocumentId:'12345']]

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('No changeDocumentId provided.')

        def calledWithParameters,
            calledWithStepName

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                if(stepName.equals("transportRequestDocIDFromGit")) {
                    calledWithParameters = params
                }
            })

        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            stageName: 'foo')

        assertThat(calledWithParameters,is(not(null)))
    }
}
