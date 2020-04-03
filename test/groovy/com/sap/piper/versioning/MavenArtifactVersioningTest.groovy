package com.sap.piper.versioning

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsReadMavenPomRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals

class MavenArtifactVersioningTest extends BasePiperTest{

    Map dockerParameters
    def commonPipelineEnvironment

    MavenArtifactVersioning av
    String version = '1.2.3'

    JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(shellRule)
        .around(new JenkinsReadMavenPomRule(this, 'test/resources/versioning/MavenArtifactVersioning'))

    @Before
    void init() {
        dockerParameters = [:]

        helper.registerAllowedMethod("dockerExecute", [Map.class, Closure.class],
            { parameters, closure ->
                dockerParameters = parameters
                closure()
            })

        shellRule.setReturnValue("mvn --file 'pom.xml' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate -Dexpression=project.version -DforceStdout -q", version)
        shellRule.setReturnValue("mvn --file 'snapshot/pom.xml' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate -Dexpression=project.version -DforceStdout -q", version)
    }

    @Test
    void testVersioning() {
        av = new MavenArtifactVersioning(nullScript, [filePath: 'pom.xml'])
        assertEquals(version, av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('mvn --file \'pom.xml\' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn org.codehaus.mojo:versions-maven-plugin:2.7:set -DnewVersion=1.2.3-20180101 -DgenerateBackupPoms=false', shellRule.shell[1])
    }

    @Test
    void testVersioningCustomFilePathSnapshot() {
        av = new MavenArtifactVersioning(nullScript, [filePath: 'snapshot/pom.xml'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')
        assertEquals('mvn --file \'snapshot/pom.xml\' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn org.codehaus.mojo:versions-maven-plugin:2.7:set -DnewVersion=1.2.3-20180101 -DgenerateBackupPoms=false', shellRule.shell[1])
    }
}
