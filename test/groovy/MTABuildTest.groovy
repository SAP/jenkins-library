import hudson.AbortException
import org.yaml.snakeyaml.Yaml
import org.yaml.snakeyaml.parser.ParserException

import org.junit.BeforeClass
import org.junit.ClassRule
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.junit.rules.TemporaryFolder

import com.lesfurets.jenkins.unit.BasePipelineTest

import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.Rules

public class MTABuildTest extends BasePipelineTest {

    @ClassRule
    public static TemporaryFolder tmp = new TemporaryFolder()

    private ExpectedException thrown = new ExpectedException()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
            .around(thrown)
            .around(jlr)
            .around(jscr)

    private static currentDir
    private static newDir
    private static mtaYaml

    def mtaBuildScript
    def cpe

    @BeforeClass
    static void createTestFiles() {

        currentDir = "${tmp.getRoot()}"
        mtaYaml = tmp.newFile('mta.yaml')
        newDir = "$currentDir/newDir"
        tmp.newFolder('newDir')
        tmp.newFile('newDir/mta.yaml') << defaultMtaYaml()
    }

    @Before
    void init() {

        mtaYaml.text = defaultMtaYaml()

        helper.registerAllowedMethod('pwd', [], { currentDir } )

        binding.setVariable('PATH', '/usr/bin')

        mtaBuildScript = loadScript('mtaBuild.groovy').mtaBuild
        cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
    }


    @Test
    void environmentPathTest() {

        mtaBuildScript.call(buildTarget: 'NEO')

        assert jscr.shell[1].contains('PATH=./node_modules/.bin:/usr/bin')
    }


    @Test
    void sedTest() {

        mtaBuildScript.call(buildTarget: 'NEO')

        assert jscr.shell[0] =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/mta.yaml"$/
    }


    @Test
    void mtarFilePathFromCommonPipelineEnviromentTest() {

        mtaBuildScript.call(script: [commonPipelineEnvironment: cpe],
                      buildTarget: 'NEO')

        def mtarFilePath = cpe.getMtarFilePath()

        assert mtarFilePath == "$currentDir/com.mycompany.northwind.mtar"
    }


    @Test
    void mtaBuildWithSurroundingDirTest() {

        helper.registerAllowedMethod('pwd', [], { newDir } )

        def mtarFilePath = mtaBuildScript.call(buildTarget: 'NEO')

        assert jscr.shell[0] =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/newDir\/mta.yaml"$/

        assert mtarFilePath == "$newDir/com.mycompany.northwind.mtar"
    }


    @Test
    void mtaJarLocationNotSetTest() {

        mtaBuildScript.call(buildTarget: 'NEO')

        assert jscr.shell[1].contains(' -jar mta.jar --mtar ')

        assert jlr.log.contains('[mtaBuild] Using MTA JAR from current working directory.')
    }


    @Test
    void mtaJarLocationAsParameterTest() {

        mtaBuildScript.call(mtaJarLocation: '/mylocation/mta', buildTarget: 'NEO')

        assert jscr.shell[1].contains(' -jar /mylocation/mta/mta.jar --mtar ')

        assert jlr.log.contains('[mtaBuild] MTA JAR "/mylocation/mta/mta.jar" retrieved from configuration.')
    }


    @Test
    void noMtaPresentTest() {

        mtaYaml.delete()
        thrown.expect(FileNotFoundException)

        mtaBuildScript.call(buildTarget: 'NEO')
    }


    @Test
    void badMtaTest() {

        thrown.expect(ParserException)
        thrown.expectMessage('while parsing a block mapping')

        mtaYaml.text = badMtaYaml()

        mtaBuildScript.call(buildTarget: 'NEO')
    }


    @Test
    void noIdInMtaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Property 'ID' not found in mta.yaml file at: '")

        mtaYaml.text = noIdMtaYaml()

        mtaBuildScript.call(buildTarget: 'NEO')
    }


    @Test
    void mtaJarLocationFromEnvironmentTest() {

        binding.setVariable('env', [:])
        binding.getVariable('env')['MTA_JAR_LOCATION'] = '/env/mta'

        mtaBuildScript.call(buildTarget: 'NEO')

        assert jscr.shell[1].contains('-jar /env/mta/mta.jar --mtar')
        assert jlr.log.contains('[mtaBuild] MTA JAR "/env/mta/mta.jar" retrieved from environment.')
    }


    @Test
    void mtaJarLocationFromCustomStepConfigurationTest() {

        cpe.configuration = [steps:[mtaBuild:[mtaJarLocation: '/step/mta']]]

        mtaBuildScript.call(script: [commonPipelineEnvironment: cpe],
                      buildTarget: 'NEO')

        assert jscr.shell[1].contains('-jar /step/mta/mta.jar --mtar')
        assert jlr.log.contains('[mtaBuild] MTA JAR "/step/mta/mta.jar" retrieved from configuration.')
    }


    @Test
    void buildTargetFromParametersTest() {

        mtaBuildScript.call(buildTarget: 'NEO')

        assert jscr.shell[1].contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')
    }


    @Test
    void buildTargetFromCustomStepConfigurationTest() {

        cpe.configuration = [steps:[mtaBuild:[buildTarget: 'NEO']]]

        mtaBuildScript.call(script: [commonPipelineEnvironment: cpe])

        assert jscr.shell[1].contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')
    }


    @Test
    void buildTargetFromDefaultStepConfigurationTest() {

        cpe.defaultConfiguration = [steps:[mtaBuild:[buildTarget: 'NEO']]]

        mtaBuildScript.call(script: [commonPipelineEnvironment: cpe])

        assert jscr.shell[1].contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')
    }


    private static defaultMtaYaml() {
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

    private badMtaYaml() {
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

    private noIdMtaYaml() {
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
