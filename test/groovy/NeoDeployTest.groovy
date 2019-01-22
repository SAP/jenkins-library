import com.sap.piper.Utils
import hudson.AbortException

import org.junit.rules.TemporaryFolder

import org.junit.BeforeClass
import org.junit.ClassRule
import org.junit.Ignore

import java.util.Map

import org.hamcrest.BaseMatcher
import org.hamcrest.Description
import org.jenkinsci.plugins.credentialsbinding.impl.CredentialNotFoundException
import org.junit.Assert
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

class NeoDeployTest extends BasePiperTest {

    def toolJavaValidateCalled = false

    @ClassRule
    public static TemporaryFolder tmp = new TemporaryFolder()

    private ExpectedException thrown = new ExpectedException().none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jlr)
        .around(shellRule)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('myCredentialsId', 'anonymous', '********')
            .withCredentials('CI_CREDENTIALS_ID', 'defaultUser', '********'))
        .around(jsr)

    private static workspacePath
    private static warArchiveName
    private static propertiesFileName
    private static archiveName


    @BeforeClass
    static void createTestFiles() {

        workspacePath = "${tmp.getRoot()}"
        warArchiveName = 'warArchive.war'
        propertiesFileName = 'config.properties'
        archiveName = 'archive.mtar'

        tmp.newFile(warArchiveName) << 'dummy war archive'
        tmp.newFile(propertiesFileName) << 'dummy properties file'
        tmp.newFile(archiveName) << 'dummy archive'
    }

    @Before
    void init() {

        helper.registerAllowedMethod('dockerExecute', [Map, Closure], null)
        helper.registerAllowedMethod('fileExists', [String], { s -> return new File(workspacePath, s).exists() })
        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionWithEnvVars(m) })

        nullScript.commonPipelineEnvironment.configuration = [steps:[neoDeploy: [host: 'test.deploy.host.com', account: 'trialuser123']]]
    }


    @Test
    void straightForwardTestConfigViaConfigProperties() {

        boolean buildStatusHasBeenSet = false
        boolean notifyOldConfigFrameworkUsed = false

        nullScript.commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
        nullScript.commonPipelineEnvironment.configuration = [:]

        nullScript.currentBuild = [setResult: {buildStatusHasBeenSet = true}]

        def utils = new Utils() {
            void pushToSWA(Map parameters, Map config) {
                notifyOldConfigFrameworkUsed = parameters.stepParam4
            }
        }

        jsr.step.neoDeploy(script: nullScript,
                       archivePath: archiveName,
                       neoCredentialsId: 'myCredentialsId',
                       utils: utils
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy-mta")
                                    .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                                    .hasSingleQuotedOption('account', 'trialuser123')
                                    .hasOption('synchronous', '')
                                    .hasSingleQuotedOption('user', 'anonymous')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*'))

        assert !buildStatusHasBeenSet
        assert notifyOldConfigFrameworkUsed
    }

    @Test
    void testConfigViaConfigPropertiesSetsBuildToUnstable() {

        def buildStatus = 'SUCCESS'

        nullScript.commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
        nullScript.commonPipelineEnvironment.configuration = [:]

        nullScript.currentBuild = [setResult: { r -> buildStatus = r}]

        System.setProperty('com.sap.piper.featureFlag.buildUnstableWhenOldConfigFrameworkIsUsedByNeoDeploy',
            Boolean.TRUE.toString())

        try {
            jsr.step.neoDeploy(script: nullScript,
                               archivePath: archiveName,
                               neoCredentialsId: 'myCredentialsId',
                               utils: utils
            )
        } finally {
            System.clearProperty('com.sap.piper.featureFlag.buildUnstableWhenOldConfigFrameworkIsUsedByNeoDeploy')
        }

        assert buildStatus == 'UNSTABLE'
    }

    @Test
    void straightForwardTestConfigViaConfiguration() {

        boolean notifyOldConfigFrameworkUsed = true

        def utils = new Utils() {
            void pushToSWA(Map parameters, Map config) {
                notifyOldConfigFrameworkUsed = parameters.stepParam4
            }
        }

        jsr.step.neoDeploy(script: nullScript,
            archivePath: archiveName,
            neoCredentialsId: 'myCredentialsId',
            utils: utils,
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy-mta")
                                    .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                                    .hasSingleQuotedOption('account', 'trialuser123')
                                    .hasOption('synchronous', '')
                                    .hasSingleQuotedOption('user', 'anonymous')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*'))

        assert !notifyOldConfigFrameworkUsed
    }

    @Test
    void straightForwardTestConfigViaConfigurationAndViaConfigProperties() {

        nullScript.commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'configProperties.deploy.host.com')
        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'configPropsUser123')

        nullScript.commonPipelineEnvironment.configuration = [steps:[neoDeploy: [host: 'configuration-frwk.deploy.host.com',
                                                account: 'configurationFrwkUser123']]]

        jsr.step.neoDeploy(script: nullScript,
            archivePath: archiveName,
            neoCredentialsId: 'myCredentialsId'
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy-mta")
                                    .hasSingleQuotedOption('host', 'configuration-frwk\\.deploy\\.host\\.com')
                                    .hasSingleQuotedOption('account', 'configurationFrwkUser123')
                                    .hasOption('synchronous', '')
                                    .hasSingleQuotedOption('user', 'anonymous')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*'))
    }

    @Test
    void archivePathFromCPETest() {
        nullScript.commonPipelineEnvironment.setMtarFilePath('archive.mtar')
        jsr.step.neoDeploy(script: nullScript)

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy-mta")
                                    .hasDoubleQuotedOption('source', 'archive.mtar'))
    }

    @Test
    void archivePathFromParamsHasHigherPrecedenceThanCPETest() {
        nullScript.commonPipelineEnvironment.setMtarFilePath('archive2.mtar')
        jsr.step.neoDeploy(script: nullScript,
                      archivePath: "archive.mtar")

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy-mta")
                                    .hasDoubleQuotedOption('source', 'archive.mtar'))
    }


    @Test
    void badCredentialsIdTest() {

        thrown.expect(CredentialNotFoundException)

        jsr.step.neoDeploy(script: nullScript,
                       archivePath: archiveName,
                       neoCredentialsId: 'badCredentialsId'
        )
    }


    @Test
    void credentialsIdNotProvidedTest() {

        jsr.step.neoDeploy(script: nullScript,
                       archivePath: archiveName
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy-mta")
                                    .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                                    .hasSingleQuotedOption('account', 'trialuser123')
                                    .hasOption('synchronous', '')
                                    .hasSingleQuotedOption('user', 'defaultUser')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*'))
    }


    @Test
    void neoHomeNotSetTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionWithPath(m) })

        jsr.step.neoDeploy(script: nullScript,
                       archivePath: archiveName
        )

        assert shellRule.shell.find { c -> c.contains('"neo.sh" deploy-mta') }
        assert jlr.log.contains('SAP Cloud Platform Console Client is on PATH.')
        assert jlr.log.contains("Using SAP Cloud Platform Console Client 'neo.sh'.")
    }


    @Test
    void neoHomeAsParameterTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionWithPath(m) })

        jsr.step.neoDeploy(script: nullScript,
                       archivePath: archiveName,
                       neoCredentialsId: 'myCredentialsId',
                       neoHome: '/param/neo'
        )

        assert shellRule.shell.find{ c -> c = "\"/param/neo/tools/neo.sh\" deploy-mta" }
        assert jlr.log.contains("SAP Cloud Platform Console Client home '/param/neo' retrieved from configuration.")
        assert jlr.log.contains("Using SAP Cloud Platform Console Client '/param/neo/tools/neo.sh'.")
    }


    @Test
    void neoHomeFromEnvironmentTest() {

        jsr.step.neoDeploy(script: nullScript,
                       archivePath: archiveName
        )

        assert shellRule.shell.find { c -> c.contains("\"/opt/neo/tools/neo.sh\" deploy-mta")}
        assert jlr.log.contains("SAP Cloud Platform Console Client home '/opt/neo' retrieved from environment.")
        assert jlr.log.contains("Using SAP Cloud Platform Console Client '/opt/neo/tools/neo.sh'.")
    }


    @Test
    void neoHomeFromCustomStepConfigurationTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> getVersionWithPath(m) })

        nullScript.commonPipelineEnvironment.configuration = [steps:[neoDeploy: [host: 'test.deploy.host.com', account: 'trialuser123', neoHome: '/config/neo']]]

        jsr.step.neoDeploy(script: nullScript,
                       archivePath: archiveName
        )

        assert shellRule.shell.find { c -> c = "\"/config/neo/tools/neo.sh\" deploy-mta"}
        assert jlr.log.contains("SAP Cloud Platform Console Client home '/config/neo' retrieved from configuration.")
        assert jlr.log.contains("Using SAP Cloud Platform Console Client '/config/neo/tools/neo.sh'.")
    }


    @Test
    void archiveNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('Archive path not configured (parameter "archivePath").')

        jsr.step.neoDeploy(script: nullScript)
    }


    @Test
    void wrongArchivePathProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('Archive cannot be found')

        jsr.step.neoDeploy(script: nullScript,
                       archivePath: 'wrongArchiveName')
    }


    @Test
    void scriptNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR host')

        nullScript.commonPipelineEnvironment.configuration = [:]

        jsr.step.neoDeploy(script: nullScript, archivePath: archiveName)
    }

    @Test
    void mtaDeployModeTest() {

        jsr.step.neoDeploy(script: nullScript, archivePath: archiveName, deployMode: 'mta')

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy-mta")
                                    .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                                    .hasSingleQuotedOption('account', 'trialuser123')
                                    .hasOption('synchronous', '')
                                    .hasSingleQuotedOption('user', 'defaultUser')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*'))

    }

    @Test
    void warFileParamsDeployModeTest() {

        jsr.step.neoDeploy(script: nullScript,
                             applicationName: 'testApp',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125',
                             deployMode: 'warParams',
                             vmSize: 'lite',
                             warAction: 'deploy',
                             archivePath: warArchiveName)

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy")
                                    .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                                    .hasSingleQuotedOption('account', 'trialuser123')
                                    .hasSingleQuotedOption('application', 'testApp')
                                    .hasSingleQuotedOption('runtime', 'neo-javaee6-wp')
                                    .hasSingleQuotedOption('runtime-version', '2\\.125')
                                    .hasSingleQuotedOption('size', 'lite')
                                    .hasSingleQuotedOption('user', 'defaultUser')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*\\.war'))

    }

    @Test
    void warFileParamsDeployModeRollingUpdateTest() {

        jsr.step.neoDeploy(script: nullScript,
                             archivePath: warArchiveName,
                             deployMode: 'warParams',
                             applicationName: 'testApp',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125',
                             warAction: 'rolling-update',
                             vmSize: 'lite')

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" rolling-update")
                                    .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                                    .hasSingleQuotedOption('account', 'trialuser123')
                                    .hasSingleQuotedOption('application', 'testApp')
                                    .hasSingleQuotedOption('runtime', 'neo-javaee6-wp')
                                    .hasSingleQuotedOption('runtime-version', '2\\.125')
                                    .hasSingleQuotedOption('size', 'lite')
                                    .hasSingleQuotedOption('user', 'defaultUser')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*\\.war'))
    }

    @Test
    void warPropertiesFileDeployModeTest() {

        jsr.step.neoDeploy(script: nullScript,
                             archivePath: warArchiveName,
                             deployMode: 'warPropertiesFile',
                             propertiesFile: propertiesFileName,
                             applicationName: 'testApp',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125',
                             warAction: 'deploy',
                             vmSize: 'lite')

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" deploy")
                                    .hasArgument("config.properties")
                                    .hasSingleQuotedOption('user', 'defaultUser')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*\\.war'))
    }

    @Test
    void warPropertiesFileDeployModeRollingUpdateTest() {

        jsr.step.neoDeploy(script: nullScript,
                             archivePath: warArchiveName,
                             deployMode: 'warPropertiesFile',
                             propertiesFile: propertiesFileName,
                             applicationName: 'testApp',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125',
                             warAction: 'rolling-update',
                             vmSize: 'lite')

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("#!/bin/bash \"/opt/neo/tools/neo.sh\" rolling-update")
                                    .hasArgument('config.properties')
                                    .hasSingleQuotedOption('user', 'defaultUser')
                                    .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                                    .hasDoubleQuotedOption('source', '.*\\.war'))
    }

    @Test
    void applicationNameNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR applicationName')

        jsr.step.neoDeploy(script: nullScript,
                             archivePath: warArchiveName,
                             deployMode: 'warParams',
                             runtime: 'neo-javaee6-wp',
                             runtimeVersion: '2.125'
            )
    }

    @Test
    void runtimeNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtime')

        jsr.step.neoDeploy(script: nullScript,
                             archivePath: warArchiveName,
                             applicationName: 'testApp',
                             deployMode: 'warParams',
                             runtimeVersion: '2.125')
    }

    @Test
    void runtimeVersionNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtimeVersion')

        jsr.step.neoDeploy(script: nullScript,
                             archivePath: warArchiveName,
                             applicationName: 'testApp',
                             deployMode: 'warParams',
                             runtime: 'neo-javaee6-wp')
    }

    @Test
    void illegalDeployModeTest() {

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid deployMode = 'illegalMode'. Valid 'deployMode' values are: [mta, warParams, warPropertiesFile]")

        jsr.step.neoDeploy(script: nullScript,
            archivePath: warArchiveName,
            deployMode: 'illegalMode',
            applicationName: 'testApp',
            runtime: 'neo-javaee6-wp',
            runtimeVersion: '2.125',
            warAction: 'deploy',
            vmSize: 'lite')
    }

    @Test
    void illegalVMSizeTest() {

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid vmSize = 'illegalVM'. Valid 'vmSize' values are: [lite, pro, prem, prem-plus].")

        jsr.step.neoDeploy(script: nullScript,
            archivePath: warArchiveName,
            deployMode: 'warParams',
            applicationName: 'testApp',
            runtime: 'neo-javaee6-wp',
            runtimeVersion: '2.125',
            warAction: 'deploy',
            vmSize: 'illegalVM')
    }

    @Test
    void illegalWARActionTest() {

        thrown.expect(Exception)
        thrown.expectMessage("[neoDeploy] Invalid warAction = 'illegalWARAction'. Valid 'warAction' values are: [deploy, rolling-update].")

        jsr.step.neoDeploy(script: nullScript,
            archivePath: warArchiveName,
            deployMode: 'warParams',
            applicationName: 'testApp',
            runtime: 'neo-javaee6-wp',
            runtimeVersion: '2.125',
            warAction: 'illegalWARAction',
            vmSize: 'lite')
    }

    @Test
    void deployHostProvidedAsDeprecatedParameterTest() {

        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'configPropsUser123')

        jsr.step.neoDeploy(script: nullScript,
                             archivePath: archiveName,
                             deployHost: "my.deploy.host.com"
        )

        assert jlr.log.contains("[WARNING][neoDeploy] Deprecated parameter 'deployHost' is used. This will not work anymore in future versions. Use parameter 'host' instead.")
    }

    @Test
    void deployAccountProvidedAsDeprecatedParameterTest() {

        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'configPropsUser123')

        jsr.step.neoDeploy(script: nullScript,
                             archivePath: archiveName,
                             host: "my.deploy.host.com",
                             deployAccount: "myAccount"
        )

        assert jlr.log.contains("Deprecated parameter 'deployAccount' is used. This will not work anymore in future versions. Use parameter 'account' instead.")
    }


    private getVersionWithEnvVars(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('neo.sh version')) {
            return '''SAP Cloud Platform Console Client
                    SDK version    : 3.39.10
                    Runtime        : neo-java-web'''
        } else {
            return getEnvVars(m)
        }
    }

    private getVersionWithPath(Map m) {

        if(m.script.contains('java -version')) {
            return '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        } else if(m.script.contains('neo.sh version')) {
            return '''SAP Cloud Platform Console Client
                    SDK version    : 3.39.10
                    Runtime        : neo-java-web'''
        } else {
            return getPath(m)
        }
    }

    private getEnvVars(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return '/opt/java'
        } else if(m.script.contains('NEO_HOME')) {
            return '/opt/neo'
        } else if (m.script.contains('which java')) {
            return 0
        } else if (m.script.contains('which neo')) {
            return 0
        } else {
            return 0
        }
    }

    private getPath(Map m) {

        if(m.script.contains('JAVA_HOME')) {
            return ''
        } else if(m.script.contains('NEO_HOME')) {
            return ''
        } else if (m.script.contains('which java')) {
            return 0
        } else if (m.script.contains('which neo')) {
            return 0
        } else {
            return 0
        }
    }

    class CommandLineMatcher extends BaseMatcher {

        String prolog
        Set<String> args = (Set)[]
        Set<MapEntry> opts = (Set) []

        String hint = ''

        CommandLineMatcher hasProlog(prolog) {
            this.prolog = prolog
            return this
        }

        CommandLineMatcher hasDoubleQuotedOption(String key, String value) {
            hasOption(key, "\"${value}\"")
            return this
        }

        CommandLineMatcher hasSingleQuotedOption(String key, String value) {
            hasOption(key, "\'${value}\'")
            return this
        }

        CommandLineMatcher hasOption(String key, String value) {
            this.opts.add(new MapEntry(key, value))
            return this
        }

        CommandLineMatcher hasArgument(String arg) {
            this.args.add(arg)
            return this
        }

        @Override
        boolean matches(Object o) {

            for(String cmd : o) {

                hint = ''
                boolean matches = true

                if(!cmd.matches(/${prolog}.*/)) {
                    hint = "A command line starting with \'${prolog}\'."
                    matches = false
                }

                for(MapEntry opt : opts) {
                    if(! cmd.matches(/.*[\s]*--${opt.key}[\s]*${opt.value}.*/)) {
                        hint = "A command line containing option \'${opt.key}\' with value \'${opt.value}\'"
                        matches = false
                    }
                }

                for(String arg : args) {
                    if(! cmd.matches(/.*[\s]*${arg}[\s]*.*/)) {
                        hint = "A command line having argument '${arg}'."
                        matches = false
                    }
                }

                if(matches)
                    return true
            }

            return false
        }

        @Override
        public void describeTo(Description description) {
            description.appendText(hint)
        }
    }
}
