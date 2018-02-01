package com.sap.piper.versioning

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.JenkinsReadMavenPomRule
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

class MavenArtifactVersioningTest extends BasePipelineTest{

    MavenArtifactVersioning av

    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this).around(jscr).around(thrown).around(new JenkinsReadMavenPomRule(this, 'test/resources/MavenArtifactVersioning'))

    @Before
    public void init() {
        prepareObjectInterceptors(this)
    }

    @Test
    void testVersioning() {
        av = new MavenArtifactVersioning(this, [filePath: 'pom.xml'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('mvn versions:set -DnewVersion=1.2.3-20180101', jscr.shell[0])
    }


    @Test
    void testVersioningCustomFilePathSnapshot() {
        av = new MavenArtifactVersioning(this, [filePath: 'snapshot/pom.xml'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('mvn versions:set -DnewVersion=1.2.3-20180101 --file snapshot/pom.xml', jscr.shell[0])
    }


    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }
}
