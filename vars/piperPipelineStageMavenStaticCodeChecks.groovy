import com.sap.piper.ConfigurationLoader
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.QualityCheck
import com.sap.piper.ReportAggregator

import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Maven modules which should be excluded by the static code checks. By default the modules 'unit-tests' and 'integration-tests' will be excluded. */
    'mavenModulesExcludes',
    /**Path to a filter file with bug definitions which should be excluded. */
    'spotBugsExcludeFilterFile',
    /**Path to a filter file with bug definitions which should be included. */
    'spotBugsIncludeFilterFile',
    /**The maximum number of failures allowed before execution fails. */
    'spotBugsMaxAllowedViolations',
    /**What priority level to fail the build on. PMD violations are assigned a priority from 1 (most severe) to 5 (least severe) according the the rule's priority. Violations at or less than this priority level are considered failures and will fail the build if failOnViolation=true and the count exceeds maxAllowedViolations. The other violations will be regarded as warnings and will be displayed in the build output if verbose=true. Setting a value of 5 will treat all violations as failures, which may cause the build to fail. Setting a value of 1 will treat all violations as warnings. Only values from 1 to 5 are valid. */
    'pmdFailurePriority',
    /**The maximum number of failures allowed before execution fails. Used in conjunction with failOnViolation=true and utilizes failurePriority. This value has no meaning if failOnViolation=false. If the number of failures is greater than this number, the build will be failed. If the number of failures is less than or equal to this value, then the build will not be failed. Defaults to 5. */
    'pmdMaxAllowedViolations'
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

        String spotBugsIncludeFilterFile = 'default_spotbugs_include_filter.xml'
        String spotBugsLocalIncludeFilterPath = ".pipeline/${spotBugsIncludeFilterFile}"
        writeFile file: spotBugsLocalIncludeFilterPath, text: libraryResource(spotBugsIncludeFilterFile)

        try {
            mavenExecuteStaticCodeChecks(script: script, spotBugsIncludeFilterFile: spotBugsLocalIncludeFilterPath)
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
    }
    ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.PmdCheck)
}
