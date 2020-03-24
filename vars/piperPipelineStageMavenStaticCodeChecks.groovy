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
    /**A comma-separated list of exclusions (.java source files) expressed as an Ant-style pattern relative to the sources root folder, i.e. application/src/main/java for maven projects. */
    'pmdExcludes',
    /**The PMD rulesets to use. See the Stock Java Rulesets for a list of available rules. Defaults to a custom ruleset provided by this maven plugin. */
    'pmdRuleSets'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Execute static code checks for Maven based projects. The plugins SpotBugs and PMD are used.
 *
 */
@GenerateStageDocumentation(defaultStageName = 'staticCodeChecks')
void call(Map parameters = [:]) {
    String stageName = 'staticCodeChecks'//parameters.stageName?:env.STAGE_NAME
    final script = checkScript(this, parameters) ?: null

    piperStageWrapper(stageName: stageName, script: script) {

        String spotBugsIncludeFilterFile = 'default_spotbugs_include_filter.xml'
        String spotBugsLocalIncludeFilterPath = ".pipeline/${spotBugsIncludeFilterFile}"
        writeFile file: spotBugsLocalIncludeFilterPath, text: libraryResource(spotBugsIncludeFilterFile)

        String defaultPmdRulesFile = 'default_pmd_rulesets.xml'
        String pmdRulesPath = ".pipeline/${defaultPmdRulesFile}"
        writeFile file: pmdRulesPath, text: libraryResource(defaultPmdRulesFile)

        mavenExecuteStaticCodeChecks(script: script,
            spotBugsIncludeFilterFile: spotBugsLocalIncludeFilterPath,
            pmdRuleSets: [pmdRulesPath])

        Map configuration = ConfigurationLoader.stageConfiguration(script, stageName)
        // the checks are executed by default, even if they are not configured. They aren't executed only in case they are turned off with `false`
        if (configuration.mavenExecuteStaticCodeChecks?.spotBugs == null || configuration.mavenExecuteStaticCodeChecks?.spotBugs == true) {
            recordIssues(failedTotalHigh: 1,
                failedTotalNormal: 10,
                blameDisabled: true,
                enabledForFailure: true,
                aggregatingResults: false,
                tool: spotBugs(pattern: '**/target/spotbugsXml.xml'))

            ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.FindbugsCheck)
        }
        if (configuration.mavenExecuteStaticCodeChecks?.pmd == null || configuration.mavenExecuteStaticCodeChecks?.pmd == true) {
            recordIssues(failedTotalHigh: 1,
                failedTotalNormal: 10,
                blameDisabled: true,
                enabledForFailure: true,
                aggregatingResults: false,
                tool: pmdParser(pattern: '**/target/pmd.xml'))

            ReportAggregator.instance.reportStaticCodeExecution(QualityCheck.PmdCheck)
        }
    }
}
