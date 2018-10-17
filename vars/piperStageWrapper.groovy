import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import groovy.transform.Field

@Field String STEP_NAME = 'piperStageWrapper'

void call(Map parameters = [:], body) {

    def script = parameters.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    // load default & individual configuration
    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixin(ConfigurationLoader.defaultStageConfiguration(this, stageName))
        .mixinGeneralConfig(script.commonPipelineEnvironment)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName)
        .mixin(parameters)
        .use()

    stageLocking(config) {
        withNode(config) {
            try {
                utils.unstashAll(config.stashContent)

                def scriptFileName = stageExitFilePath(stageName, config)
                if (fileExists(scriptFileName)) {
                    Script interceptor = load(scriptFileName)
                    echo "[${STEP_NAME}] Running interceptor '${scriptFileName}' for ${stageName}."
                    //passing config twice to keep compatibility with https://github.com/SAP/cloud-s4-sdk-pipeline-lib/blob/master/vars/runAsStage.groovy
                    interceptor(body, stageName, config, config)
                } else {
                    body()
                }
            } finally {
                deleteDir()
            }
        }
    }
}

String stageExitFilePath(String stageName, Map config) {
    return "${config.extensionLocation}${stageName.replace(' ', '_').toLowerCase()}.groovy"
}

void stageLocking(Map config, Closure body) {
    if (config.stageLocking) {
        lock(resource: "${env.JOB_NAME}/${config.ordinal}", inversePrecedence: true) {
            milestone config.ordinal
            body()
        }
    } else {
        body()
    }
}

void withNode(Map config, Closure body) {
    if (config.withNode) {
        node(config.nodeLabel) {
            body()
        }
    } else {
        body()
    }
}
