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

        // jUnit
        publishJUnit(junit)
        publishJacoco(jacoco)
        publishCobertura(cobertura)

        if (!allowUnstableBuilds)
            failUnstableBuild(currentBuild)
    }
}

def publishJUnit(Map settings = [:]) {
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

def publishJacoco(Map settings = [:]) {
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

def publishCobertura(Map settings = [:]) {
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

def touchFiles(){
    echo 'update test results'
    //def patternArray = pattern.split(',')
    //for(def i = 0; i < patternArray.length; i++){
    for (String p : pattern.split(',')) {
        sh "find . -wholename '${p.trim()}' -exec touch {} \\;"
    }
}

def failUnstableBuild(currentBuild) {
    if (currentBuild.result == 'UNSTABLE') {
        echo "Current Build Status: ${currentBuild.result}"
        currentBuild.result = 'FAILURE'
        error 'Some tests failed!'
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
