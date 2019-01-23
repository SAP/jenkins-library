import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.k8s.SystemEnv
import groovy.transform.Field
import hudson.AbortException

@Field def STEP_NAME = getClass().getName()
@Field def PLUGIN_ID_KUBERNETES = 'kubernetes'
@Field Set GENERAL_CONFIG_KEYS = ['jenkinsKubernetes']
@Field Set PARAMETER_KEYS = [
    'containerCommand', // specify start command for container created with dockerImage parameter to overwrite Piper default (`/usr/bin/tail -f /dev/null`).
    'containerCommands', //specify start command for containers to overwrite Piper default (`/usr/bin/tail -f /dev/null`). If container's default start command should be used provide empty string like: `['selenium/standalone-chrome': '']`
    'containerEnvVars', //specify environment variables per container. If not provided dockerEnvVars will be used
    'containerMap', //specify multiple images which then form a kubernetes pod, example: containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute','selenium/standalone-chrome': 'selenium']
    'containerName', //optional configuration in combination with containerMap to define the container where the commands should be executed in
    'containerPortMappings', //map which defines per docker image the port mappings, like containerPortMappings: ['selenium/standalone-chrome': [[name: 'selPort', containerPort: 4444, hostPort: 4444]]]
    'containerShell', // allows to specify the shell to be executed for container with containerName
    'containerWorkspaces', //specify workspace (=home directory of user) per container. If not provided dockerWorkspace will be used. If empty, home directory will not be set.
    'dockerImage',
    'dockerWorkspace',
    'dockerEnvVars',
    'stashContent'
]
@Field Set STEP_CONFIG_KEYS = PARAMETER_KEYS.plus(['stashIncludes', 'stashExcludes','alwaysPullImage'])

void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

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
    return [name      : 'dynamic-agent-' + config.uniqueId,
            label     : config.uniqueId,
            containers: getContainerList(config),
            alwaysPullImage: config.alwaysPullImage]
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
        if (config.containerName && config.stashContent.isEmpty()){
            config.stashContent.add(stashWorkspace(config, 'workspace'))
        }
        podTemplate(getOptions(config)) {
            node(config.uniqueId) {
                if (config.containerName) {
                    Map containerParams = [name: config.containerName]
                    if (config.containerShell) {
                        containerParams.shell = config.containerShell
                    }
                    container(containerParams){
                        try {
                            utils.unstashAll(config.stashContent)
                            body()
                        } finally {
                            stashWorkspace(config, 'container')
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

private String stashWorkspace(config, prefix) {
    def stashName = "${prefix}-${config.uniqueId}"
    try {
        // Every dockerImage used in the dockerExecuteOnKubernetes should have user id 1000
        sh "chown -R 1000:1000 ."
        stash(
            name: stashName,
            include: config.stashIncludes.workspace,
            exclude: config.stashExcludes.excludes
        )
        return stashName
    } catch (AbortException | IOException e) {
        echo "${e.getMessage()}"
    }
    return null
}

private void unstashWorkspace(config, prefix) {
    try {
        unstash "${prefix}-${config.uniqueId}"
    } catch (AbortException | IOException e) {
        echo "${e.getMessage()}"
    }
}

private List getContainerList(config) {

    result = []
    result.push(containerTemplate(
        name: 'jnlp',
        image: config.jenkinsKubernetes.jnlpAgent
    ))
    config.containerMap.each { imageName, containerName ->
        def templateParameters = [
            name: containerName.toLowerCase(),
            image: imageName,
            alwaysPullImage: true,
            envVars: getContainerEnvs(config, imageName)
        ]

        if (!config.containerCommands?.get(imageName)?.isEmpty()) {
            templateParameters.command = config.containerCommands?.get(imageName)?: '/usr/bin/tail -f /dev/null'
        }

        if (config.containerPortMappings?.get(imageName)) {
            def ports = []
            def portCounter = 0
            config.containerPortMappings.get(imageName).each {mapping ->
                mapping.name = "${containerName}${portCounter}".toString()
                ports.add(portMapping(mapping))
                portCounter ++
            }
            templateParameters.ports = ports
        }
        result.push(containerTemplate(templateParameters))
    }
    return result
}

/**
 * Returns a list of envVar object consisting of set
 * environment variables, params (Parametrized Build) and working directory.
 * (Kubernetes-Plugin only!)
 * @param config Map with configurations
 */
private List getContainerEnvs(config, imageName) {
    def containerEnv = []
    def dockerEnvVars = config.containerEnvVars?.get(imageName) ?: config.dockerEnvVars ?: [:]
    def dockerWorkspace = config.containerWorkspaces?.get(imageName) != null ? config.containerWorkspaces?.get(imageName) : config.dockerWorkspace ?: ''

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
