void call(body, stageName, config1, config2) {
    echo "Stage Name: ${stageName}"
    echo "Config 1: ${config1}"
    body()
    echo "Config 2: ${config2}"
}
