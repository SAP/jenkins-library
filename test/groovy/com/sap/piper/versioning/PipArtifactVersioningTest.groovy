package com.sap.piper.versioning

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadFileRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertEquals

class PipArtifactVersioningTest extends BasePiperTest{

    JenkinsReadFileRule jrfr = new JenkinsReadFileRule(this, 'test/resources/versioning/PipArtifactVersioning/')
    JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jrfr)
        .around(jwfr)

    @Test
    void testVersioning() {
        PipArtifactVersioning av = new PipArtifactVersioning(nullScript, [filePath: 'version.txt'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('1.2.3-20180101', jwfr.files['version.txt'])
    }
}
