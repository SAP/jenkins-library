package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsStepRule
import util.Rules

import hudson.AbortException


class EnvironmentUtilsTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)

    @Test
    void isEnvironmentVariableFailedTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> throw new AbortException('') })

        thrown.expect(AbortException)
        thrown.expectMessage("There was an error requesting the environment variable 'JAVA_HOME'.")

        EnvironmentUtils.isEnvironmentVariable(nullScript, 'JAVA_HOME')
    }

    @Test
    void isNotEnvironmentVariableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '' })

        def isEnvVar = EnvironmentUtils.isEnvironmentVariable(nullScript, 'JAVA_HOME')

        assert isEnvVar == false
    }

    @Test
    void isEnvironmentVariableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '/env/java' })

        def isEnvVar = EnvironmentUtils.isEnvironmentVariable(nullScript, 'JAVA_HOME')

        assert isEnvVar == true
    }

    @Test
    void getEnvironmentVariableFailedTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> throw new AbortException('') })

        thrown.expect(AbortException)
        thrown.expectMessage("There was an error requesting the environment variable 'JAVA_HOME'.")

        EnvironmentUtils.getEnvironmentVariable(nullScript, 'JAVA_HOME')
    }

    @Test
    void getEnvironmentVariableTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return '/env/java' })

        def envVar = EnvironmentUtils.getEnvironmentVariable(nullScript, 'JAVA_HOME')

        assert envVar == '/env/java'
    }
}
