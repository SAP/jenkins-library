import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'parserId',
    'parserName',
    'parserPattern',
    'parserScript'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([])

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {

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
        if(JenkinsUtils.addWarningsNGParser(
            configuration.parserId,
            configuration.parserName,
            configuration.parserPattern,
            configuration.parserScript
        )){
            echo "[${STEP_NAME}] New Warnings-NG plugin parser '${configuration.parserName}' configuration added."
        }

        node(){
            try{
                writeFile file: 'buildlog', text: JenkinsUtils.getFullBuildLog(script.currentBuild)
                // parse log for Piper Notifications
                recordIssues(
                    blameDisabled: true,
                    enabledForFailure: true,
                    tools: [groovyScript(
                        parserId: configuration.parserId,
                        pattern: 'buildlog'
                    )]
                )
            }finally{
                deleteDir()
            }
        }
    }
}
