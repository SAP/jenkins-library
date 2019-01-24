import util.CommandLineMatcher
import util.JenkinsLockRule
import util.JenkinsWithEnvRule

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.subString
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers
import org.junit.Assert
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import com.sap.piper.JenkinsUtils

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

class FioriOnCloudPlatformPipelineTest extends BasePiperTest {

    /*  This scenario builds a fiori app and deploys it into an neo account.
        The build is performed using mta, which delegates to grunt. grunt in
        turn makes use of the 'sap/grunt-sapui5-bestpractice-build' plugin.
        The dependencies are resolved via npm.

        In order to run the scenario the project needs to fullfill these
        prerequisites:

        Build tools:
        *   mta.jar available
        *   npm installed

        Project configuration:
        *   sap registry `@sap:registry=https://npm.sap.com` configured in
            .npmrc (either in the project or on any other suitable level)
        *   dependency to `@sap/grunt-sapui5-bestpractice-build` declared in
            package.json
        *   npmTask `@sap/grunt-sapui5-bestpractice-build` loaded inside
            Gruntfile.js and configure default tasks (e.g. lint, clean, build)
        *   mta.yaml
    */

    JenkinsStepRule stepRule = new JenkinsStepRule(this)
    JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLockRule jlr = new JenkinsLockRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(stepRule)
        .around(shellRule)
        .around(jlr)
        .around(new JenkinsWithEnvRule(this))
        .around(new JenkinsCredentialsRule(this)
        .withCredentials('CI_CREDENTIALS_ID', 'foo', 'terceSpot'))

    @Before
    void setup() {
        //
        // needed since we have dockerExecute inside mtaBuild
        JenkinsUtils.metaClass.static.isPluginActive = {def s -> false}

        //
        // Things we validate:
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*echo \\$JAVA_HOME.*', '/opt/sap/java')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*echo \\$MTA_JAR_LOCATION.*', '/opt/sap')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.*echo \\$NEO_HOME.*', '/opt/sap/neo')
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".*bin/java -version.*", '1.8.0') // the java version
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, ".*bin/java -jar .*mta.jar", '1.36.0') // the mta version

        //
        // there is a check for the mta.yaml file and for the deployable test.mtar file
        helper.registerAllowedMethod('fileExists', [String],{

            it ->

            // called inside mtaBuild, this file contains build config
            it == 'mta.yaml' ||

            // called inside neo deploy, this file gets deployed
            it == 'test.mtar'
        })

        //
        // the properties below we read out of the yaml file
        readYamlRule.registerYaml('mta.yaml', ('''
                                       |ID : "test"
                                       |PATH : "."
                                       |''' as CharSequence).stripMargin())

        //
        // we need the path variable since we extend the path in the mtaBuild step. In order
        // to be able to extend the path we have to have some initial value.
        binding.setVariable('PATH', '/usr/bin')

        helper.registerAllowedMethod('pwd', [], { return "./" })
    }

    @Test
    void straightForwardTest() {

        nullScript
            .commonPipelineEnvironment
                .configuration =  [steps:
                                    [neoDeploy:
                                        [ host: 'hana.example.com',
                                          account: 'myTestAccount',
                                        ]
                                    ]
                                ]

        stepRule.step.fioriOnCloudPlatformPipeline(script: nullScript)

        //
        // the mta build call:
        assertThat(shellRule.shell, hasItem(
                                allOf(  containsString('java -jar /opt/sap/mta.jar'),
                                        containsString('--mtar test.mtar'),
                                        containsString('--build-target=NEO'),
                                        containsString('build'))))

        //
        // the deployable is exchanged between the involved steps via this property:
        assertThat(nullScript.commonPipelineEnvironment.getMtarFilePath(), is(equalTo('test.mtar')))

        //
        // the neo deploy call:
        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher()
                .hasProlog("\"/opt/sap/neo/tools/neo.sh\" deploy-mta")
                .hasSingleQuotedOption('host', 'hana\\.example\\.com')
                .hasSingleQuotedOption('account', 'myTestAccount')
                .hasSingleQuotedOption('password', 'terceSpot')
                .hasSingleQuotedOption('user', 'foo')
                .hasSingleQuotedOption('source', 'test.mtar')
                .hasArgument('synchronous'))
    }
}
