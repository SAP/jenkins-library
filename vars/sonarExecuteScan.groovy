import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * GitHub Plugin only:
     * The URL to the Github API. see https://docs.sonarqube.org/display/PLUG/GitHub+Plugin#GitHubPlugin-Usage
     */
    'githubApiUrl', // voter only! URL to access GitHub WS API | default: https://api.github.com
    /**
     * Pull-Request voting only:
     * The Github organization.
     */
    'githubOrg',
    /**
     * Pull-Request voting only:
     * The Github repository.
     */
    'githubRepo',
    /**
     * GitHub Plugin only:
     * The Jenkins credentialId for a Github token. It is needed to report findings back to the pull-request.
     */
    'githubTokenCredentialsId',
    /**
     * The Jenkins credentialsId for a SonarQube token. It is needed for non-anonymous analysis runs. see https://sonarcloud.io/account/security
     */
    'sonarTokenCredentialsId',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * GitHub Plugin only:
     * Set to true to only enable a summary comment on the pull-request.
     */
    'disableInlineComments',
    /**
     * Name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system.
     * @see dockerExecute
     */
    'dockerImage',
    /**
     * The name of the SonarQube instance defined in the Jenkins settings.
     */
    'instance',
    /**
     * Activated the pull-request handling using the [GitHub Plugin](https://docs.sonarqube.org/display/PLUG/GitHub+Plugin) (deprecated).
     * @possibleValues `true`, `false`
     */
    'legacyPRHandling',
    /**
     * A list of options which are passed to the `sonar-scanner`.
     */
    'options',
    /**
     * SonarCloud.io only:
     * Organization that the project will be assigned to.
     */
    'organizationKey',
    /**
     *
     */
    'projectVersion'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * The step executes the [sonar-scanner](https://docs.sonarqube.org/display/SCAN/Analyzing+with+SonarQube+Scanner) cli command to scan the defined sources and publish the results to a SonarQube instance.
 */
@GenerateDocumentation
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
            .withMandatoryProperty('githubTokenCredentialsId', null, { conf -> conf.legacyPRHandling && isPullRequest() })
            .withMandatoryProperty('githubOrg', null, { isPullRequest() })
            .withMandatoryProperty('githubRepo', null, { isPullRequest() })
            .use()

        def worker = { c ->
            withSonarQubeEnv(c.instance) {
                loadSonarScanner(c)

                if(c.projectVersion && !isPullRequest()) c.options.add("-Dsonar.projectVersion=${c.projectVersion}")
                if(c.organization) c.options.add("-Dsonar.organization=${c.organization}")

                // prefix options
                c.options = c.options.collect { it.startsWith('-D') ? it : "-D${it}" }

                sh "PATH=\$PATH:${env.WORKSPACE}/.sonar-scanner/bin sonar-scanner ${c.options.join(' ')}"
            }
        }

        if(config.sonarTokenCredentialsId){
            def workerForSonarAuth = worker
            worker = { c ->
                withCredentials([string(
                    credentialsId: c.sonarTokenCredentialsId,
                    variable: 'SONAR_TOKEN'
                )]){
                    c.options.add("-Dsonar.login=$SONAR_TOKEN")
                    workerForSonarAuth(c)
                }
            }
        }

        if(isPullRequest()){
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
                        c.options.add("-Dsonar.github.pullRequest=${env.changeId}")
                        c.options.add("-Dsonar.github.repository=${c.githubOrg}/${c.githubRepo}")
                        if(c.githubApiUrl) c.options.add("-Dsonar.github.endpoint=${c.githubApiUrl}")
                        if(c.disableInlineComments) c.options.add("-Dsonar.github.disableInlineComments=${c.disableInlineComments}")
                        workerForGithubAuth(c)
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

private Boolean isPullRequest(){
    return env.CHANGE_ID
}

private void loadSonarScanner(config){
    def filename = new File(config.sonarScannerUrl).getName()
    def foldername = filename.replace('.zip', '').replace('cli-', '')

    sh """
        curl --remote-name --remote-header-name --location --silent --show-error ${config.sonarScannerUrl}
        unzip -q ${filename}
        mv ${foldername} .sonar-scanner
    """
}
