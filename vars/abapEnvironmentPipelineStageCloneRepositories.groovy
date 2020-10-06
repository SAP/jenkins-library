import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Pulls Software Components / Git repositories into the ABAP Environment instance */
    'abapEnvironmentPullGitRepo',
    /** Checks out a Branch in the pulled Software Component on the ABAP Environment instance */
    'abapEnvironmentCheckoutBranch',
    /** Clones Software Components / Git repositories into the ABAP Environment instance and checks out the respective branches */
    'abapEnvironmentCloneGitRepo'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage clones Git repositories / software components to the ABAP Environment instance and checks out the master branch
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        /*if (parameters.script.commonPipelineEnvironment.configuration.runStage?.get("Prepare System")) {
            abapEnvironmentCloneGitRepo script: parameters.script
        }*/
        if(parameters.script.commonPipelineEnvironment.getStepConfiguration('abapEnvironmentPullGitRepo', 'Clone Repositories').repositoryNames || parameters.script.commonPipelineEnvironment.configuration.runStep?.get('abapEnvironmentPullGitRepo')?) {
            abapEnvironmentPullGitRepo script: parameters.script
        } 
        if(parameters.script.commonPipelineEnvironment.configuration.runStep?.get('abapEnvironmentCloneGitRepo')?) {
            abapEnvironmentCloneGitRepo script: parameters.script
        }
        if(parameters.script.commonPipelineEnvironment.configuration.runStep?.get('abapEnvironmentCheckoutBranch')?) {
            abapEnvironmentCheckoutBranch script: parameters.script
        }
    }

}
