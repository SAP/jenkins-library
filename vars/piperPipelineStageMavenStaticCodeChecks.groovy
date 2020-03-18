
import com.sap.cloud.sdk.s4hana.pipeline.QualityCheck
import com.sap.cloud.sdk.s4hana.pipeline.ReportAggregator

def call(Map parameters = [:]) {
    String stageName = 'staticCodeChecks'
    Script script = parameters.script

    piperStageWrapper(stageName: stageName, script: script) {

        String includeFilterFile = 's4hana_findbugs_include_filter.xml'
        String localIncludeFilerPath = "s4hana_pipeline/${includeFilterFile}"
        writeFile file: localIncludeFilerPath, text: libraryResource(includeFilterFile)

        mavenExecuteStaticCodeChecks(script: script,
            spotBugsIncludeFilterFile: localIncludeFilerPath,
            pmdRuleSets: 'rulesets/s4hana-qualities.xml',
            m2Path: s4SdkGlobals.m2Directory)

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
        if (configuration.mavenExecuteStaticCodeChecks?.spotBugs == null || configuration.mavenExecuteStaticCodeChecks?.spotBugs == true) {
            ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.FindbugsCheck)
        }
        if (configuration.mavenExecuteStaticCodeChecks?.pmd == null || configuration.mavenExecuteStaticCodeChecks?.pmd == true) {
            ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.PmdCheck)
        }
    }
}
