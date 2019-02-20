import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
//import com.sap.piper.Notify

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([

    /** Control the build technology, examples 'maven', 'npm' */
    'buildType',
    /** ... */
    'containerBuildOptions',
    /** ... */
    'containerCommand',
    /** ... */
    'containerShell',
    /** ... */
    'dockerOptions',
    /** ... */
    'dockerCommand',
    /** Image to be used for builds in case they run inside a Docker container */
    'dockerImage',
    /** ... */
    'dockerImageNameAndTag'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:], body = '') {
    handleStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .dependingOn('buildType').mixin('buildParameters')
            .mixin(
                dockerImageNameAndTag: script.globalPipelineEnvironment.getDockerImageNameAndTag()
            )
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('buildType').mixin('containerBuildOptions')
            .dependingOn('buildType').mixin('containerCommand')
            .dependingOn('buildType').mixin('containerShell')
            .dependingOn('buildType').mixin('dockerCommand')
            .dependingOn('buildType').mixin('dockerImage')
            .dependingOn('buildType').mixin('dockerOptions')
            .use()

        // report to SWA
        utils.pushToSWA([stepParam1: config.buildType, 'buildType': config.buildType], config)

        switch(config.buildType){
            case 'maven':
                mavenExecute script: this
                break
            case 'npm':
                npmExecute
                break
            case 'dockerLocal':
                new ConfigurationHelper(config).withMandatoryProperty('dockerImageNameAndTag')
                def dockerBuildImage = docker.build(config.dockerImageNameAndTag, "${config.containerBuildOptions} .")
                script.globalPipelineEnvironment.setDockerBuildImage(dockerBuildImage)
                break
            //option 'kaniko' so far only suitable for PR-voting since no push to a registry
            //ToDo: push to registry
            case 'kaniko':
                dockerExecute(
                    script: this,
                    containerCommand: config.containerCommand,
                    containerShell: config.containerShell,
                    dockerImage: config.dockerImage,
                    dockerOptions: config.dockerOptions
                ) {
                    sh """#!${config.containerShell}
mv /kaniko/.docker/config.json /kaniko/.docker/config.json.bak
mv /kaniko/.config/gcloud/docker_credential_gcr_config.json /kaniko/.config/gcloud/docker_credential_gcr_config.json.bak
/kaniko/executor --dockerfile ${env.WORKSPACE}/Dockerfile --context ${env.WORKSPACE} ${config.containerBuildOptions}"""
                }
                break
            default:
                if (config.dockerImage && config.dockerCommand) {
                    dockerExecute(
                        dockerImage: config.dockerImage,
                        dockerOptions: config.dockerOptions
                    ) {
                        sh "${config.dockerCommand}"
                    }
                }
        }

        //ToDo: upload artifacts to binary repo or Jenkins archiving ...
    }
}
