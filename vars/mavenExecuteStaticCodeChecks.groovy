import com.sap.piper.ConfigurationLoader
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.QualityCheck
import com.sap.piper.ReportAggregator
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String METADATA_FILE = 'metadata/mavenStaticCodeChecks.yaml'
@Field String STEP_NAME = getClass().getName()

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

