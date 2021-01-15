import static org.hamcrest.Matchers.*

import com.sap.piper.Utils

import hudson.AbortException

import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.junit.rules.ExpectedException

import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

class UiVeri5ExecuteTestsTest extends BasePiperTest {

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(stepRule)
        .around(readYamlRule)

    @Test
    void testCallGoWrapper() {

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod(
            'piperExecuteBin',
            [Map, String, String, List],
            {
                params, stepName, metaData, creds ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
            }
        )
        helper.registerAllowedMethod('dockerExecute', [Map, Closure], null)

        stepRule.step.uiVeri5ExecuteTests(script: nullScript)

        assert calledWithParameters.size() == 1
        assert calledWithParameters.script == nullScript

        assert calledWithStepName == 'uiVeri5ExecuteTests'
        assert calledWithMetadata == 'metadata/uiVeri5ExecuteTests.yaml'
        assert calledWithCredentials.size == 1

    }
}
