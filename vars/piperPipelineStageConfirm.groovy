import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.StageNameProvider
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'confirm'

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
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    String unstableStepNames = script.commonPipelineEnvironment.getValue('unstableSteps') ? "${script.commonPipelineEnvironment.getValue('unstableSteps').join(', ')}" : ''

    boolean approval = false
    def userInput

    milestone()

    timeout(
        unit: 'HOURS',
        time: config.manualConfirmationTimeout
    ){
        if (currentBuild.result == 'UNSTABLE') {
            def minReasonLength = 10
            def acknowledgementText = 'I acknowledge that for traceability purposes the approval reason is stored together with my user name / user id'
            def reasonDescription = "Please provide a reason for overruling the failed steps ${unstableStepNames}, with ${minReasonLength} characters or more:".toString()
            def acknowledgementDescription = "${acknowledgementText}:".toString()
            while(!approval) {
                userInput = input(
                    message: 'Approve continuation of pipeline, although some steps failed.',
                    ok: 'Approve',
                    parameters: [
                        text(
                            defaultValue: '',
                            description: reasonDescription,
                            name: 'reason'
                        ),
                        booleanParam(
                            defaultValue: false,
                            description: acknowledgementDescription,
                            name: 'acknowledgement'
                        )
                    ]
                )
                approval = validateApproval(userInput.reason, minReasonLength, userInput.acknowledgement, acknowledgementText, unstableStepNames)
            }
        } else {
            input message: config.manualConfirmationMessage
        }

    }

}

private boolean validateApproval(reason, minReasonLength, acknowledgement, acknowledgementText, unstableStepNames) {
    def reasonIsLongEnough = reason?.length() >= minReasonLength
    approved = acknowledgement && reasonIsLongEnough
    if (approved) {
        echo "Failed steps\n------------\n${unstableStepNames}"
        echo "Reason\n------\n${reason}"
        echo "Acknowledgement\n---------------\n☑ ${acknowledgementText}"
    } else {
        if (!acknowledgement) {
            echo "Rejected the approval because the user didn't acknowledge that his user name or id is logged"
        }
        if (!reasonIsLongEnough) {
            echo "Rejected the approval because the provided reason has less than ${minReasonLength} characters"
        }
    }
    return approved
}
