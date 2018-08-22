package com.sap.piper

import org.junit.Rule
import org.junit.Before
import org.junit.Test
import org.junit.rules.ExpectedException

import util.JenkinsLoggingRule
import util.JenkinsShellCallRule

import com.sap.piper.Utils

class UtilsTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jscr)
        .around(jlr)
    
    private utils = new Utils()
    private parameters

    @Before
    void setup() {

        parameters = [:]
    }

    @Test
    void noValueGetMandatoryParameterTest() {

        thrown.expect(Exception)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR test")

        utils.getMandatoryParameter(parameters, 'test', null)
    }

    @Test
    void defaultValueGetMandatoryParameterTest() {

        assert  utils.getMandatoryParameter(parameters, 'test', 'default') == 'default'
    }

    @Test
    void valueGetmandatoryParameterTest() {

        parameters.put('test', 'value')

        assert utils.getMandatoryParameter(parameters, 'test', null) == 'value'
    }

    @Test
    void testSWAReporting() {
        utils.pushToSWA(
            [step: 'anything'],
            null
        )
        println("SHELL: ${jscr.shell}")
        println("LOG: ${jlr.log}")
//        assertThat(jscr, containsString)
    }
}
