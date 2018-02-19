package com.sap.piper

import org.junit.Rule
import org.junit.Before
import org.junit.Test
import org.junit.rules.ExpectedException

import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsLoggingRule

import com.sap.piper.ToolUtils

import com.lesfurets.jenkins.unit.BasePipelineTest


class ToolUtilsTest extends BasePipelineTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
                .around(thrown)
                .around(jlr)

    private cpe


    @Before
    void setup() {

        cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
    }

    @Test
    void getMtaJarFromConfigurationTest() {

        def configuration = [mtaJarLocation: '/config/mta']
        def environment = [MTA_JAR_LOCATION: '/env/mta']

        def mtaJar = ToolUtils.getMtaJar(cpe, 'test', configuration, environment)

        assert mtaJar == '/config/mta/mta.jar'
        assert jlr.log.contains('[test] MTA JAR "/config/mta/mta.jar" retrieved from configuration.')
    }

    @Test
    void getMtaJarFromEnvironmentTest() {

        def environment = [MTA_JAR_LOCATION: '/env/mta']

        def mtaJar = ToolUtils.getMtaJar(cpe, 'test', [:], environment)

        assert mtaJar == '/env/mta/mta.jar'
        assert jlr.log.contains('[test] MTA JAR "/env/mta/mta.jar" retrieved from environment.')
    }

    @Test
    void getMtaJarFromCurrentWorkingDirectoryTest() {

        def mtaJar = ToolUtils.getMtaJar(cpe, 'test', [:], [:])

        assert mtaJar == 'mta.jar'
        assert jlr.log.contains('[test] Using MTA JAR from current working directory.')
    }

    @Test
    void getNeoExecutableFromConfigurationTest() {

        def configuration = [neoHome: '/config/neo']
        def environment = [NEO_HOME: '/env/neo']

        def neoExecutable = ToolUtils.getNeoExecutable(cpe, 'test', configuration, environment)

        assert neoExecutable == '/config/neo/tools/neo.sh'
        assert jlr.log.contains('[test] Neo executable "/config/neo/tools/neo.sh" retrieved from configuration.')
    }

    @Test
    void getNeoExecutableFromEnvironmentTest() {

        def environment = [NEO_HOME: '/env/neo']

        def neoExecutable = ToolUtils.getNeoExecutable(cpe, 'test', [:], environment)

        assert neoExecutable == '/env/neo/tools/neo.sh'
        assert jlr.log.contains('[test] Neo executable "/env/neo/tools/neo.sh" retrieved from environment.')
    }

    @Test
    void getNeoExecutableFromCurrentWorkingDirectoryTest() {

        def neoExecutable = ToolUtils.getNeoExecutable(cpe, 'test', [:], [:])

        assert neoExecutable == 'neo.sh'
        assert jlr.log.contains('[test] Using Neo executable from PATH.')
    }

    @Test
    void getCmCliExecutableFromConfigurationTest() {

        def configuration = [cmCliHome: '/config/cmclient']
        def environment = [CM_CLI_HOME: '/env/cmclient']

        def cmCliExecutable = ToolUtils.getCmCliExecutable(cpe, 'test', configuration, environment)

        assert cmCliExecutable == '/config/cmclient/bin/cmclient'
        assert jlr.log.contains('[test] Change Management Command Line Interface "/config/cmclient/bin/cmclient" retrieved from configuration.')
    }

    @Test
    void getCmCliExecutableFromEnvironmentTest() {

        def environment = [CM_CLI_HOME: '/env/cmclient']

        def cmCliExecutable = ToolUtils.getCmCliExecutable(cpe, 'test', [:], environment)

        assert cmCliExecutable == '/env/cmclient/bin/cmclient'
        assert jlr.log.contains('[test] Change Management Command Line Interface "/env/cmclient/bin/cmclient" retrieved from environment.')
    }

    @Test
    void getCmCliExecutableFromCurrentWorkingDirectoryTest() {

        def cmCliExecutable = ToolUtils.getCmCliExecutable(cpe, 'test', [:], [:])

        assert cmCliExecutable == 'cmclient'
        assert jlr.log.contains('[test] Change Management Command Line Interface retrieved from current working directory.')
    }
}
