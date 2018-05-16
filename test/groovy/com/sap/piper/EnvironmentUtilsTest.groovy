package com.sap.piper.tools

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.Rules

import com.lesfurets.jenkins.unit.BasePipelineTest

import com.sap.piper.EnvironmentUtils

import hudson.AbortException


class EnvironmentUtilsTest extends BasePipelineTest {

    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
                .around(thrown)

    private script


    @Before
    void setup() {

        script = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
    }

    @Test
    void isEnvironmentVariableFailedTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> throw new AbortException('') })

        thrown.expect(AbortException)
        thrown.expectMessage("There was an error requesting the environment variable 'JAVA_HOME'.")

        EnvironmentUtils.isEnvironmentVariable(script, 'JAVA_HOME')
    }

    @Test
    void isNotEnvironmentVariableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '' })

        def isEnvVar = EnvironmentUtils.isEnvironmentVariable(script, 'JAVA_HOME')

        assert isEnvVar == false
    }

    @Test
    void isEnvironmentVariableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '/env/java' })

        def isEnvVar = EnvironmentUtils.isEnvironmentVariable(script, 'JAVA_HOME')

        assert isEnvVar == true
    }

    @Test
    void getEnvironmentVariableFailedTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> throw new AbortException('') })

        thrown.expect(AbortException)
        thrown.expectMessage("There was an error requesting the environment variable 'JAVA_HOME'.")

        EnvironmentUtils.getEnvironmentVariable(script, 'JAVA_HOME')
    }

    @Test
    void getEnvironmentVariableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '/env/java' })

        def envVar = EnvironmentUtils.getEnvironmentVariable(script, 'JAVA_HOME')

        assert envVar == '/env/java'
    }
}
