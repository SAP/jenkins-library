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

        UUID.metaClass.static.randomUUID = { -> 1}
    }

    @Test
    void testDefaults() {
        stepRule.step.kanikoExecute(
            script: nullScript
        )
        assertThat(shellRule.shell, hasItem('#!/busybox/sh rm -f /kaniko/.docker/config.json'))
        assertThat(shellRule.shell, hasItem(allOf(
            startsWith('#!/busybox/sh'),
            containsString('mv 1-config.json /kaniko/.docker/config.json'),
            containsString('/kaniko/executor'),
            containsString('--dockerfile /path/to/current/workspace/Dockerfile'),
            containsString('--context /path/to/current/workspace'),
            containsString('--skip-tls-verify-pull'),
            containsString('--no-push')
        )))

        assertThat(writeFileRule.files.values()[0], is('{"auths":{}}'))

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
        assertThat(writeFileRule.files.values()[0], allOf(
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

        assertThat(shellRule.shell, hasItem(allOf(
            startsWith('#!/busybox/sh'),
            containsString('mv 1-config.json /kaniko/.docker/config.json'),
            containsString('/kaniko/executor'),
            containsString('--dockerfile /path/to/current/workspace/Dockerfile'),
            containsString('--context /path/to/current/workspace'),
            containsString('--skip-tls-verify-pull'),
            containsString('--destination my.docker.registry/path/myImageName:myTag')
        )))
    }

    @Test
    void testPreserveDestination() {
        stepRule.step.kanikoExecute(
            script: nullScript,
            containerBuildOptions: '--destination my.docker.registry/path/myImageName:myTag'
        )

        assertThat(shellRule.shell, hasItem(allOf(
            startsWith('#!/busybox/sh'),
            containsString('mv 1-config.json /kaniko/.docker/config.json'),
            containsString('/kaniko/executor'),
            containsString('--dockerfile /path/to/current/workspace/Dockerfile'),
            containsString('--context /path/to/current/workspace'),
            containsString('--destination my.docker.registry/path/myImageName:myTag')
        )))
    }

    @Test
    void testCustomCertificates() {
        stepRule.step.kanikoExecute(
            script: nullScript,
            customTlsCertificateLinks: ['http://link.one', 'http://link.two']
        )

        assertThat(shellRule.shell, hasItem(allOf(
            startsWith('#!/busybox/sh'),
            containsString('rm -f /kaniko/.docker/config.json'),
            containsString('wget http://link.one -O - >> /kaniko/ssl/certs/ca-certificates.crt'),
            containsString('wget http://link.two -O - >> /kaniko/ssl/certs/ca-certificates.crt'),
        )))
    }
}
