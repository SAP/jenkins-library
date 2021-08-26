import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.junit.Assert.assertThat
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.allOf

public class CnbBuildTest extends BasePiperTest {

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

        stepRule.step.cnbBuild(script: nullScript, containerImage: 'foo:bar', dockerConfigJsonCredentialsId: 'DOCKER_CREDENTIALS')

        assertThat(calledWithParameters.size(), is(3))
        assertThat(calledWithParameters.script, is(nullScript))
        assertThat(calledWithParameters.containerImage, is('foo:bar'))
        assertThat(calledWithParameters.dockerConfigJsonCredentialsId, is('DOCKER_CREDENTIALS'))

        assertThat(calledWithStepName, is('cnbBuild'))
        assertThat(calledWithMetadata, is('metadata/cnbBuild.yaml'))

        assertThat(calledWithCredentials.size(), is(1))
        assertThat(calledWithCredentials[0].size(), is(3))
        assertThat(calledWithCredentials[0], allOf(hasEntry('type','file'), hasEntry('id','dockerConfigJsonCredentialsId'), hasEntry('env',['PIPER_dockerConfigJSON'])))

    }
}
