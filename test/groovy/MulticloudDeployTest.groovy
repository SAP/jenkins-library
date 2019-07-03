import com.sap.piper.DefaultValueCache
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils

import hudson.AbortException

import org.junit.Assert
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.*


class MulticloudDeployTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException().none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsMockStepRule neoDeployRule = new JenkinsMockStepRule(this, 'neoDeploy')
    private JenkinsMockStepRule cloudFoundryDeployRule = new JenkinsMockStepRule(this, 'cloudFoundryDeploy')
    private JenkinsReadMavenPomRule readMavenPomRule = new JenkinsReadMavenPomRule(this, 'test/resources/deploy')

    private Map neo1 = [:]
    private Map neo2 = [:]
    private Map cloudFoundry1 = [:]
    private Map cloudFoundry2 = [:]

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(new JenkinsFileExistsRule(this))
        .around(thrown)
        .around(stepRule)
        .around(neoDeployRule)
        .around(cloudFoundryDeployRule)
        .around(readMavenPomRule)

   private Map neoDeployParameters = [:]
   private Map cloudFoundryDeployParameters = [:]

    @Before
    void init() {

        neo1 = [
                  host: 'test.deploy.host1.com',
                  account: 'trialuser1',
                  credentialsId: 'credentialsId1'
              ]

        neo2 = [
                  host: 'test.deploy.host2.com',
                  account: 'trialuser2',
                  credentialsId: 'credentialsId2'
              ]

        cloudFoundry1 = [
                           appName:'testAppName1',
                           manifest: 'test.yml',
                           org: 'testOrg1',
                           space: 'testSpace1',
                           credentialsId: 'cfCredentialsId1'
                       ]

        cloudFoundry2 = [
                            appName:'testAppName2',
                            manifest: 'test.yml',
                            org: 'testOrg2',
                            space: 'testSpace2',
                            credentialsId: 'cfCredentialsId2'
                        ]

        DefaultValueCache.createInstance(loadDefaultPipelineEnvironment(),
        [
            general: [
                neoTargets: [
                    neo1, neo2
                ],
                cfTargets: [
                    cloudFoundry1, cloudFoundry2
                ]
            ],
            stages: [
                acceptance: [
                    org: 'testOrg',
                    space: 'testSpace',
                    deployUser: 'testUser'
                ]
            ],
            steps: [
                cloudFoundryDeploy: [
                    deployTool: 'cf_native',
                    deployType: 'blue-green',
                    keepOldInstance: true,
                    cf_native: [
                        dockerImage: 's4sdk/docker-cf-cli',
                        dockerWorkspace: '/home/piper'
                    ]
                ]
            ]
        ])
    }

    @Test
    void errorNoTargetsDefined() {

        DefaultValueCache.getInstance().getProjectConfig().general.neoTargets = []
        DefaultValueCache.getInstance().getProjectConfig().general.cfTargets = []

        thrown.expect(Exception)
        thrown.expectMessage('Deployment skipped because no targets defined!')

        stepRule.step.multicloudDeploy(
            script: nullScript,
            stage: 'test'
        )
    }

    @Test
    void errorNoSourceForNeoDeploymentTest() {

        DefaultValueCache.getInstance().getProjectConfig().general.neoTargets = [neo1]
        DefaultValueCache.getInstance().getProjectConfig().general.cfTargets = []

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR source')

        stepRule.step.multicloudDeploy(
            script: nullScript,
            stage: 'test'
        )
    }

    @Test
    void neoDeploymentTest() {

        DefaultValueCache.getInstance().getProjectConfig().general.neoTargets = [neo1]
        DefaultValueCache.getInstance().getProjectConfig().general.cfTargets = []

        stepRule.step.multicloudDeploy(
            script: nullScript,
            stage: 'test',
            source: 'file.mtar'
        )

        assert neoDeployRule.hasParameter('warAction', 'deploy')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neo1)
    }

    @Test
    void neoRollingUpdateTest() {

        DefaultValueCache.getInstance().getProjectConfig().general.neoTargets = []
        DefaultValueCache.getInstance().getProjectConfig().general.cfTargets = []

        def neoParam = [
                    host: 'test.param.deploy.host.com',
                    account: 'trialparamNeoUser',
                    credentialsId: 'paramNeoCredentialsId'
                ]

        stepRule.step.multicloudDeploy(
            script: nullScript,
            stage: 'test',
            neoTargets: [neoParam],
            source: 'file.mtar',
            enableZeroDowntimeDeployment: true
        )

        assert neoDeployRule.hasParameter('warAction', 'rolling-update')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neoParam)
    }

    @Test
    void cfDeploymentTest() {

        DefaultValueCache.getInstance().getProjectConfig().general.neoTargets = []
        DefaultValueCache.getInstance().getProjectConfig().general.cfTargets = []

        def cloudFoundry = [
                    appName:'paramTestAppName',
                    manifest: 'test.yml',
                    org: 'paramTestOrg',
                    space: 'paramTestSpace',
                    credentialsId: 'paramCfCredentialsId'
                ]

        stepRule.step.multicloudDeploy([
            script: nullScript,
            stage: 'acceptance',
            cfTargets: [cloudFoundry]
        ])

        assert cloudFoundryDeployRule.hasParameter('deployType', 'standard')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')
    }

    @Test
    void cfBlueGreenDeploymentTest() {

        DefaultValueCache.getInstance().getProjectConfig().general.neoTargets = []
        DefaultValueCache.getInstance().getProjectConfig().general.cfTargets = [cloudFoundry1]

        stepRule.step.multicloudDeploy([
            script: nullScript,
            stage: 'acceptance',
            enableZeroDowntimeDeployment: true
        ])

        assert cloudFoundryDeployRule.hasParameter('deployType', 'blue-green')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry1)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')
    }

    @Test
    void multicloudDeploymentTest() {

        stepRule.step.multicloudDeploy([
            script: nullScript,
            stage: 'acceptance',
            enableZeroDowntimeDeployment: true,
            source: 'file.mtar'
        ])

        assert neoDeployRule.hasParameter('warAction', 'rolling-update')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neo1)

        assert neoDeployRule.hasParameter('warAction', 'rolling-update')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neo2)

        assert cloudFoundryDeployRule.hasParameter('deployType', 'blue-green')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry1)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')

        assert cloudFoundryDeployRule.hasParameter('deployType', 'blue-green')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry2)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')
    }

}
