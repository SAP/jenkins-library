import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = []
@Field Set STEP_CONFIG_KEYS = []
@Field Set PARAMETER_KEYS = []
/**
 * This stage clones Git repositories / software components to the ABAP Environment instance
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    def stageName = parameters.stageName?:env.STAGE_NAME

    abapEnvironmentPullGitRepo script: parameters.script

}
