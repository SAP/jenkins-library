import hudson.AbortException
import org.junit.rules.TemporaryFolder

import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

import static ProjectSource.projectSource

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException

import com.lesfurets.jenkins.unit.BasePipelineTest


class NeoDeploymentTest extends BasePipelineTest {

    @Rule
    public ExpectedException thrown = new ExpectedException().none()

    @Rule
    public TemporaryFolder tmp = new TemporaryFolder()

    def script

    def shellCalls = []

    def pipeline
    def echoes = []
    def archivePath

    @Before
    void setup() {

        super.setUp()

        archivePath = "${tmp.newFolder("workspace").toURI().getPath()}archiveName.mtar"
        pipeline = "${tmp.newFolder("pipeline").toURI().getPath()}pipeline"

        def piperLib = library()
                .name('piper-library-os')
                .retriever(projectSource())
                .targetPath('clonePath/is/not/necessary')
                .defaultVersion('irrelevant')
                .allowOverride(true)
                .implicit(false)
                .build()
        helper.registerSharedLibrary(piperLib)

        helper.registerAllowedMethod('sh', [String], { GString s ->
            shellCalls.add(s.replaceAll(/\s+/, " ").trim())
        })
        helper.registerAllowedMethod('error', [String], { s -> throw new AbortException(s) })
        helper.registerAllowedMethod('echo', [String], { s -> echoes.add(s) })
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

        defaultPipeline()

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(archivePath) << "dummy archive"

        script = loadScript(pipeline)

        script.execute(archivePath, 'myCredentialsId')

        assert shellCalls[0] =~ /#!\/bin\/bash \/opt\/neo\/tools\/neo\.sh deploy-mta --user anonymous --host test\.deploy\.host\.com --source ".*" --account trialuser123 --password \*\*\*\*\*\*\*\* --synchronous/

        assert echoes[1] == "[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment."

    }


    @Test
    void badCredentialsIdTest() {

        defaultPipeline()

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(archivePath) << "dummy archive"

        thrown.expect(MissingPropertyException)

        script = loadScript(pipeline)

        script.execute(archivePath, 'badCredentialsId')

    }


    @Test
    void credentialsIdNotProvidedTest() {

        noCredentialsIdPipeline()

        binding.getVariable('env')['NEO_HOME'] = '/opt/neo'

        new File(archivePath) << "dummy archive"

        script = loadScript(pipeline)

        script.execute(archivePath)

        assert shellCalls[0] =~ /#!\/bin\/bash \/opt\/neo\/tools\/neo\.sh deploy-mta --user defaultUser --host test\.deploy\.host\.com --source ".*" --account trialuser123 --password \*\*\*\*\*\*\*\* --synchronous/

        assert echoes[1] == "[neoDeploy] Neo executable \"/opt/neo/tools/neo.sh\" retrieved from environment."
    }


    @Test
    void neoHomeNotSetTest() {

        noCredentialsIdPipeline()

        new File(archivePath) << "dummy archive"

        script = loadScript(pipeline)

        script.execute(archivePath)

        assert shellCalls[0] =~ /#!\/bin\/bash neo deploy-mta --user defaultUser --host test\.deploy\.host\.com --source ".*" --account trialuser123 --password \*\*\*\*\*\*\*\* --synchronous/

        assert echoes[1] == "Using Neo executable from PATH."
    }


    @Test
    void neoHomeAsParameterTest() {

        neoHomeParameterPipeline()

        new File(archivePath) << "dummy archive"

        script = loadScript(pipeline)

        script.execute(archivePath, 'myCredentialsId')

        assert shellCalls[0] =~ /#!\/bin\/bash \/etc\/neo\/tools\/neo\.sh deploy-mta --user anonymous --host test\.deploy\.host\.com --source ".*" --account trialuser123 --password \*\*\*\*\*\*\*\* --synchronous/

        assert echoes[1] == "[neoDeploy] Neo executable \"/etc/neo/tools/neo.sh\" retrieved from parameters."

    }


    @Test
    void archiveNotProvidedTest() {

        noArchivePathPipeline()

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR archivePath')

        script = loadScript(pipeline)

        script.execute()

    }


    @Test
    void wrongArchivePathProvidedTest() {

        defaultPipeline()

        thrown.expect(AbortException)
        thrown.expectMessage("Archive cannot be found with parameter archivePath: '")

        script = loadScript(pipeline)

        script.execute(archivePath, 'myCredentialsId')

    }


    @Test
    void scriptNotProvidedTest() {

        noScriptPipeline()

        new File(archivePath) << "dummy archive"

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR deployHost')

        script = loadScript(pipeline)

        script.execute(archivePath)

    }


    private defaultPipeline(){
        new File(pipeline) <<   """
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
        new File(pipeline) <<   """
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
        new File(pipeline) <<   """
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
        new File(pipeline) <<   """
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
        new File(pipeline) <<   """
                                @Library('piper-library-os')
                                
                                execute(archivePath) {
                                
                                  node() {
                                    neoDeploy archivePath: archivePath
                                  }
                                
                                }
                                
                                return this
                                """
    }

}
