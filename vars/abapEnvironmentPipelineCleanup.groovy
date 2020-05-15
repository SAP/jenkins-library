import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

/**
 * This stage cleans up the ABAP Environment Pipeline run
 */
@GenerateStageDocumentation(defaultStageName = 'Init')
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    // if (script.commonPipelineEnvironment.configuration.runStage?.get("Prepare System") == true) {
        cloudFoundryDeleteService script: parameters.script
    // }
}
