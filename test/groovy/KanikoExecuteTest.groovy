#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class KanikoExecuteTest extends BasePiperTest {
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, 'test/resources/kaniko/')
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(shellRule)
        .around(readFileRule)
        .around(writeFileRule)
        .around(dockerExecuteRule)
        .around(stepRule)

    def fileMap = [:]

    @Before
    void init() {
        binding.variables.env.WORKSPACE = '/path/to/current/workspace'

        helper.registerAllowedMethod('file', [Map], { m ->
            fileMap = m
            return m
        })

        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            binding.setProperty(fileMap.variable, 'config.json')
            try {
                c()
            } finally {
                binding.setProperty(fileMap.variable, null)
            }
        })
    }

    @Test
    void testDefaults() {
        stepRule.step.kanikoExecute(
            script: nullScript
        )
        assertThat(shellRule.shell, hasItem('#!/busybox/sh rm /kaniko/.docker/config.json'))
        assertThat(shellRule.shell, hasItem('#!/busybox/sh /kaniko/executor --dockerfile /path/to/current/workspace/Dockerfile --context /path/to/current/workspace --skip-tls-verify --skip-tls-verify-pull --no-push'))

        assertThat(writeFileRule.files, hasEntry('/kaniko/.docker/config.json', '{"auths":{}}'))

        assertThat(dockerExecuteRule.dockerParams, allOf(
            hasEntry('containerCommand', '/busybox/tail -f /dev/null'),
            hasEntry('containerShell', '/busybox/sh'),
            hasEntry('dockerImage', 'gcr.io/kaniko-project/executor:debug'),
            hasEntry('dockerOptions', "-u 0 --entrypoint=''")

        ))
    }

    @Test
    void testCustomDockerCredentials() {
        stepRule.step.kanikoExecute(
            script: nullScript,
            dockerConfigJsonCredentialsId: 'myDockerConfigJson'
        )

        assertThat(fileMap.credentialsId, is('myDockerConfigJson'))
        assertThat(writeFileRule.files.get('/kaniko/.docker/config.json'), allOf(
            containsString('docker.my.domain.com:4444'),
            containsString('"auth": "myAuth"'),
            containsString('"email": "my.user@domain.com"')
        ))
    }

    @Test
    void testCustomImage() {
        stepRule.step.kanikoExecute(
            script: nullScript,
            containerImageNameAndTag: 'my.docker.registry/path/myImageName:myTag'
        )

        assertThat(shellRule.shell, hasItem('#!/busybox/sh /kaniko/executor --dockerfile /path/to/current/workspace/Dockerfile --context /path/to/current/workspace --skip-tls-verify --skip-tls-verify-pull --destination my.docker.registry/path/myImageName:myTag'))
    }

    @Test
    void testPreserveDestination() {
        stepRule.step.kanikoExecute(
            script: nullScript,
            containerBuildOptions: '--destination my.docker.registry/path/myImageName:myTag'
        )

        assertThat(shellRule.shell, hasItem('#!/busybox/sh /kaniko/executor --dockerfile /path/to/current/workspace/Dockerfile --context /path/to/current/workspace --destination my.docker.registry/path/myImageName:myTag'))
    }
}
