#!groovy
import groovy.json.JsonSlurperClassic
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
import util.Rules

import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals

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

        assertTrue(jlr.log.contains('[cloudFoundryDeploy] General parameters: deployTool=, deployType=standard, cfApiEndpoint=https://api.cf.eu10.hana.ondemand.com, cfOrg=testOrg, cfSpace=testSpace, cfCredentialsId=myCreds, deployUser=testUser'))
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

        assertTrue(jlr.log.contains('[cloudFoundryDeploy] General parameters: deployTool=notAvailable, deployType=standard, cfApiEndpoint=https://api.cf.eu10.hana.ondemand.com, cfOrg=testOrg, cfSpace=testSpace, cfCredentialsId=myCreds, deployUser=testUser'))
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

        assertEquals('s4sdk/docker-cf-cli', jedr.dockerParams.dockerImage)
        assertEquals('/home/piper', jedr.dockerParams.dockerWorkspace)


        assertTrue(jscr.shell[1].contains('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"'))
        assertTrue(jscr.shell[1].contains('cf push "testAppName" -f "test.yml"'))
    }

    //issue https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/794
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

        assertTrue(jscr.shell[1].contains('cf login -u "test_cf" -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"'))
        assertTrue(jscr.shell[1].contains('cf push -f "test.yml"'))

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

        assertEquals('to be changed', jedr.dockerParams.dockerImage)
        assertEquals('/home/piper', jedr.dockerParams.dockerWorkspace)

        assertTrue(jscr.shell[0].contains('cf login -u test_cf -p \'********\' -a https://api.cf.eu10.hana.ondemand.com -o "testOrg" -s "testSpace"'))
        assertTrue(jscr.shell[0].contains("cf deploy target/test.mtar -f".toString()))
    }

}
