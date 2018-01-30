def call(Map parameters = [:], body) {

    def script = parameters.script
    def measurementName = parameters.get('measurementName', 'test_duration')

    //start measurement
    def start = System.currentTimeMillis()

    body()

    //record measurement
    def duration = System.currentTimeMillis() - start

    if (script != null)
        script.commonPipelineEnvironment.setPipelineMeasurement(measurementName, duration)

    return duration
}

