import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.SysEnv
import org.codehaus.groovy.GroovyException
import com.sap.piper.ConfigurationMerger
import java.util.UUID

def call(Map parameters = [:], body) {
    def STEP_NAME = 'executeDockerOnKubernetes'
    def PLUGIN_ID_KUBERNETES = 'kubernetes'

    handlePipelineStepErrors(stepName: 'executeDockerOnKubernetes', stepParameters: parameters) {
        if (!isPluginActive(PLUGIN_ID_KUBERNETES)) {
            error("[ERROR][${STEP_NAME}] not supported. Plugin '${PLUGIN_ID_KUBERNETES}' is not installed or not active.")
        }

        final script = parameters.script
        prepareDefaultValues script: script

        Set parameterKeys = ['dindImage',
                             'dockerImage',
                             'dockerWorkspace',
                             'dockerEnvVars']
        Set stepConfigurationKeys = parameterKeys

        Map config = ConfigurationMerger.merge(script, 'executeDockerOnKubernetes',
            parameters, parameterKeys,
            stepConfigurationKeys)

        config.uniqueId = UUID.randomUUID().toString()

        if (!config.dockerImage) throw new GroovyException('Docker image not specified.')

        def options = [name      : env.jaas_owner + '-jaas',
                       label     : config.uniqueId,
                       containers: getContainerList(config)]

        stashWorkspace(config)
        podTemplate(options) {
            node(config.uniqueId) {
                echo "Execute container content in Kubernetes pod"
                utils.unstashAll(config.stashContent)
                container(name: 'container-exec') {
                    body()
                }
                stashContainer(config)
            }
        }
        unstashContainer(config)
    }
}

private stashWorkspace(config) {
    if (config.stashContent.size() == 0) {
        try {
            sh "chmod -R u+w ."
            stash "workspace-${config.uniqueId}"
            config.stashContent += 'workspace-' + config.uniqueId
        } catch (hudson.AbortException e) {
            echo "${e.getMessage()}"
        } catch (java.io.IOException ioe) {
            echo "${ioe.getMessage()}"
        }
    }
}

private stashContainer(config) {
    def stashBackConfig = config.stashBackConfig
    try {
        stashBackConfig.name = "container-${config.uniqueId}"
        stash stashBackConfig
    } catch (hudson.AbortException e) {
        echo "${e.getMessage()}"
    } catch (java.io.IOException ioe) {
        echo "${ioe.getMessage()}"
    }
}

private unstashContainer(config) {
    try {
        unstash "container-${config.uniqueId}"
    } catch (hudson.AbortException e) {
        echo "${e.getMessage()}"
    } catch (java.io.IOException ioe) {
        echo "${ioe.getMessage()}"
    }
}

private getContainerList(config) {
    def envVars
    envVars = getContainerEnvs(config.dockerEnvVars, config.dockerWorkspace)
    if (config.dindImage) {
        envVars << envVar(key: 'DOCKER_HOST', value: '2375')
    }

    result = []
    result.push(containerTemplate(name: 'jnlp',
        image: 's4sdk/jnlp-k8s:latest',
        args: '${computer.jnlpmac} ${computer.name}'))
    result.push(containerTemplate(name: 'container-exec',
        image: config.dockerImage,
        alwaysPullImage: true,
        command: '/usr/bin/tail -f /dev/null',
        envVars: envVars))
    if (config.dindImage) result.push(containerTemplate(name: 'container-dind',
        image: config.dindImage,
        privileged: true))
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
    if (params) {
        for (String k : params.keySet()) {
            containerEnv << envVar(key: k, value: params[k].toString())
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

@NonCPS
private isPluginActive(String pluginId) {
    return Jenkins.instance.pluginManager.plugins.find { p -> p.isActive() && p.getShortName() == pluginId }
}
