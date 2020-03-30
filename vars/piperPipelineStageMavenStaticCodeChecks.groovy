import com.sap.piper.ConfigurationLoader
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.QualityCheck
import com.sap.piper.ReportAggregator

import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
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
        mavenExecuteStaticCodeChecks(script: script)
    }
}
