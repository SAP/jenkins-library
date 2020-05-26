import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Starts an ATC check run on the ABAP Environment instance */
    'abapEnvironmentRunATCCheck'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage runs the ATC Checks
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    def stageName = parameters.stageName?:env.STAGE_NAME
    piperStageWrapper (script: script, stageName: stageName) {

        abapEnvironmentRunATCCheck script: parameters.script

        def atcResult = readFile file: "ATCResults.xml"
        if (atcResult != '<?xml version="1.0" encoding="utf-8"?><checkstyle version="1.0"/>') {
            unstable('ATC Issues detected - setting build status to UNSTABLE')
        }
    }
}
