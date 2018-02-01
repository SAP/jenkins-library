package com.sap.piper.versioning

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.JenkinsReadFileRule
import util.JenkinsReadMavenPomRule
import util.JenkinsShellCallRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertEquals

class DockerArtifactVersioningTest extends BasePipelineTest{

    DockerArtifactVersioning av

    String passedDir

    JenkinsReadFileRule jrfr = new JenkinsReadFileRule(this, 'test/resources/DockerArtifactVersioning')
    JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jrfr)
        .around(jwfr)
        .around(thrown)

    @Before
    public void init() {

        helper.registerAllowedMethod("dir", [String.class, Closure.class], { s, closure ->
            passedDir = s
            return closure()
        })

        prepareObjectInterceptors(this)
    }

    @Test
    void testVersioningFrom() {
        av = new DockerArtifactVersioning(this, [filePath: 'Dockerfile', dockerVersionSource: 'FROM'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('1.2.3-20180101', jwfr.files['VERSION'])
    }

    @Test
    void testVersioningEnv() {
        av = new DockerArtifactVersioning(this, [filePath: 'Dockerfile', dockerVersionSource: 'TEST'])
        assertEquals('2.3.4', av.getVersion())
    }


    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }
}
