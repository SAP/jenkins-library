import com.sap.piper.ConfigurationHelper

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'newmanExecute'
@Field Set STEP_CONFIG_KEYS = [
    'dockerImage',
    'failOnError',
    'newmanCollection',
    'newmanEnvironment',
    'newmanGlobals',
    'newmanRunCommand',
    'testRepository'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        List collectionList = findFiles(glob: config.newmanCollection)?.toList()

        if (!config.dockerImage.isEmpty()) {
            if (config.testRepository)
                git config.testRepository
            dockerExecute(
                dockerImage: config.dockerImage
            ) {
                sh 'npm install newman --global --quiet'
                for(String collection : collectionList){
                    // resolve templates
                    def command = SimpleTemplateEngine.newInstance()
                        .createTemplate(config.newmanRunCommand)
                        .make([config: config.plus([newmanCollection: collection])]).toString()
                    if(!config.failOnError) command += ' --suppress-exit-code'
                    sh "newman ${command}"
                }
            }
        }
    }
}
