import org.apache.commons.exec.*
import hudson.AbortException

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

class ToolValidateTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException().none()
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(loggingRule)
        .around(stepRule)

    def home = 'home'

    @Test
    void nullHomeTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'home' can not be null or empty.")

        stepRule.step.toolValidate(tool: 'java')
    }

    @Test
    void emptyHomeTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'home' can not be null or empty.")

        stepRule.step.toolValidate(tool: 'java', home: '')
    }

    @Test
    void nullToolTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 0 })

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'tool' can not be null or empty.")

        stepRule.step.toolValidate(tool: null, home: home)
    }

    @Test
    void emptyToolTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 0 })

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'tool' can not be null or empty.")

        stepRule.step.toolValidate(tool: '', home: home)
    }

    @Test
    void invalidToolTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 0 })

        thrown.expect(AbortException)
        thrown.expectMessage("The tool 'test' is not supported.")

        stepRule.step.toolValidate(tool: 'test', home: home)
    }

    @Test
    void unableToValidateJavaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The verification of Java failed.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoVersion(m) })

        stepRule.step.toolValidate(tool: 'java', home: home)
    }

    @Test
    void unableToValidateMtaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The verification of SAP Multitarget Application Archive Builder failed.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoVersion(m) })

        stepRule.step.toolValidate(tool: 'mta', home: home)
    }

    @Test
    void unableToValidateNeoTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The verification of SAP Cloud Platform Console Client failed.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoVersion(m) })

        stepRule.step.toolValidate(tool: 'neo', home: home)
    }

    @Test
    void unableToValidateCmTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The verification of Change Management Command Line Interface failed.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getNoVersion(m) })

        stepRule.step.toolValidate(tool: 'cm', home: home)
    }

    @Test
    void validateIncompatibleVersionJavaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The installed version of Java is 1.7.0.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        stepRule.step.toolValidate(tool: 'java', home: home)
    }

    @Test
    void validateIncompatibleVersionMtaTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The installed version of SAP Multitarget Application Archive Builder is 1.0.5.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })

        stepRule.step.toolValidate(tool: 'mta', home: home)
    }

    @Test
    void validateCmIncompatibleVersionTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('The installed version of Change Management Command Line Interface is 0.0.0.')

        helper.registerAllowedMethod('sh', [Map], { Map m -> getIncompatibleVersion(m) })
        binding.setVariable('tool', 'cm')

        stepRule.step.toolValidate(tool: 'cm', home: home)
    }

    @Test
    void validateJavaTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        stepRule.step.toolValidate(tool: 'java', home: home)

        assert loggingRule.log.contains('Verifying Java version 1.8.0 or compatible version.')
        assert loggingRule.log.contains('Java version 1.8.0 is installed.')
    }

    @Test
    void validateMtaTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        stepRule.step.toolValidate(tool: 'mta', home: home)

        assert loggingRule.log.contains('Verifying SAP Multitarget Application Archive Builder version 1.0.6 or compatible version.')
        assert loggingRule.log.contains('SAP Multitarget Application Archive Builder version 1.0.6 is installed.')
    }

    @Test
    void validateNeoTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        stepRule.step.toolValidate(tool: 'neo', home: home)
    }

    @Test
    void validateCmTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersion(m) })

        stepRule.step.toolValidate(tool: 'cm', home: home)

        assert loggingRule.log.contains('Verifying Change Management Command Line Interface version 0.0.1 or compatible version.')
        assert loggingRule.log.contains('Change Management Command Line Interface version 0.0.1 is installed.')
    }


    private getToolHome(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return '/env/java'
        } else if(m.script.contains('MTA_JAR_LOCATION')) {
            return '/env/mta/mta.jar'
        } else if(m.script.contains('NEO_HOME')) {
            return '/env/neo'
        } else if(m.script.contains('CM_CLI_HOME')) {
            return '/env/cmclient'
        } else {
            return 0
        }
    }

    private getNoVersion(Map m) {

        if(m.script.contains('java -version')) {
            throw new AbortException('script returned exit code 127')
        } else if(m.script.contains('mta.jar -v')) {
            throw new AbortException('script returned exit code 127')
        } else if(m.script.contains('neo.sh version')) {
            throw new AbortException('script returned exit code 127')
        } else if(m.script.contains('cmclient -v')) {
            throw new AbortException('script returned exit code 127')
        } else {
            return getToolHome(m)
        }
    }

    private getVersion(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta.jar -v')) {
            return '1.0.6'
        } else if(m.script.contains('neo.sh version')) {
            return '''SAP Cloud Platform Console Client
                    SDK version    : 3.39.10
                    Runtime        : neo-java-web'''
        } else if(m.script.contains('cmclient -v')) {
            return '0.0.1-beta-2 : fc9729964a6acf5c1cad9c6f9cd6469727625a8e'
        } else {
            return getToolHome(m)
        }
    }

    private getIncompatibleVersion(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.7.0_121\"
                    OpenJDK Runtime Environment (build 1.7.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('mta.jar -v')) {
            return '1.0.5'
        } else if(m.script.contains('neo.sh version')) {
            return '''SAP Cloud Platform Console Client
                    SDK version    : 1.126.51
                    Runtime        : neo-java-web'''
        } else if(m.script.contains('cmclient -v')) {
            return '0.0.0-beta-1 : fc9729964a6acf5c1cad9c6f9cd6469727625a8e'
        } else {
            return getToolHome(m)
        }
    }
}
