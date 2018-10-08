import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.ConfigurationHelper
import com.sap.piper.k8s.ContainerMap
import com.sap.piper.JenkinsUtils
import groovy.transform.Field

@Field def STEP_NAME = 'dockerExecute'
@Field def PLUGIN_ID_DOCKER_WORKFLOW = 'docker-workflow'

@Field Set GENERAL_CONFIG_KEYS = ['jenkinsKubernetes']

@Field Set PARAMETER_KEYS = [
    'containerPortMappings',
    'dockerEnvVars',
    'dockerImage',
    'dockerName',
    'dockerOptions',
    'dockerWorkspace',
    'dockerVolumeBind',
    'sidecarName',
    'sidecarEnvVars',
    'sidecarImage',
    'sidecarOptions',
    'sidecarWorkspace',
    'sidecarVolumeBind'
]
@Field Set STEP_CONFIG_KEYS = PARAMETER_KEYS

void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        final script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()
        if (isKubernetes() && config.dockerImage) {
            if (env.POD_NAME && isContainerDefined(config)) {
                container(getContainerDefined(config)) {
                    echo "[INFO][${STEP_NAME}] Executing inside a Kubernetes Container."
                    body()
                    sh "chown -R 1000:1000 ."
                }
            } else {
                if (!config.sidecarImage) {
                    dockerExecuteOnKubernetes(
                        script: script,
                        dockerImage: config.dockerImage,
                        dockerEnvVars: config.dockerEnvVars,
                        dockerWorkspace: config.dockerWorkspace
                    ){
                        echo "[INFO][${STEP_NAME}] Executing inside a Kubernetes Pod"
                        body()
                    }
                } else {
                    Map paramMap = [
                        script: script,
                        containerCommands: [:],
                        containerEnvVars: [:],
                        containerMap: [:],
                        containerName: config.dockerName,
                        containerPortMappings: [:],
                        containerWorkspaces: [:]
                    ]
                    paramMap.containerCommands[config.sidecarImage] = ''

                    paramMap.containerEnvVars[config.dockerImage] = config.dockerEnvVars
                    paramMap.containerEnvVars[config.sidecarImage] = config.sidecarEnvVars

                    paramMap.containerMap[config.dockerImage] = config.dockerName
                    paramMap.containerMap[config.sidecarImage] = config.sidecarName

                    paramMap.containerPortMappings = config.containerPortMappings

                    paramMap.containerWorkspaces[config.dockerImage] = config.dockerWorkspace
                    paramMap.containerWorkspaces[config.sidecarImage] = ''

                    dockerExecuteOnKubernetes(paramMap){
                        echo "[INFO][${STEP_NAME}] Executing inside a Kubernetes Pod with sidecar container"
                        body()
                    }
                }
            }
        } else {
            boolean executeInsideDocker = true
            if (!JenkinsUtils.isPluginActive(PLUGIN_ID_DOCKER_WORKFLOW)) {
                echo "[WARNING][${STEP_NAME}] Docker not supported. Plugin '${PLUGIN_ID_DOCKER_WORKFLOW}' is not installed or not active. Configured docker image '${config.dockerImage}' will not be used."
                executeInsideDocker = false
            }

            def returnCode = sh script: 'which docker > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][${STEP_NAME}] No docker environment found (command 'which docker' did not return with '0'). Configured docker image '${config.dockerImage}' will not be used."
                executeInsideDocker = false
            }

            returnCode = sh script: 'docker ps -q > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][$STEP_NAME] Cannot connect to docker daemon (command 'docker ps' did not return with '0'). Configured docker image '${config.dockerImage}' will not be used."
                executeInsideDocker = false
            }
            if (executeInsideDocker && config.dockerImage) {
                def image = docker.image(config.dockerImage)
                image.pull()
                if (!config.sidecarImage) {
                    image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                        body()
                    }
                } else {
                    def sidecarImage = docker.image(config.sidecarImage)
                    sidecarImage.pull()
                    sidecarImage.withRun(getDockerOptions(config.sidecarEnvVars, config.sidecarVolumeBind, config.sidecarOptions)) { c ->
                        config.dockerOptions = config.dockerOptions?:[]
                        config.dockerOptions.add("--link ${c.id}:${config.sidecarName}")
                        image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                            echo "[INFO][${STEP_NAME}] Running with sidecar container."
                            body()
                        }
                    }
                }
            } else {
                echo "[INFO][${STEP_NAME}] Running on local environment."
                body()
            }
        }
    }
}



/**
 * Returns a string with docker options containing
 * environment variables (if set).
 * Possible to extend with further options.
 * @param dockerEnvVars Map with environment variables
 */
@NonCPS
private getDockerOptions(Map dockerEnvVars, Map dockerVolumeBind, def dockerOptions) {
    def specialEnvironments = [
        'http_proxy',
        'https_proxy',
        'no_proxy',
        'HTTP_PROXY',
        'HTTPS_PROXY',
        'NO_PROXY'
    ]
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

    if (dockerOptions) {
        if (dockerOptions instanceof CharSequence) {
            options.add(dockerOptions.toString())
        } else if (dockerOptions instanceof List) {
            for (String option : dockerOptions) {
                options.add "${option}"
            }
        } else {
            throw new IllegalArgumentException("Unexpected type for dockerOptions. Expected was either a list or a string. Actual type was: '${dockerOptions.getClass()}'")
        }
    }
    return options.join(' ')
}


boolean isContainerDefined(config) {
    Map containerMap = ContainerMap.instance.getMap()

    if (!containerMap.containsKey(env.POD_NAME)) {
        return false
    }

    return containerMap.get(env.POD_NAME).containsKey(config.dockerImage)
}


def getContainerDefined(config) {
    return ContainerMap.instance.getMap().get(env.POD_NAME).get(config.dockerImage).toLowerCase()
}


boolean isKubernetes() {
    return Boolean.valueOf(env.ON_K8S)
}
