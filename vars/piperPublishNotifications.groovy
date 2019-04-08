import com.sap.icd.jenkins.Utils
import com.sap.piper.JenkinsUtils

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    handleStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // report to SWA
        utils.pushToSWA([
            folder: script.globalPipelineEnvironment.getGithubOrg(),
            repository: script.globalPipelineEnvironment.getGithubRepo(),
            step: STEP_NAME
        ])

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
