import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

/**
 * This stage clones Git repositories / software components to the ABAP Environment instance
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName) {

        abapEnvironmentPullGitRepo script: parameters.script

    }
}
