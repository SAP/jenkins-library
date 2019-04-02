import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.*

class PipelineStashFilesBeforeBuildTest extends BasePiperTest {
    JenkinsStepRule stepRule = new JenkinsStepRule(this)
    JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    //JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        //.around(readJsonRule)
        .around(loggingRule)
        .around(shellRule)
        .around(stepRule)

    @Test
    void testStashBeforeBuildNoOpa() {

        stepRule.step.pipelineStashFilesBeforeBuild(script: nullScript, juStabUtils: utils)

        assertThat(loggingRule.log, containsString('Stash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: deployDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: git (include: .git/**, exclude: , useDefaultExcludes: false'))
        //assertFalse(loggingRule.log.contains('Stash content: opa5'))
        assertThat(loggingRule.log, containsString('Stash content: opensourceConfiguration'))
        assertThat(loggingRule.log, containsString('Stash content: pipelineConfigAndTests (include: .pipeline/**, exclude: , useDefaultExcludes: true'))
        assertThat(loggingRule.log, containsString('Stash content: securityDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: tests'))
    }

    @Test
    void testStashBeforeBuild() {

        stepRule.step.pipelineStashFilesBeforeBuild(script: nullScript, juStabUtils: utils, runOpaTests: true)

        // asserts
        assertThat(loggingRule.log, containsString('Stash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: deployDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: git'))
        assertThat(loggingRule.log, containsString('Stash content: opa5'))
        assertThat(loggingRule.log, containsString('Stash content: opensourceConfiguration'))
        assertThat(loggingRule.log, containsString('Stash content: pipelineConfigAndTests'))
        assertThat(loggingRule.log, containsString('Stash content: securityDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: tests'))
    }
}
