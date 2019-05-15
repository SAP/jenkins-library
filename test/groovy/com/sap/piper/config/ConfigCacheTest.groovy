package com.sap.piper.config

import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsLoggingRule
import util.Rules

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertThat

class ConfigCacheTest extends BasePiperTest {

    def loggingRule = new JenkinsLoggingRule(this)

    public ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(thrown)

    @Test
    void getPiperDefaultsTest() {
        def configCache = ConfigCache.getInstance(nullScript)
        assert configCache.getPiperDefaults() != null
        assertThat(loggingRule.log, containsString("Loading configuration file 'default_pipeline_environment.yml'"))
    }

    @Test
    void callParameterlessConstructorTest() {
        thrown.expect(NullPointerException)
        thrown.expectMessage('Steps not available.')
        new ConfigCache()
    }
}
