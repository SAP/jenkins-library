import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.DownloadCacheUtils
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
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** Artifact ID of the main build artifact. */
    'artifactId',
    /** Group ID of the main build artifact. */
    'groupId',
    /** The docker image to use for executing the step. */
    'dockerImage',
    /** The options to be passed to docker when executing the step within a docker context. */
    'dockerOptions',
])

/**
 * This stage is responsible fpr releasing/deploying artifacts to a Nexus Repository Manager.<br />
 */
@GenerateStageDocumentation(defaultStageName = 'artifactDeployment')
def call(Map parameters = [:]) {
    String stageName = 'artifactDeployment'
    final script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .withMandatoryProperty('nexus')
        .use()

    piperStageWrapper(stageName: stageName, script: script) {

        def commonPipelineEnvironment = script.commonPipelineEnvironment
        List unstableSteps = commonPipelineEnvironment?.getValue('unstableSteps') ?: []
        if (unstableSteps) {
            piperPipelineStageConfirm script: script
            unstableSteps = []
            commonPipelineEnvironment.setValue('unstableSteps', unstableSteps)
        }

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        def nexusConfig = config.nexus

        // Add all mandatory parameters
        Map nexusUploadParams = [
            script: script,
            version: nexusConfig.version,
            url: nexusConfig.url,
            repository: nexusConfig.repository,
        ]

        // Add additional parameters only if they are set. This avoids the following problem:
        // Since the context parameters will be converted to JSON, put into an ENV variable,
        // and then decoded back using the piper binary's getConfig --contextConfig command,
        // key that were present, but had a null String value, will now in fact have a value of
        // type net.sf.json.JSONNull. This in turn will not evaluate to 'false' using Groovy
        // Truth as null Strings would have.
        if (config.dockerImage) {
            nexusUploadParams.dockerImage = config.dockerImage
        }
        if (config.dockerOptions) {
            nexusUploadParams.dockerOptions = config.dockerOptions
        }

        nexusUploadParams = DownloadCacheUtils.injectDownloadCacheInMavenParameters(script as Script, nexusUploadParams)

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

        // The withEnv wrapper can be removed before merging to master.
        withEnv(['REPOSITORY_UNDER_TEST=SAP/jenkins-library','LIBRARY_VERSION_UNDER_TEST=stage-artifact-deployment']) {
            nexusUpload(nexusUploadParams)
        }

        ReportAggregator.instance.reportDeploymentToNexus()
    }
}
