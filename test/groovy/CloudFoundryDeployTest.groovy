#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml
import util.BasePiperTest
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

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(jwfr)
        .around(jedr)
        .around(jer)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() throws Throwable {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if(l[0].credentialsId == 'test_cfCredentialsId') {
                binding.setProperty('username', 'test_cf')
                binding.setProperty('password', '********')
            } else if(l[0].credentialsId == 'test_camCredentialsId') {
                binding.setProperty('username', 'test_cam')
                binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('username', null)
                binding.setProperty('password', null)
            }
        })
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
        assertThat(jscr.shell, hasItem(containsString('cf push "testAppName" -f "test.yml"')))
    }

    @Test
    void testCfNativeWithAppNameCustomApi() {
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
        assertThat(jscr.shell, hasItem(containsString('cf push "testAppName" -f "test.yml"')))
    }

    @Test
    void testCfNativeAppNameFromManifest() {
        helper.registerAllowedMethod('fileExists', [String.class], { s -> return true })
        helper.registerAllowedMethod("readYaml", [Map], { Map m ->
            if(m.text) {
                return new Yaml().load(m.text)
            } else if(m.file == 'test.yml') {
                return [applications: [[name: 'manifestAppName']]]
            } else if(m.file) {
                return new Yaml().load((m.file as File).text)
            } else {
                throw new IllegalArgumentException("Key 'text' is missing in map ${m}.")
            }
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
        assertThat(jscr.shell, hasItem(containsString('cf push -f "test.yml"')))
    }

    @Test
    void testCfNativeWithoutAppName() {
        helper.registerAllowedMethod('fileExists', [String.class], { s -> return true })
        helper.registerAllowedMethod("readYaml", [Map], { Map m ->
            if(m.text) {
                return new Yaml().load(m.text)
            } else if(m.file == 'test.yml') {
                return [applications: [[]]]
            } else if(m.file) {
                return new Yaml().load((m.file as File).text)
            } else {
                throw new IllegalArgumentException("Key 'text' is missing in map ${m}.")
            }
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
    }
}
