package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadFileRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertThat

class WhiteSourceConfigurationHelperTest extends BasePiperTest {
    JenkinsReadFileRule jrfr = new JenkinsReadFileRule(this, 'test/resources/utilsTest/')
    JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jrfr)
        .around(jwfr)

    @Before
    void init() {
        helper.registerAllowedMethod('readProperties', [Map], {return new Properties()})
    }

    @Test
    void testExtendConfigurationFileUnifiedAgentPip() {
        WhitesourceConfigurationHelper.extendUAConfigurationFile(nullScript, utils, [scanType: 'pip', configFilePath: './config', orgToken: 'abcd', productName: 'name', productToken: '1234', userKey: '0000'], "./")
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("apiKey=abcd"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("productName=name"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("productToken=1234"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("userKey=0000"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("python.resolveDependencies=true"))
    }

    @Test
    void testExtendConfigurationFileUnifiedAgentVerbose() {
        WhitesourceConfigurationHelper.extendUAConfigurationFile(nullScript, utils, [scanType: 'pip', verbose: true, configFilePath: './config', orgToken: 'abcd', productName: 'name', productToken: '1234', userKey: '0000'], "./")
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("apiKey=abcd"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("productName=name"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("productToken=1234"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("userKey=0000"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("python.resolveDependencies=true"))
        assertThat(jwfr.files['./config.847f9aec2f93de9000d5fa4e6eaace2283ae6377'], containsString("log.level=debug"))
    }
}

