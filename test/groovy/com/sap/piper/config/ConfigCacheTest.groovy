package com.sap.piper.config

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.Rules

class ConfigCacheTest extends BasePiperTest {

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))

    @Test
    void getPiperDefaultsTest() {
        def configCache = ConfigCache.getInstance(nullScript)
        assert configCache.getPiperDefaults() != null
    }
}
