import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'sonarExecute'
@Field Set STEP_CONFIG_KEYS = [
    'changeId', // voter only! the pull-request number
    'disableInlineComments', // voter only! set to true to only enable a summary comment on the pull-request
    'dockerImage', // the image to run the sonar-scanner
    'githubApiUrl', // voter only! URL to access GitHub WS API | default: https://api.github.com
    'githubOrg', // voter only!
    'githubRepo', // voter only!
    'githubTokenCredentialsId', // voter only!
    'instance', // the instance name of the Sonar server configured in the Jenkins
    'isVoter', // voter only! enables the preview mode
    'options',
    'projectVersion',
    'sonarTokenCredentialsId'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:], Closure body = null) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def utils = parameters.juStabUtils ?: new Utils()
        def script = parameters.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(
                projectVersion: script.commonPipelineEnvironment.getArtifactVersion()?.tokenize('.')?.get(0),
                githubOrg: script.commonPipelineEnvironment.getGithubOrg(),
                githubRepo: script.commonPipelineEnvironment.getGithubRepo(),
                changeId: env.CHANGE_ID
            )
            .mixin(parameters, PARAMETER_KEYS)
            .use()
        // check mandatory parameters
        if(config.isVoter){
            new ConfigurationHelper(config)
                .withMandatoryProperty('changeId')
                .withMandatoryProperty('githubTokenCredentialsId')
                .withMandatoryProperty('githubOrg')
                .withMandatoryProperty('githubRepo')
        }else{
            new ConfigurationHelper(config)
                .withMandatoryProperty('projectVersion')
        }
        // resolve templates
        config.options = SimpleTemplateEngine.newInstance().createTemplate(config.options).make([projectVersion: config.projectVersion]).toString()

        def worker = { c, b ->
            withSonarQubeEnv(c.instance) {
                if(b) b()
                sh "sonar-scanner ${c.options}"
            }
        }

        if(config.sonarTokenCredentialsId){
            def workerForSonarAuth = worker
            worker = { c, b ->
                withCredentials([string(
                    credentialsId: c.sonarTokenCredentialsId,
                    variable: 'SONAR_TOKEN'
                )]){
                    c.options += " -Dsonar.login=$SONAR_TOKEN"
                    workerForSonarAuth(c,b)
                }
            }
        }

        if(config.isVoter){
            def workerForGithubAuth = worker
            worker = { c, b ->
                withCredentials([string(
                    credentialsId: c.githubTokenCredentialsId,
                    variable: 'GITHUB_TOKEN'
                )]){
                    c.options += ' -Dsonar.analysis.mode=preview'
                    c.options += " -Dsonar.github.oauth=$GITHUB_TOKEN"
                    c.options += " -Dsonar.github.pullRequest=${c.changeId}"
                    c.options += " -Dsonar.github.repository=${c.githubOrg}/${c.githubRepo}"
                    if(c.githubApiUrl)
                        c.options += " -Dsonar.github.endpoint=${c.githubApiUrl}"
                    if(c.disableInlineComments)
                        c.options += " -Dsonar.github.disableInlineComments=${c.disableInlineComments}"

                    workerForGithubAuth(c,b)
                }
            }
        }

        dockerExecute(
            dockerImage: config.dockerImage
        ){
            worker(config, body)
        }
    }
}
