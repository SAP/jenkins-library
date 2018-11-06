package com.sap.piper.cm

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.Rules

class StepHelpersTest extends BasePiperTest {

    // Configuration is not checked by the tests here.
    // We simply assume it fits. It is the duty of the
    // step related tests to ensure the configuration is valid.
    def params = [changeManagement:
        [git: [
                from: 'HEAD~1',
                to: 'HEAD',
                format: '%b'
            ],
            transportRequestLabel: "TransportRequest:"
        ]
    ]

    private Map getTransportRequestIdReceivedParameters = [:]

    @Before
    public void setup() {
        getTransportRequestIdReceivedParameters.clear()
    }

    JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
                                  .around(jlr)

    private ChangeManagement cm = new ChangeManagement(nullScript) {
        String getTransportRequestId(
                String from,
                String to,
                String label,
                String format
        ) {
            getTransportRequestIdReceivedParameters['from'] = from
            getTransportRequestIdReceivedParameters['to'] = to
            getTransportRequestIdReceivedParameters['label'] = label
            getTransportRequestIdReceivedParameters['format'] = format
            return '097'
        }
    }

    @Test
    public void transportRequestIdViaCommitHistoryTest() {

        def transportRequestId = StepHelpers.getTransportRequestId(cm, nullScript, params)

        assert transportRequestId == '097'
        assert getTransportRequestIdReceivedParameters ==
        [
            from: 'HEAD~1',
            to: 'HEAD',
            label: 'TransportRequest:',
            format: '%b'
        ]

        // We cache the value. Otherwise we have to retrieve it each time from the
        // commit history.
        assert nullScript.commonPipelineEnvironment.getTransportRequestId() == '097'

    }

    @Test
    public void transportRequestIdViaCommonPipelineEnvironmentTest() {

        nullScript.commonPipelineEnvironment.setTransportRequestId('098')
        def transportRequestId = StepHelpers.getTransportRequestId(cm, nullScript, params)

        assert transportRequestId == '098'

        // getTransportRequestId gets not called on ChangeManagement util class
        // in this case.
        assert getTransportRequestIdReceivedParameters == [:]
    }

    @Test
    public void transportRequestIdViaParametersTest() {

        params << [transportRequestId: '099']

        def transportRequestId = StepHelpers.getTransportRequestId(cm, nullScript, params)

        assert transportRequestId == '099'

        // In case we get the transport request id via parameters we do not cache it
        // Caller knows the transport request id anyway. So the caller can provide it with
        // each call.
        assert nullScript.commonPipelineEnvironment.getTransportRequestId() == null

        // getTransportRequestId gets not called on ChangeManagement util class
        // in this case.
        assert getTransportRequestIdReceivedParameters == [:]
    }

    @Test
    public void transportRequestIdNotProvidedTest() {

        jlr.expect('Cannot retrieve transportRequestId from commit history')

        def cm = new ChangeManagement(nullScript) {
        String getTransportRequestId(
                String from,
                String to,
                String label,
                String format
        )   {
                throw new ChangeManagementException('Cannot retrieve transport request id')
            }
        }

        def transportRequestId = StepHelpers.getTransportRequestId(cm, nullScript, params)

        assert transportRequestId == null
    }
}
