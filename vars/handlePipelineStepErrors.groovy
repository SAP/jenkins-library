import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field STEP_NAME = 'handlePipelineStepErrors'

void call(Map parameters = [:], body) {
    def stepParameters = parameters.stepParameters //mandatory
    def stepName = parameters.stepName //mandatory
    def verbose = parameters.get('echoDetails', true)
    def message = ''
    try {
        if (stepParameters == null && stepName == null)
            error "step handlePipelineStepErrors requires following mandatory parameters: stepParameters, stepName"

        if (verbose)
            echo "--- BEGIN LIBRARY STEP: ${stepName} ---"

        body()
    } catch (Throwable err) {
        if (verbose)
            message += SimpleTemplateEngine.newInstance()
                .createTemplate(libraryResource('com.sap.piper/templates/error.log'))
                .make([
                    stepName: stepName,
                    stepParameters: stepParameters?.toString(),
                    error: err
                ]).toString()
        throw err
    } finally {
        if (verbose)
            message += "--- END LIBRARY STEP: ${stepName} ---"
        echo message
    }
}
