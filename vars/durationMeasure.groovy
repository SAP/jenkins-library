import com.sap.piper.GenerateDocumentation
import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.analytics.InfluxData

import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = [
    /** Defines the name of the measurement which is written to the Influx database.*/
    'measurementName'
]

/**
 * This step is used to measure the duration of a set of steps, e.g. a certain stage.
 * The duration is stored in a Map. The measurement data can then be written to an Influx database using step [influxWriteData](influxWriteData.md).
 *
 * !!! tip
 *     Measuring for example the duration of pipeline stages helps to identify potential bottlenecks within the deployment pipeline.
 *     This then helps to counter identified issues with respective optimization measures, e.g parallelization of tests.
 */
@GenerateDocumentation
def call(Map parameters = [:], body) {

    def script = checkScript(this, parameters)

    def measurementName = parameters.get('measurementName', 'test_duration')

    //start measurement
    long start = System.currentTimeMillis()

    // execute the body, catch the potential exception
    echo "--- Begin durationMeasure for ${measurementName} ---"
    Throwable caught = null
    try {
        body()
        echo "body() for '${measurementName}' successfully executed"
    } catch(Throwable t) {
        echo "body() for '${measurementName}' executed throwing ${t}"
        caught = t
    }

    // calculate and store the duration
    long duration = System.currentTimeMillis() - start
    echo "--- End durationMeasure for ${measurementName} (${duration} ms) ---"
    InfluxData.addField('pipeline_data', measurementName, duration)

    // re-throw the caught Throwable if present
    if (caught != null) {
        throw caught
    }
    return duration
}
