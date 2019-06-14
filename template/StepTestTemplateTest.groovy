import com.lesfurets.jenkins.unit.BasePipelineTest

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsStepRule

class StepTestTemplateTest extends BasePipelineTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jsr)

    @Before
    void init() throws Exception {
    }

    @Test
    void testStepTestTemplate() throws Exception {
        jsr.step.stepTestTemplate()
        // asserts
        assertTrue(true)
        assertJobStatusSuccess()
    }
}
