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
    'abapEnvironmentCloneGitRepo',
    /** Specifies the strategy that should be peformed on the ABAP Environment instance*/
    'strategy'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage clones Git repositories / software components to the ABAP Environment instance and checks out the master branch
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('strategy', parameters.script.commonPipelineEnvironment.configuration.runStage?.get(stageName))
        .use()

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {

        if(parameters.script.commonPipelineEnvironment.getStepConfiguration('abapEnvironmentPullGitRepo', 'Clone Repositories').repositoryNames || parameters.script.commonPipelineEnvironment.configuration.runStep?.get('abapEnvironmentPullGitRepo') && (!config.strategy)) {
            abapEnvironmentPullGitRepo script: parameters.script
        } 

        switch (config.strategy) {
            case 'Pull':
                abapEnvironmentPullGitRepo script: parameters.script
                break
            case 'Clone':
                abapEnvironmentCloneGitRepo script: parameters.script
                break
            case 'CheckoutPull':
                abapEnvironmentCheckoutBranch script: parameters.script
                abapEnvironmentPullGitRepo script: parameters.script
                break
            case 'PullCheckoutPull':
                abapEnvironmentPullGitRepo script: parameters.script
                abapEnvironmentCheckoutBranch script: parameters.script
                abapEnvironmentPullGitRepo script: parameters.script
                break
            default:
                abapEnvironmentPullGitRepo script: parameters.script
                break
        }

        /*if (parameters.script.commonPipelineEnvironment.configuration.runStage?.get("Prepare System")) {
            abapEnvironmentCloneGitRepo script: parameters.script
        }
        if(parameters.script.commonPipelineEnvironment.configuration.runStep?.get('abapEnvironmentCloneGitRepo')) {
            abapEnvironmentCloneGitRepo script: parameters.script
        }
        if(parameters.script.commonPipelineEnvironment.configuration.runStep?.get('abapEnvironmentCheckoutBranch')) {
            abapEnvironmentCheckoutBranch script: parameters.script
        }*/
    }

}
