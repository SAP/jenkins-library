import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

void call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: 'pipelineStashFiles', stepParameters: parameters) {

        pipelineStashFilesBeforeBuild(parameters)
        body() //execute build
        pipelineStashFilesAfterBuild(parameters)
    }
}
