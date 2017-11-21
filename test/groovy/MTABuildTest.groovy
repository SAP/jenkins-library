import hudson.AbortException
import org.jenkinsci.plugins.pipeline.utility.steps.shaded.org.yaml.snakeyaml.Yaml
import org.jenkinsci.plugins.pipeline.utility.steps.shaded.org.yaml.snakeyaml.parser.ParserException
import org.junit.Before
import org.junit.Rule
import org.junit.Test

import org.junit.rules.ExpectedException
import org.junit.rules.TemporaryFolder

public class MTABuildTest extends PiperTestBase {

    @Rule
    public ExpectedException thrown = new ExpectedException()

    @Rule
    public TemporaryFolder tmp = new TemporaryFolder()

    def currentDir
    def otherDir
    def mtaBuildShEnv


    @Before
    void setUp() {

        super.setUp()
        currentDir = tmp.newFolder().toURI().getPath()[0..-2] //omit final '/'
        otherDir = tmp.newFolder().toURI().getPath()[0..-2] //omit final '/'

        helper.registerAllowedMethod('readYaml', [Map], {
            m ->
                return new Yaml().load((m.file as File).text)
        })
        helper.registerAllowedMethod("dir", [String, Closure], {
            s, c ->
                currentDir = "${currentDir}/${s}"
                c()
        })
        helper.registerAllowedMethod('pwd', [], { currentDir } )
        helper.registerAllowedMethod("withEnv", [List.class, Closure.class],
                { l, c ->
                    mtaBuildShEnv = l
                    c()
                })
        helper.registerAllowedMethod('error', [String], { s -> throw new hudson.AbortException(s) })

        binding.setVariable('PATH', '/usr/bin')
        binding.setVariable('JAVA_HOME', '/opt/java')
        binding.setVariable('env', [:])

    }


    @Test
    public void straightForwardTest(){

        binding.getVariable('env')['MTA_JAR_LOCATION'] = '/opt/mta'

        new File("${currentDir}/mta.yaml") << defaultMtaYaml()

        def mtarFilePath = withPipeline(defaultPipeline()).execute()

        assert shellCalls[0] =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/mta.yaml"$/

        assert shellCalls[1].contains("PATH=./node_modules/.bin:/usr/bin")

        assert shellCalls[1].contains(' -jar /opt/mta/mta.jar --mtar ')

        assert mtarFilePath == "${currentDir}/com.mycompany.northwind.mtar"

        assert messages[1] == "[mtaBuild] MTA JAR \"/opt/mta/mta.jar\" retrieved from environment."
    }


    @Test
    public void mtarFilePathFromCommonPipelineEnviromentTest(){

        binding.getVariable('env')['MTA_JAR_LOCATION'] = '/opt/mta'

        new File("${currentDir}/mta.yaml") << defaultMtaYaml()

        def mtarFilePath = withPipeline(returnMtarFilePathFromCommonPipelineEnvironmentPipeline()).execute()

        assert shellCalls[0] =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/mta.yaml"$/

        assert shellCalls[1].contains("PATH=./node_modules/.bin:/usr/bin")

        assert shellCalls[1].contains(' -jar /opt/mta/mta.jar --mtar ')

        assert mtarFilePath == "${currentDir}/com.mycompany.northwind.mtar"

        assert messages[1] == "[mtaBuild] MTA JAR \"/opt/mta/mta.jar\" retrieved from environment."
    }


    @Test
    public void mtaBuildWithSurroundingDirTest(){

        binding.getVariable('env')['MTA_JAR_LOCATION'] = '/opt/mta'

        def newDirName = 'newDir'
        new File("${currentDir}/${newDirName}").mkdirs()
        new File("${currentDir}/${newDirName}/mta.yaml") << defaultMtaYaml()

        def mtarFilePath = withPipeline(withSurroundingDirPipeline()).execute(newDirName)

        assert shellCalls[0] =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/newDir\/mta.yaml"$/

        assert shellCalls[1].contains("PATH=./node_modules/.bin:/usr/bin")

        assert shellCalls[1].contains(' -jar /opt/mta/mta.jar --mtar ')

        assert mtarFilePath == "${currentDir}/com.mycompany.northwind.mtar"

        assert messages[1] == "[mtaBuild] MTA JAR \"/opt/mta/mta.jar\" retrieved from environment."
    }

    @Test
    void mtaHomeNotSetTest() {

        new File("${currentDir}/mta.yaml") << defaultMtaYaml()

        def mtarFilePath = withPipeline(defaultPipeline()).execute()

        assert shellCalls[0] =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/mta.yaml"$/

        assert shellCalls[1].contains("PATH=./node_modules/.bin:/usr/bin")

        assert shellCalls[1].contains(' -jar mta.jar --mtar ')

        assert mtarFilePath == "${currentDir}/com.mycompany.northwind.mtar"

        assert messages[1] == "[mtaBuild] Using MTA JAR from current working directory."
    }


    @Test
    void mtaHomeAsParameterTest() {

        new File("${currentDir}/mta.yaml") << defaultMtaYaml()

        def mtarFilePath = withPipeline(mtaJarLocationAsParameterPipeline()).execute()

        assert shellCalls[0] =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/mta.yaml"$/

        assert shellCalls[1].contains("PATH=./node_modules/.bin:/usr/bin")

        assert shellCalls[1].contains(' -jar /etc/mta/mta.jar --mtar ')

        assert mtarFilePath == "${currentDir}/com.mycompany.northwind.mtar"

        assert messages[1] == "[mtaBuild] MTA JAR \"/etc/mta/mta.jar\" retrieved from parameters."
    }


    @Test
    public void noMtaPresentTest(){
        thrown.expect(FileNotFoundException)

        withPipeline(defaultPipeline()).execute()
    }


    @Test
    public void badMtaTest(){
        thrown.expect(ParserException)
        thrown.expectMessage('while parsing a block mapping')

        new File("${currentDir}/mta.yaml") << badMtaYaml()

        withPipeline(defaultPipeline()).execute()
    }


    @Test
    public void noIdInMtaTest(){
        thrown.expect(AbortException)
        thrown.expectMessage("Property 'ID' not found in mta.yaml file at: '")

        new File("${currentDir}/mta.yaml") << noIdMtaYaml()

        withPipeline(defaultPipeline()).execute()
    }


    @Test
    public void noBuildTargetTest(){
        thrown.expect(Exception)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR buildTarget")

        new File("${currentDir}/mta.yaml") << defaultMtaYaml()

        withPipeline(noBuildTargetPipeline()).execute()
    }


    private defaultPipeline(){
        return '''
               @Library('piper-library-os')

               execute(){
                 mtaBuild buildTarget: 'NEO'
               }

               return this
               '''
    }

    private returnMtarFilePathFromCommonPipelineEnvironmentPipeline(){
        return '''
               @Library('piper-library-os')

               execute(){
                 mtaBuild buildTarget: 'NEO'
                 return commonPipelineEnvironment.getMtarFilePath()
               }

               return this
               '''
    }

    private mtaJarLocationAsParameterPipeline(){
        return '''
               @Library('piper-library-os')

               execute(){
                 mtaBuild mtaJarLocation: '/etc/mta', buildTarget: 'NEO'
               }

               return this
               '''
    }

    private withSurroundingDirPipeline(){
        return '''
               @Library('piper-library-os')

               execute(dirPath){
                 dir("${dirPath}"){
                   mtaBuild buildTarget: 'NEO'
                 }
               }

               return this
               '''
    }

    private noBuildTargetPipeline(){
        return '''
               @Library('piper-library-os')

               execute(){
                 mtaBuild()
               }

               return this
               '''
    }


    private defaultMtaYaml(){
        return  '''
                _schema-version: "2.0.0"
                ID: "com.mycompany.northwind"
                version: 1.0.0
                
                parameters:
                  hcp-deployer-version: "1.0.0"
                
                modules:
                  - name: "fiorinorthwind"
                    type: html5
                    path: .
                    parameters:
                       version: 1.0.0-${timestamp}
                    build-parameters:
                      builder: grunt
                build-result: dist
                '''
    }

    private badMtaYaml(){
        return  '''
                _schema-version: "2.0.0
                ID: "com.mycompany.northwind"
                version: 1.0.0
                
                parameters:
                  hcp-deployer-version: "1.0.0"
                
                modules:
                  - name: "fiorinorthwind"
                    type: html5
                    path: .
                    parameters:
                       version: 1.0.0-${timestamp}
                    build-parameters:
                      builder: grunt
                build-result: dist
                '''
    }

    private noIdMtaYaml(){
        return  '''
                _schema-version: "2.0.0"
                version: 1.0.0
                
                parameters:
                  hcp-deployer-version: "1.0.0"
                
                modules:
                  - name: "fiorinorthwind"
                    type: html5
                    path: .
                    parameters:
                       version: 1.0.0-${timestamp}
                    build-parameters:
                      builder: grunt
                build-result: dist
                '''
    }
}
