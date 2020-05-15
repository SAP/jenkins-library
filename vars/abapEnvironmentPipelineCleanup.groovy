import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

/**
 * This stage cleans up the ABAP Environment Pipeline run
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    echo "--------------------------------DELETE SYSTEM--------------------------------------"
    if (script.commonPipelineEnvironment.configuration.runStage?.get("Prepare System") == true) {
        cloudFoundryDeleteService script: parameters.script
    }
}
