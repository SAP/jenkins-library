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
import com.sap.piper.tools.JavaArchiveDescriptor

import hudson.AbortException


class JavaArchiveDescriptorTest extends BasePipelineTest {

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

        def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
        tool = new JavaArchiveDescriptor('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '/', 'mta.jar', '1.0.6', '-v', java, '-jar')
    }

    @Before
    void setup() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoEnvVars(m) })

        script = loadScript('mtaBuild.groovy').mtaBuild

        configuration = [:]
    }

    @Test
    void getToolHomeFromEnvironmentTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVars(m) })

        def toolHome = tool.getHome(script, configuration)

        assert toolHome == '/env/mta'
        assert jlr.log.contains("$tool.name home '/env/mta' retrieved from environment.")
    }

    @Test
    void getToolHomeFromConfigurationTest() {

        configuration = [mtaJarLocation: '/config/mta']

        def toolHome = tool.getHome(script, configuration)

        assert toolHome == '/config/mta'
        assert jlr.log.contains("$tool.name home '/config/mta' retrieved from configuration.")
    }

    @Test
    void getToolHomeFromCurrentWorkingDirectoryTest() {

        def toolHome = tool.getHome(script, configuration)

        assert toolHome == ''
        assert jlr.log.contains("$tool.name expected on current working directory.")
    }

    @Test
    void getToolTest() {

        configuration = [mtaJarLocation: '/config/mta']

        def toolExecutable = tool.getTool(script, configuration)

        assert toolExecutable == '/config/mta/mta.jar'
    }

    @Test
    void getToolExecutableTest() {

        configuration = [mtaJarLocation: '/config/mta']

        def toolExecutable = tool.getExecutable(script, configuration)

        assert toolExecutable == 'java -jar /config/mta/mta.jar'
        assert jlr.log.contains("Using $tool.name executable 'java -jar /config/mta/mta.jar'.")
    }

    @Test
    void verifyToolHomeTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVars(m) })

        tool.verifyHome(script, configuration)

        assert jlr.log.contains("Verifying $tool.name home '/env/mta'.")
        assert jlr.log.contains("Verification success. $tool.name home '/env/mta' exists.")
    }

    @Test
    void verifyToolExecutableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVars(m) })

        tool.verifyTool(script, configuration)

        assert jlr.log.contains("Verifying $tool.name '/env/mta/mta.jar'.")
        assert jlr.log.contains("Verification success. $tool.name '/env/mta/mta.jar' exists.")
    }

    @Test
    void verifyToolVersionTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        tool.verifyVersion(script, configuration)

        assert jlr.log.contains("Verifying $tool.name version $tool.version or compatible version.")
        assert jlr.log.contains("Verification success. $tool.name version $tool.version is installed.")
    }

    @Test
    void verifyToolVersion_FailedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The verification of $tool.name failed. Please check 'java -jar mta.jar'. script returned exit code 127.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionFailed(m) })

        tool.verifyVersion(script, configuration)
    }

    @Test
    void verifyToolVersion_IncompatibleVersionTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The installed version of $tool.name is 1.0.5. Please install version $tool.version or a compatible version.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        tool.verifyVersion(script, configuration)
    }


    private getEnvVars(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return '/env/java'
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return '/env/mta'
        } else {
            return 0
        }
    }

    private getNoEnvVars(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return ''
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return ''
        } else {
            return 0
        }
    }

    private getVersion(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta.jar -v')) {
            return '1.0.6'
        } else {
            return ''
        }
    }

    private getVersionFailed(Map m) {

        if(m.script.contains('java -version') || m.script.contains('mta.jar -v')) {
            throw new AbortException('script returned exit code 127')
        } else {
            return ''
        }
    }

    private getIncompatibleVersion(Map m) {

        if(m.script.contains('java -version') || m.script.contains('mta.jar -v')) {
            return '1.0.5'
        } else {
            return ''
        }
    }
}
