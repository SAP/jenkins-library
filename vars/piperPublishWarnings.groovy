import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'parserId',
    'parserName',
    'parserPattern',
    'parserScript',
    'recordIssuesSettings'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([])

/***/
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {

        //TODO: ensure warnings-ng plugin is installed.

        final script = checkScript(this, parameters) ?: this

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()
        // report to SWA
        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        // add Piper Notifications parser to config if missing
        if(new JenkinsUtils().addWarningsNGParser(
            configuration.parserId,
            configuration.parserName,
            configuration.parserPattern,
            configuration.parserScript
        )){
            echo "[${STEP_NAME}] Added warnings-ng plugin parser '${configuration.parserName}' configuration."
        }

        writeFile file: 'buildlog', text: JenkinsUtils.getFullBuildLog(script.currentBuild)
        // parse log for Piper Notifications
        recordIssues(
            configuration.recordIssuesSettings.plus([
                tools: [groovyScript(
                    parserId: configuration.parserId,
                    pattern: 'buildlog'
                )]
            ])
        )
    }
}
