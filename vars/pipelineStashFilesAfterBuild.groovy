import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

@Field String STEP_NAME = 'pipelineStashFilesAfterBuild'
@Field Set STEP_CONFIG_KEYS = ['runCheckmarx', 'stashIncludes', 'stashExcludes']
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, stepNameDoc: 'stashFiles') {
        def utils = parameters.juStabUtils
        if (utils == null) {
            utils = new Utils()
        }
        def script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]

        //additional includes via passing e.g. stashIncludes: [opa5: '**/*.include']
        //additional excludes via passing e.g. stashExcludes: [opa5: '**/*.exclude']

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('runCheckmarx', (script.commonPipelineEnvironment.configuration?.steps?.executeCheckmarxScan?.checkmarxProject != null && script.commonPipelineEnvironment.configuration.steps.executeCheckmarxScan.checkmarxProject.length()>0))
            .use()

        // store files to be checked with checkmarx
        if (config.runCheckmarx) {
            utils.stash('checkmarx', config.stashIncludes?.get('checkmarx')?config.stashIncludes.checkmarx:'**/*.js, **/*.scala, **/*.py, **/*.go, **/*.xml, **/*.html', config.stashExcludes?.get('checkmarx')?config.stashExcludes.checkmarx:'**/*.mockserver.js, node_modules/**/*.js')
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
