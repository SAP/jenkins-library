import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertThat

class PipelineStashFilesAfterBuildTest extends BasePiperTest {
    JenkinsStepRule stepRule = new JenkinsStepRule(this)
    JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    JenkinsReadJsonRule jrj = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jrj)
        .around(loggingRule)
        .around(stepRule)

    @Test
    void testStashAfterBuild() {
        helper.registerAllowedMethod("fileExists", [String.class], {
            searchTerm ->
                return false
        })
        stepRule.step.pipelineStashFilesAfterBuild(
            script: nullScript,
            juStabUtils: utils
        )
        // asserts
        assertFalse(loggingRule.log.contains('Stash content: checkmarx'))
        assertThat(loggingRule.log, containsString('Stash content: classFiles'))
        assertThat(loggingRule.log, containsString('Stash content: sonar'))
    }

    @Test
    void testStashAfterBuildWithCheckmarx() {
        helper.registerAllowedMethod("fileExists", [String.class], {
            searchTerm ->
                return true
        })
        stepRule.step.pipelineStashFilesAfterBuild(
            script: nullScript,
            juStabUtils: utils,
            runCheckmarx: true
        )
        // asserts
        assertThat(loggingRule.log, containsString('Stash content: checkmarx'))
        assertThat(loggingRule.log, containsString('Stash content: classFiles'))
        assertThat(loggingRule.log, containsString('Stash content: sonar'))
    }

    @Test
    void testStashAfterBuildWithCheckmarxConfig() {
        helper.registerAllowedMethod("fileExists", [String.class], {
            searchTerm ->
                return true
        })
        stepRule.step.pipelineStashFilesAfterBuild(
            script: [commonPipelineEnvironment: [configuration: [steps: [executeCheckmarxScan: [checkmarxProject: 'TestProject']]]]],
            juStabUtils: utils,
        )
        // asserts
        assertThat(loggingRule.log, containsString('Stash content: checkmarx'))
        assertThat(loggingRule.log, containsString('Stash content: classFiles'))
        assertThat(loggingRule.log, containsString('Stash content: sonar'))
    }

}
