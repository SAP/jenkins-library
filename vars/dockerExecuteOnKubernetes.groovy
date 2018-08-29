import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
import com.sap.piper.k8s.SystemEnv
import groovy.transform.Field
import hudson.AbortException

@Field def STEP_NAME = 'dockerExecuteOnKubernetes'
@Field def PLUGIN_ID_KUBERNETES = 'kubernetes'
@Field Set GENERAL_CONFIG_KEYS = ['jenkinsKubernetes']
@Field Set PARAMETER_KEYS = ['dockerImage',
                             'dockerWorkspace',
                             'dockerEnvVars',
                             'containerMap']
@Field Set STEP_CONFIG_KEYS = PARAMETER_KEYS.plus(['stashIncludes', 'stashExcludes'])

void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        if (!JenkinsUtils.isPluginActive(PLUGIN_ID_KUBERNETES)) {
            error("[ERROR][${STEP_NAME}] not supported. Plugin '${PLUGIN_ID_KUBERNETES}' is not installed or not active.")
        }
        final script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]

        ConfigurationHelper configHelper = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('uniqueId', UUID.randomUUID().toString())
        Map config = [:]

        if (parameters.containerMap) {
            config = configHelper.use()
            executeOnPodWithCustomContainerList(config: config) { body() }

        } else {
            config = configHelper
                .withMandatoryProperty('dockerImage')
                .use()
            executeOnPodWithSingleContainer(config: config) { body() }
        }
    }
}

void executeOnPodWithCustomContainerList(Map parameters, body) {
    def config = parameters.config
    podTemplate(getOptions(config)) {
        node(config.uniqueId) {
            body()
        }
    }
}

def getOptions(config) {
    return [name      : 'dynamic-agent-' + config.uniqueId,
            label     : config.uniqueId,
            containers: getContainerList(config)]
}

void executeOnPodWithSingleContainer(Map parameters, body) {
    Map containerMap = [:]
    def config = parameters.config
    containerMap[config.get('dockerImage').toString()] = 'container-exec'
    config.containerMap = containerMap
    /*
     * There could be exceptions thrown by
        - The podTemplate
        - The container method
        - The body
     * We use nested exception handling in this case.
     * In the first 2 cases, the workspace has not been modified. Hence, we can stash existing workspace as container and
     * unstash in the finally block. In case of exception thrown by the body, we need to stash the workspace from the container
     * in finally block
     */
    try {
        stashWorkspace(config, 'workspace')
        podTemplate(getOptions(config)) {
            node(config.uniqueId) {
                container(name: 'container-exec') {
                    try {
                        unstashWorkspace(config, 'workspace')
                        body()
                    } finally {
                        stashWorkspace(config, 'container')
                     }
                }
            }
        }
    } catch (e) {
        stashWorkspace(config, 'container')
        throw e
    } finally {
        unstashWorkspace(config, 'container')
    }
}

private void stashWorkspace(config, prefix) {
    try {
        // Every dockerImage used in the dockerExecuteOnKubernetes should have user id 1000
        sh "chown -R 1000:1000 ."
        stash(
            name: "${prefix}-${config.uniqueId}",
            include: config.stashIncludes.workspace,
            exclude: config.stashExcludes.excludes
        )
    } catch (AbortException | IOException e) {
        echo "${e.getMessage()}"
    }
}

private void unstashWorkspace(config, prefix) {
    try {
        unstash "${prefix}-${config.uniqueId}"
    } catch (AbortException | IOException e) {
        echo "${e.getMessage()}"
    }
}

private List getContainerList(config) {
    def envVars = getContainerEnvs(config)
    result = []
    result.push(containerTemplate(name: 'jnlp',
        image: config.jenkinsKubernetes.jnlpAgent))
    config.containerMap.each { imageName, containerName ->
        result.push(containerTemplate(name: containerName.toLowerCase(),
            image: imageName,
            alwaysPullImage: true,
            command: '/usr/bin/tail -f /dev/null',
            envVars: envVars))
    }
    return result
}

/**
 * Returns a list of envVar object consisting of set
 * environment variables, params (Parametrized Build) and working directory.
 * (Kubernetes-Plugin only!)
 * @param config Map with configurations
 */
private List getContainerEnvs(config) {
    def containerEnv = []
    def dockerEnvVars = config.dockerEnvVars ?: [:]
    def dockerWorkspace = config.dockerWorkspace ?: ''

    if (dockerEnvVars) {
        for (String k : dockerEnvVars.keySet()) {
            containerEnv << envVar(key: k, value: dockerEnvVars[k].toString())
        }
    }

    if (dockerWorkspace) {
        containerEnv << envVar(key: "HOME", value: dockerWorkspace)
    }

    // Inherit the proxy information from the master to the container
    SystemEnv systemEnv = new SystemEnv()
    for (String env : systemEnv.getEnv().keySet()) {
        containerEnv << envVar(key: env, value: systemEnv.get(env))
    }

    // ContainerEnv array can't be empty. Using a stub to avoid failure.
    if (!containerEnv) {
        containerEnv << envVar(key: "EMPTY_VAR", value: " ")
    }

    return containerEnv
}
