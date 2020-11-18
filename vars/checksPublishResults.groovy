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
        // PMD
        if (configuration.pmd.active) {
            def settings = configuration.pmd
            // publish
            def options = createCommonOptionsMap(settings)
            def toolOptions = createCommonToolOptionsMap(settings)
            options.put('tools', [pmdParser(toolOptions)])
            recordIssues(options)
            // archive check results
            archiveResults(configuration.archive && settings.get('archive'), settings.get('pattern'), true)
        }
        if (configuration.cpd.active) {
            def settings = configuration.cpd
            // publish
            def options = createCommonOptionsMap(settings)
            def toolOptions = createCommonToolOptionsMap(settings)
            options.put('tools', [cpd(toolOptions)])
            recordIssues(options)
            // archive check results
            archiveResults(configuration.archive && settings.get('archive'), settings.get('pattern'), true)
        }
        if (configuration.findbugs.active) {
            def settings = configuration.findbugs
            // publish
            def options = createCommonOptionsMap(settings)
            def toolOptions = createCommonToolOptionsMap(settings)
            options.put('tools', [findBugs(toolOptions.plus([useRankAsPriority: true]))])
            recordIssues(options)
            // archive check results
            archiveResults(configuration.archive && settings.get('archive'), settings.get('pattern'), true)
        }
        if (configuration.checkstyle.active) {
            def settings = configuration.checkstyle
            // publish
            def options = createCommonOptionsMap(settings)
            def toolOptions = createCommonToolOptionsMap(settings)
            options.put('tools', [checkStyle(toolOptions)])
            recordIssues(options)
            // archive check results
            archiveResults(configuration.archive && settings.get('archive'), settings.get('pattern'), true)
        }
        // JAVA SCRIPT
        if (configuration.eslint.active) {
            def settings = configuration.eslint
            def pattern =
            // publish
            def options = createCommonOptionsMap(settings)
            def toolOptions = createCommonToolOptionsMap(settings)
            options.put('tools', [esLint(toolOptions)])
            recordIssues(options)
            // archive check results
            archiveResults(configuration.archive && settings.get('archive'), settings.get('pattern'), true)
        }
        //TODO: check if JSLint is needed
        // PYTHON
        if (configuration.pylint.active) {
            def settings = configuration.pylint
            def pattern = settings.get('pattern')
            // publish
            def options = createCommonOptionsMap(settings)
            def toolOptions = createCommonToolOptionsMap(settings)
            options.put('tools', [pyLint(toolOptions.plus([pattern: pattern]))])
            recordIssues(options)
            // archive check results
            archiveResults(configuration.archive && settings.get('archive'), pattern, true)
        }
        // GENERAL
        if (configuration.tasks.active) {
            def settings = configuration.tasks
            def pattern = settings.get('pattern')
            // publish
            def options = createCommonOptionsMap(settings)
            def toolOptions = createCommonToolOptionsMap(settings)
            options.put('tools', [taskScanner(toolOptions.plus([
                includePattern: pattern,
                highTags: settings.get('high'),
                normalTags: settings.get('normal'),
                lowTags: settings.get('low'),
            ]))])
            recordIssues(options)
            // archive check results
            archiveResults(configuration.archive && settings.get('archive'), pattern, true)
        }
    }
}

def archiveResults(archive, pattern, allowEmpty){
    if(archive){
        echo "[${STEP_NAME}] archive ${pattern}"
        archiveArtifacts artifacts: pattern, allowEmptyArchive: allowEmpty
    }
}

@NonCPS
def createCommonOptionsMap(settings){
    Map result = [:]
    result.put('blameDisabled', true)
    result.put('enabledForFailure', true)
    result.put('aggregatingResults', false)
    if (settings.qualityGates)
        result.put('qualityGates', settings.qualityGates)
    // filter empty values
    result = result.findAll {
        return it.value != null && it.value != ''
    }
    return result
}

@NonCPS
def createCommonToolOptionsMap(settings){
    Map result = [pattern: settings.get('pattern')]
    if (settings.id)
        result.put('id ', settings.id)
    if (settings.name)
        result.put('name', settings.name)
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
