import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

void call(Map parameters = [:], body) {
    def stepParameters = parameters.stepParameters //mandatory
    def stepName = parameters.stepName //mandatory
    def verbose = parameters.get('echoDetails', true)
    def message = ''
    try {
        if (stepParameters == null && stepName == null)
            error "The step handlePipelineStepErrors requires following mandatory parameters: stepParameters, stepName"

        if (verbose)
            echo "--- Begin library step of: ${stepName} ---"

        body()
    } catch (Throwable error) {
        if (verbose)
            message += SimpleTemplateEngine.newInstance()
                .createTemplate(libraryResource('com.sap.piper/templates/error.log'))
                .make([
                    stepName: stepName,
                    stepParameters: stepParameters?.toString(),
                    error: error
                ]).toString()
        writeErrorToInfluxData(parameters, error)
        throw error
    } finally {
        if (verbose)
            message += "--- End library step of: ${stepName} ---"
        echo message
    }
}

private void writeErrorToInfluxData(config, error){
    def script = config?.stepParameters?.script

    if(script && script.commonPipelineEnvironment?.getInfluxCustomDataMapTags().build_error_message == null){
        script.commonPipelineEnvironment?.setInfluxCustomDataMapTagsEntry('pipeline_data', 'build_error_step', config.stepName)
        script.commonPipelineEnvironment?.setInfluxCustomDataMapTagsEntry('pipeline_data', 'build_error_stage', script.env?.STAGE_NAME)
        script.commonPipelineEnvironment?.setInfluxCustomDataMapEntry('pipeline_data', 'build_error_message', error.getMessage())
    }
}
