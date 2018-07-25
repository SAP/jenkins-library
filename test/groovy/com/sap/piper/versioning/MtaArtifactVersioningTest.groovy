package com.sap.piper.versioning

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals

class MtaArtifactVersioningTest extends BasePiperTest{

    JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this, 'test/resources/versioning/MtaArtifactVersioning/')
    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jryr)
        .around(jscr)

    @Test
    void testVersioning() {
        MtaArtifactVersioning av = new MtaArtifactVersioning(nullScript, [filePath: 'mta.yaml'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals("sed -i 's/version: 1.2.3/version: 1.2.3-20180101/g' mta.yaml", jscr.shell[0])
    }
}
