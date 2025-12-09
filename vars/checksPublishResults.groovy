import static com.sap.piper.Prerequisites.checkScript

import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.MapUtils
import com.sap.piper.Utils

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field Set TOOLS = [
    /**
     * Allows to publish the check results.
     * @possibleValues `true`, `false`, `Map`
     */
    'aggregation',
    /**
     * Searches and publishes TODOs in files with the [Task Scanner Plugin](https://wiki.jenkins-ci.org/display/JENKINS/Task+Scanner+Plugin).
     * @possibleValues `true`, `false`, `Map`
     */
    'tasks',
    /**
     * Publishes PMD findings with the [PMD plugin](https://plugins.jenkins.io/pmd).
     * @possibleValues `true`, `false`, `Map`
     */
    'pmd',
    /**
     * Publishes CPD findings with the [DRY plugin](https://plugins.jenkins.io/dry).
     * @possibleValues `true`, `false`, `Map`
     */
    'cpd',
    /**
     * Publishes Findbugs findings with the [Findbugs plugin](https://plugins.jenkins.io/findbugs).
     * @possibleValues `true`, `false`, `Map`
     */
    'findbugs',
    /**
     * Publishes Checkstyle findings with the [Checkstyle plugin](https://plugins.jenkins.io/checkstyle).
     * @possibleValues `true`, `false`, `Map`
     */
    'checkstyle',
    /**
     * Publishes ESLint findings (in [JSLint format](https://eslint.org/docs/user-guide/formatters/)) with the [Warnings plugin](https://plugins.jenkins.io/warnings).
     * @possibleValues `true`, `false`, `Map`
     */
    'eslint',
    /**
     * Publishes PyLint findings with the [Warnings plugin](https://plugins.jenkins.io/warnings), pylint needs to run with `--output-format=parseable` option.
     * @possibleValues `true`, `false`, `Map`
     */
    'pylint'
]

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = TOOLS.plus([
    /**
     * If it is set to `true` the step will archive reports matching the tool specific pattern.
     * @possibleValues `true`, `false`
     */
    'archive',
    /**
     * If it is set to `true` the step will fail the build if JUnit detected any failing tests.
     * @possibleValues `true`, `false`
     */
    'failOnError'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step can publish static check results from various sources.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        prepare(parameters)
        String stageName = parameters.stageName ?: env.STAGE_NAME
        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        if (configuration.aggregation && configuration.aggregation.active != false){
            error "[ERROR] Configuration of the aggregation view is no longer possible. Migrate any thresholds defined here to tool specific quality gates. (piper-lib/${STEP_NAME})"
        }

        // JAVA
        if(configuration.pmd.active) {
          report(pmdParser(createToolOptions(configuration.pmd)), configuration.pmd, configuration.archive)
        }
        if(configuration.cpd.active) {
          report(cpd(createToolOptions(configuration.cpd)), configuration.cpd, configuration.archive)
        }
        if(configuration.findbugs.active) {
          report(findBugs(createToolOptions(configuration.findbugs, [useRankAsPriority: true])), configuration.findbugs, configuration.archive)
        }
        if(configuration.checkstyle.active) {
          report(checkStyle(createToolOptions(configuration.checkstyle)), configuration.checkstyle, configuration.archive)
        }
        // JAVA SCRIPT
        if(configuration.eslint.active) {
          report(esLint(createToolOptions(configuration.eslint)), configuration.eslint, configuration.archive)
        }
        // PYTHON
        if(configuration.pylint.active) {
          report(pyLint(createToolOptions(configuration.pylint)), configuration.pylint, configuration.archive)
        }
        // GENERAL
        if(configuration.tasks.active) {
          report(taskScanner(createToolOptions(configuration.tasks, [
              includePattern: configuration.tasks.get('pattern'),
              highTags: configuration.tasks.get('high'),
              normalTags: configuration.tasks.get('normal'),
              lowTags: configuration.tasks.get('low'),
          ]).minus([pattern: configuration.tasks.get('pattern')])), configuration.tasks, configuration.archive)
        }
        if (configuration.failOnError && 'FAILURE' == script.currentBuild?.result){
            script.currentBuild.result = 'FAILURE'
            error "[${STEP_NAME}] Some checks failed!"
        }
    }
}

def report(tool, settings, doArchive){
    def options = createOptions(settings).plus([tools: [tool]])
    echo "recordIssues OPTIONS: ${options}"
    try {
        // publish
        recordIssues(options)
    } catch (e) {
        echo "recordIssues has failed. Possibly due to an outdated version of the warnings-ng plugin."
        e.printStackTrace()
    }
    // archive check results
    archiveResults(doArchive && settings.get('archive'), settings.get('pattern'), true)
}

def archiveResults(archive, pattern, allowEmpty){
    if(archive){
        echo "[${STEP_NAME}] archive ${pattern}"
        archiveArtifacts artifacts: pattern, allowEmptyArchive: allowEmpty
    }
}

@NonCPS
def createOptions(settings){
    Map result = [:]
    result.put('skipBlames', true)
    result.put('enabledForFailure', true)
    result.put('aggregatingResults', false)

    def qualityGates = []
    if (settings.qualityGates)
        qualityGates = qualityGates.plus(settings.qualityGates)

    // handle legacy thresholds
    // https://github.com/jenkinsci/warnings-ng-plugin/blob/6602c3a999b971405adda15be03979ce21cb3cbf/plugin/src/main/java/io/jenkins/plugins/analysis/core/util/QualityGate.java#L186
    def thresholdsList = settings.get('thresholds', [:])
    if (thresholdsList) {
        for (String status : ['fail', 'unstable']) {
            def thresholdsListPerStatus = thresholdsList.get(status, [:])
            if (thresholdsListPerStatus) {
                for (String severity : ['all', 'high', 'normal', 'low']) {
                    def threshold = thresholdsListPerStatus.get(severity)
                    if (threshold == null)
                        continue
                    threshold = threshold.toInteger() + 1
                    def type = "TOTAL"
                    if(severity != 'all')
                        type += "_" + severity.toUpperCase()
                    def gate = [threshold: threshold, type: type, unstable: status == 'unstable']
                    echo "[WARNING] legacy threshold found, please migrate to quality gate (piper-lib/checksPublishResults)"
                    echo "legacy threshold transformed to quality gate: ${gate}"
                    qualityGates = qualityGates.plus([gate])
                }
            }
        }
    }

    result.put('qualityGates', qualityGates)
    // filter empty values
    result = result.findAll {
        return it.value != null && it.value != ''
    }
    return result
}

@NonCPS
def createToolOptions(settings, additionalOptions = [:]){
    Map result = [pattern: settings.get('pattern')]
    if (settings.id)
        result.put('id', settings.id)
    if (settings.name)
        result.put('name', settings.name)
    result = result.plus(additionalOptions)
    // filter empty values
    result = result.findAll {
        return it.value != null && it.value != ''
    }
    return result
}

def prepare(parameters){
    // ensure tool maps are initialized correctly
    for(String tool : TOOLS){
        parameters[tool] = toMap(parameters[tool])
    }
    return parameters
}

def toMap(parameter){
    if(MapUtils.isMap(parameter))
        parameter.put('active', parameter.active == null?true:parameter.active)
    else if(Boolean.TRUE.equals(parameter))
        parameter = [active: true]
    else if(Boolean.FALSE.equals(parameter))
        parameter = [active: false]
    else
        parameter = [:]
    return parameter
}
