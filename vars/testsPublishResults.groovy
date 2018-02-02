import static java.util.Arrays.asList

import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger

import groovy.transform.Field

@Field def STEP_NAME = 'testsPublishResults'
/**
 * testResultsPublish
 *
 * @param script global script environment of the Jenkinsfile run
 * @param others document all parameters
 */
def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]
        prepareDefaultValues script: script
        Map configurationKeys = [
            'junit': [
                'patter': null,
                'updateResults': null,
                'allowEmptyResults': null,
                'archive': null,
                'active': null
            ],
            'jacoco': [
                'pattern': null,
                'include': null,
                'exclude': null,
                'allowEmptyResults': null,
                'archive': null,
                'active': null
            ],
            'cobertura': [
                'pattern': null,
                'onlyStableBuilds': null,
                'allowEmptyResults': null,
                'archive': null,
                'active': null
            ],
            'jmeter': [
                'pattern': null,
                'errorFailedThreshold': null,
                'errorUnstableThreshold': null,
                'errorUnstableResponseTimeThreshold': null,
                'relativeFailedThresholdPositive': null,
                'relativeFailedThresholdNegative': null,
                'relativeUnstableThresholdPositive': null,
                'relativeUnstableThresholdNegative': null,
                'modeOfThreshold': null,
                'modeThroughput': null,
                'nthBuildNumber': null,
                'configType': null,
                'failBuildIfNoResultFile': null,
                'compareBuildPrevious': null,
                'allowEmptyResults': null,
                'archive': null,
                'active': null
            ]
        ]
        final Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, STEP_NAME)
        final Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, STEP_NAME)
        prepare(parameters)
        Map configuration = ConfigurationMerger.mergeDeepStructure(
            parameters, configurationKeys,
            stepConfiguration, configurationKeys, stepDefaults)
        // UNIT TESTS
        publishJUnitReport(configuration.get('junit'))
        // CODE COVERAGE
        publishJacocoReport(configuration.get('jacoco'))
        publishCoberturaReport(configuration.get('cobertura'))
        // PERFORMANCE
        publishJMeterReport(configuration.get('jmeter'))
    }
}

def publishJUnitReport(Map settings = [:]) {
    if(settings.active){
        def pattern = settings.get('pattern')
        def allowEmpty = settings.get('allowEmptyResults')

        if (settings.get('updateResults'))
            touchFiles(pattern)
        junit(
        //step([
        //    $class: 'JUnitResultArchiver',
            testResults: pattern,
            allowEmptyResults: allowEmpty,
            healthScaleFactor: 100.0,
        )
        // archive results
        archiveResults(settings.get('archive'), pattern, allowEmpty)
    }
}

def publishJacocoReport(Map settings = [:]) {
    if(settings.active){
        def pattern = settings.get('pattern')
        def allowEmpty = settings.get('allowEmptyResults')

        jacoco(
        //step([
        //    $class: 'JacocoPublisher',
            execPattern: pattern,
            inclusionPattern: settings.get('include'),
            exclusionPattern: settings.get('exclude')
        )
        // archive results
        archiveResults(settings.get('archive'), pattern, allowEmpty)
    }
}

def publishCoberturaReport(Map settings = [:]) {
    if(settings.active){
        def pattern = settings.get('pattern')
        def allowEmpty = settings.get('allowEmptyResults')

        cobertura(
        //step([
        //    $class: 'CoberturaPublisher',
            coberturaReportFile: pattern,
            onlyStable: settings.get('onlyStableBuilds'),
            failNoReports: !allowEmpty,
            failUnstable: false,
            failUnhealthy: false,
            autoUpdateHealth: false,
            autoUpdateStability: false,
            maxNumberOfBuilds: 0
        )
        // archive results
        archiveResults(settings.get('archive'), pattern, allowEmpty)
    }
}

// publish Performance Report using "Jenkins Performance Plugin" https://wiki.jenkins.io/display/JENKINS/Performance+Plugin
def publishJMeterReport(Map settings = [:]){
    if(settings.active){
        def pattern = settings.get('pattern')

        step([
            $class: 'PerformancePublisher',
            errorFailedThreshold: settings.get('errorFailedThreshold'),
            errorUnstableThreshold: settings.get('errorUnstableThreshold'),
            errorUnstableResponseTimeThreshold: settings.get('errorUnstableResponseTimeThreshold'),
            relativeFailedThresholdPositive: settings.get('relativeFailedThresholdPositive'),
            relativeFailedThresholdNegative: settings.get('relativeFailedThresholdNegative'),
            relativeUnstableThresholdPositive: settings.get('relativeUnstableThresholdPositive'),
            relativeUnstableThresholdNegative: settings.get('relativeUnstableThresholdNegative'),
            modePerformancePerTestCase: false,
            modeOfThreshold: settings.get('modeOfThreshold'),
            modeThroughput: settings.get('modeThroughput'),
            nthBuildNumber: settings.get('nthBuildNumber'),
            configType: settings.get('configType'),
            failBuildIfNoResultFile: settings.get('failBuildIfNoResultFile'),
            compareBuildPrevious: settings.get('compareBuildPrevious'),
            parsers: asList(getJMeterParser().newInstance(pattern))
        ])
        // archive results
        archiveResults(settings.get('archive'), pattern, settings.get('allowEmptyResults'))
    }
}

def touchFiles(){
    echo "[${STEP_NAME}] update test results"
    def patternArray = pattern.split(',')
    for(def i = 0; i < patternArray.length; i++){
        sh "find . -wholename '${patternArray[i].trim()}' -exec touch {} \\;"
    }
}

@NonCPS
def getJMeterParser(){
    // handle package renaming of JMeterParser class
    try {
        return this.class.classLoader.loadClass("hudson.plugins.performance.parsers.JMeterParser")
    } catch (Exception e) {
        return this.class.classLoader.loadClass("hudson.plugins.performance.JMeterParser")
    }
}

def archiveResults(archive, pattern, allowEmpty) {
    if(archive){
        echo "[${STAP_NAME}] archive ${pattern}"
        archiveArtifacts artifacts: pattern, allowEmptyArchive: allowEmpty
    }
}

@NonCPS
def prepare(parameters){
    // ensure tool maps are initialized
    parameters.junit = toMap(parameters.junit)
    parameters.jacoco = toMap(parameters.jacoco)
    parameters.cobertura = toMap(parameters.cobertura)
    parameters.jmeter = toMap(parameters.jmeter)
    return parameters
}

@NonCPS
def toMap(settings){
    if(isMap(settings))
        settings.put('active', true)
    else if(Boolean.TRUE.equals(settings))
        settings = [active: true]
    else
        settings = [active: false]
    return settings
}

@NonCPS
def isMap(object){
    return object in Map
}
