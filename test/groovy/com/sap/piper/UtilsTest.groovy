package com.sap.piper

import groovy.transform.Field
import org.junit.Rule
import org.junit.Before
import org.junit.Test
import util.JenkinsDockerExecuteRule
import util.JenkinsReadJsonRule
import util.JenkinsReadYamlRule
import util.JenkinsWriteFileRule

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.startsWith
import static org.junit.Assert.assertThat
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import static org.hamcrest.Matchers.is

import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.BasePiperTest
import util.Rules

class UtilsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(shellRule)
        .around(loggingRule)
        .around(readYamlRule)
        .around(writeFileRule)
        .around(readJsonRule)
        .around(dockerExecuteRule)

    private parameters

    @Before
    void setup() {

        parameters = [:]

    }

    @Test
    void testGenerateSHA1() {
        def result = utils.generateSha1('ContinuousDelivery')
        // asserts
        // generated with "echo -n 'ContinuousDelivery' | sha1sum | sed 's/  -//'"
        assertThat(result, is('0dad6c33b6246702132454f604dee80740f399ad'))
    }

    @Test
    void testUnstashAllSkipNull() {

        def stashResult = utils.unstashAll(['a', null, 'b'])
        assert stashResult == ['a', 'b']
    }

    @Test
    void testRunGoStep() {

        List withEnvArgs = []

        helper.registerAllowedMethod("withEnv", [List.class, Closure.class], { arguments, closure ->
            arguments.each { arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })

        readYamlRule.registerYaml('metadata/aCommand.yaml', [:])


        shellRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/metadata/mavenStaticCodeChecks.yaml\'', '{"dockerImage": "maven:3.6-jdk-8-test"}')

        Script mockedStep = new Script() {

            String GO_COMMAND = 'mavenExecuteStaticCodeChecks'
            String METADATA_FILE = 'metadata/mavenStaticCodeChecks.yaml'
            String STEP_NAME = 'mavenExecuteStaticCodeChecks'
            String METADATA_FOLDER = '.pipeline'

            def run() {
                // it never runs
                throw new UnsupportedOperationException()
            }
        }

        Map parameters = [
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript
        ]

        utils.runGoStepWithDocker(mockedStep, parameters)

        // asserts
        assertThat(writeFileRule.files['.pipeline/metadata/mavenStaticCodeChecks.yaml'], containsString('name: mavenExecuteStaticCodeChecks'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellRule.shell[1], is('./piper mavenExecuteStaticCodeChecks'))

        assertThat(dockerExecuteRule.dockerParams.dockerImage, is('maven:3.6-jdk-8-test'))

    }
}
