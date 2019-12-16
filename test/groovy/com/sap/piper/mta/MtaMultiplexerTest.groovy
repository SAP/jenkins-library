package com.sap.piper.mta

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasSize
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not

import static org.junit.Assert.assertThat
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.Rules

class MtaMultiplexerTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(loggingRule)
        .around(thrown)

    @Test
    void testFilterFiles() {
        // prepare test data
        def files = [
            new File("pom.xml"),
            new File("some-ui${File.separator}pom.xml"),
            new File("some-service${File.separator}pom.xml"),
            new File("some-other-service${File.separator}pom.xml")
        ].toArray()
        // execute test
        def result = MtaMultiplexer.removeExcludedFiles(nullScript, files, ['pom.xml'])
        // asserts
        assertThat(result, not(hasItem('pom.xml')))
        assertThat(result, hasSize(3))
        assertThat(loggingRule.log, containsString('Skipping pom.xml'))
    }

    @Test
    void testCreateJobs() {
        def optionsList = []
        // prepare test data
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            if (map.glob == "**${File.separator}pom.xml") {
                return [new File("some-service${File.separator}pom.xml"), new File("some-other-service${File.separator}pom.xml")].toArray()
            }
            if (map.glob == "**${File.separator}package.json") {
                return [new File("some-ui${File.separator}package.json"), new File("somer-service-broker${File.separator}package.json")].toArray()
            }
            return [].toArray()
        })
        // execute test
        def result = MtaMultiplexer.createJobs(nullScript, ['myParameters':'value'], [], 'TestJobs', 'pom.xml', 'maven'){
            options -> optionsList.push(options)
        }
        // invoke jobs
        for(Closure c : result.values()) c()
        // asserts
        assertThat(result.size(), is(2))
        assertThat(result, hasKey('TestJobs - some-other-service'))
        assertThat(loggingRule.log, containsString("Found 2 maven descriptor files: [some-service${File.separator}pom.xml, some-other-service${File.separator}pom.xml]".toString()))
        assertThat(optionsList.get(0), hasEntry('myParameters', 'value'))
        assertThat(optionsList.get(0), hasEntry('scanType', 'maven'))
        assertThat(optionsList.get(0), hasEntry('buildDescriptorFile', "some-service${File.separator}pom.xml".toString()))
        assertThat(optionsList.get(1), hasEntry('myParameters', 'value'))
        assertThat(optionsList.get(1), hasEntry('scanType', 'maven'))
        assertThat(optionsList.get(1), hasEntry('buildDescriptorFile', "some-other-service${File.separator}pom.xml".toString()))
    }
}
