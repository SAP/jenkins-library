import com.sap.piper.SidecarUtils

import static com.sap.piper.Prerequisites.checkScript
import static com.sap.piper.BashUtils.quoteAndEscape as q

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field def PLUGIN_ID_DOCKER_WORKFLOW = 'docker-workflow'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Set this to 'false' to bypass a docker image pull. Useful during development process. Allows testing of images which are available in the local registry only.
     */
    'dockerPullImage',
    /**
     * Set this to 'false' to bypass a docker image pull. Useful during development process. Allows testing of images which are available in the local registry only.
     */
    'sidecarPullImage'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Kubernetes only:
     * Allows to specify start command for container created with dockerImage parameter to overwrite Piper default (`/usr/bin/tail -f /dev/null`).
     */
    'containerCommand',
    /**
     * Map which defines per docker image the port mappings, e.g. `containerPortMappings: ['selenium/standalone-chrome': [[name: 'selPort', containerPort: 4444, hostPort: 4444]]]`.
     */
    'containerPortMappings',
    /**
     * Kubernetes only:
     * Allows to specify the shell to be used for execution of commands.
     */
    'containerShell',
    /**
     * Kubernetes only: Allows to specify additional pod properties. For more details see step `dockerExecuteOnKubernetes`
     */
    'additionalPodProperties',
    /**
     * Environment variables to set in the container, e.g. [http_proxy: 'proxy:8080'].
     */
    'dockerEnvVars',
    /**
     * Name of the docker image that should be used.
     * Configure with empty value to execute the command directly on the Jenkins system (not using a container).
     * Omit to use the default image (cf. [default_pipeline_environment.yml](https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml))
     * Overwrite to use custom Docker image.
     */
    'dockerImage',
    /**
      * The registry used for pulling the docker image, if left empty the default registry as defined by the `docker-commons-plugin` will be used.
      */
    'dockerRegistryUrl',
    /**
      * Non Kubernetes only:
      * The credentials for the docker registry of type username/password as we rely on docker jenkins plugin. If left empty, images are pulled anonymously.
      * For Kubernetes cases, pass secret name of type `kubernetes.io/dockerconfigjson` via `additionalPodProperties` parameter (The secret should already be created and present in the environment)
      */
    'dockerRegistryCredentialsId',
    /**
      * Same as `dockerRegistryUrl`, but for the sidecar. If left empty, `dockerRegistryUrl` is used instead.
      */
    'sidecarRegistryUrl',
    /**
      * Same as `dockerRegistryCredentialsId`, but for the sidecar. If left empty `dockerRegistryCredentialsId` is used instead.
      */
    'sidecarRegistryCredentialsId',
    /**
     * Kubernetes only:
     * Name of the container launching `dockerImage`.
     * SideCar only:
     * Name of the container in local network.
     */
    'dockerName',
    /**
     * Docker only:
     * Docker options to be set when starting the container (List or String).
     */
    'dockerOptions',
    /**
     * Docker only:
     * Volumes that should be mounted into the container.
     */
    'dockerVolumeBind',
    /**
     * Kubernetes only:
     * Specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`.
     */
    'dockerWorkspace',
    /**
     * as `dockerEnvVars` for the sidecar container
     */
    'sidecarEnvVars',
    /**
     * as `dockerImage` for the sidecar container
     */
    'sidecarImage',
    /**
     * as `dockerName` for the sidecar container
     */
    'sidecarName',
    /**
     * as `dockerOptions` for the sidecar container
     */
    'sidecarOptions',
    /**
     * as `dockerVolumeBind` for the sidecar container
     */
    'sidecarVolumeBind',
    /**
     * as `dockerWorkspace` for the sidecar container
     */
    'sidecarWorkspace',
    /**
     * Command executed inside the container which returns exit code 0 when the container is ready to be used.
     */
    'sidecarReadyCommand',
    /**
     * Specific stashes that should be considered for the step execution.
     */
    'stashContent'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * In the Kubernetes case the workspace is only available to the respective Jenkins slave but not to the containers running inside the pod.<br />
     * This flag controls whether the stashing does *not* use the default exclude patterns in addition to the patterns provided in `stashExcludes`.
     * @possibleValues `true`, `false`
     */
    'stashNoDefaultExcludes',
])

@Field Map CONFIG_KEY_COMPATIBILITY = [
    dockerRegistryCredentialsId: 'dockerRegistryCredentials',
    sidecarRegistryCredentialsId: 'dockerSidecarRegistryCredentials',
]

/**
 * Executes a closure inside a docker container with the specified docker image.
 * The workspace is mounted into the docker image.
 * Proxy environment variables defined on the Jenkins machine are also available in the Docker container.
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        final script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .use()

        config = ConfigurationHelper.newInstance(this, config)
            .addIfEmpty('sidecarRegistryUrl', config.dockerRegistryUrl)
            .addIfEmpty('sidecarRegistryCredentialsId', config.dockerRegistryCredentialsId)
            .use()

        SidecarUtils sidecarUtils = new SidecarUtils(script)

        if (isKubernetes() && config.dockerImage) {
            List dockerEnvVars = []
            config.dockerEnvVars?.each { key, value ->
                dockerEnvVars << "$key=$value"
            }

            def securityContext = securityContextFromOptions(config.dockerOptions)
            def containerMountPath = containerMountPathFromVolumeBind(config.dockerVolumeBind)
            if (env.POD_NAME && isContainerDefined(config)) {
                container(getContainerDefined(config)) {
                    withEnv(dockerEnvVars) {
                        echo "[INFO][${STEP_NAME}] Executing inside a Kubernetes Container. Docker image: ${config.dockerImage}"
                        body()
                        sh "chown -R 1000:1000 ."
                    }
                }
            } else {
                if (!config.dockerName) {
                    config.dockerName = UUID.randomUUID().toString()
                }
                def dockerExecuteOnKubernetesParams = [
                    script: script,
                    additionalPodProperties: config.additionalPodProperties,
                    containerName: config.dockerName,
                    containerCommand: config.containerCommand,
                    containerShell: config.containerShell,
                    dockerImage: config.dockerImage,
                    dockerPullImage: config.dockerPullImage,
                    dockerEnvVars: config.dockerEnvVars,
                    dockerWorkspace: config.dockerWorkspace,
                    stashContent: config.stashContent,
                    stashNoDefaultExcludes: config.stashNoDefaultExcludes,
                    securityContext: securityContext,
                    containerMountPath: containerMountPath,
                ]

                if (config.sidecarImage) {
                    dockerExecuteOnKubernetesParams += [
                        containerPortMappings: config.containerPortMappings,
                        sidecarName: parameters.sidecarName,
                        sidecarImage: parameters.sidecarImage,
                        sidecarPullImage: parameters.sidecarPullImage,
                        sidecarReadyCommand: parameters.sidecarReadyCommand,
                        sidecarEnvVars: parameters.sidecarEnvVars,
                    ]
                }

                dockerExecuteOnKubernetes(dockerExecuteOnKubernetesParams) {
                    echo "[INFO][${STEP_NAME}] Executing inside a Kubernetes Pod. Docker image: ${config.dockerImage}"
                    body()
                }
            }
        } else {
            boolean executeInsideDocker = true
            if (!JenkinsUtils.isPluginActive(PLUGIN_ID_DOCKER_WORKFLOW)) {
                echo "[WARNING][${STEP_NAME}] Docker not supported. Plugin '${PLUGIN_ID_DOCKER_WORKFLOW}' is not installed or not active. Configured docker image '${config.dockerImage}' will not be used."
                executeInsideDocker = false
            }

            def returnCode = sh script: 'docker ps -q > /dev/null', returnStatus: true
            if (returnCode != 0) {
                echo "[WARNING][$STEP_NAME] Cannot connect to docker daemon (command 'docker ps' did not return with '0'). Configured docker image '${config.dockerImage}' will not be used."
                executeInsideDocker = false
            }
            if (executeInsideDocker && config.dockerImage) {
                utils.unstashAll(config.stashContent)
                def image = docker.image(config.dockerImage)
                pullWrapper(config.dockerPullImage, image, config.dockerRegistryUrl, config.dockerRegistryCredentialsId) {
                    if (!config.sidecarImage) {
                        image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                            body()
                        }
                    } else {
                        def networkName = "sidecar-${UUID.randomUUID()}"
                        sh "docker network create ${q(networkName)}"
                        try {
                            def sidecarImage = docker.image(config.sidecarImage)
                            pullWrapper(config.sidecarPullImage, sidecarImage, config.sidecarRegistryUrl, config.sidecarRegistryCredentialsId) {
                                config.sidecarOptions = config.sidecarOptions ?: []
                                if (config.sidecarName)
                                    config.sidecarOptions.add("--network-alias ${config.sidecarName}")
                                config.sidecarOptions.add("--network ${networkName}")
                                sidecarImage.withRun(getDockerOptions(config.sidecarEnvVars, config.sidecarVolumeBind, config.sidecarOptions)) { container ->
                                    config.dockerOptions = config.dockerOptions ?: []
                                    if (config.dockerName)
                                        config.dockerOptions.add("--network-alias ${config.dockerName}")
                                    config.dockerOptions.add("--network ${networkName}")
                                    if (config.sidecarReadyCommand) {
                                        sidecarUtils.waitForSidecarReadyOnDocker(container.id, config.sidecarReadyCommand)
                                    }
                                    image.inside(getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)) {
                                        echo "[INFO][${STEP_NAME}] Running with sidecar container."
                                        body()
                                    }
                                }
                            }
                        } finally {
                            sh "docker network remove ${networkName}"
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

void pullWrapper(boolean pullImage, def dockerImage, String dockerRegistryUrl, String dockerCredentialsId, Closure body) {
    if (!pullImage) {
        echo "[INFO][$STEP_NAME] Skipped pull of image '$dockerImage'."
        body()
        return
    }

    if (dockerCredentialsId) {
        // docker registry can be provided empty and will default to 'https://index.docker.io/v1/' in this case.
        docker.withRegistry(dockerRegistryUrl ?: '', dockerCredentialsId) {
            dockerImage.pull()
            body()
        }
    } else if (dockerRegistryUrl) {
        docker.withRegistry(dockerRegistryUrl) {
            dockerImage.pull()
            body()
        }
    } else {
        dockerImage.pull()
        body()
    }
}

/*
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
        dockerEnvVars.each { String k, v ->
            options.add("--env ${k}=${v.toString()}")
        }
    }

    specialEnvironments.each { String envVar ->
        if (dockerEnvVars == null || !dockerEnvVars.containsKey(envVar)) {
            options.add("--env ${envVar}")
        }
    }

    if (dockerVolumeBind) {
        dockerVolumeBind.each { String k, v ->
            options.add("--volume ${k}:${v.toString()}")
        }
    }

    if (dockerOptions) {
        options.addAll(dockerOptionsToList(dockerOptions))
    }

    return options.join(' ')
}

@NonCPS
def securityContextFromOptions(dockerOptions) {
    Map securityContext = [:]

    if (!dockerOptions) {
        return null
    }

    def userOption = dockerOptionsToList(dockerOptions).find { (it.startsWith("-u ") || it.startsWith("--user ")) }
    if (!userOption) {
        return null
    }

    def userOptionParts = userOption.split(" ")
    if (userOptionParts.size() != 2) {
        throw new IllegalArgumentException("Unexpected --user flag value in dockerOptions '${userOption}'")
    }

    def userGroupIds = userOptionParts[1].split(":")

    securityContext.runAsUser = userGroupIds[0].isInteger() ? userGroupIds[0].toInteger() : userGroupIds[0]

    if (userGroupIds.size() == 2) {
        securityContext.runAsGroup = userGroupIds[1].isInteger() ? userGroupIds[1].toInteger() : userGroupIds[1]
    }

    return securityContext
}

/*
 * Picks the first volumeBind option and translates it into containerMountPath, currently only one fix volume is supported
 */
@NonCPS
def containerMountPathFromVolumeBind(dockerVolumeBind) {
    if (dockerVolumeBind) {
        return dockerVolumeBind[0].split(":")[1]
    }
    return ""
}

boolean isContainerDefined(config) {
    Map containerMap = ContainerMap.instance.getMap()

    if (!containerMap.containsKey(env.POD_NAME)) {
        return false
    }

    if (env.SIDECAR_IMAGE != config.sidecarImage) {
        // If a sidecar image has been configured for the current stage,
        // then piperStageWrapper will have set the env.SIDECAR_IMAGE variable.
        // If the current step overrides the stage's sidecar image,
        // then a new Pod needs to be spawned.
        return false
    }

    return containerMap.get(env.POD_NAME).containsKey(config.dockerImage)
}


def getContainerDefined(config) {
    def containerMap = ContainerMap.instance.getMap()
    if (!containerMap.containsKey(env.POD_NAME)) {
        throw new IllegalStateException("POD_NAME not found in container map: ${env.POD_NAME}")
    }
    def podContainers = containerMap.get(env.POD_NAME)
    if (!podContainers.containsKey(config.dockerImage)) {
        throw new IllegalStateException("Docker image not found in pod. Image: ${config.dockerImage}, Pod: ${env.POD_NAME}")
    }
    return podContainers.get(config.dockerImage).toLowerCase()
}


boolean isKubernetes() {
    return Boolean.valueOf(env.ON_K8S)
}

@NonCPS
def dockerOptionsToList(dockerOptions) {
    def options = []
    if (!dockerOptions) {
        return options
    }

    if (dockerOptions instanceof CharSequence) {
        dockerOptions = [dockerOptions]
    }

    if (dockerOptions instanceof List) {
        dockerOptions.each { String option ->
            options << escapeBlanks(option)
        }
    } else {
        throw new IllegalArgumentException("Unexpected type for dockerOptions. Expected was either a list or a string. Actual type was: '${dockerOptions.getClass()}'")
    }

    return options
}

/*
 * Escapes blanks for values in key/value pairs
 * E.g. <code>description=Lorem ipsum</code> is
 * changed to <code>description=Lorem\ ipsum</code>.
 */

@NonCPS
def escapeBlanks(def s) {

    def EQ = '='
    def parts = s.split(EQ)

    if (parts.length == 2) {
        parts[1] = parts[1].replaceAll(' ', '\\\\ ')
        s = parts.join(EQ)
    }

    return s
}
