def call(Map closures, script) {
    handleStepErrors(stepName: 'runClosures', stepParameters: [script: script]) {
        if (isFeatureActive(script: script, feature: 'parallelTestExecution')) {
            parallel closures
        } else {
            def closuresToRun = closures.values().asList()
            for (int i = 0; i < closuresToRun.size(); i++) {
                (closuresToRun[i] as Closure).call()
            }
        }
    }
}
