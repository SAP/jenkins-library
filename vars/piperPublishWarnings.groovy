import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field List PLUGIN_ID_LIST = ['warnings-ng']

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * The id of the Groovy script parser. If the id is not present in the current Jenkins configuration it is created.
     */
    'parserId',
    /**
     * The display name for the warnings parsed by the parser.
     * Only considered if a new parser is created.
     */
    'parserName',
    /**
     * The pattern used to parse the log file.
     * Only considered if a new parser is created.
     */
    'parserPattern',
    /**
     * The script used to parse the matches produced by the pattern into issues.
     * Only considered if a new parser is created.
     * see https://github.com/jenkinsci/analysis-model/blob/master/src/main/java/edu/hm/hafner/analysis/IssueBuilder.java
     */
    'parserScript',
    /**
     * Settings that are passed to the recordIssues step of the warnings-ng plugin.
     * see https://github.com/jenkinsci/warnings-ng-plugin/blob/master/doc/Documentation.md#configuration
     */
    'recordIssuesSettings'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([])

/**
 * This step scans the current build log for messages produces by the Piper library steps and publishes them on the Jenkins job run as *Piper warnings* via the warnings-ng plugin.
 *
 * The default parser detects log entries with the following pattern: `[<SEVERITY>] <MESSAGE> (<LIBRARY>/<STEP>)`
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {

        final script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        for(String id : PLUGIN_ID_LIST){
            if (!JenkinsUtils.isPluginActive(id)) {
                error("[ERROR][${STEP_NAME}] The step requires the plugin '${id}' to be installed and activated in the Jenkins.")
            }
        }

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        // add Piper Notifications parser to config if missing
        if(new JenkinsUtils().addWarningsNGParser(
            configuration.parserId,
            configuration.parserName,
            configuration.parserPattern,
            configuration.parserScript
        )){
            echo "[${STEP_NAME}] Added warnings-ng plugin parser '${configuration.parserName}' configuration."
        }

        writeFile file: 'build.log', text: JenkinsUtils.getFullBuildLog(script.currentBuild)
        // parse log for Piper Notifications
        recordIssues(
            configuration.recordIssuesSettings.plus([
                tools: [groovyScript(
                    parserId: configuration.parserId,
                    pattern: 'build.log'
                )]
            ])
        )
    }
}
