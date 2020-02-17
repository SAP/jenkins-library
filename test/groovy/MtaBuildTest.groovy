import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.parser.ParserException

import com.sap.piper.PiperGoUtils

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsLoggingRule
import util.JenkinsReadJsonRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertThat

import java.util.List

import org.junit.After

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem

public class MtaBuildTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(thrown)
        .around(shellRule)
        .around(dockerExecuteRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(readJsonRule)

    private PiperGoUtils goUtils

    @Before
    void init() {

        goUtils = new PiperGoUtils(null) {
            void unstashPiperBin() {
            }

            public String prepareConfigurations(List configs, String configCacheFolder) {
                return ''
            }
        }

        nullScript.commonPipelineEnvironment.metaClass.readFromDisk = { def s -> }

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, "\\./piper.*--contextConfig", '{"only": "dummy"}')

        helper.registerAllowedMethod('withEnv', [List, Closure], {l, c -> env = l;  c()})
    }

    @After
    void tearDown() {
        nullScript.commonPipelineEnvironment.metaClass = null
    }

    @Test
    void callMtaPiperGo() {

        boolean piperBinUnstashed, configPrepared

        goUtils = new PiperGoUtils(null) {
            void unstashPiperBin() {
                piperBinUnstashed = true
            }

            public String prepareConfigurations(List configs, String configCacheFolder) {
                configPrepared = true
                return ''
            }
        }

        stepRule.step.mtaBuild(
            script: nullScript,
            piperGoUtils: goUtils)

        assert shellRule.shell[1].contains('./piper mtaBuild') // shell[0] is for resolving the config
        assert piperBinUnstashed
        assert configPrepared
    }

    @Test
    void callMtaPiperGoFailure() {

        thrown.expect(AbortException)
        thrown.expectMessage("mta build failed")

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, "\\./piper.*mtaBuild", { throw new AbortException("mta build failed.")})

        stepRule.step.mtaBuild(
            script: nullScript,
            piperGoUtils: goUtils)
    }

    @Test
    void dockerFromCustomStepConfigurationTest() {

        def expectedImage = 'image:test'
        def expectedEnvVars = ['env1': 'value1', 'env2': 'value2']
        def expectedOptions = '--opt1=val1 --opt2=val2 --opt3'
        def expectedWorkspace = '-w /path/to/workspace'

        nullScript.commonPipelineEnvironment.configuration = [steps:[mtaBuild:[
            dockerImage: expectedImage,
            dockerOptions: expectedOptions,
            dockerEnvVars: expectedEnvVars,
            dockerWorkspace: expectedWorkspace
            ]]]

        stepRule.step.mtaBuild(script: nullScript,
        piperGoUtils: goUtils)

        assert expectedImage == dockerExecuteRule.dockerParams.dockerImage
        assert expectedOptions == dockerExecuteRule.dockerParams.dockerOptions
        assert expectedEnvVars.equals(dockerExecuteRule.dockerParams.dockerEnvVars)
        assert expectedWorkspace == dockerExecuteRule.dockerParams.dockerWorkspace
    }

    @Test
    void canConfigureDockerImage() {

        stepRule.step.mtaBuild(script: nullScript, dockerImage: 'mta-docker-image:latest',
        piperGoUtils: goUtils)

        assert 'mta-docker-image:latest' == dockerExecuteRule.dockerParams.dockerImage
    }

    @Test
    void canConfigureDockerOptions() {

        stepRule.step.mtaBuild(script: nullScript, dockerOptions: 'something',
        piperGoUtils: goUtils)

        assert 'something' == dockerExecuteRule.dockerParams.dockerOptions
    }
}
