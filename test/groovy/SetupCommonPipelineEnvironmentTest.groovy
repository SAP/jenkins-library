import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml

import com.sap.piper.Utils
import com.sap.piper.DefaultValueCache

import util.BasePiperTest
import util.Rules
import util.JenkinsReadYamlRule
import util.JenkinsStepRule

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull


class SetupCommonPipelineEnvironmentTest extends BasePiperTest {

    def usedConfigFile

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(stepRule)
        .around(thrown)
        .around(new JenkinsReadYamlRule(this).registerYaml('.pipeline/config.yml', 'to_be_asserted: this_we_assert'))
        .around(new JenkinsFileExistsRule(this))

        def examplePipelineConfig = new File('test/resources/test_pipeline_config.yml').text

    @Test
    void testIsYamlConfigurationAvailable() throws Exception {

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript)

        assertNotNull(nullScript.commonPipelineEnvironment.configuration)
        assertEquals('this_we_assert', nullScript.commonPipelineEnvironment.configuration.to_be_asserted)
    }

    @Test
    void testCustomProjectConfigDoesNotExist() {
        thrown.expect(AbortException)
        thrown.expectMessage('Explicitly configured project config file \'.pipeline/myConfig.yml\' does not exist')

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript, projectConfig: '.pipeline/myConfig.yml')
    }
}

