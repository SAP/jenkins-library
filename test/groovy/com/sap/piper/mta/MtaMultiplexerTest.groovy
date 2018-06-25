package com.sap.piper.mta

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasSize
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
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(thrown)

    @Test
    void testFilterFiles() {
        // prepare test data
        def files = [
            new File('pom.xml'),
            new File('some-ui/pom.xml'),
            new File('some-service/pom.xml'),
            new File('some-other-service/pom.xml')
        ].toArray()
        // execute test
        def result = MtaMultiplexer.removeExcludedFiles(nullScript, files, ['pom.xml'])
        // asserts
        assertThat(result, not(hasItem('pom.xml')))
        assertThat(result, hasSize(3))
        assertThat(jlr.log, containsString('Skipping pom.xml'))
    }
}
