import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import com.sap.piper.DefaultValueCache

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

public class PrepareDefaultValuesTest extends BasePiperTest {

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)


    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(writeFileRule)
        .around(thrown)
        .around(stepRule)
        .around(loggingRule)

    @Before
    public void setup() {

        helper.registerAllowedMethod("readYaml", [Map], {  Map m ->
            def yml
            if(m.text) {
                return m.text
            } else if(m.file) {
                if(m.file == ".pipeline/default_pipeline_environment.yml") return [default: 'config']
                else if (m.file == ".pipeline/custom.yml") return [custom: 'myConfig']
            } else {
                throw new IllegalArgumentException("Key 'text' and 'file' are both missing in map ${m}.")
            }
        })

        helper.registerAllowedMethod("libraryResource", [String], { fileName ->
            if(fileName == 'default_pipeline_environment.yml') {
                return "default: 'config'"
            }
        })

    }

    @Test
    public void testDefaultPipelineEnvironmentOnly() {

        stepRule.step.prepareDefaultValues(script: nullScript)

        assert DefaultValueCache.getInstance().getDefaultValues().size() == 1
        assert DefaultValueCache.getInstance().getDefaultValues().default == 'config'
    }

    @Test
    public void testReInitializeOnCustomConfig() {

        def instance = DefaultValueCache.createInstance([key:'value'])

        // existing instance is dropped in case a custom config is provided.
        stepRule.step.prepareDefaultValues(script: nullScript, customDefaults: ['default_pipeline_environment.yml','custom.yml'])

        // this check is for checking we have another instance
        assert ! instance.is(DefaultValueCache.getInstance())

        // some additional checks that the configuration represented by the new
        // config is fine
        assert DefaultValueCache.getInstance().getDefaultValues().size() == 2
        assert DefaultValueCache.getInstance().getDefaultValues().default == 'config'
        assert DefaultValueCache.getInstance().getDefaultValues().custom == 'myConfig'
    }

    @Test
    public void testNoReInitializeWithoutCustomConfig() {

        def instance = DefaultValueCache.createInstance([key:'value'])

        stepRule.step.prepareDefaultValues(script: nullScript)

        assert instance.is(DefaultValueCache.getInstance())
        assert DefaultValueCache.getInstance().getDefaultValues().size() == 1
        assert DefaultValueCache.getInstance().getDefaultValues().key == 'value'
    }

    @Test
    public void testDefaultPipelineEnvironmentWithCustomConfigReferencedAsList() {

        stepRule.step.prepareDefaultValues(script: nullScript, customDefaults: ['default_pipeline_environment.yml','custom.yml'])

        assert DefaultValueCache.getInstance().getDefaultValues().size() == 2
        assert DefaultValueCache.getInstance().getDefaultValues().default == 'config'
        assert DefaultValueCache.getInstance().getDefaultValues().custom == 'myConfig'
    }

    @Test
    public void testAssertNoLogMessageInCaseOfNoAdditionalConfigFiles() {

        stepRule.step.prepareDefaultValues(script: nullScript)

        assert ! loggingRule.log.contains("Loading configuration file 'default_pipeline_environment.yml'")
    }

    @Test
    public void testAssertLogMessageInCaseOfMoreThanOneConfigFile() {

        stepRule.step.prepareDefaultValues(script: nullScript, customDefaults: ['default_pipeline_environment.yml','custom.yml'])

        assert loggingRule.log.contains("Loading configuration file 'default_pipeline_environment.yml'")
        assert loggingRule.log.contains("Loading configuration file 'custom.yml'")
    }
}
