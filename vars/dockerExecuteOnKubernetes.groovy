import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field
import org.codehaus.groovy.GroovyException

@Field def STEP_NAME = 'dockerExecuteOnKubernetes'
@Field def PLUGIN_ID_KUBERNETES = 'kubernetes'

@Field Set PARAMETER_KEYS = ['dindImage',
                             'dockerImage',
                             'dockerWorkspace',
                             'dockerEnvVars']
@Field Set STEP_CONFIG_KEYS = PARAMETER_KEYS.plus(['stashContent', 'stashIncludes', 'stashExcludes'])

def call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        if (!isPluginActive(PLUGIN_ID_KUBERNETES)) {
            error("[ERROR][${STEP_NAME}] not supported. Plugin '${PLUGIN_ID_KUBERNETES}' is not installed or not active.")
        }

        final script = parameters.script
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('dockerImage')
            .use()

        config.uniqueId = UUID.randomUUID().toString()

        if (!config.dockerImage) throw new GroovyException('Docker image not specified.')
        Map containersMap = [:]
        containersMap[config.get('dockerImage').toString()] = 'container-exec'

        stashWorkspace(config)
        runInsidePod(script: script, containersMap: containersMap, dockerEnvVars: config.dockerEnvVars, dockerWorkspace: config.dockerWorkspace) {
            unstashWorkspace(config)
            container(name: 'container-exec') {
                body()
            }
            stashContainer(config)
        }
        unstashContainer(config)
    }
}

private stashWorkspace(config) {
    if (config.stashContent.size() == 0) {
        try {
            sh "chown -R 1000:1000 ."
            stash name: "workspace-${config.uniqueId}", include: config.stashIncludes.all, exclude: config.stashExcludes.excludes
        } catch (hudson.AbortException e) {
            echo "${e.getMessage()}"
        } catch (java.io.IOException ioe) {
            echo "${ioe.getMessage()}"
        }
    }
}

private stashContainer(config) {
    try {
        sh "chown -R 1000:1000 ."
        stash name: "container-${config.uniqueId}", include: config.stashIncludes.all, exclude: config.stashExcludes.excludes
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

private unstashWorkspace(config) {
    try {
        unstash "workspace-${config.uniqueId}"
    } catch (hudson.AbortException e) {
        echo "${e.getMessage()}"
    } catch (java.io.IOException ioe) {
        echo "${ioe.getMessage()}"
    }
}

@NonCPS
private isPluginActive(String pluginId) {
    return Jenkins.instance.pluginManager.plugins.find { p -> p.isActive() && p.getShortName() == pluginId }
}
