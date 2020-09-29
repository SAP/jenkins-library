import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import com.sap.piper.Utils

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class SpinnakerTriggerPipelineTest extends BasePiperTest {
    private ExpectedException exception = new ExpectedException().none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule logginRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(exception)
        .around(shellRule)
        .around(logginRule)
        .around(readJsonRule)
        .around(stepRule)

    class EnvMock {
        def STAGE_NAME = 'testStage'
        Map getEnvironment() {
            return [key1: 'value1', key2: 'value2']
        }
    }

    def credentialFileList = []
    def timeout = 0

    @Before
    void init() {
        binding.setVariable('env', new EnvMock())

        credentialFileList = []
        helper.registerAllowedMethod('file', [Map], { m ->
            credentialFileList.add(m)
            return m
        })

        Map credentialFileNames = [
            'spinnaker-client-certificate': 'clientCert.file',
            'spinnaker-client-key': 'clientKey.file'
        ]

        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            l.each { fileCredentials ->
                binding.setProperty(fileCredentials.variable, credentialFileNames[fileCredentials.credentialsId])
            }
            try {
                c()
            } finally {
                l.each { fileCredentials ->
                    binding.setProperty(fileCredentials.variable, null)
                }
            }
        })

        helper.registerAllowedMethod('timeout', [Integer.class, Closure.class] , {i, body ->
            timeout = i
            return body()
        })

        //not sure where this comes from!
        helper.registerAllowedMethod('waitUntil', [Closure.class], {body ->
            List responseStatus = ['RUNNING', 'PAUSED', 'NOT_STARTED']
            while (!body()) {
                //take another round with a different response status
                responseStatus.each {status ->
                    shellRule.setReturnValue('curl -X GET https://spinnakerTest.url/testRef --silent --cert $clientCertificate --key $clientKey', "{\"status\": \"${status}\"}")
                    shellRule.setReturnValue('curl -X GET https://spinnakerTest.url/testRef --verbose --cert $clientCertificate --key $clientKey', "{\"status\": \"${status}\"}")
                }
            }
        })

        shellRule.setReturnValue('curl -H \'Content-Type: application/json\' -X POST -d \'{"parameters":{"param1":"val1"}}\' --silent --cert $clientCertificate --key $clientKey https://spinnakerTest.url/pipelines/spinnakerTestApp/spinnakerTestPipeline', '{"ref": "/testRef"}')
        shellRule.setReturnValue('curl -H \'Content-Type: application/json\' -X POST -d \'{"parameters":{"param1":"val1"}}\' --verbose --cert $clientCertificate --key $clientKey https://spinnakerTest.url/pipelines/spinnakerTestApp/spinnakerTestPipeline', '{"ref": "/testRef"}')
        shellRule.setReturnValue('curl -X GET https://spinnakerTest.url/testRef --silent --cert $clientCertificate --key $clientKey', '{"status": "SUCCEEDED"}')
        shellRule.setReturnValue('curl -X GET https://spinnakerTest.url/testRef --verbose --cert $clientCertificate --key $clientKey', '{"status": "SUCCEEDED"}')

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testDefaults() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                spinnakerGateUrl: 'https://spinnakerTest.url',
                spinnakerApplication: 'spinnakerTestApp',
                verbose: true
            ],
            stages: [
                testStage: [
                    spinnakerPipeline: 'spinnakerTestPipeline',
                    pipelineParameters: [param1: 'val1']
                ]
            ]
        ]
        stepRule.step.spinnakerTriggerPipeline(
            script: nullScript
        )

        assertThat(timeout, is(60))

        assertThat(logginRule.log, containsString('Triggering Spinnaker pipeline with parameters:'))
        assertThat(logginRule.log, containsString('Spinnaker pipeline /testRef triggered, waiting for the pipeline to finish'))
        assertThat(credentialFileList,
            hasItem(
                allOf(
                    hasEntry('credentialsId', 'spinnaker-client-key'),
                    hasEntry('variable', 'clientKey')
                )
            )
        )
        assertThat(credentialFileList,
            hasItem(
                allOf(
                    hasEntry('credentialsId', 'spinnaker-client-certificate'),
                    hasEntry('variable', 'clientCertificate')
                )
            )
        )
    }

    @Test
    void testDisabledPipelineCheck() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                spinnaker: [
                    gateUrl: 'https://spinnakerTest.url',
                    application: 'spinnakerTestApp'
                ]
            ],
            stages: [
                testStage: [
                    pipelineNameOrId: 'spinnakerTestPipeline',
                    pipelineParameters: [param1: 'val1']
                ]
            ]
        ]
        stepRule.step.spinnakerTriggerPipeline(
            script: nullScript,
            timeout: 0
        )

        assertThat(logginRule.log, containsString('Exiting without waiting for Spinnaker pipeline result.'))
        assertThat(timeout, is(0))
    }

    @Test
    void testTriggerFailure() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                spinnakerGateUrl: 'https://spinnakerTest.url',
                spinnakerApplication: 'spinnakerTestApp'
            ],
            stages: [
                testStage: [
                    spinnakerPipeline: 'spinnakerTestPipeline'
                ]
            ]
        ]

        shellRule.setReturnValue('curl -H \'Content-Type: application/json\' -X POST --silent --cert $clientCertificate --key $clientKey https://spinnakerTest.url/pipelines/spinnakerTestApp/spinnakerTestPipeline', '{}')

        exception.expectMessage('Failed to trigger Spinnaker pipeline')
        stepRule.step.spinnakerTriggerPipeline(
            script: nullScript,
        )
    }

    @Test
    void testPipelineFailure() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                spinnakerGateUrl: 'https://spinnakerTest.url',
                spinnakerApplication: 'spinnakerTestApp',
                verbose: true
            ],
            stages: [
                testStage: [
                    spinnakerPipeline: 'spinnakerTestPipeline',
                    pipelineParameters: [param1: 'val1']
                ]
            ]
        ]

        shellRule.setReturnValue('curl -X GET https://spinnakerTest.url/testRef --silent --cert $clientCertificate --key $clientKey', '{"status": "FAILED"}')
        shellRule.setReturnValue('curl -X GET https://spinnakerTest.url/testRef --verbose --cert $clientCertificate --key $clientKey', '{"status": "FAILED"}')

        exception.expect(hudson.AbortException)
        exception.expectMessage('Spinnaker pipeline failed with FAILED')
        try {
            stepRule.step.spinnakerTriggerPipeline(script: nullScript)
        } finally {
            assertThat(logginRule.log, containsString('Full Spinnaker response = '))
        }
    }
}
