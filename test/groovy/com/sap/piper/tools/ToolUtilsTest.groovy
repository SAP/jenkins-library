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

import com.sap.piper.tools.Tool
import com.sap.piper.tools.ToolUtils


class ToolUtilsTest extends BasePipelineTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
                .around(thrown)
                .around(jlr)

    private cpe

    private static java
    private static mta
    private static neo
    private static cmCli


    @BeforeClass
    static void createTools() {

        java = new Tool('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
        mta = new Tool('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '/', 'mta.jar', '1.0.6', '-v')
        neo = new Tool('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', '3.39.10', 'version')
        cmCli = new Tool('Change Management Command Line Interface', 'CM_CLI_HOME', 'cmCliHome', '/bin/', 'cmclient', '0.0.1', '-v')
    }

    @Before
    void setup() {

        cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
    }

    @Test
    void getMtaJarFromConfigurationTest() {

        def environment = [:]
        def configuration = [mtaJarLocation: '/config/mta']

        def mtaJar = ToolUtils.getToolExecutable(mta, cpe, configuration, environment)

        assert mtaJar == '/config/mta/mta.jar'
        assert jlr.log.contains("$mta.name home '/config/mta' retrieved from configuration.")
        assert jlr.log.contains("Using $mta.name executable '/config/mta/mta.jar'.")
    }

    @Test
    void getMtaJarFromEnvironmentTest() {

        def environment = [MTA_JAR_LOCATION: '/env/mta']
        def configuration = [mtaJarLocation: '/config/mta']

        def mtaJar = ToolUtils.getToolExecutable(mta, cpe, configuration, environment)

        assert mtaJar == '/env/mta/mta.jar'
        assert jlr.log.contains("$mta.name home '/env/mta' retrieved from environment.")
        assert jlr.log.contains("Using $mta.name executable '/env/mta/mta.jar'.")
    }

    @Test
    void getMtaJarFromCurrentWorkingDirectoryTest() {

        def environment = [:]
        def configuration = [:]

        def mtaJar = ToolUtils.getToolExecutable(mta, cpe, configuration, environment)

        assert mtaJar == 'mta.jar'
        assert jlr.log.contains("$mta.name expected on PATH.")
        assert jlr.log.contains("Using $mta.name executable 'mta.jar'.")
    }

    @Test
    void getNeoExecutableFromConfigurationTest() {

        def environment = [:]
        def configuration = [neoHome: '/config/neo']

        def neoExecutable = ToolUtils.getToolExecutable(neo, cpe, configuration, environment)

        assert neoExecutable == '/config/neo/tools/neo.sh'
        assert jlr.log.contains("$neo.name home '/config/neo' retrieved from configuration.")
        assert jlr.log.contains("Using $neo.name executable '/config/neo/tools/neo.sh'.")
    }

    @Test
    void getNeoExecutableFromEnvironmentTest() {

        def environment = [NEO_HOME: '/env/neo']
        def configuration = [neoHome: '/config/neo']

        def neoExecutable = ToolUtils.getToolExecutable(neo, cpe, configuration, environment)

        assert neoExecutable == '/env/neo/tools/neo.sh'
        assert jlr.log.contains("$neo.name home '/env/neo' retrieved from environment.")
        assert jlr.log.contains("Using $neo.name executable '/env/neo/tools/neo.sh'.")
    }

    @Test
    void getNeoExecutableFromCurrentWorkingDirectoryTest() {

        def environment = [:]
        def configuration = [:]

        def neoExecutable = ToolUtils.getToolExecutable(neo, cpe, configuration, environment)

        assert neoExecutable == 'neo.sh'
        assert jlr.log.contains("$neo.name expected on PATH.")
        assert jlr.log.contains("Using $neo.name executable 'neo.sh'.")
    }

    @Test
    void getCmCliExecutableFromConfigurationTest() {

        def environment = [:]
        def configuration = [cmCliHome: '/config/cmclient']

        def cmCliExecutable = ToolUtils.getToolExecutable(cmCli, cpe, configuration, environment)

        assert cmCliExecutable == '/config/cmclient/bin/cmclient'
        assert jlr.log.contains("$cmCli.name home '/config/cmclient' retrieved from configuration.")
        assert jlr.log.contains("Using $cmCli.name executable '/config/cmclient/bin/cmclient'.")
    }

    @Test
    void getCmCliExecutableFromEnvironmentTest() {

        def environment = [CM_CLI_HOME: '/env/cmclient']
        def configuration = [cmCliHome: '/config/cmclient']

        def cmCliExecutable = ToolUtils.getToolExecutable(cmCli, cpe, configuration, environment)

        assert cmCliExecutable == '/env/cmclient/bin/cmclient'
        assert jlr.log.contains("$cmCli.name home '/env/cmclient' retrieved from environment.")
        assert jlr.log.contains("Using $cmCli.name executable '/env/cmclient/bin/cmclient'.")
    }

    @Test
    void getCmCliExecutableFromCurrentWorkingDirectoryTest() {

        def environment = [:]
        def configuration = [:]

        def cmCliExecutable = ToolUtils.getToolExecutable(cmCli, cpe, configuration, environment)

        assert cmCliExecutable == 'cmclient'
        assert jlr.log.contains("$cmCli.name expected on PATH.")
        assert jlr.log.contains("Using $cmCli.name executable 'cmclient'.")
    }
}
