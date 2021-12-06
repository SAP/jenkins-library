import com.sap.piper.BuildTool
import com.sap.piper.ConfigurationLoader
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.QualityCheck
import com.sap.piper.ReportAggregator
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String METADATA_FILE = 'metadata/mavenExecuteStaticCodeChecks.yaml'
@Field String STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: null
    List credentials = []
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MAVEN)

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
