void call(Map params) {
    echo "Stage Name: ${params.stageName}"
    echo "Config: ${params.config}"
    echo "Not calling ${params.stageName}"
    echo "Branch: ${params.script.commonPipelineEnvironment.gitBranch}"
}
return this
