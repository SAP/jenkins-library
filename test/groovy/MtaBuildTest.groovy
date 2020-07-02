import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.parser.ParserException

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertThat
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem

public class MtaBuildTest extends BasePiperTest {

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

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'CF')

        assert calledWithParameters.size() == 2
        assert calledWithParameters.script == nullScript
        assert calledWithParameters.buildTarget == 'CF'

        assert calledWithStepName == 'mtaBuild'
	assert calledWithMetadata == 'metadata/mtaBuild.yaml'
        assert calledWithCredentials.isEmpty()

    }
}
