package com.sap.piper.versioning

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadFileRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertEquals

class GolangArtifactVersioningTest extends BasePiperTest{

    JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, 'test/resources/versioning/GolangArtifactVersioning/')
    JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(readFileRule)
        .around(writeFileRule)

    @Test
    void testVersioning() {
        GolangArtifactVersioning av = new GolangArtifactVersioning(nullScript, [filePath: 'VERSION'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('1.2.3-20180101', writeFileRule.files['VERSION'])
    }
}
