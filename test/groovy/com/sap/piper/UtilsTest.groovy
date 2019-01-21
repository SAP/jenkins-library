package com.sap.piper

import org.junit.Rule
import org.junit.Before
import org.junit.Ignore
import org.junit.Test
import static org.junit.Assert.assertThat
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not

import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.BasePiperTest
import util.Rules

import com.sap.piper.Utils

class UtilsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jscr)
        .around(jlr)

    private parameters

    @Before
    void setup() {

        parameters = [:]
    }

    @Test
    void noValueGetMandatoryParameterTest() {

        thrown.expect(Exception)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR test")

        utils.getMandatoryParameter(parameters, 'test', null)
    }

    @Test
    void defaultValueGetMandatoryParameterTest() {

        assert  utils.getMandatoryParameter(parameters, 'test', 'default') == 'default'
    }

    @Test
    void valueGetmandatoryParameterTest() {

        parameters.put('test', 'value')

        assert utils.getMandatoryParameter(parameters, 'test', null) == 'value'
    }

    @Test
    @Ignore("replaced by TelemetryTest#testReportingToSWA")
    void testSWAReporting() {
        utils.env = [BUILD_URL: 'something', JOB_URL: 'nothing']
        utils.pushToSWA([step: 'anything'], [collectTelemetryData: true])
        // asserts
        assertThat(jscr.shell, hasItem(containsString('curl -G -v "https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log"')))
        assertThat(jscr.shell, hasItem(containsString('action_name=Piper Library OS')))
        assertThat(jscr.shell, hasItem(containsString('custom3=anything')))
        assertThat(jscr.shell, hasItem(containsString('custom5=`echo -n \'something\' | sha1sum | sed \'s/ -//\'`')))
    }

    @Test
    @Ignore("replaced by TelemetryTest#testNotifyWithOptOut")
    void testDisabledSWAReporting() {
        utils.env = [BUILD_URL: 'something', JOB_URL: 'nothing']
        utils.pushToSWA([step: 'anything'], [collectTelemetryData: false])
        // asserts
        assertThat(jlr.log, containsString('[anything] Telemetry reporting disabled!'))
        assertThat(jscr.shell, not(hasItem(containsString('https://webanalytics.cfapps.eu10.hana.ondemand.com'))))
    }

    @Test
    @Ignore("replaced by TelemetryTest#testNotifyWithOptOutWithoutConfig")
    void testImplicitlyDisabledSWAReporting() {
        utils.env = [BUILD_URL: 'something', JOB_URL: 'nothing']
        utils.pushToSWA([step: 'anything'], null)
        // asserts
        assertThat(jlr.log, containsString('[anything] Telemetry reporting disabled!'))
    }

    @Test
    @Ignore("replaced by TelemetryTest#testNotifyWithOptOutWithEmptyConfig")
    void testImplicitlyDisabledSWAReporting2() {
        utils.env = [BUILD_URL: 'something', JOB_URL: 'nothing']
        utils.pushToSWA([step: 'anything'], [:])
        // asserts
        assertThat(jlr.log, containsString('[anything] Telemetry reporting disabled!'))
    }

    @Test
    void testGenerateSHA1() {
        def result = utils.generateSha1('ContinuousDelivery')
        // asserts
        // generated with "echo -n 'ContinuousDelivery' | sha1sum | sed 's/  -//'"
        assertThat(result, is('0dad6c33b6246702132454f604dee80740f399ad'))
    }
}
