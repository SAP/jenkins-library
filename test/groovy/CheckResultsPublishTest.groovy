import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import com.lesfurets.jenkins.unit.BasePipelineTest

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

import util.JenkinsConfigRule
import util.JenkinsSetupRule

class CheckResultsPublishTest extends BasePipelineTest {
    Map publisherStepOptions

    @Rule
    public RuleChain ruleChain = RuleChain.outerRule(new JenkinsSetupRule(this)).around(new JenkinsConfigRule(this))

    def stepUnderTest

    @Before
    void init() {
        publisherStepOptions = [:]
        // prepare checkResultsPublish step
        stepUnderTest = loadScript("checkResultsPublish.groovy").checkResultsPublish
        // add handler for generic step call
        helper.registerAllowedMethod("step", [Map.class], {
            parameters -> publisherStepOptions[parameters.$class] = parameters
        })
    }

    @Test
    void testPublishWithDefaultSettings() throws Exception {
        stepUnderTest.call()
        println(publisherStepOptions)
        assert(publisherStepOptions['AnalysisPublisher'] != null)
        assertEquals('AnalysisPublisher', publisherStepOptions['AnalysisPublisher']['$class'])
        // ensure nothing else is published
        assert(publisherStepOptions['WarningsPublisher'] == null)
        assert(publisherStepOptions['PmdPublisher'] == null)
        assert(publisherStepOptions['DryPublisher'] == null)
        assert(publisherStepOptions['FindBugsPublisher'] == null)
        assert(publisherStepOptions['CheckStylePublisher'] == null)
    }
}
