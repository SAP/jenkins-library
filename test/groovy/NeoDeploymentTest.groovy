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

    @Before
    void setup() {

        super._setUp()

        archivePath = "${tmp.newFolder("workspace").toURI().getPath()}archiveName.mtar"

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

        assert shellCalls[0] =~ /#!\/bin\/bash \/opt\/neo\/tools\/neo\.sh deploy-mta --user anonymous --host test\.deploy\.host\.com --source ".*" --account trialuser123 --password \*\*\*\*\*\*\*\* --synchronous/

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

        assert shellCalls[0] =~ /#!\/bin\/bash \/opt\/neo\/tools\/neo\.sh deploy-mta --user defaultUser --host test\.deploy\.host\.com --source ".*" --account trialuser123 --password \*\*\*\*\*\*\*\* --synchronous/

        assert messages[1] == "[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment."
    }


    @Test
    void neoHomeNotSetTest() {

        new File(archivePath) << "dummy archive"

        withPipeline(noCredentialsIdPipeline()).execute(archivePath)

        assert shellCalls[0] =~ /#!\/bin\/bash neo deploy-mta --user defaultUser --host test\.deploy\.host\.com --source ".*" --account trialuser123 --password \*\*\*\*\*\*\*\* --synchronous/

        assert messages[1] == "Using Neo executable from PATH."
    }


    @Test
    void neoHomeAsParameterTest() {

        new File(archivePath) << "dummy archive"

        withPipeline(neoHomeParameterPipeline()).execute(archivePath, 'myCredentialsId')

        assert shellCalls[0] =~ /#!\/bin\/bash \/etc\/neo\/tools\/neo\.sh deploy-mta --user anonymous --host test\.deploy\.host\.com --source ".*" --account trialuser123 --password \*\*\*\*\*\*\*\* --synchronous/

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


    private defaultPipeline(){
    { -> """
         @Library('piper-library-os')

         execute(archivePath, neoCredentialsId) {

           commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
           commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

           node() {
             neoDeploy script: this, archivePath: archivePath, neoCredentialsId: neoCredentialsId
           }

         }

         return this
         """}
    }

    private noCredentialsIdPipeline(){
        { -> """
             @Library('piper-library-os')

             execute(archivePath) {

               commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
               commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

               node() {
                 neoDeploy script: this, archivePath: archivePath
               }

             }

            return this
            """ }
    }

    private neoHomeParameterPipeline(){
        { -> """
             @Library('piper-library-os')

             execute(archivePath, neoCredentialsId) {

               commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
               commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

               node() {
                 neoDeploy script: this, archivePath: archivePath, neoCredentialsId: neoCredentialsId, neoHome: '/etc/neo'
               }

             }

             return this
             """ }
    }

    private noArchivePathPipeline(){
        { -> """
             @Library('piper-library-os')

             execute() {

               commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
               commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')

               node() {
                 neoDeploy script: this
               }

            }

            return this
            """ }
    }

    private noScriptPipeline(){
    { -> """
         @Library('piper-library-os')

         execute(archivePath) {

           node() {
             neoDeploy archivePath: archivePath
           }

         }

         return this
         """ }
    }

}
