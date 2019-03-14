import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    // DEPRECATED: SonarQube LTS
    'disableInlineComments', // voter only! set to true to only enable a summary comment on the pull-request
    'dockerImage', // the image to run the sonar-scanner
    // DEPRECATED: SonarQube LTS
    'githubApiUrl', // voter only! URL to access GitHub WS API | default: https://api.github.com
    'githubOrg', // voter only!
    'githubRepo', // voter only!
    // DEPRECATED: SonarQube LTS
    'githubTokenCredentialsId', // voter only!
    'instance', // the instance name of the Sonar server configured in the Jenkins
    'options',
    'projectVersion',
    'sonarTokenCredentialsId',
    'legacyPRHandling'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def utils = parameters.juStabUtils ?: new Utils()
        def script = checkScript(this, parameters) ?: this
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, GENERAL_CONFIG_KEYS)
            .mixin(
                projectVersion: script.commonPipelineEnvironment.getArtifactVersion()?.tokenize('.')?.get(0)
            )
            .mixin(parameters, PARAMETER_KEYS)
            // check mandatory parameters
            .withMandatoryProperty('githubTokenCredentialsId', null, { c -> return env.CHANGE_ID })
            .withMandatoryProperty('githubOrg', null, { c -> return env.CHANGE_ID })
            .withMandatoryProperty('githubRepo', null, { c -> return env.CHANGE_ID })
            .withMandatoryProperty('projectVersion', null, { c -> return !env.CHANGE_ID })
            .use()

        def worker = { c ->
            withSonarQubeEnv(c.instance) {
                loadSonarScanner(c)

                if(!env.CHANGE_ID && c.projectVersion) c.options.add("-Dsonar.projectVersion='${c.projectVersion}'")

                sh "PATH=${WORKSPACE}/.sonar-scanner/bin sonar-scanner ${c.options.join(' ')}"
            }
        }

        if(config.sonarTokenCredentialsId){
            def workerForSonarAuth = worker
            worker = { c ->
                withCredentials([string(
                    credentialsId: c.sonarTokenCredentialsId,
                    variable: 'SONAR_TOKEN'
                )]){
                    c.options.add(" -Dsonar.login=$SONAR_TOKEN")
                    workerForSonarAuth(c)
                }
            }
        }

        if(env.CHANGE_ID){
            def workerForGithubAuth = worker
            worker = { c ->
                if(c.legacyPRHandling) {
                    withCredentials([string(
                        credentialsId: c.githubTokenCredentialsId,
                        variable: 'GITHUB_TOKEN'
                    )]){
                        // support for https://docs.sonarqube.org/display/PLUG/GitHub+Plugin
                        c.options.add('-Dsonar.analysis.mode=preview')
                        c.options.add("-Dsonar.github.oauth=$GITHUB_TOKEN")
                        c.options.add("-Dsonar.github.pullRequest=${c.changeId}")
                        c.options.add("-Dsonar.github.repository=${c.githubOrg}/${c.githubRepo}")
                        if(c.githubApiUrl) c.options.add("-Dsonar.github.endpoint=${c.githubApiUrl}")
                        if(c.disableInlineComments) c.options.add("-Dsonar.github.disableInlineComments=${c.disableInlineComments}")
                    }
                } else {
                    // see https://sonarcloud.io/documentation/analysis/pull-request/
                    c.options.add("-Dsonar.pullrequest.key=${env.CHANGE_ID}")
                    c.options.add("-Dsonar.pullrequest.base=${env.CHANGE_TARGET}")
                    c.options.add("-Dsonar.pullrequest.branch=${env.BRANCH_NAME}")
                    c.options.add("-Dsonar.pullrequest.provider=${c.pullRequestProvider}")
                    switch(c.pullRequestProvider){
                        case 'github':
                            c.options.add("-Dsonar.pullrequest.github.repository=${c.githubOrg}/${c.githubRepo}")
                            break;
                        default: error "Pull-Request provider '${c.pullRequestProvider}' is not supported!"
                    }
                    workerForGithubAuth(c)
                }
            }
        }

        dockerExecute(
            script: script,
            dockerImage: config.dockerImage
        ){
            worker(config)
        }
    }
}

void loadSonarScanner(config){
    def filename = new File(config.sonarScannerUrl).getName()
    def foldername = filename.replace('.zip', '').replace('cli-', '')

    sh """
        curl --remote-name --remote-header-name --location --silent --show-error ${config.sonarScannerUrl}
        unzip -q ${filename}
        mv ${foldername} .sonar-scanner
    """
}
