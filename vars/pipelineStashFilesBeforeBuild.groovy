import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field Set STEP_CONFIG_KEYS = ['noDefaultExludes', 'stashIncludes', 'stashExcludes']
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, stepNameDoc: 'stashFiles') {

        Utils utils = parameters.juStabUtils
        if (utils == null) {
            utils = new Utils()
        }

        def script = checkScript(this, parameters)
        if (script == null)
            script = this

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        config.stashIncludes.each {stashKey, stashIncludes ->
            def useDefaultExcludes = !config.noDefaultExludes.contains(stashKey)
            utils.stashWithMessage(stashKey, "[${STEP_NAME}] no files detected for stash '${stashKey}': ", stashIncludes, config.stashExcludes[stashKey]?:'', useDefaultExcludes)
        }
    }
}
