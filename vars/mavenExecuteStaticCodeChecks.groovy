import com.sap.piper.ConfigurationLoader
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.QualityCheck
import com.sap.piper.ReportAggregator
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String METADATA_FILE = 'metadata/mavenStaticCodeChecks.yaml'
@Field String STEP_NAME = getClass().getName()


@Field Set STEP_CONFIG_KEYS = [
    /**Parameter to turn off SpotBugs.  */
    'spotBugs',
    /** Parameter to turn off PMD.*/
    'pmd',
    /** Maven modules which should be excluded by the static code checks. By default the modules 'unit-tests' and 'integration-tests' will be excluded.*/
    'mavenModulesExcludes',
    /** Path to a filter file with bug definitions which should be excluded.*/
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

/**
 * Executes static code checks for maven based projects.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: null
    List credentials = []
    parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(script, parameters)

    try {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
    } catch (Exception exception) {
        error("Maven Static Code Checks execution failed. Please examine the reports which are also available in the Jenkins user interface.")
    }
    finally {
        showIssues(script)
    }
}

private showIssues(Script script) {
    Map configuration = ConfigurationLoader.stepConfiguration(script, STEP_NAME)
    // Every check is executed by default. Only if configured with `false` the check won't be executed
    if (!(configuration.spotBugs == false)) {
        recordIssues(blameDisabled: true,
            enabledForFailure: true,
            aggregatingResults: false,
            tool: spotBugs(pattern: '**/target/spotbugsXml.xml'))

        ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.FindbugsCheck)
    }
    if (!(configuration.pmd == false)) {
        recordIssues(blameDisabled: true,
            enabledForFailure: true,
            aggregatingResults: false,
            tool: pmdParser(pattern: '**/target/pmd.xml'))
        ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.PmdCheck)
    }
}
