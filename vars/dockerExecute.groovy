import com.cloudbees.groovy.cps.NonCPS

def call(Map parameters = [:], body) {

    def STEP_NAME = 'dockerExecute'
    def PLUGIN_ID_DOCKER_WORKFLOW = 'docker-workflow'

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def dockerImage = parameters.dockerImage ?: ''
        Map dockerEnvVars = parameters.dockerEnvVars ?: [:]
        def dockerOptions = parameters.dockerOptions ?: ''
        Map dockerVolumeBind = parameters.dockerVolumeBind ?: [:]
        final script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]
        echo "Has container defined for ${dockerImage} and it is ${hasContainerDefined(script, dockerImage)} value is ${getContainerDefined(script, dockerImage)}"

        if (env.S4SDK_STAGE_NAME && hasContainerDefined(script, dockerImage)) {
            container(getContainerDefined(script, dockerImage)) {
                echo "Executing inside a Kubernetes Container"
                body()
            }
        } else if (env.jaas_owner) {
            executeDockerOnKubernetes(
                dockerImage: parameters.dockerImage,
                dockerEnvVars: parameters.dockerEnvVars,
                dockerOptions: parameters.dockerOptions,
                dockerVolumeBind: parameters.dockerVolumeBind) {
                body()
            }
        } else if (dockerImage) {

            if (!isPluginActive(PLUGIN_ID_DOCKER_WORKFLOW)) {
                echo "[WARNING][${STEP_NAME}] Docker not supported. Plugin '${PLUGIN_ID_DOCKER_WORKFLOW}' is not installed or not active. Configured docker image '${dockerImage}' will not be used."
                dockerImage = null
            }

            def returnCode = sh script: 'which docker > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][${STEP_NAME}] No docker environment found (command 'which docker' did not return with '0'). Configured docker image '${dockerImage}' will not be used."
                dockerImage = null
            }

            returnCode = sh script: 'docker ps -q > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][$STEP_NAME] Cannot connect to docker daemon (command 'docker ps' did not return with '0'). Configured docker image '${dockerImage}' will not be used."
                dockerImage = null
            }
            def image = docker.image(dockerImage)
            image.pull()
            image.inside(getDockerOptions(dockerEnvVars, dockerVolumeBind, dockerOptions)) {
                body()
            }
        }

        if (!dockerImage) {
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
boolean hasContainerDefined(script, dockerImage) {
    echo "Values for${dockerImage} is ${script?.commonPipelineEnvironment?.configuration?.k8sMapping} and stage name is ${env.S4SDK_STAGE_NAME}"
    return script?.commonPipelineEnvironment?.configuration?.k8sMapping[env.S4SDK_STAGE_NAME].containsKey(dockerImage)
}

@NonCPS
def getContainerDefined(script, dockerImage) {
    return script?.commonPipelineEnvironment?.configuration?.k8sMapping[env.S4SDK_STAGE_NAME][dockerImage]
}
