package com.sap.piper.tools

import org.junit.Ignore
import org.junit.BeforeClass
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.Rules

import hudson.AbortException

class JavaArchiveDescriptorTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
                .around(thrown)
                .around(jlr)

    private static javaArchive
    private static configuration

    private script


    @BeforeClass
    static void init() {

        def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
        javaArchive = new JavaArchiveDescriptor('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '1.0.6', '-v', java)
    }

    @Before
    void setup() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoEnvVars(m) })
        helper.registerAllowedMethod('error', [String], { s -> throw new hudson.AbortException(s) })

        script = loadScript('mtaBuild.groovy').mtaBuild

        configuration = [:] //no default configuration
    }

    @Test
    void getJavaArchiveFileFromEnvironmentTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVars(m) })

        def javaArchiveFile = javaArchive.getFile(script, configuration)

        assert javaArchiveFile == '/env/mta/mta_archive_builder-1.1.0.jar'
        assert jlr.log.contains("SAP Multitarget Application Archive Builder file '/env/mta/mta_archive_builder-1.1.0.jar' retrieved from environment.")
    }

    @Test
    void getJavaArchiveFileFromConfigurationTest() {

        configuration = [mtaJarLocation: '/config/mta/mta_archive_builder-1.1.0.jar']

        def javaArchiveFile = javaArchive.getFile(script, configuration)

        assert javaArchiveFile == '/config/mta/mta_archive_builder-1.1.0.jar'
        assert jlr.log.contains("SAP Multitarget Application Archive Builder file '/config/mta/mta_archive_builder-1.1.0.jar' retrieved from configuration.")
    }

    // Compatibility tests
    @Test
    void getJavaArchiveFileFromEnvironment_CompatibilityTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVarsWithCompatibility(m) })

        def javaArchiveFile = javaArchive.getFile(script, configuration)

        assert javaArchiveFile == '/env/mta/mta.jar'
        assert jlr.log.contains("SAP Multitarget Application Archive Builder file '/env/mta' retrieved from environment.")
    }

    @Test
    void getJavaArchiveFileFromConfiguration_CompatibilityTest() {

        configuration = [mtaJarLocation: '/config/mta']

        def javaArchiveFile = javaArchive.getFile(script, configuration)

        assert javaArchiveFile == '/config/mta/mta.jar'
        assert jlr.log.contains("SAP Multitarget Application Archive Builder file '/config/mta' retrieved from configuration.")
    }
    //

    @Test
    void getJavaArchiveFileFailedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Please, configure SAP Multitarget Application Archive Builder. SAP Multitarget Application Archive Builder can be set using the environment variable 'MTA_JAR_LOCATION', or " +
                             "using the configuration key 'mtaJarLocation'.")

        javaArchive.getFile(script, configuration)
     }

    @Ignore('while compatibility code is not removed')
    void getJavaArchiveFileFromEnvironment_UnexpectedFormatTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The value '/env/mta/mta_archive_builder-1.1.0.jarr' of the environment variable 'MTA_JAR_LOCATION' has an unexpected format.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getUnexpectedFormatEnvVars(m) })

        javaArchive.getFile(script, configuration)
    }

    @Ignore('while compatibility code is not removed')
    void getJavaArchiveFileFromConfiguration_UnexpectedFormatTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The value '/config/mta/mta_archive_builder-1.1.0.jarr' of the configuration key 'mtaJarLocation' has an unexpected format.")

        configuration = [mtaJarLocation: '/config/mta/mta_archive_builder-1.1.0.jarr']

        javaArchive.getFile(script, configuration)
    }

    @Test
    void getJavaArchiveCallTest() {

        configuration = [mtaJarLocation: '/config/mta/mta_archive_builder-1.1.0.jar']

        def javaArchiveCall = javaArchive.getCall(script, configuration)

        assert javaArchiveCall == 'java -jar /config/mta/mta_archive_builder-1.1.0.jar'
        assert jlr.log.contains("Using SAP Multitarget Application Archive Builder '/config/mta/mta_archive_builder-1.1.0.jar'.")
    }

    @Test
    void verifyJavaArchiveFileTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getEnvVars(m) })

        javaArchive.verifyFile(script, configuration)

        assert jlr.log.contains("Verifying SAP Multitarget Application Archive Builder '/env/mta/mta_archive_builder-1.1.0.jar'.")
        assert jlr.log.contains("Verification success. SAP Multitarget Application Archive Builder '/env/mta/mta_archive_builder-1.1.0.jar' exists.")
    }

    @Test
    void verifyJavaArchiveVersionTest() {

        configuration = [mtaJarLocation: 'mta_archive_builder-1.1.0.jar']

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        javaArchive.verifyVersion(script, configuration)

        assert jlr.log.contains("Verifying SAP Multitarget Application Archive Builder version 1.0.6 or compatible version.")
        assert jlr.log.contains("Verification success. SAP Multitarget Application Archive Builder version 1.0.6 is installed.")
    }

    @Test
    void verifyJavaArchiveVersion_FailedTest() {

        configuration = [mtaJarLocation: 'mta_archive_builder-1.1.0.jar']

        thrown.expect(AbortException)
        thrown.expectMessage("The verification of SAP Multitarget Application Archive Builder failed. Please check 'java -jar mta_archive_builder-1.1.0.jar'. script returned exit code 127.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionFailed(m) })

        javaArchive.verifyVersion(script, configuration)
    }

    @Test
    void verifyJavaArchiveVersion_IncompatibleVersionTest() {

        configuration = [mtaJarLocation: '/config/mta/mta_archive_builder-1.1.0.jar']

        thrown.expect(AbortException)
        thrown.expectMessage("The installed version of SAP Multitarget Application Archive Builder is 1.0.5. Please install version 1.0.6 or a compatible version.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        javaArchive.verifyVersion(script, configuration)
    }


    private getEnvVars(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return '/env/java'
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return '/env/mta/mta_archive_builder-1.1.0.jar'
        } else {
            return 0
        }
    }

    private getEnvVarsWithCompatibility(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return '/env/java'
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return '/env/mta'
        } else {
            return 0
        }
    }

    /*
    private getUnexpectedFormatEnvVars(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return '/env/java'
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return '/env/mta/mta_archive_builder-1.1.0.jarr'
        } else {
            return 0
        }
    }
    */

    private getNoEnvVars(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return ''
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return ''
        } else if(m.script.contains('which java')) {
            return 0
        } else {
            return 0
        }
    }

    private getVersion(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta_archive_builder-1.1.0.jar -v')) {
            return '1.0.6'
        } else {
            return getNoEnvVars(m)
        }
    }

    private getVersionFailed(Map m) {

        if(m.script.contains('java -version') || m.script.contains('mta_archive_builder-1.1.0.jar -v')) {
            throw new AbortException('script returned exit code 127')
        } else {
            return getNoEnvVars(m)
        }
    }

    private getIncompatibleVersion(Map m) {

        if(m.script.contains('java -version') || m.script.contains('mta_archive_builder-1.1.0.jar -v')) {
            return '1.0.5'
        } else {
            return getNoEnvVars(m)
        }
    }
}
