import hudson.AbortException
import org.junit.rules.TemporaryFolder

import com.lesfurets.jenkins.unit.BasePipelineTest

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.JenkinsConfigRule
import util.JenkinsLoggingRule
import util.JenkinsSetupRule
import util.JenkinsShellCallRule

class NeoDeploymentTest extends BasePipelineTest {

    private ExpectedException thrown = new ExpectedException().none()
    private TemporaryFolder tmp = new TemporaryFolder()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = RuleChain.outerRule(thrown)
                                              .around(tmp)
                                              .around(new JenkinsSetupRule(this))
                                              .around(jlr)
                                              .around(jscr)
                                              .around(new JenkinsConfigRule(this))

    def workspacePath
    def warArchiveName
    def propertiesFileName
    def archiveName


    def neoDeployScript
    def cpe

    @Before
    void init() {

        workspacePath = "${tmp.newFolder("workspace").toURI().getPath()}"
        warArchiveName = 'warArchive.war'
        propertiesFileName = 'config.properties'
        archiveName = "archive.mtar"

        helper.registerAllowedMethod('error', [String], { s -> throw new AbortException(s) })
        helper.registerAllowedMethod('fileExists', [String], { s -> return new File(workspacePath, s).exists() })
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if(l[0].credentialsId == 'myCredentialsId') {
                binding.setProperty('username', 'anonymous')
                binding.setProperty('password', '********')
            } else if(l[0].credentialsId == 'CI_CREDENTIALS_ID') {
                binding.setProperty('username', 'defaultUser')
                binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('username', null)
                binding.setProperty('password', null)
            }

        })

        binding.setVariable('env', [:])

        neoDeployScript = loadScript("neoDeploy.groovy").neoDeploy
        cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment


    }


    @Test
    void straightForwardTest() {

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(workspacePath, archiveName) << "dummy archive"

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                       archivePath: archiveName,
                       neoCredentialsId: 'myCredentialsId'
        )

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'anonymous' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")

    }


    @Test
    void badCredentialsIdTest() {

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(workspacePath, archiveName) << "dummy archive"

        thrown.expect(MissingPropertyException)

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                       archivePath: archiveName,
                       neoCredentialsId: 'badCredentialsId'
        )
    }


    @Test
    void credentialsIdNotProvidedTest() {

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(workspacePath, archiveName) << "dummy archive"

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                       archivePath: archiveName
        )

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }


    @Test
    void neoHomeNotSetTest() {

        new File(workspacePath, archiveName) << "dummy archive"

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                       archivePath: archiveName
        )

        assert jscr.shell[0] =~ /#!\/bin\/bash "neo" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert jlr.log.contains("Using Neo executable from PATH.")
    }


    @Test
    void neoHomeAsParameterTest() {

        new File(workspacePath, archiveName) << "dummy archive"

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                       archivePath: archiveName,
                       neoCredentialsId: 'myCredentialsId',
                       neoHome: '/etc/neo'
        )

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/etc\/neo\/tools\/neo\.sh" deploy-mta --user 'anonymous' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous.*/

        assert jlr.log.contains("[neoDeploy] Neo executable \"/etc/neo/tools/neo.sh\" retrieved from parameters.")

    }


    @Test
    void archiveNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR archivePath')

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe])
    }


    @Test
    void wrongArchivePathProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Archive cannot be found with parameter archivePath: '")

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                       archivePath: archiveName)
    }


    @Test
    void scriptNotProvidedTest() {

        new File(workspacePath, archiveName) << "dummy archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR deployHost')

        neoDeployScript.call(archivePath: archiveName)
    }

    @Test
    void mtaDeployModeTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(workspacePath, archiveName) << "dummy archive"

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe], archivePath: archiveName, deployMode: 'mta')


        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous.*/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void warFileParamsDeployModeTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(workspacePath, warArchiveName) << "dummy war archive"

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                             applicationName: 'testApp',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125',
                             deployMode: 'warParams',
                             vmSize: 'lite',
                             warAction: 'deploy',
                             archivePath: warArchiveName)

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" --host 'test\.deploy\.host\.com' --account 'trialuser123' --application 'testApp' --runtime 'neo-javaee6-wp' --runtime-version '2\.125' --size 'lite'/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void warFileParamsDeployModeRollingUpdateTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(workspacePath, warArchiveName) << "dummy war archive"

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                             archivePath: warArchiveName,
                             deployMode: 'warParams',
                             applicationName: 'testApp',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125',
                             warAction: 'rolling-update',
                             vmSize: 'lite')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" rolling-update --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" --host 'test\.deploy\.host\.com' --account 'trialuser123' --application 'testApp' --runtime 'neo-javaee6-wp' --runtime-version '2\.125' --size 'lite'/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void warPropertiesFileDeployModeTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(workspacePath, warArchiveName) << "dummy war archive"
        new File(workspacePath, propertiesFileName) << "dummy properties file"

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                             archivePath: warArchiveName,
                             deployMode: 'warPropertiesFile',
                             propertiesFile: propertiesFileName,
                             applicationName: 'testApp',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125',
                             warAction: 'deploy',
                             vmSize: 'lite')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" .*\.properties/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void warPropertiesFileDeployModeRollingUpdateTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(workspacePath, warArchiveName) << "dummy war archive"
        new File(workspacePath, propertiesFileName) << "dummy properties file"

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                             archivePath: warArchiveName,
                             deployMode: 'warPropertiesFile',
                             propertiesFile: propertiesFileName,
                             applicationName: 'testApp',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125',
                             warAction: 'rolling-update',
                             vmSize: 'lite')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" rolling-update --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" .*\.properties/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void applicationNameNotProvidedTest() {
        new File(workspacePath, warArchiveName) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR applicationName')

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                             archivePath: warArchiveName,
                             deployMode: 'warParams',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125'
            )
    }

    @Test
    void runtimeNotProvidedTest() {
        new File(workspacePath, warArchiveName) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtime')

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                             archivePath: warArchiveName,
                             applicationName: 'testApp',
                             deployMode: 'warParams',
                             runtimeVersion: '2.125')
    }

    @Test
    void runtimeVersionNotProvidedTest() {
        new File(workspacePath, warArchiveName) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtimeVersion')

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
                             archivePath: warArchiveName,
                             applicationName: 'testApp',
                             deployMode: 'warParams',
                             runtime: 'neo-javaee6-wp')
    }

    @Test
    void illegalDeployModeTest() {
        new File(workspacePath, warArchiveName) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid deployMode = 'illegalMode'. Valid 'deployMode' values are: 'mta', 'warParams' and 'warPropertiesFile'")

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
            archivePath: warArchiveName,
            deployMode: 'illegalMode',
            applicationName: 'testApp',
            runtime: 'neo-javaee6-wp',
            runtimeVersion: '2.125',
            warAction: 'deploy',
            vmSize: 'lite')
    }

    @Test
    void illegalVMSizeTest() {
        new File(workspacePath, warArchiveName) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid vmSize = 'illegalVM'. Valid 'vmSize' values are: 'lite', 'pro', 'prem' and 'prem-plus'.")

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
            archivePath: warArchiveName,
            deployMode: 'warParams',
            applicationName: 'testApp',
            runtime: 'neo-javaee6-wp',
            runtimeVersion: '2.125',
            warAction: 'deploy',
            vmSize: 'illegalVM')
    }

    @Test
    void illegalWARActionTest() {
        new File(workspacePath, warArchiveName) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid warAction = 'illegalWARAction'. Valid 'warAction' values are: 'deploy' and 'rolling-update'.")

        cpe.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        cpe.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

        neoDeployScript.call(script: [commonPipelineEnvironment: cpe],
            archivePath: warArchiveName,
            deployMode: 'warParams',
            applicationName: 'testApp',
            runtime: 'neo-javaee6-wp',
            runtimeVersion: '2.125',
            warAction: 'illegalWARAction',
            vmSize: 'lite')
    }
}
