void call(body, stageName, config) {
    echo "Stage Name: ${stageName}"
    echo "Config: ${config}"
    body()
}
return this
