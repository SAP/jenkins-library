import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = [
    /**
     * By default certain files are excluded from stashing (e.g. `.git` folder).
     * Details can be found as per [Pipeline basic step `stash](https://jenkins.io/doc/pipeline/steps/workflow-basic-steps/#stash-stash-some-files-to-be-used-later-in-the-build).
     * This parameter allows to provide a list of stash names for which the standard exclude behavior should be switched off.
     * This will allow you to also stash directories like `.git`.
     */
    'noDefaultExludes',
    /** @see pipelineStashFiles */
    'stashIncludes',
    /** @see pipelineStashFiles */
    'stashExcludes'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step stashes files that are needed in other build steps (on other nodes).
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, stepNameDoc: 'stashFiles') {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        //additional includes via passing e.g. stashIncludes: [opa5: '**/*.include']
        //additional excludes via passing e.g. stashExcludes: [opa5: '**/*.exclude']

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        config.stashIncludes.each {stashKey, stashIncludes ->
            def useDefaultExcludes = !config.noDefaultExludes.contains(stashKey)
            utils.stashWithMessage(stashKey, "[${STEP_NAME}] no files detected for stash '${stashKey}': ", stashIncludes, config.stashExcludes[stashKey]?:'', useDefaultExcludes)
        }
    }
}
