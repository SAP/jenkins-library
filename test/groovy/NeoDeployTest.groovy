import com.sap.piper.Utils
import hudson.AbortException

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.not

import org.hamcrest.Matchers
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
        helper.registerAllowedMethod('pwd', [], { return workspacePath })

        nullScript.commonPipelineEnvironment.configuration = [steps: [neoDeploy: [neo: [host: 'test.deploy.host.com', account: 'trialuser123']]]]
    }

    @Test
    void straightForwardTestConfigViaParameters() {

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
            new CommandLineMatcher().hasProlog("neo.sh deploy-mta")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasOption('synchronous', '')
                .hasSingleQuotedOption('user', 'anonymous')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*'))

        assert !notifyOldConfigFrameworkUsed
    }

    @Test
    void straightForwardTestConfigViaConfiguration() {

        nullScript.commonPipelineEnvironment.configuration = [steps: [
            neoDeploy: [
                neo: [
                    host: 'configuration-frwk.deploy.host.com',
                    account: 'configurationFrwkUser123'
                ],
                source: archiveName
            ]
        ]]

        stepRule.step.neoDeploy(script: nullScript,
            neo:[credentialsId: 'myCredentialsId']
        )

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("neo.sh deploy-mta")
                .hasSingleQuotedOption('host', 'configuration-frwk\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'configurationFrwkUser123')
                .hasOption('synchronous', '')
                .hasSingleQuotedOption('user', 'anonymous')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', archiveName))
    }

    @Test
    void archivePathFromCPETest() {
        nullScript.commonPipelineEnvironment.setMtarFilePath('archive.mtar')
        stepRule.step.neoDeploy(script: nullScript)

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("neo.sh deploy-mta")
                .hasSingleQuotedOption('source', 'archive.mtar'))
    }

    @Test
    void archivePathFromParamsHasHigherPrecedenceThanCPETest() {
        nullScript.commonPipelineEnvironment.setMtarFilePath('archive2.mtar')
        stepRule.step.neoDeploy(script: nullScript,
            source: "archive.mtar")

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("neo.sh deploy-mta")
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
            new CommandLineMatcher().hasProlog("neo.sh deploy-mta")
                .hasSingleQuotedOption('host', 'test\\.deploy\\.host\\.com')
                .hasSingleQuotedOption('account', 'trialuser123')
                .hasOption('synchronous', '')
                .hasSingleQuotedOption('user', 'defaultUser')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*')
        )
    }

    @Test
    void wrongArchivePathProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('File wrongArchiveName cannot be found')
        stepRule.step.neoDeploy(script: nullScript,
            source: 'wrongArchiveName')
    }


    @Test
    void sanityChecksDeployModeMTATest() {

        thrown.expect(Exception)
        thrown.expectMessage(
            allOf(
                containsString('ERROR - NO VALUE AVAILABLE FOR:'),
                containsString('neo/host'),
                containsString('neo/account'),
                containsString('source')))

        nullScript.commonPipelineEnvironment.configuration = [:]

        // deployMode mta is the default, but for the sake of transparency it is better to repeat it.
        stepRule.step.neoDeploy(script: nullScript, deployMode: 'mta')
    }

    @Test
    public void sanityChecksDeployModeWarPropertiesFileTest() {

        thrown.expect(IllegalArgumentException)
        // using this deploy mode 'account' and 'host' are provided by the properties file
        thrown.expectMessage(
            allOf(
                containsString('ERROR - NO VALUE AVAILABLE FOR source'),
                not(containsString('neo/host')),
                not(containsString('neo/account'))))

        nullScript.commonPipelineEnvironment.configuration = [:]

        stepRule.step.neoDeploy(script: nullScript, deployMode: 'warPropertiesFile')
    }

    @Test
    public void sanityChecksDeployModeWarParamsTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(
            allOf(
                containsString('ERROR - NO VALUE AVAILABLE FOR:'),
                containsString('source'),
                containsString('neo/application'),
                containsString('neo/runtime'),
                containsString('neo/runtimeVersion'),
                containsString('neo/host'),
                containsString('neo/account')))

        nullScript.commonPipelineEnvironment.configuration = [:]

        stepRule.step.neoDeploy(script: nullScript, deployMode: 'warParams')
    }

    @Test
    void mtaDeployModeTest() {

        stepRule.step.neoDeploy(script: nullScript, source: archiveName, deployMode: 'mta')

        Assert.assertThat(shellRule.shell,
            new CommandLineMatcher().hasProlog("neo.sh deploy-mta")
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
            new CommandLineMatcher().hasProlog("neo.sh deploy")
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
            new CommandLineMatcher().hasProlog("neo.sh rolling-update")
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
                .hasProlog("neo.sh deploy")
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
            new CommandLineMatcher().hasProlog("neo.sh deploy")
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
            new CommandLineMatcher().hasProlog("neo.sh rolling-update")
                .hasArgument('config.properties')
                .hasSingleQuotedOption('user', 'defaultUser')
                .hasSingleQuotedOption('password', '\\*\\*\\*\\*\\*\\*\\*\\*')
                .hasSingleQuotedOption('source', '.*\\.war'))
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
}
