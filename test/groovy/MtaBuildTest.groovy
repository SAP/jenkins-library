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
import util.JenkinsWriteFileRule
import util.Rules

public class MtaBuildTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this).registerYaml('mta.yaml', defaultMtaYaml() )
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(thrown)
        .around(loggingRule)
        .around(shellRule)
        .around(dockerExecuteRule)
        .around(stepRule)
        .around(writeFileRule)

    @Before
    void init() {

        helper.registerAllowedMethod('fileExists', [String], { s -> s == 'mta.yaml' })

        helper.registerAllowedMethod('httpRequest', [String.class], { s -> new SettingsStub()})

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*\\$MTA_JAR_LOCATION.*', '')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*\\$JAVA_HOME.*', '')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*which java.*', 0)
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*java -version.*', '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)''')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*mta\\.jar -v.*', '1.0.6')

    }


    @Test
    void environmentPathTest() {

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert shellRule.shell.find { c -> c.contains('PATH=./node_modules/.bin:$PATH')}
    }


    @Test
    void sedTest() {

        stepRule.step.mtaBuild(script: nullScript, buildTarget: 'NEO')

        assert shellRule.shell.find { c -> c =~ /sed -ie "s\/\\\$\{timestamp\}\/`date \+%Y%m%d%H%M%S`\/g" "mta.yaml"$/}
    }


    @Test
    void mtarFilePathFromCommonPipelineEnvironmentTest() {

        stepRule.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        def mtarFilePath = nullScript.commonPipelineEnvironment.getMtarFilePath()

        assert mtarFilePath == "com.mycompany.northwind.mtar"
    }

    @Test
    void mtaJarLocationAsParameterTest() {

        stepRule.step.mtaBuild(script: nullScript, mtaJarLocation: '/mylocation/mta/mta.jar', buildTarget: 'NEO')

        assert shellRule.shell.find { c -> c.contains('-jar /mylocation/mta/mta.jar --mtar')}
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
    void mtaJarLocationFromCustomStepConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration = [steps:[mtaBuild:[mtaJarLocation: '/config/mta/mta.jar']]]

        stepRule.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        assert shellRule.shell.find(){ c -> c.contains('java -jar /config/mta/mta.jar --mtar')}
    }


    @Test
    void mtaJarLocationFromDefaultStepConfigurationTest() {

        stepRule.step.mtaBuild(script: nullScript,
                      buildTarget: 'NEO')

        assert shellRule.shell.find(){ c -> c.contains('java -jar /opt/sap/mta/lib/mta.jar --mtar')}
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
    void canConfigureMavenUserSettings() {

        stepRule.step.mtaBuild(script: nullScript, projectSettingsFile: 'settings.xml')

        assert shellRule.shell.find(){ c -> c.contains('cp settings.xml $HOME/.m2/settings.xml')}
    }

    @Test
    void canConfigureMavenUserSettingsFromRemoteSource() {

        stepRule.step.mtaBuild(script: nullScript, projectSettingsFile: 'https://some.host/my-settings.xml')

        assert shellRule.shell.find(){ c -> c.contains('cp project-settings.xml $HOME/.m2/settings.xml')}
    }

    @Test
    void canConfigureMavenGlobalSettings() {

        stepRule.step.mtaBuild(script: nullScript, globalSettingsFile: 'settings.xml')

        assert shellRule.shell.find(){ c -> c.contains('cp settings.xml $M2_HOME/conf/settings.xml')}
    }

    @Test
    void canConfigureNpmRegistry() {

        stepRule.step.mtaBuild(script: nullScript, defaultNpmRegistry: 'myNpmRegistry.com')

        assert shellRule.shell.find(){ c -> c.contains('npm config set registry myNpmRegistry.com')}
    }

    @Test
    void canConfigureMavenGlobalSettingsFromRemoteSource() {

        stepRule.step.mtaBuild(script: nullScript, globalSettingsFile: 'https://some.host/my-settings.xml')

        assert shellRule.shell.find(){ c -> c.contains('cp global-settings.xml $M2_HOME/conf/settings.xml')}
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

    class SettingsStub {
        String getContent() {
            return "<xml>sometext</xml>"
        }
    }
}
