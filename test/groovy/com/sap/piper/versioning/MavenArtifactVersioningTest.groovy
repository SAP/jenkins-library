package com.sap.piper.versioning

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsMavenExecuteRule
import util.JenkinsReadYamlRule
import util.Rules

import static org.junit.Assert.assertEquals

class MavenArtifactVersioningTest extends BasePiperTest{

    Map dockerParameters
    def commonPipelineEnvironment

    MavenArtifactVersioning av
    String version = '1.2.3'

    JenkinsMavenExecuteRule mvnExecuteRule = new JenkinsMavenExecuteRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(mvnExecuteRule)

    @Before
    void init() {
        dockerParameters = [:]

        helper.registerAllowedMethod("dockerExecute", [Map.class, Closure.class],
            { parameters, closure ->
                dockerParameters = parameters
                closure()
            })

        mvnExecuteRule.setReturnValue([
            'pomPath': 'pom.xml',
            'goals': ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
            'defines': ['-Dexpression=project.version', '-DforceStdout', '-q'],
        ], version)

        mvnExecuteRule.setReturnValue([
            'pomPath': 'snapshot/pom.xml',
            'goals': ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
            'defines': ['-Dexpression=project.version', '-DforceStdout', '-q'],
        ], version)
    }

    @Test
    void testVersioning() {
        av = new MavenArtifactVersioning(nullScript, [filePath: 'pom.xml'])
        assertEquals(version, av.getVersion())
        av.setVersion('1.2.3-20180101')

        assertEquals(2, mvnExecuteRule.executions.size())
        assertEquals(new JenkinsMavenExecuteRule.Execution([
            pomPath: 'pom.xml',
            goals: ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
            defines: ['-Dexpression=project.version', '-DforceStdout', '-q']
        ]), mvnExecuteRule.executions[0])
        assertEquals(new JenkinsMavenExecuteRule.Execution([
            pomPath: 'pom.xml',
            goals: ['org.codehaus.mojo:versions-maven-plugin:2.7:set'],
            defines: ['-DnewVersion=1.2.3-20180101', '-DgenerateBackupPoms=false']
        ]), mvnExecuteRule.executions[1])
    }

    @Test
    void testVersioningCustomFilePathSnapshot() {
        av = new MavenArtifactVersioning(nullScript, [filePath: 'snapshot/pom.xml'])
        assertEquals('1.2.3', av.getVersion())
        av.setVersion('1.2.3-20180101')

        assertEquals(2, mvnExecuteRule.executions.size())
        assertEquals(new JenkinsMavenExecuteRule.Execution([
            pomPath: 'snapshot/pom.xml',
            goals: ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
            defines: ['-Dexpression=project.version', '-DforceStdout', '-q']
        ]), mvnExecuteRule.executions[0])
        assertEquals(new JenkinsMavenExecuteRule.Execution([
            pomPath: 'snapshot/pom.xml',
            goals: ['org.codehaus.mojo:versions-maven-plugin:2.7:set'],
            defines: ['-DnewVersion=1.2.3-20180101', '-DgenerateBackupPoms=false']
        ]), mvnExecuteRule.executions[1])
    }
}
