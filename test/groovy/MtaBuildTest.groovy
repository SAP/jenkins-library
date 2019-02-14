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
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this).registerYaml('mta.yaml', defaultMtaYaml() )

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(thrown)
        .around(loggingRule)
        .around(shellRule)
        .around(dockerExecuteRule)
        .around(stepRule)

    @Before
    void init() {

        helper.registerAllowedMethod('fileExists', [String], { s -> s == 'mta.yaml' })

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*\\$MTA_JAR_LOCATION.*', '')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*\\$JAVA_HOME.*', '')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*which java.*', 0)
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*java -version.*', '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)''')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*mta\\.jar -v.*', '1.0.6')

        binding.setVariable('PATH', '/usr/bin')
    }


    @Test
    void environmentPathTest() {

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert shellRule.shell.find { c -> c.contains('PATH=./node_modules/.bin:/usr/bin')}
    }


    @Test
    void sedTest() {

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert shellRule.shell.find { c -> c =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" "mta.yaml"$/}
    }


    @Test
    void marFilePathFromCommonPipelineEnvironmentTest() {

        stepRule.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        def mtarFilePath = nullScript.commonPipelineEnvironment.getMtarFilePath()

        assert mtarFilePath == "com.mycompany.northwind.mtar"
    }

    @Test
    void mtaJarLocationAsParameterTest() {

        stepRule.step.mtaBuild(script: nullScript, mtaJarLocation: '/mylocation/mta/mta.jar', buildTarget: 'NEO')

        assert shellRule.shell.find { c -> c.contains('-jar /mylocation/mta/mta.jar --mtar')}

        assert loggingRule.log.contains("SAP Multitarget Application Archive Builder file '/mylocation/mta/mta.jar' retrieved from configuration.")
        assert loggingRule.log.contains("Using SAP Multitarget Application Archive Builder '/mylocation/mta/mta.jar'.")
    }


    @Test
    void noMtaPresentTest() {
        helper.registerAllowedMethod('fileExists', [String], { false })
        thrown.expect(AbortException)
        thrown.expectMessage('\'mta.yaml\' not found in project sources and \'applicationName\' not provided as parameter ' +
                                '- cannot generate \'mta.yaml\' file.')

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')
    }


    @Test
    void badMtaTest() {

        thrown.expect(ParserException)
        thrown.expectMessage('while parsing a block mapping')

        readYamlRule.registerYaml('mta.yaml', badMtaYaml())

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')
    }


    @Test
    void noIdInMtaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Property 'ID' not found in mta.yaml file.")

        readYamlRule.registerYaml('mta.yaml', noIdMtaYaml() )

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')
    }


    @Test
    void mtaJarLocationFromEnvironmentTest() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*\\$MTA_JAR_LOCATION.*', '/env/mta/mta.jar')

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert shellRule.shell.find { c -> c.contains("-jar /env/mta/mta.jar --mtar")}
        assert loggingRule.log.contains("SAP Multitarget Application Archive Builder file '/env/mta/mta.jar' retrieved from environment.")
        assert loggingRule.log.contains("Using SAP Multitarget Application Archive Builder '/env/mta/mta.jar'.")
    }


    @Test
    void mtaJarLocationFromCustomStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration = [steps:[mtaBuild:[mtaJarLocation: '/config/mta/mta.jar']]]

        stepRule.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        assert shellRule.shell.find(){ c -> c.contains("-jar /config/mta/mta.jar --mtar")}
        assert loggingRule.log.contains("SAP Multitarget Application Archive Builder file '/config/mta/mta.jar' retrieved from configuration.")
        assert loggingRule.log.contains("Using SAP Multitarget Application Archive Builder '/config/mta/mta.jar'.")
    }


    @Test
    void mtaJarLocationFromDefaultStepConfigurationTest() {

        stepRule.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        assert shellRule.shell.find(){ c -> c.contains("-jar /opt/sap/mta/lib/mta.jar --mtar")}
        assert loggingRule.log.contains("SAP Multitarget Application Archive Builder file '/opt/sap/mta/lib/mta.jar' retrieved from configuration.")
        assert loggingRule.log.contains("Using SAP Multitarget Application Archive Builder '/opt/sap/mta/lib/mta.jar'.")
    }


    @Test
    void buildTargetFromParametersTest() {

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert shellRule.shell.find { c -> c.contains('java -jar /opt/sap/mta/lib/mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }


    @Test
    void buildTargetFromCustomStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration = [steps:[mtaBuild:[buildTarget: 'NEO']]]

        stepRule.step.mtaBuild(script: nullScript)

        assert shellRule.shell.find(){ c -> c.contains('java -jar /opt/sap/mta/lib/mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }

    @Test
    void canConfigureDockerImage() {

        stepRule.step.mtaBuild(script: nullScript, dockerImage: 'mta-docker-image:latest')

        assert 'mta-docker-image:latest' == dockerExecuteRule.dockerParams.dockerImage
    }

    @Test
    void canConfigureDockerOptions() {

        stepRule.step.mtaBuild(script: nullScript, dockerOptions: 'something')

        assert 'something' == dockerExecuteRule.dockerParams.dockerOptions
    }

    @Test
    void buildTargetFromDefaultStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.defaultConfiguration = [steps:[mtaBuild:[buildTarget: 'NEO']]]

        stepRule.step.mtaBuild(script: nullScript)

        assert shellRule.shell.find { c -> c.contains('java -jar /opt/sap/mta/lib/mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO build')}
    }


    @Test
    void extensionFromParametersTest() {

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO', extension: 'param_extension')

        assert shellRule.shell.find { c -> c.contains('java -jar /opt/sap/mta/lib/mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO --extension=param_extension build')}
    }


    @Test
    void extensionFromCustomStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration = [steps:[mtaBuild:[buildTarget: 'NEO', extension: 'config_extension']]]

        stepRule.step.mtaBuild(script: nullScript)

        assert shellRule.shell.find(){ c -> c.contains('java -jar /opt/sap/mta/lib/mta.jar --mtar com.mycompany.northwind.mtar --build-target=NEO --extension=config_extension build')}
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
