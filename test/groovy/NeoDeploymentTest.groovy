import hudson.AbortException
import org.junit.rules.TemporaryFolder
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.JenkinsConfigRule
import util.JenkinsLoggingRule
import util.JenkinsSetupRule
import util.JenkinsShellCallRule

class NeoDeploymentTest extends PiperTestBase {

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
    def archivePath
    def warArchivePath
    def propertiesFilePath

    @Before
    void init() {

        archivePath = "${tmp.newFolder("workspace").toURI().getPath()}archiveName.mtar"
        warArchivePath = "${tmp.getRoot().toURI().getPath()}workspace/warArchive.war"
        propertiesFilePath = "${tmp.getRoot().toURI().getPath()}workspace/config.properties"

        helper.registerAllowedMethod('error', [String], { s -> throw new AbortException(s) })
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

    }


    @Test
    void straightForwardTest() {

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(archivePath) << "dummy archive"

        withPipeline(defaultPipeline()).execute(archivePath, 'myCredentialsId')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'anonymous' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")

    }


    @Test
    void badCredentialsIdTest() {

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(archivePath) << "dummy archive"

        thrown.expect(MissingPropertyException)

        withPipeline(defaultPipeline()).execute(archivePath, 'badCredentialsId')

    }


    @Test
    void credentialsIdNotProvidedTest() {

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(archivePath) << "dummy archive"

        withPipeline(noCredentialsIdPipeline()).execute(archivePath)

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }


    @Test
    void neoHomeNotSetTest() {

        new File(archivePath) << "dummy archive"

        withPipeline(noCredentialsIdPipeline()).execute(archivePath)

        assert jscr.shell[0] =~ /#!\/bin\/bash "neo" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert jlr.log.contains("Using Neo executable from PATH.")
    }


    @Test
    void neoHomeAsParameterTest() {

        new File(archivePath) << "dummy archive"

        withPipeline(neoHomeParameterPipeline()).execute(archivePath, 'myCredentialsId')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/etc\/neo\/tools\/neo\.sh" deploy-mta --user 'anonymous' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous.*/

        assert jlr.log.contains("[neoDeploy] Neo executable \"/etc/neo/tools/neo.sh\" retrieved from parameters.")

    }


    @Test
    void archiveNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR archivePath')

        withPipeline(noArchivePathPipeline()).execute()

    }


    @Test
    void wrongArchivePathProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Archive cannot be found with parameter archivePath: '")

        withPipeline(defaultPipeline()).execute(archivePath, 'myCredentialsId')

    }


    @Test
    void scriptNotProvidedTest() {

        new File(archivePath) << "dummy archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR deployHost')

        withPipeline(noScriptPipeline()).execute(archivePath)

    }

    @Test
    void mtaDeployModeTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(archivePath) << "dummy archive"

        withPipeline(mtaDeployModePipeline()).execute(archivePath, 'mta')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous.*/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void warFileParamsDeployModeTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(warArchivePath) << "dummy war archive"

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'warParams', 'lite', 'deploy')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" --host 'test\.deploy\.host\.com' --account 'trialuser123' --application 'testApp' --runtime 'neo-javaee6-wp' --runtime-version '2\.125' --size 'lite'/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void warFileParamsDeployModeRollingUpdateTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(warArchivePath) << "dummy war archive"

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'warParams', 'lite', 'rolling-update')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" rolling-update --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" --host 'test\.deploy\.host\.com' --account 'trialuser123' --application 'testApp' --runtime 'neo-javaee6-wp' --runtime-version '2\.125' --size 'lite'/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void warPropertiesFileDeployModeTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(warArchivePath) << "dummy war archive"
        new File(propertiesFilePath) << "dummy properties file"

        withPipeline(warPropertiesFileDeployModePipeline()).execute(warArchivePath, propertiesFilePath, 'warPropertiesFile', 'deploy')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" .*\.properties/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void warPropertiesFileDeployModeRollingUpdateTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(warArchivePath) << "dummy war archive"
        new File(propertiesFilePath) << "dummy properties file"

        withPipeline(warPropertiesFileDeployModePipeline()).execute(warArchivePath, propertiesFilePath, 'warPropertiesFile', 'rolling-update')

        assert jscr.shell[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" rolling-update --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" .*\.properties/
        assert jlr.log.contains("[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment.")
    }

    @Test
    void applicationNameNotProvidedTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR applicationName')

        withPipeline(noApplicationNamePipeline()).execute(warArchivePath, 'warParams')
    }

    @Test
    void runtimeNotProvidedTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtime')

        withPipeline(noRuntimePipeline()).execute(warArchivePath, 'warParams')
    }

    @Test
    void runtimeVersionNotProvidedTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtimeVersion')

        withPipeline(noRuntimeVersionPipeline()).execute(warArchivePath, 'warParams')
    }

    @Test
    void illegalDeployModeTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid deployMode = 'illegalMode'. Valid 'deployMode' values are: 'mta', 'warParams' and 'warPropertiesFile'")

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'illegalMode', 'lite', 'deploy')
    }

    @Test
    void illegalVMSizeTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid vmSize = 'illegalVM'. Valid 'vmSize' values are: 'lite', 'pro', 'prem' and 'prem-plus'.")

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'warParams', 'illegalVM', 'deploy')
    }

    @Test
    void illegalWARActionTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid warAction = 'illegalWARAction'. Valid 'warAction' values are: 'deploy' and 'rolling-update'.")

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'warParams', 'lite', 'illegalWARAction')
    }

    private defaultPipeline(){
        return """
               @Library('piper-library-os')

               execute(archivePath, neoCredentialsId) {

                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

                 node() {
                   neoDeploy script: this, archivePath: archivePath, neoCredentialsId: neoCredentialsId
                 }

               }

               return this
               """
    }

    private noCredentialsIdPipeline(){
        return """
               @Library('piper-library-os')

               execute(archivePath) {

                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

                 node() {
                   neoDeploy script: this, archivePath: archivePath
                 }

               }

               return this
               """
    }

    private neoHomeParameterPipeline(){
        return """
               @Library('piper-library-os')

               execute(archivePath, neoCredentialsId) {

                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

                 node() {
                   neoDeploy script: this, archivePath: archivePath, neoCredentialsId: neoCredentialsId, neoHome: '/etc/neo'
                 }

               }

               return this
               """
    }

    private noArchivePathPipeline(){
        return """
               @Library('piper-library-os')

               execute() {

                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

                 node() {
                   neoDeploy script: this
                 }

               }

               return this
               """
    }

    private noScriptPipeline(){
        return """
               @Library('piper-library-os')

               execute(archivePath) {

                 node() {
                   neoDeploy archivePath: archivePath
                 }

               }

               return this
               """
    }

    private noApplicationNamePipeline() {
        return """
               @Library('piper-library-os')

               execute(warArchivePath, deployMode) {
               
                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
                 def runtime = 'neo-javaee6-wp'
                 def runtimeVersion = '2.125'

                 node() {
                   neoDeploy script: this, archivePath: warArchivePath, deployMode: deployMode, runtime: runtime, runtimeVersion: runtimeVersion
                 }

               }

               return this
               """
    }

    private noRuntimePipeline() {
        return """
               @Library('piper-library-os')

               execute(warArchivePath, deployMode) {
               
                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
                 def appName = 'testApp'
                 def runtime = 'neo-javaee6-wp'
                 def runtimeVersion = '2.125'

                 node() {
                   neoDeploy script: this, archivePath: warArchivePath, deployMode: deployMode, applicationName: appName, runtimeVersion: runtimeVersion
                 }

               }

               return this
               """
    }

    private noRuntimeVersionPipeline() {
        return """
               @Library('piper-library-os')

               execute(warArchivePath, deployMode) {
               
                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
                 def appName = 'testApp'
                 def runtime = 'neo-javaee6-wp'
                 def runtimeVersion = '2.125'

                 node() {
                   neoDeploy script: this, archivePath: warArchivePath, deployMode: deployMode, applicationName: appName, runtime: runtime
                 }

               }

               return this
               """
    }

    private warPropertiesFileDeployModePipeline() {
        return """
               @Library('piper-library-os')

               execute(warArchivePath, propertiesFilePath, deployMode, warAction) {
               
                 node() {
                   neoDeploy script: this, deployMode: deployMode, archivePath: warArchivePath, propertiesFile: propertiesFilePath, warAction: warAction
                 }

               }

               return this
               """
    }

    private warParamsDeployModePipeline() {
        return """
               @Library('piper-library-os')

               execute(warArchivePath, deployMode, vmSize, warAction) {
               
                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
                 def appName = 'testApp'
                 def runtime = 'neo-javaee6-wp'
                 def runtimeVersion = '2.125'

                 node() {
                   neoDeploy script: this, archivePath: warArchivePath, deployMode: deployMode, applicationName: appName, runtime: runtime, runtimeVersion: runtimeVersion, warAction: warAction, vmSize: vmSize
                 }

               }

               return this
               """
    }

    private mtaDeployModePipeline() {
        return """
               @Library('piper-library-os')

               execute(archivePath, deployMode) {
               
                 commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
                 commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
               
                 node() {
                   neoDeploy script: this, archivePath: archivePath, deployMode: deployMode
                 }

               }

               return this
               """
    }

}
