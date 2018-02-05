import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger

import groovy.transform.Field

@Field def STEP_NAME = 'checksPublishResults'

/**
 * checksPublishResults
 *
 * @param others document all parameters
 */
def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]
        prepareDefaultValues script: script

        Map configurationKeys = [
            'aggregation': [
                'active': null,
                'healthy': null,
                'unHealthy': null,
                'thresholds': [
                    'fail': ['all': null,'low': null,'normal': null,'high': null],
                    'unstable': ['all': null,'low': null,'normal': null,'high': null]
                ]
            ],
            'tasks': [
                'pattern': null,
                'low': null,
                'normal': null,
                'high': null,
                'archive': null,
                'active': null,
                'healthy': null,
                'unHealthy': null,
                'thresholds': [
                    'fail': ['all': null,'low': null,'normal': null,'high': null],
                    'unstable': ['all': null,'low': null,'normal': null,'high': null]
                ]
            ],
            'pmd': [
                'pattern': null,
                'archive': null,
                'active': null,
                'healthy': null,
                'unHealthy': null,
                'thresholds': [
                    'fail': ['all': null,'low': null,'normal': null,'high': null],
                    'unstable': ['all': null,'low': null,'normal': null,'high': null]
                ]
            ],
            'cpd': [
                'pattern': null,
                'archive': null,
                'active': null,
                'healthy': null,
                'unHealthy': null,
                'thresholds': [
                    'fail': ['all': null,'low': null,'normal': null,'high': null],
                    'unstable': ['all': null,'low': null,'normal': null,'high': null]
                ]
            ],
            'findbugs': [
                'pattern': null,
                'archive': null,
                'active': null,
                'healthy': null,
                'unHealthy': null,
                'thresholds': [
                    'fail': ['all': null,'low': null,'normal': null,'high': null],
                    'unstable': ['all': null,'low': null,'normal': null,'high': null]
                ]
            ],
            'checkstyle': [
                'pattern': null,
                'archive': null,
                'active': null,
                'healthy': null,
                'unHealthy': null,
                'thresholds': [
                    'fail': ['all': null,'low': null,'normal': null,'high': null],
                    'unstable': ['all': null,'low': null,'normal': null,'high': null]
                ]
            ],
            'eslint': [
                'pattern': null,
                'archive': null,
                'active': null,
                'healthy': null,
                'unHealthy': null,
                'thresholds': [
                    'fail': ['all': null,'low': null,'normal': null,'high': null],
                    'unstable': ['all': null,'low': null,'normal': null,'high': null]
                ]
            ],
            'pylint': [
                'pattern': null,
                'archive': null,
                'active': null,
                'healthy': null,
                'unHealthy': null,
                'thresholds': [
                    'fail': ['all': null,'low': null,'normal': null,'high': null],
                    'unstable': ['all': null,'low': null,'normal': null,'high': null]
                ]
            ],
            'archive': null
        ]
        final Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, STEP_NAME)
        final Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, STEP_NAME)
        prepare(parameters)
        Map configuration = ConfigurationMerger.mergeDeepStructure(parameters, configurationKeys, stepConfiguration, configurationKeys, stepDefaults)

        def doArchive = configuration.get('archive')
        // JAVA
        report('PmdPublisher', configuration.get('pmd'), doArchive)
        report('DryPublisher', configuration.get('cpd'), doArchive)
        report('FindBugsPublisher', configuration.get('findbugs'), doArchive)
        report('CheckStylePublisher', configuration.get('checkstyle'), doArchive)
        // JAVA SCRIPT
        reportWarnings('JSLint', configuration.get('eslint'), doArchive)
        // PYTHON
        reportWarnings('PyLint', configuration.get('pylint'), doArchive)
        // GENERAL
        reportTasks(configuration.get('tasks'))
        aggregate(configuration.get('aggregation'))
    }
}

def aggregate(settings){
    if (settings.active) {
        def options = createCommonOptionsMap('AnalysisPublisher', settings)
        // publish
        step(options)
    }
}

def reportTasks(settings){
    if (settings.active) {
        def options = createCommonOptionsMap('TasksPublisher', settings)
        options.put('pattern', settings.get('pattern'))
        options.put('high', settings.get('high'))
        options.put('normal', settings.get('normal'))
        options.put('low', settings.get('low'))
        // publish
        step(options)
    }
}

def report(publisherName, settings, doArchive){
    if (settings.active) {
        def pattern = settings.get('pattern')
        def options = createCommonOptionsMap(publisherName, settings)
        options.put('pattern', pattern)
        // publish
        step(options)
        // archive check results
        archiveResults(doArchive && settings.get('archive'), pattern, true)
    }
}

def reportWarnings(parserName, settings, doArchive){
    if (settings.active) {
        def pattern = settings.get('pattern')
        def options = createCommonOptionsMap('WarningsPublisher', settings)
        options.put('parserConfigurations', [[
            parserName: parserName,
            pattern: pattern
        ]])
        // publish
        step(options)
        // archive check results
        archiveResults(doArchive && settings.get('archive'), pattern, true)
    }
}

@NonCPS
def isMap(object){
    return object in Map
}

@NonCPS
def toMap(parameter){
    if(isMap(parameter))
        parameter.put('active', true)
    else if(Boolean.TRUE.equals(parameter))
        parameter = [active: true]
    else if(Boolean.FALSE.equals(parameter))
        parameter = [active: false]
    else
        parameter = [:]
    return parameter
}

def archiveResults(archive, pattern, allowEmpty){
    if(archive){
        echo "[${STEP_NAME}] archive ${pattern}"
        archiveArtifacts artifacts: pattern, allowEmptyArchive: allowEmpty
    }
}

@NonCPS
def createCommonOptionsMap(publisherName, settings){
    Map result = [:]
    def thresholds = settings.get('thresholds')
    def fail = thresholds.get('fail')
    def unstable = thresholds.get('unstable')

    result.put('$class', publisherName)
    result.put('healthy', settings.get('healthy'))
    result.put('unHealthy', settings.get('unHealthy'))
    result.put('canRunOnFailed', true)
    result.put('failedTotalAll', '' + fail.get('all'))
    result.put('failedTotalHigh', '' + fail.get('high'))
    result.put('failedTotalNormal', '' + fail.get('normal'))
    result.put('failedTotalLow', '' + fail.get('low'))
    result.put('unstableTotalAll', '' + unstable.get('all'))
    result.put('unstableTotalHigh', '' + unstable.get('high'))
    result.put('unstableTotalNormal', '' + unstable.get('normal'))
    result.put('unstableTotalLow', '' + unstable.get('low'))

    return result
}

@NonCPS
def prepare(parameters){
    // ensure tool maps are initialized
    parameters.aggregation = toMap(parameters.aggregation)
    parameters.tasks = toMap(parameters.tasks)
    parameters.pmd = toMap(parameters.pmd)
    parameters.cpd = toMap(parameters.cpd)
    parameters.findbugs = toMap(parameters.findbugs)
    parameters.checkstyle = toMap(parameters.checkstyle)
    parameters.eslint = toMap(parameters.eslint)
    parameters.pylint = toMap(parameters.pylint)
    return parameters
}
