import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsDockerExecuteRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.CoreMatchers.containsString
import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

class DetectExecuteScanTest extends BasePiperTest {

    private JenkinsDockerExecuteRule dockerRule = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    private String detectProperties = ''

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(shellRule)
        .around(dockerRule)
        .around(stepRule)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('testCredentials', 'testToken')
        )

    @Before
    void init() {

        detectProperties = ''
        helper.registerAllowedMethod('synopsys_detect', [String.class], {s ->
            detectProperties = s
        })
    }

    @Test
    void testDetectDefault() {
        stepRule.step.detectExecuteScan([
            apiTokenCredentialsId: 'testCredentials',
            projectName: 'testProject',
            serverUrl: 'https://test.blackducksoftware.com',
            juStabUtils: utils,
            script: nullScript,
            groups: 'testGroup'
        ])

        //ToDo: assert unstashing

        assertThat(detectProperties, containsString("--detect.project.name='testProject'"))
        assertThat(detectProperties, containsString("--detect.project.version.name='1'"))
        assertThat(detectProperties, containsString("--blackduck.url=https://test.blackducksoftware.com"))
        assertThat(detectProperties, containsString("--blackduck.api.token=testToken"))
        assertThat(detectProperties, containsString("--detect.blackduck.signature.scanner.paths=."))
        assertThat(detectProperties, containsString("--blackduck.signature.scanner.memory=4096"))
        assertThat(detectProperties, containsString("--blackduck.timeout=6000"))
        assertThat(detectProperties, containsString("--blackduck.trust.cert=true"))
        assertThat(detectProperties, containsString("--detect.report.timeout=4800"))
        assertThat(detectProperties, containsString("--detect.project.user.groups='testGroup'"))
    }

    @Test
    void testDetectCustomPaths() {
        stepRule.step.detectExecuteScan([
            apiTokenCredentialsId: 'testCredentials',
            projectName: 'testProject',
            scanPaths: ['test1/', 'test2/'],
            serverUrl: 'https://test.blackducksoftware.com',
            juStabUtils: utils,
            script: nullScript
        ])

        assertThat(detectProperties, containsString("--detect.blackduck.signature.scanner.paths=test1/,test2/"))
    }

    @Test
    void testDetectSourceScanOnly() {
        stepRule.step.detectExecuteScan([
            apiTokenCredentialsId: 'testCredentials',
            projectName: 'testProject',
            scanners: ['source'],
            serverUrl: 'https://test.blackducksoftware.com',
            juStabUtils: utils,
            script: nullScript
        ])

        assertThat(detectProperties, not(containsString("--detect.blackduck.signature.scanner.paths=.")))
        assertThat(detectProperties, containsString("--detect.source.path=."))
    }

    @Test
    void testDetectGolang() {
        stepRule.step.detectExecuteScan([
            buildTool: 'golang',
            apiTokenCredentialsId: 'testCredentials',
            projectName: 'testProject',
            serverUrl: 'https://test.blackducksoftware.com',
            juStabUtils: utils,
            script: nullScript
        ])

        assertThat(dockerRule.dockerParams.dockerImage, is('golang:1.12-stretch'))
        assertThat(dockerRule.dockerParams.dockerWorkspace, is(''))
        assertThat(dockerRule.dockerParams.stashContent, allOf(hasItem('buildDescriptor'),hasItem('checkmarx')))

        assertThat(shellRule.shell, hasItem('curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh'))
        assertThat(shellRule.shell, hasItem('ln --symbolic $(pwd) $GOPATH/src/hub'))
        assertThat(shellRule.shell, hasItem('cd $GOPATH/src/hub && dep ensure'))
    }

    @Test
    void testCustomScanProperties() {
        def detectProps = [
            '--blackduck.signature.scanner.memory=1024'
        ]
        stepRule.step.detectExecuteScan([
            //scanProperties: detectProps,
            scanProperties: ['--blackduck.signature.scanner.memory=1024', '--myNewOne'],
            apiTokenCredentialsId: 'testCredentials',
            projectName: 'testProject',
            serverUrl: 'https://test.blackducksoftware.com',
            juStabUtils: utils,
            script: nullScript
        ])

        assertThat(detectProperties, containsString("--detect.project.name='testProject'"))
        assertThat(detectProperties, containsString("--detect.project.version.name='1'"))
        assertThat(detectProperties, containsString("--blackduck.signature.scanner.memory=1024"))
        assertThat(detectProperties, not(containsString("--blackduck.signature.scanner.memory=4096")))
        assertThat(detectProperties, not(containsString("--detect.report.timeout=4800")))
        assertThat(detectProperties, containsString("--myNewOne"))
    }
}
