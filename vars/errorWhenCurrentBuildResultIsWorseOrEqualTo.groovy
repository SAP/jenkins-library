def call(Map parameters = [:]) {
    final ERROR_MESSAGE_PREFIX = "Build was ABORTED and marked as FAILURE. "

    handleStepErrors(stepName: 'errorWhenCurrentBuildResultIsWorseOrEqualTo', stepParameters: parameters) {
        def script = parameters.script
        def errorStatus = parameters.errorStatus
        def errorHandler = parameters.errorHandler
        def errorHandlerParameter = parameters.errorHandlerParameter
        def errorMessage = parameters.errorMessage ?: ''

        if (script.currentBuild.result && script.currentBuild.resultIsWorseOrEqualTo(errorStatus)) {
            if (errorHandler) {
                errorHandler(errorHandlerParameter)
            }
            error(ERROR_MESSAGE_PREFIX + errorMessage)
        }
    }
}
