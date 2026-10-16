package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.anyOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

class PiperPipelineStageAdditionalUnitTestsTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)


    private List stepsCalled = []

    @Before
    void init()  {

        nullScript.env.STAGE_NAME = 'Additional Unit Tests'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Additional Unit Tests'))
            return body()
        })

        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], {m ->
            stepsCalled.add('npmExecuteScripts')
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
        })
    }

    @Test
    void testAdditionalUnitTestsDefault() {

        jsr.step.piperPipelineStageAdditionalUnitTests(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, not(anyOf(hasItem('npmExecuteScripts'), hasItem('testsPublishResults'))))
    }

    @Test
    void testAdditionalUnitTestsWithNpm() {

        nullScript.commonPipelineEnvironment.configuration = [runStep: ['Additional Unit Tests': [npmExecuteScripts: true]]]

        jsr.step.piperPipelineStageAdditionalUnitTests(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('npmExecuteScripts', 'testsPublishResults'))
    }
}
