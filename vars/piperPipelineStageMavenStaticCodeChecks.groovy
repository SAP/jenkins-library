import com.sap.piper.ConfigurationLoader
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.QualityCheck
import com.sap.piper.ReportAggregator

import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Execute static code checks for Maven based projects. This stage enforces SAP Cloud SDK specific PND rulesets as well as SpotBugs include filter.  */
    'mavenExecuteStaticCodeChecks'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Execute static code checks for Maven based projects. This stage enforces SAP Cloud SDK specific PND rulesets as well as SpotBugs include filter.
 *
 */
@GenerateStageDocumentation(defaultStageName = 'mavenExecuteStaticCodeChecks')
void call(Map parameters = [:]) {
    String stageName = 'mavenExecuteStaticCodeChecks'
    final script = checkScript(this, parameters) ?: null

    piperStageWrapper(stageName: stageName, script: script) {
        try {
            mavenExecuteStaticCodeChecks(script: script)
        } catch (Exception exception) {
            error("Maven Static Code Checks execution failed. Please examine the reports which are also available in the Jenkins user interface.")
        }
        finally {
            showIssues(script, stageName)
        }
    }
}

private showIssues(Script script, String stageName) {
    Map configuration = ConfigurationLoader.stageConfiguration(script, stageName)
    // the checks are executed by default, even if they are not configured. They aren't executed only in case they are turned off with `false`
    if (configuration.mavenExecuteStaticCodeChecks?.spotBugs == null || configuration.mavenExecuteStaticCodeChecks?.spotBugs == true) {
        recordIssues(blameDisabled: true,
            enabledForFailure: true,
            aggregatingResults: false,
            tool: spotBugs(pattern: '**/target/spotbugsXml.xml'))

        ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.FindbugsCheck)
    }
    if (configuration.mavenExecuteStaticCodeChecks?.pmd == null || configuration.mavenExecuteStaticCodeChecks?.pmd == true) {
        recordIssues(blameDisabled: true,
            enabledForFailure: true,
            aggregatingResults: false,
            tool: pmdParser(pattern: '**/target/pmd.xml'))
        ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.PmdCheck)
    }
}
