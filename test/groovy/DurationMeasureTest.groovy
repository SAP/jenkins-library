import com.sap.piper.analytics.InfluxData

import org.junit.Rule
import org.junit.Test
import util.BasePiperTest

import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not

import static org.junit.Assert.assertThat
import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsReadYamlRule
import util.JenkinsStepRule


class DurationMeasureTest extends BasePiperTest {
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(stepRule)

    @Test
    void testDurationMeasurement() throws Exception {
        def bodyExecuted = false
        stepRule.step.durationMeasure(script: nullScript, measurementName: 'test') {
            bodyExecuted = true
        }
        // doesnt work
        //assertThat(InfluxData.getInstance().getFields(), hasEntry('pipeline_data', hasEntry('test', is(anything()))))
        assertThat(InfluxData.getInstance().getFields(), hasKey('pipeline_data'))
        assertThat(InfluxData.getInstance().getFields().pipeline_data, hasKey('test'))
        assertThat(InfluxData.getInstance().getFields().pipeline_data.test, is(not(null)))
        assertThat(bodyExecuted, is(true))
        assertJobStatusSuccess()
    }
}
