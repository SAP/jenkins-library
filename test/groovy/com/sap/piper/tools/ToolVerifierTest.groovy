package com.sap.piper.tools

import org.junit.BeforeClass
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

    private ExpectedException thrown = new ExpectedException().none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
                                      .around(thrown)
                                      .around(jlr)

    private script
    private static configuration
    private static tool


    @BeforeClass
    static void init() {

        configuration = [mtaJarLocation: 'home']
        tool = new Tool('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '/', 'mta.jar', '1.0.6', '-v')
    }

    @Before
    void setup() {

        script = loadScript('mtaBuild.groovy').mtaBuild
    }


    @Test
    void verifyToolHomeTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getToolHome(m) })

        ToolVerifier.verifyToolHome(tool, script, configuration)

        assert jlr.log.contains("Verifying $tool.name home '/env/mta'.")
        assert jlr.log.contains("Verification success. $tool.name home '/env/mta' exists.")
    }

    @Test
    void verifyToolExecutableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getToolHome(m) })

        ToolVerifier.verifyToolExecutable(tool, script, configuration)

        assert jlr.log.contains("Verifying $tool.name executable '/env/mta/mta.jar'.")
        assert jlr.log.contains("Verification success. $tool.name executable '/env/mta/mta.jar' exists.")
    }

    @Test
    void verifyToolVersionTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        ToolVerifier.verifyToolVersion(tool, script, configuration)

        assert jlr.log.contains("Verifying $tool.name version $tool.version or compatible version.")
        assert jlr.log.contains("Verification success. $tool.name version $tool.version is installed.")
    }

    @Test
    void verifyToolVersion_FailedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The verification of $tool.name failed.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionFailed(m) })

        ToolVerifier.verifyToolVersion(tool, script, configuration)
    }

    @Test
    void verifyToolVersion_IncompatibleVersionTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The installed version of $tool.name is 1.0.5.")

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        ToolVerifier.verifyToolVersion(tool, script, configuration)
    }


    private getToolHome(Map m) {

        if(m.script.contains('MTA_JAR_LOCATION')) {
            return '/env/mta'
        } else {
            return 0
        }
    }

    private getVersion(Map m) {

        if(m.script.contains('mta.jar -v')) {
            return '1.0.6'
        } else {
            return ''
        }
    }

    private getVersionFailed(Map m) {

        if(m.script.contains('mta.jar -v')) {
            throw new AbortException('script returned exit code 127')
        } else {
            return ''
        }
    }

    private getIncompatibleVersion(Map m) {

        if(m.script.contains('mta.jar -v')) {
            return '1.0.5'
        } else {
            return ''
        }
    }
}

