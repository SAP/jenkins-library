import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml

import com.lesfurets.jenkins.unit.BasePipelineTest

import util.Rules
import util.JenkinsStepRule
import util.JenkinsEnvironmentRule

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull

class SetupCommonPipelineEnvironmentTest extends BasePipelineTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jsr)
        .around(jer)

    def usedConfigFile

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
        jsr.step.call(script: [commonPipelineEnvironment: jer.env])

        assertEquals('.pipeline/config.yml', usedConfigFile)
        assertNotNull(jer.env.configuration)
        assertEquals('develop', jer.env.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', jer.env.configuration.steps.mavenExecute.dockerImage)
    }
}
