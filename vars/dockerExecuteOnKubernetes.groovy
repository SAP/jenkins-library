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
    'jenkinsKubernetes'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
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
     * A map of docker image to the name of the container. The pod will be created with all the images from this map and they are labled based on the value field of each map entry.
     * Example: `['maven:3.5-jdk-8-alpine': 'mavenExecute', 'selenium/standalone-chrome': 'selenium', 'famiko/jmeter-base': 'checkJMeter', 's4sdk/docker-cf-cli': 'cloudfoundry']`
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
     * Name of the docker image that should be used. If empty, Docker is not used.
     */
    'dockerImage',
    /**
     * Set this to 'false' to bypass a docker image pull.
     * Usefull during development process. Allows testing of images which are available in the local registry only.
     */
    'dockerPullImage',
    /**
     * Specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`.
     */
    'dockerWorkspace',
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
     *
     */
    'stashExcludes',
    /**
     *
     */
    'stashIncludes'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.minus([
    'stashIncludes',
    'stashExcludes'
])

/**
 * Executes a closure inside a container in a kubernetes pod.
 * Proxy environment variables defined on the Jenkins machine are also available in the container.
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        final script = checkScript(this, parameters) ?: this

        if (!JenkinsUtils.isPluginActive(PLUGIN_ID_KUBERNETES)) {
            error("[ERROR][${STEP_NAME}] not supported. Plugin '${PLUGIN_ID_KUBERNETES}' is not installed or not active.")
        }

        def utils = parameters?.juStabUtils ?: new Utils()

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('uniqueId', UUID.randomUUID().toString())
        Map config = configHelper.use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        if (!parameters.containerMap) {
            configHelper.withMandatoryProperty('dockerImage')
            config.containerName = 'container-exec'
            config.containerMap = ["${config.get('dockerImage')}": config.containerName]
            config.containerCommands = config.containerCommand ? ["${config.get('dockerImage')}": config.containerCommand] : null
        }
        executeOnPod(config, utils, body)
    }
}

def getOptions(config) {
    def namespace = config.jenkinsKubernetes.namespace
    def options = [
        name      : 'dynamic-agent-' + config.uniqueId,
        label     : config.uniqueId,
        yaml      : generatePodSpec(config)
    ]
    if (namespace) {
        options.namespace = namespace
    }
    if (config.nodeSelector) {
        options.nodeSelector = config.nodeSelector
    }

    return options
}

void executeOnPod(Map config, utils, Closure body) {
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

        def stashContent = config.stashContent
        if (config.containerName && stashContent.isEmpty()){
            stashContent = [stashWorkspace(config, 'workspace')]
        }
        podTemplate(getOptions(config)) {
            node(config.uniqueId) {
                if (config.containerName) {
                    Map containerParams = [name: config.containerName]
                    if (config.containerShell) {
                        containerParams.shell = config.containerShell
                    }
                    echo "ContainerConfig: ${containerParams}"
                    container(containerParams){
                        try {
                            utils.unstashAll(stashContent)
                            body()
                        } finally {
                            stashWorkspace(config, 'container', true)
                        }
                    }
                } else {
                    body()
                }
            }
        }
    } finally {
        if (config.containerName)
            unstashWorkspace(config, 'container')
    }
}

private String generatePodSpec(Map config) {
    def containers = getContainerList(config)
    def podSpec = [
        apiVersion: "v1",
        kind: "Pod",
        metadata: [
            lables: config.uniqueId
        ],
        spec: [
            containers: containers
        ]
    ]
    podSpec.spec.securityContext = getSecurityContext(config)

    return new JsonUtils().groovyObjectToPrettyJsonString(podSpec)
}


private String stashWorkspace(config, prefix, boolean chown = false) {
    def stashName = "${prefix}-${config.uniqueId}"
    try {
        if (chown)  {
            def securityContext = getSecurityContext(config)
            def runAsUser = securityContext?.runAsUser ?: 1000
            def fsGroup = securityContext?.fsGroup ?: 1000
            sh """#!${config.containerShell?:'/bin/sh'}
chown -R ${runAsUser}:${fsGroup} ."""
        }
        stash(
            name: stashName,
            includes: config.stashIncludes.workspace,
            excludes: config.stashExcludes.workspace,
            //inactive due to negative side-effects, we may require a dedicated git stash to be used
            //useDefaultExcludes: false
        )
        return stashName
    } catch (AbortException | IOException e) {
        echo "${e.getMessage()}"
    }
    return null
}

private Map getSecurityContext(Map config) {
    return config.securityContext ?: config.jenkinsKubernetes.securityContext ?: [:]
}

private void unstashWorkspace(config, prefix) {
    try {
        unstash "${prefix}-${config.uniqueId}"
    } catch (AbortException | IOException e) {
        echo "${e.getMessage()}"
    }
}

private List getContainerList(config) {
    def result = [[
        name: 'jnlp',
        image: config.jenkinsKubernetes.jnlpAgent
    ]]
    config.containerMap.each { imageName, containerName ->
        def containerPullImage = config.containerPullImageFlags?.get(imageName)
        def containerSpec = [
            name: containerName.toLowerCase(),
            image: imageName,
            imagePullPolicy: containerPullImage ? "Always" : "IfNotPresent",
            env: getContainerEnvs(config, imageName)
        ]

        def configuredCommand = config.containerCommands?.get(imageName)
        def shell = config.containerShell ?: '/bin/sh'
        if (configuredCommand == null) {
            containerSpec['command'] = [
                '/usr/bin/tail',
                '-f',
                '/dev/null'
            ]
        } else if(configuredCommand != "") {
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
            config.containerPortMappings.get(imageName).each {mapping ->
                def name = "${containerName}${portCounter}".toString()
                if(mapping.containerPort != mapping.hostPort) {
                    echo ("[WARNING][${STEP_NAME}]: containerPort and hostPort are different for container '${containerName}'. "
                        + "The hostPort will be ignored.")
                }
                ports.add([name: name, containerPort: mapping.containerPort])
                portCounter ++
            }
            containerSpec.ports = ports
        }
        result.push(containerSpec)
    }
    return result
}

/*
 * Returns a list of envVar object consisting of set
 * environment variables, params (Parametrized Build) and working directory.
 * (Kubernetes-Plugin only!)
 * @param config Map with configurations
 */
private List getContainerEnvs(config, imageName) {
    def containerEnv = []
    def dockerEnvVars = config.containerEnvVars?.get(imageName) ?: config.dockerEnvVars ?: [:]
    def dockerWorkspace = config.containerWorkspaces?.get(imageName) != null ? config.containerWorkspaces?.get(imageName) : config.dockerWorkspace ?: ''

    def envVar = { e ->
        [ name: e.key, value: e.value ]
    }

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

    return containerEnv
}
