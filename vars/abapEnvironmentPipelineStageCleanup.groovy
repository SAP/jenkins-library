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
 * This stage cleans up the ABAP Environment Pipeline run
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    stageName = stageName.replace('Declarative: ', '')
    System.out.println(stageName)
    piperStageWrapper (script: script, stageName: stageName, stageLocking: false) {
        if(parameters.script.commonPipelineEnvironment.configuration.runStage?.get("Prepare System")) {
            cloudFoundryDeleteService script: parameters.script
        }
    }

}
