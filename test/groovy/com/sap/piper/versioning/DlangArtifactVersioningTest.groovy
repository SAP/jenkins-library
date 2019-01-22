package com.sap.piper.versioning

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadJsonRule
import util.JenkinsWriteJsonRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

class DlangArtifactVersioningTest extends BasePiperTest{

    JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this, 'test/resources/versioning/DlangArtifactVersioning/')
    JenkinsWriteJsonRule jwjr = new JenkinsWriteJsonRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(readJsonRule)
        .around(jwjr)

    @Test
    void testVersioning() {
        DlangArtifactVersioning av = new DlangArtifactVersioning(nullScript, [filePath: 'dub.json'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertTrue(jwjr.files['dub.json'].contains('1.2.3-20180101'))
    }
}
