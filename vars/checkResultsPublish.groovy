import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.Utils

def getStepName(){return 'checkResultsPublish'}

/**
 * checkResultsPublish
 *
 * @param others document all parameters
 */
def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: getStepName(), stepParameters: parameters) {
        // GENERAL
        def tasks = parameters.get('tasks', false)
        def aggregation = parameters.get('aggregation', [:])
        def doArchive = parameters.get('archive', false)
        // JAVA
        def pmd = parameters.get('pmd', false)
        def cpd = parameters.get('cpd', false)
        def findbugs = parameters.get('findbugs', false)
        def checkstyle = parameters.get('checkstyle', false)
        // JAVA SCRIPT
        def eslint = parameters.get('eslint', false)
        // PYTHON
        def pylint = parameters.get('pylint', false)

        // report TODOs
        reportTasks(tasks, '**/*.java')
        // report PMD
        report('PmdPublisher', pmd, '**/target/pmd.xml', doArchive)
        // report CPD
        report('DryPublisher', cpd, '**/target/cpd.xml', doArchive)
        // report Findbugs
        report('FindBugsPublisher', findbugs, '**/target/findbugsXml.xml, **/target/findbugs.xml', doArchive)
        // report Checkstyle
        report('CheckStylePublisher', checkstyle, '**/target/checkstyle-result.xml', doArchive)
        // report ESLint
        reportWarnings('JSLint', eslint, '**/eslint.xml', doArchive)
        // report PyLint
        reportWarnings('PyLint', pylint, '**/pylint.log', doArchive)

        // aggregate results
        aggregate(aggregation)
    }
}

def aggregate(settings){
    if (!Boolean.FALSE.equals(settings)) {
        settings = asMap(settings)
        def options = createCommonOptionsMap('AnalysisPublisher', settings)
        // publish
        step(options)
    }
}

def report(stepName, settings, defaultPattern, doArchive){
    // exit if set to FALSE
    if(!Boolean.FALSE.equals(settings)){
        settings = asMap(settings)
        def pattern = settings.get('pattern', defaultPattern)
        def options = createCommonOptionsMap(stepName, settings)
        options.put('pattern', pattern)
        // publish
        step(options)
        // archive check results
        archiveResults(doArchive && settings.get('archive', 'true'), pattern, true)
    }
}

def reportWarnings(parserName, settings, defaultPattern, doArchive){
    // exit if set to FALSE
    if(!Boolean.FALSE.equals(settings)){
        settings = asMap(settings)
        def pattern = settings.get('pattern', defaultPattern)
        def options = createCommonOptionsMap('WarningsPublisher', settings)
        options.put('parserConfigurations', [[
            parserName: parserName,
            pattern: pattern
        ]])
        // publish
        step(options)
        // archive check results
        archiveResults(doArchive && settings.get('archive', 'true'), pattern, true)
    }
}

def reportTasks(settings, defaultPattern){
    // exit if set to FALSE
    if(!Boolean.FALSE.equals(settings)){
        settings = asMap(settings)
        def options = createCommonOptionsMap('TasksPublisher', settings)
        options.put('pattern', settings.get('pattern', defaultPattern))
        options.put('high', settings.get('high', 'FIXME'))
        options.put('normal', settings.get('normal', 'TODO,REVISE,XXX'))
        options.put('low', settings.get('low', ''))
        // publish
        step(options)
    }
}

@NonCPS
def ensureMap(parameters, name){
    def value = parameters.get(name, [:])
    if(!isMap(value))
        error "Expected parameter ${name} to be a map."
    return value
}

@NonCPS
def asMap(parameter){
    if(Boolean.TRUE.equals(parameter))
        return [:]
    return parameter
}

@NonCPS
def isMap(object){
    return object in Map
}

def archiveResults(archive, pattern, allowEmpty){
    if(archive){
        echo "[${getStepName()}] archive ${pattern}"
        archiveArtifacts artifacts: pattern, allowEmptyArchive: allowEmpty
    }
}

@NonCPS
def createCommonOptionsMap(publisherName, settings){
    Map result = [:]
    def thresholds = ensureMap(settings, 'thresholds')
    def fail = ensureMap(thresholds, 'fail')
    def unstable = ensureMap(thresholds, 'unstable')

    result.put('$class', publisherName)
    result.put('healthy', settings.get('healthy', ''))
    result.put('unHealthy', settings.get('unHealthy', ''))
    result.put('canRunOnFailed', true)
    result.put('failedTotalAll', '' + fail.get('all', ''))
    result.put('failedTotalHigh', '' + fail.get('high', ''))
    result.put('failedTotalNormal', '' + fail.get('normal', ''))
    result.put('failedTotalLow', '' + fail.get('low', ''))
    result.put('unstableTotalAll', '' + unstable.get('all', ''))
    result.put('unstableTotalHigh', '' + unstable.get('high', ''))
    result.put('unstableTotalNormal', '' + unstable.get('normal', ''))
    result.put('unstableTotalLow', '' + unstable.get('low', ''))

    return result
}
