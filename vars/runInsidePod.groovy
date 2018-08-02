import com.sap.piper.ConfigurationHelper
import com.sap.piper.SysEnv
import groovy.transform.Field

@Field def STEP_NAME = 'runInsidePod'

@Field Set GENERAL_CONFIG_KEYS = ['jenkinsKubernetes']

@Field Set PARAMETER_KEYS = ['dockerOptions',
                             'dockerWorkspace',
                             'dockerEnvVars',
                             'containersMap']

def call(Map parameters = [:], body) {
    def uniqueId = UUID.randomUUID().toString()

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = parameters.script

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('containersMap')
            .use()

        def options = [name      : 'dynamic-agent-' + uniqueId,
                       label     : uniqueId,
                       containers: getContainerList(config)]
        podTemplate(options) {
            node(uniqueId) {
                body()
            }
        }
    }
}

private getContainerList(config) {
    def envVars
    envVars = getContainerEnvs(config)
    result = []
    result.push(containerTemplate(name: 'jnlp',
        image: config.jenkinsKubernetes.jnlpAgent,
        args: '${computer.jnlpmac} ${computer.name}'))

    config.containersMap.each { imageName, containerName  ->
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
private getContainerEnvs(config) {
    def containerEnv = []
    def dockerEnvVars = config.dockerEnvVars ?: [:]
    def dockerWorkspace = config.dockerWorkspace ?: ''
    if (dockerEnvVars) {
        for (String k : dockerEnvVars.keySet()) {
            containerEnv << envVar(key: k, value: dockerEnvVars[k].toString())
        }
    }

    if (dockerWorkspace) containerEnv << envVar(key: "HOME", value: dockerWorkspace)

    // Inherit the proxy information from the master to the container
    def systemEnv = new SysEnv()
    def envList = systemEnv.getEnv().keySet()
    for (String env : envList) {
        containerEnv << envVar(key: env, value: systemEnv.get(env))
    }

    // ContainerEnv array can't be empty. Using a stub to avoid failure.
    if (!containerEnv) containerEnv << envVar(key: "EMPTY_VAR", value: " ")

    return containerEnv
}
