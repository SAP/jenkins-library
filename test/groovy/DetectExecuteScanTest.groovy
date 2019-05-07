#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
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

    @Before
    void init() {

        detectProperties = ''
        helper.registerAllowedMethod('string', [Map.class], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List.class, Closure.class], { l, c ->
            if (l[0].credentialsId == 'testCredentials') {
                binding.setProperty(l[0].variable, 'testToken')
            }
            try {
                c()
            } finally {
                binding.setProperty(l[0].variable, null)
            }
        })

        helper.registerAllowedMethod('synopsys_detect', [String.class], {s ->
            detectProperties = s
        })
    }

    @Test
    void testDetectDefault() {
        stepRule.step.detectExecuteScan([
            userTokenCredentialsId: 'testCredentials',
            projectName: 'testProject',
            serverUrl: 'https://test.blackducksoftware.com',
            juStabUtils: utils,
            script: nullScript
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
    }

    @Test
    void testDetectGolang() {
        stepRule.step.detectExecuteScan([
            buildTool: 'golang',
            userTokenCredentialsId: 'testCredentials',
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
    void testOverwriteScanProperty() {
        def detectProps = [
            '--blackduck.signature.scanner.memory=1024'
        ]
        stepRule.step.detectExecuteScan([
            scanProperties: detectProps,
            userTokenCredentialsId: 'testCredentials',
            projectName: 'testProject',
            serverUrl: 'https://test.blackducksoftware.com',
            juStabUtils: utils,
            script: nullScript
        ])

        assertThat(detectProperties, containsString("--blackduck.signature.scanner.memory=1024"))
        assertThat(detectProperties, not(containsString("--blackduck.signature.scanner.memory=4096")))
    }
}
