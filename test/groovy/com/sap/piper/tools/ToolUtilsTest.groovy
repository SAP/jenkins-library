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

    private script

    private static tool
    private static configuration


    @BeforeClass
    static void init() {

        tool = new Tool('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '/', 'mta.jar', '1.0.6', '-v')
    }

    @Before
    void setup() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '' })

        script = loadScript('mtaBuild.groovy').mtaBuild

        configuration = [:]
    }

    @Test
    void getToolHomeFromEnvironmentTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '/env/mta' })

        def toolHome = ToolUtils.getToolHome(tool, script, configuration)

        assert toolHome == '/env/mta'
        assert jlr.log.contains("$tool.name home '/env/mta' retrieved from environment.")
    }

    @Test
    void getToolHomeFromConfigurationTest() {

        configuration = [mtaJarLocation: '/config/mta']

        def toolHome = ToolUtils.getToolHome(tool, script, configuration)

        assert toolHome == '/config/mta'
        assert jlr.log.contains("$tool.name home '/config/mta' retrieved from configuration.")
    }

    @Test
    void getToolHomeFromCurrentWorkingDirectoryTest() {

        def toolHome = ToolUtils.getToolHome(tool, script, configuration)

        assert toolHome == ''
        assert jlr.log.contains("$tool.name expected on PATH or current working directory.")
    }

    @Test
    void getToolExecutableTest() {

        configuration = [mtaJarLocation: '/config/mta']

        def toolExecutable = ToolUtils.getToolExecutable(tool, script, configuration)

        assert toolExecutable == '/config/mta/mta.jar'
        assert jlr.log.contains("Using $tool.name executable '/config/mta/mta.jar'.")
    }
}
