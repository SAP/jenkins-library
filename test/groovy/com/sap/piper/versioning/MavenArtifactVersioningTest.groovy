package com.sap.piper.versioning

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadMavenPomRule
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals

class MavenArtifactVersioningTest extends BasePiperTest{

    Map dockerParameters
    def commonPipelineEnvironment

    MavenArtifactVersioning av

    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jscr)
        .around(new JenkinsReadMavenPomRule(this, 'test/resources/MavenArtifactVersioning'))

    @Before
    void init() {
        dockerParameters = [:]

        helper.registerAllowedMethod("dockerExecute", [Map.class, Closure.class],
            { parameters, closure ->
                dockerParameters = parameters
                closure()
            })
    }

    @Test
    void testVersioning() {
        av = new MavenArtifactVersioning(nullScript, [filePath: 'pom.xml'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('mvn --file \'pom.xml\' versions:set -DnewVersion=1.2.3-20180101', jscr.shell[0])
    }

    @Test
    void testVersioningCustomFilePathSnapshot() {
        av = new MavenArtifactVersioning(nullScript, [filePath: 'snapshot/pom.xml'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('mvn --file \'snapshot/pom.xml\' versions:set -DnewVersion=1.2.3-20180101', jscr.shell[0])
    }
}
