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
    private JenkinsMockStepRule cloudFoundryCreateServiceRule = new JenkinsMockStepRule(this, 'cloudFoundryCreateService')
    private JenkinsReadMavenPomRule readMavenPomRule = new JenkinsReadMavenPomRule(this, 'test/resources/deploy')

    private Map neo1 = [:]
    private Map neo2 = [:]
    private Map cloudFoundry1 = [:]
    private Map cloudFoundry2 = [:]

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(stepRule)
        .around(neoDeployRule)
        .around(cloudFoundryDeployRule)
        .around(cloudFoundryCreateServiceRule)
        .around(readMavenPomRule)

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

        nullScript.commonPipelineEnvironment.configuration = [
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
                        dockerImage: 'ppiper/cf-cli',
                        dockerWorkspace: '/home/piper'
                    ]
                ]
            ]
        ]

        helper.registerAllowedMethod('echo', [CharSequence.class], {})

    }

    @Test
    void errorNoTargetsDefined() {

        nullScript.commonPipelineEnvironment.configuration.general.neoTargets = []
        nullScript.commonPipelineEnvironment.configuration.general.cfTargets = []

        thrown.expect(Exception)
        thrown.expectMessage('Deployment skipped because no targets defined!')

        stepRule.step.multicloudDeploy(
            script: nullScript
        )
    }

    @Test
    void errorNoSourceForNeoDeploymentTest() {

        nullScript.commonPipelineEnvironment.configuration.general.neoTargets = [neo1]
        nullScript.commonPipelineEnvironment.configuration.general.cfTargets = []

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR source')

        stepRule.step.multicloudDeploy(
            script: nullScript
        )
    }

    @Test
    void neoDeploymentTest() {

        nullScript.commonPipelineEnvironment.configuration.general.neoTargets = [neo1]
        nullScript.commonPipelineEnvironment.configuration.general.cfTargets = []

        stepRule.step.multicloudDeploy(
            script: nullScript,
            source: 'file.mtar'
        )

        assert neoDeployRule.hasParameter('script', nullScript)
        assert neoDeployRule.hasParameter('warAction', 'deploy')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neo1)
    }

    @Test
    void neoRollingUpdateTest() {

        nullScript.commonPipelineEnvironment.configuration.general.neoTargets = []
        nullScript.commonPipelineEnvironment.configuration.general.cfTargets = []

        def neoParam = [
                    host: 'test.param.deploy.host.com',
                    account: 'trialparamNeoUser',
                    credentialsId: 'paramNeoCredentialsId'
                ]

        stepRule.step.multicloudDeploy(
            script: nullScript,
            neoTargets: [neoParam],
            source: 'file.mtar',
            enableZeroDowntimeDeployment: true
        )

        assert neoDeployRule.hasParameter('script', nullScript)
        assert neoDeployRule.hasParameter('warAction', 'rolling-update')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neoParam)
    }

    @Test
    void cfDeploymentTest() {

        nullScript.commonPipelineEnvironment.configuration.general.neoTargets = []
        nullScript.commonPipelineEnvironment.configuration.general.cfTargets = []

        def cloudFoundry = [
                    appName:'paramTestAppName',
                    manifest: 'test.yml',
                    org: 'paramTestOrg',
                    space: 'paramTestSpace',
                    credentialsId: 'paramCfCredentialsId'
                ]

        stepRule.step.multicloudDeploy([
            script: nullScript,
            cfTargets: [cloudFoundry]
        ])

        assert cloudFoundryDeployRule.hasParameter('script', nullScript)
        assert cloudFoundryDeployRule.hasParameter('deployType', 'standard')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')
    }

    @Test
    void cfBlueGreenDeploymentTest() {

        nullScript.commonPipelineEnvironment.configuration.general.neoTargets = []
        nullScript.commonPipelineEnvironment.configuration.general.cfTargets = [cloudFoundry1]

        stepRule.step.multicloudDeploy([
            script: nullScript,
            enableZeroDowntimeDeployment: true
        ])

        assert cloudFoundryDeployRule.hasParameter('script', nullScript)
        assert cloudFoundryDeployRule.hasParameter('deployType', 'blue-green')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry1)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')
    }

    @Test
    void multicloudDeploymentTest() {

        stepRule.step.multicloudDeploy([
            script: nullScript,
            enableZeroDowntimeDeployment: true,
            source: 'file.mtar'
        ])

        assert neoDeployRule.hasParameter('script', nullScript)
        assert neoDeployRule.hasParameter('warAction', 'rolling-update')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neo1)

        assert neoDeployRule.hasParameter('script', nullScript)
        assert neoDeployRule.hasParameter('warAction', 'rolling-update')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neo2)

        assert cloudFoundryDeployRule.hasParameter('script', nullScript)
        assert cloudFoundryDeployRule.hasParameter('deployType', 'blue-green')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry1)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')

        assert cloudFoundryDeployRule.hasParameter('script', nullScript)
        assert cloudFoundryDeployRule.hasParameter('deployType', 'blue-green')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry2)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')
    }

    @Test
    void multicloudParallelDeploymentTest() {

        stepRule.step.multicloudDeploy([
            script: nullScript,
            enableZeroDowntimeDeployment: true,
            parallelExecution: true,
            source: 'file.mtar'
        ])

        assert neoDeployRule.hasParameter('script', nullScript)
        assert neoDeployRule.hasParameter('warAction', 'rolling-update')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neo1)

        assert neoDeployRule.hasParameter('script', nullScript)
        assert neoDeployRule.hasParameter('warAction', 'rolling-update')
        assert neoDeployRule.hasParameter('source', 'file.mtar')
        assert neoDeployRule.hasParameter('neo', neo2)

        assert cloudFoundryDeployRule.hasParameter('script', nullScript)
        assert cloudFoundryDeployRule.hasParameter('deployType', 'blue-green')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry1)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')

        assert cloudFoundryDeployRule.hasParameter('script', nullScript)
        assert cloudFoundryDeployRule.hasParameter('deployType', 'blue-green')
        assert cloudFoundryDeployRule.hasParameter('cloudFoundry', cloudFoundry2)
        assert cloudFoundryDeployRule.hasParameter('mtaPath', nullScript.commonPipelineEnvironment.mtarFilePath)
        assert cloudFoundryDeployRule.hasParameter('deployTool', 'cf_native')
    }

    @Test
    void 'cfCreateServices calls cloudFoundryCreateService step with correct parameters'() {
        stepRule.step.multicloudDeploy([
            script          : nullScript,
            cfCreateServices: [[apiEndpoint: 'http://mycf.org', serviceManifest: 'services-manifest.yml', manifestVariablesFiles: 'vars.yml', space: 'PerformanceTests', org: 'MyOrg', credentialsId: 'MyCred']],
            source          : 'file.mtar'
        ])

        assert cloudFoundryCreateServiceRule.hasParameter('cloudFoundry', [
            serviceManifest       : 'services-manifest.yml',
            space                 : 'PerformanceTests',
            org                   : 'MyOrg',
            credentialsId         : 'MyCred',
            apiEndpoint           : 'http://mycf.org',
            manifestVariablesFiles: 'vars.yml'
        ])
    }

    @Test
    void 'cfCreateServices with parallelTestExecution defined in compatible parameter - must run in parallel'() {
        def closureRun = null

        helper.registerAllowedMethod('error', [String.class], { s->
            if (s == "Deployment skipped because no targets defined!") {
                // This error is ok because in this test we're not interested in the deployment
            } else {
                throw new RuntimeException("Unexpected error in test with message: ${s}")
            }
        })
        helper.registerAllowedMethod('parallel', [Map.class], {m -> closureRun = m})

        nullScript.commonPipelineEnvironment.configuration.general['features'] = [parallelTestExecution: true]

        stepRule.step.multicloudDeploy([
            script: nullScript,
            cfCreateServices: [[serviceManifest: 'services-manifest.yml', space: 'PerformanceTests', org: 'foo', credentialsId: 'foo']],
            source: 'file.mtar'
        ])

        assert closureRun != null
    }

    @Test
    void 'cfCreateServices with parallelExecution defined - must run in parallel'() {
        def closureRun = null

        helper.registerAllowedMethod('error', [String.class], { s->
            if (s == "Deployment skipped because no targets defined!") {
                // This error is ok because in this test we're not interested in the deployment
            } else {
                throw new RuntimeException("Unexpected error in test with message: ${s}")
            }
        })
        helper.registerAllowedMethod('parallel', [Map.class], {m -> closureRun = m})

        nullScript.commonPipelineEnvironment.configuration.general = [parallelExecution: true]

        stepRule.step.multicloudDeploy([
            script: nullScript,
            cfCreateServices: [[serviceManifest: 'services-manifest.yml', space: 'PerformanceTests', org: 'foo', credentialsId: 'foo']],
            source: 'file.mtar'
        ])

        assert closureRun != null
    }

}
