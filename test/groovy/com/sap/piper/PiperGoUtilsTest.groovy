package com.sap.piper

import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsHttpRequestRule
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.PluginMock
import util.Rules

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

class PiperGoUtilsTest extends BasePiperTest {

    public ExpectedException exception = ExpectedException.none()
    public JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    public JenkinsHttpRequestRule httpRequestRule = new JenkinsHttpRequestRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(shellCallRule)
        .around(exception)
        .around(loggingRule)
        .around(httpRequestRule)

    @Before
    void init() {
        helper.registerAllowedMethod("retry", [Integer, Closure], null)
        JenkinsUtils.metaClass.static.isPluginActive = { def s -> new PluginMock(s).isActive() }
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
        piperGoUtils.metaClass.getLibrariesInfo = { -> return [[name: 'piper-lib-os', version: 'master']] }

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        httpRequestRule.mockUrl("https://github.com/SAP/jenkins-library/releases/latest/download/piper_master", {})

        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(2))
        assertThat(shellCallRule.shell[0].toString(), is('chmod +x piper'))
        assertThat(shellCallRule.shell[1].toString(), is('./piper version'))

        assertThat(httpRequestRule.requests.size(), is(1))
        assertThat(httpRequestRule.requests[0].url, is('https://github.com/SAP/jenkins-library/releases/latest/download/piper_master'))
    }

    @Test
    void testUnstashPiperBinNonMaster() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = { -> return [[name: 'piper-lib-os', version: 'testTag']] }

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })


        httpRequestRule.mockUrl("https://github.com/SAP/jenkins-library/releases/download/testTag/piper", {})

        piperGoUtils.unstashPiperBin()

        assertThat(shellCallRule.shell.size(), is(2))
        assertThat(shellCallRule.shell[0].toString(), is('chmod +x piper'))

        assertThat(httpRequestRule.requests.size(), is(1))
        assertThat(httpRequestRule.requests[0].url.toString(), is('https://github.com/SAP/jenkins-library/releases/download/testTag/piper'))

    }

    @Test
    void testUnstashPiperBinFallback() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = { -> return [[name: 'piper-lib-os', version: 'notAvailable']] }

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        httpRequestRule.mockUrl("https://github.com/SAP/jenkins-library/releases/download/notAvailable/piper", {
            throw new AbortException("Fail: the returned code 404 is not in the accepted range");
        })
        httpRequestRule.mockUrl("https://github.com/SAP/jenkins-library/releases/latest/download/piper_master", {})

        shellCallRule.setReturnValue('./piper version', "1.2.3")

        piperGoUtils.unstashPiperBin()

        assertThat(shellCallRule.shell.size(), is(2))
        assertThat(shellCallRule.shell[0].toString(), is('chmod +x piper'))
        assertThat(shellCallRule.shell[1].toString(), is ('./piper version'))

        assertThat(httpRequestRule.requests.size(), is(2))
        assertThat(httpRequestRule.requests[0].url.toString(), is('https://github.com/SAP/jenkins-library/releases/download/notAvailable/piper'))
        assertThat(httpRequestRule.requests[1].url.toString(), is('https://github.com/SAP/jenkins-library/releases/latest/download/piper_master'))
    }

    @Test
    void testDownloadFailedWithErrorCode() {
        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = { -> return [[name: 'piper-lib-os', version: 'notAvailable']] }

        httpRequestRule.mockUrl("https://github.com/SAP/jenkins-library/releases/download/notAvailable/piper", {
            throw new AbortException("Fail: the returned code 404 is not in the accepted range");
        })
        httpRequestRule.mockUrl("https://github.com/SAP/jenkins-library/releases/latest/download/piper_master", {
            throw new AbortException("Fail: the returned code 500 is not in the accepted range");
        })

        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        exception.expectMessage(containsString('Download of Piper go binary failed'))
        piperGoUtils.unstashPiperBin()
    }

    @Test
    void testDownloadFailedWithError() {
        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = { -> return [[name: 'piper-lib-os', version: 'notAvailable']] }

        helper.registerAllowedMethod('sh', [Map.class], { m -> throw new AbortException('download failed') })

        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        exception.expectMessage(containsString('Download of Piper go binary failed'))
        piperGoUtils.unstashPiperBin()
    }
}

