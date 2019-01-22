import com.sap.piper.Utils
import hudson.AbortException
import org.jenkinsci.plugins.credentialsbinding.impl.CredentialNotFoundException
import org.junit.Assert
import org.junit.Before
import org.junit.BeforeClass
import org.junit.ClassRule
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.junit.rules.TemporaryFolder
import util.BasePiperTest
import util.CommandLineMatcher
import util.JenkinsCredentialsRule
import util.JenkinsLockRule
import util.JenkinsLoggingRule
import util.JenkinsPropertiesRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsShellCallRule.Type
import util.JenkinsStepRule
import util.JenkinsWithEnvRule
import util.Rules

class NeoDeployTest extends BasePiperTest {

    def toolJavaValidateCalled = false

    @ClassRule
    public static TemporaryFolder tmp = new TemporaryFolder()

    private ExpectedException thrown = new ExpectedException().none()
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLockRule lockRule = new JenkinsLockRule(this)


    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(new JenkinsPropertiesRule(this, propertiesFileName, configProperties))
        .around(thrown)
        .around(loggingRule)
        .around(shellRule)
        .around(new JenkinsCredentialsRule(this)
        .withCredentials('myCredentialsId', 'anonymous', '********')
        .withCredentials('CI_CREDENTIALS_ID', 'defaultUser', '********'))
        .around(stepRule)
        .around(lockRule)
        .around(new JenkinsWithEnvRule(this))


    private static workspacePath
    private static warArchiveName
    private static propertiesFileName
    private static archiveName
    private static configProperties


    @BeforeClass
    static void createTestFiles() {

        workspacePath = "${tmp.getRoot()}"
        warArchiveName = 'warArchive.war'
        propertiesFileName = 'config.properties'
        archiveName = 'archive.mtar'

        configProperties = new Properties()
        configProperties.put('account', 'trialuser123')
        configProperties.put('host', 'test.deploy.host.com')
        configProperties.put('application', 'testApp')

        tmp.newFile(warArchiveName) << 'dummy war archive'
        tmp.newFile(propertiesFileName) << 'dummy properties file'
        tmp.newFile(archiveName) << 'dummy archive'
    }

    @Before
    void init() {

        helper.registerAllowedMethod('dockerExecute', [Map, Closure], null)
        helper.registerAllowedMethod('fileExists', [String], { s -> return new File(workspacePath, s).exists() })
        mockShellCommands()

        nullScript.commonPipelineEnvironment.configuration = [steps: [neoDeploy: [neo: [host: 'test.deploy.host.com', account: 'trialuser123']]]]
    }

    @Test
    void straightForwardTestCompatibilityConfiguration(){
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.* status .*', 'Status: STARTED')

        nullScript.commonPipelineEnvironment.configuration = [
            steps: [
                neoDeploy: [
                    host: 'test.deploy.host.com',
                    account: 'trialuser123',
                    neoCredentialsId: 'myCredentialsId'
                ]]]

        stepRule.step.neoDeploy(script: nullScript,
            archivePath: warArchiveName,
            deployMode: 'warParams',
            applicationName: 'testApp',
            runtime: 'neo-javaee6-wp',
            runtimeVersion: '2.125',
            warAction: 'rolling-update',
            vmSize: 'lite')

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" rolling-update")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasSingleQuotedOption('application', 'testApp')
                .hasSingleQuotedOption('runtime', 'neo-javaee6-wp')
                .hasSingleQuotedOption('runtime-version', '2\\.125')
                .hasSingleQuotedOption('size', 'lite')
                .hasSingleQuotedOption('user', 'anonymous')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*\\.war'))
    }

    @Test
    void straightForwardTestConfigViaConfigProperties() {

        boolean buildStatusHasBeenSet = false
        boolean notifyOldConfigFrameworkUsed = false

        nullScript.commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
        nullScript.commonPipelineEnvironment.configuration = [:]

        nullScript.currentBuild = [setResult: { buildStatusHasBeenSet = true }]

        def utils = new Utils() {
            void pushToSWA(Map parameters, Map config) {
                notifyOldConfigFrameworkUsed = parameters.stepParam4
            }
        }

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName,
            neo: [credentialsId: 'myCredentialsId'],
            utils: utils
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy-mta")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasOption('synchronous', '')
                .hasSingleQuotedOption('user', 'anonymous')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*'))

        assert !buildStatusHasBeenSet
        assert notifyOldConfigFrameworkUsed
    }

    @Test
    void testConfigViaConfigPropertiesSetsBuildToUnstable() {

        def buildStatus = 'SUCCESS'

        nullScript.commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'test.deploy.host.com')
        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'trialuser123')
        nullScript.commonPipelineEnvironment.configuration = [:]

        nullScript.currentBuild = [setResult: { r -> buildStatus = r }]

        System.setProperty('com.sap.piper.featureFlag.buildUnstableWhenOldConfigFrameworkIsUsedByNeoDeploy',
            Boolean.TRUE.toString())

        try {
            stepRule.step.neoDeploy(script: nullScript,
                source: archiveName,
                neo:[credentialsId: 'myCredentialsId'],
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

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName,
            neo:[credentialsId: 'myCredentialsId'],
            utils: utils,
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy-mta")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasOption('synchronous', '')
                .hasSingleQuotedOption('user', 'anonymous')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*'))

        assert !notifyOldConfigFrameworkUsed
    }

    @Test
    void straightForwardTestConfigViaConfigurationAndViaConfigProperties() {

        nullScript.commonPipelineEnvironment.setConfigProperty('DEPLOY_HOST', 'configProperties.deploy.host.com')
        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'configPropsUser123')

        nullScript.commonPipelineEnvironment.configuration = [steps: [neoDeploy: [neo: [host   : 'configuration-frwk.deploy.host.com',
                                                                                  account: 'configurationFrwkUser123']]]]

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName,
            neo:[credentialsId: 'myCredentialsId']
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy-mta")
                .hasSingleQuotedOption('host', 'configuration-frwk\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'configurationFrwkUser123')
                .hasOption('synchronous', '')
                .hasSingleQuotedOption('user', 'anonymous')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*'))
    }

    @Test
    void archivePathFromCPETest() {
        nullScript.commonPipelineEnvironment.setMtarFilePath('archive.mtar')
        stepRule.step.neoDeploy(script: nullScript)

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy-mta")
                .hasSingleQuotedOption('source', 'archive.mtar'))
    }

    @Test
    void archivePathFromParamsHasHigherPrecedenceThanCPETest() {
        nullScript.commonPipelineEnvironment.setMtarFilePath('archive2.mtar')
        stepRule.step.neoDeploy(script: nullScript,
            source: "archive.mtar")

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy-mta")
                .hasSingleQuotedOption('source', 'archive.mtar'))
    }


    @Test
    void badCredentialsIdTest() {

        thrown.expect(CredentialNotFoundException)

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName,
            neo:[credentialsId: 'badCredentialsId']
        )
    }


    @Test
    void credentialsIdNotProvidedTest() {

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy-mta")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasOption('synchronous', '')
                .hasSingleQuotedOption('user', 'defaultUser')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*')
        )
    }


    @Test
    void neoHomeNotSetTest() {

        mockHomeVariablesNotSet()

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName
        )

        assert shellRule.shell.find { c -> c.contains('"neo.sh" deploy-mta') }
        assert loggingRule.log.contains('SAP Cloud Platform Console Client is on PATH.')
        assert loggingRule.log.contains("Using SAP Cloud Platform Console Client 'neo.sh'.")
    }


    @Test
    void neoHomeAsParameterTest() {

        mockHomeVariablesNotSet()

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName,
            neo:[credentialsId: 'myCredentialsId'],
            neoHome: '/param/neo'
        )

        assert shellRule.shell.find { c -> c = "\"/param/neo/tools/neo.sh\" deploy-mta" }
        assert loggingRule.log.contains("SAP Cloud Platform Console Client home '/param/neo' retrieved from configuration.")
        assert loggingRule.log.contains("Using SAP Cloud Platform Console Client '/param/neo/tools/neo.sh'.")
    }


    @Test
    void neoHomeFromEnvironmentTest() {

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName
        )

        assert shellRule.shell.find { c -> c.contains("\"/opt/neo/tools/neo.sh\" deploy-mta") }
        assert loggingRule.log.contains("SAP Cloud Platform Console Client home '/opt/neo' retrieved from environment.")
        assert loggingRule.log.contains("Using SAP Cloud Platform Console Client '/opt/neo/tools/neo.sh'.")
    }


    @Test
    void neoHomeFromCustomStepConfigurationTest() {

        mockHomeVariablesNotSet()

        nullScript.commonPipelineEnvironment.configuration = [steps: [neoDeploy: [neo: [host: 'test.deploy.host.com', account: 'trialuser123'], neoHome: '/config/neo']]]

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName
        )

        assert shellRule.shell.find { c -> c = "\"/config/neo/tools/neo.sh\" deploy-mta" }
        assert loggingRule.log.contains("SAP Cloud Platform Console Client home '/config/neo' retrieved from configuration.")
        assert loggingRule.log.contains("Using SAP Cloud Platform Console Client '/config/neo/tools/neo.sh'.")
    }


    @Test
    void archiveNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR source')

        stepRule.step.neoDeploy(script: nullScript)
    }


    @Test
    void wrongArchivePathProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('File wrongArchiveName cannot be found')
        stepRule.step.neoDeploy(script: nullScript,
            source: 'wrongArchiveName')
    }


    @Test
    void scriptNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR host')

        nullScript.commonPipelineEnvironment.configuration = [:]

        stepRule.step.neoDeploy(script: nullScript, source: archiveName)
    }

    @Test
    void mtaDeployModeTest() {

        stepRule.step.neoDeploy(script: nullScript, source: archiveName, deployMode: 'mta')

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy-mta")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasOption('synchronous', '')
                .hasSingleQuotedOption('user', 'defaultUser')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*'))

    }

    @Test
    void warFileParamsDeployModeTest() {

        stepRule.step.neoDeploy(script: nullScript,
            neo: [
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125',
                size: 'lite',
            ],
            deployMode: 'warParams',
            warAction: 'deploy',
            source: warArchiveName)

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasSingleQuotedOption('application', 'testApp')
                .hasSingleQuotedOption('runtime', 'neo-javaee6-wp')
                .hasSingleQuotedOption('runtime-version', '2\\.125')
                .hasSingleQuotedOption('size', 'lite')
                .hasSingleQuotedOption('user', 'defaultUser')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*\\.war'))

    }

    @Test
    void warFileParamsDeployModeRollingUpdateTest() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.* status .*', 'Status: STARTED')

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'warParams',
            warAction: 'rolling-update',
            neo: [
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125',
                size: 'lite'
            ]
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" rolling-update")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasSingleQuotedOption('application', 'testApp')
                .hasSingleQuotedOption('runtime', 'neo-javaee6-wp')
                .hasSingleQuotedOption('runtime-version', '2\\.125')
                .hasSingleQuotedOption('size', 'lite')
                .hasSingleQuotedOption('user', 'defaultUser')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*\\.war'))
    }

    @Test
    void warFirstTimeRollingUpdateTest() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.* status .*', 'ERROR: Application [testApp] not found')

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'warParams',
            warAction: 'rolling-update',
            neo: [
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125'
            ]
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher()
                .hasProlog("\"/opt/neo/tools/neo.sh\" deploy")
                .hasSingleQuotedOption('application', 'testApp'))
    }

    void warNotStartedRollingUpdateTest() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.* status .*', 'Status: STOPPED')

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'warParams',
            warAction: 'rolling-update',
            neo: [
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125'
            ]
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher()
                .hasProlog("\"/opt/neo/tools/neo.sh\" deploy")
                .hasSingleQuotedOption('application', 'testApp'))
    }

    @Test
    void showLogsOnFailingDeployment() {

        thrown.expect(Exception)
        shellRule.failExecution(Type.REGEX, '.* deploy .*')

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'warParams',
            warAction: 'deploy',
            neo: [
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125'
            ]
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("cat /var/log/neo/*"))
    }

    @Test
    void warPropertiesFileDeployModeTest() {

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'warPropertiesFile',
            warAction: 'deploy',
            neo: [
                propertiesFile: propertiesFileName,
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125',
                size: 'lite'
            ]
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" deploy")
                .hasArgument("config.properties")
                .hasSingleQuotedOption('user', 'defaultUser')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*\\.war'))
    }

    @Test
    void warPropertiesFileDeployModeRollingUpdateTest() {

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, '.* status .*', 'Status: STARTED')

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'warPropertiesFile',
            warAction: 'rolling-update',
            neo: [
                propertiesFile: propertiesFileName,
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125',
                size: 'lite'
            ])

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("\"/opt/neo/tools/neo.sh\" rolling-update")
                .hasArgument('config.properties')
                .hasSingleQuotedOption('user', 'defaultUser')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*\\.war'))
    }

    @Test
    void applicationNameNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR application')

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'warParams',
            neo: [
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125'
            ]
        )
    }

    @Test
    void runtimeNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtime')

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            neo: [
                application: 'testApp',
                runtimeVersion: '2.125'
            ],
            deployMode: 'warParams')
    }

    @Test
    void runtimeVersionNotProvidedTest() {

        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR runtimeVersion')

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            neo: [
                application: 'testApp',
                runtime: 'neo-javaee6-wp'
            ],
            deployMode: 'warParams')
    }

    @Test
    void illegalDeployModeTest() {

        thrown.expect(Exception)
        thrown.expectMessage("Invalid deployMode = 'illegalMode'. Valid 'deployMode' values are: [mta, warParams, warPropertiesFile].")

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'illegalMode',
            warAction: 'deploy',
            neo: [
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125',
                size: 'lite'
            ])
    }

    @Test
    void illegalWARActionTest() {

        thrown.expect(Exception)
        thrown.expectMessage("Invalid warAction = 'illegalWARAction'. Valid 'warAction' values are: [deploy, rolling-update].")

        stepRule.step.neoDeploy(script: nullScript,
            source: warArchiveName,
            deployMode: 'warParams',
            warAction: 'illegalWARAction',
            neo: [
                application: 'testApp',
                runtime: 'neo-javaee6-wp',
                runtimeVersion: '2.125',
                size: 'lite'
            ])
    }

    @Test
    void deployHostProvidedAsDeprecatedParameterTest() {

        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'configPropsUser123')

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName,
            deployHost: "my.deploy.host.com"
        )

        assert loggingRule.log.contains("[WARNING][neoDeploy] Deprecated parameter 'deployHost' is used. This will not work anymore in future versions. Use parameter 'host' instead.")
    }

    @Test
    void deployAccountProvidedAsDeprecatedParameterTest() {

        nullScript.commonPipelineEnvironment.setConfigProperty('CI_DEPLOY_ACCOUNT', 'configPropsUser123')

        stepRule.step.neoDeploy(script: nullScript,
            source: archiveName,
            neo: [
                host: "my.deploy.host.com",
            ],
            deployAccount: "myAccount"
        )

        assert loggingRule.log.contains("Deprecated parameter 'deployAccount' is used. This will not work anymore in future versions. Use parameter 'account' instead.")
    }

    private mockShellCommands() {
        String javaVersion = '''openjdk version \"1.8.0_121\"
                    OpenJDK Runtime Environment (build 1.8.0_121-8u121-b13-1~bpo8+1-b13)
                    OpenJDK 64-Bit Server VM (build 25.121-b13, mixed mode)'''
        shellRule.setReturnValue(Type.REGEX, '.*java -version.*', javaVersion)

        String neoVersion = '''SAP Cloud Platform Console Client
                    SDK version    : 3.39.10
                    Runtime        : neo-java-web'''
        shellRule.setReturnValue(Type.REGEX, '.*neo.sh version.*', neoVersion)

        shellRule.setReturnValue(Type.REGEX, '.*JAVA_HOME.*', '/opt/java')
        shellRule.setReturnValue(Type.REGEX, '.*NEO_HOME.*', '/opt/neo')
        shellRule.setReturnValue(Type.REGEX, '.*which java.*', 0)
        shellRule.setReturnValue(Type.REGEX, '.*which neo.*', 0)
    }

    private mockHomeVariablesNotSet() {
        shellRule.setReturnValue(Type.REGEX, '.*JAVA_HOME.*', '')
        shellRule.setReturnValue(Type.REGEX, '.*NEO_HOME.*', '')
        shellRule.setReturnValue(Type.REGEX, '.*which java.*', 0)
        shellRule.setReturnValue(Type.REGEX, '.*which neo.*', 0)
    }
}
