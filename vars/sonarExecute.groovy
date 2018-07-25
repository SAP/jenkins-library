import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'sonarExecute'
@Field Set STEP_CONFIG_KEYS = [
    'dockerImage', // the image to run the sonar-scanner
    'instance', // the instance name of the Sonar server configured in the Jenkins
    'options',
    'projectVersion',
    // needed for voter
    'disableInlineComments', // set to true to only enable a summary comment on the pull-request
    'isVoter', // enables the preview mode
    'changeId', // the pull-request number
    'githubApiUrl', // URL to access GitHub WS API | default: https://api.github.com
    'githubOrg',
    'githubRepo',
    'githubTokenCredentialsId'
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

        dockerExecute(
            dockerImage: config.dockerImage,
        ){
            withSonarQubeEnv(config.instance) {
                if(config.isVoter){
                    withCredentials([string(
                        credentialsId: config.githubTokenCredentialsId,
                        variable: 'GITHUB_TOKEN'
                    )]){
                        def options = [
                            '-Dsonar.analysis.mode=preview'
                            "-Dsonar.github.oauth=$GITHUB_TOKEN"
                            "-Dsonar.github.pullRequest=${config.changeId}"
                            "-Dsonar.github.repository=${config.githubOrg}/${config.githubRepo}"
                        ]
                        if(config.githubApiUrl)
                            options.push("-Dsonar.github.endpoint=${config.githubApiUrl}")
                        if(config.disableInlineComments)
                            options.push("-Dsonar.github.disableInlineComments=${config.disableInlineComments}")

                        if(body) body()
                        sh "sonar-scanner ${config.options} ${options.join(' ')}"
                    }
                }else{
                    if(body) body()
                    sh "sonar-scanner ${config.options}"
                }
            }
        }
    }
}
