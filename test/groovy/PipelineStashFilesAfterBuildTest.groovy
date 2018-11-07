import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertThat

class PipelineStashFilesAfterBuildTest extends BasePiperTest {
    JenkinsStepRule jsr = new JenkinsStepRule(this)
    JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    JenkinsReadJsonRule jrj = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jrj)
        .around(jlr)
        .around(jsr)

    @Test
    void testStashAfterBuild() {
        helper.registerAllowedMethod("fileExists", [String.class], {
            searchTerm ->
                return false
        })
        jsr.step.pipelineStashFilesAfterBuild(
            script: nullScript,
            juStabUtils: utils
        )
        // asserts
        assertFalse(jlr.log.contains('Stash content: checkmarx'))
        assertThat(jlr.log, containsString('Stash content: classFiles'))
        assertThat(jlr.log, containsString('Stash content: sonar'))
    }

    @Test
    void testStashAfterBuildWithCheckmarx() {
        helper.registerAllowedMethod("fileExists", [String.class], {
            searchTerm ->
                return true
        })
        jsr.step.pipelineStashFilesAfterBuild(
            script: nullScript,
            juStabUtils: utils,
            runCheckmarx: true
        )
        // asserts
        assertThat(jlr.log, containsString('Stash content: checkmarx'))
        assertThat(jlr.log, containsString('Stash content: classFiles'))
        assertThat(jlr.log, containsString('Stash content: sonar'))
    }

    @Test
    void testStashAfterBuildWithCheckmarxConfig() {
        helper.registerAllowedMethod("fileExists", [String.class], {
            searchTerm ->
                return true
        })
        jsr.step.pipelineStashFilesAfterBuild(
            script: [commonPipelineEnvironment: [configuration: [steps: [executeCheckmarxScan: [checkmarxProject: 'TestProject']]]]],
            juStabUtils: utils,
        )
        // asserts
        assertThat(jlr.log, containsString('Stash content: checkmarx'))
        assertThat(jlr.log, containsString('Stash content: classFiles'))
        assertThat(jlr.log, containsString('Stash content: sonar'))
    }

}
