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
    //JenkinsReadJsonRule jrj = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        //.around(jrj)
        .around(loggingRule)
        .around(shellRule)
        .around(stepRule)

    @Test
    void testStashBeforeBuildNoOpa() {

        stepRule.step.pipelineStashFilesBeforeBuild(script: nullScript, juStabUtils: utils)

        // asserts
        assertEquals('mkdir -p gitmetadata', shellRule.shell[0])
        assertEquals('cp -rf .git/* gitmetadata', shellRule.shell[1])
        assertEquals('chmod -R u+w gitmetadata', shellRule.shell[2])

        assertThat(loggingRule.log, containsString('Stash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: deployDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: git'))
        assertFalse(loggingRule.log.contains('Stash content: opa5'))
        assertThat(loggingRule.log, containsString('Stash content: opensourceConfiguration'))
        assertThat(loggingRule.log, containsString('Stash content: pipelineConfigAndTests'))
        assertThat(loggingRule.log, containsString('Stash content: securityDescriptor'))
        assertThat(loggingRule.log, containsString('Stash content: tests'))
    }

    @Test
    void testStashBeforeBuildOpa() {

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
