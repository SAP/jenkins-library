import groovy.transform.Field
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    'abapEnvironmentAssembleConfirm'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage builds an AddOn for the SAP Cloud Platform ABAP Environment
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        abapEnvironmentAssembleConfirm script: parameters.script
    }

}
