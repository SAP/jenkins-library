package com.sap.piper

import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.Rules

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

class PiperGoUtilsTest extends BasePiperTest {

    public ExpectedException exception = ExpectedException.none()
    public JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(shellCallRule)
        .around(exception)
        .around(loggingRule)

    @Before
    void init() {
        helper.registerAllowedMethod("retry", [Integer, Closure], null)
    }

    @Test
    void testUnstashPiperBinAvailable() {

        def piperBinStash = 'piper-bin'

        // this mocks utils.unstash
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            if (stashFileName != piperBinStash) {
                return []
            }
            return [piperBinStash]
        })

        def piperGoUtils = new PiperGoUtils(nullScript, utils)

        piperGoUtils.unstashPiperBin()
    }


    @Test
    void testUnstashPiperBinMaster() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib-os', version: 'master']]}

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/latest/download/piper\'', '200')

        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(4))
        assertThat(shellCallRule.shell[1].toString(), is('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/latest/download/piper\''))
        assertThat(shellCallRule.shell[2].toString(), is('chmod +x piper'))
        assertThat(shellCallRule.shell[3].toString(), is('./piper version'))
    }

    @Test
    void testUnstashPiperBinNonMaster() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib-os', version: 'testTag']]}

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/download/testTag/piper\'', '200')

        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(4))
        assertThat(shellCallRule.shell[1].toString(), is('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/download/testTag/piper\''))
        assertThat(shellCallRule.shell[2].toString(), is('chmod +x piper'))
        assertThat(shellCallRule.shell[3].toString(), is('./piper version'))
    }

    @Test
    void testUnstashPiperBinFallback() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib-os', version: 'notAvailable']]}

        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('./piper version', "1.2.3")
        shellCallRule.setReturnValue('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/download/notAvailable/piper\'', '404')
        shellCallRule.setReturnValue('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/latest/download/piper\'', '200')

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(5))
        assertThat(shellCallRule.shell[0].toString(), is('[ -x ./piper ]'))
        assertThat(shellCallRule.shell[1].toString(), is('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/download/notAvailable/piper\''))
        assertThat(shellCallRule.shell[2].toString(), is('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/latest/download/piper\''))
        assertThat(shellCallRule.shell[3].toString(), is('chmod +x piper'))
        assertThat(shellCallRule.shell[4].toString(), is ('./piper version'))
    }

    @Test
    void testDownloadFailedWithErrorCode() {
        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib-os', version: 'notAvailable']]}

        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/download/notAvailable/piper\'', '404')
        shellCallRule.setReturnValue('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/latest/download/piper\'', '500')

        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        exception.expectMessage(containsString('Download of Piper go binary failed'))
        piperGoUtils.unstashPiperBin()
    }

    @Test
    void testDownloadFailedWithHTTPCode() {
        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib-os', version: 'notAvailable']]}

        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/download/notAvailable/piper\'', '404')
        shellCallRule.setReturnValue('curl --insecure --silent --retry 5 --retry-max-time 240 --location --write-out \'%{http_code}\' --output piper \'https://github.com/SAP/jenkins-library/releases/latest/download/piper\'', '500')

        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        exception.expectMessage(containsString('Download of Piper go binary failed'))
        piperGoUtils.unstashPiperBin()
    }

    @Test
    void testDownloadFailedWithError() {
        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib-os', version: 'notAvailable']]}

        shellCallRule.setReturnValue('[ -x ./piper ]', 1)

        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        exception.expectMessage(containsString('Download of Piper go binary failed'))
        piperGoUtils.unstashPiperBin()
    }
}
