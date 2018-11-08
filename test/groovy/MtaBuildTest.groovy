import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.parser.ParserException

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

public class MtaBuildTest extends BasePiperTest {

    def toolMtaValidateCalled = false
    def toolJavaValidateCalled = false

    private ExpectedException thrown = new ExpectedException()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule jder = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this).registerYaml('mta.yaml', defaultMtaYaml() )

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jryr)
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(jder)
        .around(jsr)

    @Before
    void init() {

        helper.registerAllowedMethod('fileExists', [String], { s -> false })
        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionWithoutEnvVars(m) })

        binding.setVariable('PATH', '/usr/bin')
    }


    @Test
    void environmentPathTest() {

        jsr.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains('PATH=./node_modules/.bin:/usr/bin')}
    }


    @Test
    void sedTest() {

        jsr.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert jscr.shell.find { c -> c =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" "mta.yaml"$/}
    }


    @Test
    void mtarFilePathFromCommonPipelineEnviromentTest() {

        jsr.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        def mtarFilePath = nullScript.commonPipelineEnvironment.getMtarFilePath()

        assert mtarFilePath == "com.mycompany.northwind.mtar"
    }

    @Test
    void mtaJarLocationAsParameterTest() {

        jsr.step.mtaBuild(script: nullScript, mtaJarLocation: '/mylocation/mta/mta.jar', buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains('-jar /mylocation/mta/mta.jar --mtar')}

        assert jlr.log.contains("SAP Multitarget Application Archive Builder file '/mylocation/mta/mta.jar' retrieved from configuration.")
        assert jlr.log.contains("Using SAP Multitarget Application Archive Builder '/mylocation/mta/mta.jar'.")
    }


    @Test
    void noMtaPresentTest() {

        jryr.registerYaml('mta.yaml', { throw new FileNotFoundException() })
        thrown.expect(FileNotFoundException)

        jsr.step.mtaBuild(script: nullScript, buildTarget: 'NEO')
    }


    @Test
    void badMtaTest() {

        thrown.expect(ParserException)
        thrown.expectMessage('while parsing a block mapping')

        jryr.registerYaml('mta.yaml', badMtaYaml())

        jsr.step.mtaBuild(script: nullScript, buildTarget: 'NEO')
    }


    @Test
    void noIdInMtaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Property 'ID' not found in mta.yaml file.")

        jryr.registerYaml('mta.yaml', noIdMtaYaml() )

        jsr.step.mtaBuild(script: nullScript, buildTarget: 'NEO')
    }


    @Test
    void mtaJarLocationFromEnvironmentTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionWithEnvVars(m) })

        jsr.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains("-jar /env/mta/mta.jar --mtar")}
        assert jlr.log.contains("SAP Multitarget Application Archive Builder file '/env/mta/mta.jar' retrieved from environment.")
        assert jlr.log.contains("Using SAP Multitarget Application Archive Builder '/env/mta/mta.jar'.")
    }


    @Test
    void mtaJarLocationFromCustomStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration = [steps:[mtaBuild:[mtaJarLocation: '/config/mta/mta.jar']]]

        jsr.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        assert jscr.shell.find(){ c -> c.contains("-jar /config/mta/mta.jar --mtar")}
        assert jlr.log.contains("SAP Multitarget Application Archive Builder file '/config/mta/mta.jar' retrieved from configuration.")
        assert jlr.log.contains("Using SAP Multitarget Application Archive Builder '/config/mta/mta.jar'.")
    }


    @Test
    void mtaJarLocationFromDefaultStepConfigurationTest() {

        jsr.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        assert jscr.shell.find(){ c -> c.contains("-jar mta.jar --mtar")}
        assert jlr.log.contains("SAP Multitarget Application Archive Builder file 'mta.jar' retrieved from configuration.")
        assert jlr.log.contains("Using SAP Multitarget Application Archive Builder 'mta.jar'.")
    }


    @Test
    void buildTargetFromParametersTest() {

        jsr.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert jscr.shell.find { c -> c.contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }


    @Test
    void buildTargetFromCustomStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration = [steps:[mtaBuild:[buildTarget: 'NEO']]]

        jsr.step.mtaBuild(script: nullScript)

        assert jscr.shell.find(){ c -> c.contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }

    @Test
    void canConfigureDockerImage() {

        jsr.step.mtaBuild(script: nullScript, dockerImage: 'mta-docker-image:latest')

        assert 'mta-docker-image:latest' == jder.dockerParams.dockerImage
    }

    @Test
    void canConfigureDockerOptions() {

        jsr.step.mtaBuild(script: nullScript, dockerOptions: 'something')

        assert 'something' == jder.dockerParams.dockerOptions
    }

    @Test
    void buildTargetFromDefaultStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.defaultConfiguration = [steps:[mtaBuild:[buildTarget: 'NEO']]]

        jsr.step.mtaBuild(script: nullScript)

        assert jscr.shell.find { c -> c.contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }


    @Test
    void extensionFromParametersTest() {

        jsr.step.mtaBuild(script: nullScript, buildTarget: 'NEO', extension: 'param_extension')

        assert jscr.shell.find { c -> c.contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO --extension=param_extension build')}
    }


    @Test
    void extensionFromCustomStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration = [steps:[mtaBuild:[buildTarget: 'NEO', extension: 'config_extension']]]

        jsr.step.mtaBuild(script: nullScript)

        assert jscr.shell.find(){ c -> c.contains('java -jar mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO --extension=config_extension build')}
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

    private getVersionWithEnvVars(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta.jar -v')) {
            return '1.0.6'
        } else {
            return getEnvVars(m)
        }
    }

    private getVersionWithoutEnvVars(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta.jar -v')) {
            return '1.0.6'
        } else {
            return getNoEnvVars(m)
        }
    }

    private getVersionWithoutEnvVarsAndNotInCurrentDir(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta.jar -v')) {
            return '1.0.6'
        } else {
            return getNoEnvVarsAndNotInCurrentDir(m)
        }
    }

    private getEnvVars(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return ''
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return '/env/mta/mta.jar'
        } else if(m.script.contains('which java')) {
            return 0
        } else {
            return 0
        }
    }

    private getNoEnvVars(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return ''
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return ''
        } else if(m.script.contains('which java')) {
            return 0
        } else {
            return 0
        }
    }

    private getNoEnvVarsAndNotInCurrentDir(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return ''
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return ''
        } else if(m.script.contains('which java')) {
            return 0
        } else {
            return 1
        }
    }
}
