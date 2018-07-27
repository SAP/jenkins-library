import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.ConfigurationLoader

def call(Map parameters = [:], body) {

    def STEP_NAME = 'dockerExecute'
    def PLUGIN_ID_DOCKER_WORKFLOW = 'docker-workflow'

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        final script = parameters.script
        ConfigurationHelper parameters = new ConfigurationHelper(parameters)
        ConfigurationLoader generalConfig = ConfigurationLoader.generalConfiguration(script)

        Set parameterKeys = ['dockerImage',
                             'dockerOptions',
                             'dockerEnvVars',
                             'dockerVolumeBind']

        Set generalConfigKeys = ['kubernetes']

        Map config = ConfigurationMerger.merge(parameters, parameterKeys, generalConfig, generalConfigKeys)

        if (isKubernetes(config)) {
            if (env.POD_NAME && isContainerDefined(config, dockerImage)) {
                container(getContainerDefined(config, dockerImage)) {
                    echo "Executing inside a Kubernetes Container"
                    body()
                    sh "chown -R 1000:1000 ."
                }
            } else {
                dockerExecuteOnKubernetes(script: script,
                    dockerImage: config.dockerImage,
                    dockerEnvVars: config.dockerEnvVars,
                    dockerOptions: config.dockerOptions,
                    dockerVolumeBind: config.dockerVolumeBind) {
                    body()
                }
            }
        } else if (config.dockerImage) {

            if (!isPluginActive(PLUGIN_ID_DOCKER_WORKFLOW)) {
                echo "[WARNING][${STEP_NAME}] Docker not supported. Plugin '${PLUGIN_ID_DOCKER_WORKFLOW}' is not installed or not active. Configured docker image '${config.dockerImage}' will not be used."
                config.dockerImage = null
            }

            def returnCode = sh script: 'which docker > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][${STEP_NAME}] No docker environment found (command 'which docker' did not return with '0'). Configured docker image '${config.dockerImage}' will not be used."
                config.dockerImage = null
            }

            returnCode = sh script: 'docker ps -q > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][$STEP_NAME] Cannot connect to docker daemon (command 'docker ps' did not return with '0'). Configured docker image '${config.dockerImage}' will not be used."
                config.dockerImage = null
            }
            def image = docker.image(config.dockerImage)
            image.pull()
            image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                body()
            }
        } else {
            echo "[INFO][${STEP_NAME}] Running on local environment."
            body()
        }
    }
}

@NonCPS
private isPluginActive(String pluginId) {
    return Jenkins.instance.pluginManager.plugins.find { p -> p.isActive() && p.getShortName() == pluginId }
}

/**
 * Returns a string with docker options containing
 * environment variables (if set).
 * Possible to extend with further options.
 * @param dockerEnvVars Map with environment variables
 */
@NonCPS
private getDockerOptions(Map dockerEnvVars, Map dockerVolumeBind, def dockerOptions) {
    def specialEnvironments = ['http_proxy',
                               'https_proxy',
                               'no_proxy',
                               'HTTP_PROXY',
                               'HTTPS_PROXY',
                               'NO_PROXY']
    def options = []
    if (dockerEnvVars) {
        for (String k : dockerEnvVars.keySet()) {
            options.add("--env ${k}=${dockerEnvVars[k].toString()}")
        }
    }

    for (String envVar : specialEnvironments) {
        if (dockerEnvVars == null || !dockerEnvVars.containsKey(envVar)) {
            options.add("--env ${envVar}")
        }
    }

    if (dockerVolumeBind) {
        for (String k : dockerVolumeBind.keySet()) {
            options.add("--volume ${k}:${dockerVolumeBind[k].toString()}")
        }
    }

    if (dockerOptions instanceof CharSequence) {
        options.add(dockerOptions.toString())
    } else if (dockerOptions instanceof List) {
        for (String option : dockerOptions) {
            options.add "${option}"
        }
    } else {
        throw new IllegalArgumentException("Unexpected type for dockerOptions. Expected was either a list or a string. Actual type was: '${dockerOptions.getClass()}'")
    }

    return options.join(' ')
}

@NonCPS
boolean isContainerDefined(config, dockerImage) {
    def k8sMapping = config.k8sMapping ?: [:]
    if (k8sMapping.containsKey(env.POD_NAME)) {
        return k8sMapping[env.POD_NAME].containsKey(dockerImage)
    }
    return false
}

@NonCPS
def getContainerDefined(config, dockerImage) {
    def k8sMapping = config.k8sMapping
    return k8sMapping[env.POD_NAME]?.get(dockerImage)
}

@NonCPS
boolean isKubernetes(config) {
    return config.kubernetes?.enabled ?: false
}
