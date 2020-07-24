import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*
import com.sap.piper.Utils

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertThat


class PipelineStashFilesAfterBuildTest extends BasePiperTest {
    JenkinsStepRule stepRule = new JenkinsStepRule(this)
    JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(readJsonRule)
        .around(loggingRule)
        .around(stepRule)

    @Before
    public void setup() {
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testStashAfterBuild() {
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
}
