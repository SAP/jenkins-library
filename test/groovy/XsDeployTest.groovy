import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.nullValue
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers
import org.hamcrest.core.IsNull
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.PiperGoUtils

import hudson.AbortException
import util.BasePiperTest
import util.CommandLineMatcher
import util.JenkinsCredentialsRule
import util.JenkinsDockerExecuteRule
import util.JenkinsLockRule
import util.JenkinsReadJsonRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

class XsDeployTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLockRule lockRule = new JenkinsLockRule(this)
    private JenkinsDockerExecuteRule dockerRule = new JenkinsDockerExecuteRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)

    List env

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
                                        .around(new JenkinsReadYamlRule(this))
                                        .around(new JenkinsReadJsonRule(this))
                                        .around(stepRule)
                                        .around(dockerRule)
                                        .around(writeFileRule)
                                        .around(new JenkinsCredentialsRule(this)
                                            .withCredentials('myCreds', 'cred_xs', 'topSecret'))
                                        .around(lockRule)
                                        .around(shellRule)
                                        .around(thrown)

    private PiperGoUtils goUtils = new PiperGoUtils(null) {
        void unstashPiperBin() {
        }
    }

    @Before
    public void init() {
        helper.registerAllowedMethod('withEnv', [List, Closure], {l, c -> env = l;  c()})
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --contextConfig --stepMetadata.*', '{"dockerImage": "xs", "credentialsId":"myCreds"}')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --stepMetadata.*', '{"mode": "BG_DEPLOY", "action": "NONE", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')
        nullScript.commonPipelineEnvironment.xsDeploymentId = null
    }

    @Test
    public void testDeployFailed() {

        thrown.expect(AbortException)
        thrown.expectMessage('script returned exit code 1')

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*xsDeploy .*', { throw new AbortException('script returned exit code 1')})

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils,
        )
    }

    @Test
    public void testInvalidDeploymentModeProvided() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('No enum constant')

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --stepMetadata.*', '{"mode": "DOES_NOT_EXIST", "action": "NONE", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils,
        )
    }

    @Test
    public void testParametersViaSignature() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*xsDeploy .*', '{"operationId": "1234"}')

        stepRule.step.xsDeploy(
            script: nullScript,
            apiUrl: 'https://example.org/xs',
            org: 'myOrg',
            space: 'mySpace',
            credentialsId: 'myCreds',
            deployOpts: '-t 60',
            mtaPath: 'myApp.mta',
            mode: 'DEPLOY',
            action: 'NONE',
            piperGoUtils: goUtils
        )

        
        // nota bene: script and piperGoUtils are not contained in the json below.
        assertThat(env*.toString(), contains('PIPER_parametersJSON={"apiUrl":"https://example.org/xs","org":"myOrg","space":"mySpace","credentialsId":"myCreds","deployOpts":"-t 60","mtaPath":"myApp.mta","mode":"DEPLOY","action":"NONE"}'))
    }

    @Test
    public void testBlueGreenDeployInit() {

        //
        // Only difference between bg deploy and standard deploy is in the config.
        // The surrounding behavior is the same. Hence there is no dedicated test here
        // in the groovy layer for standard deploy
        //

        boolean unstashCalled

        assertThat(nullScript.commonPipelineEnvironment.xsDeploymentId, nullValue())
        
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*xsDeploy .*', '{"operationId": "1234"}')

        goUtils = new PiperGoUtils(null) {
            void unstashPiperBin() {
                unstashCalled = true
            }
        }
        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils
        )

        assertThat(unstashCalled, equalTo(true))

        assertThat(nullScript.commonPipelineEnvironment.xsDeploymentId, is('1234'))

        assertThat(writeFileRule.files.keySet(), contains('metadata/xsDeploy.yaml'))
        
        assertThat(dockerRule.dockerParams.dockerImage, equalTo('xs'))
        assertThat(dockerRule.dockerParams.dockerPullImage, equalTo(false))
        
        assertThat(shellRule.shell,
            allOf(
                new CommandLineMatcher()
                    .hasProlog('./piper version'),
                new CommandLineMatcher()
                    .hasProlog('./piper getConfig --contextConfig --stepMetadata \'metadata/xsDeploy.yaml\''),
                new CommandLineMatcher()
                    .hasProlog('./piper getConfig --stepMetadata \'metadata/xsDeploy.yaml\''),
                new CommandLineMatcher()
                    .hasProlog('#!/bin/bash ./piper xsDeploy --user \\$\\{USERNAME\\} --password \\$\\{PASSWORD\\}'),
                not(new CommandLineMatcher()
                    .hasProlog('#!/bin/bash ./piper xsDeploy --user \\$\\{USERNAME\\} --password \\$\\{PASSWORD\\}  --operationId'))
            )
        )

        assertThat(lockRule.getLockResources(), contains('xsDeploy:https://example.org/xs:myOrg:mySpace'))
    }

    @Test
    public void testBlueGreenDeployResume() {

        nullScript.commonPipelineEnvironment.xsDeploymentId = '1234'

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --stepMetadata.*', '{"mode": "BG_DEPLOY", "action": "RESUME", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils
        )

        assertThat(shellRule.shell,
            new CommandLineMatcher()
                .hasProlog('#!/bin/bash ./piper xsDeploy --user \\$\\{USERNAME\\} --password \\$\\{PASSWORD\\} --operationId 1234')
        )

        assertThat(lockRule.getLockResources(), contains('xsDeploy:https://example.org/xs:myOrg:mySpace'))
    }

    @Test
    public void testBlueGreenDeployResumeWithoutDeploymentId() {

        // this happens in case we would like to complete a deployment without having a (successful) deployments before.

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(
            allOf(
                containsString('No operationId provided'),
                containsString('Was there a deployment before?')))

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig --stepMetadata.*', '{"mode": "BG_DEPLOY", "action": "RESUME", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        assertThat(nullScript.commonPipelineEnvironment.xsDeploymentId, nullValue())

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils,
            failOnError: true,
        )
    }
}