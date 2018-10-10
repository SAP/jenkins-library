package com.sap.piper.versioning

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadFileRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

class DockerArtifactVersioningTest extends BasePiperTest{

    DockerArtifactVersioning av

    String passedDir

    JenkinsReadFileRule jrfr = new JenkinsReadFileRule(this, 'test/resources/versioning/DockerArtifactVersioning')
    JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jrfr)
        .around(jwfr)
        .around(jlr)
        .around(thrown)

    @Before
    public void init() {

        helper.registerAllowedMethod("dir", [String.class, Closure.class], { s, closure ->
            passedDir = s
            return closure()
        })
    }

    @Test
    void testVersioningFrom() {
        av = new DockerArtifactVersioning(nullScript, [filePath: 'Dockerfile', dockerVersionSource: 'FROM'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('1.2.3-20180101', jwfr.files['VERSION'])
        assertTrue(jlr.log.contains('[DockerArtifactVersioning] Version from Docker base image tag: 1.2.3'))
    }

    @Test
    void testVersioningEnv() {
        av = new DockerArtifactVersioning(nullScript, [filePath: 'Dockerfile', dockerVersionSource: 'TEST'])
        assertEquals('2.3.4', av.getVersion())
        assertTrue(jlr.log.contains('[DockerArtifactVersioning] Version from Docker environment variable TEST: 2.3.4'))
    }
}
