import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.ReportAggregator
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STAGE_STEP_KEYS = []
@Field Set STAGE_CONFIG_KEYS = STAGE_STEP_KEYS.plus([
    /** Parameters for deployment to a Nexus Repository Manager. */
    'nexus'
])
@Field Set PARAMETER_KEYS = STAGE_CONFIG_KEYS

/**
 * This stage is responsible for releasing/deploying artifacts to a Nexus Repository Manager.<br />
 */
@GenerateStageDocumentation(defaultStageName = 'artifactDeployment')
void call(Map parameters = [:]) {
    String stageName = 'artifactDeployment'
    final script = checkScript(this, parameters) ?: this

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STAGE_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .withMandatoryProperty('nexus')
        .use()

    piperStageWrapper(stageName: stageName, script: script) {

        def commonPipelineEnvironment = script.commonPipelineEnvironment
        List unstableSteps = commonPipelineEnvironment?.getValue('unstableSteps') ?: []
        if (unstableSteps) {
            piperPipelineStageConfirm script: script
            unstableSteps = []
            commonPipelineEnvironment.setValue('unstableSteps', unstableSteps)
        }

        Map nexusConfig = config.nexus as Map

        // Pull additionalClassifiers param from resolved config here for legacy compatibility.
        // The parameter will become obsolete soon.
        Map nexusUploadParams = [
            script: script,
            additionalClassifiers: nexusConfig.additionalClassifiers,
        ]

        withEnv(["STAGE_NAME=${stageName}"]) {
            nexusUpload(nexusUploadParams)
        }

        ReportAggregator.instance.reportDeploymentToNexus()
    }
}
