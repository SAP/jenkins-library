void call(script, body, stageName, config) {
    echo "Stage Name: ${stageName}"
    echo "Config: ${config}"
    body()
    echo "Branch: ${script.commonPipelineEnvironment.gitBranch}"
}
return this
