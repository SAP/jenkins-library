def call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: 'pipelineStashFiles', stepParameters: parameters) {

        pipelineStashFilesBeforeBuild(parameters)
        body() //execute build
        pipelineStashFilesAfterBuild(parameters)
    }
}
