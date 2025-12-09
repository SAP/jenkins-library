import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/hadolintExecute.yaml'

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'qualityGates',
    'reportFile',
    'reportName'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: null
    List credentialInfo = [
        [type: 'usernamePassword', id: 'configurationCredentialsId', env: ['PIPER_configurationUsername', 'PIPER_configurationPassword']],
    ]

    issuesWrapper(parameters, script){
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentialInfo)
    }
}

def issuesWrapper(Map parameters = [:], Script script, Closure body){
    String stageName = parameters.stageName ?: env.STAGE_NAME
    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults([:], stageName)
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    try {
        body()
    } finally {
        try {
            recordIssues(
                skipBlames: true,
                enabledForFailure: true,
                aggregatingResults: false,
                qualityGates: config.qualityGates,
                tool: checkStyle(
                    id: config.reportName,
                    name: config.reportName,
                    pattern: config.reportFile
                )
            )
        } catch (e) {
            echo "recordIssues has failed. Possibly due to an outdated version of the warnings-ng plugin."
            e.printStackTrace()
        }
    }
}
