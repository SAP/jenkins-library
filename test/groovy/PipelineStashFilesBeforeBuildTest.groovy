import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.*

class PipelineStashFilesBeforeBuildTest extends BasePiperTest {
    JenkinsStepRule jsr = new JenkinsStepRule(this)
    JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    //JenkinsReadJsonRule jrj = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        //.around(jrj)
        .around(jlr)
        .around(jscr)
        .around(jsr)

    @Test
    void testStashBeforeBuildNoOpa() {

        jsr.step.pipelineStashFilesBeforeBuild(script: nullScript, juStabUtils: utils)

        // asserts
        assertEquals('mkdir -p gitmetadata', jscr.shell[0])
        assertEquals('cp -rf .git/* gitmetadata', jscr.shell[1])
        assertEquals('chmod -R u+w gitmetadata', jscr.shell[2])

        assertThat(jlr.log, containsString('Stash content: buildDescriptor'))
        assertThat(jlr.log, containsString('Stash content: deployDescriptor'))
        assertThat(jlr.log, containsString('Stash content: git'))
        assertFalse(jlr.log.contains('Stash content: opa5'))
        assertThat(jlr.log, containsString('Stash content: opensourceConfiguration'))
        assertThat(jlr.log, containsString('Stash content: pipelineConfigAndTests'))
        assertThat(jlr.log, containsString('Stash content: securityDescriptor'))
        assertThat(jlr.log, containsString('Stash content: tests'))
    }

    @Test
    void testStashBeforeBuildOpa() {

        jsr.step.pipelineStashFilesBeforeBuild(script: nullScript, juStabUtils: utils, runOpaTests: true)

        // asserts
        assertThat(jlr.log, containsString('Stash content: buildDescriptor'))
        assertThat(jlr.log, containsString('Stash content: deployDescriptor'))
        assertThat(jlr.log, containsString('Stash content: git'))
        assertThat(jlr.log, containsString('Stash content: opa5'))
        assertThat(jlr.log, containsString('Stash content: opensourceConfiguration'))
        assertThat(jlr.log, containsString('Stash content: pipelineConfigAndTests'))
        assertThat(jlr.log, containsString('Stash content: securityDescriptor'))
        assertThat(jlr.log, containsString('Stash content: tests'))
    }
}
