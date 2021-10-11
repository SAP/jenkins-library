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

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import static org.hamcrest.Matchers.hasEntry

public class isChangeInDevelopmentTest extends BasePiperTest {

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
            calledWithMetadata
        List calledWithCredentials

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

        stepRule.step.isChangeInDevelopment(script: nullScript, abc: 'CF')

        assertThat(calledWithParameters.size(), is(2))
        assertThat(calledWithParameters.script, is(nullScript))
        assertThat(calledWithParameters.abc, is('CF'))

        assertThat(calledWithStepName, is('isChangeInDevelopment'))
        assertThat(calledWithMetadata, is('metadata/isChangeInDevelopment.yaml'))
        assertThat(calledWithCredentials[0].size(), is(3))
        assertThat(calledWithCredentials[0], allOf(hasEntry('type','usernamePassword'), hasEntry('id','credentialsId'), hasEntry('env',['PIPER_username', 'PIPER_password'])))
    }
}
