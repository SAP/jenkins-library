import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

@Field String STEP_NAME = 'pipelineStashFilesAfterBuild'
@Field Set STEP_CONFIG_KEYS = ['runCheckmarx', 'stashIncludes', 'stashExcludes']
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, stepNameDoc: 'stashFiles') {
        def utils = parameters.juStabUtils
        if (utils == null) {
            utils = new Utils()
        }

        def script = checkScript(this, parameters)
        if (script == null)
            script = this

        //additional includes via passing e.g. stashIncludes: [opa5: '**/*.include']
        //additional excludes via passing e.g. stashExcludes: [opa5: '**/*.exclude']

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin([
                runCheckmarx: (script.commonPipelineEnvironment.configuration?.steps?.executeCheckmarxScan?.checkmarxProject != null && script.commonPipelineEnvironment.configuration.steps.executeCheckmarxScan.checkmarxProject.length()>0)
            ])
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME,
                               stepParam1: parameters?.script == null], config)

        // store files to be checked with checkmarx
        if (config.runCheckmarx) {
            utils.stash(
                'checkmarx',
                config.stashIncludes.checkmarx,
                config.stashExcludes.checkmarx
            )
        }

        utils.stashWithMessage(
            'classFiles',
            '[${STEP_NAME}] Failed to stash class files.',
            config.stashIncludes.classFiles,
            config.stashExcludes.classFiles
        )

        utils.stashWithMessage(
            'sonar',
            '[${STEP_NAME}] Failed to stash sonar files.',
            config.stashIncludes.sonar,
            config.stashExcludes.sonar
        )
    }
}
