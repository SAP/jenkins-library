import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.ReportAggregator
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STAGE_STEP_KEYS = []
@Field Set STAGE_CONFIG_KEYS = STAGE_STEP_KEYS
@Field Set PARAMETER_KEYS = STAGE_CONFIG_KEYS

/**
 * This stage is responsible for releasing/deploying artifacts to a Nexus Repository Manager.<br />
 */
@GenerateStageDocumentation(defaultStageName = 'artifactDeployment')
void call(Map parameters = [:]) {
    String stageName = 'artifactDeployment'
    final script = checkScript(this, parameters) ?: this

    piperStageWrapper(stageName: stageName, script: script) {

        def commonPipelineEnvironment = script.commonPipelineEnvironment
        List unstableSteps = commonPipelineEnvironment?.getValue('unstableSteps') ?: []
        if (unstableSteps) {
            piperPipelineStageConfirm script: script
            unstableSteps = []
            commonPipelineEnvironment.setValue('unstableSteps', unstableSteps)
        }

        withEnv(["STAGE_NAME=${stageName}"]) {
            nexusUpload(script: script)
        }

        ReportAggregator.instance.reportDeploymentToNexus()
    }
}
