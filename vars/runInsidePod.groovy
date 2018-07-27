import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.SysEnv

import java.util.UUID

def call(Map parameters = [:], body) {
    def uniqueId = UUID.randomUUID().toString()

    ConfigurationHelper configurationHelper = new ConfigurationHelper(parameters)
    def script = configurationHelper.getConfigProperty('script')
    def containersMap = configurationHelper.getConfigProperty('containersMap',[:])
    def dockerEnvVars = configurationHelper.getConfigProperty('dockerEnvVars',[:])
    def dockerWorkspace = configurationHelper.getConfigProperty('dockerWorkspace','')

    handleStepErrors(stepName: 'runInsidePod', stepParameters: [:]) {
        def options = [name      : 'dynamic-agent-' + uniqueId,
                       label     : uniqueId,
                       containers: getContainerList(script, containersMap, dockerEnvVars, dockerWorkspace)]
        podTemplate(options) {
            node(uniqueId) {
                body()
            }
        }
    }
}

private getContainerList(script, containersMap, dockerEnvVars, dockerWorkspace) {
    def envVars
    def jnlpAgent = ConfigurationLoader.generalConfiguration(script).kubernetes.jnlpAgent

    envVars = getContainerEnvs(dockerEnvVars, dockerWorkspace)
    result = []
    result.push(containerTemplate(name: 'jnlp',
        image: jnlpAgent,
        args: '${computer.jnlpmac} ${computer.name}'))

    containersMap.each { k, v ->
        result.push(containerTemplate(name: v,
            image: k,
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
 * @param dockerEnvVars Map with environment variables
 * @param dockerWorkspace Path to working dir
 */
private getContainerEnvs(dockerEnvVars, dockerWorkspace) {
    def containerEnv = []

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
