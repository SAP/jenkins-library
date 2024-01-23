import com.sap.piper.SidecarUtils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.k8s.SystemEnv
import com.sap.piper.JsonUtils

import groovy.transform.Field
import hudson.AbortException

@Field def STEP_NAME = getClass().getName()
@Field def PLUGIN_ID_KUBERNETES = 'kubernetes'

@Field Set GENERAL_CONFIG_KEYS = [
    'jenkinsKubernetes',
        /**
         * Jnlp agent Docker images which should be used to create new pods.
         * @parentConfigKey jenkinsKubernetes
         */
        'jnlpAgent',
        /**
         * Namespace that should be used to create a new pod
         * @parentConfigKey jenkinsKubernetes
         */
        'namespace',
        /**
         * Name of the pod template that should be inherited from.
         * The pod template can be defined in the Jenkins UI
         * @parentConfigKey jenkinsKubernetes
         */
        'inheritFrom',
        'additionalPodProperties',
        'resources',
        'annotations',
    /**
     * Set this to 'false' to bypass a docker image pull.
     * Useful during development process. Allows testing of images which are available in the local registry only.
     */
    'dockerPullImage',
    /**
     * Set this to 'false' to bypass a docker image pull.
     * Useful during development process. Allows testing of images which are available in the local registry only.
     */
    'sidecarPullImage',
    /**
     * Print more detailed information into the log.
     * @possibleValues `true`, `false`
     */
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([

    /**
     * Additional pod specific configuration. Map with the properties names
     * as key and the corresponding value as value. The value can also be
     * a nested structure.
     * The properties will be added to the pod spec inside node `spec` at the
     * same level like e.g. `containers`
     * for eg., additionalPodProperties: [
     *               imagePullSecrets: ['secret-name']
     *        ]
     * This property provides some kind of an expert mode. Any property
     * which is not handled otherwise by the step can be set. It is not
     * possible to overwrite e.g. the `containers` property or to
     * overwrite the `securityContext` property.
     * Alternate way for providing `additionalPodProperties` is via
     * `general/jenkinsKubernetes/additionalPodProperties` in the project configuration.
     * Providing the resources map as parameter to the step call takes
     * precedence.
     * This freedom comes with great responsibility. The property
     * `additionalPodProperties` should only be used in case you
     * really know what you are doing.
     */
    'additionalPodProperties',
    /**
    * Adds annotations in the metadata section of the PodSpec
    */
    'annotations',
    /**
     * Allows to specify start command for container created with dockerImage parameter to overwrite Piper default (`/usr/bin/tail -f /dev/null`).
     */
    'containerCommand',
    /**
     * Specifies start command for containers to overwrite Piper default (`/usr/bin/tail -f /dev/null`).
     * If container's defaultstart command should be used provide empty string like: `['selenium/standalone-chrome': '']`.
     */
    'containerCommands',
    /**
     * Specifies environment variables per container. If not provided `dockerEnvVars` will be used.
     */
    'containerEnvVars',
    /**
     * A map of docker image to the name of the container. The pod will be created with all the images from this map and they are labelled based on the value field of each map entry.
     * Example: `['maven:3.5-jdk-8-alpine': 'mavenExecute', 'selenium/standalone-chrome': 'selenium', 'famiko/jmeter-base': 'checkJMeter', 'ppiper/cf-cli:6': 'cloudfoundry']`
     */
    'containerMap',
    /**
     * Optional configuration in combination with containerMap to define the container where the commands should be executed in.
     */
    'containerName',
    /**
     * Map which defines per docker image the port mappings, e.g. `containerPortMappings: ['selenium/standalone-chrome': [[name: 'selPort', containerPort: 4444, hostPort: 4444]]]`.
     */
    'containerPortMappings',
    /**
     * Specifies the pullImage flag per container.
     */
    'containerPullImageFlags',
    /**
     * Allows to specify the shell to be executed for container with containerName.
     */
    'containerShell',
    /**
     * Specifies a dedicated user home directory per container which will be passed as value for environment variable `HOME`. If not provided `dockerWorkspace` will be used.
     */
    'containerWorkspaces',
    /**
     * Environment variables to set in the container, e.g. [http_proxy:'proxy:8080'].
     */
    'dockerEnvVars',
    /**
     * Optional name of the docker image that should be used. If no docker image is provided, the closure will be executed in the jnlp agent container.
     */
    'dockerImage',
    /**
     * Specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`.
     */
    'dockerWorkspace',
    /**
     * as `dockerImage` for the sidecar container
     */
    'sidecarImage',
    /**
     * SideCar only:
     * Name of the container in local network.
     */
    'sidecarName',
    /**
     * Command executed inside the container which returns exit code 0 when the container is ready to be used.
     */
    'sidecarReadyCommand',
    /**
     * as `dockerEnvVars` for the sidecar container
     */
    'sidecarEnvVars',
    /**
     * as `dockerWorkspace` for the sidecar container
     */
    'sidecarWorkspace',

    /** Defines the Kubernetes nodeSelector as per [https://github.com/jenkinsci/kubernetes-plugin](https://github.com/jenkinsci/kubernetes-plugin).*/
    'nodeSelector',
    /**
     * Kubernetes Security Context used for the pod.
     * Can be used to specify uid and fsGroup.
     * See: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
     */
    'securityContext',
    /**
     * Specific stashes that should be considered for the step execution.
     */
    'stashContent',
    /**
     * In the Kubernetes case the workspace is only available to the respective Jenkins slave but not to the containers running inside the pod.<br />
     * This configuration defines exclude pattern for stashing from Jenkins workspace to working directory in container and back.
     * Following excludes can be set:
     *
     * * `workspace`: Pattern for stashing towards container
     * * `stashBack`: Pattern for bringing data from container back to Jenkins workspace. If not set: defaults to setting for `workspace`.
     */
    'stashExcludes',
    /**
     * In the Kubernetes case the workspace is only available to the respective Jenkins slave but not to the containers running inside the pod.<br />
     * This configuration defines include pattern for stashing from Jenkins workspace to working directory in container and back.
     * Following includes can be set:
     *
     * * `workspace`: Pattern for stashing towards container
     * * `stashBack`: Pattern for bringing data from container back to Jenkins workspace. If not set: defaults to setting for `workspace`.
     */
    'stashIncludes',
    /**
     * In the Kubernetes case the workspace is only available to the respective Jenkins slave but not to the containers running inside the pod.<br />
     * This configuration defines include pattern for stashing from Jenkins workspace to working directory in container and back.
     * This flag controls whether the stashing does *not* use the default exclude patterns in addition to the patterns provided in `stashExcludes`.
     * @possibleValues `true`, `false`
     */
    'stashNoDefaultExcludes',
    /**
     * A map containing the resources per container. The key is the
     * container name. The value is a map defining valid resources.
     * An entry with key `DEFAULT` can be used for defining resources
     * for all containers which does not have resources specified otherwise.
     * Alternate way for providing resources is via `general/jenkinsKubernetes/resources`
     * in the project configuration. Providing the resources map as parameter
     * to the step call takes precedence.
     */
    'resources',
    /**
     * The path to which a volume should be mounted to. This volume will be available at the same
     * mount path in each container of the provided containerMap. The volume is of type emptyDir
     * and has the name 'volume'. With the additionalPodProperties parameter one can for example
     * use this volume in an initContainer.
     */
    'containerMountPath',
     /**
     * The docker image to run as initContainer.
     */
    'initContainerImage',
    /**
     * Command executed inside the init container shell. Please enter command without providing any "sh -c" prefix. For example for an echo message, simply enter: echo `HelloWorld`
     */
    'initContainerCommand',

])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.minus([
    'stashIncludes',
    'stashExcludes'
])

/**
 * Executes a closure inside a container in a kubernetes pod.
 * Proxy environment variables defined on the Jenkins machine are also available in the container.
 *
 * By default jnlp agent defined for kubernetes-plugin will be used (see [https://github.com/jenkinsci/kubernetes-plugin#pipeline-support](https://github.com/jenkinsci/kubernetes-plugin#pipeline-support)).
 *
 * It is possible to define a custom jnlp agent image by
 *
 * 1. Defining the jnlp image via environment variable JENKINS_JNLP_IMAGE in the Kubernetes landscape
 * 2. Defining the image via config (`jenkinsKubernetes.jnlpAgent`)
 *
 * Option 1 will take precedence over option 2.
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        final script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        if (!JenkinsUtils.isPluginActive(PLUGIN_ID_KUBERNETES)) {
            error("[ERROR][${STEP_NAME}] not supported. Plugin '${PLUGIN_ID_KUBERNETES}' is not installed or not active.")
        }

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('uniqueId', UUID.randomUUID().toString())
            .use()

        if (!config.containerMap && config.dockerImage) {
            config.containerName = 'container-exec'
            config.containerMap = [(config.get('dockerImage')): config.containerName]
            config.containerCommands = config.containerCommand ? [(config.get('dockerImage')): config.containerCommand] : null
        }
        executeOnPod(config, utils, body, script)
    }
}

def getOptions(config) {
    def namespace = config.jenkinsKubernetes.namespace
    def options = [
        name : 'dynamic-agent-' + config.uniqueId,
        label: config.uniqueId,
        yaml : generatePodSpec(config)
    ]
    if (namespace) {
        options.namespace = namespace
    }
    if (config.nodeSelector) {
        options.nodeSelector = config.nodeSelector
    }
    if (!config.verbose) {
        options.showRawYaml = false
    }

    if(config.jenkinsKubernetes.inheritFrom){
        options.inheritFrom = config.jenkinsKubernetes.inheritFrom
        options.yamlMergeStrategy  = merge()
    }
    return options
}

void executeOnPod(Map config, utils, Closure body, Script script) {
    /*
     * There could be exceptions thrown by
        - The podTemplate
        - The container method
        - The body
     * We use nested exception handling in this case.
     * In the first 2 cases, the 'container' stash is not created because the inner try/finally is not reached.
     * However, the workspace has not been modified and don't need to be restored.
     * In case third case, we need to create the 'container' stash to bring the modified content back to the host.
     */
    try {
        SidecarUtils sidecarUtils = new SidecarUtils(script)
        def stashContent = config.stashContent
        boolean defaultStashCreated = false
        if (config.containerName && stashContent.isEmpty()) {
            stashContent = [stashWorkspace(config, utils, 'workspace')]
            defaultStashCreated = true
        }
        podTemplate(getOptions(config)) {
            node(config.uniqueId) {
                if (config.sidecarReadyCommand) {
                    sidecarUtils.waitForSidecarReadyOnKubernetes(config.sidecarName, config.sidecarReadyCommand)
                }
                if (config.containerName) {
                    Map containerParams = [name: config.containerName]
                    if (config.containerShell) {
                        containerParams.shell = config.containerShell
                    }
                    echo "ContainerConfig: ${containerParams}"
                    container(containerParams) {
                        try {
                            utils.unstashAll(stashContent)
                            if (config.verbose) {
                                lsDir('Directory content before body execution')
                            }
                            if (defaultStashCreated) {
                                invalidateStash(config, 'workspace', utils)
                            }
                            def result = body()
                            if (config.verbose) {
                                lsDir('Directory content after body execution')
                            }
                            return result
                        } finally {
                            stashWorkspace(config, utils, 'container', true, true)
                        }
                    }
                } else {
                    body()
                }
            }
        }
    } finally {
        if (config.containerName)
            unstashWorkspace(config, utils, 'container')
    }
}

private void lsDir(String message) {
  echo "[DEBUG] Begin of ${message}"
  // some images might not contain the find command. In that case the build must not be aborted.
  catchError (message: 'Cannot list directory content', buildResult: 'SUCCESS', stageResult: 'SUCCESS') {
    // no -ls option since this is not available for some images
    sh  'find . -mindepth 1 -maxdepth 2'
  }
  echo "[DEBUG] End of ${message}"
}

private String generatePodSpec(Map config) {
    def podSpec = [
        apiVersion: "v1",
        kind      : "Pod",
        metadata  : [
            lables: config.uniqueId,
            annotations: [:]
        ],
        spec      : [:]
    ]
    podSpec.metadata.annotations = getAnnotations(config)
    podSpec.spec += getAdditionalPodProperties(config)
    podSpec.spec.initContainers = getInitContainerList(config)
    podSpec.spec.containers = getContainerList(config)
    podSpec.spec.securityContext = getSecurityContext(config)

    if (config.containerMountPath) {
        podSpec.spec.volumes = [[
                                    name    : "volume",
                                    emptyDir: [:]
                                ]]
    }

    return new JsonUtils().groovyObjectToPrettyJsonString(podSpec)
}

private String stashWorkspace(config, utils, prefix, boolean chown = false, boolean stashBack = false) {
    def stashName = "${prefix}-${config.uniqueId}"
    try {
        if (chown) {
            def securityContext = getSecurityContext(config)
            def runAsUser = securityContext?.runAsUser ?: 1000
            def fsGroup = securityContext?.fsGroup ?: 1000
            sh """#!${config.containerShell ?: '/bin/sh'}
chown -R ${runAsUser}:${fsGroup} ."""
        }

        def includes, excludes

        if (config.verbose) {
            echo "stashIncludes config: ${config.stashIncludes}"
            echo "stashExcludes config: ${config.stashExcludes}"
        }

        if (stashBack) {
            includes = config.stashIncludes.stashBack ?: config.stashIncludes.workspace
            excludes = config.stashExcludes.stashBack ?: config.stashExcludes.workspace
        } else {
            includes = config.stashIncludes.workspace
            excludes = config.stashExcludes.workspace
        }

        if (config.verbose) {
            echo "stash effective (includes): ${includes}"
            echo "stash effective (excludes): ${excludes}"
        }

        utils.stash(
            name: stashName,
            includes: includes,
            excludes: excludes,
            // 'true' by default due to negative side-effects, but can be overwritten via parameters
            // (as done by artifactPrepareVersion to preserve the .git folder)
            useDefaultExcludes: !config.stashNoDefaultExcludes,
            allowEmpty: true
        )
        return stashName
    } catch (AbortException | IOException e) {
        echo "${e.getMessage()}"
    } catch (Throwable e) {
        echo "Unstash workspace failed with throwable ${e.getMessage()}"
        throw e
    }
    return null
}

private Map getAnnotations(Map config){
  return config.annotations ?: config.jenkinsKubernetes.annotations ?: [:]
}

private Map getAdditionalPodProperties(Map config) {
    Map podProperties = config.additionalPodProperties ?: config.jenkinsKubernetes.additionalPodProperties ?: [:]
    if(podProperties) {
        echo "Additional pod properties found (${podProperties.keySet()})." +
        ' Providing additional pod properties is some kind of expert mode. In case of any problems caused by these' +
        ' additional properties only limited support can be provided.'
    }
    return podProperties
}

private Map getSecurityContext(Map config) {
    return config.securityContext ?: config.jenkinsKubernetes.securityContext ?: [:]
}

private void unstashWorkspace(config, utils, prefix) {
    try {
        utils.unstash "${prefix}-${config.uniqueId}"
    } catch (AbortException | IOException e) {
        echo "${e.getMessage()}\n${e.getCause()}"
    } catch (Throwable e) {
        echo "Unstash workspace failed with throwable ${e.getMessage()}"
        throw e
    } finally {
        invalidateStash(config, prefix, utils)
    }
}

private List getInitContainerList(config){
    def initContainerSpecList = []
    if (config.initContainerImage && config.containerMountPath) {
        // regex [\W_] matches any non-word character equivalent to [^a-zA-Z0-9_]
        def initContainerName = config.initContainerImage.toLowerCase().replaceAll(/[\W_]/,"-" )
        def initContainerSpec = [
            name           : initContainerName,
            image          : config.initContainerImage
            ]
        if (config.containerMountPath) {
            initContainerSpec.volumeMounts = [[name: "volume", mountPath: config.containerMountPath]]
        }
        if (config.initContainerCommand == null) {
            initContainerSpec['command'] = [
                '/usr/bin/tail',
                '-f',
                '/dev/null'
            ]
        } else {
            initContainerSpec['command'] = [
                'sh',
                '-c',
                config.initContainerCommand
            ]
        }
        initContainerSpecList.push(initContainerSpec)
    }
    return initContainerSpecList
}

private List getContainerList(config) {

    //If no custom jnlp agent provided as default jnlp agent (jenkins/jnlp-slave) as defined in the plugin, see https://github.com/jenkinsci/kubernetes-plugin#pipeline-support
    def result = []
    //allow definition of jnlp image via environment variable JENKINS_JNLP_IMAGE in the Kubernetes landscape or via config as fallback
    if (env.JENKINS_JNLP_IMAGE || config.jenkinsKubernetes.jnlpAgent) {

        def jnlpContainerName = 'jnlp'

        def jnlpSpec = [
            name : jnlpContainerName,
            image: env.JENKINS_JNLP_IMAGE ?: config.jenkinsKubernetes.jnlpAgent
        ]

        def resources = getResources(jnlpContainerName, config)
        if(resources) {
            jnlpSpec.resources = resources
        }

        result.push(jnlpSpec)
    }
    config.containerMap.each { imageName, containerName ->
        def containerPullImage = config.containerPullImageFlags?.get(imageName)
        boolean pullImage = containerPullImage != null ? containerPullImage : config.dockerPullImage
        def containerSpec = [
            name           : containerName.toLowerCase(),
            image          : imageName,
            imagePullPolicy: pullImage ? "Always" : "IfNotPresent",
            env            : getContainerEnvs(config, imageName, config.dockerEnvVars, config.dockerWorkspace)
        ]
        if (config.containerMountPath) {
            containerSpec.volumeMounts = [[name: "volume", mountPath: config.containerMountPath]]
        }

        def configuredCommand = config.containerCommands?.get(imageName)
        def shell = config.containerShell ?: '/bin/sh'
        if (configuredCommand == null) {
            containerSpec['command'] = [
                '/usr/bin/tail',
                '-f',
                '/dev/null'
            ]
        } else if (configuredCommand != "") {
            // apparently "" is used as a flag for not settings container commands !?
            containerSpec['command'] =
                (configuredCommand in List) ? configuredCommand : [
                    shell,
                    '-c',
                    configuredCommand
                ]
        }

        if (config.containerPortMappings?.get(imageName)) {
            def ports = []
            def portCounter = 0
            config.containerPortMappings.get(imageName).each { mapping ->
                def name = "${containerName}${portCounter}".toString()
                if (mapping.containerPort != mapping.hostPort) {
                    echo("[WARNING][${STEP_NAME}]: containerPort and hostPort are different for container '${containerName}'. "
                        + "The hostPort will be ignored.")
                }
                ports.add([name: name, containerPort: mapping.containerPort])
                portCounter++
            }
            containerSpec.ports = ports
        }
        def resources = getResources(containerName.toLowerCase(), config)
        if(resources) {
            containerSpec.resources = resources
        }
        result.push(containerSpec)
    }
    if (config.sidecarImage) {
        def sideCarContainerName = config.sidecarName.toLowerCase()
        def containerSpec = [
            name           : sideCarContainerName,
            image          : config.sidecarImage,
            imagePullPolicy: config.sidecarPullImage ? "Always" : "IfNotPresent",
            env            : getContainerEnvs(config, config.sidecarImage, config.sidecarEnvVars, config.sidecarWorkspace),
            command        : []
        ]
        def resources = getResources(sideCarContainerName, config)
        if (resources) {
            containerSpec.resources = resources
        }
        if (config.containerMountPath) {
            containerSpec.volumeMounts = [[name: "volume", mountPath: config.containerMountPath]]
        }
        result.push(containerSpec)
    }
    return result
}

private Map getResources(String containerName, Map config) {
    Map resources = config.resources
    if(resources == null) {
        resources = config?.jenkinsKubernetes.resources
    }
    if(resources == null) {
        return null
    }
    Map res = resources.get(containerName)
    if(res == null) {
        res = resources.get('DEFAULT')
    }
    return res
}

/*
 * Returns a list of envVar object consisting of set
 * environment variables, params (Parametrized Build) and working directory.
 * (Kubernetes-Plugin only!)
 * @param config Map with configurations
 */

private List getContainerEnvs(config, imageName, defaultEnvVars, defaultConfig) {
    def containerEnv = []
    def dockerEnvVars = config.containerEnvVars?.get(imageName) ?: defaultEnvVars ?: [:]
    def dockerWorkspace = config.containerWorkspaces?.get(imageName) != null ? config.containerWorkspaces?.get(imageName) : defaultConfig ?: ''

    def envVar = { e ->
        [name: e.key, value: e.value]
    }

    if (dockerEnvVars) {
        dockerEnvVars.each {
            k, v ->
            containerEnv << envVar(key: k, value: v.toString())
        }
    }

    if (dockerWorkspace) {
        containerEnv << envVar(key: "HOME", value: dockerWorkspace)
    }

    // Inherit the proxy information from the master to the container
    SystemEnv systemEnv = new SystemEnv()
    systemEnv.getEnv().each {
        k, v ->
            containerEnv << envVar(key: k, value: v)
    }

    return containerEnv
}

private void invalidateStash(def config, String prefix, def utils) {
    String name = "${prefix}-${config.uniqueId}"
    echo "invalidate stash ${name}"
    utils.stash name: name, excludes: '**/*', allowEmpty: true
}
