void call(Closure originalStage, String stageName, Map configuration, Map params) {
    echo "Stage Name: $stageName"
    echo "Config: $configuration"
    originalStage()
    echo "Branch: ${params.script.commonPipelineEnvironment.gitBranch}"
}
return this
