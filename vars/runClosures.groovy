void call(script, Map closures, Boolean parallelExecution, String label = "closures") {
    handlePipelineStepErrors(stepName: 'runClosures', stepParameters: [script: script]) {
        echo "Executing $label"
        if (parallelExecution) {
            echo "Executing $label in parallel"
            parallel closures
        } else {
            echo "Executing $label in sequence"
            def closuresToRun = closures.values().asList()
            for (int i = 0; i < closuresToRun.size(); i++) {
                (closuresToRun[i] as Closure)()
            }
        }
    }
}
