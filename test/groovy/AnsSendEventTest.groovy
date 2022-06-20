import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

public class AnsSendEventTest extends BasePiperTest {

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

        stepRule.step.ansSendEvent(script: nullScript, abc: 'ABC')

        assertThat(calledWithParameters.size(), is(2))
        assertThat(calledWithParameters.script, is(nullScript))
        assertThat(calledWithParameters.abc, is('ABC'))

        assertThat(calledWithStepName, is('ansSendEvent'))
        assertThat(calledWithMetadata, is('metadata/ansSendEvent.yaml'))
        assertThat(calledWithCredentials[0].size(), is(3))
        assertThat(calledWithCredentials[0], allOf(hasEntry('type', 'token'), hasEntry('id', 'ansServiceKeyCredentialsId'), hasEntry('env', ['PIPER_ansServiceKey'])))
    }
}
