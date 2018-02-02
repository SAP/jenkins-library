import static java.util.Arrays.asList

import com.cloudbees.groovy.cps.NonCPS

/**
 * testResultsPublish
 *
 * @param script global script environment of the Jenkinsfile run
 * @param others document all parameters
 */
def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: 'testResultsPublish', stepParameters: parameters) {
        // GENERAL
        def allowUnstableBuilds = parameters.get('allowUnstableBuilds', false)
        // UNIT TESTS
        def junit = parameters.get('junit', false)
        // CODE COVERAGE
        def jacoco = parameters.get('jacoco', false)
        def cobertura = parameters.get('cobertura', false)
        // PERFORMANCE
        def jmeter = parameters.get('jmeter', false)

        // jUnit
        publishJUnitReport(junit)
        publishJacocoReport(jacoco)
        publishCoberturaReport(cobertura)
        publishJMeterReport(jmeter)

        if (!allowUnstableBuilds)
            failUnstableBuild(currentBuild)
    }
}

def publishJUnitReport(Map settings = [:]) {
    if(!Boolean.FALSE.equals(settings)){
        settings = asMap(settings)
        def pattern = settings.get('pattern', '**/target/surefire-reports/*.xml')
        def archive = settings.get('archive', false)
        def allowEmpty = settings.get('allowEmptyResults', true)
        def updateResults = settings.get('updateResults', false)

        if (updateResults)
            touchFiles(pattern)
        junit(
            testResults: pattern,
            allowEmptyResults: allowEmpty,
            healthScaleFactor: 100.0,
        )
        // archive results
        archiveResults(archive, pattern, allowEmpty)
    }
}

def publishJacocoReport(Map settings = [:]) {
    if(!Boolean.FALSE.equals(settings)){
        settings = asMap(settings)
        def pattern = settings.get('pattern', '**/target/*.exec')
        def archive = settings.get('archive', false)
        def allowEmpty = settings.get('allowEmptyResults', true)
        def include = settings.get('include', '')
        def exclude = settings.get('exclude', '')

        jacoco(
            execPattern: pattern,
            inclusionPattern: include,
            exclusionPattern: exclude
        )

        // archive results
        archiveResults(archive, pattern, allowEmpty)
    }
}

def publishCoberturaReport(Map settings = [:]) {
    if(!Boolean.FALSE.equals(settings)){
        settings = asMap(settings)
        def pattern = settings.get('pattern', '**/target/coverage/cobertura-coverage.xml')
        def archive = settings.get('archive', false)
        def allowEmpty = settings.get('allowEmptyResults', true)
        def onlyStable = settings.get('onlyStableBuilds', true)

        cobertura(
            coberturaReportFile: pattern,
            onlyStable: (onlyStable?true:false),
            failNoReports: (allowEmpty?false:true),
            failUnstable: false,
            failUnhealthy: false,
            autoUpdateHealth: false,
            autoUpdateStability: false,
            maxNumberOfBuilds: 0
        )

        // archive results
        archiveResults(archive, pattern, allowEmpty)
    }
}

// publish Performance Report using "Jenkins Performance Plugin" https://wiki.jenkins.io/display/JENKINS/Performance+Plugin
def publishJMeterReport(Map settings = [:]){
    if(!Boolean.FALSE.equals(settings)){
        settings = asMap(settings)
        def pattern = settings.get('pattern', '**/*.jtl')
        def archive = settings.get('archive', false)
        def allowEmpty = settings.get('allowEmptyResults', true)

        step([
            $class: 'PerformancePublisher',
            errorFailedThreshold: settings.get('errorFailedThreshold', 20),
            errorUnstableThreshold: settings.get('errorUnstableThreshold', 10),
            errorUnstableResponseTimeThreshold: settings.get('errorUnstableResponseTimeThreshold', ""),
            relativeFailedThresholdPositive: settings.get('relativeFailedThresholdPositive', 0),
            relativeFailedThresholdNegative: settings.get('relativeFailedThresholdNegative', 0),
            relativeUnstableThresholdPositive: settings.get('relativeUnstableThresholdPositive', 0),
            relativeUnstableThresholdNegative: settings.get('relativeUnstableThresholdNegative', 0),
            modePerformancePerTestCase: false,
            modeOfThreshold: settings.get('modeOfThreshold', false),
            modeThroughput: settings.get('modeThroughput', false),
            nthBuildNumber: settings.get('nthBuildNumber', 0),
            configType: settings.get('configType', "PRT"),
            failBuildIfNoResultFile: settings.get('failBuildIfNoResultFile', false),
            compareBuildPrevious: settings.get('compareBuildPrevious', true),
            parsers: asList(getJMeterParser().newInstance(pattern))
        ])

        // archive results
        archiveResults(archive, pattern, allowEmpty)
    }
}

def touchFiles(){
    echo 'update test results'
    def patternArray = pattern.split(',')
    for(def i = 0; i < patternArray.length; i++){
        sh "find . -wholename '${patternArray[i].trim()}' -exec touch {} \\;"
    }
}

def failUnstableBuild(currentBuild) {
    if (currentBuild.result == 'UNSTABLE') {
        echo "Current Build Status: ${currentBuild.result}"
        currentBuild.result = 'FAILURE'
        error 'Some tests failed!'
    }
}

def getJMeterParser(){
    // handle package renaming of JMeterParser class
    try {
        return this.class.classLoader.loadClass("hudson.plugins.performance.parsers.JMeterParser")
    } catch (Exception e) {
        return this.class.classLoader.loadClass("hudson.plugins.performance.JMeterParser")
    }
}

@NonCPS
def asMap(parameter) {
    return Boolean.TRUE.equals(parameter)
        ?[:]
        :parameter
}

def archiveResults(archive, pattern, allowEmpty) {
    if(archive){
        echo "archive ${pattern}"
        archiveArtifacts artifacts: pattern, allowEmptyArchive: allowEmpty
    }
}
