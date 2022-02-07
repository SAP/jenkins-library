import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
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
            calledWithCredentials,
            calledWithFailOnError

        helper.registerAllowedMethod(
            'piperExecuteBin',
            [Map, String, String, List, Boolean, Boolean, Boolean],
            {
                params, stepName, metaData, creds, failOnMissingReports, failOnMissingLinks, failOnError  ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
                calledWithFailOnError = failOnError
            }
        )

        stepRule.step.cnbBuild(
            script: nullScript,
            buildpacks: ['test1', 'test2'],
            containerImageName: 'foo',
            containerImageTag: 'bar',
            containerRegistryUrl: 'test',
            dockerConfigJsonCredentialsId: 'DOCKER_CREDENTIALS',
            buildEnvVars: ['foo=bar', 'bar=baz']
        )

        assertThat(calledWithParameters.size(), is(7))
        assertThat(calledWithParameters.script, is(nullScript))
        assertThat(calledWithParameters.buildpacks, is(['test1', 'test2']))
        assertThat(calledWithParameters.containerImageName, is('foo'))
        assertThat(calledWithParameters.containerImageTag, is('bar'))
        assertThat(calledWithParameters.containerRegistryUrl, is('test'))
        assertThat(calledWithParameters.dockerConfigJsonCredentialsId, is('DOCKER_CREDENTIALS'))
        assertThat(calledWithParameters.buildEnvVars, is(['foo=bar', 'bar=baz']))

        assertThat(calledWithStepName, is('cnbBuild'))
        assertThat(calledWithMetadata, is('metadata/cnbBuild.yaml'))

        assertThat(calledWithCredentials.size(), is(1))
        assertThat(calledWithCredentials[0].size(), is(3))
        assertThat(calledWithCredentials[0], allOf(hasEntry('type','file'), hasEntry('id','dockerConfigJsonCredentialsId'), hasEntry('env',['PIPER_dockerConfigJSON'])))

        assertTrue(calledWithFailOnError)

    }
}
