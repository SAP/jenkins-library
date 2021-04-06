import com.sap.piper.JenkinsUtils
import hudson.AbortException
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsDockerExecuteRule
import util.JenkinsEnvironmentRule
import util.JenkinsFileExistsRule
import util.JenkinsLoggingRule
import util.JenkinsReadFileRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.stringContainsInOrder
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class CloudFoundryDeployTest extends BasePiperTest {

    private File tmpDir = File.createTempDir()
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, tmpDir.getAbsolutePath())
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)


    private writeInfluxMap = [:]

    class JenkinsUtilsMock extends JenkinsUtils {
        def isJobStartedByUser() {
            return true
        }
    }

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(thrown)
        .around(loggingRule)
        .around(shellRule)
        .around(writeFileRule)
        .around(readFileRule)
        .around(fileExistsRule)
        .around(dockerExecuteRule)
        .around(environmentRule)
        .around(credentialsRule)
        .around(stepRule) // needs to be activated after dockerExecuteRule, otherwise executeDocker is not mocked

    @Before
    void init() {
        // removing additional credentials tests might have added; adding default credentials
        credentialsRule.reset()
            .withCredentials('test_cfCredentialsId', 'test_cf', '********')

        UUID.metaClass.static.randomUUID = { -> 1 }
        helper.registerAllowedMethod('influxWriteData', [Map.class], { m ->
            writeInfluxMap = m
        })

        helper.registerAllowedMethod('findFiles', [Map.class], { m ->
            return [].toArray()
        })
    }

    @After
    void tearDown() {
        UUID.metaClass = null
    }

    @Test
    void testNoTool() throws Exception {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                camSystemRole: 'testRole',
                cfCredentialsId: 'test_cfCredentialsId'
            ],
            stages: [
                acceptance: [
                    cfOrg: 'testOrg',
                    cfSpace: 'testSpace',
                    deployUser: 'testUser',
                ]
            ],
            steps: [
                cloudFoundryDeploy: []
            ]
        ]
        nullScript.commonPipelineEnvironment.setBuildTool('mta')
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            mtaPath: 'target/test.mtar',
            stageName: 'acceptance',
        ])
        // asserts
        assertThat(loggingRule.log, containsString('[cloudFoundryDeploy] General parameters: deployTool=mtaDeployPlugin, deployType=standard, cfApiEndpoint=https://api.cf.eu10.hana.ondemand.com, cfOrg=testOrg, cfSpace=testSpace, cfCredentialsId=test_cfCredentialsId'))
    }

    @Test
    void testNotAvailableTool() throws Exception {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                cfCredentialsId: 'myCreds'
            ],
            stages: [
                acceptance: [
                    cfOrg: 'testOrg',
                    cfSpace: 'testSpace',
                    deployUser: 'testUser',
                ]
            ],
            steps: [
                cloudFoundryDeploy: []
            ]
        ]

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'notAvailable',
            stageName: 'acceptance'
        ])
        // asserts
        assertThat(loggingRule.log, containsString('[cloudFoundryDeploy] General parameters: deployTool=notAvailable, deployType=standard, cfApiEndpoint=https://api.cf.eu10.hana.ondemand.com, cfOrg=testOrg, cfSpace=testSpace, cfCredentialsId=myCreds'))
        assertThat(loggingRule.log, containsString('[cloudFoundryDeploy] WARNING! Found unsupported deployTool. Skipping deployment.'))
    }



    @Test
    void testCfNativeWithAppName() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])
        // asserts
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfNativeWithAppNameCustomApi() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfApiEndpoint: 'https://customApi',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])
        // asserts
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://customApi -o "testOrg" -s "testSpace"')))
    }

    @Test
    void testCfNativeWithAppNameCompatible() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
                credentialsId: 'test_cfCredentialsId',
                appName: 'testAppName',
                manifest: 'test.yml'
            ]
        ])
        // asserts
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfNativeWithDockerImage() {
        // adding additional credentials for Docker registry authorization
        credentialsRule.withCredentials('test_cfDockerCredentialsId', 'test_cf_docker', '********')
        readYamlRule.registerYaml('test.yml', "applications: [[name: 'manifestAppName']]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployDockerImage: 'repo/image:tag',
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
                credentialsId: 'test_cfCredentialsId',
                appName: 'testAppName'
            ]
        ])

        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString('cf push testAppName --docker-image repo/image:tag')))
        assertThat(shellRule.shell, hasItem(containsString('cf logout')))
    }

    @Test
    void testCfNativeWithDockerImageAndCredentials() {
        // adding additional credentials for Docker registry authorization
        credentialsRule.withCredentials('test_cfDockerCredentialsId', 'test_cf_docker', '********')
        readYamlRule.registerYaml('test.yml', "applications: [[name: 'manifestAppName']]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployDockerImage: 'repo/image:tag',
            dockerCredentialsId: 'test_cfDockerCredentialsId',
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
                credentialsId: 'test_cfCredentialsId',
                appName: 'testAppName'
            ]
        ])
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry(equalTo('CF_DOCKER_PASSWORD'), equalTo("${'********'}")))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString('cf push testAppName --docker-image repo/image:tag --docker-username test_cf_docker')))
        assertThat(shellRule.shell, hasItem(containsString('cf logout')))
    }

    @Test
    void testCfNativeWithManifestAndDockerCredentials() {
        // Docker image can be done via manifest.yml; if a private Docker registry is used, --docker-username and DOCKER_PASSWORD
        // must be set; this is checked by this test

        // adding additional credentials for Docker registry authorization
        credentialsRule.withCredentials('test_cfDockerCredentialsId', 'test_cf_docker', '********')
        readYamlRule.registerYaml('test.yml', "applications: [[name: 'manifestAppName']]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            dockerCredentialsId: 'test_cfDockerCredentialsId',
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
                credentialsId: 'test_cfCredentialsId',
                appName: 'testAppName',
                manifest: 'manifest.yml'
            ]
        ])
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry(equalTo('CF_DOCKER_PASSWORD'), equalTo("${'********'}")))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName -f 'manifest.yml' --docker-username test_cf_docker")))
        assertThat(shellRule.shell, hasItem(containsString('cf logout')))
    }

    @Test
    void testCfNativeBlueGreenWithManifestAndDockerCredentials() {
        // Blue Green Deploy cf cli plugin does not support --docker-username and --docker-image parameters
        // docker username and docker image have to be set in the manifest file
        // if a private docker repository is used the CF_DOCKER_PASSWORD env variable must be set

        credentialsRule.withCredentials('test_cfDockerCredentialsId', 'test_cf_docker', '********')
        readYamlRule.registerYaml('manifest.yml', "applications: [[name: 'manifestAppName']]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            dockerCredentialsId: 'test_cfDockerCredentialsId',
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
                credentialsId: 'test_cfCredentialsId',
                appName: 'testAppName',
                manifest: 'manifest.yml'
            ]
        ])
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry(equalTo('CF_DOCKER_PASSWORD'), equalTo("${'********'}")))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf blue-green-deploy testAppName --delete-old-apps -f 'manifest.yml'")))
        assertThat(shellRule.shell, hasItem(containsString('cf logout')))
    }

    @Test
    void testCfNativeAppNameFromManifest() {
        fileExistsRule.registerExistingFile('test.yml')
        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfManifest: 'test.yml'
        ])
        // asserts
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfNativeWithoutAppName() {
        fileExistsRule.registerExistingFile('test.yml')
        readYamlRule.registerYaml('test.yml', "applications: [{}]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryDeploy] ERROR: No appName available in manifest test.yml.')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfManifest: 'test.yml'
        ])
    }

    @Test
    void testCfNativeBlueGreenDefaultDeleteOldInstance() {

        readYamlRule.registerYaml('test.yml', "applications: [{}]")

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))

        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf blue-green-deploy testAppName --delete-old-apps -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))

    }

    @Test
    void testCfNativeBlueGreenExplicitDeleteOldInstance() {

        readYamlRule.registerYaml('test.yml', "applications: [{}]")

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            keepOldInstance: false,
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))

        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf blue-green-deploy testAppName --delete-old-apps -f 'test.yml'")))
        assertThat(shellRule.shell, not(hasItem(containsString("cf stop testAppName-old &>"))))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))

    }

    @Test
    void testCfNativeBlueGreenKeepOldInstance() {

        readYamlRule.registerYaml('test.yml', "applications: [{}]")

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            keepOldInstance: true,
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))

        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf blue-green-deploy testAppName -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf stop testAppName-old &>")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfNativeBlueGreenMultipleApplications() {

        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName1'},{name: 'manifestAppName2'}]")
        fileExistsRule.registerExistingFile('test.yml')

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[cloudFoundryDeploy] Your manifest contains more than one application. For blue green deployments your manifest file may contain only one application.")

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])
    }

    @Test
    void testCfNativeBlueGreenWithNoRoute() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName1', no-route: true}]")
        fileExistsRule.registerExistingFile('test.yml')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName -f 'test.yml'")))
    }

    @Test
    void testCfNativeBlueGreenKeepOldInstanceShouldThrowErrorOnStopError(){
        new File(tmpDir, '1-cfStopOutput.txt').write('any error message')

        helper.registerAllowedMethod("sh", [String], { cmd ->
            if (cmd.toString().contains('cf stop testAppName-old'))
                throw new Exception('fail')
        })

        readYamlRule.registerYaml('test.yml', "applications: [{}]")

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[cloudFoundryDeploy] ERROR: Could not stop application testAppName-old. Error: any error message")

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            keepOldInstance: true,
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))

        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf blue-green-deploy testAppName -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf stop testAppName-old &> 1-cfStopOutput.txt")))
    }

    @Test
    void testCfNativeStandardShouldNotStopInstance() {

        readYamlRule.registerYaml('test.yml', "applications: [{}]")

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'standard',
            keepOldInstance: true,
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(shellRule.shell, not(hasItem(containsString("cf stop testAppName-old &>"))))
    }

    @Test
    void testCfNativeWithoutAppNameBlueGreen() {

        fileExistsRule.registerExistingFile('test.yml')
        readYamlRule.registerYaml('test.yml', "applications: [{}]")

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryDeploy] ERROR: Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfManifest: 'test.yml'
        ])
    }

    @Test
    void testCfNativeFailureInShellCall() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        helper.registerAllowedMethod("sh", [String], { cmd ->
            if (cmd.toString().contains('cf login -u "test_cf"'))
                throw new Exception('fail')
        })

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryDeploy] ERROR: The execution of the deploy command failed, see the log for details.')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }


    @Test
    void testMta() {
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            mtaPath: 'target/test.mtar'
        ])
        // asserts
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString('cf deploy target/test.mtar -f')))
        assertThat(shellRule.shell, hasItem(containsString('cf logout')))
    }

    @Test
    void useMtaFilePathFromPipelineEnvironment() {
        environmentRule.env.mtarFilePath = 'target/test.mtar'

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin'
        ])
        // asserts
        assertThat(shellRule.shell, hasItem(containsString('cf deploy target/test.mtar -f')))
    }

    @Test
    void testMtaBlueGreen() {

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            deployType: 'blue-green',
            mtaPath: 'target/test.mtar'
        ])

        assertThat(shellRule.shell, hasItem(stringContainsInOrder(["cf login -u \"test_cf\"", 'cf bg-deploy', '-f', '--no-confirm'])))
    }

    @Test
    void testInfluxReporting() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3')
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])
        // asserts
        assertThat(writeInfluxMap.customDataMap.deployment_data.artifactUrl, is('n/a'))
        assertThat(writeInfluxMap.customDataMap.deployment_data.deployTime, containsString(new Date().format( 'MMM dd, yyyy')))
        assertThat(writeInfluxMap.customDataMap.deployment_data.jobTrigger, is('USER'))

        assertThat(writeInfluxMap.customDataMapTags.deployment_data.artifactVersion, is('1.2.3'))
        assertThat(writeInfluxMap.customDataMapTags.deployment_data.deployUser, is('test_cf'))
        assertThat(writeInfluxMap.customDataMapTags.deployment_data.deployResult, is('SUCCESS'))
        assertThat(writeInfluxMap.customDataMapTags.deployment_data.cfApiEndpoint, is('https://api.cf.eu10.hana.ondemand.com'))
        assertThat(writeInfluxMap.customDataMapTags.deployment_data.cfOrg, is('testOrg'))
        assertThat(writeInfluxMap.customDataMapTags.deployment_data.cfSpace, is('testSpace'))
    }

    @Test
    void testCfPushDeploymentWithVariableSubstitutionFromFile() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: '((appName))'}]")
        fileExistsRule.registerExistingFile('test.yml')
        fileExistsRule.registerExistingFile('vars.yml')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml',
            cfManifestVariablesFiles: ['vars.yml']
        ])

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName --vars-file 'vars.yml' -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
        assertThat(loggingRule.log,containsString("We will add the following string to the cf push call: --vars-file 'vars.yml' !"))
        assertThat(loggingRule.log,not(containsString("We will add the following string to the cf push call:  !")))
    }

    @Test
    void testCfPushDeploymentWithVariableSubstitutionFromNotExistingFilePrintsWarning() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: '((appName))'}]")
        fileExistsRule.registerExistingFile('test.yml')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml',
            cfManifestVariablesFiles: ['vars.yml']
        ])

        // asserts
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName -f 'test.yml'")))
        assertThat(loggingRule.log, containsString("[WARNING] We skip adding not-existing file 'vars.yml' as a vars-file to the cf create-service-push call"))
    }

    @Test
    void testCfPushDeploymentWithVariableSubstitutionFromVarsList() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: '((appName))'}]")
        List varsList = [["appName" : "testApplicationFromVarsList"]]

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml',
            cfManifestVariables: varsList
        ])

        // asserts
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName --var appName='testApplicationFromVarsList' -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
        assertThat(loggingRule.log,containsString("We will add the following string to the cf push call: --var appName='testApplicationFromVarsList' !"))
        assertThat(loggingRule.log,not(containsString("We will add the following string to the cf push call:  !")))
    }

    @Test
    void testCfPushDeploymentWithVariableSubstitutionFromVarsListNotAList() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: '((appName))'}]")

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryDeploy] ERROR: Parameter config.cloudFoundry.manifestVariables is not a List!')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml',
            cfManifestVariables: 'notAList'
        ])

    }

    @Test
    void testCfPushDeploymentWithVariableSubstitutionFromVarsListAndVarsFile() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: '((appName))'}]")
        List varsList = [["appName" : "testApplicationFromVarsList"]]
        fileExistsRule.registerExistingFile('vars.yml')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml',
            cfManifestVariablesFiles: ['vars.yml'],
            cfManifestVariables: varsList
        ])

        // asserts
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName --var appName='testApplicationFromVarsList' --vars-file 'vars.yml' -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfPushDeploymentWithoutVariableSubstitution() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: '((appName))'}]")

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        // asserts
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf push testAppName -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfBlueGreenDeploymentWithVariableSubstitution() {

        readYamlRule.registerYaml('test.yml', "applications: [{name: '((appName))'}]")
        readYamlRule.registerYaml('vars.yml', "[appName: 'testApplication']")

        fileExistsRule.registerExistingFile("test.yml")
        fileExistsRule.registerExistingFile("vars.yml")

        boolean testYamlWritten = false
        def testYamlData = null
        helper.registerAllowedMethod('writeYaml', [Map], { Map m ->
            if (m.file.equals("test.yml")) {
                testYamlWritten = true
                testYamlData = m.data
            }
        })

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml',
            cfManifestVariablesFiles: ['vars.yml']
        ])

        // asserts
        assertTrue(testYamlWritten)
        assertNotNull(testYamlData)
        assertThat(testYamlData.get("applications").get(0).get("name"), is("testApplication"))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf blue-green-deploy testAppName --delete-old-apps -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfBlueGreenDeploymentWithVariableSubstitutionFromVarsList() {
        readYamlRule.registerYaml('test.yml', "applications: [{name: '((appName))'}]")
        readYamlRule.registerYaml('vars.yml', "[appName: 'testApplication']")
        List varsList = [["appName" : "testApplicationFromVarsList"]]

        fileExistsRule.registerExistingFile("test.yml")
        fileExistsRule.registerExistingFile("vars.yml")

        boolean testYamlWritten = false
        def testYamlData = null
        helper.registerAllowedMethod('writeYaml', [Map], { Map m ->
            if (m.file.equals("test.yml")) {
                testYamlWritten = true
                testYamlData = m.data
            }
        })

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            deployType: 'blue-green',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml',
            cfManifestVariablesFiles: ['vars.yml'],
            cfManifestVariables: varsList
        ])

        // asserts
        assertTrue(testYamlWritten)
        assertNotNull(testYamlData)
        assertThat(testYamlData.get("applications").get(0).get("name"), is("testApplicationFromVarsList"))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli:6'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(dockerExecuteRule.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString("cf blue-green-deploy testAppName --delete-old-apps -f 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testTraceOutputOnVerbose() {

        fileExistsRule.existingFiles.addAll(
            'test.yml',
            'cf.log'
        )

        new File(tmpDir, 'cf.log') << 'Hello SAP'

        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
                manifest: 'test.yml',
                ],
            deployTool: 'cf_native',
            cfCredentialsId: 'test_cfCredentialsId',
            verbose: true
        ])

        assertThat(loggingRule.log, allOf(
            containsString('### START OF CF CLI TRACE OUTPUT ###'),
            containsString('Hello SAP'),
            containsString('### END OF CF CLI TRACE OUTPUT ###')))
    }

    @Test
    void testTraceNoTraceFileWritten() {

        fileExistsRule.existingFiles.addAll(
            'test.yml',
        )

        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
                manifest: 'test.yml',
                ],
            deployTool: 'cf_native',
            cfCredentialsId: 'test_cfCredentialsId',
            verbose: true
        ])

        assertThat(loggingRule.log, containsString('No trace file found'))
    }

    @Test
    void testAdditionCfNativeOpts() {

        readYamlRule.registerYaml('test.yml', "applications: [{name: 'manifestAppName'}]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3')
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            loginParameters: '--some-login-opt value',
            cfNativeDeployParameters: '--some-deploy-opt cf-value',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(shellRule.shell, hasItem(
            stringContainsInOrder([
                'cf login ', '--some-login-opt value',
                'cf push', '--some-deploy-opt cf-value'])))

    }

    @Test
    void testAdditionMtaOpts() {

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
            ],
            apiParameters: '--some-api-opt value',
            loginParameters: '--some-login-opt value',
            mtaDeployParameters: '--some-deploy-opt mta-value',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            deployType: 'blue-green',
            mtaPath: 'target/test.mtar'
        ])

        assertThat(shellRule.shell, hasItem(
            stringContainsInOrder([
                'cf api', '--some-api-opt value',
                'cf login ', '--some-login-opt value',
                'cf bg-deploy', '--some-deploy-opt mta-value'])))

    }

    @Test
    void 'appName with underscores should throw an error'() {
        String expected = "Your application name my_invalid_app_name contains a '_' (underscore) which is not allowed, only letters, dashes and numbers can be used. Please change the name to fit this requirement.\n" +
            "For more details please visit https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html#basic-settings."
        String actual = ""
        helper.registerAllowedMethod('error', [String.class], {s -> actual = s})

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'irrelevant',
                space: 'irrelevant',
                appName: 'my_invalid_app_name'
            ],
            cfCredentialsId: 'test_cfCredentialsId',
            mtaPath: 'irrelevant'
        ])

        assertEquals(expected, actual)
    }

    @Test
    void 'appName with alpha-numeric chars and leading dash should throw an error'() {
        String expected = "Your application name -my-Invalid-AppName123 contains a starts or ends with a '-' (dash) which is not allowed, only letters, dashes and numbers can be used. Please change the name to fit this requirement.\nFor more details please visit https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html#basic-settings."
        String actual = ""
        helper.registerAllowedMethod('error', [String.class], {s -> actual = s})

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'irrelevant',
                space: 'irrelevant',
                appName: '-my-Invalid-AppName123'
            ],
            cfCredentialsId: 'test_cfCredentialsId',
            mtaPath: 'irrelevant'
        ])

        assertEquals(expected, actual)
    }

    @Test
    void 'appName with alpha-numeric chars and trailing dash should throw an error'() {
        String expected = "Your application name my-Invalid-AppName123- contains a starts or ends with a '-' (dash) which is not allowed, only letters, dashes and numbers can be used. Please change the name to fit this requirement.\nFor more details please visit https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html#basic-settings."
        String actual = ""
        helper.registerAllowedMethod('error', [String.class], {s -> actual = s})

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'irrelevant',
                space: 'irrelevant',
                appName: 'my-Invalid-AppName123-'
            ],
            cfCredentialsId: 'test_cfCredentialsId',
            mtaPath: 'irrelevant'
        ])

        assertEquals(expected, actual)
    }

    @Test
    void 'appName with alpha-numeric chars should work'() {
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'irrelevant',
                space: 'irrelevant',
                appName: 'myValidAppName123'
            ],
            deployTool: 'cf_native',
            cfCredentialsId: 'test_cfCredentialsId',
            mtaPath: 'irrelevant'
        ])

        assertTrue(loggingRule.log.contains("cfAppName=myValidAppName123"))
    }

    @Test
    void 'appName with alpha-numeric chars and dash should work'() {
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'irrelevant',
                space: 'irrelevant',
                appName: 'my-Valid-AppName123'
            ],
            deployTool: 'cf_native',
            cfCredentialsId: 'test_cfCredentialsId',
            mtaPath: 'irrelevant'
        ])

        assertTrue(loggingRule.log.contains("cfAppName=my-Valid-AppName123"))
    }

    @Test
    void testMtaExtensionDescriptor() {
        fileExistsRule.existingFiles.addAll(
            'globalMtaDescriptor.mtaext',
        )
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace'
            ],
            mtaExtensionDescriptor: 'globalMtaDescriptor.mtaext',
            mtaDeployParameters: '--some-deploy-opt mta-value',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            deployType: 'blue-green',
            mtaPath: 'target/test.mtar'
        ])

        assertThat(shellRule.shell, hasItem(containsString("-e globalMtaDescriptor.mtaext")))
    }

    @Test
    void testTargetMtaExtensionDescriptor() {
        fileExistsRule.existingFiles.addAll(
            'targetMtaDescriptor.mtaext',
        )
        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cloudFoundry: [
                org: 'testOrg',
                space: 'testSpace',
                mtaExtensionDescriptor: 'targetMtaDescriptor.mtaext'
            ],
            mtaDeployParameters: '--some-deploy-opt mta-value',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            deployType: 'blue-green',
            mtaPath: 'target/test.mtar'
        ])
        assertThat(shellRule.shell, hasItem(containsString("-e targetMtaDescriptor.mtaext")))
    }

    @Test
    void testMtaExtensionCredentials() {
        fileExistsRule.existingFiles.addAll(
            'mtaext.mtaext',
        )
        credentialsRule.withCredentials("mtaExtCredTest","token")

        helper.registerAllowedMethod('readFile', [String], {file ->
            if (file == 'mtaext.mtaext') {
                return '_schema-version: \'3.1\'\n' +
                    'ID: test.ext\n' +
                    'extends: test\n' +
                    '\n' +
                    'parameters:\n' +
                    '  test-credentials: "<%= testCred %>"'
            }
            return ''
        })

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            mtaPath: 'target/test.mtar',
            mtaExtensionDescriptor: "mtaext.mtaext",
            mtaExtensionCredentials: [
                testCred: 'mtaExtCredTest'
            ]
        ])

        assertThat(shellRule.shell, hasItem(containsString('cp mtaext.mtaext mtaext.mtaext.original')))
        assertThat(shellRule.shell, hasItem(containsString('mv --force mtaext.mtaext.original mtaext.mtaext')))
        assertThat(writeFileRule.files['mtaext.mtaext'], is('_schema-version: \'3.1\'\n' +
            'ID: test.ext\n' +
            'extends: test\n' +
            '\n' +
            'parameters:\n' +
            '  test-credentials: "token"'))
    }

    @Test
    void testMtaExtensionDescriptorNotFound() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryDeploy] The mta extension descriptor file mtaext.mtaext does not exist at the configured location.')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            mtaPath: 'target/test.mtar',
            mtaExtensionDescriptor: "mtaext.mtaext"
        ])
    }

    @Test
    void testMtaExtensionDescriptorReadFails() {
        fileExistsRule.existingFiles.addAll(
            'mtaext.mtaext',
        )

        thrown.expect(Exception)
        thrown.expectMessage('[cloudFoundryDeploy] Unable to read mta extension file mtaext.mtaext.')

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            mtaPath: 'target/test.mtar',
            mtaExtensionDescriptor: "mtaext.mtaext",
            mtaExtensionCredentials: [
                testCred: 'mtaExtCredTest'
            ],
        ])
    }

    @Test
    void testGoStepFeatureToggleOn() {
        String calledStep = ''
        String usedMetadataFile = ''
        List credInfo = []
        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            Map parameters, String stepName,
            String metadataFile, List credentialInfo ->
                calledStep = stepName
                usedMetadataFile = metadataFile
                credInfo = credentialInfo
        })

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            useGoStep: true,
            deployTool: 'irrelevant',
            cfOrg: 'irrelevant',
            cfSpace: 'irrelevant',
            cfCredentialsId: 'irrelevant',
            mtaExtensionCredentials: [myCred: 'Mta.ExtensionCredential~Credential_Id1'],
        ])

        assertEquals('cloudFoundryDeploy', calledStep)
        assertEquals('metadata/cloudFoundryDeploy.yaml', usedMetadataFile)

        // contains assertion does not work apparently when comparing a list of lists agains an expected list.
        boolean found = false
            credInfo.each { entry ->
                if (entry == [type:'token', id:'Mta.ExtensionCredential~Credential_Id1', env:['MTA_EXTENSION_CREDENTIAL_CREDENTIAL_ID1'], resolveCredentialsId:false]) {
                    found = true
            }
	    }
        assertTrue(found)
    }

    @Test
    void testGoStepFeatureToggleOff() {
        String calledStep = ''
        String usedMetadataFile = ''
        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            Map parameters, String stepName,
            String metadataFile, List credentialInfo ->
                calledStep = stepName
                usedMetadataFile = metadataFile
        })

        stepRule.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            useGoStep: 'false',
            deployTool: 'irrelevant',
            cfOrg: 'irrelevant',
            cfSpace: 'irrelevant',
            cfCredentialsId: 'irrelevant',
        ])

        assertEquals('', calledStep)
        assertEquals('', usedMetadataFile)
    }
}
