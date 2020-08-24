package com.sap.piper.analytics

import org.junit.Rule
import org.junit.Before
import org.junit.Test
import static org.junit.Assert.assertThat
import static org.junit.Assume.assumeThat
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.empty
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.hasEntry

import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsReadFileRule
import util.BasePiperTest
import util.Rules

class InfluxDataTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, null)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jscr)
        .around(jlr)
        .around(readFileRule)

    @Before
    void setup() {
        InfluxData.instance = null
    }

    @Test
    void testCreateInstance() {
        InfluxData.getInstance()
        // asserts
        assertThat(InfluxData.instance.fields, allOf(
            is(not(null)),
            hasKey('jenkins_custom_data'),
            hasKey('pipeline_data'),
            hasKey('step_data')
        ))
        assertThat(InfluxData.instance.fields.jenkins_custom_data, is([:]))
        assertThat(InfluxData.instance.fields.pipeline_data, is([:]))
        assertThat(InfluxData.instance.fields.step_data, is([:]))
        assertThat(InfluxData.instance.tags, allOf(
            is(not(null)),
            hasKey('jenkins_custom_data'),
            hasKey('pipeline_data'),
            hasKey('step_data')
        ))
        assertThat(InfluxData.instance.tags.jenkins_custom_data, is([:]))
        assertThat(InfluxData.instance.tags.pipeline_data, is([:]))
        assertThat(InfluxData.instance.tags.step_data, is([:]))
    }

    @Test
    void testAddToDefaultMeasurement() {
        InfluxData.addField('step_data', 'anyKey', 'anyValue')
        InfluxData.addTag('step_data', 'anyKey', 'anyTag')
        // asserts
        assertThat(InfluxData.instance.fields.jenkins_custom_data, is([:]))
        assertThat(InfluxData.instance.fields.pipeline_data, is([:]))
        assertThat(InfluxData.instance.fields.step_data, is(['anyKey': 'anyValue']))
        assertThat(InfluxData.instance.tags.jenkins_custom_data, is([:]))
        assertThat(InfluxData.instance.tags.pipeline_data, is([:]))
        assertThat(InfluxData.instance.tags.step_data, is(['anyKey': 'anyTag']))
    }

    @Test
    void testAddToNewMeasurement() {
        InfluxData.addField('new_measurement_data', 'anyKey', 'anyValue')
        InfluxData.addTag('new_measurement_data', 'anyKey', 'anyTag')
        // asserts
        assertThat(InfluxData.instance.fields.new_measurement_data, is(['anyKey': 'anyValue']))
        assertThat(InfluxData.instance.fields.jenkins_custom_data, is([:]))
        assertThat(InfluxData.instance.fields.pipeline_data, is([:]))
        assertThat(InfluxData.instance.fields.step_data, is([:]))
        assertThat(InfluxData.instance.tags.new_measurement_data, is(['anyKey': 'anyTag']))
        assertThat(InfluxData.instance.tags.jenkins_custom_data, is([:]))
        assertThat(InfluxData.instance.tags.pipeline_data, is([:]))
        assertThat(InfluxData.instance.tags.step_data, is([:]))
    }

    @Test
    void testResetInstance() {
        InfluxData.addField('step_data', 'anyKey', 'anyValue')
        assumeThat(InfluxData.instance.fields.step_data, is(['anyKey': 'anyValue']))
        InfluxData.reset()
        // asserts
        assertThat(InfluxData.instance.fields.jenkins_custom_data, is([:]))
        assertThat(InfluxData.instance.fields.pipeline_data, is([:]))
        assertThat(InfluxData.instance.fields.step_data, is([:]))
    }

    @Test
    void testReadFromDisk() {
        // init
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            if(map.glob == '.pipeline/influx/**')
                return [
                    new File(".pipeline/influx/step_data/fields/sonar"),
                    new File(".pipeline/influx/step_data/fields/protecode"),
                    new File("cst/test2.yml"),
                ].toArray()
            return [].toArray()
        })
        readFileRule.files.putAll([
            '.pipeline/influx/step_data/fields/sonar': 'true',
            '.pipeline/influx/step_data/fields/protecode': 'false',
        ])

        // tests
        InfluxData.readFromDisk(nullScript)
        // asserts
        assertThat(InfluxData.instance.fields.step_data, is([sonar: true, protecode: false]))
    }
}
