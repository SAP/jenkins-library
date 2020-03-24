import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.ReportAggregator
import com.sap.piper.Utils
import groovy.transform.Field

import static groovy.json.JsonOutput.toJson

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STAGE_STEP_KEYS = [
    /** Parameters for deployment to a Nexus Repository Manager. */
    'nexus',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This stage is responsible fpr releasing/deploying artifacts to a Nexus Repository Manager.<br />
 */
@GenerateStageDocumentation(defaultStageName = 'artifactDeployment')
def call(Map parameters = [:]) {
//    def script = checkScript(this, parameters) ?: this
//    def utils = parameters.juStabUtils ?: new Utils()
//
//    def stageName = parameters.stageName?:env.STAGE_NAME
//
//    Map config = ConfigurationHelper.newInstance(this)
//        .loadStepDefaults()
//        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
//        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
//        .mixin(parameters, PARAMETER_KEYS)
//        .addIfEmpty('nexus', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.nexus)
//        .use()
//
//    piperStageWrapper (script: script, stageName: stageName) {
//        // telemetry reporting
//        utils.pushToSWA([step: STEP_NAME], config)
//
//        if (!config.nexus) {
//            error("Can't deploy to nexus because the configuration is missing. " +
//                "Please ensure the `$stageName` section has a `nexus` sub-section.")
//        }
//
//        nexusUpload(nexusUploadParams)
//
//        ReportAggregator.instance.reportDeploymentToNexus()
//    }

    String stageName = 'artifactDeployment'
    final script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    piperStageWrapper(stageName: stageName, script: script) {
        Map defaultConfig = ConfigurationLoader.defaultStageConfiguration(script, stageName)
        Map stageConfig = ConfigurationLoader.stageConfiguration(script, stageName)

        def commonPipelineEnvironment = script.commonPipelineEnvironment
        List unstableSteps = commonPipelineEnvironment?.getValue('unstableSteps') ?: []
        if (unstableSteps) {
            piperPipelineStageConfirm script: script
            unstableSteps = []
            commonPipelineEnvironment.setValue('unstableSteps', unstableSteps)
        }

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], stageConfig)

        if (!stageConfig.nexus) {
            error("Can't deploy to nexus because the configuration is missing. " +
                "Please ensure the `$stageName` section has a `nexus` sub-section.")
        }

        Set nexusConfigKeys = [
            'dockerImage',
            'url',
            'repository',
            'version',
            'credentialsId',
            'additionalClassifiers',
            'artifactId',
            'groupId',
        ]

        Map nexusConfig = ConfigurationMerger.merge(stageConfig.nexus, nexusConfigKeys, defaultConfig.nexus)

        String url = nexusConfig.url
        if (!url) {
            error "You need to configure the key 'url' in the nexus configuration"
        }

        def nexusUrlWithoutProtocol = url.replaceFirst("^https?://", "")

        Map mavenConfig = ConfigurationMerger.merge(
            ConfigurationLoader.stepConfiguration(script, 'mavenExecute'),
            ['dockerImage', 'dockerOptions', 'globalSettingsFile'] as Set,
            ConfigurationLoader.defaultStepConfiguration(script, 'mavenExecute'))

        // Merge maven and nexus config for nexusUpload step
        Map nexusUploadParams = [
            script: script,
            url: nexusUrlWithoutProtocol,
            version: nexusConfig.version,
            repository: nexusConfig.repository,
            groupId: nexusConfig.groupId,

            m2Path: mavenConfig.m2Directory,
            globalSettingsFile: mavenConfig.globalSettingsFile,
            dockerImage: mavenConfig.dockerImage,
            dockerOptions: mavenConfig.dockerOptions
        ]

        // Set artifactId if configured, fall-back to artifactId from CPE if set
        if (nexusConfig.artifactId) {
            nexusUploadParams.artifactId = nexusConfig.artifactId
        } else if (script.commonPipelineEnvironment.configuration.artifactId) {
            nexusUploadParams.artifactId = script.commonPipelineEnvironment.configuration.artifactId
        }
        // Replace 'additionalClassifiers' List with JSON encoded String
        if (nexusConfig.additionalClassifiers) {
            nexusUploadParams.additionalClassifiers = "${toJson(nexusConfig.additionalClassifiers as List)}"
        }
        // Only add 'nexusCredentialsId' if 'credentialsId' is not empty
        if (nexusConfig.credentialsId) {
            nexusUploadParams.nexusCredentialsId = nexusConfig.credentialsId
        }

        // The withEnv wrapper can be removed before merging to master.
        withEnv(['REPOSITORY_UNDER_TEST=SAP/jenkins-library','LIBRARY_VERSION_UNDER_TEST=nexus-upload-cmd']) {
            nexusUpload(nexusUploadParams)
        }

        ReportAggregator.instance.reportDeploymentToNexus()
    }
}
