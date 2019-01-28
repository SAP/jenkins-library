import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field Set STEP_CONFIG_KEYS = [
    'dockerImage'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS + ['dockerOptions']

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        dockerExecute(script: script, dockerImage: configuration.dockerImage, dockerOptions: configuration.dockerOptions) {
            sh '''npm run build'''
        }
    }
}
