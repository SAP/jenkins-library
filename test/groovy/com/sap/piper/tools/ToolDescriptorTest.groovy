package com.sap.piper.tools

import org.junit.BeforeClass
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.JenkinsLoggingRule
import util.Rules

import com.lesfurets.jenkins.unit.BasePipelineTest

import com.sap.piper.tools.ToolDescriptor

import hudson.AbortException


class ToolDescriptorTest extends BasePipelineTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
                .around(thrown)
                .around(jlr)

    private static tool
    private static configuration

    private script


    @BeforeClass
    static void init() {

        tool = new ToolDescriptor('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', '3.39.10', 'version')
    }

    @Before
    void setup() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoEnvVars(m) })

        script = loadScript('neoDeploy.groovy').neoDeploy

        configuration = [:]
    }

    @Test
    void getToolHomeFromEnvironmentTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVars(m) })

        def toolHome = tool.getToolLocation(script, configuration)

        assert toolHome == '/env/neo'
        assert jlr.log.contains("SAP Cloud Platform Console Client home '/env/neo' retrieved from environment.")
    }

    @Test
    void getToolHomeFromConfigurationTest() {

        configuration = [neoHome: '/config/neo']

        def toolHome = tool.getToolLocation(script, configuration)

        assert toolHome == '/config/neo'
        assert jlr.log.contains("SAP Cloud Platform Console Client home '/config/neo' retrieved from configuration.")
    }

    @Test
    void getToolHomeFromCurrentWorkingDirectoryTest() {

        def toolHome = tool.getToolLocation(script, configuration)

        assert toolHome == ''
        assert jlr.log.contains("SAP Cloud Platform Console Client is on PATH.")
    }

    @Test
    void getToolTest() {

        configuration = [neoHome: '/config/neo']

        def toolExecutable = tool.getTool(script, configuration)

        assert toolExecutable == '/config/neo/tools/neo.sh'
    }

    @Test
    void getToolExecutableTest() {

        configuration = [neoHome: '/config/neo']

        def toolExecutable = tool.getToolExecutable(script, configuration)

        assert toolExecutable == '/config/neo/tools/neo.sh'
        assert jlr.log.contains("Using SAP Cloud Platform Console Client '/config/neo/tools/neo.sh'.")
    }

    @Test
    void verifyToolHomeTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVars(m) })

        tool.verifyToolLocation(script, configuration)

        assert jlr.log.contains("Verifying SAP Cloud Platform Console Client location '/env/neo'.")
        assert jlr.log.contains("Verification success. SAP Cloud Platform Console Client location '/env/neo' exists.")
    }

    @Test
    void verifyToolExecutableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVars(m) })

        tool.verifyToolExecutable(script, configuration)

        assert jlr.log.contains("Verifying SAP Cloud Platform Console Client '/env/neo/tools/neo.sh'.")
        assert jlr.log.contains("Verification success. SAP Cloud Platform Console Client '/env/neo/tools/neo.sh' exists.")
    }

    @Test
    void verifyToolVersionTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        tool.verifyVersion(script, configuration)

        assert jlr.log.contains("Verifying SAP Cloud Platform Console Client version 3.39.10 or compatible version.")
        assert jlr.log.contains("Verification success. SAP Cloud Platform Console Client version 3.39.10 is installed.")
    }

    @Test
    void verifyToolVersion_FailedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The verification of SAP Cloud Platform Console Client failed. Please check 'neo.sh'. script returned exit code 127.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionFailed(m) })

        tool.verifyVersion(script, configuration)
    }

    @Test
    void verifyToolVersion_IncompatibleVersionTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The installed version of SAP Cloud Platform Console Client is 1.0.5. Please install version 3.39.10 or a compatible version.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        tool.verifyVersion(script, configuration)
    }

    @Test
    void verifyToolVersion_WithMultipleVersionsTest() {

        def neoVersions = ['neo-java-web': '3.39.10', 'neo-javaee6-wp': '2.132.6', 'neo-javaee7-wp': '1.21.13']
        def tool = new ToolDescriptor('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', neoVersions, 'version')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        tool.verifyVersion(script, configuration)

        assert jlr.log.contains("Verifying SAP Cloud Platform Console Client version 3.39.10 or compatible version.")
        assert jlr.log.contains("Verification success. SAP Cloud Platform Console Client version 3.39.10 is installed.")
    }

    @Test
    void verifyToolVersion_WithMultipleVersions_FailedTest() {

        def neoVersions = ['neo-java-web': '3.39.10', 'neo-javaee6-wp': '2.132.6', 'neo-javaee7-wp': '1.21.13']
        def tool = new ToolDescriptor('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', neoVersions, 'version')

        thrown.expect(AbortException)
        thrown.expectMessage("The verification of SAP Cloud Platform Console Client failed. Please check 'neo.sh'. script returned exit code 127.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionFailed(m) })

        tool.verifyVersion(script, configuration)
    }

    @Test
    void verifyToolVersion_WithMultipleVersions_IncompatibleVersionTest() {

        def neoVersions = ['neo-java-web': '3.39.10', 'neo-javaee6-wp': '2.132.6', 'neo-javaee7-wp': '1.21.13']
        def tool = new ToolDescriptor('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', neoVersions, 'version')

        thrown.expect(AbortException)
        thrown.expectMessage("The installed version of SAP Cloud Platform Console Client is 1.0.5. Please install version 3.39.10 or a compatible version.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        tool.verifyVersion(script, configuration)
    }


    private getEnvVars(Map m) {

        if(m.script.contains('NEO_HOME')) {
            return '/env/neo'
        } else {
            return 0
        }
    }

    private getNoEnvVars(Map m) {

        if(m.script.contains('NEO_HOME')) {
            return ''
        } else if(m.script.contains('which neo')) {
            return 0
        } else {
            return 0
        }
    }

    private getVersion(Map m) {

        if(m.script.contains('neo.sh version')) {
            return '''SAP Cloud Platform Console Client
                    SDK version    : 3.39.10
                    Runtime        : neo-java-web'''
        } else {
            return getNoEnvVars(m)
        }
    }

    private getVersionFailed(Map m) {

        if(m.script.contains('neo.sh version')) {
            throw new AbortException('script returned exit code 127')
        } else {
            return getNoEnvVars(m)
        }
    }

    private getIncompatibleVersion(Map m) {

        if(m.script.contains('neo.sh version')) {
            return '''SAP Cloud Platform Console Client
                    SDK version    : 1.0.5
                    Runtime        : neo-java-web'''
        } else {
            return getNoEnvVars(m)
        }
    }
}
