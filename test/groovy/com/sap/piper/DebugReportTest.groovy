package com.sap.piper

import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.Rules

class DebugReportTest extends BasePiperTest {

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(new JenkinsLoggingRule(this))

    @Test
    void testInitFromEnvironment() {
        Map env = createEnv()
        DebugReport.instance.initFromEnvironment(env)

        Assert.assertTrue(DebugReport.instance.environment.containsKey('build_details'))
        Assert.assertEquals('Kubernetes', DebugReport.instance.environment.get('environment'))

        Set<String> buildDetails = DebugReport.instance.environment.build_details as Set<String>
        Assert.assertTrue(buildDetails.size() > 0)

        boolean foundJenkinsVersion = false
        boolean foundJavaVersion = false

        for (String details in buildDetails) {
            if (details.contains(env.get('JENKINS_VERSION') as String))
                foundJenkinsVersion = true
            if (details.contains(env.get('JAVA_VERSION') as String))
                foundJavaVersion = true
        }
        Assert.assertTrue(foundJenkinsVersion)
        Assert.assertTrue(foundJavaVersion)
    }

    @Test
    void testLogOutput() {
        DebugReport.instance.initFromEnvironment(createEnv())
        DebugReport.instance.setGitRepoInfo('GIT_URL' : 'git://url', 'GIT_LOCAL_BRANCH' : 'some-branch')

        String debugReport = DebugReport.instance.generateReport(mockScript(), false)

        Assert.assertTrue(debugReport.contains('## Pipeline Environment'))
        Assert.assertTrue(debugReport.contains('## Local Extensions'))
        Assert.assertTrue(debugReport.contains('#### Environment\n' +
            '`Kubernetes`'))
        Assert.assertFalse(debugReport.contains('Repository | Branch'))
        Assert.assertFalse(debugReport.contains('some-branch'))
    }

    @Test
    void testLogOutputConfidential() {
        DebugReport.instance.initFromEnvironment(createEnv())
        DebugReport.instance.setGitRepoInfo('GIT_URL' : 'git://url', 'GIT_LOCAL_BRANCH' : 'some-branch')

        String debugReport = DebugReport.instance.generateReport(mockScript(), true)

        Assert.assertTrue(debugReport.contains('## Pipeline Environment'))
        Assert.assertTrue(debugReport.contains('## Local Extensions'))
        Assert.assertTrue(debugReport.contains('#### Environment\n' +
            '`Kubernetes`'))
        Assert.assertTrue(debugReport.contains('Repository | Branch'))
        Assert.assertTrue(debugReport.contains('some-branch'))
    }

    private Script mockScript() {
        helper.registerAllowedMethod("libraryResource", [String.class], { path ->

            File resource = new File(new File('resources'), path)
            if (resource.exists()) {
                return resource.getText()
            }

            return ''
        })

        return nullScript
    }

    private static Map createEnv() {
        Map env = [:]
        env.put('JENKINS_VERSION', '42')
        env.put('JAVA_VERSION', '8')
        env.put('ON_K8S', 'true')
        return env
    }
}
