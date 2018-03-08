import hudson.AbortException
import org.yaml.snakeyaml.Yaml
import org.yaml.snakeyaml.parser.ParserException

import org.junit.BeforeClass
import org.junit.ClassRule
import org.junit.Ignore
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.junit.rules.TemporaryFolder

import com.lesfurets.jenkins.unit.BasePipelineTest

import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsEnvironmentRule
import util.Rules

public class MtaBuildTest extends BasePipelineTest {

    def toolMtaValidateCalled = false
    def toolJavaValidateCalled = false

    @ClassRule
    public static TemporaryFolder tmp = new TemporaryFolder()

    private ExpectedException thrown = new ExpectedException()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(jsr)
        .around(jer)

    private static currentDir
    private static newDir
    private static mtaYaml

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

        //
        // needs to be after loading the scripts. Here we have a different behaviour
        // for usual steps and for steps contained in the shared lib itself.
        //
        // toolValidate mocked here since we are not interested in testing
        // toolValidate here. This is expected to be done in a test class for
        // toolValidate.
        //
        helper.registerAllowedMethod('toolValidate', [Map], { m ->

                                                              if(m.tool == 'mta')
                                                                  toolMtaValidateCalled = true

                                                              if(m.tool == 'java')
                                                                  toolJavaValidateCalled = true
                                                            })
    }


    @Test
    void environmentPathTest() {

        jsr.step.call(buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains('PATH=./node_modules/.bin:/usr/bin')}
    }


    @Test
    void sedTest() {

        jsr.step.call(buildTarget: 'NEO')

        assert jscr.shell.find { c -> c =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/mta.yaml"$/}
    }


    @Test
    void mtarFilePathFromCommonPipelineEnviromentTest() {

        jsr.step.call(script: [commonPipelineEnvironment: jer.env],
                      buildTarget: 'NEO')

        def mtarFilePath = jer.env.getMtarFilePath()

        assert mtarFilePath == "$currentDir/com.mycompany.northwind.mtar"
    }


    @Test
    void mtaBuildWithSurroundingDirTest() {

        helper.registerAllowedMethod('pwd', [], { newDir } )

        def mtarFilePath = jsr.step.call(buildTarget: 'NEO')

        assert jscr.shell.find { c -> c =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" ".*\/newDir\/mta.yaml"$/}

        assert mtarFilePath == "$newDir/com.mycompany.northwind.mtar"
    }


    @Test
    void mtaJarLocationNotSetTest() {

        jsr.step.call(buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains(' -jar mta.jar --mtar ')}

        assert jlr.log.contains('[mtaBuild] Using MTA JAR from current working directory.')
    }


    @Test
    void mtaJarLocationAsParameterTest() {

        jsr.step.call(mtaJarLocation: '/mylocation/mta', buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains(' -jar /mylocation/mta/mta.jar --mtar ')}

        assert jlr.log.contains('[mtaBuild] MTA JAR "/mylocation/mta/mta.jar" retrieved from configuration.')
    }


    @Test
    void noMtaPresentTest() {

        mtaYaml.delete()
        thrown.expect(FileNotFoundException)

        jsr.step.call(buildTarget: 'NEO')
    }


    @Test
    void badMtaTest() {

        thrown.expect(ParserException)
        thrown.expectMessage('while parsing a block mapping')

        mtaYaml.text = badMtaYaml()

        jsr.step.call(buildTarget: 'NEO')
    }


    @Test
    void noIdInMtaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Property 'ID' not found in mta.yaml file at: '")

        mtaYaml.text = noIdMtaYaml()

        jsr.step.call(buildTarget: 'NEO')
    }


    @Test
    void mtaJarLocationFromEnvironmentTest() {

        binding.setVariable('env', [:])
        binding.getVariable('env')['MTA_JAR_LOCATION'] = '/env/mta'

        jsr.step.call(buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains('-jar /env/mta/mta.jar --mtar')}
        assert jlr.log.contains('[mtaBuild] MTA JAR "/env/mta/mta.jar" retrieved from environment.')
    }


    @Test
    void mtaJarLocationFromCustomStepConfigurationTest() {

        jer.env.configuration = [steps:[mtaBuild:[mtaJarLocation: '/step/mta']]]

        jsr.step.call(script: [commonPipelineEnvironment: jer.env],
                      buildTarget: 'NEO')

        assert jscr.shell.find(){ c -> c.contains('-jar /step/mta/mta.jar --mtar')}
        assert jlr.log.contains('[mtaBuild] MTA JAR "/step/mta/mta.jar" retrieved from configuration.')
    }


    @Test
    void buildTargetFromParametersTest() {

        jsr.step.call(buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }


    @Test
    void buildTargetFromCustomStepConfigurationTest() {

        jer.env.configuration = [steps:[mtaBuild:[buildTarget: 'NEO']]]

        jsr.step.call(script: [commonPipelineEnvironment: jer.env])

        assert jscr.shell.find(){ c -> c.contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }


    @Test
    void buildTargetFromDefaultStepConfigurationTest() {

        jer.env.defaultConfiguration = [steps:[mtaBuild:[buildTarget: 'NEO']]]

        jsr.step.call(script: [commonPipelineEnvironment: jer.env])

        assert jscr.shell.find { c -> c.contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }

    @Ignore('Tool validation disabled since it does not work properly in conjunction with slaves.')
    @Test
    void skipValidationInCaseMtarJarFileIsUsedFromWorkingDir() {
        jscr.setReturnValue('ls mta.jar', 0)
        jsr.step.call(script: [commonPipelineEnvironment: jer.env])
        assert !toolMtaValidateCalled
    }

    @Ignore('Tool validation disabled since it does not work properly in conjunction with slaves.')
    @Test
    void performValidationInCaseMtarJarFileIsNotUsedFromWorkingDir() {
        jscr.setReturnValue('ls mta.jar', 1)
        jsr.step.call(script: [commonPipelineEnvironment: jer.env])
        assert toolMtaValidateCalled
    }

    @Ignore('Tool validation disabled since it does not work properly in conjunction with slaves.')
    @Test
    void toolJavaValidateCalled() {

        jsr.step.call(buildTarget: 'NEO')

        assert toolJavaValidateCalled
    }

    @Ignore('Tool validation disabled since it does not work properly in conjunction with slaves.')
    @Test
    void toolValidateNotCalledWhenJavaHomeIsUnsetButJavaIsInPath() {

        jscr.setReturnValue('which java', 0)
        jsr.step.call(buildTarget: 'NEO')

        assert !toolJavaValidateCalled
        assert jlr.log.contains('Tool validation (java) skipped. JAVA_HOME not set, but java executable in path.')
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
