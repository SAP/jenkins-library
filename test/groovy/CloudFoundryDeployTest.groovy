#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsEnvironmentRule
import util.JenkinsDockerExecuteRule
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.JenkinsReadYamlRule
import util.Rules

import static org.junit.Assert.assertThat

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.containsString

class CloudFoundryDeployTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    private JenkinsDockerExecuteRule jedr = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jryr)
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(jwfr)
        .around(jedr)
        .around(jer)
        .around(new JenkinsCredentialsRule(this).withCredentials('test_cfCredentialsId', 'test_cf', '********'))
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

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

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: '',
            stageName: 'acceptance',
        ])
        // asserts
        assertThat(jlr.log, containsString('[cloudFoundryDeploy] General parameters: deployTool=, deployType=standard, cfApiEndpoint=https://api.cf.eu10.hana.ondemand.com, cfOrg=testOrg, cfSpace=testSpace, cfCredentialsId=myCreds, deployUser=testUser'))
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

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'notAvailable',
            stageName: 'acceptance'
        ])
        // asserts
        assertThat(jlr.log, containsString('[cloudFoundryDeploy] General parameters: deployTool=notAvailable, deployType=standard, cfApiEndpoint=https://api.cf.eu10.hana.ondemand.com, cfOrg=testOrg, cfSpace=testSpace, cfCredentialsId=myCreds, deployUser=testUser'))
    }

    @Test
    void testCfNativeWithAppName() {
        jryr.registerYaml('test.yml', "applications: [[name: 'manifestAppName']]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])
        // asserts
        assertThat(jedr.dockerParams, hasEntry('dockerImage', 's4sdk/docker-cf-cli'))
        assertThat(jedr.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(jedr.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(jscr.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(jscr.shell, hasItem(containsString("cf push testAppName -f 'test.yml'")))
        assertThat(jscr.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfNativeWithAppNameCustomApi() {
        jryr.registerYaml('test.yml', "applications: [[name: 'manifestAppName']]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            cfApiEndpoint: 'https://customApi',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])
        // asserts
        assertThat(jscr.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://customApi -o "testOrg" -s "testSpace"')))
    }

    @Test
    void testCfNativeWithAppNameCompatible() {
        jryr.registerYaml('test.yml', "applications: [[name: 'manifestAppName']]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
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
        assertThat(jedr.dockerParams, hasEntry('dockerImage', 's4sdk/docker-cf-cli'))
        assertThat(jedr.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(jedr.dockerParams.dockerEnvVars, hasEntry('STATUS_CODE', "${200}"))
        assertThat(jscr.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(jscr.shell, hasItem(containsString("cf push testAppName -f 'test.yml'")))
        assertThat(jscr.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfNativeAppNameFromManifest() {
        helper.registerAllowedMethod('fileExists', [String.class], { s -> return true })
        jryr.registerYaml('test.yml', "applications: [[name: 'manifestAppName']]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfManifest: 'test.yml'
        ])
        // asserts
        assertThat(jscr.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(jscr.shell, hasItem(containsString("cf push -f 'test.yml'")))
        assertThat(jscr.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfNativeWithoutAppName() {
        helper.registerAllowedMethod('fileExists', [String.class], { s -> return true })
        jryr.registerYaml('test.yml', "applications: [[]]")
        helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryDeploy] ERROR: No appName available in manifest test.yml.')

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfManifest: 'test.yml'
        ])
    }

    @Test
    void testCfNativeBlueGreenDefaultDeleteOldInstance() {

        jryr.registerYaml('test.yml', "applications: [[]]")

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            deployType: 'blue-green',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(jedr.dockerParams, hasEntry('dockerImage', 's4sdk/docker-cf-cli'))
        assertThat(jedr.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))

        assertThat(jscr.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(jscr.shell, hasItem(containsString("cf blue-green-deploy testAppName --delete-old-apps -f 'test.yml'")))
        assertThat(jscr.shell, hasItem(containsString("cf logout")))

    }

    @Test
    void testCfNativeBlueGreenExplicitDeleteOldInstance() {

        jryr.registerYaml('test.yml', "applications: [[]]")

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            deployType: 'blue-green',
            keepOldInstance: false,
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(jedr.dockerParams, hasEntry('dockerImage', 's4sdk/docker-cf-cli'))
        assertThat(jedr.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))

        assertThat(jscr.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(jscr.shell, hasItem(containsString("cf blue-green-deploy testAppName --delete-old-apps -f 'test.yml'")))
        assertThat(jscr.shell, hasItem(containsString("cf logout")))

    }

    @Test
    void testCfNativeBlueGreenKeepOldInstance() {

        jryr.registerYaml('test.yml', "applications: [[]]")

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            deployType: 'blue-green',
            keepOldInstance: true,
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(jedr.dockerParams, hasEntry('dockerImage', 's4sdk/docker-cf-cli'))
        assertThat(jedr.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))

        assertThat(jscr.shell, hasItem(containsString('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(jscr.shell, hasItem(containsString("cf blue-green-deploy testAppName -f 'test.yml'")))
        assertThat(jscr.shell, hasItem(containsString("cf stop testAppName-old")))
        assertThat(jscr.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testCfNativeStandardShouldNotStopInstance() {

        jryr.registerYaml('test.yml', "applications: [[]]")

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            deployType: 'standard',
            keepOldInstance: true,
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfAppName: 'testAppName',
            cfManifest: 'test.yml'
        ])

        assertThat(jscr.shell, not(hasItem(containsString("cf stop testAppName-old"))))
    }

    @Test
    void testCfNativeWithoutAppNameBlueGreen() {

        helper.registerAllowedMethod('fileExists', [String.class], { s -> return true })
        jryr.registerYaml('test.yml', "applications: [[]]")

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryDeploy] ERROR: Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)')

        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            deployTool: 'cf_native',
            deployType: 'blue-green',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfManifest: 'test.yml'
        ])
    }


    @Test
    void testMta() {
        jsr.step.cloudFoundryDeploy([
            script: nullScript,
            juStabUtils: utils,
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            deployTool: 'mtaDeployPlugin',
            mtaPath: 'target/test.mtar'
        ])
        // asserts
        assertThat(jedr.dockerParams, hasEntry('dockerImage', 's4sdk/docker-cf-cli'))
        assertThat(jedr.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(jscr.shell, hasItem(containsString('cf login -u test_cf -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"')))
        assertThat(jscr.shell, hasItem(containsString('cf deploy target/test.mtar -f')))
        assertThat(jscr.shell, hasItem(containsString('cf logout')))
    }
}
