import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([])
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

        Map piperNotificationsSettings = [
            parserName: 'Piper Notifications Parser',
            parserLinkName: 'Piper Notifications',
            parserTrendName: 'Piper Notifications',
            parserRegexp: '\\[(INFO|WARNING|ERROR)\\] (.*) \\(([^) ]*)\\/([^) ]*)\\)',
            parserExample: ''
        ]
        piperNotificationsSettings.parserScript = '''import hudson.plugins.warnings.parser.Warning
        import hudson.plugins.analysis.util.model.Priority

        Priority priority = Priority.LOW
        String message = matcher.group(2)
        String libraryName = matcher.group(3)
        String stepName = matcher.group(4)
        String fileName = 'Jenkinsfile'

        switch(matcher.group(1)){
            case 'WARNING': priority = Priority.NORMAL; break;
            case 'ERROR': priority = Priority.HIGH; break;
        }

        return new Warning(fileName, 0, libraryName, stepName, message, priority);
        '''

        // add Piper Notifications parser to config if missing
        if(JenkinsUtils.addWarningsParser(piperNotificationsSettings)){
            echo "[${STEP_NAME}] New Warnings plugin parser '${piperNotificationsSettings.parserName}' configuration added."
        }

        node(){
            try{
                // parse log for Piper Notifications
                warnings(canRunOnFailed: true, consoleParsers: [[ parserName: piperNotificationsSettings.parserName ]])
            }finally{
                deleteDir()
            }
        }
    }
}
