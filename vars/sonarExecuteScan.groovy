import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Pull-Request voting only:
     * The URL to the Github API. see https://docs.sonarqube.org/display/PLUG/GitHub+Plugin#GitHubPlugin-Usage
     * deprecated: only supported in LTS / < 7.2
     */
    'githubApiUrl',
    /**
     * Pull-Request voting only:
     * The Github organization.
     * @default: `commonPipelineEnvironment.getGithubOrg()`
     */
    'githubOrg',
    /**
     * Pull-Request voting only:
     * The Github repository.
     * @default: `commonPipelineEnvironment.getGithubRepo()`
     */
    'githubRepo',
    /**
     * Pull-Request voting only:
     * The Jenkins credentialId for a Github token. It is needed to report findings back to the pull-request.
     * deprecated: only supported in LTS / < 7.2
     * @possibleValues Jenkins credential id
     */
    'githubTokenCredentialsId',
    /**
     * The Jenkins credentialsId for a SonarQube token. It is needed for non-anonymous analysis runs. see https://sonarcloud.io/account/security
     * @possibleValues Jenkins credential id
     */
    'sonarTokenCredentialsId',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Pull-Request voting only:
     * Disables the pull-request decoration with inline comments.
     * deprecated: only supported in LTS / < 7.2
     * @possibleValues `true`, `false`
     */
    'disableInlineComments',
    /**
     * Name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system.
     * see dockerExecute
     */
    'dockerImage',
    /**
     * The name of the SonarQube instance defined in the Jenkins settings.
     */
    'instance',
    /**
     * Pull-Request voting only:
     * Activates the pull-request handling using the [GitHub Plugin](https://docs.sonarqube.org/display/PLUG/GitHub+Plugin) (deprecated).
     * deprecated: only supported in LTS / < 7.2
     * @possibleValues `true`, `false`
     */
    'legacyPRHandling',
    /**
     * A list of options which are passed to the `sonar-scanner`.
     */
    'options',
    /**
     * Organization that the project will be assigned to in SonarCloud.io.
     */
    'organization',
    /**
     * The project version that is reported to SonarQube.
     * @default: major number of `commonPipelineEnvironment.getArtifactVersion()`
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
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, GENERAL_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('projectVersion', script.commonPipelineEnvironment.getArtifactVersion()?.tokenize('.')?.get(0))
            .addIfEmpty('githubOrg', script.commonPipelineEnvironment.getGithubOrg())
            .addIfEmpty('githubRepo', script.commonPipelineEnvironment.getGithubRepo())
            // check mandatory parameters
            .withMandatoryProperty('githubTokenCredentialsId', null, { config -> config.legacyPRHandling && isPullRequest() })
            .withMandatoryProperty('githubOrg', null, { isPullRequest() })
            .withMandatoryProperty('githubRepo', null, { isPullRequest() })
            .use()

        def worker = { config ->
            withSonarQubeEnv(config.instance) {
                loadSonarScanner(config)

                if(config.organization) config.options.add("sonar.organization=${config.organization}")
                if(config.projectVersion) config.options.add("sonar.projectVersion=${config.projectVersion}")
                // prefix options
                config.options = config.options.collect { it.startsWith('-D') ? it : "-D${it}" }

                sh "PATH=\$PATH:${env.WORKSPACE}/.sonar-scanner/bin sonar-scanner ${config.options.join(' ')}"
            }
        }

        if(configuration.sonarTokenCredentialsId){
            def workerForSonarAuth = worker
            worker = { config ->
                withCredentials([string(
                    credentialsId: config.sonarTokenCredentialsId,
                    variable: 'SONAR_TOKEN'
                )]){
                    config.options.add("sonar.login=$SONAR_TOKEN")
                    workerForSonarAuth(config)
                }
            }
        }

        if(isPullRequest()){
            def workerForGithubAuth = worker
            worker = { config ->
                if(config.legacyPRHandling) {
                    withCredentials([string(
                        credentialsId: config.githubTokenCredentialsId,
                        variable: 'GITHUB_TOKEN'
                    )]){
                        // support for https://docs.sonarqube.org/display/PLUG/GitHub+Plugin
                        config.options.add('sonar.analysis.mode=preview')
                        config.options.add("sonar.github.oauth=$GITHUB_TOKEN")
                        config.options.add("sonar.github.pullRequest=${env.CHANGE_ID}")
                        config.options.add("sonar.github.repository=${config.githubOrg}/${config.githubRepo}")
                        if(config.githubApiUrl) config.options.add("sonar.github.endpoint=${config.githubApiUrl}")
                        if(config.disableInlineComments) config.options.add("sonar.github.disableInlineComments=${config.disableInlineComments}")
                        workerForGithubAuth(config)
                    }
                } else {
                    // see https://sonarcloud.io/documentation/analysis/pull-request/
                    config.options.add("sonar.pullrequest.key=${env.CHANGE_ID}")
                    config.options.add("sonar.pullrequest.base=${env.CHANGE_TARGET}")
                    config.options.add("sonar.pullrequest.branch=${env.BRANCH_NAME}")
                    config.options.add("sonar.pullrequest.provider=${config.pullRequestProvider}")
                    switch(config.pullRequestProvider){
                        case 'github':
                            config.options.add("sonar.pullrequest.github.repository=${config.githubOrg}/${config.githubRepo}")
                            break;
                        default: error "Pull-Request provider '${config.pullRequestProvider}' is not supported!"
                    }
                    workerForGithubAuth(config)
                }
            }
        }

        dockerExecute(
            script: script,
            dockerImage: configuration.dockerImage
        ){
            worker(configuration)
        }
    }
}

private Boolean isPullRequest(){
    return env.CHANGE_ID
}

private void loadSonarScanner(config){
    def filename = new File(config.sonarScannerDownloadUrl).getName()
    def foldername = filename.replace('.zip', '').replace('cli-', '')

    sh """
        curl --remote-name --remote-header-name --location --silent --show-error ${config.sonarScannerDownloadUrl}
        unzip -q ${filename}
        mv ${foldername} .sonar-scanner
    """
}
