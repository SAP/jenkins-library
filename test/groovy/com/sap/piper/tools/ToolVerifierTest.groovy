package com.sap.piper.tools

import org.junit.BeforeClass
import org.junit.ClassRule
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.TemporaryFolder
import org.junit.rules.RuleChain

import util.JenkinsLoggingRule
import util.Rules

import com.lesfurets.jenkins.unit.BasePipelineTest

import com.sap.piper.tools.Tool
import com.sap.piper.tools.ToolVerifier

import hudson.AbortException


class ToolVerifierTest extends BasePipelineTest {

    @ClassRule
    public static TemporaryFolder tmp = new TemporaryFolder()

    private ExpectedException thrown = new ExpectedException().none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
                                      .around(thrown)
                                      .around(jlr)

    private static home
    private static mtaJar
    private static neoExecutable
    private static cmCliExecutable

    private script
    private static stepName
    private static configuration
    private static environment

    private static java
    private static mta
    private static neo
    private static cmCli


    @BeforeClass
    static void createTestFiles() {

        home = "${tmp.getRoot()}"
        tmp.newFolder('bin')
        tmp.newFolder('bin', 'java')
        tmp.newFile('mta.jar')
        tmp.newFolder('tools')
        tmp.newFile('tools/neo.sh')
        tmp.newFile('bin/cmclient')

        mtaJar = "$home/mta.jar"
        neoExecutable = "$home/tools/neo.sh"
        cmCliExecutable = "$home/bin/cmclient"

        java = new Tool('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
        mta = new Tool('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '/', 'mta.jar', '1.0.6', '-v')
        neo = new Tool('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', '3.39.10', 'version')
        cmCli = new Tool('Change Management Command Line Interface', 'CM_CLI_HOME', 'cmCliHome', '/bin/', 'cmclient', '0.0.1', '-v')

        configuration = [mtaJarLocation: home, neoHome: home, cmCliHome: home]
        environment = [JAVA_HOME: home]
    }

    @Before
    void init() {

        script = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
    }


    @Test
    void unableToVerifyJavaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The verification of Java failed.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoVersion(m) })

        ToolVerifier.verifyToolVersion(java, script, configuration, environment)
    }

    @Test
    void unableToVerifyMtaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The verification of SAP Multitarget Application Archive Builder failed.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoVersion(m) })

        ToolVerifier.verifyToolVersion(mta, script, configuration, environment)
    }

    @Test
    void unableToVerifyNeoTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The verification of SAP Cloud Platform Console Client failed.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoVersion(m) })

        ToolVerifier.verifyToolVersion(neo, script, configuration, environment)
    }

    @Test
    void unableToVefifyCmTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The verification of Change Management Command Line Interface failed.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoVersion(m) })

        ToolVerifier.verifyToolVersion(cmCli, script, configuration, environment)

        script.execute()
    }

    @Test
    void verifyIncompatibleVersionJavaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The installed version of Java is 1.7.0.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        ToolVerifier.verifyToolVersion(java, script, configuration, environment)
    }

    @Test
    void verifyIncompatibleVersionMtaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The installed version of SAP Multitarget Application Archive Builder is 1.0.5.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        ToolVerifier.verifyToolVersion(mta, script, configuration, environment)
    }

    @Test
    void verifyNeoIncompatibleVersionTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The installed version of SAP Cloud Platform Console Client is 1.126.51.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        ToolVerifier.verifyToolVersion(neo, script, configuration, environment)
    }

    @Test
    void verifyCmIncompatibleVersionTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The installed version of Change Management Command Line Interface is 0.0.0.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })
        binding.setVariable('tool', 'cm')

        ToolVerifier.verifyToolVersion(cmCli, script, configuration, environment)
    }

    @Test
    void verifyJavaTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        ToolVerifier.verifyToolVersion(java, script, configuration, environment)

        assert jlr.log.contains('Verifying Java version 1.8.0 or compatible version.')
        assert jlr.log.contains('Java version 1.8.0 is installed.')
    }

    @Test
    void verifyMtaTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        ToolVerifier.verifyToolVersion(mta, script, configuration, environment)

        assert jlr.log.contains('Verifying SAP Multitarget Application Archive Builder version 1.0.6 or compatible version.')
        assert jlr.log.contains('SAP Multitarget Application Archive Builder version 1.0.6 is installed.')
    }

    @Test
    void verifyNeoTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        ToolVerifier.verifyToolVersion(neo, script, configuration, environment)

        assert jlr.log.contains('Verifying SAP Cloud Platform Console Client version 3.39.10 or compatible version.')
        assert jlr.log.contains('SAP Cloud Platform Console Client version 3.39.10 is installed.')
    }

    @Test
    void verifyCmTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        ToolVerifier.verifyToolVersion(cmCli, script, configuration, environment)

        assert jlr.log.contains('Verifying Change Management Command Line Interface version 0.0.1 or compatible version.')
        assert jlr.log.contains('Change Management Command Line Interface version 0.0.1 is installed.')
    }


    private getNoVersion(Map m) { 
        throw new AbortException('script returned exit code 127')
    }

    private getVersion(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta.jar -v')) {
            return '1.0.6'
        } else if(m.script.contains('neo.sh version')) {
            return '''SAP Cloud Platform Console Client
                    SDK version    : 3.39.10
                    Runtime        : neo-java-web'''
        } else if(m.script.contains('cmclient -v')) {
            return '0.0.1-beta-2 : fc9729964a6acf5c1cad9c6f9cd6469727625a8e'
        }
    }

    private getIncompatibleVersion(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.7.0_121\"
                    OpenJDK Runtime Environment (build 1.7.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta.jar -v')) {
            return '1.0.5'
        } else if(m.script.contains('neo.sh version')) {
            return '''SAP Cloud Platform Console Client
                    SDK version    : 1.126.51
                    Runtime        : neo-java-web'''
        } else if(m.script.contains('cmclient -v')) {
            return '0.0.0-beta-1 : fc9729964a6acf5c1cad9c6f9cd6469727625a8e'
        }
    }
}

