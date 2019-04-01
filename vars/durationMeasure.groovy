import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.analytics.InfluxData

import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

def call(Map parameters = [:], body) {

    def script = checkScript(this, parameters)

    def measurementName = parameters.get('measurementName', 'test_duration')

    //start measurement
    def start = System.currentTimeMillis()

    body()

    //record measurement
    def duration = System.currentTimeMillis() - start

    InfluxData.addField('pipeline_data', measurementName, duration)

    return duration
}

