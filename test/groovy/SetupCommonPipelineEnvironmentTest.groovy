import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml

import util.BasePiperTest
import util.Rules
import util.JenkinsStepRule

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull

class SetupCommonPipelineEnvironmentTest extends BasePiperTest {
    def usedConfigFile

    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jsr)

    @Before
    void init() {
        def examplePipelineConfig = new File('test/resources/test_pipeline_config.yml').text

        helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
            Yaml yamlParser = new Yaml()
            if(parameters.text) {
                return yamlParser.load(parameters.text)
            }
            usedConfigFile = parameters.file
            return yamlParser.load(examplePipelineConfig)
        })
        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })
    }

    @Test
    void testIsConfigurationAvailable() throws Exception {
        jsr.step.call(script: nullScript)

        assertEquals('.pipeline/config.yml', usedConfigFile)
        assertNotNull(nullScript.commonPipelineEnvironment.configuration)
        assertEquals('develop', nullScript.commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', nullScript.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
    }
}
