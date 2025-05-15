import com.sap.piper.ConfigurationHelper
import com.sap.piper.DebugReport
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    /**
     * Flag to control whether potentially confidential information will be included in the
     * debug_report.txt. Default value is `false`. Additional information written to the log
     * when this flag is `true` includes MTA modules, NPM modules, the GitHub repository and
     * branch, the global extension repository if used, a shared config file path, and all
     * used global and local shared libraries.
     * @possibleValues `true`, `false`
     */
    'shareConfidentialInformation'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS + [
    /**
     * Flag to enable printing the generated debug_report.txt also to the console.
     */
    'printToConsole'
]
/**
 * Archives the debug_report.txt artifact which facilitates analyzing pipeline errors by collecting
 * information about the Jenkins environment in which the pipeline was run. There is a single
 * config option 'shareConfidentialInformation' to enable including (possibly) confidential
 * information in the debug report, which could be helpful depending on the specific error.
 * By default this information is not included.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    try {
        String stageName = parameters.stageName ?: env.STAGE_NAME
        // ease handling extension
        stageName = stageName?.replace('Declarative: ', '')
        def utils = parameters.juStabUtils ?: new Utils()

        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        boolean shareConfidentialInformation = configuration?.get('shareConfidentialInformation') ?: false

        Map result = DebugReport.instance.generateReport(script, shareConfidentialInformation)

        if (parameters.printToConsole) {
            echo result.contents
        }

        script.writeFile file: result.fileName, text: result.contents
        script.archiveArtifacts artifacts: result.fileName
        echo "Successfully archived debug report as '${result.fileName}'"
    } catch (Exception e) {
        println("WARNING: The debug report was not created, it threw the following error message:")
        println("${e}")
    }
}
