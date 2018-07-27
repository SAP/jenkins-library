import com.sap.piper.ConfigurationMerger
import com.sap.piper.Utils
import org.codehaus.groovy.GroovyException

def call(Map parameters = [:], body) {
    def STEP_NAME = 'executeDockerOnKubernetes'
    def PLUGIN_ID_KUBERNETES = 'kubernetes'

    handlePipelineStepErrors(stepName: 'executeDockerOnKubernetes', stepParameters: parameters) {
        def utils= new Utils()
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


        if (!config.dockerImage) throw new GroovyException('Docker image not specified.')
        Map containersMap = [:]
        containersMap[config.get('dockerImage').toString()] = 'container-exec'
        stashWorkspace(config)
        runInsidePod(script: script, containersMap: containersMap, dockerEnvVars: config.dockerEnvVars, dockerWorkspace: config.dockerWorkspace) {
            echo "Execute container content in Kubernetes pod"
            utils.unstashAll(config.stashContent)
            body()
            stashContainer(config)
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

