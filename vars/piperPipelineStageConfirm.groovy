import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Specifies if a manual confirmation is active before running the __Promote__ and __Release__ stages of the pipeline.
     * @possibleValues `true`, `false`
     */
    'manualConfirmation',
    /** Defines message displayed as default manual confirmation. Please note: only used in case pipeline is in state __SUCCESSFUL__ */
    'manualConfirmationMessage',
    /** Defines how many hours a manual confirmation is possible for a dedicated pipeline. */
    'manualConfirmationTimeout'

]
@Field STAGE_STEP_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage a manual confirmation is requested before processing subsequent stages like __Promote__ and __Release__.
 *
 * This stage will be active in two scenarios:
 * - manual activation of this stage
 * - in case of an 'UNSTABLE' build (even when manual confirmation is inactive)
 */
@GenerateStageDocumentation(defaultStageName = 'Confirm')
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this, script)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    String unstableStepNames = script.commonPipelineEnvironment.getValue('unstableSteps') ? "${script.commonPipelineEnvironment.getValue('unstableSteps').join(':\n------\n')}:" : ''

    boolean approval = false
    def userInput

    timeout(
        unit: 'HOURS',
        time: config.manualConfirmationTimeout
    ){
        if (currentBuild.result == 'UNSTABLE') {
            while(!approval) {
                userInput = input(
                    message: 'Approve continuation of pipeline, although some steps failed.',
                    ok: 'Approve',
                    parameters: [
                        text(
                            defaultValue: unstableStepNames,
                            description: 'Please provide a reason for overruling following failed steps:',
                            name: 'reason'
                        ),
                        booleanParam(
                            defaultValue: false,
                            description: 'I acknowledge that for traceability purposes the approval reason is stored together with my user name / user id:',
                            name: 'acknowledgement'
                        )
                    ]
                )
                approval = userInput.acknowledgement && userInput.reason?.length() > (unstableStepNames.length() + 10)
            }
            echo "Reason:\n-------------\n${userInput.reason}"
            echo "Acknowledged:\n-------------\n${userInput.acknowledgement}"
        } else {
            input message: config.manualConfirmationMessage
        }

    }

}
