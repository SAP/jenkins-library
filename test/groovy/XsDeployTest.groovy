import com.sap.piper.DefaultValueCache
import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.containsInAnyOrder
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.nullValue
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers
import org.hamcrest.core.IsNull
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

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

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*getConfig.*--contextConfig.*', '{"dockerImage": "xs", "dockerPullImage": false, "credentialsId":"myCreds"}')

        // what we set on the shell rule and on the null script is the same. We read it on the groovy level, and also via go getConfig, hence we need it twice.
        nullScript.commonPipelineEnvironment.configuration = [steps: [xsDeploy: [mode: 'BG_DEPLOY', action: 'NONE', apiUrl: 'https://example.org/xs', org: 'myOrg', space: 'mySpace']]]
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'getConfig.* (?!--contextConfig)', '{"mode": "BG_DEPLOY", "action": "NONE", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        nullScript.commonPipelineEnvironment.xsDeploymentId = null

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
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

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'getConfig.* (?!--contextConfig)', '{"mode": "DOES_NOT_EXIST", "action": "NONE", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')
        nullScript.commonPipelineEnvironment.configuration = [steps: [xsDeploy: [mode: 'DOES_NOT_EXIST', action: 'NONE', apiUrl: 'https://example.org/xs', org: 'myOrg', space: 'mySpace']]]

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils,
        )
    }

    @Test
    public void testDeployableViaCPE() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*xsDeploy .*', ' ')

        nullScript.commonPipelineEnvironment.mtarFilePath = "my.mtar"

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

        assertThat(shellRule.shell,
                new CommandLineMatcher()
                    .hasProlog('#!/bin/bash ./piper xsDeploy')
                    // explicitly provided, it is not contained in project config.
                    .hasOption('mtaPath', 'my.mtar'))
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

        assertThat(writeFileRule.files.keySet(), containsInAnyOrder(
            '.pipeline/additionalConfigs/default_pipeline_environment.yml',
            '.pipeline/metadata/xsDeploy.yaml',
            ))

        assertThat(dockerRule.dockerParams.dockerImage, equalTo('xs'))
        assertThat(dockerRule.dockerParams.dockerPullImage, equalTo(false))

        assertThat(shellRule.shell,
            allOf(
                new CommandLineMatcher()
                    .hasProlog('./piper version'),
                new CommandLineMatcher()
                    .hasProlog('./piper getConfig')
                    .hasArgument('--contextConfig'),
                new CommandLineMatcher()
                    .hasProlog('./piper getConfig --stepMetadata \'.pipeline/metadata/xsDeploy.yaml\''),
                new CommandLineMatcher()
                    .hasProlog('#!/bin/bash ./piper xsDeploy --defaultConfig ".pipeline/additionalConfigs/default_pipeline_environment.yml" --username \\$\\{USERNAME\\} --password \\$\\{PASSWORD\\}'),
                not(new CommandLineMatcher()
                    .hasProlog('#!/bin/bash ./piper xsDeploy')
                    .hasOption('operationId', '1234'))
            )
        )

        assertThat(lockRule.getLockResources(), contains('xsDeploy:https://example.org/xs:myOrg:mySpace'))
    }

    @Test
    public void testBlueGreenDeployResume() {

        nullScript.commonPipelineEnvironment.xsDeploymentId = '1234'

        nullScript.commonPipelineEnvironment.configuration = [steps: [xsDeploy: [mode: 'BG_DEPLOY', action: 'RESUME', apiUrl: 'https://example.org/xs', org: 'myOrg', space: 'mySpace']]]
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'getConfig.* (?!--contextConfig)', '{"mode": "BG_DEPLOY", "action": "RESUME", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils
        )

        assertThat(shellRule.shell,
            new CommandLineMatcher()
                .hasProlog('#!/bin/bash ./piper xsDeploy')
                .hasOption('operationId', '1234')
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

        nullScript.commonPipelineEnvironment.configuration = [steps: [xsDeploy: [mode: 'BG_DEPLOY', action: 'RESUME', apiUrl: 'https://example.org/xs', org: 'myOrg', space: 'mySpace']]]
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'getConfig.* (?!--contextConfig)', '{"mode": "BG_DEPLOY", "action": "RESUME", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        assertThat(nullScript.commonPipelineEnvironment.xsDeploymentId, nullValue())

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils,
            failOnError: true,
        )
    }

    @Test
    public void testBlueGreenDeployResumeOperationIdViaSignature() {

        // this happens in case we would like to complete a deployment without having a (successful) deployments before.

        nullScript.commonPipelineEnvironment.configuration = [steps: [xsDeploy: [mode: 'BG_DEPLOY', action: 'RESUME', apiUrl: 'https://example.org/xs', org: 'myOrg', space: 'mySpace']]]
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, 'getConfig.* (?!--contextConfig)', '{"mode": "BG_DEPLOY", "action": "RESUME", "apiUrl": "https://example.org/xs", "org": "myOrg", "space": "mySpace"}')

        assertThat(nullScript.commonPipelineEnvironment.xsDeploymentId, nullValue())

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils,
            failOnError: true,
            operationId: '1357'
        )

        assertThat(shellRule.shell,
            new CommandLineMatcher()
                .hasProlog('#!/bin/bash ./piper xsDeploy')
                .hasOption('operationId', '1357')
        )
    }

    @Test
    public void testDockerParamsViaProjectConfig() {


        nullScript.commonPipelineEnvironment.configuration = [steps:
            [xsDeploy:
                [
                    dockerImage: 'xs1',
                    dockerPullImage: true
                ]
            ]
        ]

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils
        )

        // 'xs' provided on the context config is superseded by the value set in the project
        assertThat(dockerRule.dockerParams.dockerImage, equalTo('xs1'))
        assertThat(dockerRule.dockerParams.dockerPullImage, equalTo(true))
    }

    @Test
    public void testDockerParamsViaProjectConfigNested() {


        nullScript.commonPipelineEnvironment.configuration = [steps:
            [xsDeploy:
                [
                    docker: [
                        dockerImage: 'xs1',
                        dockerPullImage: true
                    ]
                ]
            ]
        ]

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils
        )

        // 'xs' provided on the context config is superseded by the value set in the project
        assertThat(dockerRule.dockerParams.dockerImage, equalTo('xs1'))
        assertThat(dockerRule.dockerParams.dockerPullImage, equalTo(true))
    }

    @Test
    public void testDockerParamsViaSignature() {


        nullScript.commonPipelineEnvironment.configuration = [steps:
            [xsDeploy:
                [
                    dockerImage: 'xs1'
                ]
            ]
        ]

        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils,
            dockerImage: 'xs2',
        )

        // 'xs' provided on the context config and 'xs1' provided by project config
        // is superseded by the value set in the project
        assertThat(dockerRule.dockerParams.dockerImage, equalTo('xs2'))
    }

    @Test
    public void testAdditionalCustomConfigLayers() {

        def resources = ['a.yml': '- x: y}', 'b.yml' : '- a: b}']

        helper.registerAllowedMethod('libraryResource', [String], {

            r ->

            def resource = resources[r]
            if(resource) return resource

            File res = new File(new File('resources'), r)
            if (res.exists()) {
                return res.getText()
            }

            throw new RuntimeException("Resource '${r}' not found.")
        })

        assertThat(nullScript.commonPipelineEnvironment.xsDeploymentId, nullValue())

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*xsDeploy .*', '{"operationId": "1234"}')

        DefaultValueCache.createInstance([:], ['a.yml', 'b.yml'])

        goUtils = new PiperGoUtils(null) {
            void unstashPiperBin() {
            }
        }
        stepRule.step.xsDeploy(
            script: nullScript,
            piperGoUtils: goUtils
        )

        assertThat(writeFileRule.files.keySet(), containsInAnyOrder(
            '.pipeline/additionalConfigs/a.yml',
            '.pipeline/additionalConfigs/b.yml',
            '.pipeline/additionalConfigs/default_pipeline_environment.yml',
            '.pipeline/metadata/xsDeploy.yaml',
            ))

        assertThat(shellRule.shell,
            allOf(
                new CommandLineMatcher()
                    .hasProlog('./piper getConfig')
                    .hasArgument('--contextConfig')
                    .hasArgument('--defaultConfig ".pipeline/additionalConfigs/b.yml" ".pipeline/additionalConfigs/a.yml" ".pipeline/additionalConfigs/default_pipeline_environment.yml"'),
                new CommandLineMatcher()
                    .hasProlog('./piper getConfig --stepMetadata \'.pipeline/metadata/xsDeploy.yaml\''),
            )
        )
    }
}
