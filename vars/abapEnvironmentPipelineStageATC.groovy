import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

/**
 * This stage runs the ATC Checks
 */
@GenerateStageDocumentation(defaultStageName = 'Init')
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    abapEnvironmentRunATCCheck script: parameters.script
    try {
        recordIssues(tools: [checkStyle(pattern: 'ATCResults.xml')])
    } catch (ex) {
        // if plugin not available - do nothing
    }

    def atcResult = readFile file: "ATCResults.xml"
    if (atcResult != "") {
        unstable('ATC Issues')
    }
}
