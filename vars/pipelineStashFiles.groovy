import groovy.transform.Field

@Field STEP_NAME = 'pipelineStashFiles'

void call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: 'pipelineStashFiles', stepParameters: parameters) {

        pipelineStashFilesBeforeBuild(parameters)
        body() //execute build
        pipelineStashFilesAfterBuild(parameters)
    }
}
