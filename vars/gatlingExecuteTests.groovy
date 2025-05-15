import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = [
    /**
     * Defines the behavior in case tests fail. When this is set to `true` test results cannot be recorded using the `publishTestResults` step afterwards.
     * @possibleValues `true`, `false`
     */
    'failOnError',
    /**
     * Path to the pom.xml file containing the performance test Maven module, for example `performance-tests/pom.xml`.
     */
    'pomPath',
    /**
     * Optional List of app URLs and corresponding Jenkins credential IDs.
     */
    'appUrls'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this step Gatling performance tests are executed.
 * Requires the [Jenkins Gatling plugin](https://plugins.jenkins.io/gatling/) to be installed.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        Script script = checkScript(this, parameters) ?: this
        Utils utils = parameters.juStabUtils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('pomPath')
            .use()

        def appUrls = parameters.get('appUrls')
        if (appUrls && !(appUrls instanceof List)) {
            error "The optional parameter 'appUrls' needs to be a List of Maps, where each Map contains the two entries 'url' and 'credentialsId'."
        }

        if (!fileExists(config.pomPath)) {
            error "The file '${config.pomPath}' does not exist."
        }

        utils.unstashAll(config.stashContent)

        try {
            if (appUrls) {
                for (int i = 0; i < appUrls.size(); i++) {
                    def appUrl = appUrls.get(i)
                    if (!(appUrl instanceof Map)) {
                        error "The entry at index $i in 'appUrls' is not a Map. It needs to be a Map containing the two entries 'url' and 'credentialsId'."
                    }
                    executeTestsWithAppUrlAndCredentials(script, appUrl.url, appUrl.credentialsId, config.pomPath)
                }
            } else {
                mavenExecute script: script, flags: ['--update-snapshots'], pomPath: config.pomPath, goals: ['test']
            }
        } finally {
            gatlingArchive()
        }
    }
}

void executeTestsWithAppUrlAndCredentials(Script script, url, credentialsId, pomPath) {
    withCredentials([
        [
            $class: 'UsernamePasswordMultiBinding',
            credentialsId: credentialsId,
            passwordVariable: 'PERFORMANCE_TEST_PASSWORD',
            usernameVariable: 'PERFORMANCE_TEST_USERNAME'
        ]
    ]) {
        List defines = [
            "-DappUrl=$url",
            "-Dusername=$PERFORMANCE_TEST_USERNAME",
            "-Dpassword=$PERFORMANCE_TEST_PASSWORD"
        ]
        mavenExecute script: script, flags: ['--update-snapshots'], pomPath: pomPath, goals: ['test'], defines: defines
    }
}
