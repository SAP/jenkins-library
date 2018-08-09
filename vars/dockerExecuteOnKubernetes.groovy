import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
import groovy.transform.Field


@Field def STEP_NAME = 'dockerExecuteOnKubernetes'
@Field def PLUGIN_ID_KUBERNETES = 'kubernetes'
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set PARAMETER_KEYS = ['dindImage',
                             'dockerImage',
                             'dockerWorkspace',
                             'dockerEnvVars']
@Field Set STEP_CONFIG_KEYS = PARAMETER_KEYS.plus(['stashContent', 'stashIncludes', 'stashExcludes'])

void call(Map parameters = [:], body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def jUtils = new JenkinsUtils()
        if (!jUtils.isPluginActive(PLUGIN_ID_KUBERNETES)) {
            error("[ERROR][${STEP_NAME}] not supported. Plugin '${PLUGIN_ID_KUBERNETES}' is not installed or not active.")
        }

        final script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('dockerImage')
            .use()

        config.uniqueId = UUID.randomUUID().toString()

        Map containersMap = [:]
        containersMap[config.get('dockerImage').toString()] = 'container-exec'

        stashWorkspace(config)
        containerExecuteInsidePod(script: script, containersMap: containersMap, dockerEnvVars: config.dockerEnvVars, dockerWorkspace: config.dockerWorkspace) {
            container(name: 'container-exec') {
                unstashWorkspace(config)
                try {
                    body()
                } finally {
                    stashContainer(config)
                }
            }
        }
        unstashContainer(config)
    }
}

private stashWorkspace(config) {
    try {
        sh "chmod -R u+w ."
        stash name: "workspace-${config.uniqueId}", include: config.stashIncludes.all, exclude: config.stashExcludes.excludes
    } catch (hudson.AbortException e) {
        echo "${e.getMessage()}"
    } catch (java.io.IOException ioe) {
        echo "${ioe.getMessage()}"
    }
}

private stashContainer(config) {
    try {
        sh "chmod -R u+w ."
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

