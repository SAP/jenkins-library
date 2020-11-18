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
@Field Set STEP_CONFIG_KEYS = TOOLS.plus(['archive'])
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

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        // JAVA
        report(pmdParser(createToolOptions(configuration.pmd)), configuration.pmd, configuration.archive)
        report(cpd(createToolOptions(configuration.cpd)), configuration.cpd, configuration.archive)
        report(findBugs(createToolOptions(configuration.findbugs, [useRankAsPriority: true])), configuration.findbugs, configuration.archive)
        report(checkStyle(createToolOptions(configuration.checkstyle)), configuration.checkstyle, configuration.archive)
        // JAVA SCRIPT
        //TODO: check if ESLint (in Checkstyle format) is sufficient or if JSLint is needed
        report(esLint(createToolOptions(configuration.eslint)), configuration.eslint, configuration.archive)
        // PYTHON
        report(pyLint(createToolOptions(configuration.pylint)), configuration.pylint, configuration.archive)
        // GENERAL
        report(taskScanner(createToolOptions(configuration.tasks, [
            includePattern: configuration.tasks.get('pattern'),
            highTags: configuration.tasks.get('high'),
            normalTags: configuration.tasks.get('normal'),
            lowTags: configuration.tasks.get('low'),
        ]).minus([pattern: configuration.tasks.get('pattern')])), configuration.tasks, configuration.archive)
    }
}

def report(tool, settings, doArchive){
    if (settings.active) {
        def options = createOptions(settings)
        // publish
        recordIssues(options.plus([tools: [tool]]))
        // archive check results
        archiveResults(doArchive && settings.get('archive'), settings.get('pattern'), true)
    }
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
    result.put('blameDisabled', true)
    result.put('enabledForFailure', true)
    result.put('aggregatingResults', false)
    if (settings.qualityGates)
        result.put('qualityGates', settings.qualityGates)


    def thresholds = settings.get('thresholds', [:])
    def fail = thresholds.get('fail', [:])
    def unstable = thresholds.get('unstable', [:])

    result.put('failedTotalAll', fail.get('all'))
    result.put('failedTotalHigh', fail.get('high'))
    result.put('failedTotalNormal', fail.get('normal'))
    result.put('failedTotalLow', fail.get('low'))
    result.put('unstableTotalAll', unstable.get('all'))
    result.put('unstableTotalHigh', unstable.get('high'))
    result.put('unstableTotalNormal', unstable.get('normal'))
    result.put('unstableTotalLow', unstable.get('low'))
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
        result.put('id ', settings.id)
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
