void call(Map params) {
    echo "Stage Name: ${params.stageName}"
    echo "Config: ${params.config}"
    params.originalStage()
    echo "Branch: ${params.script.commonPipelineEnvironment.gitBranch}"
}
return this
