import com.sap.piper.ConfigurationLoader
import com.sap.piper.QualityCheck
import com.sap.piper.ReportAggregator

import static com.sap.piper.Prerequisites.checkScript

def call(Map parameters = [:]) {
    final String stageName = 'staticCodeChecks'
    final script = checkScript(this, parameters) ?: null

    piperStageWrapper(stageName: stageName, script: script) {

        String spotBugsIncludeFilterFile = 'default_spotbugs_include_filter.xml'
        String spotBugsLocalIncludeFilerPath = ".pipeline/${spotBugsIncludeFilterFile}"
        writeFile file: spotBugsLocalIncludeFilerPath, text: libraryResource(spotBugsIncludeFilterFile)

        String defaultPmdRulesFile = 'default_pmd_rulesets.xml'
        String pmdRulesPath = ".pipeline/${defaultPmdRulesFile}"
        writeFile file: pmdRulesPath, text: libraryResource(defaultPmdRulesFile)

        mavenExecuteStaticCodeChecks(script: script,
            spotBugsIncludeFilterFile: spotBugsLocalIncludeFilerPath,
            pmdRuleSets: [pmdRulesPath])

        executeWithLockedCurrentBuildResult(
            script: script,
            errorStatus: 'FAILURE',
            errorHandler: script.buildFailureReason.setFailureReason,
            errorHandlerParameter: 'Maven Static Code Checks',
            errorMessage: "Please examine the SpotBugs/PMD reports."
        ) {
            recordIssues failedTotalHigh: 1,
                failedTotalNormal: 10,
                blameDisabled: true,
                enabledForFailure: true,
                aggregatingResults: false,
                tool: spotBugs(pattern: '**/target/spotbugsXml.xml')

            recordIssues failedTotalHigh: 1,
                failedTotalNormal: 10,
                blameDisabled: true,
                enabledForFailure: true,
                aggregatingResults: false,
                tool: pmdParser(pattern: '**/target/pmd.xml')
        }

        Map configuration = ConfigurationLoader.stageConfiguration(script, stageName)
        // the checks are executed by default, even if they are not configured. They aren't executed only in case they are turned off with `false`
        if (configuration.mavenExecuteStaticCodeChecks?.spotBugs == null || configuration.mavenExecuteStaticCodeChecks?.spotBugs == true) {
            ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.FindbugsCheck)
        }
        if (configuration.mavenExecuteStaticCodeChecks?.pmd == null || configuration.mavenExecuteStaticCodeChecks?.pmd == true) {
            ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.PmdCheck)
        }
    }
}
