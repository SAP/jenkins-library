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
            changeDocumentLabel: "ChangeDocument:",
            transportRequestLabel: "TransportRequest:",
        ]
    ]

    private Map getChangeDocumentIdReceivedParameters = [:]
    private Map getTransportRequestIdReceivedParameters = [:]

    @Before
    public void setup() {
        getChangeDocumentIdReceivedParameters.clear()
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

        String getChangeDocumentId(
                String from,
                String to,
                String label,
                String format
        ) {
            getChangeDocumentIdReceivedParameters['from'] = from
            getChangeDocumentIdReceivedParameters['to'] = to
            getChangeDocumentIdReceivedParameters['label'] = label
            getChangeDocumentIdReceivedParameters['format'] = format
            return '001'
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

    public void changeDocumentIdViaCommitHistoryTest() {

        def changeDocumentId = StepHelpers.getChangeDocumentId(cm, nullScript, params)

        assert changeDocumentId == '001'
        assert getChangeDocumentIdReceivedParameters ==
        [
            from: 'HEAD~1',
            to: 'HEAD',
            label: 'ChangeDocument:',
            format: '%b'
        ]

        // We cache the value. Otherwise we have to retrieve it each time from the
        // commit history.
        assert nullScript.commonPipelineEnvironment.getChangeDocumentId() == '001'
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

    public void changeDocumentIdViaCommonPipelineEnvironmentTest() {

        nullScript.commonPipelineEnvironment.setChangeDocumentId('002')
        def transportRequestId = StepHelpers.getChangeDocumentId(cm, nullScript, params)

        assert transportRequestId == '002'

        // getChangeDocumentId gets not called on ChangeManagement util class
        // in this case.
        assert getChangeDocumentIdReceivedParameters == [:]
    }

    @Test
    public void changeDocumentIdViaParametersTest() {

        params << [changeDocumentId: '003']

        def transportRequestId = StepHelpers.getChangeDocumentId(cm, nullScript, params)

        assert transportRequestId == '003'

        // In case we get the change document id via parameters we do not cache it
        // Caller knows the change document id anyway. So the caller can provide it with
        // each call.
        assert nullScript.commonPipelineEnvironment.getChangeDocumentId() == null

        // getChangeDocumentId gets not called on ChangeManagement util class
        // in this case.
        assert getChangeDocumentIdReceivedParameters == [:]
    }

    @Test
    public void changeDocumentIdNotProvidedTest() {

        jlr.expect('[WARN] Cannot retrieve changeDocumentId from commit history')

        def cm = new ChangeManagement(nullScript) {
            String getChangeDocumentId(
                    String from,
                    String to,
                    String label,
                    String format
            ) {
                throw new ChangeManagementException('Cannot retrieve change document ids')
            }
        }

        def changeDocumentId = StepHelpers.getChangeDocumentId(cm, nullScript, params)

        assert changeDocumentId == null
    }
}
