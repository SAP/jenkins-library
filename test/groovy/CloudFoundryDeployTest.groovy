import com.sap.piper.JenkinsUtils
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
            deployTool: '',
            stageName: 'acceptance',
        ])
        // asserts
        assertThat(loggingRule.log, containsString('[cloudFoundryDeploy] General parameters: deployTool=, deployType=standard, cfApiEndpoint=https://api.cf.eu10.hana.ondemand.com, cfOrg=testOrg, cfSpace=testSpace, cfCredentialsId=myCreds'))
        assertThat(loggingRule.log, containsString('[cloudFoundryDeploy] WARNING! Found unsupported deployTool. Skipping deployment.'))
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
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(shellRule.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(shellRule.shell, hasItem(containsString('cf deploy target/test.mtar -f')))
        assertThat(shellRule.shell, hasItem(containsString('cf logout')))
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

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
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

}
