import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import com.sap.piper.Utils

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.*

class PipelineStashFilesBeforeBuildTest extends BasePiperTest {
    JenkinsStepRule stepRule = new JenkinsStepRule(this)
    JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    //JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)

    @Before
    public void setup() {
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        //.around(readJsonRule)
        .around(loggingRule)
        .around(shellRule)
        .around(stepRule)

    @Test
    void testStashBeforeBuild() {

        stepRule.step.pipelineStashFilesBeforeBuild(script: nullScript, juStabUtils: utils, runOpaTests: true)

        // asserts
        assertThat(loggingRule.log, containsString('Stash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: deployDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: git'))
        assertThat(loggingRule.log, containsString('Stash content: opensourceConfiguration'))
        assertThat(loggingRule.log, containsString('Stash content: pipelineConfigAndTests'))
        assertThat(loggingRule.log, containsString('Stash content: securityDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: tests'))
    }

    @Test
    void testStashBeforeBuildCustomConfig() {

        stepRule.step.pipelineStashFilesBeforeBuild(script: nullScript, juStabUtils: utils, runOpaTests: true, stashIncludes: ['myStash': '**.myTest'])

        // asserts
        assertThat(loggingRule.log, containsString('Stash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: deployDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: git'))
        assertThat(loggingRule.log, containsString('Stash content: opensourceConfiguration'))
        assertThat(loggingRule.log, containsString('Stash content: pipelineConfigAndTests'))
        assertThat(loggingRule.log, containsString('Stash content: securityDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: tests'))
        assertThat(loggingRule.log, containsString('Stash content: myStash'))
    }
}
