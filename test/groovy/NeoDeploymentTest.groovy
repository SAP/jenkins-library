import hudson.AbortException
import org.junit.rules.TemporaryFolder

import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

import static ProjectSource.projectSource

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException

class NeoDeploymentTest extends PiperTestBase {

    @Rule
    public ExpectedException thrown = new ExpectedException().none()

    @Rule
    public TemporaryFolder tmp = new TemporaryFolder()

    def archivePath
    def warArchivePath
    def propertiesFilePath

    @Before
    void setUp() {

        super.setUp()

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

        assert shellCalls[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'anonymous' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert messages[1] == "[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment."

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

        assert shellCalls[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert messages[1] == "[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment."
    }


    @Test
    void neoHomeNotSetTest() {

        new File(archivePath) << "dummy archive"

        withPipeline(noCredentialsIdPipeline()).execute(archivePath)

        assert shellCalls[0] =~ /#!\/bin\/bash "neo" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous/

        assert messages[1] == "Using Neo executable from PATH."
    }


    @Test
    void neoHomeAsParameterTest() {

        new File(archivePath) << "dummy archive"

        withPipeline(neoHomeParameterPipeline()).execute(archivePath, 'myCredentialsId')

        assert shellCalls[0] =~ /#!\/bin\/bash "\/etc\/neo\/tools\/neo\.sh" deploy-mta --user 'anonymous' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous.*/

        assert messages[1] == "[neoDeploy] Neo executable \"/etc/neo/tools/neo.sh\" retrieved from parameters."

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

        withPipeline(mtaDeployModePipeline()).execute(archivePath, 'MTA')

        assert shellCalls[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy-mta --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*" --host 'test\.deploy\.host\.com' --account 'trialuser123' --synchronous.*/
        assert messages[1] == "[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment."
    }

    @Test
    void warFileParamsDeployModeTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(warArchivePath) << "dummy war archive"

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'WAR_PARAMS', 'lite', 'deploy')

        assert shellCalls[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" --host 'test\.deploy\.host\.com' --account 'trialuser123' --application 'testApp' --runtime 'neo-javaee6-wp' --runtime-version '2\.125' --size 'lite'/
        assert messages[1] == "[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment."
    }

    @Test
    void warPropertiesFileDeployModeTest() {
        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'
        new File(warArchivePath) << "dummy war archive"
        new File(propertiesFilePath) << "dummy properties file"

        withPipeline(warPropertiesFileDeployModePipeline()).execute(warArchivePath, propertiesFilePath, 'WAR_PROPERTIESFILE')

        assert shellCalls[0] =~ /#!\/bin\/bash "\/opt\/neo\/tools\/neo\.sh" deploy --user 'defaultUser' --password '\*\*\*\*\*\*\*\*' --source ".*\.war" .*\.properties/
        assert messages[1] == "[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment."
    }

    @Test
    void applicationNameNotProvidedTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR applicationName')

        withPipeline(noApplicationNamePipeline()).execute(warArchivePath, 'WAR_PARAMS')
    }

    @Test
    void runtimeNotProvidedTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtime')

        withPipeline(noRuntimePipeline()).execute(warArchivePath, 'WAR_PARAMS')
    }

    @Test
    void runtimeVersionNotProvidedTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtimeVersion')

        withPipeline(noRuntimeVersionPipeline()).execute(warArchivePath, 'WAR_PARAMS')
    }

    @Test
    void illegalDeployModeTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("[neoDeploy] Invalid deployMode = 'ILLEGAL_MODE'. Valid 'deployMode' values are: 'MTA', 'WAR_PARAMS' and 'WAR_PROPERTIESFILE'")

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'ILLEGAL_MODE', 'lite', 'deploy')
    }

    @Test
    void illegalVMSizeTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("[neoDeploy] Invalid vmSize = 'illegalVM'. Valid 'vmSize' values are: 'lite', 'pro', 'prem' and 'prem-plus'.")

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'WAR_PARAMS', 'illegalVM', 'deploy')
    }

    @Test
    void illegalWARActionTest() {
        new File(warArchivePath) << "dummy war archive"

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("[neoDeploy] Invalid warAction = 'illegalWARAction'. Valid 'warAction' values are: 'deploy' and 'rolling-update'.")

        withPipeline(warParamsDeployModePipeline()).execute(warArchivePath, 'WAR_PARAMS', 'lite', 'illegalWARAction')
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

               execute(warArchivePath, propertiesFilePath, deployMode) {
               
                 node() {
                   neoDeploy script: this, deployMode: deployMode, archivePath: warArchivePath, propertiesFile: propertiesFilePath
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
