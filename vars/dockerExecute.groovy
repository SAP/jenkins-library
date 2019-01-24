import static com.sap.piper.Prerequisites.checkScript

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field def PLUGIN_ID_DOCKER_WORKFLOW = 'docker-workflow'
@Field Set GENERAL_CONFIG_KEYS = ['jenkinsKubernetes']
@Field Set STEP_CONFIG_KEYS = [
    'containerPortMappings',
    'containerCommand',
    'containerShell',
    'dockerEnvVars',
    'dockerImage',
    'dockerName',
    'dockerOptions',
    'dockerWorkspace',
    'dockerVolumeBind',
    'dockerAlwaysPullImage',
    'sidecarEnvVars',
    'sidecarImage',
    'sidecarName',
    'sidecarOptions',
    'sidecarWorkspace',
    'sidecarVolumeBind',
    'stashContent'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this

        def utils = parameters?.juStabUtils ?: new Utils()

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
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
                        containerCommand: config.containerCommand,
                        containerShell: config.containerShell,
                        dockerImage: config.dockerImage,
                        dockerEnvVars: config.dockerEnvVars,
                        dockerWorkspace: config.dockerWorkspace,
                        stashContent: config.stashContent
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
                        containerWorkspaces: [:],
                        stashContent: config.stashContent
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

            returnCode = sh script: 'docker ps -q > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][$STEP_NAME] Cannot connect to docker daemon (command 'docker ps' did not return with '0'). Configured docker image '${config.dockerImage}' will not be used."
                executeInsideDocker = false
            }
            if (executeInsideDocker && config.dockerImage) {
                utils.unstashAll(config.stashContent)
                def image = docker.image(config.dockerImage)
                if(config.dockerAlwaysPullImage) image.pull()
                if (!config.sidecarImage) {
                    image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                        body()
                    }
                } else {
                    def networkName = "sidecar-${UUID.randomUUID()}"
                    sh "docker network create ${networkName}"
                    try{
                        def sidecarImage = docker.image(config.sidecarImage)
                        if(config.dockerAlwaysPullImage) sidecarImage.pull()
                        config.sidecarOptions = config.sidecarOptions?:[]
                        if(config.sidecarName)
                            config.sidecarOptions.add("--network-alias ${config.sidecarName}")
                        config.sidecarOptions.add("--network ${networkName}")
                        sidecarImage.withRun(getDockerOptions(config.sidecarEnvVars, config.sidecarVolumeBind, config.sidecarOptions)) { c ->
                            config.dockerOptions = config.dockerOptions?:[]
                            if(config.dockerName)
                                config.dockerOptions.add("--network-alias ${config.dockerName}")
                            config.dockerOptions.add("--network ${networkName}")
                            image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                                echo "[INFO][${STEP_NAME}] Running with sidecar container."
                                body()
                            }
                        }
                    }finally{
                        sh "docker network remove ${networkName}"
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
            dockerOptions = [dockerOptions]
        }
        if (dockerOptions instanceof List) {
            for (String option : dockerOptions) {
                options << escapeBlanks(option)
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

/**
 * Escapes blanks for values in key/value pairs
 * E.g. <code>description=Lorem ipsum</code> is
 * changed to <code>description=Lorem\ ipsum</code>.
 */
@NonCPS
def escapeBlanks(def s) {

    def EQ='='
    def parts=s.split(EQ)

    if(parts.length == 2) {
        parts[1]=parts[1].replaceAll(' ', '\\\\ ')
        s = parts.join(EQ)
    }

    return s
}
