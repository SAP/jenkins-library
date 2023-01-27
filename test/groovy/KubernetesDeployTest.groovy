import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class KubernetesDeployTest extends BasePiperTest {

    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)

    private List withEnvArgs = []
    private List credentials = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(readJsonRule)
        .around(shellCallRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(dockerExecuteRule)

    @Before
    void init() {
        credentials = []
        helper.registerAllowedMethod("withEnv", [List.class, Closure.class], {arguments, closure ->
            arguments.each {arg ->
                withEnvArgs.add(arg.toString())
            }
            return closure()
        })

        helper.registerAllowedMethod("writePipelineEnv", [Map.class], {m -> return })
        helper.registerAllowedMethod("readPipelineEnv", [Map.class], {m -> return })
        helper.registerAllowedMethod('fileExists', [Map.class], {
            return false
        })

        helper.registerAllowedMethod('file', [Map], { m -> return m })
        helper.registerAllowedMethod('string', [Map], { m -> return m })
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            l.each {m ->
                credentials.add(m)
                if (m.credentialsId == 'kubeConfig') {
                    binding.setProperty('PIPER_kubeConfig', 'myKubeConfig')
                } else if (m.credentialsId == 'kubeToken') {
                    binding.setProperty('PIPER_kubeToken','myKubeToken')
                } else if (m.credentialsId == 'dockerCredentials') {
                    binding.setProperty('PIPER_containerRegistryUser', 'registryUser')
                    binding.setProperty('PIPER_containerRegistryPassword', '********')
                }
            }
            try {
                c()
            } finally {
                binding.setProperty('PIPER_kubeConfig', null)
                binding.setProperty('PIPER_kubeToken', null)
                binding.setProperty('PIPER_containerRegistryUser', null)
                binding.setProperty('PIPER_containerRegistryPassword', null)
            }
        })

        helper.registerAllowedMethod('fileExists', [String.class], {return true})
        helper.registerAllowedMethod('findFiles', [Map.class], {return null})
    }

    @Test
    void testKubernetesDeployAllCreds() {
        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/kubernetesDeploy.yaml\'', '{"kubeConfigFileCredentialsId":"kubeConfig", "kubeTokenCredentialsId":"kubeToken", "dockerCredentialsId":"dockerCredentials", "dockerImage":"my.Registry/K8S:latest"}')

        stepRule.step.kubernetesDeploy(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            testParam: "This is test content",
            script: nullScript
        )
        // asserts
        assertThat(writeFileRule.files['.pipeline/tmp/metadata/kubernetesDeploy.yaml'], containsString('name: kubernetesDeploy'))
        assertThat(withEnvArgs[0], allOf(startsWith('PIPER_parametersJSON'), containsString('"testParam":"This is test content"')))
        assertThat(shellCallRule.shell[2], is('./piper kubernetesDeploy'))
        assertThat(credentials.size(), is(3))

        assertThat(dockerExecuteRule.dockerParams.dockerImage, is('my.Registry/K8S:latest'))
    }

    @Test
    void testKubernetesDeploySomeCreds() {
        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/kubernetesDeploy.yaml\'', '{"kubeTokenCredentialsId":"kubeToken", "dockerCredentialsId":"dockerCredentials"}')
        stepRule.step.kubernetesDeploy(
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript
        )
        // asserts
        assertThat(shellCallRule.shell[2], is('./piper kubernetesDeploy'))
        assertThat(credentials.size(), is(2))
    }
}
